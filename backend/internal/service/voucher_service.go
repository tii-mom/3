package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrVoucherDisabled          = infraerrors.Forbidden("VOUCHER_DISABLED", "balance vouchers are disabled")
	ErrVoucherNotFound          = infraerrors.NotFound("VOUCHER_NOT_FOUND", "balance voucher not found")
	ErrVoucherInvalid           = infraerrors.BadRequest("VOUCHER_INVALID", "balance voucher is invalid")
	ErrVoucherUnavailable       = infraerrors.Conflict("VOUCHER_UNAVAILABLE", "balance voucher is no longer available")
	ErrVoucherSelfRedeem        = infraerrors.BadRequest("VOUCHER_SELF_REDEEM", "voucher issuer cannot redeem this voucher")
	ErrVoucherInsufficient      = infraerrors.BadRequest("VOUCHER_INSUFFICIENT_TRANSFERABLE_CREDIT", "insufficient transferable credit")
	ErrVoucherLimitExceeded     = infraerrors.TooManyRequests("VOUCHER_LIMIT_EXCEEDED", "voucher creation limit exceeded")
	ErrVoucherStepUpRequired    = infraerrors.Forbidden("VOUCHER_STEP_UP_REQUIRED", "two-factor verification is required")
	ErrVoucherAmountOutOfRange  = infraerrors.BadRequest("VOUCHER_AMOUNT_OUT_OF_RANGE", "voucher amount is outside the configured limits")
	ErrCreditBucketsNotEnforced = infraerrors.Conflict("CREDIT_BUCKETS_NOT_ENFORCED", "credit buckets must be enforced before enabling this feature")
)

const voucherTenantID int64 = 1

type VoucherService struct {
	db          *sql.DB
	settings    SettingRepository
	totpService *TotpService
}

type Voucher struct {
	ID             int64      `json:"id"`
	IssuerUserID   int64      `json:"issuer_user_id"`
	RedeemerUserID *int64     `json:"redeemer_user_id,omitempty"`
	CodeLast4      string     `json:"code_last4"`
	FaceValue      string     `json:"face_value"`
	FeeAmount      string     `json:"fee_amount"`
	FeeRateBPS     int        `json:"fee_rate_bps"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	RedeemedAt     *time.Time `json:"redeemed_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	Code           string     `json:"code,omitempty"`
}

type CreateVoucherInput struct {
	Amount   string
	TOTPCode string
}

type voucherConfig struct {
	enabled         bool
	bucketsEnforced bool
	feeBPS          int64
	minimum         decimal.Decimal
	maximum         decimal.Decimal
	dailyMaximum    decimal.Decimal
	dailyCount      int64
	expiryDays      int
	stepUpMinimum   decimal.Decimal
}

func NewVoucherService(db *sql.DB, settings SettingRepository, totpService *TotpService) *VoucherService {
	return &VoucherService{db: db, settings: settings, totpService: totpService}
}

func (s *VoucherService) AdminConfig(ctx context.Context) (map[string]any, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"enabled": config.enabled, "fee_bps": config.feeBPS,
		"minimum_usd": config.minimum.String(), "maximum_usd": config.maximum.String(),
		"daily_maximum_usd": config.dailyMaximum.String(), "daily_count": config.dailyCount,
		"expiry_days": config.expiryDays, "step_up_minimum_usd": config.stepUpMinimum.String(),
	}, nil
}

func (s *VoucherService) UpdateEnabled(ctx context.Context, enabled bool) error {
	if enabled {
		config, err := s.loadConfig(ctx)
		if err != nil {
			return err
		}
		if !config.bucketsEnforced {
			return ErrCreditBucketsNotEnforced
		}
	}
	return s.settings.Set(ctx, "balance_voucher_enabled", strconv.FormatBool(enabled))
}

