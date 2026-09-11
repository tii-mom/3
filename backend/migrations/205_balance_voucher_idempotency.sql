ALTER TABLE balance_vouchers
    ADD COLUMN IF NOT EXISTS idempotency_key_hash CHAR(64),
    ADD COLUMN IF NOT EXISTS idempotency_request_hash CHAR(64),
    ADD COLUMN IF NOT EXISTS code_encrypted TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_balance_vouchers_issuer_idempotency
    ON balance_vouchers(tenant_id, issuer_user_id, idempotency_key_hash)
    WHERE idempotency_key_hash IS NOT NULL;
