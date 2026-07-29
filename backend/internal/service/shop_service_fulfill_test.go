package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAdminFulfillOrderUpdatesOrderAndPayment(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status,
       snapshot_name, snapshot_description, snapshot_image_url, snapshot_product_type,
       snapshot_price_cny_minor, snapshot_grant_usd_amount::text, snapshot_commission_bps,
       fulfillment_note, paid_at, fulfilled_at
FROM shop_orders
WHERE tenant_id = 1 AND id = $1
FOR UPDATE`)).
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "product_id", "payment_order_id", "status", "fulfillment_status", "commission_status",
			"snapshot_name", "snapshot_description", "snapshot_image_url", "snapshot_product_type",
			"snapshot_price_cny_minor", "snapshot_grant_usd_amount", "snapshot_commission_bps",
			"fulfillment_note", "paid_at", "fulfilled_at",
		}).AddRow(int64(88), int64(21), int64(55), int64(910), "paid", "pending", "none", "X", "Y", "/img.png", "virtual", int64(1299), "0", int64(1000), "", nil, nil))
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE shop_orders
SET status = 'fulfilled',
    fulfillment_status = 'fulfilled',
    fulfillment_note = $2,
    fulfilled_at = COALESCE(fulfilled_at, $3),
    updated_at = NOW()
WHERE id = $1`)).
		WithArgs(int64(88), "manual delivery note", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(910), OrderStatusCompleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = svc.AdminFulfillOrder(context.Background(), 88, "manual delivery note")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminFulfillOrderRejectsBalanceProducts(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewShopService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status,
       snapshot_name, snapshot_description, snapshot_image_url, snapshot_product_type,
       snapshot_price_cny_minor, snapshot_grant_usd_amount::text, snapshot_commission_bps,
       fulfillment_note, paid_at, fulfilled_at
FROM shop_orders
WHERE tenant_id = 1 AND id = $1
FOR UPDATE`)).
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "product_id", "payment_order_id", "status", "fulfillment_status", "commission_status",
			"snapshot_name", "snapshot_description", "snapshot_image_url", "snapshot_product_type",
			"snapshot_price_cny_minor", "snapshot_grant_usd_amount", "snapshot_commission_bps",
			"fulfillment_note", "paid_at", "fulfilled_at",
		}).AddRow(int64(88), int64(21), int64(55), int64(910), "paid", "pending", "none", "X", "Y", "/img.png", ShopProductTypePlatformUSDBalance, int64(1299), "12.5", int64(1000), "", nil, nil))
	mock.ExpectRollback()

	err = svc.AdminFulfillOrder(context.Background(), 88, "manual delivery note")
	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminFulfillOrderRequiresNote(t *testing.T) {
	svc := NewShopService(&sql.DB{})
	err := svc.AdminFulfillOrder(context.Background(), 88, "   ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "fulfillment note is required")
}
