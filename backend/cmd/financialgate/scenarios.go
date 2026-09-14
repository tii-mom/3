package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditledger"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const scenarioWorkers = 12

func runFinancialScenarios(ctx context.Context, db *sql.DB) (_ map[string]string, err error) {
	// 只统计「可登录账号」：迁移 206 会种一个 guest@internal.3api.invalid
	// 系统游客账号（password_hash 为 nologin: 前缀，永不可登录），
	// 若把系统账号也算进去，任何跑过迁移的库都会被判定为非空。
	var existingUsers int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE password_hash NOT LIKE 'nologin:%'`).Scan(&existingUsers); err != nil {
		return nil, err
	}
	if existingUsers != 0 {
		return nil, fmt.Errorf("refusing destructive scenarios: expected an empty users table (ignoring nologin system accounts), found %d login-capable rows", existingUsers)
	}
	// 迁移会种下系统账号（如 206 的游客账号），它们没有信用桶；而开启推广计划
	// 会校验「每个用户都有信用桶」。这里先补齐，否则场景永远跑不起来。
	if err := ensureCreditAccountsForAllUsers(ctx, db); err != nil {
		return nil, err
	}

	settings := databaseSettings{db: db}
	distribution := service.NewDistributionService(db, nil)
	defer func() {
		_, _ = db.ExecContext(context.Background(), `UPDATE distribution_programs SET enabled = FALSE, stack_with_legacy = FALSE WHERE tenant_id = 1 AND code = 'compute_company'`)
		_ = settings.Set(context.Background(), "credit_bucket_enforce_enabled", "false")
		_ = settings.Set(context.Background(), "balance_voucher_enabled", "false")
		_ = settings.Set(context.Background(), "distribution_enabled", "false")
		_ = settings.Set(context.Background(), "saas_control_plane_enabled", "false")
	}()

	users, err := seedScenarioUsers(ctx, db)
	if err != nil {
		return nil, err
	}
	// 整条邀请链都要建（含 users[6]）：商城佣金场景的买家是 users[6]，
	// 他必须有邀请人才会结算佣金。少建一个，佣金就静默变成 'none'。
	if err := seedDistributionChain(ctx, db, users); err != nil {
		return nil, err
	}
	if err := distribution.UpdateFinancialRuntimeConfig(ctx, true); err != nil {
		return nil, err
	}
	if err := distribution.UpdateProgramConfig(ctx, true, false); err != nil {
		return nil, err
	}

	orderID, err := seedRechargePrincipal(ctx, db, users[5])
	if err != nil {
		return nil, err
	}
	if err := runConcurrentRecharge(ctx, distribution, orderID, users[5]); err != nil {
		return nil, err
	}
	if err := assertDistributionScenario(ctx, db, orderID, users[:6]); err != nil {
		return nil, err
	}
	// 商城才是现在唯一会「产生佣金」的地方：余额充值不再发佣金。
	shopResult, err := runShopMoneyScenario(ctx, db, users)
	if err != nil {
		return nil, err
	}
	withdrawalResult, err := runWithdrawalScenario(ctx, db, users[0], users[1])
	if err != nil {
		return nil, err
	}

	if err := settings.Set(ctx, "balance_voucher_enabled", "true"); err != nil {
		return nil, err
	}
	voucherResult, err := runVoucherScenario(ctx, db, settings, users[1], users[6])
	if err != nil {
		return nil, err
	}

	if err := settings.Set(ctx, "saas_control_plane_enabled", "true"); err != nil {
		return nil, err
	}
	wholesaleResult, err := runWholesaleScenario(ctx, db, settings, users[0])
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"distribution_concurrency": fmt.Sprintf("%d duplicate deliveries -> 1 recharge event, 0 commissions (recharge no longer pays commission)", scenarioWorkers),
		"distribution_shop_money":  shopResult,
		"distribution_withdrawal":  withdrawalResult,
		"first_recharge_bonus":     "USD 10,000 principal -> USD 1,000 non-transferable bonus",
		"voucher_lifecycle":        voucherResult,
		"wholesale_billing":        wholesaleResult,
	}, nil
}

// ensureCreditAccountsForAllUsers 给所有还没有信用桶的用户补桶。
// 迁移种下的系统账号（206 的游客账号等）默认没有信用桶，
// 而「开启推广计划」会要求全库用户都有桶，不补齐场景就无法推进。
func ensureCreditAccountsForAllUsers(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `
SELECT u.id
FROM users u
LEFT JOIN user_credit_accounts a ON a.user_id = u.id
WHERE a.user_id IS NULL
ORDER BY u.id`)
	if err != nil {
		return fmt.Errorf("find users without credit accounts: %w", err)
	}
	var pending []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return err
		}
		pending = append(pending, userID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, userID := range pending {
		if err := creditledger.EnsureAccount(ctx, db, userID); err != nil {
			return fmt.Errorf("initialize credit account for user %d: %w", userID, err)
		}
	}
	return nil
}

func seedScenarioUsers(ctx context.Context, db *sql.DB) ([]int64, error) {
	users := make([]int64, 7)
	for i := range users {
		err := db.QueryRowContext(ctx, `
INSERT INTO users (email, password_hash, role, status, balance, concurrency, total_recharged)
VALUES ($1, 'financial-gate-only', 'user', 'active', 0, 5, 0)
RETURNING id`, fmt.Sprintf("financial-gate-%d@example.invalid", i+1)).Scan(&users[i])
		if err != nil {
			return nil, fmt.Errorf("insert scenario user %d: %w", i+1, err)
		}
		if err := creditledger.EnsureAccount(ctx, db, users[i]); err != nil {
			return nil, fmt.Errorf("initialize scenario user %d credit account: %w", i+1, err)
		}
	}
	return users, nil
}

func seedDistributionChain(ctx context.Context, db *sql.DB, users []int64) error {
	for i, userID := range users {
		var inviter any
		if i > 0 {
			inviter = users[i-1]
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO user_affiliates (user_id, aff_code, inviter_id) VALUES ($1, $2, $3)`, userID, fmt.Sprintf("FGATE%02d", i+1), inviter); err != nil {
			return fmt.Errorf("insert affiliate relation %d: %w", i+1, err)
		}
	}
	var programID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO distribution_members (program_id, tenant_id, user_id)
