-- Enable the finance features only for a genuinely empty installation.
-- Existing environments keep their current operator-controlled values.
DO $$
DECLARE
    user_count BIGINT;
    payment_order_count BIGINT;
    credit_account_count BIGINT;
    credit_ledger_count BIGINT;
    voucher_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO payment_order_count FROM payment_orders;
    SELECT COUNT(*) INTO credit_account_count FROM user_credit_accounts;
    SELECT COUNT(*) INTO credit_ledger_count FROM user_credit_ledger;
    SELECT COUNT(*) INTO voucher_count FROM balance_vouchers;

    IF user_count = 0
       AND payment_order_count = 0
       AND credit_account_count = 0
       AND credit_ledger_count = 0
       AND voucher_count = 0 THEN
        UPDATE settings
        SET value = 'true', updated_at = NOW()
        WHERE key IN (
            'credit_bucket_enforce_enabled',
            'balance_voucher_enabled',
            'distribution_enabled'
        );

        UPDATE distribution_programs
        SET enabled = TRUE, stack_with_legacy = FALSE, updated_at = NOW()
        WHERE tenant_id = 1 AND code = 'compute_company';
    END IF;
END
$$;
