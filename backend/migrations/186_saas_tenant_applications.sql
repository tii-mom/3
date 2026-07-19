CREATE TABLE IF NOT EXISTS saas_tenant_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    brand_name VARCHAR(120) NOT NULL,
    contact_name VARCHAR(80) NOT NULL,
    contact_channel VARCHAR(24) NOT NULL
        CHECK (contact_channel IN ('email', 'phone', 'telegram', 'whatsapp', 'wechat', 'other')),
    contact_value VARCHAR(255) NOT NULL,
    desired_domain VARCHAR(255),
    expected_monthly_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (expected_monthly_usd >= 0),
    expected_users INTEGER NOT NULL DEFAULT 0 CHECK (expected_users >= 0),
    business_description TEXT NOT NULL DEFAULT '',
    referral_code VARCHAR(32) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'SUBMITTED'
        CHECK (status IN ('SUBMITTED', 'CONTACTED', 'APPROVED', 'REJECTED')),
    tenant_id BIGINT UNIQUE REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    reviewer_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    review_note TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_saas_tenant_applications_active_user
    ON saas_tenant_applications(user_id)
    WHERE status IN ('SUBMITTED', 'CONTACTED');

CREATE INDEX IF NOT EXISTS idx_saas_tenant_applications_admin_queue
    ON saas_tenant_applications(status, submitted_at, id);

CREATE TABLE IF NOT EXISTS saas_tenant_application_events (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES saas_tenant_applications(id) ON DELETE RESTRICT,
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL
        CHECK (to_status IN ('SUBMITTED', 'CONTACTED', 'APPROVED', 'REJECTED')),
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    actor_type VARCHAR(16) NOT NULL CHECK (actor_type IN ('user', 'admin', 'system')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saas_tenant_application_events_application
    ON saas_tenant_application_events(application_id, id);

INSERT INTO settings (key, value, updated_at) VALUES
    ('saas_application_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;
