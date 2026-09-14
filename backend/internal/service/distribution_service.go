package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditledger"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrDistributionDisabled    = infraerrors.Forbidden("DISTRIBUTION_DISABLED", "distribution program is disabled")
	ErrPayoutAccountRequired   = infraerrors.BadRequest("PAYOUT_ACCOUNT_REQUIRED", "payout account is required")
	ErrWithdrawalAmountInvalid = infraerrors.BadRequest("WITHDRAWAL_AMOUNT_INVALID", "withdrawal amount is invalid")
	ErrWithdrawalLimitExceeded = infraerrors.TooManyRequests("WITHDRAWAL_LIMIT_EXCEEDED", "daily withdrawal limit exceeded")
	ErrWithdrawalInsufficient  = infraerrors.BadRequest("WITHDRAWAL_INSUFFICIENT", "insufficient commission balance")
	ErrWithdrawalStateInvalid  = infraerrors.Conflict("WITHDRAWAL_STATE_INVALID", "withdrawal state transition is invalid")
	ErrCommissionDebt          = infraerrors.Conflict("COMMISSION_DEBT_OUTSTANDING", "commission debt must be repaid before withdrawal")
	ErrReversalInvalid         = infraerrors.BadRequest("DISTRIBUTION_REVERSAL_INVALID", "invalid distribution reversal request")
)

// 「推广计划」结算深度说明
//
// 历史沿革：该程序（code 仍为 compute_company）原本是 5 层多级分销，按团队业绩
// 分 T0~T3 逐层拨出，合计 10%~40%；其后一度收敛为「单层 + 三档阶梯（5%/8%/10%）」。
//
// 现在起点更简：推广计划只有**一层**，且**不设任何档位**——返佣只有一个来源，
// 就是被购买的商品在上架时设置的返佣比例（shop_products.commission_bps，
// 下单时快照到 shop_orders.snapshot_commission_bps）。
//
// distribution_relations 表仍保留 0~5 层的完整数据（历史佣金与账本要能追溯），
// 但业务结算一律只看 depth = 1。

type DistributionService struct {
	db        *sql.DB
	encryptor SecretEncryptor
}

type DistributionProcessResult struct {
	Enabled         bool
	StackWithLegacy bool
}

// DistributionDashboard 是推广员后台的概览数据。
//
// 返佣只有一个来源：商城商品订单（按商品上架时设置的返佣比例结算），
// 因此这里没有档位、没有层级构成。金额一律是人民币（minor），
// 与只能调 API 的美元额度（user_credit_accounts）是两种钱，不要混用。
type DistributionDashboard struct {
	Enabled                   bool   `json:"enabled"`
	BalanceRechargeMultiplier string `json:"balance_recharge_multiplier"`
	USDToCNYRate              string `json:"usd_to_cny_rate"`
	CommissionFreezeHours     int    `json:"commission_freeze_hours"`
	WithdrawalMinMinor        int64  `json:"withdrawal_min_cny_minor"`
	WithdrawalDailyLimit      int    `json:"withdrawal_daily_limit"`
	// InviteeCount = 直接邀请人数；InviteeSpendMinor = 这些人在商城的实付总额。
	InviteeCount      int64 `json:"invitee_count"`
	InviteeSpendMinor int64 `json:"team_volume_cny_minor"`
	AvailableMinor    int64 `json:"available_cny_minor"`
	FrozenMinor       int64 `json:"frozen_cny_minor"`
	WithdrawingMinor  int64 `json:"withdrawing_cny_minor"`
	DebtMinor         int64 `json:"debt_cny_minor"`
	LifetimeMinor     int64 `json:"lifetime_earned_cny_minor"`
}

type DistributionAnalyticsPoint struct {
	Date            string `json:"date"`
	SpendMinor      int64  `json:"spend_cny_minor"`
	CommissionMinor int64  `json:"commission_cny_minor"`
}

type DistributionAnalyticsSummary struct {
	SpendMinor              int64   `json:"spend_cny_minor"`
	CommissionMinor         int64   `json:"commission_cny_minor"`
	PreviousSpendMinor      int64   `json:"previous_spend_cny_minor"`
	PreviousCommissionMinor int64   `json:"previous_commission_cny_minor"`
	SpendGrowthPercent      float64 `json:"spend_growth_percent"`
	CommissionGrowthPercent float64 `json:"commission_growth_percent"`
}

type DistributionForecastHorizon struct {
	Eligible                 bool    `json:"eligible"`
	Reason                   string  `json:"reason,omitempty"`
	EstimatedSpendMinor      int64   `json:"estimated_spend_cny_minor"`
	EstimatedCommissionMinor int64   `json:"estimated_commission_cny_minor"`
	SpendGrowthPercent       float64 `json:"spend_growth_percent"`
	CommissionGrowthPercent  float64 `json:"commission_growth_percent"`
}

type DistributionForecast struct {
	Method     string                      `json:"method"`
	SevenDays  DistributionForecastHorizon `json:"seven_days"`
	ThirtyDays DistributionForecastHorizon `json:"thirty_days"`
}

type DistributionAnalytics struct {
	AsOf      time.Time                    `json:"as_of"`
	RangeDays int                          `json:"range_days"`
	Series    []DistributionAnalyticsPoint `json:"series"`
	Summary   DistributionAnalyticsSummary `json:"summary"`
	Forecast  DistributionForecast         `json:"forecast"`
}

type DistributionPolicyInput struct {
	CommissionFreezeHours int    `json:"commission_freeze_hours"`
	WithdrawalMinMinor    int64  `json:"withdrawal_min_cny_minor"`
	WithdrawalDailyLimit  int    `json:"withdrawal_daily_limit"`
	WithdrawalFeeBPS      int    `json:"withdrawal_fee_bps"`
	FirstRechargeBonusBPS int    `json:"first_recharge_bonus_bps"`
	FirstRechargeBonusCap string `json:"first_recharge_bonus_cap_usd"`
}

type DistributionConversion struct {
	ID             int64     `json:"id"`
	AmountCNYMinor int64     `json:"amount_cny_minor"`
	USDAmount      string    `json:"usd_amount"`
	CNYToUSDRate   string    `json:"cny_to_usd_rate"`
	RateSource     string    `json:"rate_source"`
	USDToCNYRate   string    `json:"usd_to_cny_rate"`
	ConfigVersion  int       `json:"config_version"`
	CreatedAt      time.Time `json:"created_at"`
}

type DistributionTreeNode struct {
	UserID         int64  `json:"user_id"`
	ParentUserID   int64  `json:"parent_user_id"`
	EmailMasked    string `json:"email_masked"`
	Username       string `json:"username"`
	DirectChildren int64  `json:"direct_children"`
	// TeamVolumeMinor 是这名直属邀请成员在商城的实付总额（人民币 minor）。
	// 返佣没有档位，这里只是「他贡献了多少业绩」的展示字段。
	TeamVolumeMinor int64 `json:"team_volume_cny_minor"`
}

// DistributionCommission 是佣金明细的一行。
//
// 返点现在只有两个可能的来源，用 Source 区分（两表的自增 id 会重号，
// 前端必须用 source + id 组合做 key）：
//   - "shop"         商城订单佣金（shop_commission_records，当前唯一在写入的表）
//   - "distribution" 历史充值返佣（distribution_commissions，已停止写入，仅留痕）
type DistributionCommission struct {
	ID              int64     `json:"id"`
	Source          string    `json:"source"`
	SourceOrderID   int64     `json:"source_order_id"`
	SourceUserID    int64     `json:"source_user_id"`
	Depth           int       `json:"depth"`
	Tier            int       `json:"tier"`
	RateBPS         int       `json:"rate_bps"`
	BaseMinor       int64     `json:"base_cny_minor"`
	AmountMinor     int64     `json:"amount_cny_minor"`
	TeamVolumeMinor int64     `json:"team_volume_cny_minor"`
	Status          string    `json:"status"`
	FrozenUntil     time.Time `json:"frozen_until"`
	CreatedAt       time.Time `json:"created_at"`
}

type AdminDistributionCommission struct {
	DistributionCommission
	BeneficiaryUserID int64 `json:"beneficiary_user_id"`
}

