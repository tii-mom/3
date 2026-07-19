// Command financialgate validates financial migrations and reconciliation
// invariants against an explicitly selected PostgreSQL database.
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/migrations"
	_ "github.com/lib/pq"
)

var requiredFinancialMigrations = []string{
	"175_disable_refunds_platform_wide.sql",
	"176_credit_accounts_and_vouchers.sql",
	"177_distribution_program.sql",
	"178_saas_control_plane.sql",
	"179_financial_runtime_controls.sql",
	"180_distribution_reversals.sql",
}

type gateReport struct {
	DatabaseVersion       string            `json:"database_version"`
	DatabaseName          string            `json:"database_name"`
	DatabaseHost          string            `json:"database_host"`
	MigrationsBefore      int64             `json:"migrations_before"`
	MigrationsAfterFirst  int64             `json:"migrations_after_first"`
	MigrationsAfterSecond int64             `json:"migrations_after_second"`
	RequiredMigrations    []string          `json:"required_migrations"`
	Reconciliation        map[string]int64  `json:"reconciliation"`
	OutboxByStatus        map[string]int64  `json:"outbox_by_status"`
	Scenarios             map[string]string `json:"scenarios,omitempty"`
	Duration              string            `json:"duration"`
}

type zeroCheck struct {
	name  string
	query string
}

var reconciliationChecks = []zeroCheck{
	{
		name: "credit_bucket_balance_mismatch",
		query: `SELECT COUNT(*) FROM users u JOIN user_credit_accounts a ON a.user_id = u.id
WHERE u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt`,
	},
	{
		name:  "migration_audit_unreconciled",
		query: `SELECT COUNT(*) FROM financial_balance_migration_audit WHERE reconciliation_status <> 'RECONCILED'`,
	},
	{
		name:  "open_reconciliation_issues",
		query: `SELECT COUNT(*) FROM financial_reconciliation_issues WHERE status = 'OPEN'`,
	},
	{
		name: "voucher_without_ledger",
		query: `SELECT COUNT(*) FROM balance_vouchers v
WHERE NOT EXISTS (SELECT 1 FROM balance_voucher_ledger l WHERE l.voucher_id = v.id)`,
	},
	{
		name: "negative_distribution_wallet",
		query: `SELECT COUNT(*) FROM distribution_cash_wallets
WHERE available_cny_minor < 0 OR frozen_cny_minor < 0 OR withdrawing_cny_minor < 0 OR debt_cny_minor < 0`,
	},
	{
		name: "reversed_recharge_without_single_reversal",
		query: `SELECT COUNT(*) FROM (
SELECT e.id FROM distribution_recharge_events e
LEFT JOIN distribution_reversal_events r ON r.recharge_event_id = e.id
WHERE e.status = 'REVERSED' GROUP BY e.id HAVING COUNT(r.id) <> 1
) anomalies`,
	},
	{
		name: "duplicate_distribution_commission",
		query: `SELECT COUNT(*) FROM (
SELECT program_id, source_order_id, beneficiary_user_id, depth
FROM distribution_commissions
GROUP BY program_id, source_order_id, beneficiary_user_id, depth HAVING COUNT(*) > 1
) anomalies`,
	},
	{
		name: "invalid_distribution_relation",
		query: `SELECT COUNT(*) FROM distribution_relations
WHERE depth < 0 OR depth > 5
   OR (depth = 0 AND ancestor_user_id <> descendant_user_id)
   OR (depth > 0 AND ancestor_user_id = descendant_user_id)`,
	},
	{
		name:  "negative_wholesale_wallet",
		query: `SELECT COUNT(*) FROM saas_wholesale_wallets WHERE balance_usd < 0`,
	},
	{
		name: "negative_partner_wallet",
		query: `SELECT COUNT(*) FROM saas_partner_wallets
WHERE available_cny_minor < 0 OR frozen_cny_minor < 0 OR withdrawing_cny_minor < 0`,
	},
}