func (s *VoucherService) Create(ctx context.Context, userID int64, input CreateVoucherInput) (*Voucher, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !config.enabled {
		return nil, ErrVoucherDisabled
	}
	if !config.bucketsEnforced {
		return nil, ErrCreditBucketsNotEnforced
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(input.Amount))
	if err != nil || amount.LessThan(config.minimum) || amount.GreaterThan(config.maximum) || amount.Exponent() < -8 {
		return nil, ErrVoucherAmountOutOfRange
	}
	if amount.GreaterThanOrEqual(config.stepUpMinimum) {
		if strings.TrimSpace(input.TOTPCode) == "" || s.totpService == nil {
			return nil, ErrVoucherStepUpRequired
		}
		if err := s.totpService.VerifyCode(ctx, userID, input.TOTPCode); err != nil {
			return nil, err
		}
	}
	fee := calculateVoucherFee(amount, config.feeBPS)
	total := amount.Add(fee)
	code, hash, last4, err := generateVoucherCode()
	if err != nil {
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
	var count int64
	var daily string
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*), COALESCE(SUM(face_value), 0)::text
FROM balance_vouchers
WHERE tenant_id = $1 AND issuer_user_id = $2 AND created_at >= date_trunc('day', NOW())
`, voucherTenantID, userID).Scan(&count, &daily); err != nil {
		return nil, err
	}
	dailyAmount, err := decimal.NewFromString(daily)
	if err != nil {
		return nil, err
	}
	if count >= config.dailyCount || dailyAmount.Add(amount).GreaterThan(config.dailyMaximum) {
		return nil, ErrVoucherLimitExceeded
	}
	account, err := lockCreditAccount(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if account.transferable.LessThan(total) {
		return nil, ErrVoucherInsufficient
	}
	account.transferable = account.transferable.Sub(total)
	if err := persistVoucherCreditChange(ctx, tx, userID, account, "voucher_reserve", "balance_voucher", "", total.Neg(), decimal.Zero, fmt.Sprintf("voucher-create:%s", hash)); err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().AddDate(0, 0, config.expiryDays)
	var voucherID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO balance_vouchers (
    tenant_id, issuer_user_id, code_hash, code_last4, face_value, fee_amount,
    fee_rate_bps, status, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, 'ISSUED', $8)
RETURNING id`, voucherTenantID, userID, hash, last4, amount.String(), fee.String(), config.feeBPS, expiresAt).Scan(&voucherID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO balance_voucher_ledger (voucher_id, tenant_id, user_id, action, face_value, fee_amount)
VALUES ($1, $2, $3, 'ISSUED', $4, $5)`, voucherID, voucherTenantID, userID, amount.String(), fee.String()); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_credit_ledger SET source_id = $1 WHERE idempotency_key = $2`, strconv.FormatInt(voucherID, 10), fmt.Sprintf("voucher-create:%s", hash)); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "balance_voucher", strconv.FormatInt(voucherID, 10), "voucher.issued", fmt.Sprintf("voucher:%d:issued", voucherID), map[string]any{"issuer_user_id": userID}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Voucher{ID: voucherID, IssuerUserID: userID, CodeLast4: last4, FaceValue: amount.StringFixed(8), FeeAmount: fee.StringFixed(8), FeeRateBPS: int(config.feeBPS), Status: "ISSUED", ExpiresAt: expiresAt, CreatedAt: time.Now().UTC(), Code: code}, nil
}

