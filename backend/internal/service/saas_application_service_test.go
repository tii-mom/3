package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type saasApplicationSettingRepoStub struct {
	values           map[string]string
	setMultipleCalls int
}

func (s *saasApplicationSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *saasApplicationSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *saasApplicationSettingRepoStub) Set(_ context.Context, key, value string) error {
	s.values[key] = value
	return nil
}

func (s *saasApplicationSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (s *saasApplicationSettingRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	s.setMultipleCalls++
	for key, value := range values {
		s.values[key] = value
	}
	return nil
}

func TestUpdateSaaSFeatureFlagsUsesOneBulkWrite(t *testing.T) {
	settings := &saasApplicationSettingRepoStub{values: map[string]string{}}
	controlPlane := true
	applications := true
	require.NoError(t, NewSaaSService(nil, settings, nil, nil).UpdateFeatureFlags(context.Background(), &controlPlane, &applications))
	require.Equal(t, 1, settings.setMultipleCalls)
	require.Equal(t, "true", settings.values["saas_control_plane_enabled"])
	require.Equal(t, "true", settings.values["saas_application_enabled"])
}

func (s *saasApplicationSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *saasApplicationSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

type saasApplicationEncryptor struct{}

func (saasApplicationEncryptor) Encrypt(value string) (string, error) {
	return "encrypted:" + value, nil
}
func (saasApplicationEncryptor) Decrypt(value string) (string, error) { return value, nil }

func TestSubmitSaaSApplicationDoesNotProvisionTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &saasApplicationSettingRepoStub{values: map[string]string{"saas_application_enabled": "true"}}
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenant_applications WHERE user_id = $1 AND status IN ('SUBMITTED', 'CONTACTED'))`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`INSERT INTO saas_tenant_applications`).
		WithArgs(int64(42), "Acme AI", "Alice", "email", "alice@example.com", "api.acme.test", "2500", 300, "Reseller network", "REF123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "brand_name", "contact_name", "contact_channel", "contact_value",
			"desired_domain", "expected_monthly_usd", "expected_users", "business_description",
			"referral_code", "status", "tenant_id", "reviewer_user_id", "review_note",
			"submitted_at", "reviewed_at", "created_at", "updated_at",
		}).AddRow(7, 42, "Acme AI", "Alice", "email", "alice@example.com", "api.acme.test", "2500", 300, "Reseller network", "REF123", "SUBMITTED", nil, nil, "", now, nil, now, now))
	mock.ExpectExec(`INSERT INTO saas_tenant_application_events`).WithArgs(int64(7), int64(42)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	item, err := NewSaaSService(db, settings, nil, nil).SubmitApplication(context.Background(), 42, SubmitSaaSApplicationInput{
		BrandName: " Acme AI ", ContactName: " Alice ", ContactChannel: "EMAIL",
		ContactValue: " alice@example.com ", DesiredDomain: "API.ACME.TEST.",
		ExpectedMonthlyUSD: "2500", ExpectedUsers: 300,
		BusinessDescription: " Reseller network ", ReferralCode: "ref123",
	})
	require.NoError(t, err)
	require.Equal(t, int64(7), item.ID)
	require.Equal(t, "SUBMITTED", item.Status)
	require.Nil(t, item.TenantID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaaSApplicationOverviewReturnsEmptyState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &saasApplicationSettingRepoStub{values: map[string]string{"saas_application_enabled": "true"}}
	mock.ExpectQuery(`FROM saas_tenant_applications WHERE user_id = \$1`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`FROM saas_tenants t`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	overview, err := NewSaaSService(db, settings, nil, nil).ApplicationOverview(context.Background(), 42)
	require.NoError(t, err)
	require.True(t, overview.ApplicationsEnabled)
	require.Nil(t, overview.Application)
	require.Nil(t, overview.Tenant)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApproveSaaSApplicationCreatesTenantAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &saasApplicationSettingRepoStub{values: map[string]string{"saas_control_plane_enabled": "true"}}
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT user_id, brand_name, referral_code, status`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "brand_name", "referral_code", "status"}).AddRow(42, "Acme AI", "", "CONTACTED"))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`)).WithArgs("acme-ai").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE slug = $1)`)).
		WithArgs("acme-ai").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`INSERT INTO saas_tenants`).
		WithArgs("acme-ai", "Acme AI", "Acme API", "https://cdn.example/logo.png", int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name", "status", "site_name", "site_logo", "primary_domain", "core_user_id", "created_at"}).
			AddRow(9, "acme-ai", "Acme AI", "active", "Acme API", "https://cdn.example/logo.png", "", 42, now))
	mock.ExpectExec(`INSERT INTO saas_wholesale_wallets`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO saas_tenant_configs`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO api_keys`).WithArgs(int64(42), sqlmock.AnyArg(), "Wholesale / Acme AI", int64(9)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO saas_provisioning_jobs`).WithArgs(int64(9), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE saas_tenant_applications`).WithArgs(int64(7), int64(9), int64(1), "approved pilot").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO saas_tenant_application_events`).WithArgs(int64(7), "CONTACTED", int64(1), "approved pilot").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM saas_tenant_applications WHERE id = \$1`).WithArgs(int64(7)).
		WillReturnRows(applicationRows().AddRow(7, 42, "Acme AI", "Alice", "email", "alice@example.com", "api.acme.test", "2500", 300, "Reseller network", "", "APPROVED", 9, 1, "approved pilot", now, now, now, now))
	mock.ExpectCommit()

	result, err := NewSaaSService(db, settings, nil, saasApplicationEncryptor{}).ApproveApplication(context.Background(), 7, 1, ApproveSaaSApplicationInput{
		Slug: "acme-ai", SiteName: "Acme API", SiteLogo: "https://cdn.example/logo.png", ReviewNote: "approved pilot",
	})
	require.NoError(t, err)
	require.Equal(t, int64(9), result.Tenant.ID)
	require.Equal(t, "APPROVED", result.Application.Status)
	require.Contains(t, result.WholesaleKey, "sk-wholesale-")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApproveSaaSApplicationRejectsExistingTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &saasApplicationSettingRepoStub{values: map[string]string{"saas_control_plane_enabled": "true"}}
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT user_id, brand_name, referral_code, status`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "brand_name", "referral_code", "status"}).AddRow(42, "Acme AI", "", "SUBMITTED"))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`)).WithArgs("acme-ai").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	_, err = NewSaaSService(db, settings, nil, saasApplicationEncryptor{}).ApproveApplication(context.Background(), 7, 1, ApproveSaaSApplicationInput{Slug: "acme-ai"})
	require.True(t, errors.Is(err, ErrSaaSAlreadyTenant))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApproveSaaSApplicationReturnsSlugConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &saasApplicationSettingRepoStub{values: map[string]string{"saas_control_plane_enabled": "true"}}
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT user_id, brand_name, referral_code, status`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "brand_name", "referral_code", "status"}).AddRow(42, "Acme AI", "", "SUBMITTED"))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`)).WithArgs("acme-ai").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`)).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE slug = $1)`)).
		WithArgs("acme-ai").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	_, err = NewSaaSService(db, settings, nil, saasApplicationEncryptor{}).ApproveApplication(context.Background(), 7, 1, ApproveSaaSApplicationInput{Slug: "acme-ai"})
	require.True(t, errors.Is(err, ErrSaaSTenantSlugExists))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaaSApplicationTransitionsAreForwardOnly(t *testing.T) {
	require.True(t, allowedSaaSApplicationTransition("SUBMITTED", "CONTACTED"))
	require.True(t, allowedSaaSApplicationTransition("CONTACTED", "APPROVED"))
	require.False(t, allowedSaaSApplicationTransition("REJECTED", "SUBMITTED"))
	require.False(t, allowedSaaSApplicationTransition("APPROVED", "REJECTED"))
}

func applicationRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user_id", "brand_name", "contact_name", "contact_channel", "contact_value",
		"desired_domain", "expected_monthly_usd", "expected_users", "business_description",
		"referral_code", "status", "tenant_id", "reviewer_user_id", "review_note",
		"submitted_at", "reviewed_at", "created_at", "updated_at",
	})
}
