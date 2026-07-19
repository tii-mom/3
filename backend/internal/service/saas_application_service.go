package service

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrSaaSApplicationDisabled         = infraerrors.Forbidden("SAAS_APPLICATION_DISABLED", "SaaS applications are disabled")
	ErrSaaSApplicationNotFound         = infraerrors.NotFound("SAAS_APPLICATION_NOT_FOUND", "SaaS application not found")
	ErrSaaSApplicationExists           = infraerrors.Conflict("SAAS_APPLICATION_EXISTS", "an active SaaS application already exists")
	ErrSaaSApplicationTransition       = infraerrors.Conflict("SAAS_APPLICATION_TRANSITION_INVALID", "invalid SaaS application status transition")
	ErrSaaSAlreadyTenant               = infraerrors.Conflict("SAAS_TENANT_ALREADY_EXISTS", "user already owns a SaaS tenant")
	ErrSaaSApplicationApprovalRequired = infraerrors.Conflict(
		"SAAS_APPLICATION_APPROVAL_REQUIRED",
		"user has a SaaS application and must be provisioned through the approval workflow",
	)
)

type SaaSTenantApplication struct {
	ID                  int64      `json:"id"`
	UserID              int64      `json:"user_id"`
	Username            string     `json:"username,omitempty"`
	UserEmail           string     `json:"user_email,omitempty"`
	BrandName           string     `json:"brand_name"`
	ContactName         string     `json:"contact_name"`
	ContactChannel      string     `json:"contact_channel"`
	ContactValue        string     `json:"contact_value"`
	DesiredDomain       string     `json:"desired_domain"`
	ExpectedMonthlyUSD  string     `json:"expected_monthly_usd"`
	ExpectedUsers       int        `json:"expected_users"`
	BusinessDescription string     `json:"business_description"`
	ReferralCode        string     `json:"referral_code,omitempty"`
	Status              string     `json:"status"`
	TenantID            *int64     `json:"tenant_id,omitempty"`
	ReviewerUserID      *int64     `json:"reviewer_user_id,omitempty"`
	ReviewNote          string     `json:"review_note"`
	SubmittedAt         time.Time  `json:"submitted_at"`
	ReviewedAt          *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type SubmitSaaSApplicationInput struct {
	BrandName           string
	ContactName         string
	ContactChannel      string
	ContactValue        string
	DesiredDomain       string
	ExpectedMonthlyUSD  string
	ExpectedUsers       int
	BusinessDescription string
	ReferralCode        string
}

type ApproveSaaSApplicationInput struct {
	Slug       string
	SiteName   string
	SiteLogo   string
	ReviewNote string
}

type SaaSApplicationOverview struct {
	ApplicationsEnabled bool                   `json:"applications_enabled"`
	Application         *SaaSTenantApplication `json:"application,omitempty"`
	Tenant              *SaaSTenant            `json:"tenant,omitempty"`
}

type SaaSApplicationApprovalResult struct {
	Application  SaaSTenantApplication `json:"application"`
	Tenant       SaaSTenant            `json:"tenant"`
	WholesaleKey string                `json:"wholesale_api_key"`
}

func (s *SaaSService) ApplicationEnabled(ctx context.Context) (bool, error) {
	value, err := s.settings.GetValue(ctx, "saas_application_enabled")
	return value == "true", err
}

func (s *SaaSService) UpdateApplicationEnabled(ctx context.Context, enabled bool) error {
	return s.settings.Set(ctx, "saas_application_enabled", strconv.FormatBool(enabled))
}

func (s *SaaSService) UpdateFeatureFlags(ctx context.Context, controlPlane, applications *bool) error {
	values := make(map[string]string, 2)
	if controlPlane != nil {
		values["saas_control_plane_enabled"] = strconv.FormatBool(*controlPlane)
	}
	if applications != nil {
		values["saas_application_enabled"] = strconv.FormatBool(*applications)
	}
	if len(values) == 0 {
		return nil
	}
	if err := s.settings.SetMultiple(ctx, values); err != nil {
		return err
	}
	if controlPlane != nil {
		s.invalidateWholesaleAuthCache(ctx, 0)
	}
	return nil
}

func (s *SaaSService) SubmitApplication(ctx context.Context, userID int64, input SubmitSaaSApplicationInput) (*SaaSTenantApplication, error) {
	enabled, err := s.ApplicationEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrSaaSApplicationDisabled
	}
	input, monthly, err := normalizeSaaSApplicationInput(input)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSaaSAlreadyTenant
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenant_applications WHERE user_id = $1 AND status IN ('SUBMITTED', 'CONTACTED'))`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSaaSApplicationExists
	}

	item := &SaaSTenantApplication{}
	var desiredDomain sql.NullString
	err = tx.QueryRowContext(ctx, `