// AdminDistributionMember 是后台「推广成员」列表的一行：只有真正做过推广
// （直接邀请过至少 1 人）的用户才会出现。没有档位，只有人数、业绩与钱包余额。
type AdminDistributionMember struct {
	UserID            int64  `json:"user_id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	InviteeCount      int64  `json:"invitee_count"`
	InviteeSpendMinor int64  `json:"team_volume_cny_minor"`
	AvailableMinor    int64  `json:"available_cny_minor"`
	FrozenMinor       int64  `json:"frozen_cny_minor"`
	LifetimeMinor     int64  `json:"lifetime_earned_cny_minor"`
}

type DistributionRechargeEvent struct {
	ID                    int64      `json:"id"`
	SourceOrderID         int64      `json:"source_order_id"`
	UserID                int64      `json:"user_id"`
	BaseMinor             int64      `json:"base_cny_minor"`
	CreditedUSD           string     `json:"credited_usd"`
	FirstRechargeBonusUSD string     `json:"first_recharge_bonus_usd"`
	ConfigVersion         int        `json:"config_version"`
	Status                string     `json:"status"`
	ReversalReason        string     `json:"reversal_reason,omitempty"`
	ReversedAt            *time.Time `json:"reversed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

type DistributionReversal struct {
	ID              int64     `json:"id"`
	RechargeEventID int64     `json:"recharge_event_id"`
	SourceOrderID   int64     `json:"source_order_id"`
	UserID          int64     `json:"user_id"`
	ReversalType    string    `json:"reversal_type"`
	BaseMinor       int64     `json:"base_cny_minor"`
	PrincipalUSD    string    `json:"principal_usd"`
	BonusUSD        string    `json:"bonus_usd"`
	LegacyRebateUSD string    `json:"legacy_rebate_usd"`
	CommissionMinor int64     `json:"commission_cny_minor"`
	Reason          string    `json:"reason"`
	OperatorUserID  int64     `json:"operator_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type DistributionRelationAudit struct {
	AncestorUserID   int64     `json:"ancestor_user_id"`
	DescendantUserID int64     `json:"descendant_user_id"`
	Depth            int       `json:"depth"`
	CreatedAt        time.Time `json:"created_at"`
}

type DistributionConversionAudit struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	AmountCNYMinor int64     `json:"amount_cny_minor"`
	USDAmount      string    `json:"usd_amount"`
	CNYToUSDRate   string    `json:"cny_to_usd_rate,omitempty"`
	RateSource     string    `json:"rate_source,omitempty"`
	USDToCNYRate   string    `json:"usd_to_cny_rate"`
	ConfigVersion  int       `json:"config_version"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type PayoutAccount struct {
	AccountType  string `json:"account_type"`
	AccountMask  string `json:"account_mask"`
	RealNameMask string `json:"real_name_mask"`
}

type AdminPayoutDetails struct {
	WithdrawalID int64  `json:"withdrawal_id"`
	UserID       int64  `json:"user_id"`
	AccountType  string `json:"account_type"`
	Account      string `json:"account"`
	RealName     string `json:"real_name"`
}

type Withdrawal struct {
	ID               int64      `json:"id"`
	AmountMinor      int64      `json:"amount_cny_minor"`
	FeeMinor         int64      `json:"fee_cny_minor"`
	FeeRateBPS       int        `json:"fee_rate_bps"`
	ConfigVersion    int        `json:"config_version"`
	Status           string     `json:"status"`
	RejectReason     string     `json:"reject_reason,omitempty"`
	PaymentReference string     `json:"payment_reference,omitempty"`
	ProofURL         string     `json:"proof_url,omitempty"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty"`
}

func NewDistributionService(db *sql.DB, encryptor SecretEncryptor) *DistributionService {
	return &DistributionService{db: db, encryptor: encryptor}
}

func (s *DistributionService) IsEnabled(ctx context.Context) bool {
	if s == nil || s.db == nil {
		return false
	}
	var enabled bool
	return s.db.QueryRowContext(ctx, `SELECT enabled FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&enabled) == nil && enabled
}

func (s *DistributionService) ProgramConfig(ctx context.Context) (map[string]any, error) {
	var enabled, stack bool
	var freezeHours, dailyLimit, feeBPS, bonusBPS, version int
	var minimum int64
	var bonusCap string
	err := s.db.QueryRowContext(ctx, `SELECT enabled, stack_with_legacy, commission_freeze_hours, withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps, first_recharge_bonus_bps, first_recharge_bonus_cap_usd::text, current_config_version FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&enabled, &stack, &freezeHours, &minimum, &dailyLimit, &feeBPS, &bonusBPS, &bonusCap, &version)
	if err != nil {
		return nil, err
	}
	multiplier, err := s.balanceRechargeMultiplier(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"enabled": enabled, "stack_with_legacy": false, "commission_freeze_hours": freezeHours,
		"withdrawal_min_cny_minor": minimum, "withdrawal_daily_limit": dailyLimit,
		"withdrawal_fee_bps": feeBPS, "first_recharge_bonus_bps": bonusBPS,
		"first_recharge_bonus_cap_usd": bonusCap, "current_config_version": version,
		"balance_recharge_multiplier": multiplier.String(),
		// Kept for clients that still read the legacy field. New conversions use
		// the explicit CNY-to-USD multiplier above.
		"usd_to_cny_rate": decimal.NewFromInt(1).Div(multiplier).String(),
	}, nil
}

// balanceRechargeMultiplier reads the same setting used by payment recharge
// orders: how many USD of platform balance one CNY produces. Keeping this in
// the distribution service prevents the earnings conversion from drifting
// away from the purchase configuration.
func (s *DistributionService) balanceRechargeMultiplier(ctx context.Context) (decimal.Decimal, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, SettingBalanceRechargeMult).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return decimal.NewFromInt(1), nil
	}
	if err != nil {
		return decimal.Zero, err
	}
	multiplier, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || !multiplier.IsPositive() {
		return decimal.Zero, infraerrors.BadRequest("BALANCE_RECHARGE_MULTIPLIER_INVALID", "balance recharge multiplier is invalid")
	}
	return multiplier, nil
}

func (s *DistributionService) UpdateUSDToCNYRate(ctx context.Context, rate decimal.Decimal) error {
	if !rate.IsPositive() || rate.GreaterThan(decimal.NewFromInt(1000)) || rate.Exponent() < -10 {
		return infraerrors.BadRequest("DISTRIBUTION_EXCHANGE_RATE_INVALID", "distribution USD to CNY rate is invalid")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings (key, value, updated_at) VALUES ('distribution_usd_to_cny_rate', $1, NOW()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, rate.String())
	return err
}

func (s *DistributionService) FinancialRuntimeConfig(ctx context.Context) (map[string]any, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = 'credit_bucket_enforce_enabled'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]any{"credit_bucket_enforce_enabled": false}, nil
	}
	return map[string]any{"credit_bucket_enforce_enabled": strings.EqualFold(strings.TrimSpace(raw), "true")}, err
}

func (s *DistributionService) UpdateFinancialRuntimeConfig(ctx context.Context, enforce bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings (key, value, updated_at) VALUES ('credit_bucket_enforce_enabled', $1, NOW()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, strconv.FormatBool(enforce))
	return err
}

func (s *DistributionService) CreatePolicyVersion(ctx context.Context, operatorID int64, input DistributionPolicyInput) (int, error) {
	bonusCap, err := validateDistributionPolicy(input)
	if err != nil {
		return 0, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var programID int64
	var currentVersion int
	if err := tx.QueryRowContext(ctx, `SELECT id, current_config_version FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' FOR UPDATE`).Scan(&programID, &currentVersion); err != nil {
		return 0, err
	}
	version := currentVersion + 1
	// 政策版本只承载提现规则与冻结期：返佣比例归商品，不再有档位配置。
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_policy_versions (program_id, config_version, commission_freeze_hours, withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps, first_recharge_bonus_bps, first_recharge_bonus_cap_usd, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, programID, version, input.CommissionFreezeHours, input.WithdrawalMinMinor, input.WithdrawalDailyLimit, input.WithdrawalFeeBPS, input.FirstRechargeBonusBPS, bonusCap.String(), operatorID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_programs SET commission_freeze_hours = $2, withdrawal_min_cny_minor = $3, withdrawal_daily_limit = $4, withdrawal_fee_bps = $5, first_recharge_bonus_bps = $6, first_recharge_bonus_cap_usd = $7, current_config_version = $8, updated_at = NOW() WHERE id = $1`, programID, input.CommissionFreezeHours, input.WithdrawalMinMinor, input.WithdrawalDailyLimit, input.WithdrawalFeeBPS, input.FirstRechargeBonusBPS, bonusCap.String(), version); err != nil {
		return 0, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "distribution_policy", strconv.Itoa(version), "distribution.policy_version_created", fmt.Sprintf("distribution-policy:%d", version), map[string]any{"operator_user_id": operatorID, "config_version": version}); err != nil {
		return 0, err
	}
	return version, tx.Commit()
}

func (s *DistributionService) UpdateProgramConfig(ctx context.Context, enabled, stack bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if enabled {
		if err := ensureCreditBucketsReady(ctx, tx); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO settings (key, value, updated_at)
VALUES ('credit_bucket_enforce_enabled', 'true', NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`); err != nil {
			return err
		}
	}
	// Legacy rewards are no longer part of the active compute-company policy.
	// Keep the column for rollback compatibility, but never enable stacking.
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_programs SET enabled = $1, stack_with_legacy = FALSE, updated_at = NOW() WHERE tenant_id = 1 AND code = 'compute_company'`, enabled); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO settings (key, value, updated_at) VALUES ('distribution_enabled', $1, NOW()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, strconv.FormatBool(enabled)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DistributionService) ProcessRecharge(ctx context.Context, orderID, userID int64, payAmount, feeRate, creditedUSD decimal.Decimal) (DistributionProcessResult, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return DistributionProcessResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return DistributionProcessResult{}, err
	}
	var programID int64
	var enabled, stack bool
	var configVersion, bonusBPS int
	var bonusCapRaw string
	err = tx.QueryRowContext(ctx, `SELECT id, enabled, stack_with_legacy, current_config_version, first_recharge_bonus_bps, first_recharge_bonus_cap_usd::text FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' FOR SHARE`).Scan(&programID, &enabled, &stack, &configVersion, &bonusBPS, &bonusCapRaw)
	if errors.Is(err, sql.ErrNoRows) || !enabled {
		return DistributionProcessResult{Enabled: false, StackWithLegacy: false}, nil
	}
	if err != nil {
		return DistributionProcessResult{}, err
	}
	// The legacy stack flag is retained in the schema for compatibility, but
	// active settlements are always single-source compute-company settlements.
	result := DistributionProcessResult{Enabled: true, StackWithLegacy: false}
	baseMinor := rechargeBaseMinor(payAmount, feeRate)
	if baseMinor <= 0 {
		return result, nil
	}
	if err := ensureDistributionMemberTx(ctx, tx, programID, userID); err != nil {
		return result, err
	}

	bonus := decimal.Zero
	var inviterID sql.NullInt64
	_ = tx.QueryRowContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id = $1`, userID).Scan(&inviterID)
	var priorEvents int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_recharge_events WHERE program_id = $1 AND user_id = $2 AND status = 'APPLIED'`, programID, userID).Scan(&priorEvents); err != nil {
		return result, err
	}
	if priorEvents == 0 && inviterID.Valid {
		capAmount, _ := decimal.NewFromString(bonusCapRaw)
		bonus = calculateFirstRechargeBonus(creditedUSD, int64(bonusBPS), capAmount)
	}
	var eventID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO distribution_recharge_events (program_id, tenant_id, source_order_id, user_id, base_cny_minor, credited_usd, first_recharge_bonus_usd, config_version)
