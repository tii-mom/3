-- Production rollout controls and versioned withdrawal policy. This migration
-- is additive so environments that have already applied 176-178 do not need
-- checksum exceptions.
INSERT INTO settings (key, value, updated_at)
VALUES ('credit_bucket_enforce_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS financial_reconciliation_issues (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128),
    compatibility_balance DECIMAL(20,8) NOT NULL,
    bucket_balance DECIMAL(20,8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'RESOLVED')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_financial_reconciliation_issues_open
    ON financial_reconciliation_issues(status, created_at, id)
    WHERE status = 'OPEN';

ALTER TABLE distribution_programs
    ADD COLUMN IF NOT EXISTS withdrawal_fee_bps INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS distribution_policy_versions (
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    config_version INTEGER NOT NULL,
    commission_freeze_hours INTEGER NOT NULL CHECK (commission_freeze_hours >= 0),
    withdrawal_min_cny_minor BIGINT NOT NULL CHECK (withdrawal_min_cny_minor > 0),
    withdrawal_daily_limit INTEGER NOT NULL CHECK (withdrawal_daily_limit > 0),
    withdrawal_fee_bps INTEGER NOT NULL CHECK (withdrawal_fee_bps BETWEEN 0 AND 10000),
    first_recharge_bonus_bps INTEGER NOT NULL CHECK (first_recharge_bonus_bps BETWEEN 0 AND 10000),
    first_recharge_bonus_cap_usd DECIMAL(20,8) NOT NULL CHECK (first_recharge_bonus_cap_usd >= 0),
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (program_id, config_version)
);

INSERT INTO distribution_policy_versions (
    program_id, config_version, commission_freeze_hours,
    withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps,
    first_recharge_bonus_bps, first_recharge_bonus_cap_usd
)
SELECT id, current_config_version, commission_freeze_hours,
       withdrawal_min_cny_minor, withdrawal_daily_limit, withdrawal_fee_bps,
       first_recharge_bonus_bps, first_recharge_bonus_cap_usd
FROM distribution_programs
ON CONFLICT (program_id, config_version) DO NOTHING;

ALTER TABLE distribution_withdrawals
    ADD COLUMN IF NOT EXISTS config_version INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS fee_rate_bps INTEGER NOT NULL DEFAULT 0;

ALTER TABLE saas_partner_withdrawals
    ADD COLUMN IF NOT EXISTS config_version INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS fee_rate_bps INTEGER NOT NULL DEFAULT 0;
