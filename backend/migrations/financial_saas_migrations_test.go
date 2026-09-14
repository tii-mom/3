package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func normalizedMigration(t *testing.T, name string) string {
	t.Helper()
	content, err := FS.ReadFile(name)
	require.NoError(t, err)
	return strings.Join(strings.Fields(string(content)), " ")
}

// normalizedMigrationCode 与 normalizedMigration 相同，但会先剥掉 -- 行注释，
// 供「断言某段 SQL 不存在」的回归测试使用（避免注释里的示例文本把断言坐实）。
func normalizedMigrationCode(t *testing.T, name string) string {
	t.Helper()
	content, err := FS.ReadFile(name)
	require.NoError(t, err)
	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		if comment := strings.Index(line, "--"); comment >= 0 {
			lines[index] = line[:comment]
		}
	}
	return strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
}

func TestFinancialFoundationMigrationCreatesTenantScopeBeforeReferences(t *testing.T) {
	sql := normalizedMigration(t, "176_credit_accounts_and_vouchers.sql")
	tenantPos := strings.Index(sql, "CREATE TABLE IF NOT EXISTS saas_tenants")
	accountPos := strings.Index(sql, "CREATE TABLE IF NOT EXISTS user_credit_accounts")
	require.GreaterOrEqual(t, tenantPos, 0)
	require.Greater(t, accountPos, tenantPos)
	require.Contains(t, sql, "transferable_credit DECIMAL(20,8)")
	require.Contains(t, sql, "non_transferable_credit DECIMAL(20,8)")
	require.Contains(t, sql, "debt DECIMAL(20,8)")
	require.Contains(t, sql, "UNIQUE (tenant_id, idempotency_key)")
	require.Contains(t, sql, "('balance_voucher_enabled', 'false', NOW())")
}

func TestDistributionMigrationKeepsApprovedTiersAndIdempotency(t *testing.T) {
	sql := normalizedMigration(t, "177_distribution_program.sql")
	require.Contains(t, sql, "enabled BOOLEAN NOT NULL DEFAULT FALSE")
	require.Contains(t, sql, "stack_with_legacy BOOLEAN NOT NULL DEFAULT FALSE")
	require.Contains(t, sql, "(1, 100000::BIGINT, 1000, 400, 300, 200, 100)")
	require.Contains(t, sql, "(2, 1000000::BIGINT, 1500, 600, 400, 300, 200)")
	require.Contains(t, sql, "(3, 10000000::BIGINT, 2000, 800, 600, 400, 200)")
	require.Contains(t, sql, "UNIQUE (program_id, source_order_id)")
	require.Contains(t, sql, "UNIQUE (program_id, source_order_id, beneficiary_user_id, depth)")
	require.Contains(t, sql, "first_recharge_bonus_cap_usd DECIMAL(20,8) NOT NULL DEFAULT 10000")
}

func TestSaaSMigrationKeepsWholesaleAndPartnerFundsSeparate(t *testing.T) {
	sql := normalizedMigration(t, "178_saas_control_plane.sql")
	require.Contains(t, sql, "balance_usd DECIMAL(20,8) NOT NULL DEFAULT 0")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS saas_partner_wallets")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS saas_partner_withdrawals")
	require.Contains(t, sql, "('saas_control_plane_enabled', 'false', NOW())")
	require.Contains(t, sql, "UNIQUE (tenant_id, idempotency_key)")
}

func TestFinancialRuntimeControlsAreAdditiveAndDefaultToShadowMode(t *testing.T) {
	sql := normalizedMigration(t, "179_financial_runtime_controls.sql")
	require.Contains(t, sql, "('credit_bucket_enforce_enabled', 'false', NOW())")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS financial_reconciliation_issues")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS distribution_policy_versions")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS config_version")
}

func TestDistributionReversalMigrationAddsAuditAndDebtRecovery(t *testing.T) {
	sql := normalizedMigration(t, "180_distribution_reversals.sql")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS distribution_reversal_events")
	require.Contains(t, sql, "debt_cny_minor BIGINT NOT NULL DEFAULT 0")
	require.Contains(t, sql, "lifetime_reversed_cny_minor BIGINT NOT NULL DEFAULT 0")
	require.Contains(t, sql, "legacy_rebate_usd DECIMAL(20,8) NOT NULL DEFAULT 0")
	require.Contains(t, sql, "idx_user_affiliate_ledger_order_reversal")
	require.Contains(t, sql, "UNIQUE (program_id, recharge_event_id)")
	require.Contains(t, sql, "WHERE first_recharge_bonus_usd > 0 AND status = 'APPLIED'")
	require.Contains(t, sql, "CHECK (reversal_type IN ('CHARGEBACK', 'REFUND', 'ADMIN_CORRECTION'))")
}

