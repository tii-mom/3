//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func expectUsageDebit(mock sqlmock.Sqlmock, userID, apiKeyID int64, requestID, transferable, nonTransferable, debt, amount, expectedTransferable, expectedNonTransferable, expectedDebt, expectedBalance string) {
	mock.ExpectExec(`INSERT INTO user_credit_accounts`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT 1 FROM user_credit_ledger`).
		WithArgs("usage:" + sqlmockAnyInt(apiKeyID) + ":" + requestID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery(`SELECT transferable_credit::text, non_transferable_credit::text, debt::text`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"transferable_credit", "non_transferable_credit", "debt"}).AddRow(transferable, nonTransferable, debt))
	mock.ExpectQuery(`SELECT value FROM settings`).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("false"))
	mock.ExpectExec(`UPDATE user_credit_accounts`).
		WithArgs(userID, expectedTransferable, expectedNonTransferable, expectedDebt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`UPDATE users`).
		WithArgs(userID, expectedBalance, "0", false, false, amount).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))
	mock.ExpectExec(`INSERT INTO user_credit_ledger`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func sqlmockAnyInt(value int64) string {
	return fmt.Sprintf("%d", value)
}

func TestDeductUsageBillingBalance_UsesCreditBuckets(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	expectUsageDebit(mock, 42, 7, "request-sufficient", "0", "10", "0", "-2.5", "0", "7.5", "0", "7.5")
	mock.ExpectCommit()

	newBalance, sufficient, err := deductUsageBillingBalance(ctx, tx, 42, 2.5, 7, "request-sufficient")
	require.NoError(t, err)
	require.True(t, sufficient)
	require.InDelta(t, 7.5, newBalance, 0.000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeductUsageBillingBalance_RecordsDebtOnOverdraft(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	expectUsageDebit(mock, 42, 7, "request-overdraft", "0", "5", "0", "-10", "0", "0", "5", "-5")
	mock.ExpectCommit()

	newBalance, sufficient, err := deductUsageBillingBalance(ctx, tx, 42, 10, 7, "request-overdraft")
	require.NoError(t, err)
	require.False(t, sufficient)
	require.InDelta(t, -5, newBalance, 0.000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeductUsageBillingBalance_ReturnsUserNotFound(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectExec(`INSERT INTO user_credit_accounts`).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT 1 FROM user_credit_accounts`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectRollback()

	_, _, err = deductUsageBillingBalance(ctx, tx, 42, 10, 7, "request-missing-user")
	require.ErrorIs(t, err, service.ErrUserNotFound)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCaptureUsageBillingBatchImageBalance_RejectsActualCostOverHold(t *testing.T) {
	_, err := captureUsageBillingBatchImageBalance(context.Background(), &sql.Tx{}, &service.BatchImageBalanceHoldCommand{
		UserID: 42, HoldAmount: 1, ActualAmount: 1.1,
	})
	require.ErrorIs(t, err, service.ErrBatchImageSettlementCostExceedsHold)
}
