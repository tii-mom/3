-- Refunds are intentionally unsupported by platform policy. Clear historical
-- provider flags and prevent application or manual writes from re-enabling them.
UPDATE payment_provider_instances
SET refund_enabled = FALSE,
    allow_user_refund = FALSE
WHERE refund_enabled = TRUE OR allow_user_refund = TRUE;

ALTER TABLE payment_provider_instances
    DROP CONSTRAINT IF EXISTS payment_provider_instances_refunds_disabled;

ALTER TABLE payment_provider_instances
    ADD CONSTRAINT payment_provider_instances_refunds_disabled
    CHECK (refund_enabled = FALSE AND allow_user_refund = FALSE) NOT VALID;

ALTER TABLE payment_provider_instances
    VALIDATE CONSTRAINT payment_provider_instances_refunds_disabled;
