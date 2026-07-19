package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestReverseDistributionCommissionCreatesDebtForWithdrawnFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	mock.ExpectExec(`INSERT INTO distribution_cash_wallets`).
		WithArgs(int64(7), int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT available_cny_minor, frozen_cny_minor, debt_cny_minor`).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"available_cny_minor", "frozen_cny_minor", "debt_cny_minor"}).AddRow(100, 0, 0))
	mock.ExpectExec(`UPDATE distribution_cash_wallets`).
		WithArgs(int64(7), int64(42), int64(0), int64(0), int64(200), int64(300)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE distribution_commissions`).
		WithArgs(int64(91), "provider chargeback", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO distribution_wallet_ledger`).
		WithArgs(int64(7), int64(42), "commission_reversal", int64(-300), "distribution_commission", "91", "distribution:7:commission:91:reversal").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = reverseDistributionCommissionTx(context.Background(), tx, 7, 1, 91, 42, 300, "AVAILABLE", "provider chargeback")
	require.NoError(t, err)
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReverseRechargeRejectsIncompleteSecurityContext(t *testing.T) {
	service := &DistributionService{}
	_, err := service.ReverseRecharge(context.Background(), 7, 1, "CHARGEBACK", "")
	require.ErrorIs(t, err, ErrReversalInvalid)
}

func TestReverseLegacyAffiliateRebateRemovesFrozenAccrual(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	mock.ExpectQuery(`FROM user_affiliate_ledger`).
		WithArgs(int64(8122)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "source_user_id", "amount", "frozen_until"}).AddRow(33, 42, 77, "5", time.Now().Add(time.Hour)))
	mock.ExpectQuery(`SELECT aff_quota::text, aff_frozen_quota::text, aff_history_quota::text`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"aff_quota", "aff_frozen_quota", "aff_history_quota"}).AddRow("1", "5", "10"))
	mock.ExpectExec(`UPDATE user_affiliates`).
		WithArgs(int64(42), "1", "0", "5").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO user_affiliate_ledger`).
		WithArgs(int64(42), "-5", int64(77), int64(8122), "1", "0", "5").
		WillReturnResult(sqlmock.NewResult(44, 1))
	mock.ExpectExec(`UPDATE user_affiliate_ledger SET reversed_at`).
		WithArgs(int64(33), "provider chargeback", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	total, err := reverseLegacyAffiliateRebateTx(context.Background(), tx, 8122, 1, "provider chargeback")
	require.NoError(t, err)
	require.Equal(t, "5", total.String())
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