VALUES ($1, 1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (program_id, source_order_id) DO NOTHING
	RETURNING id`, programID, orderID, userID, baseMinor, creditedUSD.String(), bonus.String(), configVersion).Scan(&eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if bonus.IsPositive() {
		_, _, err = creditledger.Apply(ctx, tx, userID, bonus, creditctx.Metadata{
			EntryType: "first_recharge_bonus", SourceType: "distribution_recharge", SourceID: strconv.FormatInt(eventID, 10),
			IdempotencyKey: fmt.Sprintf("distribution:%d:first-bonus", eventID), Transferable: false,
		}, false)
		if err != nil {
			return result, err
		}
	}

	// 余额充值不再产生推广佣金。
	//
	// 返佣只有一个来源：被购买商品上架时设置的返佣比例，由商城服务
	// （shop_service.issueCommissionTx）在订单支付成功时按商品快照结算。
	// 这里只保留充值事件本身：它承担首充奖励的幂等判断、冲正审计与后台查询，
	// 不再展开邀请关系、不再累计团队业绩、不再写档位、不再发放任何佣金。
	//
	// 历史数据不受影响：distribution_commissions 与 distribution_recharge_events
	// 里已有的记录照旧可查、可冲正。
	if _, err := tx.ExecContext(ctx, `INSERT INTO financial_outbox_events (tenant_id, aggregate_type, aggregate_id, event_type, payload, idempotency_key) VALUES (1, 'distribution_recharge', $1, 'distribution.recharge_processed', jsonb_build_object('order_id', $2::bigint, 'user_id', $3::bigint), $4) ON CONFLICT DO NOTHING`, strconv.FormatInt(eventID, 10), orderID, userID, fmt.Sprintf("distribution-recharge:%d", orderID)); err != nil {
		return result, err
	}
	return result, tx.Commit()
}

func (s *DistributionService) ReverseRecharge(ctx context.Context, eventID, operatorUserID int64, reversalType, reason string) (*DistributionReversal, error) {
	reversalType = strings.ToUpper(strings.TrimSpace(reversalType))
	reason = strings.TrimSpace(reason)
	if eventID <= 0 || operatorUserID <= 0 || reason == "" || (reversalType != "CHARGEBACK" && reversalType != "REFUND" && reversalType != "ADMIN_CORRECTION") {
		return nil, ErrReversalInvalid
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var programID, orderID, userID, baseMinor int64
	var creditedRaw, bonusRaw, status string
	err = tx.QueryRowContext(ctx, `
SELECT e.program_id, e.source_order_id, e.user_id, e.base_cny_minor,
       e.credited_usd::text, e.first_recharge_bonus_usd::text, e.status
FROM distribution_recharge_events e
WHERE e.id = $1
FOR UPDATE OF e`, eventID).Scan(&programID, &orderID, &userID, &baseMinor, &creditedRaw, &bonusRaw, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("DISTRIBUTION_RECHARGE_NOT_FOUND", "distribution recharge event not found")
	}
	if err != nil {
		return nil, err
	}
	if status == "REVERSED" {
		item, loadErr := loadDistributionReversalTx(ctx, tx, programID, eventID)
		if loadErr != nil {
			return nil, loadErr
		}
		return item, tx.Commit()
	}
	if status != "APPLIED" {
		return nil, infraerrors.Conflict("DISTRIBUTION_REVERSAL_STATE_INVALID", "distribution recharge event cannot be reversed")
	}
	lockRows, err := tx.QueryContext(ctx, `
SELECT user_id
FROM (
    SELECT $1::bigint AS user_id
    UNION
    SELECT user_id FROM user_affiliate_ledger
    WHERE source_order_id = $2 AND action = 'accrue' AND reversed_at IS NULL
    UNION
    SELECT beneficiary_user_id FROM distribution_commissions
    WHERE program_id = $3 AND source_order_id = $2
) locked_users
ORDER BY user_id`, userID, orderID, programID)
	if err != nil {
		return nil, err
	}
	lockUserIDs := make([]int64, 0, 2)
	for lockRows.Next() {
		var lockUserID int64
		if err := lockRows.Scan(&lockUserID); err != nil {
			_ = lockRows.Close()
			return nil, err
		}
		lockUserIDs = append(lockUserIDs, lockUserID)
	}
	if err := lockRows.Close(); err != nil {
		return nil, err
	}
	for _, lockUserID := range lockUserIDs {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockUserID); err != nil {
			return nil, err
		}
	}
	creditedUSD, err := decimal.NewFromString(creditedRaw)
	if err != nil || !creditedUSD.IsPositive() {
		return nil, fmt.Errorf("invalid recharge principal for event %d", eventID)
	}
	bonusUSD, err := decimal.NewFromString(bonusRaw)
	if err != nil || bonusUSD.IsNegative() {
		return nil, fmt.Errorf("invalid recharge bonus for event %d", eventID)
	}
	if bonusUSD.IsPositive() {
		if _, _, err := creditledger.Apply(ctx, tx, userID, bonusUSD.Neg(), creditctx.Metadata{
			EntryType: "first_recharge_bonus_reversal", SourceType: "distribution_reversal", SourceID: strconv.FormatInt(eventID, 10),
			IdempotencyKey: fmt.Sprintf("distribution:%d:first-bonus-reversal", eventID),
		}, false); err != nil {
			return nil, err
		}
	}
	if _, _, err := creditledger.Apply(ctx, tx, userID, creditedUSD.Neg(), creditctx.Metadata{
		EntryType: "recharge_principal_reversal", SourceType: "distribution_reversal", SourceID: strconv.FormatInt(eventID, 10),
		IdempotencyKey: fmt.Sprintf("distribution:%d:principal-reversal", eventID), CountRecharge: true, DebitTransferableFirst: true,
	}, false); err != nil {
		return nil, err
	}

	// 余额充值不再产生推广佣金，也不再累计「团队业绩 / 档位」，
	// 所以冲正只需要退回额度 + 首充奖励，以及回收历史（充值返佣时期）可能存在的佣金。
	type commissionReversalItem struct {
		id, beneficiaryUserID, amountMinor int64
		status                             string
	}
	commissionRows, err := tx.QueryContext(ctx, `SELECT id, beneficiary_user_id, amount_cny_minor, status FROM distribution_commissions WHERE program_id = $1 AND source_order_id = $2 ORDER BY beneficiary_user_id, id FOR UPDATE`, programID, orderID)
	if err != nil {
		return nil, err
	}
	commissions := make([]commissionReversalItem, 0, 5)
	for commissionRows.Next() {
		var item commissionReversalItem
		if err := commissionRows.Scan(&item.id, &item.beneficiaryUserID, &item.amountMinor, &item.status); err != nil {
			_ = commissionRows.Close()
			return nil, err
		}
		commissions = append(commissions, item)
	}
	if err := commissionRows.Close(); err != nil {
		return nil, err
	}
	var commissionMinor int64
	for _, commission := range commissions {
		if commission.status == "REVERSED" {
			continue
		}
		if err := reverseDistributionCommissionTx(ctx, tx, programID, operatorUserID, commission.id, commission.beneficiaryUserID, commission.amountMinor, commission.status, reason); err != nil {
			return nil, err
		}
		commissionMinor += commission.amountMinor
	}
	legacyRebateUSD, err := reverseLegacyAffiliateRebateTx(ctx, tx, orderID, operatorUserID, reason)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_recharge_events SET status = 'REVERSED', reversed_at = $2, reversal_reason = $3, reversal_operator_user_id = $4 WHERE id = $1 AND status = 'APPLIED'`, eventID, now, reason, operatorUserID); err != nil {
		return nil, err
	}
	var reversalID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO distribution_reversal_events (
    program_id, tenant_id, recharge_event_id, source_order_id, user_id,
    reversal_type, base_cny_minor, principal_usd, bonus_usd, legacy_rebate_usd,
    commission_cny_minor, reason, operator_user_id
) VALUES ($1, 1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id`, programID, eventID, orderID, userID, reversalType, baseMinor, creditedUSD.String(), bonusUSD.String(), legacyRebateUSD.String(), commissionMinor, reason, operatorUserID).Scan(&reversalID); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "distribution_reversal", strconv.FormatInt(reversalID, 10), "distribution.recharge_reversed", fmt.Sprintf("distribution-reversal:%d", eventID), map[string]any{"order_id": orderID, "user_id": userID, "reversal_type": reversalType}); err != nil {
		return nil, err
	}
	item, err := loadDistributionReversalTx(ctx, tx, programID, eventID)
	if err != nil {
		return nil, err
	}
	return item, tx.Commit()
}

func (s *DistributionService) Dashboard(ctx context.Context, userID int64) (*DistributionDashboard, error) {
	dashboard := &DistributionDashboard{}
	var programID int64
	err := s.db.QueryRowContext(ctx, `SELECT id, enabled, commission_freeze_hours, withdrawal_min_cny_minor, withdrawal_daily_limit FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID, &dashboard.Enabled, &dashboard.CommissionFreezeHours, &dashboard.WithdrawalMinMinor, &dashboard.WithdrawalDailyLimit)
	if err != nil {
		return nil, err
	}
	// 邀请业绩与返佣全部来自商城订单：返佣比例是商品上架时设置的，
	// 所以这里没有任何「档位 / 团队业绩 / 层级构成」的概念，
	// 也不再读 distribution_members（那张表只服务已下线的充值返佣）。
	if err := s.db.QueryRowContext(ctx, `
SELECT (SELECT COUNT(*) FROM user_affiliates WHERE inviter_id = $1),
       COALESCE((
           SELECT SUM(o.payable_cny_minor)
           FROM shop_orders o
           JOIN user_affiliates ua ON ua.user_id = o.user_id AND ua.inviter_id = $1
           WHERE o.tenant_id = 1 AND o.status IN ('paid', 'fulfilled')
       ), 0)`, userID).Scan(&dashboard.InviteeCount, &dashboard.InviteeSpendMinor); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, lifetime_earned_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`, programID, userID).Scan(&dashboard.AvailableMinor, &dashboard.FrozenMinor, &dashboard.WithdrawingMinor, &dashboard.DebtMinor, &dashboard.LifetimeMinor); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	multiplier, err := s.balanceRechargeMultiplier(ctx)
	if err != nil {
		return nil, err
	}
	dashboard.BalanceRechargeMultiplier = multiplier.String()
	dashboard.USDToCNYRate = decimal.NewFromInt(1).Div(multiplier).String()
	return dashboard, nil
}

func (s *DistributionService) Analytics(ctx context.Context, userID int64, days int) (*DistributionAnalytics, error) {
	if days != 7 && days != 30 && days != 90 {
		days = 30
	}
	end := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.Add(-time.Duration(days*2) * 24 * time.Hour)
	// 趋势同样只看商城：spend = 你直接邀请的人在商城的实付额（payable，抵扣部分不计），
	// commission = 这些订单按商品返佣比例结给你的佣金。不涉任何档位与充值返佣。
	rows, err := s.db.QueryContext(ctx, `
