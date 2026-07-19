package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrSaaSDisabled          = infraerrors.Forbidden("SAAS_CONTROL_PLANE_DISABLED", "SaaS control plane is disabled")
	ErrSaaSTenantNotFound    = infraerrors.NotFound("SAAS_TENANT_NOT_FOUND", "SaaS tenant not found")
	ErrSaaSDomainInvalid     = infraerrors.BadRequest("SAAS_DOMAIN_INVALID", "invalid tenant domain")
	ErrSaaSDomainNotVerified = infraerrors.Conflict("SAAS_DOMAIN_NOT_VERIFIED", "domain ownership verification failed")
	ErrSaaSCoreUserRequired  = infraerrors.BadRequest("SAAS_CORE_USER_REQUIRED", "core user is required for wholesale access")
	ErrSaaSTenantSlugExists  = infraerrors.Conflict("SAAS_TENANT_SLUG_EXISTS", "SaaS tenant slug is already in use")
)

type SaaSService struct {
	db        *sql.DB
	settings  SettingRepository
	authCache APIKeyAuthCacheInvalidator
	encryptor SecretEncryptor
}

type SaaSTenant struct {
	ID            int64     `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	SiteName      string    `json:"site_name"`
	SiteLogo      string    `json:"site_logo"`
	PrimaryDomain string    `json:"primary_domain,omitempty"`
	CoreUserID    *int64    `json:"core_user_id,omitempty"`
	WholesaleUSD  string    `json:"wholesale_balance_usd"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateSaaSTenantInput struct {
	Slug         string
	Name         string
	SiteName     string
	SiteLogo     string
	CoreUserID   int64
	ReferralCode string
}

type CreateSaaSTenantResult struct {
	Tenant       SaaSTenant `json:"tenant"`
	WholesaleKey string     `json:"wholesale_api_key"`
}

type SaaSDomain struct {
	ID                int64      `json:"id"`
	TenantID          int64      `json:"tenant_id"`
	Domain            string     `json:"domain"`
	VerificationToken string     `json:"verification_token"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	TLSStatus         string     `json:"tls_status"`
	Status            string     `json:"status"`
}

type SaaSPlan struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	BillingPeriod string    `json:"billing_period"`
	PriceMinor    int64     `json:"price_cny_minor"`
	Enabled       bool      `json:"enabled"`
	Limits        string    `json:"limits"`
	CreatedAt     time.Time `json:"created_at"`
}

type SaaSSubscription struct {
	ID               int64     `json:"id"`
	TenantID         int64     `json:"tenant_id"`
	PlanID           int64     `json:"plan_id"`
	Status           string    `json:"status"`
	PaidMinor        int64     `json:"paid_cny_minor"`
	PaymentReference string    `json:"payment_reference"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type SaaSResourceAllocation struct {
	ID               int64  `json:"id"`
	TenantID         int64  `json:"tenant_id"`
	GroupID          int64  `json:"group_id"`
	AllocationType   string `json:"allocation_type"`
	ConcurrencyLimit int    `json:"concurrency_limit"`
	MonthlyLimitUSD  string `json:"monthly_limit_usd"`
}

type ProvisioningJob struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PartnerDashboard struct {
	AvailableMinor   int64 `json:"available_cny_minor"`
	FrozenMinor      int64 `json:"frozen_cny_minor"`
	WithdrawingMinor int64 `json:"withdrawing_cny_minor"`
	LifetimeMinor    int64 `json:"lifetime_earned_cny_minor"`
}

type TenantControl struct {
	Tenant            SaaSTenant `json:"tenant"`
	RetailMultiplier  string     `json:"retail_multiplier"`
	PaymentProvider   string     `json:"payment_provider"`
	PaymentConfigured bool       `json:"payment_configured"`
	InstanceConfig    string     `json:"instance_config"`
}

func NewSaaSService(db *sql.DB, settings SettingRepository, authCache APIKeyAuthCacheInvalidator, encryptor SecretEncryptor) *SaaSService {
	return &SaaSService{db: db, settings: settings, authCache: authCache, encryptor: encryptor}
}

func (s *SaaSService) Enabled(ctx context.Context) (bool, error) {
	value, err := s.settings.GetValue(ctx, "saas_control_plane_enabled")
	return value == "true", err
}

func (s *SaaSService) UpdateEnabled(ctx context.Context, enabled bool) error {
	if err := s.settings.Set(ctx, "saas_control_plane_enabled", strconv.FormatBool(enabled)); err != nil {
		return err
	}
	s.invalidateWholesaleAuthCache(ctx, 0)
	return nil
}

func (s *SaaSService) CreateTenant(ctx context.Context, input CreateSaaSTenantInput) (*CreateSaaSTenantResult, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Name = strings.TrimSpace(input.Name)
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`).MatchString(input.Slug) || input.Name == "" {
		return nil, infraerrors.BadRequest("SAAS_TENANT_INVALID", "invalid tenant slug or name")
	}
	if input.CoreUserID <= 0 {
		return nil, ErrSaaSCoreUserRequired
	}
	prepared, err := s.prepareSaaSTenantProvisioning(input.Slug)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, input.CoreUserID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, input.Slug); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`, input.CoreUserID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSaaSCoreUserRequired
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`, input.CoreUserID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSaaSAlreadyTenant
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE slug = $1)`, input.Slug).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSaaSTenantSlugExists
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenant_applications WHERE user_id = $1)`, input.CoreUserID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSaaSApplicationApprovalRequired
	}
	if input.SiteName == "" {
		input.SiteName = input.Name
	}
	tenant, err := s.provisionTenantTx(ctx, tx, input, prepared)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &CreateSaaSTenantResult{Tenant: *tenant, WholesaleKey: prepared.apiKey}, nil
}

func (s *SaaSService) TenantControl(ctx context.Context, ownerUserID int64) (*TenantControl, error) {
	var control TenantControl
	var paymentConfig sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT t.id, t.slug, t.name, t.status, t.site_name, t.site_logo, COALESCE(t.primary_domain, ''), t.core_user_id, COALESCE(w.balance_usd, 0)::text, t.created_at, c.retail_multiplier::text, c.payment_provider, c.payment_config_encrypted, c.instance_config::text FROM saas_tenants t JOIN saas_tenant_configs c ON c.tenant_id = t.id LEFT JOIN saas_wholesale_wallets w ON w.tenant_id = t.id WHERE t.core_user_id = $1 AND t.id <> 1`, ownerUserID).Scan(&control.Tenant.ID, &control.Tenant.Slug, &control.Tenant.Name, &control.Tenant.Status, &control.Tenant.SiteName, &control.Tenant.SiteLogo, &control.Tenant.PrimaryDomain, &control.Tenant.CoreUserID, &control.Tenant.WholesaleUSD, &control.Tenant.CreatedAt, &control.RetailMultiplier, &control.PaymentProvider, &paymentConfig, &control.InstanceConfig)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	control.PaymentConfigured = paymentConfig.Valid && paymentConfig.String != ""
	return &control, nil
}

