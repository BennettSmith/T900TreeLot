package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/audit"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestWriterPersistsAuditEvent(t *testing.T) {
	db := testdb.OpenMigrated(t)
	writer := audit.NewWriter(db)
	now := time.Date(2026, 7, 31, 22, 30, 0, 0, time.UTC)

	err := writer.Write(context.Background(), audit.Event{
		ActorID:       "identity-1",
		Action:        "identity.bootstrap.completed",
		TargetType:    "identity",
		TargetID:      "identity-1",
		CorrelationID: "credential-1",
		Payload:       map[string]any{"email_normalized": "first@example.org"},
		CreatedAt:     now,
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var action, payload string
	if err := db.QueryRow(context.Background(), `
		SELECT action, payload->>'email_normalized'
		FROM audit_events
		WHERE correlation_id = 'credential-1'
	`).Scan(&action, &payload); err != nil {
		t.Fatalf("read audit row: %v", err)
	}
	if action != "identity.bootstrap.completed" || payload != "first@example.org" {
		t.Fatalf("audit row = %q/%q", action, payload)
	}
}

func TestWriterDefaultsEmptyPayloadAndTimestamp(t *testing.T) {
	db := testdb.OpenMigrated(t)
	writer := audit.NewWriter(db)

	if err := writer.Write(context.Background(), audit.Event{
		Action:        "identity.bootstrap.started",
		TargetType:    "bootstrap",
		TargetID:      "1",
		CorrelationID: "correlation-1",
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var payload string
	if err := db.QueryRow(context.Background(), `
		SELECT payload::text
		FROM audit_events
		WHERE correlation_id = 'correlation-1'
	`).Scan(&payload); err != nil {
		t.Fatalf("read audit row: %v", err)
	}
	if payload != "{}" {
		t.Fatalf("payload = %q, want {}", payload)
	}
}