WITH day_grid AS (
    SELECT generate_series($2::timestamptz, $3::timestamptz - INTERVAL '1 day', INTERVAL '1 day')::date AS day
), spend_totals AS (
    SELECT (o.paid_at AT TIME ZONE 'UTC')::date AS day,
           COALESCE(SUM(o.payable_cny_minor), 0)::bigint AS spend_cny_minor
    FROM shop_orders o
    JOIN user_affiliates ua ON ua.user_id = o.user_id AND ua.inviter_id = $1
    WHERE o.tenant_id = 1
      AND o.status IN ('paid', 'fulfilled')
      AND o.paid_at >= $2
      AND o.paid_at < $3
    GROUP BY 1
), commission_totals AS (
    SELECT (c.created_at AT TIME ZONE 'UTC')::date AS day,
           COALESCE(SUM(CASE WHEN c.status <> 'REVERSED' THEN c.amount_cny_minor ELSE 0 END), 0)::bigint AS commission_cny_minor
    FROM shop_commission_records c
    WHERE c.tenant_id = 1
      AND c.beneficiary_user_id = $1
      AND c.created_at >= $2
      AND c.created_at < $3
    GROUP BY 1
)
SELECT TO_CHAR(g.day, 'YYYY-MM-DD'),
       COALESCE(s.spend_cny_minor, 0),
       COALESCE(c.commission_cny_minor, 0)
FROM day_grid g
LEFT JOIN spend_totals s ON s.day = g.day
LEFT JOIN commission_totals c ON c.day = g.day
ORDER BY g.day`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	allSeries := make([]DistributionAnalyticsPoint, 0, days*2)
	for rows.Next() {
		var point DistributionAnalyticsPoint
		if err := rows.Scan(&point.Date, &point.SpendMinor, &point.CommissionMinor); err != nil {
			return nil, err
		}
		allSeries = append(allSeries, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(allSeries) < days {
		return nil, fmt.Errorf("distribution analytics returned incomplete date series")
	}
	current := allSeries[len(allSeries)-days:]
	previous := allSeries[:len(allSeries)-days]
	summary := DistributionAnalyticsSummary{}
	for _, point := range current {
		summary.SpendMinor += point.SpendMinor
		summary.CommissionMinor += point.CommissionMinor
	}
	for _, point := range previous {
		summary.PreviousSpendMinor += point.SpendMinor
		summary.PreviousCommissionMinor += point.CommissionMinor
	}
	summary.SpendGrowthPercent = growthPercent(summary.SpendMinor, summary.PreviousSpendMinor)
	summary.CommissionGrowthPercent = growthPercent(summary.CommissionMinor, summary.PreviousCommissionMinor)
	return &DistributionAnalytics{
		AsOf:      time.Now().UTC(),
		RangeDays: days,
		Series:    append([]DistributionAnalyticsPoint(nil), current...),
		Summary:   summary,
		Forecast: DistributionForecast{
			Method:     "weighted_recent_trend_v1",
			SevenDays:  forecastHorizon(current, 7, 7),
			ThirtyDays: forecastHorizon(current, 30, 14),
		},
	}, nil
}

func growthPercent(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (float64(current-previous) / float64(previous)) * 100
}

func forecastHorizon(series []DistributionAnalyticsPoint, horizon, minimumActiveDays int) DistributionForecastHorizon {
	activeDays := 0
	for _, point := range series {
		if point.SpendMinor > 0 || point.CommissionMinor > 0 {
			activeDays++
		}
	}
	if len(series) < horizon {
		return DistributionForecastHorizon{Reason: "insufficient_history"}
	}
	if activeDays < minimumActiveDays {
		return DistributionForecastHorizon{Reason: "insufficient_activity"}
	}
	recentWindow := 7
	if horizon > 7 {
		recentWindow = 14
	}
	if recentWindow > len(series) {
		recentWindow = len(series)
	}
	recharge, rechargeGrowth := projectMetric(series, recentWindow, horizon, func(point DistributionAnalyticsPoint) int64 { return point.SpendMinor })
	commission, commissionGrowth := projectMetric(series, recentWindow, horizon, func(point DistributionAnalyticsPoint) int64 { return point.CommissionMinor })
	return DistributionForecastHorizon{
		Eligible:                 true,
		EstimatedSpendMinor:      recharge,
		EstimatedCommissionMinor: commission,
		SpendGrowthPercent:       rechargeGrowth,
		CommissionGrowthPercent:  commissionGrowth,
	}
}

func projectMetric(series []DistributionAnalyticsPoint, recentWindow, horizon int, value func(DistributionAnalyticsPoint) int64) (int64, float64) {
	start := len(series) - recentWindow
	var recentTotal, previousTotal int64
	for index := start; index < len(series); index++ {
		recentTotal += value(series[index])
	}
	previousStart := start - recentWindow
	if previousStart < 0 {
		previousStart = 0
	}
	for index := previousStart; index < start; index++ {
		previousTotal += value(series[index])
	}
	recentAverage := float64(recentTotal) / float64(recentWindow)
	previousCount := start - previousStart
	previousAverage := recentAverage
	if previousCount > 0 {
		previousAverage = float64(previousTotal) / float64(previousCount)
	}
	slope := (recentAverage - previousAverage) / float64(recentWindow)
	var projected float64
	for step := 1; step <= horizon; step++ {
		projected += math.Max(0, recentAverage+slope*float64(step))
	}
	baseline := recentAverage * float64(horizon)
	growth := 0.0
	if baseline > 0 {
		growth = ((projected / baseline) - 1) * 100
	}
	return int64(math.Round(projected)), growth
}

func (s *DistributionService) Tree(ctx context.Context, ownerUserID, parentUserID int64, search string, page, pageSize int) ([]DistributionTreeNode, int64, error) {
	if parentUserID <= 0 {
		parentUserID = ownerUserID
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if parentUserID != ownerUserID {
		// 推广计划已收敛为单层：只展示「我直接邀请的人」，不再支持下钻更深的层级。
		// 下钻会暴露第二代伙伴，与单层叙事不符，也会诱发多层分销的预期。
		return nil, 0, infraerrors.Forbidden("DISTRIBUTION_TREE_FORBIDDEN", "team drilling is disabled for the single-level promotion program")
	}
	pattern := "%" + strings.TrimSpace(search) + "%"
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_affiliates ua JOIN users u ON u.id = ua.user_id WHERE ua.inviter_id = $1 AND ($2 = '%%' OR u.email ILIKE $2 OR u.username ILIKE $2)`, parentUserID, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}
	// 业绩（team_volume）直接取这名成员在商城的实付额；不再读 distribution_members，
	// 因为那张表只承载已下线的「充值返佣 + 档位」语义。
	rows, err := s.db.QueryContext(ctx, `
SELECT u.id, $1::bigint, u.email, u.username,
       (SELECT COUNT(*) FROM user_affiliates c WHERE c.inviter_id = u.id),
       COALESCE((
           SELECT SUM(o.payable_cny_minor)
           FROM shop_orders o
           WHERE o.tenant_id = 1 AND o.user_id = u.id AND o.status IN ('paid', 'fulfilled')
       ), 0)
FROM user_affiliates ua JOIN users u ON u.id = ua.user_id
WHERE ua.inviter_id = $1 AND ($2 = '%%' OR u.email ILIKE $2 OR u.username ILIKE $2)
ORDER BY u.id LIMIT $3 OFFSET $4`, parentUserID, pattern, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]DistributionTreeNode, 0, pageSize)
	for rows.Next() {
		var node DistributionTreeNode
		var email string
		if err := rows.Scan(&node.UserID, &node.ParentUserID, &email, &node.Username, &node.DirectChildren, &node.TeamVolumeMinor); err != nil {
			return nil, 0, err
		}
		node.EmailMasked = maskDistributionEmail(email)
		items = append(items, node)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) Ledger(ctx context.Context, userID int64, page, pageSize int) ([]DistributionCommission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var programID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return nil, 0, err
	}
	var total int64
	// 明细要同时覆盖「商城订单佣金」和「历史充值返佣」两张表：
	// 前者是当前唯一在写入的佣金来源，后者只留痕但历史用户的钱在里面。
	if err := s.db.QueryRowContext(ctx, `
SELECT (SELECT COUNT(*) FROM distribution_commissions WHERE program_id = $1 AND beneficiary_user_id = $2)
     + (SELECT COUNT(*) FROM shop_commission_records WHERE tenant_id = 1 AND beneficiary_user_id = $2)`, programID, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH merged AS (
    SELECT 'distribution'::text AS source, id, source_order_id, source_user_id, depth, tier, rate_bps,
           base_cny_minor, amount_cny_minor, team_volume_cny_minor, status, frozen_until, created_at
    FROM distribution_commissions
    WHERE program_id = $1 AND beneficiary_user_id = $2
    UNION ALL
    -- 商城佣金：商品上架时定死的单层返佣比例。
    -- depth 恒为 1（只结算直接邀请人）、tier 恒为 0（已无档位体系）；
    -- team_volume 列对单层模型没有意义，与 base 取同值仅为保持列形状一致。
    SELECT 'shop'::text AS source, id, shop_order_id, buyer_user_id, 1, 0, commission_bps,
           base_cny_minor, amount_cny_minor, base_cny_minor, status, frozen_until, created_at
    FROM shop_commission_records
    WHERE tenant_id = 1 AND beneficiary_user_id = $2
)
SELECT source, id, source_order_id, source_user_id, depth, tier, rate_bps,
       base_cny_minor, amount_cny_minor, team_volume_cny_minor, status, frozen_until, created_at
FROM merged
ORDER BY created_at DESC, source DESC, id DESC
LIMIT $3 OFFSET $4`, programID, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]DistributionCommission, 0, pageSize)
	for rows.Next() {
		var item DistributionCommission
		if err := rows.Scan(&item.Source, &item.ID, &item.SourceOrderID, &item.SourceUserID, &item.Depth, &item.Tier, &item.RateBPS, &item.BaseMinor, &item.AmountMinor, &item.TeamVolumeMinor, &item.Status, &item.FrozenUntil, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) SavePayoutAccount(ctx context.Context, userID int64, alipayAccount, realName string) (*PayoutAccount, error) {
	alipayAccount = strings.TrimSpace(alipayAccount)
	realName = strings.TrimSpace(realName)
	if alipayAccount == "" || len(alipayAccount) > 160 || realName == "" || len(realName) > 80 {
		return nil, infraerrors.BadRequest("PAYOUT_ACCOUNT_INVALID", "invalid Alipay account or real name")
	}
	if s.encryptor == nil {
		return nil, infraerrors.ServiceUnavailable("PAYOUT_ENCRYPTION_UNAVAILABLE", "payout account encryption is unavailable")
	}
	encryptedAccount, err := s.encryptor.Encrypt(alipayAccount)
	if err != nil {
		return nil, err
	}
	encryptedName, err := s.encryptor.Encrypt(realName)
	if err != nil {
		return nil, err
	}
	last4 := alipayAccount
	if len(last4) > 4 {
		last4 = last4[len(last4)-4:]
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO distribution_payout_accounts (tenant_id, user_id, account_type, account_encrypted, account_last4, real_name_encrypted)
VALUES (1, $1, 'alipay', $2, $3, $4)
ON CONFLICT (tenant_id, user_id, account_type) DO UPDATE
SET account_encrypted = EXCLUDED.account_encrypted, account_last4 = EXCLUDED.account_last4,
    real_name_encrypted = EXCLUDED.real_name_encrypted, verified_at = NULL, updated_at = NOW()`, userID, encryptedAccount, last4, encryptedName)
	if err != nil {
		return nil, err
	}
	return &PayoutAccount{AccountType: "alipay", AccountMask: "****" + last4, RealNameMask: maskRealName(realName)}, nil
}

