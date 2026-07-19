-- Chargeback and payment reversal support for the compute-company program.
-- Historical financial rows remain immutable; current state is corrected by
-- explicit reversal records and signed wallet ledger entries.

ALTER TABLE distribution_recharge_events
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'APPLIED',
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
    ADD COLUMN IF NOT EXISTS reversal_operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'distribution_recharge_events_status_check'
    ) THEN
        ALTER TABLE distribution_recharge_events
            ADD CONSTRAINT distribution_recharge_events_status_check
            CHECK (status IN ('APPLIED', 'REVERSED'));
    END IF;
END $$;

-- A reversed payment no longer consumes the user's first-recharge reward.
DROP INDEX IF EXISTS idx_distribution_first_recharge_bonus;
CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_first_recharge_bonus
    ON distribution_recharge_events(program_id, user_id)
    WHERE first_recharge_bonus_usd > 0 AND status = 'APPLIED';

ALTER TABLE distribution_cash_wallets
    ADD COLUMN IF NOT EXISTS debt_cny_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lifetime_reversed_cny_minor BIGINT NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'distribution_cash_wallets_debt_nonnegative'
    ) THEN
        ALTER TABLE distribution_cash_wallets
            ADD CONSTRAINT distribution_cash_wallets_debt_nonnegative
            CHECK (debt_cny_minor >= 0);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'distribution_cash_wallets_reversed_nonnegative'
    ) THEN
        ALTER TABLE distribution_cash_wallets
            ADD CONSTRAINT distribution_cash_wallets_reversed_nonnegative
            CHECK (lifetime_reversed_cny_minor >= 0);
    END IF;
END $$;

ALTER TABLE distribution_commissions
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
    ADD COLUMN IF NOT EXISTS reversal_operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE distribution_wallet_ledger
    ADD COLUMN IF NOT EXISTS debt_after BIGINT NOT NULL DEFAULT 0;

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
    ADD COLUMN IF NOT EXISTS reversal_operator_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_ledger_order_reversal
    ON user_affiliate_ledger(source_order_id, user_id, action)
    WHERE source_order_id IS NOT NULL AND action = 'reverse';

CREATE TABLE IF NOT EXISTS distribution_reversal_events (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES distribution_programs(id) ON DELETE RESTRICT,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    recharge_event_id BIGINT NOT NULL REFERENCES distribution_recharge_events(id) ON DELETE RESTRICT,
    source_order_id BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reversal_type VARCHAR(24) NOT NULL,
    base_cny_minor BIGINT NOT NULL CHECK (base_cny_minor > 0),
    principal_usd DECIMAL(20,8) NOT NULL CHECK (principal_usd > 0),
    bonus_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (bonus_usd >= 0),
    legacy_rebate_usd DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (legacy_rebate_usd >= 0),
    commission_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (commission_cny_minor >= 0),
    reason TEXT NOT NULL,
    operator_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, recharge_event_id),
    UNIQUE (program_id, source_order_id),
    CHECK (reversal_type IN ('CHARGEBACK', 'REFUND', 'ADMIN_CORRECTION'))
);

CREATE INDEX IF NOT EXISTS idx_distribution_reversals_user_created
    ON distribution_reversal_events(program_id, user_id, created_at DESC);
