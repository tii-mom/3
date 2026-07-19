package main

import "testing"

func TestValidateTargetRejectsNonLocalByDefault(t *testing.T) {
	err := validateTarget("postgres://user:pass@db.example.com:5432/app?sslmode=require", false)
	if err == nil {
		t.Fatal("expected non-local target to be rejected")
	}
}

func TestValidateTargetAllowsLoopback(t *testing.T) {
	for _, dsn := range []string{
		"postgres://user:pass@127.0.0.1:5432/app?sslmode=disable",
		"postgresql://user:pass@localhost:5432/app?sslmode=disable",
		"postgres://user:pass@[::1]:5432/app?sslmode=disable",
	} {
		if err := validateTarget(dsn, false); err != nil {
			t.Fatalf("validateTarget(%q): %v", dsn, err)
		}
	}
}

func TestValidateTargetRequiresExplicitDatabase(t *testing.T) {
	if err := validateTarget("postgres://user:pass@127.0.0.1:5432", false); err == nil {
		t.Fatal("expected missing database name to be rejected")
	}
}

func TestValidateTargetRequiresPostgresURL(t *testing.T) {
	if err := validateTarget("host=127.0.0.1 dbname=app", false); err == nil {
		t.Fatal("expected keyword DSN to be rejected")
	}
}