SELECT $1, 1, unnest($2::bigint[]) ON CONFLICT DO NOTHING`, programID, pq.Array(users)); err != nil {
		return fmt.Errorf("insert distribution members: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO distribution_relations (program_id, tenant_id, ancestor_user_id, descendant_user_id, depth)
SELECT $1, 1, id, id, 0 FROM unnest($2::bigint[]) AS ids(id) ON CONFLICT DO NOTHING`, programID, pq.Array(users)); err != nil {
		return fmt.Errorf("insert self relations: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
WITH RECURSIVE chain AS (
    SELECT inviter_id AS ancestor_user_id, user_id AS descendant_user_id, 1 AS depth
    FROM user_affiliates WHERE user_id = ANY($2::bigint[]) AND inviter_id IS NOT NULL
    UNION ALL
    SELECT parent.inviter_id, chain.descendant_user_id, chain.depth + 1
    FROM chain JOIN user_affiliates parent ON parent.user_id = chain.ancestor_user_id
    WHERE chain.depth < 5 AND parent.inviter_id IS NOT NULL
)
INSERT INTO distribution_relations (program_id, tenant_id, ancestor_user_id, descendant_user_id, depth)
SELECT $1, 1, ancestor_user_id, descendant_user_id, depth FROM chain
ON CONFLICT DO NOTHING`, programID, pq.Array(users)); err != nil {
		return fmt.Errorf("insert closure relations: %w", err)
	}
	return nil
}

func seedRechargePrincipal(ctx context.Context, db *sql.DB, buyerID int64) (int64, error) {
	if _, _, err := creditledger.Apply(ctx, db, buyerID, decimal.NewFromInt(10000), creditctx.Metadata{
		EntryType: "paid_recharge", SourceType: "financial_gate", SourceID: "principal",
		IdempotencyKey: "financial-gate:recharge-principal", Transferable: true, CountRecharge: true,
	}, false); err != nil {
		return 0, fmt.Errorf("credit recharge principal: %w", err)
	}
	var orderID int64
	err := db.QueryRowContext(ctx, `
INSERT INTO payment_orders (
    user_id, user_email, amount, pay_amount, fee_rate, recharge_code,
    payment_type, payment_trade_no, order_type, status, out_trade_no,
    expires_at, paid_at, completed_at
) VALUES ($1, 'financial-gate@example.invalid', 10000, 1000, 0, 'FGATE',
          'gate', 'financial-gate-trade', 'balance', 'COMPLETED', 'financial-gate-order',
          NOW() + INTERVAL '1 hour', NOW(), NOW())
RETURNING id`, buyerID).Scan(&orderID)
	return orderID, err
}

func runConcurrentRecharge(ctx context.Context, distribution *service.DistributionService, orderID, buyerID int64) error {
	var wg sync.WaitGroup
	errs := make(chan error, scenarioWorkers)
	for range scenarioWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := distribution.ProcessRecharge(ctx, orderID, buyerID, decimal.NewFromInt(1000), decimal.Zero, decimal.NewFromInt(10000))
			if err == nil && !result.Enabled {
				err = errors.New("distribution unexpectedly disabled")
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return fmt.Errorf("concurrent recharge: %w", err)
		}
	}
	return nil
}

func assertDistributionScenario(ctx context.Context, db *sql.DB, orderID int64, users []int64) error {
	var events int64
	var bonus string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(first_recharge_bonus_usd), 0)::text FROM distribution_recharge_events WHERE source_order_id = $1`, orderID).Scan(&events, &bonus); err != nil {
		return err
	}
	if events != 1 || bonus != "1000.00000000" {
		return fmt.Errorf("unexpected recharge event result: events=%d bonus=%s", events, bonus)
	}
	// 余额充值不再产生推广佣金（返佣只跟着商品的 commission_bps 走）。
	// 重复投递 12 次的旧断言是「1 事件 + 5 层佣金」，现在必须是「1 事件 + 0 佣金」。
	var rechargeCommissions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_commissions WHERE source_order_id = $1`, orderID).Scan(&rechargeCommissions); err != nil {
		return err
	}
	if rechargeCommissions != 0 {
		return fmt.Errorf("recharge must not settle commission, found %d rows", rechargeCommissions)
	}
	var shopCommissions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shop_commission_records`).Scan(&shopCommissions); err != nil {
		return err
	}
	if shopCommissions != 0 {
		return fmt.Errorf("recharge fixtures must not touch shop commissions, found %d rows", shopCommissions)
	}
	var transferable, nonTransferable, balance string
	if err := db.QueryRowContext(ctx, `SELECT a.transferable_credit::text, a.non_transferable_credit::text, u.balance::text FROM user_credit_accounts a JOIN users u ON u.id = a.user_id WHERE a.user_id = $1`, users[5]).Scan(&transferable, &nonTransferable, &balance); err != nil {
		return err
	}
	if transferable != "10000.00000000" || nonTransferable != "1000.00000000" || balance != "11000.00000000" {
		return fmt.Errorf("first recharge buckets mismatch: transferable=%s non_transferable=%s balance=%s", transferable, nonTransferable, balance)
	}
	return nil
}

