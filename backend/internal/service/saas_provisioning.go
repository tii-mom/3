package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type preparedSaaSTenantProvisioning struct {
	apiKey        string
	desiredConfig string
}

func (s *SaaSService) prepareSaaSTenantProvisioning(slug string) (*preparedSaaSTenantProvisioning, error) {
	apiKey, err := randomPrefixedSecret("sk-wholesale-", 24)
	if err != nil {
		return nil, err
	}
	databasePassword, err := randomPrefixedSecret("db-", 24)
	if err != nil {
		return nil, err
	}
	if s.encryptor == nil {
		return nil, infraerrors.ServiceUnavailable("PROVISIONING_ENCRYPTION_UNAVAILABLE", "provisioning encryption is unavailable")
	}
	encryptedDatabasePassword, err := s.encryptor.Encrypt(databasePassword)
	if err != nil {
		return nil, err
	}
	desiredConfig, err := json.Marshal(map[string]any{
		"isolation": "dedicated_instance", "instance_name": "3api-" + slug,
		"database_schema":             "tenant_" + strings.ReplaceAll(slug, "-", "_"),
		"database_password_encrypted": encryptedDatabasePassword,
		"redis_namespace":             "tenant:" + slug,
		"object_prefix":               "tenants/" + slug,
		"log_label":                   "tenant=" + slug,
		"docker":                      map[string]any{"memory": "512m", "cpus": 1},
		"caddy":                       map[string]any{"tls": "dns_verified_only"},
	})
	if err != nil {
		return nil, err
	}
	return &preparedSaaSTenantProvisioning{apiKey: apiKey, desiredConfig: string(desiredConfig)}, nil
}

func (s *SaaSService) provisionTenantTx(
	ctx context.Context,
	tx *sql.Tx,
	input CreateSaaSTenantInput,
	prepared *preparedSaaSTenantProvisioning,
) (*SaaSTenant, error) {
	var tenant SaaSTenant
	err := tx.QueryRowContext(ctx, `
INSERT INTO saas_tenants (slug, name, status, site_name, site_logo, core_user_id)
VALUES ($1, $2, 'active', $3, $4, $5)
RETURNING id, slug, name, status, site_name, site_logo, COALESCE(primary_domain, ''), core_user_id, created_at`,
		input.Slug, input.Name, input.SiteName, input.SiteLogo, input.CoreUserID,
	).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.SiteName,
		&tenant.SiteLogo, &tenant.PrimaryDomain, &tenant.CoreUserID, &tenant.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO saas_wholesale_wallets (tenant_id) VALUES ($1)`, tenant.ID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO saas_tenant_configs (tenant_id) VALUES ($1)`, tenant.ID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO api_keys (user_id, key, name, status, key_type, tenant_id, created_at, updated_at)
VALUES ($1, $2, $3, 'active', 'tenant_wholesale', $4, NOW(), NOW())`,
		input.CoreUserID, prepared.apiKey, "Wholesale / "+input.Name, tenant.ID); err != nil {
		return nil, err
	}
	if referral := strings.ToUpper(strings.TrimSpace(input.ReferralCode)); referral != "" {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO saas_partner_referrals (referrer_user_id, referred_tenant_id, referral_code, commission_rate_bps, valid_until)
SELECT user_id, $1, $2, 1000, NOW() + INTERVAL '12 months'
FROM user_affiliates WHERE aff_code = $2 AND user_id <> $3
ON CONFLICT (referred_tenant_id) DO NOTHING`, tenant.ID, referral, input.CoreUserID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO saas_provisioning_jobs (tenant_id, action, desired_config)
VALUES ($1, 'provision', $2::jsonb)`, tenant.ID, prepared.desiredConfig); err != nil {
		return nil, err
	}
	tenant.WholesaleUSD = "0.00000000"
	return &tenant, nil
}
