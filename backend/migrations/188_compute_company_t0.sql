-- Unify the former direct reward into the compute-company program as T0.
-- Existing policy versions remain immutable; the current program moves to a
-- new four-tier policy version below.

DO $$
DECLARE
    compute_program_id BIGINT;
    previous_version INTEGER;
    next_version INTEGER;
BEGIN
    SELECT id, current_config_version
    INTO compute_program_id, previous_version
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
    FOR UPDATE;

    IF compute_program_id IS NULL THEN
        RAISE EXCEPTION 'compute_company distribution program is missing';
    END IF;

    next_version := previous_version + 1;

    -- tier 0 is now a valid active tier. Program enabled state, rather than
    -- tier value, distinguishes an inactive program from T0.
    ALTER TABLE distribution_tier_configs DROP CONSTRAINT IF EXISTS distribution_tier_configs_tier_check;
    ALTER TABLE distribution_tier_configs DROP CONSTRAINT IF EXISTS distribution_tier_configs_threshold_cny_minor_check;
    ALTER TABLE distribution_tier_configs
        ADD CONSTRAINT distribution_tier_configs_tier_check CHECK (tier BETWEEN 0 AND 3),
        ADD CONSTRAINT distribution_tier_configs_threshold_cny_minor_check CHECK (threshold_cny_minor >= 0);

    ALTER TABLE distribution_members DROP CONSTRAINT IF EXISTS distribution_members_tier_override_check;
    ALTER TABLE distribution_members
        ADD CONSTRAINT distribution_members_tier_override_check CHECK (tier_override IS NULL OR tier_override BETWEEN 0 AND 3);

    ALTER TABLE distribution_commissions DROP CONSTRAINT IF EXISTS distribution_commissions_tier_check;
    ALTER TABLE distribution_commissions
        ADD CONSTRAINT distribution_commissions_tier_check CHECK (tier BETWEEN 0 AND 3);

    INSERT INTO distribution_tier_configs
        (program_id, config_version, tier, threshold_cny_minor, level1_bps, level2_bps, level3_bps, level4_bps, level5_bps)
    VALUES
        (compute_program_id, next_version, 0, 0, 1000, 0, 0, 0, 0),
        (compute_program_id, next_version, 1, 100000, 1000, 400, 300, 200, 100),
        (compute_program_id, next_version, 2, 1000000, 1500, 600, 400, 300, 200),
        (compute_program_id, next_version, 3, 10000000, 2000, 800, 600, 400, 200)
    ON CONFLICT (program_id, config_version, tier) DO NOTHING;

    INSERT INTO distribution_policy_versions
        (program_id, config_version, commission_freeze_hours, withdrawal_min_cny_minor,
         withdrawal_daily_limit, withdrawal_fee_bps, first_recharge_bonus_bps,
         first_recharge_bonus_cap_usd)
    SELECT p.id, next_version, p.commission_freeze_hours, 2000,
           1, withdrawal_fee_bps, first_recharge_bonus_bps,
           first_recharge_bonus_cap_usd
    FROM distribution_programs p
    WHERE p.id = compute_program_id
    ON CONFLICT (program_id, config_version) DO NOTHING;

    UPDATE distribution_programs
    SET current_config_version = next_version,
        withdrawal_min_cny_minor = 2000,
        withdrawal_daily_limit = 1,
        stack_with_legacy = FALSE,
        updated_at = NOW()
    WHERE id = compute_program_id;
END $$;

-- The conversion rate is intentionally separate from subscription pricing.
INSERT INTO settings (key, value, updated_at)
VALUES ('distribution_usd_to_cny_rate', '7.15', NOW())
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_usd_conversions (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
    usd_amount DECIMAL(20,8) NOT NULL CHECK (usd_amount > 0),
    usd_to_cny_rate DECIMAL(20,10) NOT NULL CHECK (usd_to_cny_rate > 0),
    config_version INTEGER NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_distribution_usd_conversions_user
    ON distribution_usd_conversions(program_id, user_id, created_at DESC);