// runShopMoneyScenario 覆盖当前唯一会「产生佣金」与「花掉返点」的地方 —— 商城订单，
// 全部走真实的 service 代码路径（不自己拼佣金 SQL）：
//
//	A. 外部全额支付：恰好 1 笔商城佣金，基数 = 实付额，费率 = 商品 commission_bps；
//	B. 返点余额全额抵扣：不再产生佣金（否则「返点 → 抵扣下单 → 再返点」会无锚增发），
//	   且买家返点余额被真实扣减；
//	C. 余额不足以全额抵扣：抵扣被回滚、订单作废、余额退回，且不留佣金。
//
// 这里不覆盖「部分抵扣 + 外部补付」的支付单链路：那一段需要真实支付渠道，
// 由生产环境的支付回调链路负责；本场景只锁死「钱与佣金」的守恒关系。
func runShopMoneyScenario(ctx context.Context, db *sql.DB, users []int64) (string, error) {
	if len(users) < 7 {
		return "", errors.New("shop money scenario requires seven users")
	}
	shop := service.NewShopService(db)

	var programID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return "", err
	}
	const (
		priceMinor = 10000 // ¥100.00
		commission = 800   // 8%
		wantAmount = 800   // ¥8.00
	)
	seedWallet := func(userID, available int64) error {
		_, err := db.ExecContext(ctx, `
INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id, available_cny_minor)
VALUES ($1, 1, $2, $3)
ON CONFLICT (program_id, user_id) DO UPDATE
SET available_cny_minor = EXCLUDED.available_cny_minor, updated_at = NOW()`, programID, userID, available)
		return err
	}
	walletAvailable := func(userID int64) (int64, error) {
		var available int64
		err := db.QueryRowContext(ctx, `SELECT available_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`, programID, userID).Scan(&available)
		return available, err
	}

	var productID int64
	if err := db.QueryRowContext(ctx, `
INSERT INTO shop_products (tenant_id, name, description, product_type, price_cny_minor, commission_bps, status)
VALUES (1, 'financial-gate-sku', 'gate fixture', 'virtual', $1, $2, 'published')
RETURNING id`, priceMinor, commission).Scan(&productID); err != nil {
		return "", fmt.Errorf("insert shop product: %w", err)
	}

	// A. 外部全额支付：buyer = users[6]，邀请人 = users[5]，佣金归 users[5]。
	buyerA, beneficiary := users[6], users[5]
	var orderA int64
	if err := db.QueryRowContext(ctx, `
INSERT INTO shop_orders (tenant_id, user_id, product_id, snapshot_name, snapshot_description, snapshot_image_url,
    snapshot_product_type, snapshot_price_cny_minor, snapshot_grant_usd_amount, snapshot_commission_bps, payable_cny_minor)
VALUES (1, $1, $2, 'financial-gate-sku', '', '', 'virtual', $3, 0, $4, $3)
RETURNING id`, buyerA, productID, priceMinor, commission).Scan(&orderA); err != nil {
		return "", fmt.Errorf("insert externally paid shop order: %w", err)
	}
	if err := shop.FulfillWalletPaidOrder(ctx, orderA); err != nil {
		return "", fmt.Errorf("fulfil externally paid shop order: %w", err)
	}
	var base, bps, amount int64
	var commissionStatus, beneficiaryID string
	if err := db.QueryRowContext(ctx, `
SELECT base_cny_minor, commission_bps, amount_cny_minor, status, beneficiary_user_id::text
FROM shop_commission_records WHERE shop_order_id = $1`, orderA).Scan(&base, &bps, &amount, &commissionStatus, &beneficiaryID); err != nil {
		return "", fmt.Errorf("read shop commission: %w", err)
	}
	if base != priceMinor || bps != commission || amount != wantAmount || commissionStatus != "FROZEN" {
		return "", fmt.Errorf("shop commission mismatch: base=%d bps=%d amount=%d status=%s", base, bps, amount, commissionStatus)
	}
	var wantBeneficiary string
	if err := db.QueryRowContext(ctx, `SELECT $1::text`, beneficiary).Scan(&wantBeneficiary); err != nil {
		return "", err
	}
	if beneficiaryID != wantBeneficiary {
		return "", fmt.Errorf("shop commission beneficiary=%s want=%s", beneficiaryID, wantBeneficiary)
	}
	var orderAStatus, orderACommission string
	if err := db.QueryRowContext(ctx, `SELECT status, commission_status FROM shop_orders WHERE id = $1`, orderA).Scan(&orderAStatus, &orderACommission); err != nil {
		return "", err
	}
	if orderAStatus != "paid" || orderACommission != "frozen" {
		return "", fmt.Errorf("externally paid order state mismatch: status=%s commission=%s", orderAStatus, orderACommission)
	}

	// B. 返点余额全额抵扣：buyer = users[4]（邀请人 users[3]），余额充足。
	buyerB := users[4]
	if err := seedWallet(buyerB, priceMinor+5000); err != nil {
		return "", err
	}
	resultB, err := shop.CreateOrderAndPayment(ctx, buyerB, productID, "", "", "", "", "", "", false, false, "", true)
	if err != nil {
		return "", fmt.Errorf("full wallet deduction order: %w", err)
	}
	if !resultB.FullyPaidByWallet || resultB.PayableCNYMinor != 0 || resultB.WalletAppliedCNYMinor != priceMinor {
		return "", fmt.Errorf("full deduction result mismatch: fully=%t payable=%d applied=%d", resultB.FullyPaidByWallet, resultB.PayableCNYMinor, resultB.WalletAppliedCNYMinor)
	}
	availableB, err := walletAvailable(buyerB)
	if err != nil {
		return "", err
	}
	if availableB != 5000 {
		return "", fmt.Errorf("full deduction did not debit wallet: available=%d want=5000", availableB)
	}
	var orderBStatus, orderBCommission string
	if err := db.QueryRowContext(ctx, `SELECT status, commission_status FROM shop_orders WHERE id = $1`, resultB.ShopOrderID).Scan(&orderBStatus, &orderBCommission); err != nil {
		return "", err
	}
	// 虚拟商品的订单在「已付款、待人工交付」时停在 paid/pending（与外部支付一致），
	// 关键断言是它没有结算佣金。
	if orderBStatus != "paid" || orderBCommission != "none" {
		return "", fmt.Errorf("fully deducted order state mismatch: status=%s commission=%s", orderBStatus, orderBCommission)
	}

	// C. 余额不足：全额抵扣失败 → 抵扣回滚、订单作废、余额退回，且不留佣金。
	buyerC := users[2]
	if err := seedWallet(buyerC, 3000); err != nil {
		return "", err
	}
	if _, err := shop.CreateOrderAndPayment(ctx, buyerC, productID, "alipay", "", "", "", "", "", false, false, "", true); err == nil {
		return "", errors.New("insufficient balance must not create an externally payable order without a payment provider")
	}
	availableC, err := walletAvailable(buyerC)
	if err != nil {
		return "", err
	}
	if availableC != 3000 {
		return "", fmt.Errorf("aborted order did not refund wallet: available=%d want=3000", availableC)
	}
	var orderC int64
	var orderCStatus string
	if err := db.QueryRowContext(ctx, `SELECT id, status FROM shop_orders WHERE user_id = $1 ORDER BY id DESC LIMIT 1`, buyerC).Scan(&orderC, &orderCStatus); err != nil {
		return "", err
	}
	if orderCStatus != "failed" {
		return "", fmt.Errorf("aborted shop order status=%s want=failed", orderCStatus)
	}
	var refunds int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_wallet_ledger WHERE idempotency_key = $1`, fmt.Sprintf("shop:order:%d:wallet-refund", orderC)).Scan(&refunds); err != nil {
		return "", err
	}
	if refunds != 1 {
		return "", fmt.Errorf("aborted order wallet refunds=%d want=1", refunds)
	}

	// 全程只应存在 A 留下的那一笔佣金。
	var totalCommissions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shop_commission_records`).Scan(&totalCommissions); err != nil {
		return "", err
	}
	if totalCommissions != 1 {
		return "", fmt.Errorf("shop commissions total=%d want=1 (deducted portion must not pay commission)", totalCommissions)
	}

	return "CNY 100 order -> 1 commission of CNY 8.00 on paid base; full wallet deduction pays 0 commission; aborted deduction refunded", nil
}

