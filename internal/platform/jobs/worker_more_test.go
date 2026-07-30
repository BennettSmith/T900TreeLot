package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/jobs"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestWorkerClaimsBackgroundJob(t *testing.T) {
	db := testdb.OpenMigrated(t)
	_, err := db.Exec(context.Background(), `
		INSERT INTO background_jobs (job_type, payload, available_at)
		VALUES ('enqueue-reminders', '{}'::jsonb, $1)
	`, time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}

	worker := jobs.NewWorker(db, clock.System(), "worker-1")
	claimed, err := worker.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if claimed != 1 {
		t.Fatalf("claimed = %d, want 1", claimed)
	}

	var status string
	if err := db.QueryRow(context.Background(), `SELECT status FROM background_jobs LIMIT 1`).Scan(&status); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if status != "completed" {
		t.Fatalf("status = %q, want completed", status)
	}
}
