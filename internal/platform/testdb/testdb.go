// Package testdb provides a serialized PostgreSQL helper for platform tests.
package testdb

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/postgres"
	"golang.org/x/sys/unix"
)

const defaultTestDatabaseURL = "postgres://treelot:treelot@localhost:5432/treelot_test?sslmode=disable"

// OpenEmpty returns an exclusive empty test database with foundation tables removed.
func OpenEmpty(t *testing.T) *postgres.DB {
	t.Helper()
	unlock := flock(t)
	t.Cleanup(unlock)

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultTestDatabaseURL
	}
	if err := ValidateTestDatabaseURL(databaseURL); err != nil {
		t.Fatalf("unsafe TEST_DATABASE_URL: %v", err)
	}

	db, err := postgres.Open(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(context.Background(), `
		DROP TABLE IF EXISTS sessions CASCADE;
		DROP TABLE IF EXISTS background_jobs CASCADE;
		DROP TABLE IF EXISTS outbox_messages CASCADE;
		DROP TABLE IF EXISTS audit_events CASCADE;
		DROP TABLE IF EXISTS schema_migrations CASCADE;
	`)
	if err != nil {
		t.Fatalf("reset test database: %v", err)
	}
	return db
}

// OpenMigrated returns an exclusive migrated test database connection.
func OpenMigrated(t *testing.T) *postgres.DB {
	t.Helper()
	db := OpenEmpty(t)

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		migrationsDir = filepath.Join("..", "..", "migrations")
	}
	if _, err := migrate.Up(context.Background(), db, migrationsDir); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	return db
}

func flock(t *testing.T) func() {
	t.Helper()
	path := filepath.Join(os.TempDir(), "treelot-testdb.lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open testdb lock: %v", err)
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		_ = file.Close()
		t.Fatalf("acquire testdb lock: %v", err)
	}
	return func() {
		_ = unix.Flock(int(file.Fd()), unix.LOCK_UN)
		_ = file.Close()
	}
}