func (s *SaaSService) UpdateTenantControl(ctx context.Context, ownerUserID int64, siteName, siteLogo, retailMultiplierRaw, paymentProvider, paymentConfig, instanceConfig string) (*TenantControl, error) {
	retailMultiplier, err := decimal.NewFromString(strings.TrimSpace(retailMultiplierRaw))
	if err != nil || !retailMultiplier.IsPositive() || retailMultiplier.GreaterThan(decimal.NewFromInt(100)) {
		return nil, infraerrors.BadRequest("RETAIL_MULTIPLIER_INVALID", "invalid retail multiplier")
	}
	if strings.TrimSpace(instanceConfig) == "" {
		instanceConfig = "{}"
	}
	var encryptedConfig any
	if strings.TrimSpace(paymentConfig) != "" {
		if s.encryptor == nil {
			return nil, infraerrors.ServiceUnavailable("PAYMENT_ENCRYPTION_UNAVAILABLE", "payment encryption unavailable")
		}
		encrypted, err := s.encryptor.Encrypt(paymentConfig)
		if err != nil {
			return nil, err
		}
		encryptedConfig = encrypted
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var tenantID int64
	if err := tx.QueryRowContext(ctx, `UPDATE saas_tenants SET site_name = $2, site_logo = $3, updated_at = NOW() WHERE core_user_id = $1 AND id <> 1 RETURNING id`, ownerUserID, strings.TrimSpace(siteName), strings.TrimSpace(siteLogo)).Scan(&tenantID); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSTenantNotFound
	} else if err != nil {
		return nil, err
	}
	if encryptedConfig == nil {
		_, err = tx.ExecContext(ctx, `UPDATE saas_tenant_configs SET retail_multiplier = $2, payment_provider = $3, instance_config = $4::jsonb, updated_at = NOW() WHERE tenant_id = $1`, tenantID, retailMultiplier.String(), strings.TrimSpace(paymentProvider), instanceConfig)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE saas_tenant_configs SET retail_multiplier = $2, payment_provider = $3, payment_config_encrypted = $4, instance_config = $5::jsonb, updated_at = NOW() WHERE tenant_id = $1`, tenantID, retailMultiplier.String(), strings.TrimSpace(paymentProvider), encryptedConfig, instanceConfig)
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.TenantControl(ctx, ownerUserID)
}

func (s *SaaSService) ListTenants(ctx context.Context, page, pageSize int) ([]SaaSTenant, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM saas_tenants`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.slug, t.name, t.status, t.site_name, t.site_logo, COALESCE(t.primary_domain, ''), t.core_user_id, COALESCE(w.balance_usd, 0)::text, t.created_at FROM saas_tenants t LEFT JOIN saas_wholesale_wallets w ON w.tenant_id = t.id ORDER BY t.id DESC LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSTenant, 0, pageSize)
	for rows.Next() {
		var item SaaSTenant
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name, &item.Status, &item.SiteName, &item.SiteLogo, &item.PrimaryDomain, &item.CoreUserID, &item.WholesaleUSD, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *SaaSService) FundWholesaleWallet(ctx context.Context, tenantID int64, amountRaw, reference string) (string, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(amountRaw))
	if err != nil || !amount.IsPositive() || amount.Exponent() < -8 {
		return "", infraerrors.BadRequest("WHOLESALE_AMOUNT_INVALID", "invalid wholesale funding amount")
	}
	if strings.TrimSpace(reference) == "" {
		return "", infraerrors.BadRequest("WHOLESALE_REFERENCE_REQUIRED", "funding reference is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	var balance string
	err = tx.QueryRowContext(ctx, `INSERT INTO saas_wholesale_wallets (tenant_id, balance_usd, lifetime_funded_usd) VALUES ($1, $2, $2) ON CONFLICT (tenant_id) DO UPDATE SET balance_usd = saas_wholesale_wallets.balance_usd + EXCLUDED.balance_usd, lifetime_funded_usd = saas_wholesale_wallets.lifetime_funded_usd + EXCLUDED.lifetime_funded_usd, updated_at = NOW() RETURNING balance_usd::text`, tenantID, amount.String()).Scan(&balance)
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO saas_wholesale_ledger (tenant_id, action, amount_usd, balance_after, source_type, source_id, idempotency_key) VALUES ($1, 'fund', $2, $3, 'admin_funding', $4, $5)`, tenantID, amount.String(), balance, reference, "wholesale-fund:"+reference); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	s.invalidateWholesaleAuthCache(ctx, tenantID)
	return balance, nil
}

func (s *SaaSService) invalidateWholesaleAuthCache(ctx context.Context, tenantID int64) {
	if s.authCache == nil || s.db == nil {
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key FROM api_keys WHERE ($1 = 0 OR tenant_id = $1) AND key_type = 'tenant_wholesale' AND deleted_at IS NULL`, tenantID)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var key string
		if rows.Scan(&key) == nil {
			s.authCache.InvalidateAuthCacheByKey(ctx, key)
		}
	}
}

