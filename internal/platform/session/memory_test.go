package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
)

func TestMemoryStoreLifecycle(t *testing.T) {
	t.Parallel()

	controllable := clock.NewControllable(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	store := session.NewMemoryStore(controllable, 0)

	created, token, err := store.Create(context.Background())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	loaded, err := store.Get(context.Background(), token)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if loaded.CSRFToken != created.CSRFToken {
		t.Fatalf("csrf mismatch")
	}
	if err := store.Revoke(context.Background(), created.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := store.Get(context.Background(), token); err == nil {
		t.Fatal("expected revoked session to be missing")
	}
	if err := store.Revoke(context.Background(), 999); err == nil {
		t.Fatal("expected revoke of unknown session to fail")
	}
}

func TestMemoryStoreExpires(t *testing.T) {
	t.Parallel()

	controllable := clock.NewControllable(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	store := session.NewMemoryStore(controllable, time.Minute)
	_, token, err := store.Create(context.Background())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	controllable.Advance(2 * time.Minute)
	if _, err := store.Get(context.Background(), token); err == nil {
		t.Fatal("expected expired session to be missing")
	}
}