func (s *DistributionService) GetPayoutAccount(ctx context.Context, userID int64) (*PayoutAccount, error) {
	var last4, encryptedName string
	err := s.db.QueryRowContext(ctx, `SELECT account_last4, real_name_encrypted FROM distribution_payout_accounts WHERE tenant_id = 1 AND user_id = $1 AND account_type = 'alipay'`, userID).Scan(&last4, &encryptedName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayoutAccountRequired
	}
	if err != nil {
		return nil, err
	}
	realName, err := s.encryptor.Decrypt(encryptedName)
	if err != nil {
		return nil, err
	}
	return &PayoutAccount{AccountType: "alipay", AccountMask: "****" + last4, RealNameMask: maskRealName(realName)}, nil
}

func (s *DistributionService) CreateWithdrawal(ctx context.Context, userID, amountMinor int64) (*Withdrawal, error) {
	if err := s.Thaw(ctx, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return nil, err
	}
	var programID int64
	var enabled bool
	var minimum int64
	var dailyLimit, feeBPS, configVersion int
	if err := tx.QueryRowContext(ctx, `SELECT id, enabled, withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps, current_config_version FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID, &enabled, &minimum, &dailyLimit, &feeBPS, &configVersion); err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrDistributionDisabled
	}
	if amountMinor < minimum {
		return nil, ErrWithdrawalAmountInvalid
	}
	var payoutAccountID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM distribution_payout_accounts WHERE tenant_id = 1 AND user_id = $1 AND account_type = 'alipay'`, userID).Scan(&payoutAccountID); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayoutAccountRequired
	} else if err != nil {
		return nil, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_withdrawals WHERE program_id = $1 AND user_id = $2 AND submitted_at >= date_trunc('day', NOW())`, programID, userID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= dailyLimit {
		return nil, ErrWithdrawalLimitExceeded
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return nil, err
	}
	var available, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT available_cny_minor, debt_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 FOR UPDATE`, programID, userID).Scan(&available, &debt); err != nil {
		return nil, err
	}
	if debt > 0 {
		return nil, ErrCommissionDebt
	}
	if available < amountMinor {
		return nil, ErrWithdrawalInsufficient
	}
	feeMinor := calculateWithdrawalFee(amountMinor, feeBPS)
	if feeMinor >= amountMinor {
		return nil, ErrWithdrawalAmountInvalid
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET available_cny_minor = available_cny_minor - $3, withdrawing_cny_minor = withdrawing_cny_minor + $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, amountMinor); err != nil {
		return nil, err
	}
	var withdrawal Withdrawal
	if err := tx.QueryRowContext(ctx, `INSERT INTO distribution_withdrawals (program_id, tenant_id, user_id, payout_account_id, amount_cny_minor, fee_cny_minor, fee_rate_bps, config_version) VALUES ($1, 1, $2, $3, $4, $5, $6, $7) RETURNING id, amount_cny_minor, fee_cny_minor, fee_rate_bps, config_version, status, submitted_at`, programID, userID, payoutAccountID, amountMinor, feeMinor, feeBPS, configVersion).Scan(&withdrawal.ID, &withdrawal.AmountMinor, &withdrawal.FeeMinor, &withdrawal.FeeRateBPS, &withdrawal.ConfigVersion, &withdrawal.Status, &withdrawal.SubmittedAt); err != nil {
		return nil, err
	}
	if err := insertDistributionWalletLedger(ctx, tx, programID, userID, "withdrawal_submitted", -amountMinor, "distribution_withdrawal", strconv.FormatInt(withdrawal.ID, 10), fmt.Sprintf("distribution:%d:withdrawal:%d:submit", programID, withdrawal.ID)); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "distribution_withdrawal", strconv.FormatInt(withdrawal.ID, 10), "distribution.withdrawal_submitted", fmt.Sprintf("distribution-withdrawal:%d:submitted", withdrawal.ID), map[string]any{"user_id": userID, "amount_cny_minor": amountMinor}); err != nil {
		return nil, err
	}
	return &withdrawal, tx.Commit()
}

