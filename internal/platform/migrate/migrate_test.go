package migrate_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestUpAppliesFoundationMigrationOnce(t *testing.T) {
	db := testdb.OpenMigrated(t)

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	applied, err := migrate.Up(context.Background(), db, migrationsDir)
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if applied != 0 {
		t.Fatalf("applied = %d, want 0 after OpenMigrated", applied)
	}

	version, err := migrate.CurrentVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("CurrentVersion: %v", err)
	}
	if version != 6 {
		t.Fatalf("version = %d, want 6", version)
	}
}

func TestEnsureCompatibleRejectsMismatch(t *testing.T) {
	db := testdb.OpenMigrated(t)

	if err := migrate.EnsureCompatible(context.Background(), db, 6); err != nil {
		t.Fatalf("EnsureCompatible(6): %v", err)
	}
	if err := migrate.EnsureCompatible(context.Background(), db, 5); err == nil {
		t.Fatal("EnsureCompatible(5) succeeded")
	}
}
