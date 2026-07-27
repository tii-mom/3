package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestValidateTargetRejectsNonLocalByDefault(t *testing.T) {
	err := validateTarget("postgres://user:pass@db.example.com:5432/app?sslmode=require", false)
	if err == nil {
		t.Fatal("expected non-local target to be rejected")
	}
}

func TestValidateTargetAllowsLoopback(t *testing.T) {
	for _, dsn := range []string{
		"postgres://user:pass@127.0.0.1:5432/app?sslmode=disable",
		"postgresql://user:pass@localhost:5432/app?sslmode=disable",
		"postgres://user:pass@[::1]:5432/app?sslmode=disable",
	} {
		if err := validateTarget(dsn, false); err != nil {
			t.Fatalf("validateTarget(%q): %v", dsn, err)
		}
	}
}

func TestValidateTargetRequiresExplicitDatabase(t *testing.T) {
	if err := validateTarget("postgres://user:pass@127.0.0.1:5432", false); err == nil {
		t.Fatal("expected missing database name to be rejected")
	}
}

func TestValidateTargetRequiresPostgresURL(t *testing.T) {
	if err := validateTarget("host=127.0.0.1 dbname=app", false); err == nil {
		t.Fatal("expected keyword DSN to be rejected")
	}
}

func TestVerifyFeatureStateAllowsPostRolloutEnabledFeatures(t *testing.T) {
	db, mock := newFeatureStateMock(t)
	defer func() { _ = db.Close() }()

	expectFeatureSetting(mock, "credit_bucket_enforce_enabled", "true")
	expectFeatureSetting(mock, "balance_voucher_enabled", "true")
	expectFeatureSetting(mock, "distribution_enabled", "true")
	expectFeatureSetting(mock, "saas_control_plane_enabled", "false")
	expectComputeCompanyState(mock, true, false)

	require.NoError(t, verifyFeatureState(context.Background(), db))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyFeatureStateRejectsFinancialFeatureWithoutCreditBuckets(t *testing.T) {
	db, mock := newFeatureStateMock(t)
	defer func() { _ = db.Close() }()

	expectFeatureSetting(mock, "credit_bucket_enforce_enabled", "false")
	expectFeatureSetting(mock, "balance_voucher_enabled", "true")
	expectFeatureSetting(mock, "distribution_enabled", "false")
	expectFeatureSetting(mock, "saas_control_plane_enabled", "false")

	err := verifyFeatureState(context.Background(), db)
	require.ErrorContains(t, err, "credit buckets must be enforced")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyFeatureStateRejectsLegacyStacking(t *testing.T) {
	db, mock := newFeatureStateMock(t)
	defer func() { _ = db.Close() }()

	expectFeatureSetting(mock, "credit_bucket_enforce_enabled", "true")
	expectFeatureSetting(mock, "balance_voucher_enabled", "true")
	expectFeatureSetting(mock, "distribution_enabled", "true")
	expectFeatureSetting(mock, "saas_control_plane_enabled", "false")
	expectComputeCompanyState(mock, true, true)

	err := verifyFeatureState(context.Background(), db)
	require.ErrorContains(t, err, "must not stack")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyFeatureStateRejectsDistributionSettingMismatch(t *testing.T) {
	db, mock := newFeatureStateMock(t)
	defer func() { _ = db.Close() }()

	expectFeatureSetting(mock, "credit_bucket_enforce_enabled", "true")
	expectFeatureSetting(mock, "balance_voucher_enabled", "true")
	expectFeatureSetting(mock, "distribution_enabled", "true")
	expectFeatureSetting(mock, "saas_control_plane_enabled", "false")
	expectComputeCompanyState(mock, false, false)

	err := verifyFeatureState(context.Background(), db)
	require.ErrorContains(t, err, "disagree")
	require.NoError(t, mock.ExpectationsWereMet())
}

func newFeatureStateMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func expectFeatureSetting(mock sqlmock.Sqlmock, key, value string) {
	mock.ExpectQuery(`SELECT value FROM settings WHERE key = \$1`).
		WithArgs(key).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(value))
}

func expectComputeCompanyState(mock sqlmock.Sqlmock, enabled, stack bool) {
	mock.ExpectQuery(`SELECT enabled, stack_with_legacy FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).
		WillReturnRows(sqlmock.NewRows([]string{"enabled", "stack_with_legacy"}).AddRow(enabled, stack))
}