func (s *SaaSService) AddDomain(ctx context.Context, tenantID int64, domain string) (*SaaSDomain, error) {
	domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
	if !validTenantDomain(domain) {
		return nil, ErrSaaSDomainInvalid
	}
	token, err := randomPrefixedSecret("3api-domain-", 18)
	if err != nil {
		return nil, err
	}
	var item SaaSDomain
	err = s.db.QueryRowContext(ctx, `INSERT INTO saas_tenant_domains (tenant_id, domain, verification_token) VALUES ($1, $2, $3) RETURNING id, tenant_id, domain, verification_token, verified_at, tls_status, status`, tenantID, domain, token).Scan(&item.ID, &item.TenantID, &item.Domain, &item.VerificationToken, &item.VerifiedAt, &item.TLSStatus, &item.Status)
	return &item, err
}

func (s *SaaSService) VerifyDomain(ctx context.Context, domainID int64) (*SaaSDomain, error) {
	var item SaaSDomain
	err := s.db.QueryRowContext(ctx, `SELECT id, tenant_id, domain, verification_token, verified_at, tls_status, status FROM saas_tenant_domains WHERE id = $1`, domainID).Scan(&item.ID, &item.TenantID, &item.Domain, &item.VerificationToken, &item.VerifiedAt, &item.TLSStatus, &item.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	records, err := net.LookupTXT("_3api-verification." + item.Domain)
	if err != nil {
		return nil, ErrSaaSDomainNotVerified
	}
	verified := false
	for _, record := range records {
		if strings.TrimSpace(record) == item.VerificationToken {
			verified = true
			break
		}
	}
	if !verified {
		return nil, ErrSaaSDomainNotVerified
	}
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE saas_tenant_domains SET verified_at = $2, status = 'verified', tls_status = 'pending', updated_at = NOW() WHERE id = $1`, item.ID, now); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE saas_tenants SET primary_domain = $2, updated_at = NOW() WHERE id = $1`, item.TenantID, item.Domain); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	item.VerifiedAt, item.Status = &now, "verified"
	return &item, nil
}