func (s *VoucherService) List(ctx context.Context, userID int64, page, pageSize int) ([]Voucher, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if err := s.ExpireDue(ctx, 100); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM balance_vouchers WHERE tenant_id = $1 AND issuer_user_id = $2`, voucherTenantID, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, voucherSelectSQL+`
WHERE tenant_id = $1 AND issuer_user_id = $2
ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`, voucherTenantID, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Voucher, 0, pageSize)
	for rows.Next() {
		voucher, scanErr := scanVoucher(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, voucher)
	}
	return items, total, rows.Err()
}

func (s *VoucherService) AdminList(ctx context.Context, status string, page, pageSize int) ([]Voucher, int64, error) {
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
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM balance_vouchers WHERE status LIKE $1`, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, voucherSelectSQL+` WHERE status LIKE $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`, pattern, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Voucher, 0, pageSize)
	for rows.Next() {
		item, err := scanVoucher(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *VoucherService) SetRiskLock(ctx context.Context, voucherID int64, locked bool, reason string) (*Voucher, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	voucher, err := getVoucherForUpdate(ctx, tx, `id = $1`, voucherID)
	if err != nil {
		return nil, err
	}
	target := "RISK_LOCKED"
	if locked {
		if voucher.Status != "ISSUED" {
			return nil, ErrVoucherUnavailable
		}
		if strings.TrimSpace(reason) == "" {
			return nil, infraerrors.BadRequest("VOUCHER_RISK_REASON_REQUIRED", "risk reason is required")
		}
		_, err = tx.ExecContext(ctx, `UPDATE balance_vouchers SET status = 'RISK_LOCKED', risk_locked_at = NOW(), risk_reason = $2, updated_at = NOW() WHERE id = $1`, voucherID, reason)
	} else {
		if voucher.Status != "RISK_LOCKED" {
			return nil, ErrVoucherUnavailable
		}
		target = "ISSUED"
		_, err = tx.ExecContext(ctx, `UPDATE balance_vouchers SET status = 'ISSUED', risk_locked_at = NULL, risk_reason = NULL, updated_at = NOW() WHERE id = $1`, voucherID)
	}
	if err != nil {
		return nil, err
	}
	face, _ := decimal.NewFromString(voucher.FaceValue)
	fee, _ := decimal.NewFromString(voucher.FeeAmount)
	action := target
	if !locked {
		action = "RISK_UNLOCKED"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO balance_voucher_ledger (voucher_id, tenant_id, user_id, action, face_value, fee_amount, metadata) VALUES ($1, $2, $3, $4, $5, $6, jsonb_build_object('reason', $7))`, voucher.ID, voucherTenantID, voucher.IssuerUserID, action, face.String(), fee.String(), reason); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	voucher.Status = target
	return voucher, nil
}

func (s *VoucherService) Cancel(ctx context.Context, userID, voucherID int64) (*Voucher, error) {
	return s.transitionToRefund(ctx, userID, voucherID, "CANCELLED")
}

func (s *VoucherService) Redeem(ctx context.Context, userID int64, code string) (*Voucher, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !config.enabled {
		return nil, ErrVoucherDisabled
	}
	if !config.bucketsEnforced {
		return nil, ErrCreditBucketsNotEnforced
	}
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if !strings.HasPrefix(normalized, "VCH-") {
		return nil, ErrVoucherInvalid
	}
	hashBytes := sha256.Sum256([]byte(normalized))
	hash := hex.EncodeToString(hashBytes[:])
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	voucher, err := getVoucherForUpdate(ctx, tx, `code_hash = $1 AND tenant_id = $2`, hash, voucherTenantID)
	if err != nil {
		return nil, err
	}
	if voucher.IssuerUserID == userID {
		return nil, ErrVoucherSelfRedeem
	}
	if voucher.Status != "ISSUED" {
		return nil, ErrVoucherUnavailable
	}
	if !voucher.ExpiresAt.After(time.Now()) {
		face, _ := decimal.NewFromString(voucher.FaceValue)
		fee, _ := decimal.NewFromString(voucher.FeeAmount)
		refund := face.Add(fee)
		issuerAccount, lockErr := lockCreditAccount(ctx, tx, voucher.IssuerUserID)
		if lockErr != nil {
			return nil, lockErr
		}
		issuerAccount.transferable = issuerAccount.transferable.Add(refund)
		if err := persistVoucherCreditChange(ctx, tx, voucher.IssuerUserID, issuerAccount, "voucher_refund", "balance_voucher", strconv.FormatInt(voucher.ID, 10), refund, decimal.Zero, fmt.Sprintf("voucher:%d:expired", voucher.ID)); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE balance_vouchers SET status = 'EXPIRED', updated_at = NOW() WHERE id = $1`, voucher.ID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO balance_voucher_ledger (voucher_id, tenant_id, user_id, action, face_value, fee_amount) VALUES ($1, $2, $3, 'EXPIRED', $4, $5)`, voucher.ID, voucherTenantID, voucher.IssuerUserID, face.String(), fee.String()); err != nil {
			return nil, err
		}
		if err := insertFinancialOutboxEvent(ctx, tx, "balance_voucher", strconv.FormatInt(voucher.ID, 10), "voucher.expired", fmt.Sprintf("voucher:%d:expired", voucher.ID), map[string]any{"issuer_user_id": voucher.IssuerUserID}); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, ErrVoucherUnavailable
	}
	face, _ := decimal.NewFromString(voucher.FaceValue)
	account, err := lockCreditAccount(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	account.nonTransferable = account.nonTransferable.Add(face)
	if err := persistVoucherCreditChange(ctx, tx, userID, account, "voucher_redeem", "balance_voucher", strconv.FormatInt(voucher.ID, 10), decimal.Zero, face, fmt.Sprintf("voucher:%d:redeem", voucher.ID)); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE balance_vouchers SET status = 'REDEEMED', redeemer_user_id = $1, redeemed_at = $2, updated_at = NOW() WHERE id = $3`, userID, now, voucher.ID); err != nil {
		return nil, err
	}
	fee, _ := decimal.NewFromString(voucher.FeeAmount)
	if _, err := tx.ExecContext(ctx, `INSERT INTO balance_voucher_ledger (voucher_id, tenant_id, user_id, action, face_value, fee_amount) VALUES ($1, $2, $3, 'REDEEMED', $4, $5)`, voucher.ID, voucherTenantID, userID, face.String(), fee.String()); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "balance_voucher", strconv.FormatInt(voucher.ID, 10), "voucher.redeemed", fmt.Sprintf("voucher:%d:redeemed", voucher.ID), map[string]any{"redeemer_user_id": userID}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	voucher.Status, voucher.RedeemerUserID, voucher.RedeemedAt = "REDEEMED", &userID, &now
	return voucher, nil
}

func (s *VoucherService) ExpireDue(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, issuer_user_id FROM balance_vouchers WHERE tenant_id = $1 AND status = 'ISSUED' AND expires_at <= NOW() ORDER BY id LIMIT $2`, voucherTenantID, limit)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	type due struct{ id, userID int64 }
	items := make([]due, 0)
	for rows.Next() {
		var item due
		if err := rows.Scan(&item.id, &item.userID); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := s.transitionToRefund(ctx, item.userID, item.id, "EXPIRED"); err != nil && !errors.Is(err, ErrVoucherUnavailable) {
			return err
		}
	}
	return nil
}

func (s *VoucherService) transitionToRefund(ctx context.Context, userID, voucherID int64, target string) (*Voucher, error) {
	if target != "CANCELLED" && target != "EXPIRED" {
		return nil, ErrVoucherInvalid
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	voucher, err := getVoucherForUpdate(ctx, tx, `id = $1 AND tenant_id = $2 AND issuer_user_id = $3`, voucherID, voucherTenantID, userID)
	if err != nil {
		return nil, err
	}
	if voucher.Status != "ISSUED" {
		return nil, ErrVoucherUnavailable
	}
	face, _ := decimal.NewFromString(voucher.FaceValue)
	fee, _ := decimal.NewFromString(voucher.FeeAmount)
	refund := face.Add(fee)
	account, err := lockCreditAccount(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	account.transferable = account.transferable.Add(refund)
	if err := persistVoucherCreditChange(ctx, tx, userID, account, "voucher_refund", "balance_voucher", strconv.FormatInt(voucher.ID, 10), refund, decimal.Zero, fmt.Sprintf("voucher:%d:%s", voucher.ID, strings.ToLower(target))); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	updateSQL := `UPDATE balance_vouchers SET status = $1, updated_at = NOW() WHERE id = $2`
	updateArgs := []any{target, voucher.ID}
	if target == "CANCELLED" {
		updateSQL = `UPDATE balance_vouchers SET status = $1, cancelled_at = $2, updated_at = NOW() WHERE id = $3`
		updateArgs = []any{target, now, voucher.ID}
	}
	if _, err := tx.ExecContext(ctx, updateSQL, updateArgs...); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO balance_voucher_ledger (voucher_id, tenant_id, user_id, action, face_value, fee_amount) VALUES ($1, $2, $3, $4, $5, $6)`, voucher.ID, voucherTenantID, userID, target, face.String(), fee.String()); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "balance_voucher", strconv.FormatInt(voucher.ID, 10), "voucher."+strings.ToLower(target), fmt.Sprintf("voucher:%d:%s", voucher.ID, strings.ToLower(target)), map[string]any{"issuer_user_id": userID}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	voucher.Status = target
	if target == "CANCELLED" {
		voucher.CancelledAt = &now
	}
	return voucher, nil
}

