-- While credit buckets run in shadow mode, the compatibility balance remains
-- authoritative. Reconcile any drift without reducing a user's visible credit,
-- and retain an immutable ledger record for the correction.
WITH mismatches AS (
    SELECT
        u.id AS user_id,
        u.balance AS target_balance,
        a.transferable_credit AS transferable_before,
        a.non_transferable_credit AS non_transferable_before,
        a.debt AS debt_before,
        u.balance - (a.transferable_credit + a.non_transferable_credit - a.debt) AS delta
    FROM users u
    JOIN user_credit_accounts a ON a.user_id = u.id
    WHERE u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt
), reconciled AS (
    UPDATE user_credit_accounts a
    SET transferable_credit = CASE
            WHEN m.delta >= 0 THEN a.transferable_credit + m.delta
            ELSE GREATEST(
                a.transferable_credit - GREATEST(-m.delta - a.non_transferable_credit, 0),
                0
            )
        END,
        non_transferable_credit = CASE
            WHEN m.delta >= 0 THEN a.non_transferable_credit
            ELSE GREATEST(a.non_transferable_credit + m.delta, 0)
        END,
        debt = CASE
            WHEN m.delta >= 0 THEN a.debt
            ELSE a.debt + GREATEST(-m.delta - a.non_transferable_credit - a.transferable_credit, 0)
        END,
        updated_at = NOW()
    FROM mismatches m
    WHERE a.user_id = m.user_id
    RETURNING
        a.user_id,
        m.target_balance,
        m.transferable_before,
        m.non_transferable_before,
        m.debt_before,
        a.transferable_credit AS transferable_after,
        a.non_transferable_credit AS non_transferable_after,
        a.debt AS debt_after
)
INSERT INTO user_credit_ledger (
    tenant_id, user_id, entry_type, source_type, source_id,
    transferable_delta, non_transferable_delta, debt_delta,
    transferable_after, non_transferable_after, debt_after, balance_after,
    idempotency_key, metadata
)
SELECT
    1,
    r.user_id,
    'migration_reconciliation',
    'legacy_balance_reconciliation',
    '191_reconcile_legacy_credit_balances',
    r.transferable_after - r.transferable_before,
    r.non_transferable_after - r.non_transferable_before,
    r.debt_after - r.debt_before,
    r.transferable_after,
    r.non_transferable_after,
    r.debt_after,
    r.target_balance,
    'migration:191:user:' || r.user_id,
    jsonb_build_object('policy', 'legacy_balance_authoritative_while_shadow_mode')
FROM reconciled r
ON CONFLICT DO NOTHING;

UPDATE financial_balance_migration_audit audit
SET bucket_balance = u.balance,
    reconciliation_status = 'RECONCILED'
FROM users u
JOIN user_credit_accounts a ON a.user_id = u.id
WHERE audit.user_id = u.id
  AND u.balance = a.transferable_credit + a.non_transferable_credit - a.debt;
