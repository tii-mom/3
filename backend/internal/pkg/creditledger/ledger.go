package creditledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"
	"github.com/shopspring/decimal"
)

var ErrUserNotFound = errors.New("credit ledger user not found")

func IsSchemaUnavailable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such table: user_credit_accounts") ||
		strings.Contains(message, `relation "user_credit_accounts" does not exist`)
}

type QueryExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Account struct {
	Transferable    decimal.Decimal
	NonTransferable decimal.Decimal
	Debt            decimal.Decimal
}

func (a Account) Balance() decimal.Decimal {
	return a.Transferable.Add(a.NonTransferable).Sub(a.Debt)
}

// EnsureAccount creates a bucket account from the compatibility balance. This
// is also used for users created after the backfill migration.
func EnsureAccount(ctx context.Context, db QueryExecer, userID int64) error {
	result, err := db.ExecContext(ctx, `
INSERT INTO user_credit_accounts (user_id, tenant_id, transferable_credit, non_transferable_credit, debt)
SELECT id, 1, GREATEST(balance, 0), 0, GREATEST(-balance, 0)
FROM users
WHERE id = $1 AND deleted_at IS NULL
ON CONFLICT (user_id) DO NOTHING`, userID)
	if err != nil {
		return fmt.Errorf("ensure credit account: %w", err)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		rows, queryErr := db.QueryContext(ctx, `SELECT 1 FROM user_credit_accounts WHERE user_id = $1`, userID)
		if queryErr != nil {
			return queryErr
		}
		defer func() { _ = rows.Close() }()
		if !rows.Next() {
			return ErrUserNotFound
		}
	}
	return nil
}