// ConvertToPlatformBalance exchanges available company CNY earnings into a
// non-transferable USD platform balance. The exchange is one-way and is
// recorded with the rate and an idempotency key for auditability.
func (s *DistributionService) ConvertToPlatformBalance(ctx context.Context, userID, amountMinor int64, idempotencyKey string) (*DistributionConversion, error) {
	if userID <= 0 || amountMinor <= 0 || strings.TrimSpace(idempotencyKey) == "" {
		return nil, infraerrors.BadRequest("DISTRIBUTION_CONVERSION_INVALID", "conversion amount and idempotency key are required")
	}
	if err := s.Thaw(ctx, userID); err != nil {
		return nil, err
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > 160 {
		return nil, infraerrors.BadRequest("DISTRIBUTION_CONVERSION_INVALID", "idempotency key is too long")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return nil, err
	}
	var conversion DistributionConversion
	var programID int64
	var enabled bool
	if err := tx.QueryRowContext(ctx, `SELECT id, enabled FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' FOR SHARE`).Scan(&programID, &enabled); err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrDistributionDisabled
	}
	if err := tx.QueryRowContext(ctx, `
SELECT id, amount_cny_minor, usd_amount::text, COALESCE(cny_to_usd_rate::text, ''), COALESCE(rate_source, ''), usd_to_cny_rate::text, config_version, created_at
FROM distribution_usd_conversions
WHERE program_id = $1 AND user_id = $2 AND idempotency_key = $3`, programID, userID, idempotencyKey).Scan(
		&conversion.ID, &conversion.AmountCNYMinor, &conversion.USDAmount,
		&conversion.CNYToUSDRate, &conversion.RateSource, &conversion.USDToCNYRate, &conversion.ConfigVersion, &conversion.CreatedAt,
	); err == nil {
		if conversion.AmountCNYMinor != amountMinor {
			return nil, infraerrors.Conflict("DISTRIBUTION_CONVERSION_IDEMPOTENCY_CONFLICT", "idempotency key was already used with a different amount")
		}
		return &conversion, tx.Commit()
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	rateRaw, err := settingValueTx(ctx, tx, SettingBalanceRechargeMult)
	if err != nil {
		return nil, err
	}
	multiplier, err := decimal.NewFromString(rateRaw)
	if err != nil || !multiplier.IsPositive() {
		return nil, infraerrors.BadRequest("BALANCE_RECHARGE_MULTIPLIER_INVALID", "balance recharge multiplier is invalid")
	}
	legacyRate := decimal.NewFromInt(1).Div(multiplier).Round(10)
	var configVersion int
	if err := tx.QueryRowContext(ctx, `SELECT current_config_version FROM distribution_programs WHERE id = $1`, programID).Scan(&configVersion); err != nil {
		return nil, err
	}
	usdAmount := decimal.NewFromInt(amountMinor).Div(decimal.NewFromInt(100)).Mul(multiplier).Round(8)
	if !usdAmount.IsPositive() {
		return nil, infraerrors.BadRequest("DISTRIBUTION_CONVERSION_INVALID", "conversion amount is too small")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return nil, err
	}
	var available, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT available_cny_minor, debt_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 FOR UPDATE`, programID, userID).Scan(&available, &debt); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrWithdrawalInsufficient
	} else if err != nil {
		return nil, err
	}
	if debt > 0 {
		return nil, ErrCommissionDebt
	}
	if available < amountMinor {
		return nil, ErrWithdrawalInsufficient
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET available_cny_minor = available_cny_minor - $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, amountMinor); err != nil {
		return nil, err
	}
	if err := tx.QueryRowContext(ctx, `
INSERT INTO distribution_usd_conversions (program_id, tenant_id, user_id, amount_cny_minor, usd_amount, cny_to_usd_rate, rate_source, usd_to_cny_rate, config_version, idempotency_key)
VALUES ($1, 1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at`, programID, userID, amountMinor, usdAmount.String(), multiplier.String(), "balance_recharge_multiplier", legacyRate.String(), configVersion, idempotencyKey).Scan(&conversion.ID, &conversion.CreatedAt); err != nil {
		return nil, err
	}
	_, applied, err := creditledger.Apply(ctx, tx, userID, usdAmount, creditctx.Metadata{
		EntryType: "distribution_commission_conversion", SourceType: "distribution_usd_conversion", SourceID: strconv.FormatInt(conversion.ID, 10),
		IdempotencyKey: fmt.Sprintf("distribution:usd-conversion:%d:%s", userID, idempotencyKey), Transferable: false,
	}, false)
	if err != nil {
		return nil, err
	}
	if !applied {
		return nil, infraerrors.Conflict("DISTRIBUTION_CONVERSION_IDEMPOTENCY_CONFLICT", "conversion ledger idempotency key was already used")
	}
	if err := insertDistributionWalletLedger(ctx, tx, programID, userID, "commission_converted", -amountMinor, "distribution_usd_conversion", strconv.FormatInt(conversion.ID, 10), fmt.Sprintf("distribution:usd-conversion:%d:%s", userID, idempotencyKey)); err != nil {
		return nil, err
	}
	conversion.AmountCNYMinor = amountMinor
	conversion.USDAmount = usdAmount.String()
	conversion.CNYToUSDRate = multiplier.String()
	conversion.RateSource = "balance_recharge_multiplier"
	conversion.USDToCNYRate = legacyRate.String()
	conversion.ConfigVersion = configVersion
	return &conversion, tx.Commit()
}

func settingValueTx(ctx context.Context, tx *sql.Tx, key string) (string, error) {
	var raw string
	if err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&raw); errors.Is(err, sql.ErrNoRows) {
		if key == SettingBalanceRechargeMult {
			return "1", nil
		}
		if key == "distribution_usd_to_cny_rate" {
			return "7.15", nil
		}
		return "", err
	} else if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func (s *DistributionService) ListWithdrawals(ctx context.Context, userID int64, page, pageSize int) ([]Withdrawal, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var programID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_withdrawals WHERE program_id = $1 AND user_id = $2`, programID, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, withdrawalSelectSQL+` WHERE w.program_id = $1 AND w.user_id = $2 ORDER BY w.submitted_at DESC, w.id DESC LIMIT $3 OFFSET $4`, programID, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Withdrawal, 0, pageSize)
	for rows.Next() {
		item, err := scanWithdrawal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) AdminListWithdrawals(ctx context.Context, status string, page, pageSize int) ([]Withdrawal, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	pattern := "%"
	if status != "" {
		pattern = status
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_withdrawals WHERE status LIKE $1`, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, withdrawalSelectSQL+` WHERE w.status LIKE $1 ORDER BY w.submitted_at ASC, w.id ASC LIMIT $2 OFFSET $3`, pattern, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Withdrawal, 0, pageSize)
	for rows.Next() {
		item, err := scanWithdrawal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) AdminListCommissions(ctx context.Context, page, pageSize int) ([]AdminDistributionCommission, int64, error) {
	page, pageSize = normalizeFinancialPage(page, pageSize)
	var total int64
	// 与用户端明细同一口径：商城佣金 + 历史充值返佣都要能看到，
	// 否则后台在「充值返佣已下线」之后会是一条空表。
	if err := s.db.QueryRowContext(ctx, `
SELECT (SELECT COUNT(*) FROM distribution_commissions)
     + (SELECT COUNT(*) FROM shop_commission_records)`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH merged AS (
    SELECT 'distribution'::text AS source, id, source_order_id, source_user_id, beneficiary_user_id,
           depth, tier, rate_bps, base_cny_minor, amount_cny_minor, team_volume_cny_minor,
           status, frozen_until, created_at
    FROM distribution_commissions
    UNION ALL
    SELECT 'shop'::text AS source, id, shop_order_id, buyer_user_id, beneficiary_user_id,
           1, 0, commission_bps, base_cny_minor, amount_cny_minor, base_cny_minor,
           status, frozen_until, created_at
    FROM shop_commission_records
    WHERE tenant_id = 1
)
SELECT source, id, source_order_id, source_user_id, beneficiary_user_id, depth, tier, rate_bps,
       base_cny_minor, amount_cny_minor, team_volume_cny_minor, status, frozen_until, created_at
FROM merged
ORDER BY created_at DESC, source DESC, id DESC
LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]AdminDistributionCommission, 0, pageSize)
	for rows.Next() {
		var item AdminDistributionCommission
		if err := rows.Scan(&item.Source, &item.ID, &item.SourceOrderID, &item.SourceUserID, &item.BeneficiaryUserID, &item.Depth, &item.Tier, &item.RateBPS, &item.BaseMinor, &item.AmountMinor, &item.TeamVolumeMinor, &item.Status, &item.FrozenUntil, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) AdminListRechargeEvents(ctx context.Context, page, pageSize int) ([]DistributionRechargeEvent, int64, error) {
	page, pageSize = normalizeFinancialPage(page, pageSize)
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_recharge_events`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
	SELECT id, source_order_id, user_id, base_cny_minor, credited_usd::text,
	       first_recharge_bonus_usd::text, config_version, status,
	       COALESCE(reversal_reason, ''), reversed_at, created_at
FROM distribution_recharge_events
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]DistributionRechargeEvent, 0, pageSize)
	for rows.Next() {
		var item DistributionRechargeEvent
		if err := rows.Scan(&item.ID, &item.SourceOrderID, &item.UserID, &item.BaseMinor, &item.CreditedUSD, &item.FirstRechargeBonusUSD, &item.ConfigVersion, &item.Status, &item.ReversalReason, &item.ReversedAt, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) AdminListRelations(ctx context.Context, page, pageSize int) ([]DistributionRelationAudit, int64, error) {
	page, pageSize = normalizeFinancialPage(page, pageSize)
	var programID, total int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return nil, 0, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_relations WHERE program_id = $1 AND depth BETWEEN 1 AND 5`, programID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT ancestor_user_id, descendant_user_id, depth, created_at
FROM distribution_relations
WHERE program_id = $1 AND depth BETWEEN 1 AND 5
ORDER BY created_at DESC, ancestor_user_id, descendant_user_id
LIMIT $2 OFFSET $3`, programID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]DistributionRelationAudit, 0, pageSize)
	for rows.Next() {
		var item DistributionRelationAudit
		if err := rows.Scan(&item.AncestorUserID, &item.DescendantUserID, &item.Depth, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *DistributionService) AdminListConversions(ctx context.Context, page, pageSize int) ([]DistributionConversionAudit, int64, error) {
	page, pageSize = normalizeFinancialPage(page, pageSize)
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_usd_conversions`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_id, amount_cny_minor, usd_amount::text,
       COALESCE(cny_to_usd_rate::text, ''), COALESCE(rate_source, ''), usd_to_cny_rate::text,
       config_version, idempotency_key, created_at
FROM distribution_usd_conversions
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]DistributionConversionAudit, 0, pageSize)
	for rows.Next() {
		var item DistributionConversionAudit
		if err := rows.Scan(&item.ID, &item.UserID, &item.AmountCNYMinor, &item.USDAmount, &item.CNYToUSDRate, &item.RateSource, &item.USDToCNYRate, &item.ConfigVersion, &item.IdempotencyKey, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// AdminListMembers 列出「正在做推广的人」（至少直接邀请过 1 人）及其业绩与收益。
//
// 推广计划没有档位、也没有人工调档，所以这里不再有任何 tier 字段；
// 业绩与佣金分别来自商城订单（按商品返佣比例结算）与佣金钱包。
// 业绩口径同样是订单实付额（payable），抵扣部分不计。
func (s *DistributionService) AdminListMembers(ctx context.Context, search string, page, pageSize int) ([]AdminDistributionMember, int64, error) {
	page, pageSize = normalizeFinancialPage(page, pageSize)
	search = strings.TrimSpace(search)
	var total int64
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM users u
WHERE u.deleted_at IS NULL AND u.status = 'active'
  AND EXISTS (SELECT 1 FROM user_affiliates ua WHERE ua.inviter_id = u.id)
  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%' OR COALESCE(u.username, '') ILIKE '%' || $1 || '%')`, search).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT u.id, u.email, COALESCE(u.username, ''),
       (SELECT COUNT(*) FROM user_affiliates ua WHERE ua.inviter_id = u.id),
       COALESCE((
           SELECT SUM(o.payable_cny_minor)
           FROM shop_orders o
           JOIN user_affiliates ua ON ua.user_id = o.user_id AND ua.inviter_id = u.id
           WHERE o.tenant_id = 1 AND o.status IN ('paid', 'fulfilled')
       ), 0),
       COALESCE(cw.available_cny_minor, 0),
       COALESCE(cw.frozen_cny_minor, 0),
       COALESCE(cw.lifetime_earned_cny_minor, 0)
FROM users u
LEFT JOIN distribution_cash_wallets cw ON cw.tenant_id = 1 AND cw.user_id = u.id
WHERE u.deleted_at IS NULL AND u.status = 'active'
  AND EXISTS (SELECT 1 FROM user_affiliates ua WHERE ua.inviter_id = u.id)
  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%' OR COALESCE(u.username, '') ILIKE '%' || $1 || '%')
ORDER BY u.id DESC
LIMIT $2 OFFSET $3`, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]AdminDistributionMember, 0, pageSize)
	for rows.Next() {
		var item AdminDistributionMember
		if err := rows.Scan(&item.UserID, &item.Email, &item.Username, &item.InviteeCount, &item.InviteeSpendMinor, &item.AvailableMinor, &item.FrozenMinor, &item.LifetimeMinor); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func normalizeFinancialPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func (s *DistributionService) AdminPayoutDetails(ctx context.Context, withdrawalID int64) (*AdminPayoutDetails, error) {
	var details AdminPayoutDetails
	var encryptedAccount, encryptedName string
	err := s.db.QueryRowContext(ctx, `SELECT w.id, w.user_id, p.account_type, p.account_encrypted, p.real_name_encrypted FROM distribution_withdrawals w JOIN distribution_payout_accounts p ON p.id = w.payout_account_id WHERE w.id = $1`, withdrawalID).Scan(&details.WithdrawalID, &details.UserID, &details.AccountType, &encryptedAccount, &encryptedName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	}
	if err != nil {
		return nil, err
	}
	details.Account, err = s.encryptor.Decrypt(encryptedAccount)
	if err != nil {
		return nil, err
	}
	details.RealName, err = s.encryptor.Decrypt(encryptedName)
	if err != nil {
		return nil, err
	}
	return &details, nil
}

func (s *DistributionService) AdminPartnerPayoutDetails(ctx context.Context, withdrawalID int64) (*AdminPayoutDetails, error) {
	var details AdminPayoutDetails
	var encryptedAccount, encryptedName string
	err := s.db.QueryRowContext(ctx, `SELECT w.id, w.user_id, p.account_type, p.account_encrypted, p.real_name_encrypted FROM saas_partner_withdrawals w JOIN distribution_payout_accounts p ON p.id = w.payout_account_id WHERE w.id = $1`, withdrawalID).Scan(&details.WithdrawalID, &details.UserID, &details.AccountType, &encryptedAccount, &encryptedName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	}
	if err != nil {
		return nil, err
	}
	details.Account, err = s.encryptor.Decrypt(encryptedAccount)
	if err != nil {
		return nil, err
	}
	details.RealName, err = s.encryptor.Decrypt(encryptedName)
	if err != nil {
		return nil, err
	}
	return &details, nil
}

func (s *DistributionService) AdminTransitionWithdrawal(ctx context.Context, withdrawalID, operatorUserID int64, target, reason, reference, proofURL string) (*Withdrawal, error) {
	target = strings.ToUpper(strings.TrimSpace(target))
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	var programID, userID, amountMinor int64
	if err := tx.QueryRowContext(ctx, `SELECT program_id, user_id, amount_cny_minor, status FROM distribution_withdrawals WHERE id = $1 FOR UPDATE`, withdrawalID).Scan(&programID, &userID, &amountMinor, &current); errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	} else if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	switch {
	case target == "APPROVED" && current == "SUBMITTED":
		_, err = tx.ExecContext(ctx, `UPDATE distribution_withdrawals SET status = 'APPROVED', operator_user_id = $2, approved_at = $3, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorUserID, now)
	case target == "PAID" && current == "APPROVED":
		if strings.TrimSpace(reference) == "" {
			return nil, infraerrors.BadRequest("PAYMENT_REFERENCE_REQUIRED", "payment reference is required")
		}
		// Reversals acquire the same beneficiary lock before creating debt. This
		// prevents a chargeback from racing a payout after the debt is observed.
		if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
			return nil, err
		}
		var debt int64
		if err = tx.QueryRowContext(ctx, `SELECT debt_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 FOR UPDATE`, programID, userID).Scan(&debt); errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWithdrawalInsufficient
		} else if err != nil {
			return nil, err
		}
		if debt > 0 {
			return nil, ErrCommissionDebt
		}
		_, err = tx.ExecContext(ctx, `UPDATE distribution_withdrawals SET status = 'PAID', operator_user_id = $2, payment_reference = $3, proof_url = NULLIF($4, ''), paid_at = $5, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorUserID, reference, proofURL, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET withdrawing_cny_minor = withdrawing_cny_minor - $3, lifetime_withdrawn_cny_minor = lifetime_withdrawn_cny_minor + $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, amountMinor)
		}
	case target == "REJECTED" && (current == "SUBMITTED" || current == "APPROVED"):
		if strings.TrimSpace(reason) == "" {
			return nil, infraerrors.BadRequest("REJECT_REASON_REQUIRED", "reject reason is required")
		}
		_, err = tx.ExecContext(ctx, `UPDATE distribution_withdrawals SET status = 'REJECTED', operator_user_id = $2, reject_reason = $3, rejected_at = $4, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorUserID, reason, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET withdrawing_cny_minor = withdrawing_cny_minor - $3, debt_cny_minor = GREATEST(debt_cny_minor - $3, 0), available_cny_minor = available_cny_minor + GREATEST($3 - debt_cny_minor, 0), updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, amountMinor)
		}
	default:
		return nil, ErrWithdrawalStateInvalid
	}
	if err != nil {
		return nil, err
	}
	if target == "PAID" || target == "REJECTED" {
		action, ledgerAmount := "withdrawal_paid", amountMinor
		if target == "REJECTED" {
			action, ledgerAmount = "withdrawal_rejected", amountMinor
		}
		if err := insertDistributionWalletLedger(ctx, tx, programID, userID, action, ledgerAmount, "distribution_withdrawal", strconv.FormatInt(withdrawalID, 10), fmt.Sprintf("distribution:%d:withdrawal:%d:%s", programID, withdrawalID, strings.ToLower(target))); err != nil {
			return nil, err
		}
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "distribution_withdrawal", strconv.FormatInt(withdrawalID, 10), "distribution.withdrawal_"+strings.ToLower(target), fmt.Sprintf("distribution-withdrawal:%d:%s", withdrawalID, strings.ToLower(target)), map[string]any{"operator_user_id": operatorUserID}); err != nil {
		return nil, err
	}
	item, err := scanWithdrawal(tx.QueryRowContext(ctx, withdrawalSelectSQL+` WHERE w.id = $1`, withdrawalID))
	if err != nil {
		return nil, err
	}
	return &item, tx.Commit()
}