func main() {
	var dsn string
	var timeout time.Duration
	var runScenarios bool
	var stressOrders, stressConcurrency int
	flag.StringVar(&dsn, "database-url", os.Getenv("DATABASE_URL"), "explicit PostgreSQL DATABASE_URL")
	flag.DurationVar(&timeout, "timeout", 5*time.Minute, "overall gate timeout")
	flag.BoolVar(&runScenarios, "run-scenarios", false, "run destructive financial scenarios (requires an empty local database)")
	flag.IntVar(&stressOrders, "stress-orders", 0, "process this many additional fixture recharge orders")
	flag.IntVar(&stressConcurrency, "stress-concurrency", 32, "workers used by -stress-orders")
	flag.Parse()

	if strings.TrimSpace(dsn) == "" {
		fatal(errors.New("database URL is required through -database-url or DATABASE_URL"))
	}
	if err := validateTarget(dsn, os.Getenv("FINANCIAL_GATE_ALLOW_NON_LOCAL") == "true"); err != nil {
		fatal(err)
	}

	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fatal(fmt.Errorf("open database: %w", err))
	}
	defer db.Close()
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	if err := db.PingContext(ctx); err != nil {
		fatal(fmt.Errorf("ping database: %w", err))
	}

	report, err := runGate(ctx, db)
	if err != nil {
		fatal(err)
	}
	if runScenarios {
		report.Scenarios, err = runFinancialScenarios(ctx, db)
		if err != nil {
			fatal(fmt.Errorf("financial scenarios: %w", err))
		}
		if err := runReconciliationChecks(ctx, db, report.Reconciliation); err != nil {
			fatal(err)
		}
	}
	if stressOrders > 0 {
		if report.Scenarios == nil {
			report.Scenarios = make(map[string]string)
		}
		report.Scenarios["distribution_stress"], err = runDistributionStress(ctx, db, stressOrders, stressConcurrency)
		if err != nil {
			fatal(fmt.Errorf("distribution stress: %w", err))
		}
		if err := runReconciliationChecks(ctx, db, report.Reconciliation); err != nil {
			fatal(err)
		}
	}
	if err := readOutboxStatus(ctx, db, report.OutboxByStatus); err != nil {
		fatal(err)
	}
	report.Duration = time.Since(started).Round(time.Millisecond).String()
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(fmt.Errorf("encode report: %w", err))
	}
	fmt.Println(string(encoded))
}

func runGate(ctx context.Context, db *sql.DB) (*gateReport, error) {
	report := &gateReport{
		RequiredMigrations: append([]string(nil), requiredFinancialMigrations...),
		Reconciliation:     make(map[string]int64, len(reconciliationChecks)),
		OutboxByStatus:     make(map[string]int64),
	}
	if err := db.QueryRowContext(ctx, `SELECT version(), current_database(), COALESCE(inet_server_addr()::text, 'local-socket')`).
		Scan(&report.DatabaseVersion, &report.DatabaseName, &report.DatabaseHost); err != nil {
		return nil, fmt.Errorf("read database identity: %w", err)
	}

	report.MigrationsBefore, _ = migrationCount(ctx, db)
	if err := repository.ApplyMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("first migration pass: %w", err)
	}
	var err error
	report.MigrationsAfterFirst, err = migrationCount(ctx, db)
	if err != nil {
		return nil, err
	}
	if err := verifyRequiredMigrations(ctx, db); err != nil {
		return nil, err
	}
	if err := repository.ApplyMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("second migration pass: %w", err)
	}
	report.MigrationsAfterSecond, err = migrationCount(ctx, db)
	if err != nil {
		return nil, err
	}
	if report.MigrationsAfterFirst != report.MigrationsAfterSecond {
		return nil, fmt.Errorf("migration pass was not idempotent: first=%d second=%d", report.MigrationsAfterFirst, report.MigrationsAfterSecond)
	}
	if err := verifyFeatureDefaults(ctx, db); err != nil {
		return nil, err
	}
	if err := runReconciliationChecks(ctx, db, report.Reconciliation); err != nil {
		return nil, err
	}
	if err := readOutboxStatus(ctx, db, report.OutboxByStatus); err != nil {
		return nil, err
	}
	return report, nil
}

