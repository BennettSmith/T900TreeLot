package main

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestRunRejectsIncompatibleSchema(t *testing.T) {
	db := testdb.OpenMigrated(t)
	configureWorkerEnv(t)

	_, _ = db.Exec(context.Background(), `DELETE FROM schema_migrations`)
	_, _ = db.Exec(context.Background(), `INSERT INTO schema_migrations (version) VALUES (99)`)

	status := run(context.Background(), slog.New(slog.DiscardHandler), 1)
	if status != 1 {
		t.Fatalf("status = %d, want 1", status)
	}
}

func TestRunProcessesIdleTick(t *testing.T) {
	_ = testdb.OpenMigrated(t)
	configureWorkerEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := run(ctx, slog.New(slog.DiscardHandler), 1)
	if status != 0 {
		t.Fatalf("status = %d, want 0", status)
	}
}

func configureWorkerEnv(t *testing.T) {
	t.Helper()
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
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
}