INSERT INTO saas_tenant_applications (
    user_id, brand_name, contact_name, contact_channel, contact_value,
    desired_domain, expected_monthly_usd, expected_users, business_description, referral_code
) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10)
RETURNING id, user_id, brand_name, contact_name, contact_channel, contact_value,
    desired_domain, expected_monthly_usd::text, expected_users, business_description,
    referral_code, status, tenant_id, reviewer_user_id, review_note, submitted_at,
    reviewed_at, created_at, updated_at`,
		userID, input.BrandName, input.ContactName, input.ContactChannel, input.ContactValue,
		input.DesiredDomain, monthly.String(), input.ExpectedUsers, input.BusinessDescription, input.ReferralCode,
	).Scan(
		&item.ID, &item.UserID, &item.BrandName, &item.ContactName, &item.ContactChannel, &item.ContactValue,
		&desiredDomain, &item.ExpectedMonthlyUSD, &item.ExpectedUsers, &item.BusinessDescription,
		&item.ReferralCode, &item.Status, &item.TenantID, &item.ReviewerUserID, &item.ReviewNote, &item.SubmittedAt,
		&item.ReviewedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.DesiredDomain = desiredDomain.String
	if _, err := tx.ExecContext(ctx, `
INSERT INTO saas_tenant_application_events (application_id, from_status, to_status, actor_user_id, actor_type)
VALUES ($1, NULL, 'SUBMITTED', $2, 'user')`, item.ID, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *SaaSService) ApplicationOverview(ctx context.Context, userID int64) (*SaaSApplicationOverview, error) {
	enabled, err := s.ApplicationEnabled(ctx)
	if err != nil {
		return nil, err
	}
	overview := &SaaSApplicationOverview{ApplicationsEnabled: enabled}
	item, err := s.applicationByUser(ctx, userID)
	if err != nil && !errors.Is(err, ErrSaaSApplicationNotFound) {
		return nil, err
	}
	if err == nil {
		overview.Application = item
	}

	var tenant SaaSTenant
	err = s.db.QueryRowContext(ctx, `
SELECT t.id, t.slug, t.name, t.status, t.site_name, t.site_logo,
       COALESCE(t.primary_domain, ''), t.core_user_id, COALESCE(w.balance_usd, 0)::text, t.created_at
FROM saas_tenants t
LEFT JOIN saas_wholesale_wallets w ON w.tenant_id = t.id
WHERE t.core_user_id = $1 AND t.id <> 1
ORDER BY t.id DESC LIMIT 1`, userID).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.SiteName, &tenant.SiteLogo,
		&tenant.PrimaryDomain, &tenant.CoreUserID, &tenant.WholesaleUSD, &tenant.CreatedAt,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		overview.Tenant = &tenant
	}
	return overview, nil
}

func (s *SaaSService) AdminListApplications(ctx context.Context, status string, page, pageSize int) ([]SaaSTenantApplication, int64, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" && !validSaaSApplicationStatus(status) {
		return nil, 0, infraerrors.BadRequest("SAAS_APPLICATION_STATUS_INVALID", "invalid SaaS application status")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM saas_tenant_applications WHERE ($1 = '' OR status = $1)`, status).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT a.id, a.user_id, u.username, u.email, a.brand_name, a.contact_name,
       a.contact_channel, a.contact_value, COALESCE(a.desired_domain, ''),
       a.expected_monthly_usd::text, a.expected_users, a.business_description,
       a.referral_code, a.status, a.tenant_id, a.reviewer_user_id, a.review_note,
       a.submitted_at, a.reviewed_at, a.created_at, a.updated_at