func readOutboxStatus(ctx context.Context, db *sql.DB, result map[string]int64) error {
	clear(result)
	rows, err := db.QueryContext(ctx, `SELECT status, COUNT(*) FROM financial_outbox_events GROUP BY status ORDER BY status`)
	if err != nil {
		return fmt.Errorf("read outbox status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return err
		}
		result[status] = count
	}
	return rows.Err()
}

func runReconciliationChecks(ctx context.Context, db *sql.DB, results map[string]int64) error {
	for _, check := range reconciliationChecks {
		var count int64
		if err := db.QueryRowContext(ctx, check.query).Scan(&count); err != nil {
			return fmt.Errorf("run reconciliation check %s: %w", check.name, err)
		}
		results[check.name] = count
		if count != 0 {
			return fmt.Errorf("reconciliation check %s found %d anomalies", check.name, count)
		}
	}
	return nil
}

func migrationCount(ctx context.Context, db *sql.DB) (int64, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('schema_migrations') IS NOT NULL`).Scan(&exists); err != nil {
		return 0, fmt.Errorf("detect migration table: %w", err)
	}
	if !exists {
		return 0, nil
	}
	var count int64
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count migrations: %w", err)
	}
	return count, nil
}

func verifyRequiredMigrations(ctx context.Context, db *sql.DB) error {
	for _, name := range requiredFinancialMigrations {
		content, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read embedded migration %s: %w", name, err)
		}
		sum := sha256.Sum256([]byte(strings.TrimSpace(string(content))))
		expected := hex.EncodeToString(sum[:])
		var actual string
		if err := db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE filename = $1`, name).Scan(&actual); err != nil {
			return fmt.Errorf("required migration %s missing: %w", name, err)
		}
		if actual != expected {
			return fmt.Errorf("required migration %s checksum mismatch: db=%s embedded=%s", name, actual, expected)
		}
	}
	return nil
}

func verifyFeatureDefaults(ctx context.Context, db *sql.DB) error {
	keys := []string{
		"credit_bucket_enforce_enabled",
		"balance_voucher_enabled",
		"distribution_enabled",
		"saas_control_plane_enabled",
	}
	for _, key := range keys {
		var value string
		if err := db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&value); err != nil {
			return fmt.Errorf("read feature default %s: %w", key, err)
		}
		if !strings.EqualFold(strings.TrimSpace(value), "false") {
			return fmt.Errorf("feature %s must remain disabled before rollout, got %q", key, value)
		}
	}
	var enabled, stack bool
	if err := db.QueryRowContext(ctx, `SELECT enabled, stack_with_legacy FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`).Scan(&enabled, &stack); err != nil {
		return fmt.Errorf("read distribution defaults: %w", err)
	}
	if enabled || stack {
		return fmt.Errorf("distribution defaults are unsafe: enabled=%t stack_with_legacy=%t", enabled, stack)
	}
	return nil
}

func validateTarget(dsn string, allowNonLocal bool) error {
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		return errors.New("financialgate requires a postgres:// or postgresql:// DATABASE_URL")
	}
	if parsed.Hostname() == "" || parsed.Path == "" || parsed.Path == "/" {
		return errors.New("DATABASE_URL must include an explicit host and database name")
	}
	if allowNonLocal {
		return nil
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("refusing non-local database host %q; use an isolated local database or explicitly set FINANCIAL_GATE_ALLOW_NON_LOCAL=true", host)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "financialgate:", err)
	os.Exit(1)
}
