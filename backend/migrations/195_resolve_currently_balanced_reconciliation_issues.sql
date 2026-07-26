-- Re-run the guarded stale-issue cleanup for reconciliation alerts created
-- after migration 194 had already been applied in production.
--
-- This is audit-only: it never changes user balances, credit buckets, orders,
-- or ledger rows. An issue is closed only when the current authoritative
-- invariant proves that the legacy visible balance and the credit bucket
-- balance are already identical. Any real mismatch remains OPEN and continues
-- to block financialgate and feature enablement.
UPDATE financial_reconciliation_issues AS i
SET status = 'RESOLVED',
    resolved_at = COALESCE(i.resolved_at, NOW()),
    metadata = COALESCE(i.metadata, '{}'::jsonb) || jsonb_build_object(
        'auto_resolved_by', '195_resolve_currently_balanced_reconciliation_issues',
        'auto_resolved_reason', 'current_credit_bucket_balance_matches_legacy_balance',
        'auto_resolved_at', NOW()
    )
FROM users AS u
JOIN user_credit_accounts AS a ON a.user_id = u.id
WHERE i.status = 'OPEN'
  AND i.user_id = u.id
  AND u.balance = a.transferable_credit + a.non_transferable_credit - a.debt;
