package postgres_test

import (
	"context"
	"testing"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestDBHelpers(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()

	if _, err := db.Exec(ctx, `SELECT 1`); err != nil {
		t.Fatalf("Exec: %v", err)
	}
	var value int
	if err := db.QueryRow(ctx, `SELECT 1`).Scan(&value); err != nil || value != 1 {
		t.Fatalf("QueryRow: value=%d err=%v", value, err)
	}
	rows, err := db.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("expected migration row")
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if db.Pool() == nil {
		t.Fatal("Pool is nil")
	}
}