func (s *SaaSService) CreatePlan(ctx context.Context, name, period string, priceMinor int64, limits string) (*SaaSPlan, error) {
	name, period = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(period))
	if name == "" || (period != "month" && period != "year") || priceMinor < 0 {
		return nil, infraerrors.BadRequest("SAAS_PLAN_INVALID", "invalid SaaS plan")
	}
	if strings.TrimSpace(limits) == "" {
		limits = "{}"
	}
	var item SaaSPlan
	err := s.db.QueryRowContext(ctx, `INSERT INTO saas_plans (name, billing_period, price_cny_minor, limits) VALUES ($1, $2, $3, $4::jsonb) RETURNING id, name, billing_period, price_cny_minor, enabled, limits::text, created_at`, name, period, priceMinor, limits).Scan(&item.ID, &item.Name, &item.BillingPeriod, &item.PriceMinor, &item.Enabled, &item.Limits, &item.CreatedAt)
	return &item, err
}

func (s *SaaSService) ListPlans(ctx context.Context) ([]SaaSPlan, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, billing_period, price_cny_minor, enabled, limits::text, created_at FROM saas_plans ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSPlan, 0)
	for rows.Next() {
		var item SaaSPlan
		if err := rows.Scan(&item.ID, &item.Name, &item.BillingPeriod, &item.PriceMinor, &item.Enabled, &item.Limits, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) ListSubscriptions(ctx context.Context, tenantID int64) ([]SaaSSubscription, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, tenant_id, plan_id, status, paid_cny_minor, COALESCE(payment_reference, ''), expires_at, created_at FROM saas_tenant_subscriptions WHERE ($1 = 0 OR tenant_id = $1) ORDER BY id DESC LIMIT 500`, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSSubscription, 0)
	for rows.Next() {
		var item SaaSSubscription
		if err := rows.Scan(&item.ID, &item.TenantID, &item.PlanID, &item.Status, &item.PaidMinor, &item.PaymentReference, &item.ExpiresAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) ListDomains(ctx context.Context, tenantID int64) ([]SaaSDomain, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, tenant_id, domain, verification_token, verified_at, tls_status, status FROM saas_tenant_domains WHERE ($1 = 0 OR tenant_id = $1) ORDER BY id DESC LIMIT 500`, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSDomain, 0)
	for rows.Next() {
		var item SaaSDomain
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Domain, &item.VerificationToken, &item.VerifiedAt, &item.TLSStatus, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) ListResourceAllocations(ctx context.Context, tenantID int64) ([]SaaSResourceAllocation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, tenant_id, group_id, allocation_type, concurrency_limit, COALESCE(monthly_limit_usd, 0)::text FROM saas_resource_pool_allocations WHERE ($1 = 0 OR tenant_id = $1) ORDER BY id DESC LIMIT 500`, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSResourceAllocation, 0)
	for rows.Next() {
		var item SaaSResourceAllocation
		if err := rows.Scan(&item.ID, &item.TenantID, &item.GroupID, &item.AllocationType, &item.ConcurrencyLimit, &item.MonthlyLimitUSD); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) RecordPaidSubscription(ctx context.Context, tenantID, planID, paidMinor int64, reference string) (int64, error) {
	if paidMinor <= 0 || strings.TrimSpace(reference) == "" {
		return 0, infraerrors.BadRequest("SAAS_SUBSCRIPTION_PAYMENT_INVALID", "invalid subscription payment")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var period string
	if err := tx.QueryRowContext(ctx, `SELECT billing_period FROM saas_plans WHERE id = $1 AND enabled = TRUE`, planID).Scan(&period); errors.Is(err, sql.ErrNoRows) {
		return 0, infraerrors.NotFound("SAAS_PLAN_NOT_FOUND", "SaaS plan not found")
	} else if err != nil {
		return 0, err
	}
	expires := time.Now().UTC().AddDate(0, 1, 0)
	if period == "year" {
		expires = time.Now().UTC().AddDate(1, 0, 0)
	}
	var subscriptionID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO saas_tenant_subscriptions (tenant_id, plan_id, status, paid_cny_minor, payment_reference, expires_at) VALUES ($1, $2, 'active', $3, $4, $5) RETURNING id`, tenantID, planID, paidMinor, reference, expires).Scan(&subscriptionID)
	if err != nil {
		return 0, err
	}
	var referralID, beneficiaryID int64
	var rate int64
	err = tx.QueryRowContext(ctx, `SELECT id, referrer_user_id, commission_rate_bps FROM saas_partner_referrals WHERE referred_tenant_id = $1 AND valid_until >= NOW()`, tenantID).Scan(&referralID, &beneficiaryID, &rate)
	if err == nil {
		amount := decimal.NewFromInt(paidMinor).Mul(decimal.NewFromInt(rate)).Div(decimal.NewFromInt(10000)).Round(0).IntPart()
		if amount > 0 {
			var commissionID int64
			if err := tx.QueryRowContext(ctx, `INSERT INTO saas_partner_commissions (referral_id, tenant_subscription_id, beneficiary_user_id, base_cny_minor, rate_bps, amount_cny_minor) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`, referralID, subscriptionID, beneficiaryID, paidMinor, rate, amount).Scan(&commissionID); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}
			if commissionID > 0 {
				if _, err := tx.ExecContext(ctx, `INSERT INTO saas_partner_wallets (tenant_id, user_id, frozen_cny_minor, lifetime_earned_cny_minor) VALUES (1, $1, $2, $2) ON CONFLICT (tenant_id, user_id) DO UPDATE SET frozen_cny_minor = saas_partner_wallets.frozen_cny_minor + EXCLUDED.frozen_cny_minor, lifetime_earned_cny_minor = saas_partner_wallets.lifetime_earned_cny_minor + EXCLUDED.lifetime_earned_cny_minor, updated_at = NOW()`, beneficiaryID, amount); err != nil {
					return 0, err
				}
				if err := insertPartnerWalletLedger(ctx, tx, beneficiaryID, "commission_frozen", amount, "saas_partner_commission", strconv.FormatInt(commissionID, 10), fmt.Sprintf("saas-partner-commission:%d", commissionID)); err != nil {
					return 0, err
				}
			}
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "saas_subscription", strconv.FormatInt(subscriptionID, 10), "saas.subscription_paid", fmt.Sprintf("saas-subscription:%d:paid", subscriptionID), map[string]any{"tenant_id": tenantID, "paid_cny_minor": paidMinor}); err != nil {
		return 0, err
	}
	return subscriptionID, tx.Commit()
}

