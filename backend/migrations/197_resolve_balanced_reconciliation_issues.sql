-- Resolve balanced reconciliation alerts that may be created after earlier
-- cleanup migrations have already run in production.
--
-- This migration is audit-only: it preserves every issue row and never changes
-- users, credit buckets, orders, wallets, vouchers, or ledger amounts. Any real
-- mismatch remains OPEN and continues to block financialgate and feature
-- enablement.
UPDATE financial_reconciliation_issues AS i
SET status = 'RESOLVED',
    resolved_at = COALESCE(i.resolved_at, NOW()),
    metadata = COALESCE(i.metadata, '{}'::jsonb) || jsonb_build_object(
        'auto_resolved_by', '197_resolve_balanced_reconciliation_issues',
        'auto_resolved_reason', 'current_credit_bucket_balance_matches_legacy_balance',
        'auto_resolved_at', NOW()
    )
FROM users AS u
JOIN user_credit_accounts AS a ON a.user_id = u.id
WHERE i.status = 'OPEN'
  AND i.user_id = u.id
  AND u.balance = a.transferable_credit + a.non_transferable_credit - a.debt;
