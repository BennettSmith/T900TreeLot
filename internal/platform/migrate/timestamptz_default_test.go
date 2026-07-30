package migrate_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestTimestampDefaultsPreserveInstantAcrossSessionTimeZones(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()

	for _, zone := range []string{"UTC", "America/Los_Angeles", "Asia/Tokyo"} {
		t.Run(zone, func(t *testing.T) {
			tx, err := db.Begin(ctx)
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()

			if _, err := tx.Exec(ctx, `SET LOCAL TIME ZONE `+quoteLiteral(zone)); err != nil {
				t.Fatalf("SET LOCAL TIME ZONE: %v", err)
			}

			var before, created, after time.Time
			if err := tx.QueryRow(ctx, `SELECT now()`).Scan(&before); err != nil {
				t.Fatalf("read before: %v", err)
			}
			if err := tx.QueryRow(ctx, `
				INSERT INTO audit_events (action, target_type, target_id, correlation_id)
				VALUES ('test.default', 'fixture', '1', 'corr')
				RETURNING created_at
			`).Scan(&created); err != nil {
				t.Fatalf("insert: %v", err)
			}
			if err := tx.QueryRow(ctx, `SELECT now()`).Scan(&after); err != nil {
				t.Fatalf("read after: %v", err)
			}

			if created.Before(before.Add(-2*time.Second)) || created.After(after.Add(2*time.Second)) {
				t.Fatalf("created_at=%v outside [%v, %v] with TIME ZONE %s (likely timestamptz default shift)", created, before, after, zone)
			}
			skew := math.Abs(created.Sub(before).Seconds())
			if skew > 2 {
				t.Fatalf("created_at skewed by %.1fs from now() under TIME ZONE %s", skew, zone)
			}
		})
	}
}

func TestFoundationMigrationUsesNowForTimestamptzDefaults(t *testing.T) {
	t.Parallel()

	contents, err := readMigrationsFile(t)
	if err != nil {
		t.Fatal(err)
	}
	if containsAntiPattern(contents) {
		t.Fatal("migrations must use DEFAULT now() for TIMESTAMPTZ, not (now() AT TIME ZONE 'utc')")
	}
}

func quoteLiteral(value string) string {
	return "'" + value + "'"
}
