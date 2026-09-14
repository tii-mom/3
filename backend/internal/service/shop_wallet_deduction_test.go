package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// 推广计划未启用（或不存在）时不暴露抵扣入口：Enabled 为 false 且余额为 0。
func TestWalletBalanceDisabledWhenProgramMissing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' AND enabled = TRUE`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	balance, err := svc.WalletBalance(context.Background(), 7)
	require.NoError(t, err)
	require.NotNil(t, balance)
	require.False(t, balance.Enabled)
	require.Zero(t, balance.AvailableCNYMinor)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 有余额时读到可用/冻结分额，且 Enabled 为 true（前端据此展示抵扣开关）。
func TestWalletBalanceReadsAvailableAndFrozen(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' AND enabled = TRUE`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT available_cny_minor, frozen_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`)).
		WithArgs(int64(3), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"available_cny_minor", "frozen_cny_minor"}).AddRow(int64(2500), int64(700)))

	balance, err := svc.WalletBalance(context.Background(), 7)
	require.NoError(t, err)
	require.True(t, balance.Enabled)
	require.Equal(t, int64(2500), balance.AvailableCNYMinor)
	require.Equal(t, int64(700), balance.FrozenCNYMinor)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 订单关闭（取消/超时/失败）必须把已抵扣的返点余额退回可用余额，且退款流水幂等。