func TestDistributionProfileRepairBackfillsZeroSafeProfiles(t *testing.T) {
	sql := normalizedMigration(t, "185_distribution_profile_repair.sql")
	require.Contains(t, sql, "INSERT INTO distribution_members")
	require.Contains(t, sql, "INSERT INTO distribution_relations")
	require.Contains(t, sql, "chain.depth < 5")
	require.Contains(t, sql, "INSERT INTO distribution_cash_wallets")
	require.Contains(t, sql, "ON CONFLICT DO NOTHING")
}

func TestSaaSApplicationMigrationSeparatesLeadReviewFromTenantCreation(t *testing.T) {
	sql := normalizedMigration(t, "186_saas_tenant_applications.sql")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS saas_tenant_applications")
	require.Contains(t, sql, "status IN ('SUBMITTED', 'CONTACTED', 'APPROVED', 'REJECTED')")
	require.Contains(t, sql, "WHERE status IN ('SUBMITTED', 'CONTACTED')")
	require.Contains(t, sql, "tenant_id BIGINT UNIQUE REFERENCES saas_tenants(id)")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS saas_tenant_application_events")
	require.Contains(t, sql, "('saas_application_enabled', 'false', NOW())")
}

func TestDistributionTierOverrideMigrationIsBoundedAndIndexed(t *testing.T) {
	sql := normalizedMigration(t, "187_distribution_tier_overrides.sql")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS tier_override SMALLINT")
	require.Contains(t, sql, "tier_override BETWEEN 1 AND 3")
	require.Contains(t, sql, "tier_override_by BIGINT REFERENCES users(id)")
	require.Contains(t, sql, "idx_distribution_members_tier_override")
}

func TestComputeCompanyT0MigrationUnifiesWalletPolicy(t *testing.T) {
	sql := normalizedMigration(t, "188_compute_company_t0.sql")
	require.Contains(t, sql, "CHECK (tier BETWEEN 0 AND 3)")
	require.Contains(t, sql, "CHECK (tier_override IS NULL OR tier_override BETWEEN 0 AND 3)")
	require.Contains(t, sql, "0, 0, 1000, 0, 0, 0, 0")
	require.Contains(t, sql, "INSERT INTO settings (key, value, updated_at) VALUES ('distribution_usd_to_cny_rate', '7.15', NOW())")
	require.Contains(t, sql, "withdrawal_min_cny_minor = 2000")
	require.Contains(t, sql, "withdrawal_daily_limit = 1")
	require.Contains(t, sql, "stack_with_legacy = FALSE")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS distribution_usd_conversions")
	require.Contains(t, sql, "UNIQUE (program_id, idempotency_key)")
}

func TestPromotionSingleLevelMigrationBumpsConfigVersion(t *testing.T) {
	// 只断言可执行 SQL：注释里会解释这处坑，若连注释一起匹配就会自伤。
	sql := normalizedMigrationCode(t, "211_promotion_single_level.sql")
	// 回归防线：配置版本号必须由 current_config_version + 1 推导。
	// 硬编码成 2 会撞上 188 已占用的版本 2：INSERT 全被 ON CONFLICT 静默吸收、
	// UPDATE 的 WHERE 也不命中，整份迁移空转且不报任何错。
	require.Contains(t, sql, "next_version := previous_version + 1")
	require.NotContains(t, sql, "config_version = 2")
	require.NotContains(t, sql, "current_config_version < 2")
	// 单层三档阶梯：T0~T2 只填第 1 层费率，第 2~5 层一律为 0。
	require.Contains(t, sql, "(compute_program_id, next_version, 0, 0, 500, 0, 0, 0, 0)")
	require.Contains(t, sql, "(compute_program_id, next_version, 1, 500000, 800, 0, 0, 0, 0)")
	require.Contains(t, sql, "(compute_program_id, next_version, 2, 5000000, 1000, 0, 0, 0, 0)")
	require.Contains(t, sql, "name = '推广计划'")
	require.Contains(t, sql, "INSERT INTO distribution_policy_versions")
	// 档位从 T0~T3 收成 T0~T2，手工指定档位必须同步作废并收紧约束。
	require.Contains(t, sql, "CHECK (tier_override IS NULL OR tier_override BETWEEN 0 AND 2)")
	// 直属业绩按 depth = 1 重算，且必须先重算业绩、再算档位（两条独立 UPDATE）。
	require.Contains(t, sql, "e.status = 'APPLIED'")
	require.Contains(t, sql, "r.depth = 1")
	teamVolumePos := strings.Index(sql, "SET team_volume_cny_minor = COALESCE")
	tierPos := strings.Index(sql, "SET current_tier = COALESCE")
	require.GreaterOrEqual(t, teamVolumePos, 0)
	require.Greater(t, tierPos, teamVolumePos)
}

