package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestStoreCreatesAndLoadsSession(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), 0)

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
	if _, err := store.Get(context.Background(), rawToken); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get revoked error = %v, want ErrNotFound", err)
	}
}

func TestStoreGetUnknownTokenReturnsErrNotFound(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)

	if _, err := store.Get(context.Background(), ""); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get empty token error = %v, want ErrNotFound", err)
	}
	if _, err := store.Get(context.Background(), "missing-token"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get missing token error = %v, want ErrNotFound", err)
	}
}

func TestStoreGetPropagatesDatabaseErrors(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)

	_, rawToken, err := store.Create(context.Background())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err = store.Get(context.Background(), rawToken)
	if err == nil {
		t.Fatal("Get succeeded after database close")
	}
	if errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get mapped infrastructure failure to ErrNotFound: %v", err)
	}
}

func TestStoreRotatesSessionForIdentity(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)
	ctx := context.Background()

	seedIdentity(t, db, "identity-1")
	created, oldToken, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rotated, newToken, err := store.RotateForIdentity(ctx, created.ID, "identity-1")
	if err != nil {
		t.Fatalf("RotateForIdentity: %v", err)
	}
	if newToken == "" || newToken == oldToken {
		t.Fatal("expected a new raw session token")
	}
	if rotated.IdentityID != "identity-1" || rotated.AuthenticatedAt == nil {
		t.Fatalf("rotated = %#v", rotated)
	}
	if _, err := store.Get(ctx, oldToken); err == nil {
		t.Fatal("old token still loads after rotation")
	}
	loaded, err := store.Get(ctx, newToken)
	if err != nil {
		t.Fatalf("Get rotated: %v", err)
	}
	if loaded.IdentityID != "identity-1" {
		t.Fatalf("loaded.IdentityID = %q", loaded.IdentityID)
	}
}

func TestStoreCreatesAndBindsIdentitySessions(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)
	ctx := context.Background()
	seedIdentity(t, db, "identity-1")

	authenticated, rawToken, err := store.CreateForIdentity(ctx, "identity-1")
	if err != nil {
		t.Fatalf("CreateForIdentity: %v", err)
	}
	if authenticated.IdentityID != "identity-1" || authenticated.AuthenticatedAt == nil {
		t.Fatalf("authenticated session = %#v", authenticated)
	}
	loaded, err := store.Get(ctx, rawToken)
	if err != nil {
		t.Fatalf("Get authenticated: %v", err)
	}
	if loaded.IdentityID != "identity-1" {
		t.Fatalf("loaded.IdentityID = %q", loaded.IdentityID)
	}

	anonymous, _, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create anonymous: %v", err)
	}
	if err := store.BindIdentity(ctx, anonymous.ID, "identity-1"); err != nil {
		t.Fatalf("BindIdentity: %v", err)
	}
}

func TestStoreIdentitySessionValidationErrors(t *testing.T) {
	db := testdb.OpenMigrated(t)
	store := session.NewStore(db, clock.System(), time.Hour)
	ctx := context.Background()

	if _, _, err := store.CreateForIdentity(ctx, ""); err == nil {
		t.Fatal("CreateForIdentity succeeded with empty identity")
	}
	if err := store.BindIdentity(ctx, 999, "identity-1"); err != session.ErrNotFound {
		t.Fatalf("BindIdentity missing error = %v, want ErrNotFound", err)
	}
	if _, _, err := store.RotateForIdentity(ctx, 999, "identity-1"); err != session.ErrNotFound {
		t.Fatalf("RotateForIdentity missing error = %v, want ErrNotFound", err)
	}
	if err := store.Revoke(ctx, 999); err != session.ErrNotFound {
		t.Fatalf("Revoke missing error = %v, want ErrNotFound", err)
	}
}

func seedIdentity(t *testing.T, db *postgres.DB, identityID string) {
	t.Helper()
	personID := "person-" + identityID
	if _, err := db.Exec(context.Background(), `
		INSERT INTO people (id, first_name, last_name, created_at, updated_at)
		VALUES ($1, 'First', 'Admin', now(), now())
	`, personID); err != nil {
		t.Fatalf("seed person: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
		INSERT INTO identities (id, person_id, created_at, updated_at)
		VALUES ($1, $2, now(), now())
	`, identityID, personID); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
}