func (s *SaaSService) PartnerDashboard(ctx context.Context, userID int64) (*PartnerDashboard, error) {
	if err := s.thawPartner(ctx, userID); err != nil {
		return nil, err
	}
	dashboard := &PartnerDashboard{}
	err := s.db.QueryRowContext(ctx, `SELECT available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, lifetime_earned_cny_minor FROM saas_partner_wallets WHERE tenant_id = 1 AND user_id = $1`, userID).Scan(&dashboard.AvailableMinor, &dashboard.FrozenMinor, &dashboard.WithdrawingMinor, &dashboard.LifetimeMinor)
	if errors.Is(err, sql.ErrNoRows) {
		return dashboard, nil
	}
	return dashboard, err
}

func (s *SaaSService) thawPartner(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var amount int64
	if err := tx.QueryRowContext(ctx, `WITH thawed AS (UPDATE saas_partner_commissions SET status = 'AVAILABLE', thawed_at = NOW() WHERE beneficiary_user_id = $1 AND status = 'FROZEN' AND frozen_until <= NOW() RETURNING amount_cny_minor) SELECT COALESCE(SUM(amount_cny_minor), 0) FROM thawed`, userID).Scan(&amount); err != nil {
		return err
	}
	if amount > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE saas_partner_wallets SET frozen_cny_minor = frozen_cny_minor - $2, available_cny_minor = available_cny_minor + $2, updated_at = NOW() WHERE tenant_id = 1 AND user_id = $1`, userID, amount); err != nil {
			return err
		}
		if err := insertPartnerWalletLedger(ctx, tx, userID, "commission_thaw", amount, "commission_batch", "", fmt.Sprintf("saas-partner-thaw:%d:%d", userID, time.Now().UnixNano())); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SaaSService) ThawPartnersDue(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT beneficiary_user_id
FROM saas_partner_commissions
WHERE status = 'FROZEN' AND frozen_until <= NOW()
ORDER BY beneficiary_user_id
LIMIT $1`, limit)
	if err != nil {
		return err
	}
	userIDs := make([]int64, 0, limit)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := s.thawPartner(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *SaaSService) CreatePartnerWithdrawal(ctx context.Context, userID, amountMinor int64) (*Withdrawal, error) {
	if err := s.thawPartner(ctx, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return nil, err
	}
	var minimum int64
	var dailyLimit, feeBPS, configVersion int
	if err := tx.QueryRowContext(ctx, `SELECT withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps, current_config_version FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&minimum, &dailyLimit, &feeBPS, &configVersion); err != nil {
		return nil, err
	}
	if amountMinor < minimum {
		return nil, ErrWithdrawalAmountInvalid
	}
	var payoutID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM distribution_payout_accounts WHERE tenant_id = 1 AND user_id = $1 AND account_type = 'alipay'`, userID).Scan(&payoutID); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayoutAccountRequired
	} else if err != nil {
		return nil, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM saas_partner_withdrawals WHERE user_id = $1 AND submitted_at >= date_trunc('day', NOW())`, userID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= dailyLimit {
		return nil, ErrWithdrawalLimitExceeded
	}
	var available int64
	if err := tx.QueryRowContext(ctx, `SELECT available_cny_minor FROM saas_partner_wallets WHERE tenant_id = 1 AND user_id = $1 FOR UPDATE`, userID).Scan(&available); errors.Is(err, sql.ErrNoRows) || available < amountMinor {
		return nil, ErrWithdrawalInsufficient
	} else if err != nil {
		return nil, err
	}
	feeMinor := calculateWithdrawalFee(amountMinor, feeBPS)
	if feeMinor >= amountMinor {
		return nil, ErrWithdrawalAmountInvalid
	}
	if _, err := tx.ExecContext(ctx, `UPDATE saas_partner_wallets SET available_cny_minor = available_cny_minor - $2, withdrawing_cny_minor = withdrawing_cny_minor + $2, updated_at = NOW() WHERE tenant_id = 1 AND user_id = $1`, userID, amountMinor); err != nil {
		return nil, err
	}
	var item Withdrawal
	if err := tx.QueryRowContext(ctx, `INSERT INTO saas_partner_withdrawals (tenant_id, user_id, payout_account_id, amount_cny_minor, fee_cny_minor, fee_rate_bps, config_version) VALUES (1, $1, $2, $3, $4, $5, $6) RETURNING id, amount_cny_minor, fee_cny_minor, fee_rate_bps, config_version, status, submitted_at`, userID, payoutID, amountMinor, feeMinor, feeBPS, configVersion).Scan(&item.ID, &item.AmountMinor, &item.FeeMinor, &item.FeeRateBPS, &item.ConfigVersion, &item.Status, &item.SubmittedAt); err != nil {
		return nil, err
	}
	if err := insertPartnerWalletLedger(ctx, tx, userID, "withdrawal_submitted", -amountMinor, "saas_partner_withdrawal", strconv.FormatInt(item.ID, 10), fmt.Sprintf("saas-partner-withdrawal:%d:submit", item.ID)); err != nil {
		return nil, err
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "saas_partner_withdrawal", strconv.FormatInt(item.ID, 10), "saas.partner_withdrawal_submitted", fmt.Sprintf("saas-partner-withdrawal:%d:submitted", item.ID), map[string]any{"user_id": userID, "amount_cny_minor": amountMinor}); err != nil {
		return nil, err
	}
	return &item, tx.Commit()
}

func (s *SaaSService) ListPartnerWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	rows, err := s.db.QueryContext(ctx, partnerWithdrawalSelectSQL+` WHERE w.user_id = $1 ORDER BY w.id DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Withdrawal, 0)
	for rows.Next() {
		item, err := scanWithdrawal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) AdminListPartnerWithdrawals(ctx context.Context, status string) ([]Withdrawal, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	pattern := "%"
	if status != "" {
		pattern = status
	}
	rows, err := s.db.QueryContext(ctx, partnerWithdrawalSelectSQL+` WHERE w.status LIKE $1 ORDER BY w.id ASC LIMIT 200`, pattern)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]Withdrawal, 0)
	for rows.Next() {
		item, err := scanWithdrawal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) AdminTransitionPartnerWithdrawal(ctx context.Context, withdrawalID, operatorID int64, target, reason, reference, proofURL string) (*Withdrawal, error) {
	target = strings.ToUpper(strings.TrimSpace(target))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	var userID, amount int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id, amount_cny_minor, status FROM saas_partner_withdrawals WHERE id = $1 FOR UPDATE`, withdrawalID).Scan(&userID, &amount, &current); errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	} else if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	switch {
	case target == "APPROVED" && current == "SUBMITTED":
		_, err = tx.ExecContext(ctx, `UPDATE saas_partner_withdrawals SET status = 'APPROVED', operator_user_id = $2, approved_at = $3, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorID, now)
	case target == "PAID" && current == "APPROVED" && strings.TrimSpace(reference) != "":
		_, err = tx.ExecContext(ctx, `UPDATE saas_partner_withdrawals SET status = 'PAID', operator_user_id = $2, payment_reference = $3, proof_url = NULLIF($4, ''), paid_at = $5, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorID, reference, proofURL, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE saas_partner_wallets SET withdrawing_cny_minor = withdrawing_cny_minor - $2, lifetime_withdrawn_cny_minor = lifetime_withdrawn_cny_minor + $2, updated_at = NOW() WHERE tenant_id = 1 AND user_id = $1`, userID, amount)
		}
	case target == "REJECTED" && (current == "SUBMITTED" || current == "APPROVED") && strings.TrimSpace(reason) != "":
		_, err = tx.ExecContext(ctx, `UPDATE saas_partner_withdrawals SET status = 'REJECTED', operator_user_id = $2, reject_reason = $3, rejected_at = $4, updated_at = NOW() WHERE id = $1`, withdrawalID, operatorID, reason, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE saas_partner_wallets SET withdrawing_cny_minor = withdrawing_cny_minor - $2, available_cny_minor = available_cny_minor + $2, updated_at = NOW() WHERE tenant_id = 1 AND user_id = $1`, userID, amount)
		}
	default:
		return nil, ErrWithdrawalStateInvalid
	}
	if err != nil {
		return nil, err
	}
	if target == "PAID" || target == "REJECTED" {
		if err := insertPartnerWalletLedger(ctx, tx, userID, "withdrawal_"+strings.ToLower(target), amount, "saas_partner_withdrawal", strconv.FormatInt(withdrawalID, 10), fmt.Sprintf("saas-partner-withdrawal:%d:%s", withdrawalID, strings.ToLower(target))); err != nil {
			return nil, err
		}
	}
	if err := insertFinancialOutboxEvent(ctx, tx, "saas_partner_withdrawal", strconv.FormatInt(withdrawalID, 10), "saas.partner_withdrawal_"+strings.ToLower(target), fmt.Sprintf("saas-partner-withdrawal:%d:%s", withdrawalID, strings.ToLower(target)), map[string]any{"operator_user_id": operatorID}); err != nil {
		return nil, err
	}
	item, err := scanWithdrawal(tx.QueryRowContext(ctx, partnerWithdrawalSelectSQL+` WHERE w.id = $1`, withdrawalID))
	if err != nil {
		return nil, err
	}
	return &item, tx.Commit()
}

func insertPartnerWalletLedger(ctx context.Context, tx *sql.Tx, userID int64, action string, amount int64, sourceType, sourceID, idempotency string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO saas_partner_wallet_ledger (tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, idempotency_key) SELECT 1, $1, $2, $3, $4, $5, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, $6 FROM saas_partner_wallets WHERE tenant_id = 1 AND user_id = $1 ON CONFLICT DO NOTHING`, userID, action, amount, sourceType, sourceID, idempotency)
	return err
}

