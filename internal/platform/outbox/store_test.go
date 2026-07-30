package outbox_test

import (
	"context"
	"testing"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/outbox"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestStoreEnqueueAndStatus(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := outbox.NewStore(db, clock.System())

	if err := store.Enqueue(context.Background(), "", "groupsio"); err == nil {
		t.Fatal("expected empty key error")
	}
	if err := store.Enqueue(context.Background(), "key-a", ""); err == nil {
		t.Fatal("expected empty channel error")
	}
	if err := store.Enqueue(context.Background(), "key-a", "groupsio"); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if err := store.Enqueue(context.Background(), "key-a", "groupsio"); err != nil {
		t.Fatalf("duplicate Enqueue: %v", err)
	}
	message, err := store.Status(context.Background(), "key-a")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if message.Status != "pending" || message.Channel != "groupsio" {
		t.Fatalf("message = %+v", message)
	}
	if _, err := store.Status(context.Background(), "missing"); err == nil {
		t.Fatal("expected missing key error")
	}
}
