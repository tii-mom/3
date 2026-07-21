-- Scope conversion idempotency to the user while retaining one conversion
-- result per user/key. The original global key rejected legitimate retries
-- from different users that generated the same client key.
ALTER TABLE distribution_usd_conversions
    DROP CONSTRAINT IF EXISTS distribution_usd_conversions_program_id_idempotency_key_key;
CREATE UNIQUE INDEX IF NOT EXISTS distribution_usd_conversions_program_user_idempotency_key
    ON distribution_usd_conversions(program_id, user_id, idempotency_key);
