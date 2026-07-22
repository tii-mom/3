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