FROM saas_tenant_applications a
JOIN users u ON u.id = a.user_id
WHERE ($1 = '' OR a.status = $1)
ORDER BY a.submitted_at DESC, a.id DESC
LIMIT $2 OFFSET $3`, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]SaaSTenantApplication, 0, pageSize)
	for rows.Next() {
		var item SaaSTenantApplication
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Username, &item.UserEmail, &item.BrandName, &item.ContactName,
			&item.ContactChannel, &item.ContactValue, &item.DesiredDomain, &item.ExpectedMonthlyUSD,
			&item.ExpectedUsers, &item.BusinessDescription, &item.ReferralCode, &item.Status,
			&item.TenantID, &item.ReviewerUserID, &item.ReviewNote, &item.SubmittedAt,
			&item.ReviewedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *SaaSService) ReviewApplication(ctx context.Context, applicationID, reviewerUserID int64, targetStatus, note string) (*SaaSTenantApplication, error) {
	targetStatus = strings.ToUpper(strings.TrimSpace(targetStatus))
	if targetStatus != "CONTACTED" && targetStatus != "REJECTED" {
		return nil, ErrSaaSApplicationTransition
	}
	if len(strings.TrimSpace(note)) > 4000 {
		return nil, infraerrors.BadRequest("SAAS_APPLICATION_NOTE_INVALID", "review note is too long")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM saas_tenant_applications WHERE id = $1 FOR UPDATE`, applicationID).Scan(&current); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSApplicationNotFound
	} else if err != nil {
		return nil, err
	}
	if !allowedSaaSApplicationTransition(current, targetStatus) {
		return nil, ErrSaaSApplicationTransition
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE saas_tenant_applications
SET status = $2, reviewer_user_id = $3, review_note = $4, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $1`, applicationID, targetStatus, reviewerUserID, strings.TrimSpace(note)); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO saas_tenant_application_events (application_id, from_status, to_status, actor_user_id, actor_type, note)
VALUES ($1, $2, $3, $4, 'admin', $5)`, applicationID, current, targetStatus, reviewerUserID, strings.TrimSpace(note)); err != nil {
		return nil, err
	}
	item, err := scanSaaSApplication(tx.QueryRowContext(ctx, saasApplicationSelectSQL+` WHERE id = $1`, applicationID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *SaaSService) ApproveApplication(ctx context.Context, applicationID, reviewerUserID int64, input ApproveSaaSApplicationInput) (*SaaSApplicationApprovalResult, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.SiteName = strings.TrimSpace(input.SiteName)
	input.SiteLogo = strings.TrimSpace(input.SiteLogo)
	input.ReviewNote = strings.TrimSpace(input.ReviewNote)
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`).MatchString(input.Slug) {
		return nil, infraerrors.BadRequest("SAAS_TENANT_INVALID", "invalid tenant slug")
	}
	if len(input.ReviewNote) > 4000 {
		return nil, infraerrors.BadRequest("SAAS_APPLICATION_NOTE_INVALID", "review note is too long")
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
	var userID int64
	var brandName, referralCode, current string
	if err := tx.QueryRowContext(ctx, `
SELECT user_id, brand_name, referral_code, status
FROM saas_tenant_applications WHERE id = $1 FOR UPDATE`, applicationID).
		Scan(&userID, &brandName, &referralCode, &current); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSApplicationNotFound
	} else if err != nil {
		return nil, err
	}
	if !allowedSaaSApplicationTransition(current, "APPROVED") {
		return nil, ErrSaaSApplicationTransition
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, input.Slug); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saas_tenants WHERE core_user_id = $1 AND id <> 1)`, userID).Scan(&exists); err != nil {
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
	if input.SiteName == "" {
		input.SiteName = brandName
	}
	tenant, err := s.provisionTenantTx(ctx, tx, CreateSaaSTenantInput{
		Slug: input.Slug, Name: brandName, SiteName: input.SiteName, SiteLogo: input.SiteLogo,
		CoreUserID: userID, ReferralCode: referralCode,
	}, prepared)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE saas_tenant_applications
SET status = 'APPROVED', tenant_id = $2, reviewer_user_id = $3,
    review_note = $4, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $1`, applicationID, tenant.ID, reviewerUserID, input.ReviewNote); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO saas_tenant_application_events (application_id, from_status, to_status, actor_user_id, actor_type, note)
VALUES ($1, $2, 'APPROVED', $3, 'admin', $4)`, applicationID, current, reviewerUserID, input.ReviewNote); err != nil {
		return nil, err
	}
	application, err := scanSaaSApplication(tx.QueryRowContext(ctx, saasApplicationSelectSQL+` WHERE id = $1`, applicationID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &SaaSApplicationApprovalResult{Application: *application, Tenant: *tenant, WholesaleKey: prepared.apiKey}, nil
}

func (s *SaaSService) applicationByUser(ctx context.Context, userID int64) (*SaaSTenantApplication, error) {
	return scanSaaSApplication(s.db.QueryRowContext(ctx, saasApplicationSelectSQL+` WHERE user_id = $1 ORDER BY id DESC LIMIT 1`, userID))
}

type saasApplicationScanner interface {
	Scan(dest ...any) error
}

func scanSaaSApplication(row saasApplicationScanner) (*SaaSTenantApplication, error) {
	item := &SaaSTenantApplication{}
	err := row.Scan(
		&item.ID, &item.UserID, &item.BrandName, &item.ContactName, &item.ContactChannel,
		&item.ContactValue, &item.DesiredDomain, &item.ExpectedMonthlyUSD, &item.ExpectedUsers,
		&item.BusinessDescription, &item.ReferralCode, &item.Status, &item.TenantID,
		&item.ReviewerUserID, &item.ReviewNote, &item.SubmittedAt, &item.ReviewedAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaaSApplicationNotFound
	}
	return item, err
}

const saasApplicationSelectSQL = `
SELECT id, user_id, brand_name, contact_name, contact_channel, contact_value,
       COALESCE(desired_domain, ''), expected_monthly_usd::text, expected_users,
       business_description, referral_code, status, tenant_id, reviewer_user_id,
       review_note, submitted_at, reviewed_at, created_at, updated_at
FROM saas_tenant_applications`

func normalizeSaaSApplicationInput(input SubmitSaaSApplicationInput) (SubmitSaaSApplicationInput, decimal.Decimal, error) {
	input.BrandName = strings.TrimSpace(input.BrandName)
	input.ContactName = strings.TrimSpace(input.ContactName)
	input.ContactChannel = strings.ToLower(strings.TrimSpace(input.ContactChannel))
	input.ContactValue = strings.TrimSpace(input.ContactValue)
	input.DesiredDomain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(input.DesiredDomain), "."))
	input.BusinessDescription = strings.TrimSpace(input.BusinessDescription)
	input.ReferralCode = strings.ToUpper(strings.TrimSpace(input.ReferralCode))
	if input.BrandName == "" || len(input.BrandName) > 120 || input.ContactName == "" || len(input.ContactName) > 80 || input.ContactValue == "" || len(input.ContactValue) > 255 {
		return input, decimal.Zero, infraerrors.BadRequest("SAAS_APPLICATION_INVALID", "brand and contact details are required")
	}
	if !validSaaSContactChannel(input.ContactChannel) {
		return input, decimal.Zero, infraerrors.BadRequest("SAAS_APPLICATION_CONTACT_INVALID", "invalid contact channel")
	}
	if input.DesiredDomain != "" && !validTenantDomain(input.DesiredDomain) {
		return input, decimal.Zero, ErrSaaSDomainInvalid
	}
	if input.ExpectedUsers < 0 || input.ExpectedUsers > 10000000 || len(input.BusinessDescription) > 4000 || len(input.ReferralCode) > 32 {
		return input, decimal.Zero, infraerrors.BadRequest("SAAS_APPLICATION_INVALID", "application details are outside the allowed limits")
	}
	monthly := decimal.Zero
	if strings.TrimSpace(input.ExpectedMonthlyUSD) != "" {
		var err error
		monthly, err = decimal.NewFromString(strings.TrimSpace(input.ExpectedMonthlyUSD))
		if err != nil || monthly.IsNegative() || monthly.Exponent() < -8 || monthly.GreaterThan(decimal.NewFromInt(1000000000)) {
			return input, decimal.Zero, infraerrors.BadRequest("SAAS_APPLICATION_USAGE_INVALID", "invalid expected monthly usage")
		}
	}
	return input, monthly, nil
}

func validSaaSContactChannel(value string) bool {
	switch value {
	case "email", "phone", "telegram", "whatsapp", "wechat", "other":
		return true
	default:
		return false
	}
}

func validSaaSApplicationStatus(value string) bool {
	switch value {
	case "SUBMITTED", "CONTACTED", "APPROVED", "REJECTED":
		return true
	default:
		return false
	}
}

func allowedSaaSApplicationTransition(from, to string) bool {
	switch from {
	case "SUBMITTED":
		return to == "CONTACTED" || to == "APPROVED" || to == "REJECTED"
	case "CONTACTED":
		return to == "APPROVED" || to == "REJECTED"
	default:
		return false
	}
}
