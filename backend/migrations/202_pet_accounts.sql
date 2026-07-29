-- 202_pet_accounts.sql
-- TAI Protocol pet compute accounts
-- Pets from TAI Protocol get provisioned 3api accounts to purchase AI compute with TAI tokens.

CREATE TABLE IF NOT EXISTS pet_accounts (
    id BIGSERIAL PRIMARY KEY,
    pet_id VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    api_key_id BIGINT REFERENCES api_keys(id),
    group_id BIGINT,
    owner_tg_id VARCHAR(64),
    tai_spent_total DECIMAL(20,8) NOT NULL DEFAULT 0,
    compute_credits_total DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    daily_tai_limit DECIMAL(20,8) NOT NULL DEFAULT 100,
    daily_tai_used DECIMAL(20,8) NOT NULL DEFAULT 0,
    daily_reset_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pet_accounts_user_id ON pet_accounts(user_id);
CREATE INDEX idx_pet_accounts_owner_tg_id ON pet_accounts(owner_tg_id);
CREATE INDEX idx_pet_accounts_status ON pet_accounts(status);

COMMENT ON TABLE pet_accounts IS 'TAI Protocol AI pet compute accounts - pets spend TAI tokens to buy API compute';
COMMENT ON COLUMN pet_accounts.pet_id IS 'External pet identifier from TAI Protocol (e.g. gen0-001)';
COMMENT ON COLUMN pet_accounts.tai_spent_total IS 'Cumulative TAI tokens spent by this pet on compute';
COMMENT ON COLUMN pet_accounts.daily_tai_limit IS 'Anti-abuse: max TAI a pet can spend per day';
