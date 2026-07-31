package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/application"
	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestAccountQueriesFindsDisplayNameAndEmail(t *testing.T) {
	db := testdb.OpenMigrated(t)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Commander", "ada@example.org")

	profile, err := identitypostgres.NewAccountQueries(db).FindAccountProfile(context.Background(), "identity-1")
	if err != nil {
		t.Fatalf("FindAccountProfile: %v", err)
	}
	if profile.IdentityID != "identity-1" || profile.DisplayName != "Commander" || profile.PrimaryEmail != "ada@example.org" {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestAccountQueriesReturnsNotFound(t *testing.T) {
	db := testdb.OpenMigrated(t)

	_, err := identitypostgres.NewAccountQueries(db).FindAccountProfile(context.Background(), "missing")
	if !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("err = %v, want ErrAccountNotFound", err)
	}
}

func TestBootstrapResetClearsIdentityStateAndRevokesSessions(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "", "ada@example.org")
	seedStatements := []struct {
		sql  string
		args []any
	}{
		{sql: `INSERT INTO identity_roles (identity_id, role, created_at) VALUES ('identity-1', 'admin', $1)`, args: []any{now}},
		{sql: `INSERT INTO passkey_credentials (id, identity_id, credential_id, public_key, attestation_type, aaguid, sign_count, created_at) VALUES ('credential-1', 'identity-1', '\x01'::bytea, '\x02'::bytea, 'packed', 'aaguid', 0, $1)`, args: []any{now}},
		{sql: `UPDATE bootstrap_state SET closed_at = $1, closed_by_identity_id = 'identity-1' WHERE id = 1`, args: []any{now}},
		{sql: `INSERT INTO rate_limit_buckets (bucket_key, window_started_at, count, updated_at) VALUES ('bootstrap:127.0.0.1', $1, 3, $1)`, args: []any{now}},
		{sql: `INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at) VALUES ('\x03'::bytea, 'csrf', $2, $1, 'identity-1', $1)`, args: []any{now, now.Add(time.Hour)}},
	}
	for _, statement := range seedStatements {
		if _, err := db.Exec(context.Background(), statement.sql, statement.args...); err != nil {
			t.Fatalf("seed reset data: %v", err)
		}
	}

	if err := identitypostgres.NewTestControl(db).ResetBootstrap(context.Background()); err != nil {
		t.Fatalf("ResetBootstrap: %v", err)
	}

	for table, want := range map[string]int{
		"people":              0,
		"identities":          0,
		"identity_emails":     0,
		"identity_roles":      0,
		"passkey_credentials": 0,
		"rate_limit_buckets":  0,
	} {
		var count int
		if err := db.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != want {
			t.Fatalf("%s count = %d, want %d", table, count, want)
		}
	}
	var closedAt *time.Time
	if err := db.QueryRow(context.Background(), `SELECT closed_at FROM bootstrap_state WHERE id = 1`).Scan(&closedAt); err != nil {
		t.Fatalf("read bootstrap state: %v", err)
	}
	if closedAt != nil {
		t.Fatal("bootstrap state is still closed")
	}
	var revokedAt *time.Time
	var identityID *string
	if err := db.QueryRow(context.Background(), `SELECT revoked_at, identity_id FROM sessions LIMIT 1`).Scan(&revokedAt, &identityID); err != nil {
		t.Fatalf("read session: %v", err)
	}
	if revokedAt == nil || identityID != nil {
		t.Fatalf("session revoked_at=%v identity_id=%v", revokedAt, identityID)
	}
}

func seedIdentity(t *testing.T, db *platformpostgres.DB, identityID, personID, first, last, preferred, email string) {
	t.Helper()
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	if _, err := db.Exec(context.Background(), `
		INSERT INTO people (id, first_name, last_name, preferred_display_name, created_at, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $5)
	`, personID, first, last, preferred, now); err != nil {
		t.Fatalf("seed person: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
		INSERT INTO identities (id, person_id, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
	`, identityID, personID, now); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
		INSERT INTO identity_emails (identity_id, email, email_normalized, active, created_at, updated_at)
		VALUES ($1, $2, lower($2), true, $3, $3)
	`, identityID, email, now); err != nil {
		t.Fatalf("seed email: %v", err)
	}
}
