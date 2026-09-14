package service

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// 佣金明细必须同时覆盖两张表：
//   - shop_commission_records  商城订单佣金（当前唯一在写入的来源）
//   - distribution_commissions 历史充值返佣（已停止写入，仅留痕）
//
// 只读后者会导致「充值返佣下线后，明细永远为空」。
func TestDistributionLedgerMergesShopAndLegacyCommissions(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT (SELECT COUNT(*) FROM distribution_commissions WHERE program_id = $1 AND beneficiary_user_id = $2)
     + (SELECT COUNT(*) FROM shop_commission_records WHERE tenant_id = 1 AND beneficiary_user_id = $2)`)).
		WithArgs(int64(9), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	now := time.Now().UTC()
	mock.ExpectQuery(`WITH merged AS`).
		WithArgs(int64(9), int64(42), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"source", "id", "source_order_id", "source_user_id", "depth", "tier", "rate_bps",
			"base_cny_minor", "amount_cny_minor", "team_volume_cny_minor", "status", "frozen_until", "created_at",
		}).
			// 商城订单佣金：单层（depth=1）+ 无档位（tier=0），费率取商品返佣比例。
			AddRow("shop", int64(7), int64(501), int64(88), 1, 0, 800,
				int64(36850), int64(2948), int64(36850), "FROZEN", now, now).
			// 历史充值返佣：保留原有 depth / tier，便于追溯。
			AddRow("distribution", int64(7), int64(9001), int64(77), 2, 1, 400,
				int64(10000), int64(400), int64(10000), "AVAILABLE", now, now))

	items, total, err := NewDistributionService(db, nil).Ledger(context.Background(), 42, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)

	// 两表 id 会重号（都是 7），必须靠 source 区分 —— 前端 key 也依赖它。
	require.Equal(t, "shop", items[0].Source)
	require.Equal(t, "distribution", items[1].Source)
	require.Equal(t, int64(7), items[0].ID)
	require.Equal(t, int64(7), items[1].ID)

	require.Equal(t, int64(36850), items[0].BaseMinor)
	require.Equal(t, int64(2948), items[0].AmountMinor)
	require.Equal(t, 800, items[0].RateBPS)
	require.Equal(t, 1, items[0].Depth)

	require.NoError(t, mock.ExpectationsWereMet())
}

// 后台佣金列表同样要合并两张表，否则「充值返佣已下线」后后台看到的是一条空表。
func TestAdminListCommissionsMergesShopAndLegacyCommissions(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT (SELECT COUNT(*) FROM distribution_commissions)
     + (SELECT COUNT(*) FROM shop_commission_records)`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	now := time.Now().UTC()
	mock.ExpectQuery(`WITH merged AS`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"source", "id", "source_order_id", "source_user_id", "beneficiary_user_id",
			"depth", "tier", "rate_bps", "base_cny_minor", "amount_cny_minor",
			"team_volume_cny_minor", "status", "frozen_until", "created_at",
		}).
			AddRow("shop", int64(3), int64(501), int64(88), int64(42), 1, 0, 800,
				int64(36850), int64(2948), int64(36850), "FROZEN", now, now).
			AddRow("distribution", int64(3), int64(9001), int64(77), int64(41), 2, 1, 400,
				int64(10000), int64(400), int64(10000), "AVAILABLE", now, now))

	items, total, err := NewDistributionService(db, nil).AdminListCommissions(context.Background(), 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)
	require.Equal(t, "shop", items[0].Source)
	require.Equal(t, "distribution", items[1].Source)
	require.Equal(t, int64(42), items[0].BeneficiaryUserID)
	require.Equal(t, int64(41), items[1].BeneficiaryUserID)

	require.NoError(t, mock.ExpectationsWereMet())
}