func runWithdrawalScenario(ctx context.Context, db *sql.DB, paidUserID, rejectedUserID int64) (string, error) {
	distribution := service.NewDistributionService(db, gateEncryptor{})
	var programID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&programID); err != nil {
		return "", err
	}
	seedWallet := func(userID int64) error {
		_, err := db.ExecContext(ctx, `
INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor)
VALUES ($1, 1, $2, 5000, 0, 0, 0)
ON CONFLICT (program_id, user_id) DO UPDATE
SET available_cny_minor = 5000, frozen_cny_minor = 0, withdrawing_cny_minor = 0, debt_cny_minor = 0, updated_at = NOW()`, programID, userID)
		return err
	}
	if err := seedWallet(paidUserID); err != nil {
		return "", err
	}
	if err := seedWallet(rejectedUserID); err != nil {
		return "", err
	}
	if _, err := distribution.SavePayoutAccount(ctx, paidUserID, "paid@example.invalid", "Paid User"); err != nil {
		return "", err
	}
	if _, err := distribution.SavePayoutAccount(ctx, rejectedUserID, "reject@example.invalid", "Rejected User"); err != nil {
		return "", err
	}
	paid, err := distribution.CreateWithdrawal(ctx, paidUserID, 2000)
	if err != nil {
		return "", err
	}
	if _, err := distribution.AdminTransitionWithdrawal(ctx, paid.ID, paidUserID, "APPROVED", "", "paid-ref", ""); err != nil {
		return "", err
	}
	if _, err := distribution.AdminTransitionWithdrawal(ctx, paid.ID, paidUserID, "PAID", "", "paid-ref", "proof"); err != nil {
		return "", err
	}
	if _, err := distribution.CreateWithdrawal(ctx, paidUserID, 2000); !errors.Is(err, service.ErrWithdrawalLimitExceeded) {
		return "", fmt.Errorf("daily withdrawal limit returned %v", err)
	}
	rejected, err := distribution.CreateWithdrawal(ctx, rejectedUserID, 2000)
	if err != nil {
		return "", err
	}
	if _, err := distribution.AdminTransitionWithdrawal(ctx, rejected.ID, rejectedUserID, "APPROVED", "", "", ""); err != nil {
		return "", err
	}
	if _, err := distribution.AdminTransitionWithdrawal(ctx, rejected.ID, rejectedUserID, "REJECTED", "manual review", "", ""); err != nil {
		return "", err
	}
	var paidAvailable, paidWithdrawing, paidLifetime int64
	if err := db.QueryRowContext(ctx, `SELECT available_cny_minor, withdrawing_cny_minor, lifetime_withdrawn_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`, programID, paidUserID).Scan(&paidAvailable, &paidWithdrawing, &paidLifetime); err != nil {
		return "", err
	}
	var rejectedAvailable, rejectedWithdrawing int64
	if err := db.QueryRowContext(ctx, `SELECT available_cny_minor, withdrawing_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`, programID, rejectedUserID).Scan(&rejectedAvailable, &rejectedWithdrawing); err != nil {
		return "", err
	}
	if paidAvailable != 3000 || paidWithdrawing != 0 || paidLifetime != 2000 || rejectedAvailable != 5000 || rejectedWithdrawing != 0 {
		return "", fmt.Errorf("withdrawal wallet mismatch: paid=(%d,%d,%d) rejected=(%d,%d)", paidAvailable, paidWithdrawing, paidLifetime, rejectedAvailable, rejectedWithdrawing)
	}
	return "paid and rejected withdrawals settle wallet balances; daily limit enforced", nil
}

