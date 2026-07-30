package migrate_test

import (
	"context"
	"testing"

	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestEnsureCompatibleDoesNotMutateSchema(t *testing.T) {
	db := testdb.OpenEmpty(t)
	ctx := context.Background()

	err := migrate.EnsureCompatible(ctx, db, 1)
	if err == nil {
		t.Fatal("EnsureCompatible succeeded on an empty database")
	}

	var exists bool
	if scanErr := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&exists); scanErr != nil {
		t.Fatalf("probe schema_migrations: %v", scanErr)
	}
	if exists {
		t.Fatal("EnsureCompatible created schema_migrations; only migrate may mutate schema")
	}
}

func TestCurrentVersionIsReadOnlyWhenBookkeepingMissing(t *testing.T) {
	db := testdb.OpenEmpty(t)
	ctx := context.Background()

	version, err := migrate.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("CurrentVersion: %v", err)
	}
	if version != 0 {
		t.Fatalf("version = %d, want 0 when schema_migrations is absent", version)
	}

	var exists bool
	if scanErr := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&exists); scanErr != nil {
		t.Fatalf("probe schema_migrations: %v", scanErr)
	}
	if exists {
		t.Fatal("CurrentVersion created schema_migrations")
	}
}