const withdrawalSelectSQL = `SELECT w.id, w.amount_cny_minor, w.fee_cny_minor, w.fee_rate_bps, w.config_version, w.status, COALESCE(w.reject_reason, ''), COALESCE(w.payment_reference, ''), COALESCE(w.proof_url, ''), w.submitted_at, w.approved_at, w.paid_at, w.rejected_at FROM distribution_withdrawals w`

func scanWithdrawal(scanner rowScanner) (Withdrawal, error) {
	var item Withdrawal
	var approved, paid, rejected sql.NullTime
	err := scanner.Scan(&item.ID, &item.AmountMinor, &item.FeeMinor, &item.FeeRateBPS, &item.ConfigVersion, &item.Status, &item.RejectReason, &item.PaymentReference, &item.ProofURL, &item.SubmittedAt, &approved, &paid, &rejected)
	if approved.Valid {
		item.ApprovedAt = &approved.Time
	}
	if paid.Valid {
		item.PaidAt = &paid.Time
	}
	if rejected.Valid {
		item.RejectedAt = &rejected.Time
	}
	return item, err
}

func maskRealName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "***"
	}
	return string(runes[0]) + "**"
}

func (s *DistributionService) Thaw(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var programID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return err
	}
	var amount int64
	if err := tx.QueryRowContext(ctx, `
WITH distribution_thawed AS (
    UPDATE distribution_commissions
    SET status = 'AVAILABLE', thawed_at = NOW()
    WHERE program_id = $1 AND beneficiary_user_id = $2 AND status = 'FROZEN' AND frozen_until <= NOW()
    RETURNING amount_cny_minor
), shop_thawed AS (
    UPDATE shop_commission_records
    SET status = 'AVAILABLE', updated_at = NOW()
    WHERE tenant_id = 1 AND beneficiary_user_id = $2 AND status = 'FROZEN' AND frozen_until <= NOW()
    RETURNING amount_cny_minor
)
SELECT COALESCE((SELECT SUM(amount_cny_minor) FROM distribution_thawed), 0)
     + COALESCE((SELECT SUM(amount_cny_minor) FROM shop_thawed), 0)`, programID, userID).Scan(&amount); err != nil {
		return err
	}
	if amount > 0 {
		result, err := tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET frozen_cny_minor = frozen_cny_minor - $3, debt_cny_minor = GREATEST(debt_cny_minor - $3, 0), available_cny_minor = available_cny_minor + GREATEST($3 - debt_cny_minor, 0), updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, amount)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err == nil && affected != 1 {
			return ErrWithdrawalInsufficient
		}
		batchID := fmt.Sprintf("%d-%d-%d", programID, userID, time.Now().UTC().UnixNano())
		if err := insertDistributionWalletLedger(ctx, tx, programID, userID, "commission_thaw", amount, "commission_batch", batchID, "distribution:thaw:"+batchID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *DistributionService) ThawDue(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT beneficiary_user_id
FROM (
    SELECT DISTINCT beneficiary_user_id
    FROM distribution_commissions
    WHERE program_id = (SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company')
      AND status = 'FROZEN' AND frozen_until <= NOW()
    UNION
    SELECT DISTINCT beneficiary_user_id
    FROM shop_commission_records
    WHERE tenant_id = 1 AND status = 'FROZEN' AND frozen_until <= NOW()
) due_users
ORDER BY beneficiary_user_id
LIMIT $1`, limit)
	if err != nil {
		return err
	}
	userIDs := make([]int64, 0, limit)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := s.Thaw(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

func rechargeBaseMinor(payAmount, feeRate decimal.Decimal) int64 {
	divisor := decimal.NewFromInt(1).Add(feeRate.Div(decimal.NewFromInt(100)))
	if !payAmount.IsPositive() || !divisor.IsPositive() {
		return 0
	}
	return payAmount.Div(divisor).Round(2).Shift(2).Round(0).IntPart()
}

func calculateFirstRechargeBonus(credited decimal.Decimal, bps int64, capAmount decimal.Decimal) decimal.Decimal {
	bonus := credited.Mul(decimal.NewFromInt(bps)).Div(decimal.NewFromInt(10000)).Round(8)
	return decimal.Min(bonus, capAmount)
}

func calculateWithdrawalFee(amountMinor int64, feeBPS int) int64 {
	if amountMinor <= 0 || feeBPS <= 0 {
		return 0
	}
	return decimal.NewFromInt(amountMinor).Mul(decimal.NewFromInt(int64(feeBPS))).Div(decimal.NewFromInt(10000)).Round(0).IntPart()
}

func calculateCommissionMinor(baseMinor int64, rateBPS int64) int64 {
	if baseMinor <= 0 || rateBPS <= 0 {
		return 0
	}
	return decimal.NewFromInt(baseMinor).Mul(decimal.NewFromInt(rateBPS)).Div(decimal.NewFromInt(10000)).Round(0).IntPart()
}

// validateDistributionPolicy 校验推广计划的「钱包与提现」参数。
//
// 推广计划已没有档位：返佣比例不在这里配置，而是每个商品上架时各自设定。
// 所以这个校验里不再有任何 tiers 相关断言。
func validateDistributionPolicy(input DistributionPolicyInput) (decimal.Decimal, error) {
	bonusCap, err := decimal.NewFromString(strings.TrimSpace(input.FirstRechargeBonusCap))
	if err != nil || bonusCap.IsNegative() || bonusCap.Exponent() < -8 || input.CommissionFreezeHours < 0 || input.WithdrawalMinMinor <= 0 || input.WithdrawalDailyLimit <= 0 || input.WithdrawalFeeBPS < 0 || input.WithdrawalFeeBPS >= 10000 || input.FirstRechargeBonusBPS < 0 || input.FirstRechargeBonusBPS > 10000 {
		return decimal.Zero, infraerrors.BadRequest("DISTRIBUTION_POLICY_INVALID", "invalid distribution policy")
	}
	return bonusCap, nil
}

func ensureDistributionMemberTx(ctx context.Context, tx *sql.Tx, programID, userID int64) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_members (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_relations (program_id, tenant_id, ancestor_user_id, descendant_user_id, depth) VALUES ($1, 1, $2, $2, 0) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID)
	return err
}

func reverseDistributionCommissionTx(ctx context.Context, tx *sql.Tx, programID, operatorUserID, commissionID, userID, amount int64, status, reason string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return err
	}
	var available, frozen, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT available_cny_minor, frozen_cny_minor, debt_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 FOR UPDATE`, programID, userID).Scan(&available, &frozen, &debt); err != nil {
		return err
	}
	remaining := amount
	if status == "FROZEN" {
		fromFrozen := minInt64(frozen, remaining)
		frozen -= fromFrozen
		remaining -= fromFrozen
	}
	fromAvailable := minInt64(available, remaining)
	available -= fromAvailable
	remaining -= fromAvailable
	debt += remaining
	if _, err := tx.ExecContext(ctx, `
UPDATE distribution_cash_wallets
SET available_cny_minor = $3, frozen_cny_minor = $4, debt_cny_minor = $5,
    lifetime_earned_cny_minor = GREATEST(lifetime_earned_cny_minor - $6, 0),
    lifetime_reversed_cny_minor = lifetime_reversed_cny_minor + $6,
    updated_at = NOW()
WHERE program_id = $1 AND user_id = $2`, programID, userID, available, frozen, debt, amount); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_commissions SET status = 'REVERSED', reversed_at = NOW(), reversal_reason = $2, reversal_operator_user_id = $3 WHERE id = $1 AND status <> 'REVERSED'`, commissionID, reason, operatorUserID); err != nil {
		return err
	}
	return insertDistributionWalletLedger(ctx, tx, programID, userID, "commission_reversal", -amount, "distribution_commission", strconv.FormatInt(commissionID, 10), fmt.Sprintf("distribution:%d:commission:%d:reversal", programID, commissionID))
}

func reverseLegacyAffiliateRebateTx(ctx context.Context, tx *sql.Tx, orderID, operatorUserID int64, reason string) (decimal.Decimal, error) {
	type legacyAccrual struct {
		id, userID, sourceUserID int64
		amount                   decimal.Decimal
		frozenUntil              sql.NullTime
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id, user_id, COALESCE(source_user_id, 0), amount::text, frozen_until
FROM user_affiliate_ledger
WHERE source_order_id = $1 AND action = 'accrue' AND reversed_at IS NULL
ORDER BY id
FOR UPDATE`, orderID)
	if err != nil {
		return decimal.Zero, err
	}
	items := make([]legacyAccrual, 0, 1)
	for rows.Next() {
		var item legacyAccrual
		var amountRaw string
		if err := rows.Scan(&item.id, &item.userID, &item.sourceUserID, &amountRaw, &item.frozenUntil); err != nil {
			_ = rows.Close()
			return decimal.Zero, err
		}
		item.amount, err = decimal.NewFromString(amountRaw)
		if err != nil {
			_ = rows.Close()
			return decimal.Zero, err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, item := range items {
		if !item.amount.IsPositive() {
			continue
		}
		var availableRaw, frozenRaw, historyRaw string
		if err := tx.QueryRowContext(ctx, `SELECT aff_quota::text, aff_frozen_quota::text, aff_history_quota::text FROM user_affiliates WHERE user_id = $1 FOR UPDATE`, item.userID).Scan(&availableRaw, &frozenRaw, &historyRaw); err != nil {
			return decimal.Zero, err
		}
		available, err := decimal.NewFromString(availableRaw)
		if err != nil {
			return decimal.Zero, err
		}
		frozen, err := decimal.NewFromString(frozenRaw)
		if err != nil {
			return decimal.Zero, err
		}
		history, err := decimal.NewFromString(historyRaw)
		if err != nil {
			return decimal.Zero, err
		}
		remaining := item.amount
		if item.frozenUntil.Valid {
			fromFrozen := decimal.Min(frozen, remaining)
			frozen = frozen.Sub(fromFrozen)
			remaining = remaining.Sub(fromFrozen)
		}
		fromAvailable := decimal.Min(available, remaining)
		available = available.Sub(fromAvailable)
		remaining = remaining.Sub(fromAvailable)
		if remaining.IsPositive() {
			if _, _, err := creditledger.Apply(ctx, tx, item.userID, remaining.Neg(), creditctx.Metadata{
				EntryType: "legacy_affiliate_reversal", SourceType: "distribution_reversal", SourceID: strconv.FormatInt(orderID, 10),
				IdempotencyKey: fmt.Sprintf("legacy-affiliate:%d:%d:reversal", orderID, item.userID),
			}, false); err != nil {
				return decimal.Zero, err
			}
		}
		history = decimal.Max(history.Sub(item.amount), decimal.Zero)
		if _, err := tx.ExecContext(ctx, `UPDATE user_affiliates SET aff_quota = $2, aff_frozen_quota = $3, aff_history_quota = $4, updated_at = NOW() WHERE user_id = $1`, item.userID, available.String(), frozen.String(), history.String()); err != nil {
			return decimal.Zero, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO user_affiliate_ledger (
    user_id, action, amount, source_user_id, source_order_id,
    aff_quota_after, aff_frozen_quota_after, aff_history_quota_after,
    created_at, updated_at
) VALUES ($1, 'reverse', $2, NULLIF($3, 0), $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT DO NOTHING`, item.userID, item.amount.Neg().String(), item.sourceUserID, orderID, available.String(), frozen.String(), history.String()); err != nil {
			return decimal.Zero, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE user_affiliate_ledger SET reversed_at = NOW(), reversal_reason = $2, reversal_operator_user_id = $3, updated_at = NOW() WHERE id = $1`, item.id, reason, operatorUserID); err != nil {
			return decimal.Zero, err
		}
		total = total.Add(item.amount)
	}
	return total, nil
}

func loadDistributionReversalTx(ctx context.Context, tx *sql.Tx, programID, eventID int64) (*DistributionReversal, error) {
	var item DistributionReversal
	err := tx.QueryRowContext(ctx, `
SELECT id, recharge_event_id, source_order_id, user_id, reversal_type,
       base_cny_minor, principal_usd::text, bonus_usd::text, legacy_rebate_usd::text,
       commission_cny_minor, reason, operator_user_id, created_at
FROM distribution_reversal_events
WHERE program_id = $1 AND recharge_event_id = $2`, programID, eventID).Scan(
		&item.ID, &item.RechargeEventID, &item.SourceOrderID, &item.UserID, &item.ReversalType,
		&item.BaseMinor, &item.PrincipalUSD, &item.BonusUSD, &item.LegacyRebateUSD, &item.CommissionMinor,
		&item.Reason, &item.OperatorUserID, &item.CreatedAt,
	)
	return &item, err
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func insertDistributionWalletLedger(ctx context.Context, tx *sql.Tx, programID, userID int64, action string, amount int64, sourceType, sourceID, idempotencyKey string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger (program_id, tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, debt_after, idempotency_key) SELECT $1, 1, $2, $3, $4, $5, $6, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, $7 FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 ON CONFLICT DO NOTHING`, programID, userID, action, amount, sourceType, sourceID, idempotencyKey)
	return err
}

func maskDistributionEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if len(local) > 2 {
		local = local[:2]
	}
	return local + "***@" + parts[1]
}
