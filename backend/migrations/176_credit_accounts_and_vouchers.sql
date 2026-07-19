-- Tenant scope used by newly introduced financial domains. Existing core tables
-- remain physically single-tenant during the managed-instance MVP.
CREATE TABLE IF NOT EXISTS saas_tenants (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'active',
    site_name VARCHAR(120) NOT NULL DEFAULT '',
    site_logo TEXT NOT NULL DEFAULT '',
    primary_domain VARCHAR(255),
    core_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO saas_tenants (id, slug, name, site_name)
VALUES (1, '3api', '3API', '3API')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('saas_tenants', 'id'), GREATEST((SELECT MAX(id) FROM saas_tenants), 1));

CREATE TABLE IF NOT EXISTS user_credit_accounts (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    transferable_credit DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (transferable_credit >= 0),
    non_transferable_credit DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (non_transferable_credit >= 0),
    debt DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (debt >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_credit_accounts_tenant ON user_credit_accounts(tenant_id, user_id);

-- The approved migration policy treats every positive legacy balance as
-- transferable. Negative balances become explicit debt.
INSERT INTO user_credit_accounts (user_id, tenant_id, transferable_credit, non_transferable_credit, debt)
SELECT id, 1, GREATEST(balance, 0), 0, GREATEST(-balance, 0)
FROM users
ON CONFLICT (user_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS financial_balance_migration_audit (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    legacy_balance DECIMAL(20,8) NOT NULL,
    bucket_balance DECIMAL(20,8) NOT NULL,
    legacy_total_recharged DECIMAL(20,8) NOT NULL,
    payment_total_recharged DECIMAL(20,8) NOT NULL,
    reconciliation_status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO financial_balance_migration_audit (
    user_id, legacy_balance, bucket_balance, legacy_total_recharged,
    payment_total_recharged, reconciliation_status
)
SELECT u.id,
       u.balance,
       a.transferable_credit + a.non_transferable_credit - a.debt,
       u.total_recharged,
       COALESCE(p.total, 0),
       CASE WHEN u.balance = a.transferable_credit + a.non_transferable_credit - a.debt
            THEN 'RECONCILED' ELSE 'REVIEW' END
FROM users u
JOIN user_credit_accounts a ON a.user_id = u.id
LEFT JOIN (
    SELECT user_id, SUM(amount) AS total
    FROM payment_orders
    WHERE order_type = 'balance' AND status = 'COMPLETED'
    GROUP BY user_id
) p ON p.user_id = u.id
ON CONFLICT (user_id) DO NOTHING;

UPDATE users u
SET total_recharged = audit.payment_total_recharged,
    updated_at = NOW()
FROM financial_balance_migration_audit audit
WHERE audit.user_id = u.id;

CREATE TABLE IF NOT EXISTS user_credit_ledger (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_type VARCHAR(40) NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128),
    transferable_delta DECIMAL(20,8) NOT NULL DEFAULT 0,
    non_transferable_delta DECIMAL(20,8) NOT NULL DEFAULT 0,
    debt_delta DECIMAL(20,8) NOT NULL DEFAULT 0,
    transferable_after DECIMAL(20,8) NOT NULL,
    non_transferable_after DECIMAL(20,8) NOT NULL,
    debt_after DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    idempotency_key VARCHAR(160),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_credit_ledger_user_created ON user_credit_ledger(user_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_credit_ledger_idempotency
    ON user_credit_ledger(tenant_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_credit_holds (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hold_key VARCHAR(160) NOT NULL,
    transferable_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (transferable_amount >= 0),
    non_transferable_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (non_transferable_amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'HELD' CHECK (status IN ('HELD', 'CAPTURED', 'RELEASED')),
    actual_amount DECIMAL(20,8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ,
    UNIQUE (tenant_id, hold_key)
);

CREATE TABLE IF NOT EXISTS financial_outbox_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    aggregate_type VARCHAR(48) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    idempotency_key VARCHAR(160) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_financial_outbox_pending
    ON financial_outbox_events(status, available_at, id) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS balance_vouchers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    issuer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    redeemer_user_id BIGINT REFERENCES users(id) ON DELETE RESTRICT,
    code_hash CHAR(64) NOT NULL UNIQUE,
    code_last4 CHAR(4) NOT NULL,
    face_value DECIMAL(20,8) NOT NULL CHECK (face_value > 0),
    fee_amount DECIMAL(20,8) NOT NULL CHECK (fee_amount >= 0),
    fee_rate_bps INTEGER NOT NULL CHECK (fee_rate_bps >= 0 AND fee_rate_bps <= 10000),
    status VARCHAR(24) NOT NULL DEFAULT 'ISSUED',
    expires_at TIMESTAMPTZ NOT NULL,
    redeemed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    risk_locked_at TIMESTAMPTZ,
    risk_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('ISSUED', 'REDEEMED', 'CANCELLED', 'EXPIRED', 'RISK_LOCKED'))
);

CREATE INDEX IF NOT EXISTS idx_balance_vouchers_issuer ON balance_vouchers(tenant_id, issuer_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_balance_vouchers_expiry ON balance_vouchers(status, expires_at) WHERE status = 'ISSUED';

CREATE TABLE IF NOT EXISTS balance_voucher_ledger (
    id BIGSERIAL PRIMARY KEY,
    voucher_id BIGINT NOT NULL REFERENCES balance_vouchers(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(32) NOT NULL,
    face_value DECIMAL(20,8) NOT NULL,
    fee_amount DECIMAL(20,8) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_balance_voucher_ledger_voucher ON balance_voucher_ledger(voucher_id, created_at, id);

INSERT INTO settings (key, value, updated_at) VALUES
    ('balance_voucher_enabled', 'false', NOW()),
    ('balance_voucher_fee_bps', '800', NOW()),
    ('balance_voucher_min_usd', '10', NOW()),
    ('balance_voucher_max_usd', '10000', NOW()),
    ('balance_voucher_daily_usd', '30000', NOW()),
    ('balance_voucher_daily_count', '10', NOW()),
    ('balance_voucher_expiry_days', '30', NOW()),
    ('balance_voucher_step_up_usd', '1000', NOW())
ON CONFLICT (key) DO NOTHING;