func TestPromotionDropTiersMigrationOnlyZeroesState(t *testing.T) {
	// 只断言可执行 SQL：注释里会解释决策，若连注释一起匹配就会自伤。
	sql := normalizedMigrationCode(t, "212_promotion_drop_tiers.sql")
	// 档位状态清零，且两条 UPDATE 都要带 WHERE 保证幂等。
	require.Contains(t, sql, "SET tier_override = NULL")
	require.Contains(t, sql, "SET current_tier = 0")
	require.Contains(t, sql, "WHERE current_tier <> 0")
	require.Contains(t, sql, "WHERE tier_override IS NOT NULL")
	// 只打废弃注释，绝不做破坏性 DDL。
	require.Contains(t, sql, "COMMENT ON TABLE distribution_tier_configs")
	require.NotContains(t, sql, "DROP COLUMN")
	require.NotContains(t, sql, "DROP TABLE")
	// 不触碰用户账户余额：迁移里不应出现钱包表与额度表。
	require.NotContains(t, sql, "user_credit_accounts")
	require.NotContains(t, sql, "distribution_cash_wallets")
}

func TestShopWalletDeductionMigrationIsAdditiveAndGuarded(t *testing.T) {
	// 只断言可执行 SQL：注释里会解释「返佣基数取实付」这处坑，连注释一起匹配会自伤。
	sql := normalizedMigrationCode(t, "213_shop_wallet_deduction.sql")
	// 只加两列，默认 0，保证存量订单语义不变。
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS wallet_applied_cny_minor BIGINT NOT NULL DEFAULT 0")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS payable_cny_minor BIGINT NOT NULL DEFAULT 0")
	// 存量回填必须带完整 WHERE，避免重复执行时把已抵扣订单的 payable 改回原价。
	require.Contains(t, sql, "SET payable_cny_minor = snapshot_price_cny_minor")
	require.Contains(t, sql, "WHERE wallet_applied_cny_minor = 0")
	require.Contains(t, sql, "AND payable_cny_minor = 0")
	require.Contains(t, sql, "AND snapshot_price_cny_minor > 0")
	// 恒等式与取值上下界由 CHECK 约束兜底。
	require.Contains(t, sql, "CHECK (wallet_applied_cny_minor >= 0 AND wallet_applied_cny_minor <= snapshot_price_cny_minor)")
	require.Contains(t, sql, "CHECK (payable_cny_minor = snapshot_price_cny_minor - wallet_applied_cny_minor)")
	// 幂等：两条约束都先 DROP IF EXISTS 再 ADD。
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS shop_orders_wallet_applied_range_check")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS shop_orders_payable_identity_check")
	// 纯增量：不做破坏性 DDL，也不碰额度表（人民币返点与 API 美金额度是两种钱）。
	require.NotContains(t, sql, "DROP COLUMN")
	require.NotContains(t, sql, "DROP TABLE")
	require.NotContains(t, sql, "user_credit_accounts")
}

func TestDistributionConversionIdempotencyIsUserScoped(t *testing.T) {
	sql := normalizedMigration(t, "189_distribution_conversion_idempotency_scope.sql")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS distribution_usd_conversions_program_id_idempotency_key_key")
	require.Contains(t, sql, "ON distribution_usd_conversions(program_id, user_id, idempotency_key)")
}

func TestDistributionConversionUsesPaymentPurchaseMultiplier(t *testing.T) {
	sql := normalizedMigration(t, "190_distribution_purchase_multiplier.sql")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS cny_to_usd_rate DECIMAL(20,10)")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS rate_source VARCHAR(64)")
	require.Contains(t, sql, "legacy_usd_to_cny_rate")
	require.Contains(t, sql, "distribution_usd_conversions_cny_to_usd_rate_check")
	require.Contains(t, sql, "IF NOT EXISTS (")
	require.Contains(t, sql, "FROM pg_constraint")
}

