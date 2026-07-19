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
