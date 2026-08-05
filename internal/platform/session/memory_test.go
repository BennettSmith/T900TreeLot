package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
)

func TestMemoryStoreMarkStepUp(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	created, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	if err := store.MarkStepUp(context.Background(), created.ID, at); err != nil {
		t.Fatalf("MarkStepUp: %v", err)
	}
	loaded, err := store.Get(context.Background(), rawToken)
	if err != nil || loaded.StepUpAt == nil || !loaded.StepUpAt.Equal(at) {
		t.Fatalf("loaded = %#v err=%v", loaded, err)
	}
	if err := store.MarkStepUp(context.Background(), 999, at); err != session.ErrNotFound {
		t.Fatalf("missing session err = %v", err)
	}
}

func TestMemoryStoreRevokeAndDefaults(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), 0, session.TestKey)
	created, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Revoke(context.Background(), created.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := store.Get(context.Background(), rawToken); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("revoked get err = %v, want ErrNotFound", err)
	}
	if _, err := store.Get(context.Background(), ""); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("empty token get err = %v, want ErrNotFound", err)
	}
	if _, err := store.Get(context.Background(), "missing-token"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("missing token get err = %v, want ErrNotFound", err)
	}
	if err := store.Revoke(context.Background(), 999); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("missing revoke err = %v, want ErrNotFound", err)
	}
	if _, _, err := store.CreateForIdentity(context.Background(), ""); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("empty identity err = %v, want ErrNotFound", err)
	}
}
