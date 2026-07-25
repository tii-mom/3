package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestInsertVoucherRiskLedgerCastsReasonParameter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(`(?s)INSERT INTO balance_voucher_ledger.*jsonb_build_object\('reason', \$7::text\)`).
		WithArgs(int64(7), voucherTenantID, int64(42), "RISK_LOCKED", "10", "1", "manual review").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = insertVoucherRiskLedger(context.Background(), db, 7, 42, "RISK_LOCKED", "10", "1", "manual review")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
