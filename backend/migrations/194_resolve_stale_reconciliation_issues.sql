-- Resolve stale reconciliation alerts only when the live credit-account
-- invariant already proves the user's legacy balance and bucket balance match.
-- This keeps the audit rows, records why they were closed, and still leaves
-- any real balance mismatch OPEN for financialgate to block deployment.
UPDATE financial_reconciliation_issues AS i
SET status = 'RESOLVED',
    resolved_at = COALESCE(i.resolved_at, NOW()),
    metadata = COALESCE(i.metadata, '{}'::jsonb) || jsonb_build_object(
        'auto_resolved_by', '194_resolve_stale_reconciliation_issues',
        'auto_resolved_reason', 'current_credit_bucket_balance_matches_legacy_balance',
        'auto_resolved_at', NOW()
    )
FROM users AS u
JOIN user_credit_accounts AS a ON a.user_id = u.id
WHERE i.status = 'OPEN'
  AND i.user_id = u.id
  AND u.balance = a.transferable_credit + a.non_transferable_credit - a.debt;
