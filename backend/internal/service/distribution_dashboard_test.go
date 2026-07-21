package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDistributionDashboardDisabledUserWithoutProfileReturnsZeroValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, enabled, commission_freeze_hours, withdrawal_min_cny_minor, withdrawal_daily_limit FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "enabled", "freeze_hours", "minimum", "daily_limit"}).AddRow(7, false, 168, 2000, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT team_volume_cny_minor, current_tier, tier_override FROM distribution_members WHERE program_id = $1 AND user_id = $2`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"team_volume_cny_minor", "current_tier", "tier_override"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT depth, COUNT(*) FROM distribution_relations WHERE program_id = $1 AND ancestor_user_id = $2 AND depth BETWEEN 1 AND 5 GROUP BY depth`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"depth", "count"}))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT r.depth, COUNT(DISTINCT r.descendant_user_id), COALESCE(SUM(e.base_cny_minor), 0)
FROM distribution_relations r
LEFT JOIN distribution_recharge_events e
  ON e.program_id = r.program_id
 AND e.user_id = r.descendant_user_id
 AND e.status = 'APPLIED'
WHERE r.program_id = $1 AND r.ancestor_user_id = $2 AND r.depth BETWEEN 1 AND 5
GROUP BY r.depth`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"depth", "members", "recharge"}))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT depth,
       COALESCE(SUM(CASE WHEN status <> 'REVERSED' THEN amount_cny_minor ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'AVAILABLE' THEN amount_cny_minor ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'FROZEN' THEN amount_cny_minor ELSE 0 END), 0)
FROM distribution_commissions
WHERE program_id = $1 AND beneficiary_user_id = $2 AND depth BETWEEN 1 AND 5
GROUP BY depth`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"depth", "commission", "available", "frozen"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, lifetime_earned_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"available", "frozen", "withdrawing", "debt", "lifetime"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tier, threshold_cny_minor, level1_bps, level2_bps, level3_bps, level4_bps, level5_bps FROM distribution_tier_configs WHERE program_id = $1 AND config_version = (SELECT current_config_version FROM distribution_programs WHERE id = $1) ORDER BY tier`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"tier", "threshold", "l1", "l2", "l3", "l4", "l5"}).
			AddRow(0, 0, 1000, 0, 0, 0, 0).
			AddRow(1, 100000, 1000, 400, 300, 200, 100).
			AddRow(2, 1000000, 1500, 600, 400, 300, 200).
			AddRow(3, 10000000, 2000, 800, 600, 400, 200))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM settings WHERE key = 'distribution_usd_to_cny_rate'`)).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("7.15"))

	dashboard, err := NewDistributionService(db, nil).Dashboard(context.Background(), 42)
	require.NoError(t, err)
	require.False(t, dashboard.Enabled)
	require.Zero(t, dashboard.TeamVolumeMinor)
	require.Zero(t, dashboard.CurrentTier)
	require.Zero(t, dashboard.AvailableMinor)
	require.Zero(t, dashboard.FrozenMinor)
	require.Empty(t, dashboard.LevelCounts)
	require.Len(t, dashboard.Levels, 5)
	require.Len(t, dashboard.Tiers, 4)
	require.NoError(t, mock.ExpectationsWereMet())
}
