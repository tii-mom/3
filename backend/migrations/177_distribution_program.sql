CREATE TABLE IF NOT EXISTS distribution_programs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(120) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    stack_with_legacy BOOLEAN NOT NULL DEFAULT FALSE,
    commission_freeze_hours INTEGER NOT NULL DEFAULT 168,
    withdrawal_min_cny_minor BIGINT NOT NULL DEFAULT 10000,
    withdrawal_daily_limit INTEGER NOT NULL DEFAULT 1,
    first_recharge_bonus_bps INTEGER NOT NULL DEFAULT 1000,
    first_recharge_bonus_cap_usd DECIMAL(20,8) NOT NULL DEFAULT 10000,
    current_config_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

INSERT INTO distribution_programs (tenant_id, code, name)
VALUES (1, 'compute_company', '算力公司')
ON CONFLICT (tenant_id, code) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_tier_configs (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    config_version INTEGER NOT NULL,
    tier INTEGER NOT NULL CHECK (tier BETWEEN 1 AND 3),
    threshold_cny_minor BIGINT NOT NULL CHECK (threshold_cny_minor > 0),
    level1_bps INTEGER NOT NULL,
    level2_bps INTEGER NOT NULL,
    level3_bps INTEGER NOT NULL,
    level4_bps INTEGER NOT NULL,
    level5_bps INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, config_version, tier)
);

INSERT INTO distribution_tier_configs
    (program_id, config_version, tier, threshold_cny_minor, level1_bps, level2_bps, level3_bps, level4_bps, level5_bps)
SELECT id, 1, tier, threshold, l1, l2, l3, l4, l5
FROM distribution_programs
CROSS JOIN (VALUES
    (1, 100000::BIGINT, 1000, 400, 300, 200, 100),
    (2, 1000000::BIGINT, 1500, 600, 400, 300, 200),
    (3, 10000000::BIGINT, 2000, 800, 600, 400, 200)
) AS tiers(tier, threshold, l1, l2, l3, l4, l5)
WHERE code = 'compute_company'
ON CONFLICT (program_id, config_version, tier) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_relations (
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    ancestor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    descendant_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    depth SMALLINT NOT NULL CHECK (depth BETWEEN 0 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (program_id, ancestor_user_id, descendant_user_id),
    CHECK ((depth = 0 AND ancestor_user_id = descendant_user_id) OR (depth > 0 AND ancestor_user_id <> descendant_user_id))
);

CREATE INDEX IF NOT EXISTS idx_distribution_relations_descendant ON distribution_relations(program_id, descendant_user_id, depth);
CREATE INDEX IF NOT EXISTS idx_distribution_relations_ancestor_depth ON distribution_relations(program_id, ancestor_user_id, depth, descendant_user_id);

CREATE TABLE IF NOT EXISTS distribution_members (
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_volume_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (team_volume_cny_minor >= 0),
    current_tier SMALLINT NOT NULL DEFAULT 0 CHECK (current_tier BETWEEN 0 AND 3),
    activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (program_id, user_id)
);

INSERT INTO distribution_members (program_id, tenant_id, user_id)
SELECT p.id, p.tenant_id, u.id FROM distribution_programs p CROSS JOIN users u
WHERE p.code = 'compute_company' AND u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO distribution_relations (program_id, tenant_id, ancestor_user_id, descendant_user_id, depth)
SELECT p.id, p.tenant_id, u.id, u.id, 0
FROM distribution_programs p CROSS JOIN users u
WHERE p.code = 'compute_company' AND u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

WITH RECURSIVE chain AS (
    SELECT ua.inviter_id AS ancestor_user_id, ua.user_id AS descendant_user_id, 1 AS depth
    FROM user_affiliates ua
    WHERE ua.inviter_id IS NOT NULL AND ua.inviter_id <> ua.user_id
    UNION ALL
    SELECT parent.inviter_id, chain.descendant_user_id, chain.depth + 1
    FROM chain
    JOIN user_affiliates parent ON parent.user_id = chain.ancestor_user_id
    WHERE chain.depth < 5 AND parent.inviter_id IS NOT NULL
      AND parent.inviter_id <> chain.descendant_user_id
)
INSERT INTO distribution_relations (program_id, tenant_id, ancestor_user_id, descendant_user_id, depth)
SELECT p.id, p.tenant_id, chain.ancestor_user_id, chain.descendant_user_id, chain.depth
FROM distribution_programs p CROSS JOIN chain
WHERE p.code = 'compute_company' AND chain.ancestor_user_id <> chain.descendant_user_id
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_recharge_events (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    source_order_id BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    base_cny_minor BIGINT NOT NULL CHECK (base_cny_minor > 0),
    credited_usd DECIMAL(20,8) NOT NULL CHECK (credited_usd > 0),
    first_recharge_bonus_usd DECIMAL(20,8) NOT NULL DEFAULT 0,
    config_version INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, source_order_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_first_recharge_bonus
    ON distribution_recharge_events(program_id, user_id) WHERE first_recharge_bonus_usd > 0;

CREATE TABLE IF NOT EXISTS distribution_cash_wallets (
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    available_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (available_cny_minor >= 0),
    frozen_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (frozen_cny_minor >= 0),
    withdrawing_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (withdrawing_cny_minor >= 0),
    lifetime_earned_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (lifetime_earned_cny_minor >= 0),
    lifetime_withdrawn_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (lifetime_withdrawn_cny_minor >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (program_id, user_id)
);

CREATE TABLE IF NOT EXISTS distribution_commissions (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    source_order_id BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,
    source_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    beneficiary_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    depth SMALLINT NOT NULL CHECK (depth BETWEEN 1 AND 5),
    tier SMALLINT NOT NULL CHECK (tier BETWEEN 1 AND 3),
    rate_bps INTEGER NOT NULL CHECK (rate_bps > 0),
    base_cny_minor BIGINT NOT NULL CHECK (base_cny_minor > 0),
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
    team_volume_cny_minor BIGINT NOT NULL CHECK (team_volume_cny_minor >= 0),
    config_version INTEGER NOT NULL,
    frozen_until TIMESTAMPTZ NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'FROZEN' CHECK (status IN ('FROZEN', 'AVAILABLE', 'REVERSED')),
	thawed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, source_order_id, beneficiary_user_id, depth)
);

CREATE INDEX IF NOT EXISTS idx_distribution_commissions_beneficiary ON distribution_commissions(program_id, beneficiary_user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS distribution_wallet_ledger (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(40) NOT NULL,
    amount_cny_minor BIGINT NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128),
    available_after BIGINT NOT NULL,
    frozen_after BIGINT NOT NULL,
    withdrawing_after BIGINT NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS distribution_payout_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_type VARCHAR(20) NOT NULL DEFAULT 'alipay',
    account_encrypted TEXT NOT NULL,
    account_last4 VARCHAR(8) NOT NULL,
    real_name_encrypted TEXT NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, account_type)
);

CREATE TABLE IF NOT EXISTS distribution_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payout_account_id BIGINT NOT NULL REFERENCES distribution_payout_accounts(id) ON DELETE RESTRICT,
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
    fee_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (fee_cny_minor >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'SUBMITTED',
    operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reject_reason TEXT,
    payment_reference VARCHAR(160),
    proof_url TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('SUBMITTED', 'APPROVED', 'PAID', 'REJECTED'))
);

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_user ON distribution_withdrawals(program_id, user_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_status ON distribution_withdrawals(status, submitted_at, id);