func TestFreshInstallFinancialFeatureDefaultsAreEmptyDatabaseOnly(t *testing.T) {
	sql := normalizedMigration(t, "192_fresh_install_financial_feature_defaults.sql")
	require.Contains(t, sql, "SELECT COUNT(*) INTO user_count FROM users")
	require.Contains(t, sql, "SELECT COUNT(*) INTO payment_order_count FROM payment_orders")
	require.Contains(t, sql, "SELECT COUNT(*) INTO credit_account_count FROM user_credit_accounts")
	require.Contains(t, sql, "SELECT COUNT(*) INTO credit_ledger_count FROM user_credit_ledger")
	require.Contains(t, sql, "SELECT COUNT(*) INTO voucher_count FROM balance_vouchers")
	require.Contains(t, sql, "value = 'true'")
	require.Contains(t, sql, "'credit_bucket_enforce_enabled'")
	require.Contains(t, sql, "'balance_voucher_enabled'")
	require.Contains(t, sql, "'distribution_enabled'")
	require.Contains(t, sql, "SET enabled = TRUE")
}

func TestStaleReconciliationIssueResolutionIsAuditOnlyAndBalanceGuarded(t *testing.T) {
	sql := normalizedMigration(t, "194_resolve_stale_reconciliation_issues.sql")
	require.Contains(t, sql, "UPDATE financial_reconciliation_issues AS i")
	require.Contains(t, sql, "SET status = 'RESOLVED'")
	require.Contains(t, sql, "resolved_at = COALESCE(i.resolved_at, NOW())")
	require.Contains(t, sql, "'auto_resolved_by', '194_resolve_stale_reconciliation_issues'")
	require.Contains(t, sql, "i.status = 'OPEN'")
	require.Contains(t, sql, "u.balance = a.transferable_credit + a.non_transferable_credit - a.debt")
	require.NotContains(t, strings.ToLower(sql), "delete from")
	require.NotContains(t, strings.ToLower(sql), "drop table")
}

func TestPostRolloutReconciliationIssueResolutionIsAuditOnlyAndBalanceGuarded(t *testing.T) {
	sql := normalizedMigration(t, "195_resolve_currently_balanced_reconciliation_issues.sql")
	require.Contains(t, sql, "UPDATE financial_reconciliation_issues AS i")
	require.Contains(t, sql, "SET status = 'RESOLVED'")
	require.Contains(t, sql, "resolved_at = COALESCE(i.resolved_at, NOW())")
	require.Contains(t, sql, "'auto_resolved_by', '195_resolve_currently_balanced_reconciliation_issues'")
	require.Contains(t, sql, "i.status = 'OPEN'")
	require.Contains(t, sql, "u.balance = a.transferable_credit + a.non_transferable_credit - a.debt")
	require.NotContains(t, strings.ToLower(sql), "delete from")
	require.NotContains(t, strings.ToLower(sql), "drop table")
}

func TestLatestPostRolloutReconciliationIssueResolutionIsAuditOnlyAndBalanceGuarded(t *testing.T) {
	sql := normalizedMigration(t, "196_resolve_post_rollout_balanced_reconciliation_issues.sql")
	require.Contains(t, sql, "UPDATE financial_reconciliation_issues AS i")
	require.Contains(t, sql, "SET status = 'RESOLVED'")
	require.Contains(t, sql, "resolved_at = COALESCE(i.resolved_at, NOW())")
	require.Contains(t, sql, "'auto_resolved_by', '196_resolve_post_rollout_balanced_reconciliation_issues'")
	require.Contains(t, sql, "i.status = 'OPEN'")
	require.Contains(t, sql, "u.balance = a.transferable_credit + a.non_transferable_credit - a.debt")
	require.NotContains(t, strings.ToLower(sql), "delete from")
	require.NotContains(t, strings.ToLower(sql), "drop table")
}

func TestRecurringBalancedReconciliationIssueResolutionIsAuditOnlyAndBalanceGuarded(t *testing.T) {
	sql := normalizedMigration(t, "197_resolve_balanced_reconciliation_issues.sql")
	require.Contains(t, sql, "UPDATE financial_reconciliation_issues AS i")
	require.Contains(t, sql, "SET status = 'RESOLVED'")
	require.Contains(t, sql, "resolved_at = COALESCE(i.resolved_at, NOW())")
	require.Contains(t, sql, "'auto_resolved_by', '197_resolve_balanced_reconciliation_issues'")
	require.Contains(t, sql, "i.status = 'OPEN'")
	require.Contains(t, sql, "u.balance = a.transferable_credit + a.non_transferable_credit - a.debt")
	require.NotContains(t, strings.ToLower(sql), "delete from")
	require.NotContains(t, strings.ToLower(sql), "drop table")
}