func runVoucherScenario(ctx context.Context, db *sql.DB, settings databaseSettings, issuerID, redeemerID int64) (string, error) {
	if _, _, err := creditledger.Apply(ctx, db, issuerID, decimal.NewFromInt(500), creditctx.Metadata{
		EntryType: "scenario_funding", SourceType: "financial_gate", SourceID: "voucher",
		IdempotencyKey: "financial-gate:voucher-funding", Transferable: true,
	}, false); err != nil {
		return "", err
	}
	vouchers := service.NewVoucherService(db, settings, nil)
	first, err := vouchers.Create(ctx, issuerID, service.CreateVoucherInput{Amount: "100"})
	if err != nil {
		return "", fmt.Errorf("create cancellable voucher: %w", err)
	}
	fee, feeErr := decimal.NewFromString(first.FeeAmount)
	if feeErr != nil || !fee.Equal(decimal.NewFromInt(8)) {
		return "", fmt.Errorf("voucher fee mismatch: %s", first.FeeAmount)
	}
	if _, err := vouchers.Cancel(ctx, issuerID, first.ID); err != nil {
		return "", fmt.Errorf("cancel voucher: %w", err)
	}
	second, err := vouchers.Create(ctx, issuerID, service.CreateVoucherInput{Amount: "100"})
	if err != nil {
		return "", fmt.Errorf("create redeemable voucher: %w", err)
	}
	if _, err := vouchers.Redeem(ctx, issuerID, second.Code); !errors.Is(err, service.ErrVoucherSelfRedeem) {
		return "", fmt.Errorf("self redemption was not rejected: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for range scenarioWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, redeemErr := vouchers.Redeem(ctx, redeemerID, second.Code); redeemErr == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		return "", fmt.Errorf("concurrent voucher redemption successes=%d, want 1", successes)
	}
	var issuerTransferable, redeemerNonTransferable string
	if err := db.QueryRowContext(ctx, `SELECT transferable_credit::text FROM user_credit_accounts WHERE user_id = $1`, issuerID).Scan(&issuerTransferable); err != nil {
		return "", err
	}
	if err := db.QueryRowContext(ctx, `SELECT non_transferable_credit::text FROM user_credit_accounts WHERE user_id = $1`, redeemerID).Scan(&redeemerNonTransferable); err != nil {
		return "", err
	}
	if issuerTransferable != "392.00000000" || redeemerNonTransferable != "100.00000000" {
		return "", fmt.Errorf("voucher buckets mismatch: issuer=%s redeemer=%s", issuerTransferable, redeemerNonTransferable)
	}
	return fmt.Sprintf("USD 100 + USD 8 fee; cancel refunded; %d concurrent redemptions -> 1 success", scenarioWorkers), nil
}

func runWholesaleScenario(ctx context.Context, db *sql.DB, settings databaseSettings, ownerID int64) (string, error) {
	saas := service.NewSaaSService(db, settings, nil, gateEncryptor{})
	created, err := saas.CreateTenant(ctx, service.CreateSaaSTenantInput{Slug: "financial-gate", Name: "Financial Gate", CoreUserID: ownerID})
	if err != nil {
		return "", fmt.Errorf("create SaaS tenant: %w", err)
	}
	if created.WholesaleKey == "" {
		return "", errors.New("wholesale key was not returned at creation")
	}
	if _, err := saas.FundWholesaleWallet(ctx, created.Tenant.ID, "1000", "financial-gate-funding"); err != nil {
		return "", fmt.Errorf("fund wholesale wallet: %w", err)
	}
	var apiKeyID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM api_keys WHERE tenant_id = $1 AND key_type = 'tenant_wholesale'`, created.Tenant.ID).Scan(&apiKeyID); err != nil {
		return "", err
	}
	repo := repository.NewUsageBillingRepository(nil, db)
	var wg sync.WaitGroup
	var mu sync.Mutex
	applied := 0
	errs := make(chan error, scenarioWorkers)
	for range scenarioWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, applyErr := repo.Apply(ctx, &service.UsageBillingCommand{
				RequestID: "financial-gate-request", APIKeyID: apiKeyID,
				WholesaleTenantID: &created.Tenant.ID, WholesaleCost: 1.25,
			})
			if applyErr == nil && result.Applied {
				mu.Lock()
				applied++
				mu.Unlock()
			}
			errs <- applyErr
		}()
	}
	wg.Wait()
	close(errs)
	for applyErr := range errs {
		if applyErr != nil {
			return "", fmt.Errorf("concurrent wholesale billing: %w", applyErr)
		}
	}
	if applied != 1 {
		return "", fmt.Errorf("wholesale billing applied=%d, want 1", applied)
	}
	var balance string
	if err := db.QueryRowContext(ctx, `SELECT balance_usd::text FROM saas_wholesale_wallets WHERE tenant_id = $1`, created.Tenant.ID).Scan(&balance); err != nil {
		return "", err
	}
	if balance != "998.75000000" {
		return "", fmt.Errorf("wholesale balance mismatch: %s", balance)
	}
	overdrawRequestID := "financial-gate-overdraw"
	_, err = repo.Apply(ctx, &service.UsageBillingCommand{
		RequestID: overdrawRequestID, APIKeyID: apiKeyID,
		WholesaleTenantID: &created.Tenant.ID, WholesaleCost: 2000,
	})
	if !errors.Is(err, service.ErrTenantWholesaleBalanceInsufficient) {
		return "", fmt.Errorf("wholesale overdraft returned %v, want %v", err, service.ErrTenantWholesaleBalanceInsufficient)
	}
	var balanceAfterOverdraw string
	if err := db.QueryRowContext(ctx, `SELECT balance_usd::text FROM saas_wholesale_wallets WHERE tenant_id = $1`, created.Tenant.ID).Scan(&balanceAfterOverdraw); err != nil {
		return "", err
	}
	var overdrawClaims int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM usage_billing_dedup WHERE request_id = $1 AND api_key_id = $2`, overdrawRequestID, apiKeyID).Scan(&overdrawClaims); err != nil {
		return "", err
	}
	if balanceAfterOverdraw != balance || overdrawClaims != 0 {
		return "", fmt.Errorf("wholesale overdraft was not atomic: balance=%s claims=%d", balanceAfterOverdraw, overdrawClaims)
	}
	return fmt.Sprintf("%d duplicate requests -> 1 debit; USD 1,000 - USD 1.25 = USD %s; overdraft rejected atomically", scenarioWorkers, balance), nil
}