func TestMarkPaymentOrderClosedRefundsWalletDeduction(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM shop_orders WHERE tenant_id = 1 AND payment_order_id = $1 AND status = 'pending' FOR UPDATE`)).
		WithArgs(int64(910)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE shop_orders SET status = $2, fulfillment_status = 'failed', updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(42), "cancelled").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// refundShopOrderWalletTx：读回抵扣金额 → 找推广计划 → 幂等检查 → 回补余额 → 写退款流水。
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, wallet_applied_cny_minor FROM shop_orders WHERE tenant_id = 1 AND id = $1 FOR UPDATE`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "wallet_applied_cny_minor"}).AddRow(int64(7), int64(500)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' AND enabled = TRUE`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM distribution_wallet_ledger WHERE program_id = $1 AND idempotency_key = $2)`)).
		WithArgs(int64(3), "shop:order:42:wallet-refund").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`)).
		WithArgs(int64(3), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE distribution_cash_wallets SET available_cny_minor = available_cny_minor + $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`)).
		WithArgs(int64(3), int64(7), int64(500)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO distribution_wallet_ledger (program_id, tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, debt_after, idempotency_key, metadata)
SELECT $1, 1, $2, 'shop_wallet_deduction_refund', $3, 'shop_order', $4, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, $5, jsonb_build_object('label', '订单关闭退回抵扣')
FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2
ON CONFLICT DO NOTHING`)).
		WithArgs(int64(3), int64(7), int64(500), "42", "shop:order:42:wallet-refund").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = svc.MarkPaymentOrderClosed(context.Background(), 910, "cancelled")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 幂等：退款流水已存在时不再重复回补余额。
func TestMarkPaymentOrderClosedSkipsDuplicateRefund(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM shop_orders WHERE tenant_id = 1 AND payment_order_id = $1 AND status = 'pending' FOR UPDATE`)).
		WithArgs(int64(910)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE shop_orders SET status = $2, fulfillment_status = 'failed', updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(42), "failed").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, wallet_applied_cny_minor FROM shop_orders WHERE tenant_id = 1 AND id = $1 FOR UPDATE`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "wallet_applied_cny_minor"}).AddRow(int64(7), int64(500)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' AND enabled = TRUE`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM distribution_wallet_ledger WHERE program_id = $1 AND idempotency_key = $2)`)).
		WithArgs(int64(3), "shop:order:42:wallet-refund").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectCommit()

	err = svc.MarkPaymentOrderClosed(context.Background(), 910, "failed")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 全额返点余额抵扣的订单没有外部支付单：交付链路必须照样跑完，
// 且因为实付为 0，不产生任何返佣（避免「返点→抵扣→再返点」的无锚增发）。
func TestFulfillWalletPaidOrderSkipsPaymentWriteAndCommission(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status, snapshot_name, snapshot_description,
       snapshot_image_url, snapshot_product_type, snapshot_price_cny_minor, wallet_applied_cny_minor, payable_cny_minor,
       snapshot_grant_usd_amount::text,
       snapshot_commission_bps, fulfillment_note, paid_at
FROM shop_orders
WHERE tenant_id = 1 AND id = $1
FOR UPDATE`)).
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "product_id", "payment_order_id", "status", "fulfillment_status", "commission_status",
			"snapshot_name", "snapshot_description", "snapshot_image_url", "snapshot_product_type",
			"snapshot_price_cny_minor", "wallet_applied_cny_minor", "payable_cny_minor",
			"snapshot_grant_usd_amount", "snapshot_commission_bps", "fulfillment_note", "paid_at",
		}).AddRow(int64(88), int64(21), int64(55), nil, "pending", "pending", "pending",
			"X", "Y", "/img.png", ShopProductTypeVirtual,
			int64(1299), int64(1299), int64(0), "0", int64(500), "", nil))
	// 实付为 0 → 直接判定为无返佣，不写 shop_commission_records。
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE shop_orders SET commission_status = 'none', updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(88)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE shop_products SET stock_quantity = CASE WHEN stock_quantity IS NULL THEN NULL ELSE stock_quantity - 1 END, sold_count = sold_count + 1, updated_at = NOW() WHERE id = $1 AND (stock_quantity IS NULL OR stock_quantity > 0)`)).
		WithArgs(int64(55)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE shop_orders SET status = 'paid', fulfillment_status = 'pending', paid_at = COALESCE(paid_at, $2), updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(88), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 关键回归：没有外部支付单时不写 payment_orders，且事务只提交一次。
	mock.ExpectCommit()

	err = svc.FulfillWalletPaidOrder(context.Background(), 88)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 幂等重放：订单已是 paid + pending 时只回写支付单并提交，
// 不允许在 completeLinkedPaymentOrder 之后再 Commit 一次（会得到 sql.ErrTxDone）。
func TestFulfillPaidPaymentOrderIdempotentReplayCommitsOnce(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status, snapshot_name, snapshot_description,
       snapshot_image_url, snapshot_product_type, snapshot_price_cny_minor, wallet_applied_cny_minor, payable_cny_minor,
       snapshot_grant_usd_amount::text,
       snapshot_commission_bps, fulfillment_note, paid_at
FROM shop_orders
WHERE tenant_id = 1 AND payment_order_id = $1
FOR UPDATE`)).
		WithArgs(int64(910)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "product_id", "payment_order_id", "status", "fulfillment_status", "commission_status",
			"snapshot_name", "snapshot_description", "snapshot_image_url", "snapshot_product_type",
			"snapshot_price_cny_minor", "wallet_applied_cny_minor", "payable_cny_minor",
			"snapshot_grant_usd_amount", "snapshot_commission_bps", "fulfillment_note", "paid_at",
		}).AddRow(int64(88), int64(21), int64(55), int64(910), "paid", "pending", "pending",
			"X", "Y", "/img.png", ShopProductTypeVirtual,
			int64(1299), int64(0), int64(1299), "0", int64(500), "", nil))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1 AND status <> $2`)).
		WithArgs(int64(910), OrderStatusCompleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = svc.FulfillPaidPaymentOrder(context.Background(), 910)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 非 pending 的商城订单无需处理：直接返回（可能已支付或已被关闭）。
func TestMarkPaymentOrderClosedNoopWhenOrderNotPending(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM shop_orders WHERE tenant_id = 1 AND payment_order_id = $1 AND status = 'pending' FOR UPDATE`)).
		WithArgs(int64(910)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()

	err = svc.MarkPaymentOrderClosed(context.Background(), 910, "cancelled")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
