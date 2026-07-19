CREATE TABLE IF NOT EXISTS saas_plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    billing_period VARCHAR(20) NOT NULL DEFAULT 'month',
    price_cny_minor BIGINT NOT NULL CHECK (price_cny_minor >= 0),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    limits JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_tenant_domains (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES saas_tenants(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    verification_token VARCHAR(96) NOT NULL,
    verified_at TIMESTAMPTZ,
    tls_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_tenant_configs (
    tenant_id BIGINT PRIMARY KEY REFERENCES saas_tenants(id) ON DELETE CASCADE,
    retail_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1 CHECK (retail_multiplier > 0),
    payment_provider VARCHAR(40) NOT NULL DEFAULT '',
    payment_config_encrypted TEXT,
    instance_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_tenant_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    plan_id BIGINT NOT NULL REFERENCES saas_plans(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    paid_cny_minor BIGINT NOT NULL DEFAULT 0,
    payment_reference VARCHAR(160),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_saas_subscription_payment_reference
    ON saas_tenant_subscriptions(tenant_id, payment_reference)
    WHERE payment_reference IS NOT NULL AND payment_reference <> '';

CREATE TABLE IF NOT EXISTS saas_wholesale_wallets (
    tenant_id BIGINT PRIMARY KEY REFERENCES saas_tenants(id) ON DELETE CASCADE,
    balance_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (balance_usd >= 0),
    lifetime_funded_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (lifetime_funded_usd >= 0),
    lifetime_used_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (lifetime_used_usd >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_wholesale_ledger (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    action VARCHAR(32) NOT NULL,
    amount_usd DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128),
    idempotency_key VARCHAR(160) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, idempotency_key)
);

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_type VARCHAR(24) NOT NULL DEFAULT 'user';
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES saas_tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_type ON api_keys(tenant_id, key_type) WHERE tenant_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS saas_resource_pool_allocations (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES saas_tenants(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    allocation_type VARCHAR(20) NOT NULL DEFAULT 'shared',
    concurrency_limit INTEGER NOT NULL DEFAULT 0,
    monthly_limit_usd DECIMAL(20,8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, group_id)
);

CREATE TABLE IF NOT EXISTS saas_provisioning_jobs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES saas_tenants(id) ON DELETE CASCADE,
    action VARCHAR(32) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    desired_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_partner_referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referred_tenant_id BIGINT NOT NULL UNIQUE REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    referral_code VARCHAR(32) NOT NULL,
    commission_rate_bps INTEGER NOT NULL DEFAULT 1000,
    valid_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_partner_commissions (
    id BIGSERIAL PRIMARY KEY,
    referral_id BIGINT NOT NULL REFERENCES saas_partner_referrals(id) ON DELETE RESTRICT,
    tenant_subscription_id BIGINT NOT NULL REFERENCES saas_tenant_subscriptions(id) ON DELETE RESTRICT,
    beneficiary_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    base_cny_minor BIGINT NOT NULL CHECK (base_cny_minor > 0),
    rate_bps INTEGER NOT NULL DEFAULT 1000,
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
	status VARCHAR(20) NOT NULL DEFAULT 'FROZEN' CHECK (status IN ('FROZEN', 'AVAILABLE', 'REVERSED')),
	frozen_until TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
	thawed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_subscription_id, beneficiary_user_id)
);

CREATE TABLE IF NOT EXISTS saas_partner_wallets (
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    available_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (available_cny_minor >= 0),
    frozen_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (frozen_cny_minor >= 0),
    withdrawing_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (withdrawing_cny_minor >= 0),
    lifetime_earned_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (lifetime_earned_cny_minor >= 0),
    lifetime_withdrawn_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (lifetime_withdrawn_cny_minor >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);

CREATE TABLE IF NOT EXISTS saas_partner_wallet_ledger (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(40) NOT NULL,
    amount_cny_minor BIGINT NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128),
    available_after BIGINT NOT NULL,
    frozen_after BIGINT NOT NULL,
    withdrawing_after BIGINT NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS saas_partner_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payout_account_id BIGINT NOT NULL REFERENCES distribution_payout_accounts(id) ON DELETE RESTRICT,
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
    fee_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (fee_cny_minor >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'SUBMITTED' CHECK (status IN ('SUBMITTED', 'APPROVED', 'PAID', 'REJECTED')),
    operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reject_reason TEXT,
    payment_reference VARCHAR(160),
    proof_url TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saas_partner_withdrawals_status ON saas_partner_withdrawals(status, submitted_at, id);

INSERT INTO settings (key, value, updated_at) VALUES
    ('distribution_enabled', 'false', NOW()),
    ('saas_control_plane_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;