func runDistributionStress(ctx context.Context, db *sql.DB, orderCount, concurrency int) (_ string, err error) {
	if orderCount <= 0 || orderCount > 100000 {
		return "", fmt.Errorf("stress order count must be between 1 and 100000")
	}
	if concurrency <= 0 || concurrency > 256 {
		return "", fmt.Errorf("stress concurrency must be between 1 and 256")
	}
	var buyerID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = 'financial-gate-6@example.invalid' AND deleted_at IS NULL`).Scan(&buyerID); err != nil {
		return "", fmt.Errorf("financial scenario fixture is required before stress mode: %w", err)
	}
	settings := databaseSettings{db: db}
	distribution := service.NewDistributionService(db, nil)
	defer func() {
		_, _ = db.ExecContext(context.Background(), `UPDATE distribution_programs SET enabled = FALSE, stack_with_legacy = FALSE WHERE tenant_id = 1 AND code = 'compute_company'`)
		_ = settings.Set(context.Background(), "credit_bucket_enforce_enabled", "false")
		_ = settings.Set(context.Background(), "distribution_enabled", "false")
	}()
	if err := distribution.UpdateFinancialRuntimeConfig(ctx, true); err != nil {
		return "", err
	}
	if err := distribution.UpdateProgramConfig(ctx, true, false); err != nil {
		return "", err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	orderIDs := make([]int64, orderCount)
	runID := time.Now().UTC().UnixNano()
	for i := range orderIDs {
		err := tx.QueryRowContext(ctx, `
INSERT INTO payment_orders (
    user_id, user_email, amount, pay_amount, fee_rate, recharge_code,
    payment_type, payment_trade_no, order_type, status, out_trade_no,
    expires_at, paid_at, completed_at
) VALUES ($1, 'financial-gate@example.invalid', 10, 1, 0, 'FGATE-STRESS',
          'gate', $2, 'balance', 'COMPLETED', $3,
          NOW() + INTERVAL '1 hour', NOW(), NOW())
RETURNING id`, buyerID, fmt.Sprintf("stress-trade-%d-%d", runID, i), fmt.Sprintf("stress-order-%d-%d", runID, i)).Scan(&orderIDs[i])
		if err != nil {
			return "", fmt.Errorf("insert stress order %d: %w", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	tx = nil

	started := time.Now()
	jobs := make(chan int64)
	errs := make(chan error, concurrency)
	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for orderID := range jobs {
				_, processErr := distribution.ProcessRecharge(ctx, orderID, buyerID, decimal.NewFromInt(1), decimal.Zero, decimal.NewFromInt(10))
				if processErr != nil {
					errs <- fmt.Errorf("order %d: %w", orderID, processErr)
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, orderID := range orderIDs {
			select {
			case jobs <- orderID:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Wait()
	close(errs)
	for processErr := range errs {
		return "", processErr
	}
	duration := time.Since(started)

	var eventCount, commissionCount, commissionMinor int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_recharge_events WHERE source_order_id = ANY($1::bigint[])`, pq.Array(orderIDs)).Scan(&eventCount); err != nil {
		return "", err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(amount_cny_minor), 0) FROM distribution_commissions WHERE source_order_id = ANY($1::bigint[])`, pq.Array(orderIDs)).Scan(&commissionCount, &commissionMinor); err != nil {
		return "", err
	}
	if eventCount != int64(orderCount) || commissionCount != 0 || commissionMinor != 0 {
		return "", fmt.Errorf("stress conservation mismatch: want events=%d commissions=0, got events=%d commissions=%d amount_minor=%d", orderCount, eventCount, commissionCount, commissionMinor)
	}
	rate := float64(orderCount) / duration.Seconds()
	return fmt.Sprintf("%d orders, %d workers, %d events, %d commissions, %s, %.1f orders/s", orderCount, concurrency, eventCount, commissionCount, duration.Round(time.Millisecond), rate), nil
}

type databaseSettings struct{ db *sql.DB }

func (s databaseSettings) Get(ctx context.Context, key string) (*service.Setting, error) {
	item := &service.Setting{Key: key}
	if err := s.db.QueryRowContext(ctx, `SELECT id, value, updated_at FROM settings WHERE key = $1`, key).Scan(&item.ID, &item.Value, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrSettingNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s databaseSettings) GetValue(ctx context.Context, key string) (string, error) {
	item, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return item.Value, nil
}

func (s databaseSettings) Set(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, key, value)
	return err
}

func (s databaseSettings) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		value, err := s.GetValue(ctx, key)
		if errors.Is(err, service.ErrSettingNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func (s databaseSettings) SetMultiple(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if err := s.Set(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (s databaseSettings) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, rows.Err()
}

func (s databaseSettings) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM settings WHERE key = $1`, key)
	return err
}

type gateEncryptor struct{}

func (gateEncryptor) Encrypt(plaintext string) (string, error) {
	return "gate:" + base64.RawStdEncoding.EncodeToString([]byte(plaintext)), nil
}

func (gateEncryptor) Decrypt(ciphertext string) (string, error) {
	const prefix = "gate:"
	if len(ciphertext) < len(prefix) || ciphertext[:len(prefix)] != prefix {
		return "", errors.New("invalid gate ciphertext")
	}
	decoded, err := base64.RawStdEncoding.DecodeString(ciphertext[len(prefix):])
	return string(decoded), err
}

var _ service.SettingRepository = databaseSettings{}
var _ service.SecretEncryptor = gateEncryptor{}