type voucherCreditAccount struct {
	transferable    decimal.Decimal
	nonTransferable decimal.Decimal
	debt            decimal.Decimal
}

func (a voucherCreditAccount) balance() decimal.Decimal {
	return a.transferable.Add(a.nonTransferable).Sub(a.debt)
}

func lockCreditAccount(ctx context.Context, tx *sql.Tx, userID int64) (voucherCreditAccount, error) {
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_credit_accounts (user_id, tenant_id, transferable_credit, non_transferable_credit, debt) SELECT id, $2, GREATEST(balance, 0), 0, GREATEST(-balance, 0) FROM users WHERE id = $1 AND deleted_at IS NULL ON CONFLICT (user_id) DO NOTHING`, userID, voucherTenantID); err != nil {
		return voucherCreditAccount{}, err
	}
	var transferable, nonTransferable, debt string
	if err := tx.QueryRowContext(ctx, `SELECT transferable_credit::text, non_transferable_credit::text, debt::text FROM user_credit_accounts WHERE user_id = $1 FOR UPDATE`, userID).Scan(&transferable, &nonTransferable, &debt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return voucherCreditAccount{}, ErrUserNotFound
		}
		return voucherCreditAccount{}, err
	}
	account := voucherCreditAccount{}
	account.transferable, _ = decimal.NewFromString(transferable)
	account.nonTransferable, _ = decimal.NewFromString(nonTransferable)
	account.debt, _ = decimal.NewFromString(debt)
	return account, nil
}

func persistVoucherCreditChange(ctx context.Context, tx *sql.Tx, userID int64, account voucherCreditAccount, entryType, sourceType, sourceID string, transferableDelta, nonTransferableDelta decimal.Decimal, idempotencyKey string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE user_credit_accounts SET transferable_credit = $2, non_transferable_credit = $3, debt = $4, updated_at = NOW() WHERE user_id = $1`, userID, account.transferable.String(), account.nonTransferable.String(), account.debt.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, userID, account.balance().String()); err != nil {
		return err
	}
	var nullableSource any
	if sourceID != "" {
		nullableSource = sourceID
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO user_credit_ledger (
    tenant_id, user_id, entry_type, source_type, source_id,
    transferable_delta, non_transferable_delta, debt_delta,
    transferable_after, non_transferable_after, debt_after, balance_after,
    idempotency_key
) VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, $9, $10, $11, $12)`, voucherTenantID, userID, entryType, sourceType, nullableSource, transferableDelta.String(), nonTransferableDelta.String(), account.transferable.String(), account.nonTransferable.String(), account.debt.String(), account.balance().String(), idempotencyKey)
	return err
}

const voucherSelectSQL = `SELECT id, issuer_user_id, redeemer_user_id, code_last4, face_value::text, fee_amount::text, fee_rate_bps, status, expires_at, redeemed_at, cancelled_at, created_at FROM balance_vouchers `

type rowScanner interface{ Scan(dest ...any) error }

func scanVoucher(scanner rowScanner) (Voucher, error) {
	var voucher Voucher
	var redeemer sql.NullInt64
	var redeemedAt, cancelledAt sql.NullTime
	err := scanner.Scan(&voucher.ID, &voucher.IssuerUserID, &redeemer, &voucher.CodeLast4, &voucher.FaceValue, &voucher.FeeAmount, &voucher.FeeRateBPS, &voucher.Status, &voucher.ExpiresAt, &redeemedAt, &cancelledAt, &voucher.CreatedAt)
	if redeemer.Valid {
		voucher.RedeemerUserID = &redeemer.Int64
	}
	if redeemedAt.Valid {
		voucher.RedeemedAt = &redeemedAt.Time
	}
	if cancelledAt.Valid {
		voucher.CancelledAt = &cancelledAt.Time
	}
	return voucher, err
}

func getVoucherForUpdate(ctx context.Context, tx *sql.Tx, where string, args ...any) (*Voucher, error) {
	voucher, err := scanVoucher(tx.QueryRowContext(ctx, voucherSelectSQL+`WHERE `+where+` FOR UPDATE`, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVoucherNotFound
	}
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func generateVoucherCode() (code, hash, last4 string, err error) {
	random := make([]byte, 20)
	if _, err = rand.Read(random); err != nil {
		return "", "", "", err
	}
	raw := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(random)
	parts := make([]string, 0, 8)
	for len(raw) > 0 {
		n := 4
		if len(raw) < n {
			n = len(raw)
		}
		parts = append(parts, raw[:n])
		raw = raw[n:]
	}
	code = "VCH-" + strings.Join(parts, "-")
	digest := sha256.Sum256([]byte(code))
	hash = hex.EncodeToString(digest[:])
	last4 = strings.ReplaceAll(code, "-", "")
	last4 = last4[len(last4)-4:]
	return code, hash, last4, nil
}

func (s *VoucherService) loadConfig(ctx context.Context) (voucherConfig, error) {
	values, err := s.settings.GetMultiple(ctx, []string{
		"balance_voucher_enabled", "balance_voucher_fee_bps", "balance_voucher_min_usd",
		"balance_voucher_max_usd", "balance_voucher_daily_usd", "balance_voucher_daily_count",
		"balance_voucher_expiry_days", "balance_voucher_step_up_usd", "credit_bucket_enforce_enabled",
	})
	if err != nil {
		return voucherConfig{}, err
	}
	config := voucherConfig{
		enabled: values["balance_voucher_enabled"] == "true", bucketsEnforced: values["credit_bucket_enforce_enabled"] == "true", feeBPS: 800,
		minimum: decimal.NewFromInt(10), maximum: decimal.NewFromInt(10000), dailyMaximum: decimal.NewFromInt(30000),
		dailyCount: 10, expiryDays: 30, stepUpMinimum: decimal.NewFromInt(1000),
	}
	config.feeBPS = parseInt64Default(values["balance_voucher_fee_bps"], config.feeBPS)
	config.dailyCount = parseInt64Default(values["balance_voucher_daily_count"], config.dailyCount)
	config.expiryDays = int(parseInt64Default(values["balance_voucher_expiry_days"], int64(config.expiryDays)))
	config.minimum = parseDecimalDefault(values["balance_voucher_min_usd"], config.minimum)
	config.maximum = parseDecimalDefault(values["balance_voucher_max_usd"], config.maximum)
	config.dailyMaximum = parseDecimalDefault(values["balance_voucher_daily_usd"], config.dailyMaximum)
	config.stepUpMinimum = parseDecimalDefault(values["balance_voucher_step_up_usd"], config.stepUpMinimum)
	return config, nil
}

func parseInt64Default(raw string, fallback int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func parseDecimalDefault(raw string, fallback decimal.Decimal) decimal.Decimal {
	value, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || value.IsNegative() {
		return fallback
	}
	return value
}

func calculateVoucherFee(amount decimal.Decimal, feeBPS int64) decimal.Decimal {
	return amount.Mul(decimal.NewFromInt(feeBPS)).Div(decimal.NewFromInt(10000)).Round(8)
}
