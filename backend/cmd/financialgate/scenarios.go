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
	var existingUsers int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&existingUsers); err != nil {
		return nil, err
	}
	if existingUsers != 0 {
		return nil, fmt.Errorf("refusing destructive scenarios: expected an empty users table, found %d rows", existingUsers)
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
	if err := seedDistributionChain(ctx, db, users[:6]); err != nil {
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
		"distribution_concurrency": fmt.Sprintf("%d duplicate deliveries -> 1 event, 5 commissions, CNY 200.00 total", scenarioWorkers),
		"first_recharge_bonus":     "USD 10,000 principal -> USD 1,000 non-transferable bonus",
		"voucher_lifecycle":        voucherResult,
		"wholesale_billing":        wholesaleResult,
	}, nil
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
	var events, commissions int64
	var bonus string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(first_recharge_bonus_usd), 0)::text FROM distribution_recharge_events WHERE source_order_id = $1`, orderID).Scan(&events, &bonus); err != nil {
		return err
	}
	if events != 1 || bonus != "1000.00000000" {
		return fmt.Errorf("unexpected recharge event result: events=%d bonus=%s", events, bonus)
	}
	var amount int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(amount_cny_minor), 0) FROM distribution_commissions WHERE source_order_id = $1`, orderID).Scan(&commissions, &amount); err != nil {
		return err
	}
	if commissions != 5 || amount != 20000 {
		return fmt.Errorf("unexpected commissions: count=%d amount_minor=%d", commissions, amount)
	}
	var matching int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM distribution_commissions
WHERE source_order_id = $1
  AND (depth, rate_bps, amount_cny_minor) IN ((1,1000,10000),(2,400,4000),(3,300,3000),(4,200,2000),(5,100,1000))`, orderID).Scan(&matching); err != nil {
		return err
	}
	if matching != 5 {
		return fmt.Errorf("commission depth/rate schedule mismatch: matched=%d", matching)
	}
	var relationCount, tierOneMembers int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_relations WHERE descendant_user_id = $1 AND depth BETWEEN 0 AND 5`, users[5]).Scan(&relationCount); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_members WHERE user_id = ANY($1::bigint[]) AND team_volume_cny_minor = 100000 AND current_tier = 1`, pq.Array(users)).Scan(&tierOneMembers); err != nil {
		return err
	}
	if relationCount != 6 || tierOneMembers != 6 {
		return fmt.Errorf("distribution relation/team mismatch: relations=%d tier1_members=%d", relationCount, tierOneMembers)
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
	if eventCount != int64(orderCount) || commissionCount != int64(orderCount*5) || commissionMinor != int64(orderCount*20) {
		return "", fmt.Errorf("stress conservation mismatch: events=%d commissions=%d amount_minor=%d", eventCount, commissionCount, commissionMinor)
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
