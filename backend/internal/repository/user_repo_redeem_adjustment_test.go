package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func newRedeemAdjustmentRepoMock(t *testing.T) (*userRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })
	return newUserRepositoryWithSQL(client, db), mock
}

func TestApplyRedeemBalanceAdjustment_UsesAtomicFloor(t *testing.T) {
	repo, mock := newRedeemAdjustmentRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO user_credit_accounts`).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT transferable_credit::text, non_transferable_credit::text, debt::text FROM user_credit_accounts`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"transferable_credit", "non_transferable_credit", "debt"}).AddRow("10", "5", "0"))
	mock.ExpectQuery(`SELECT value FROM settings WHERE key = 'credit_bucket_enforce_enabled'`).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("false"))
	mock.ExpectExec(`UPDATE user_credit_accounts`).
		WithArgs(int64(42), "8", "0", "0").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`UPDATE users`).
		WithArgs(int64(42), "8", "0", false, true, "-7").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow("8"))
	mock.ExpectExec(`INSERT INTO user_credit_ledger`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.ApplyRedeemBalanceAdjustment(context.Background(), 42, -7))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyRedeemConcurrencyAdjustment_UsesAtomicFloor(t *testing.T) {
	repo, mock := newRedeemAdjustmentRepoMock(t)
	mock.ExpectExec(`UPDATE users SET concurrency = GREATEST\(concurrency \+ \$1, 0\), updated_at = NOW\(\) WHERE id = \$2 AND deleted_at IS NULL`).
		WithArgs(-7, int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.ApplyRedeemConcurrencyAdjustment(context.Background(), 42, -7))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyRedeemAdjustment_MissingUser(t *testing.T) {
	repo, mock := newRedeemAdjustmentRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO user_credit_accounts`).
		WithArgs(int64(404)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT 1 FROM user_credit_accounts`).
		WithArgs(int64(404)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectRollback()

	err := repo.ApplyRedeemBalanceAdjustment(context.Background(), 404, -1)
	require.ErrorIs(t, err, service.ErrUserNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
