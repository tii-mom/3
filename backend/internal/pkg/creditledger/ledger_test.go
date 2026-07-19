package creditledger

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestBucketEnforcementDefaultsToShadowAndReadsExplicitMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(`SELECT value FROM settings WHERE key = 'credit_bucket_enforce_enabled'`).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("false"))
	enforced, err := bucketEnforcementEnabled(context.Background(), db)
	require.NoError(t, err)
	require.False(t, enforced)

	mock.ExpectQuery(`SELECT value FROM settings WHERE key = 'credit_bucket_enforce_enabled'`).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TRUE"))
	enforced, err = bucketEnforcementEnabled(context.Background(), db)
	require.NoError(t, err)
	require.True(t, enforced)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyPrincipalReversalDebitsTransferableFirstAndDecrementsRecharge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(`INSERT INTO user_credit_accounts`).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT 1 FROM user_credit_accounts`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectQuery(`SELECT transferable_credit::text, non_transferable_credit::text, debt::text`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"transferable_credit", "non_transferable_credit", "debt"}).AddRow("10", "5", "0"))
	mock.ExpectQuery(`SELECT value FROM settings`).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("false"))
	mock.ExpectExec(`UPDATE user_credit_accounts`).WithArgs(int64(42), "0", "3", "0").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`UPDATE users`).WithArgs(int64(42), "3", "-12", false, false, "-12").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow("3"))
	mock.ExpectExec(`INSERT INTO user_credit_ledger`).WillReturnResult(sqlmock.NewResult(1, 1))

	account, applied, err := Apply(context.Background(), db, 42, decimal.NewFromInt(-12), creditctx.Metadata{
		EntryType: "recharge_principal_reversal", SourceType: "distribution_reversal",
		CountRecharge: true, DebitTransferableFirst: true,
	}, false)
	require.NoError(t, err)
	require.True(t, applied)
	require.True(t, account.Transferable.IsZero())
	require.Equal(t, "3", account.NonTransferable.String())
	require.True(t, account.Debt.IsZero())
	require.NoError(t, mock.ExpectationsWereMet())
}
