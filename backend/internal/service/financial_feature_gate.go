package service

import (
	"context"
	"database/sql"
)

// financialQueryExecutor keeps the readiness check usable inside the
// distribution transaction as well as for voucher settings updates.
type financialQueryExecutor interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// ensureCreditBucketsReady verifies that enabling a user-facing financial
// feature cannot strand a user on the legacy balance path. The check is
// intentionally read-only; callers decide when to persist the enforcement
// setting and the feature flag.
func ensureCreditBucketsReady(ctx context.Context, db financialQueryExecutor) error {
	var missingAccounts int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM users u
LEFT JOIN user_credit_accounts a ON a.user_id = u.id
WHERE a.user_id IS NULL`).Scan(&missingAccounts); err != nil {
		return err
	}
	if missingAccounts > 0 {
		return ErrCreditBucketsNotEnforced
	}

	var openIssues int64
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM financial_reconciliation_issues i
LEFT JOIN users u ON u.id = i.user_id
LEFT JOIN user_credit_accounts a ON a.user_id = i.user_id
WHERE i.status = 'OPEN'
  AND (
      u.id IS NULL
      OR a.user_id IS NULL
      OR u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt
  )`).Scan(&openIssues); err != nil {
		return err
	}

	if openIssues > 0 {
		return ErrCreditBucketsNotEnforced
	}
	return nil
}