func InitializeNewAccount(ctx context.Context, db QueryExecer, userID int64, balance decimal.Decimal) error {
	transferable, nonTransferable, debt := decimal.Zero, decimal.Zero, decimal.Zero
	if balance.IsNegative() {
		debt = balance.Abs()
	} else {
		nonTransferable = balance
	}
	result, err := db.ExecContext(ctx, `INSERT INTO user_credit_accounts (user_id, tenant_id, transferable_credit, non_transferable_credit, debt) VALUES ($1, 1, $2, $3, $4) ON CONFLICT (user_id) DO NOTHING`, userID, transferable.String(), nonTransferable.String(), debt.String())
	if err != nil {
		return fmt.Errorf("initialize credit account: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 || balance.IsZero() {
		return nil
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_credit_ledger (tenant_id, user_id, entry_type, source_type, transferable_delta, non_transferable_delta, debt_delta, transferable_after, non_transferable_after, debt_after, balance_after, idempotency_key) VALUES (1, $1, 'account_opening', 'user_creation', $2, $3, $4, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`, userID, transferable.String(), nonTransferable.String(), debt.String(), balance.String(), fmt.Sprintf("user:%d:account-opening", userID))
	return err
}

// Apply updates credit buckets, the users.balance compatibility snapshot, and
// the immutable ledger while the caller's transaction is holding a row lock.
func Apply(ctx context.Context, db QueryExecer, userID int64, amount decimal.Decimal, metadata creditctx.Metadata, floorAtZero bool) (Account, bool, error) {
	if amount.IsZero() {
		return loadAccount(ctx, db, userID, false)
	}
	if err := EnsureAccount(ctx, db, userID); err != nil {
		return Account{}, false, err
	}
	if metadata.IdempotencyKey != "" {
		exists, err := ledgerEntryExists(ctx, db, metadata.IdempotencyKey)
		if err != nil {
			return Account{}, false, err
		}
		if exists {
			account, _, loadErr := loadAccount(ctx, db, userID, false)
			return account, false, loadErr
		}
	}

	account, _, err := loadAccount(ctx, db, userID, true)
	if err != nil {
		return Account{}, false, err
	}
	before := account
	enforceBuckets, err := bucketEnforcementEnabled(ctx, db)
	if err != nil {
		return Account{}, false, err
	}
	remaining := amount
	if amount.IsPositive() {
		repaid := decimal.Min(account.Debt, remaining)
		account.Debt = account.Debt.Sub(repaid)
		remaining = remaining.Sub(repaid)
		if metadata.Transferable {
			account.Transferable = account.Transferable.Add(remaining)
		} else {
			account.NonTransferable = account.NonTransferable.Add(remaining)
		}
	} else {
		debit := amount.Abs()
		if metadata.DebitTransferableFirst {
			fromTransferable := decimal.Min(account.Transferable, debit)
			account.Transferable = account.Transferable.Sub(fromTransferable)
			debit = debit.Sub(fromTransferable)
			fromNonTransferable := decimal.Min(account.NonTransferable, debit)
			account.NonTransferable = account.NonTransferable.Sub(fromNonTransferable)
			debit = debit.Sub(fromNonTransferable)
		} else {
			fromNonTransferable := decimal.Min(account.NonTransferable, debit)
			account.NonTransferable = account.NonTransferable.Sub(fromNonTransferable)
			debit = debit.Sub(fromNonTransferable)
			fromTransferable := decimal.Min(account.Transferable, debit)
			account.Transferable = account.Transferable.Sub(fromTransferable)
			debit = debit.Sub(fromTransferable)
		}
		if !floorAtZero {
			account.Debt = account.Debt.Add(debit)
		}
	}

	metadataJSON, err := json.Marshal(metadata.Attributes)
	if err != nil {
		return Account{}, false, fmt.Errorf("marshal credit metadata: %w", err)
	}
	if len(metadataJSON) == 0 || string(metadataJSON) == "null" {
		metadataJSON = []byte("{}")
	}
	var sourceID any
	if metadata.SourceID != "" {
		sourceID = metadata.SourceID
	}
	var idempotencyKey any
	if metadata.IdempotencyKey != "" {
		idempotencyKey = metadata.IdempotencyKey
	}

	result, err := db.ExecContext(ctx, `
UPDATE user_credit_accounts
SET transferable_credit = $2, non_transferable_credit = $3, debt = $4, updated_at = NOW()
WHERE user_id = $1`, userID, account.Transferable.String(), account.NonTransferable.String(), account.Debt.String())
	if err != nil {
		return Account{}, false, fmt.Errorf("update credit account: %w", err)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected != 1 {
		return Account{}, false, ErrUserNotFound
	}

	rechargeDelta := decimal.Zero
	if metadata.CountRecharge {
		rechargeDelta = amount
	}
	compatibilityBalance, err := updateCompatibilityBalance(ctx, db, userID, amount, account.Balance(), rechargeDelta, floorAtZero, enforceBuckets)
	if err != nil {
		return Account{}, false, fmt.Errorf("update compatibility balance: %w", err)
	}
	if !compatibilityBalance.Equal(account.Balance()) {
		if err := recordReconciliationIssue(ctx, db, userID, metadata, compatibilityBalance, account.Balance()); err != nil {
			return Account{}, false, err
		}
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO user_credit_ledger (
    tenant_id, user_id, entry_type, source_type, source_id,
    transferable_delta, non_transferable_delta, debt_delta,
    transferable_after, non_transferable_after, debt_after, balance_after,
    idempotency_key, metadata
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb)`,
		userID, metadata.EntryType, metadata.SourceType, sourceID,
		account.Transferable.Sub(before.Transferable).String(),
		account.NonTransferable.Sub(before.NonTransferable).String(),
		account.Debt.Sub(before.Debt).String(),
		account.Transferable.String(), account.NonTransferable.String(), account.Debt.String(), account.Balance().String(),
		idempotencyKey, string(metadataJSON),
	)
	if err != nil {
		return Account{}, false, fmt.Errorf("insert credit ledger: %w", err)
	}
	return account, true, nil
}

func bucketEnforcementEnabled(ctx context.Context, db QueryExecer) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT value FROM settings WHERE key = 'credit_bucket_enforce_enabled' LIMIT 1`)
	if err != nil {
		return false, fmt.Errorf("load credit bucket enforcement setting: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, rows.Err()
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(raw), "true"), rows.Err()
}

func updateCompatibilityBalance(ctx context.Context, db QueryExecer, userID int64, amount, bucketBalance, rechargeDelta decimal.Decimal, floorAtZero, enforce bool) (decimal.Decimal, error) {
	query := `
UPDATE users
SET balance = CASE
        WHEN $4 THEN $2
        WHEN $5 THEN GREATEST(balance + $6, 0)
        ELSE balance + $6
    END,
    total_recharged = GREATEST(total_recharged + $3, 0),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING balance`
	rows, err := db.QueryContext(ctx, query, userID, bucketBalance.String(), rechargeDelta.String(), enforce, floorAtZero, amount.String())
	if err != nil {
		return decimal.Zero, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return decimal.Zero, err
		}
		return decimal.Zero, ErrUserNotFound
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromString(raw)
}

func recordReconciliationIssue(ctx context.Context, db QueryExecer, userID int64, metadata creditctx.Metadata, compatibilityBalance, bucketBalance decimal.Decimal) error {
	var sourceID any
	if metadata.SourceID != "" {
		sourceID = metadata.SourceID
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO financial_reconciliation_issues (
    tenant_id, user_id, source_type, source_id, compatibility_balance, bucket_balance, metadata
) VALUES (1, $1, $2, $3, $4, $5, jsonb_build_object('entry_type', $6))`,
		userID, metadata.SourceType, sourceID, compatibilityBalance.String(), bucketBalance.String(), metadata.EntryType)
	if err != nil {
		return fmt.Errorf("record credit reconciliation issue: %w", err)
	}
	return nil
}

func loadAccount(ctx context.Context, db QueryExecer, userID int64, forUpdate bool) (Account, bool, error) {
	query := `SELECT transferable_credit::text, non_transferable_credit::text, debt::text FROM user_credit_accounts WHERE user_id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return Account{}, false, fmt.Errorf("load credit account: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return Account{}, false, ErrUserNotFound
	}
	var transferable, nonTransferable, debt string
	if err := rows.Scan(&transferable, &nonTransferable, &debt); err != nil {
		return Account{}, false, err
	}
	account := Account{}
	if account.Transferable, err = decimal.NewFromString(transferable); err != nil {
		return Account{}, false, err
	}
	if account.NonTransferable, err = decimal.NewFromString(nonTransferable); err != nil {
		return Account{}, false, err
	}
	if account.Debt, err = decimal.NewFromString(debt); err != nil {
		return Account{}, false, err
	}
	return account, true, rows.Err()
}

func ledgerEntryExists(ctx context.Context, db QueryExecer, key string) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT 1 FROM user_credit_ledger WHERE tenant_id = 1 AND idempotency_key = $1`, key)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	return rows.Next(), rows.Err()
}
