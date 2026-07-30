package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestStoreCreatesAndLoadsSession(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), 24*time.Hour)

	created, rawToken, err := store.Create(context.Background())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rawToken == "" || created.CSRFToken == "" {
		t.Fatal("expected opaque token and csrf token")
	}

	loaded, err := store.Get(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if loaded.CSRFToken != created.CSRFToken {
		t.Fatalf("CSRFToken = %q, want %q", loaded.CSRFToken, created.CSRFToken)
	}
}

func TestStoreRejectsRevokedSession(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)

	created, rawToken, err := store.Create(context.Background())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.Revoke(context.Background(), created.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := store.Get(context.Background(), rawToken); err == nil {
		t.Fatal("Get succeeded for revoked session")
	}
}