const partnerWithdrawalSelectSQL = `SELECT w.id, w.amount_cny_minor, w.fee_cny_minor, w.fee_rate_bps, w.config_version, w.status, COALESCE(w.reject_reason, ''), COALESCE(w.payment_reference, ''), COALESCE(w.proof_url, ''), w.submitted_at, w.approved_at, w.paid_at, w.rejected_at FROM saas_partner_withdrawals w`

func (s *SaaSService) ListProvisioningJobs(ctx context.Context, tenantID int64) ([]ProvisioningJob, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, tenant_id, action, status, created_at FROM saas_provisioning_jobs WHERE ($1 = 0 OR tenant_id = $1) ORDER BY id DESC LIMIT 100`, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]ProvisioningJob, 0)
	for rows.Next() {
		var item ProvisioningJob
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Action, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SaaSService) AssignResourcePool(ctx context.Context, tenantID, groupID int64, allocationType string, concurrencyLimit int, monthlyLimitRaw string) error {
	allocationType = strings.ToLower(strings.TrimSpace(allocationType))
	if allocationType != "shared" && allocationType != "dedicated" {
		return infraerrors.BadRequest("RESOURCE_ALLOCATION_INVALID", "invalid resource allocation type")
	}
	if concurrencyLimit < 0 {
		return infraerrors.BadRequest("RESOURCE_ALLOCATION_INVALID", "invalid concurrency limit")
	}
	var monthly any
	if strings.TrimSpace(monthlyLimitRaw) != "" {
		value, err := decimal.NewFromString(monthlyLimitRaw)
		if err != nil || value.IsNegative() {
			return infraerrors.BadRequest("RESOURCE_ALLOCATION_INVALID", "invalid monthly limit")
		}
		monthly = value.String()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO saas_resource_pool_allocations (tenant_id, group_id, allocation_type, concurrency_limit, monthly_limit_usd) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (tenant_id, group_id) DO UPDATE SET allocation_type = EXCLUDED.allocation_type, concurrency_limit = EXCLUDED.concurrency_limit, monthly_limit_usd = EXCLUDED.monthly_limit_usd, updated_at = NOW()`, tenantID, groupID, allocationType, concurrencyLimit, monthly); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE api_keys SET group_id = $2, updated_at = NOW() WHERE tenant_id = $1 AND key_type = 'tenant_wholesale' AND deleted_at IS NULL`, tenantID, groupID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_allowed_groups (user_id, group_id, created_at) SELECT core_user_id, $2, NOW() FROM saas_tenants WHERE id = $1 AND core_user_id IS NOT NULL ON CONFLICT DO NOTHING`, tenantID, groupID); err != nil {
		return err
	}
	if concurrencyLimit > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE users u SET concurrency = $2, updated_at = NOW() FROM saas_tenants t WHERE t.id = $1 AND t.core_user_id = u.id`, tenantID, concurrencyLimit); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if s.authCache != nil {
		rows, err := s.db.QueryContext(ctx, `SELECT key FROM api_keys WHERE tenant_id = $1 AND key_type = 'tenant_wholesale' AND deleted_at IS NULL`, tenantID)
		if err == nil {
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var key string
				if rows.Scan(&key) == nil {
					s.authCache.InvalidateAuthCacheByKey(ctx, key)
				}
			}
		}
	}
	return nil
}

func (s *SaaSService) requireEnabled(ctx context.Context) error {
	value, err := s.settings.GetValue(ctx, "saas_control_plane_enabled")
	if err != nil {
		return err
	}
	if value != "true" {
		return ErrSaaSDisabled
	}
	return nil
}

func randomPrefixedSecret(prefix string, bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buffer), nil
}

func validTenantDomain(domain string) bool {
	if len(domain) < 4 || len(domain) > 253 || net.ParseIP(domain) != nil || strings.Contains(domain, "..") {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	labelPattern := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	for _, label := range labels {
		if !labelPattern.MatchString(label) {
			return false
		}
	}
	return true
}
