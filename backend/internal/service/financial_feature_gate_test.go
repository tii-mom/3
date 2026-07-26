package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureCreditBucketsReady(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM users u
LEFT JOIN user_credit_accounts a ON a.user_id = u.id
WHERE a.user_id IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM financial_reconciliation_issues i
LEFT JOIN users u ON u.id = i.user_id
LEFT JOIN user_credit_accounts a ON a.user_id = i.user_id
WHERE i.status = 'OPEN'
  AND (
      u.id IS NULL
      OR a.user_id IS NULL
      OR u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt
  )`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	if err := ensureCreditBucketsReady(context.Background(), db); err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureCreditBucketsReadyRejectsBlockingOpenIssues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM users u
LEFT JOIN user_credit_accounts a ON a.user_id = u.id
WHERE a.user_id IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM financial_reconciliation_issues i
LEFT JOIN users u ON u.id = i.user_id
LEFT JOIN user_credit_accounts a ON a.user_id = i.user_id
WHERE i.status = 'OPEN'
  AND (
      u.id IS NULL
      OR a.user_id IS NULL
      OR u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt
  )`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if err := ensureCreditBucketsReady(context.Background(), db); err != ErrCreditBucketsNotEnforced {
		t.Fatalf("expected credit bucket guard, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureCreditBucketsReadyRejectsMissingAccounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM users u
LEFT JOIN user_credit_accounts a ON a.user_id = u.id
WHERE a.user_id IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if err := ensureCreditBucketsReady(context.Background(), db); err != ErrCreditBucketsNotEnforced {
		t.Fatalf("expected credit bucket guard, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
