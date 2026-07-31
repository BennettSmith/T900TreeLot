package main

import (
	"log/slog"
	"os"
	"testing"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestRunRequiresUpArgument(t *testing.T) {
	if status := run(slog.New(slog.DiscardHandler), nil); status != 2 {
		t.Fatalf("status = %d, want 2", status)
	}
}

func TestRunAppliesMigrations(t *testing.T) {
	_ = testdb.OpenEmpty(t)
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://treelot:treelot@localhost:5432/treelot_test?sslmode=disable"
	}
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", url)
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	if status := run(slog.New(slog.DiscardHandler), []string{"up"}); status != 0 {
		t.Fatalf("status = %d, want 0", status)
	}
}
