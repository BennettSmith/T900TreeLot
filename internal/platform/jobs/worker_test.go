package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/jobs"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestWorkerTickIsIdleWhenQueuesEmpty(t *testing.T) {
	db := testdb.OpenMigrated(t)
	worker := jobs.NewWorker(db, clock.System(), "worker-1")

	claimed, err := worker.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if claimed != 0 {
		t.Fatalf("claimed = %d, want 0", claimed)
	}
}

func TestWorkerClaimsOutboxMessage(t *testing.T) {
	db := testdb.OpenMigrated(t)
	_, err := db.Exec(context.Background(), `
		INSERT INTO outbox_messages (idempotency_key, channel, payload, available_at)
		VALUES ('key-1', 'groupsio', '{}'::jsonb, $1)
	`, time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatalf("insert outbox: %v", err)
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
	if err := db.QueryRow(context.Background(), `SELECT status FROM outbox_messages WHERE idempotency_key = 'key-1'`).Scan(&status); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("status = %q, want delivered", status)
	}
}
