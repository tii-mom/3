package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// 推广计划已收敛为「单层 + 无档位」：概览数据只看
// 「我直接邀请了多少人 / 他们贡献了多少业绩」以及佣金钱包余额，
// 因此查询里不再出现 distribution_members / distribution_tier_configs。
func TestDistributionDashboardDisabledUserWithoutProfileReturnsZeroValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, enabled, commission_freeze_hours, withdrawal_min_cny_minor, withdrawal_daily_limit FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "enabled", "freeze_hours", "minimum", "daily_limit"}).AddRow(7, false, 168, 2000, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT (SELECT COUNT(*) FROM user_affiliates WHERE inviter_id = $1),
       COALESCE((
           SELECT SUM(o.payable_cny_minor)
           FROM shop_orders o
           JOIN user_affiliates ua ON ua.user_id = o.user_id AND ua.inviter_id = $1
           WHERE o.tenant_id = 1 AND o.status IN ('paid', 'fulfilled')
       ), 0)`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"invitee_count", "team_volume_cny_minor"}).AddRow(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, lifetime_earned_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"available", "frozen", "withdrawing", "debt", "lifetime"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM settings WHERE key = $1`)).
		WithArgs(SettingBalanceRechargeMult).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("0.14"))

	dashboard, err := NewDistributionService(db, nil).Dashboard(context.Background(), 42)
	require.NoError(t, err)
	require.False(t, dashboard.Enabled)
	require.Zero(t, dashboard.InviteeCount)
	require.Zero(t, dashboard.InviteeSpendMinor)
	require.Zero(t, dashboard.AvailableMinor)
	require.Zero(t, dashboard.FrozenMinor)
	require.Equal(t, "0.14", dashboard.BalanceRechargeMultiplier)
	require.NoError(t, mock.ExpectationsWereMet())
}
