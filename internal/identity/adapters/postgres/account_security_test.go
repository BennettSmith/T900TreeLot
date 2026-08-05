package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestAccountSecurityPersistsStepUpAndPasskeyDelete(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	if _, err := db.Exec(context.Background(), `
		INSERT INTO passkey_credentials (
			id, identity_id, credential_id, public_key, attestation_type, aaguid, sign_count, created_at
		) VALUES
			('cred-1', 'identity-1', '\x01'::bytea, '\x02'::bytea, 'none', 'aaguid', 0, $1),
			('cred-2', 'identity-1', '\x03'::bytea, '\x04'::bytea, 'none', 'aaguid', 0, $1)
	`, now); err != nil {
		t.Fatal(err)
	}
	var sessionID int64
	if err := db.QueryRow(context.Background(), `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at)
		VALUES ('\x11'::bytea, 'csrf', $2, $1, 'identity-1', $1)
		RETURNING id
	`, now, now.Add(time.Hour)).Scan(&sessionID); err != nil {
		t.Fatal(err)
	}

	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		stepUpAt, err := repos.LoadSessionStepUp(ctx, sessionID, "identity-1")
		if err != nil || stepUpAt != nil {
			t.Fatalf("initial step-up = %v err=%v", stepUpAt, err)
		}
		if err := repos.MarkSessionStepUp(ctx, sessionID, "identity-1", now); err != nil {
			return err
		}
		stepUpAt, err = repos.LoadSessionStepUp(ctx, sessionID, "identity-1")
		if err != nil || stepUpAt == nil || !stepUpAt.Equal(now) {
			t.Fatalf("marked step-up = %v err=%v", stepUpAt, err)
		}
		if err := repos.StoreStepUpCeremony(ctx, application.AssertionCeremony{
			ID: "step-1", SessionID: sessionID, Challenge: []byte("challenge"), IdentityID: "identity-1",
			UserHandle: []byte("handle"), ExpiresAt: now.Add(time.Minute),
		}); err != nil {
			return err
		}
		ceremony, err := repos.LockStepUpCeremony(ctx, "step-1")
		if err != nil || ceremony.IdentityID != "identity-1" {
			t.Fatalf("lock step-up = %#v err=%v", ceremony, err)
		}
		if err := repos.ConsumeStepUpCeremony(ctx, "step-1", now); err != nil {
			return err
		}
		passkeys, err := repos.ListPasskeys(ctx, "identity-1")
		if err != nil || len(passkeys) != 2 {
			t.Fatalf("passkeys = %#v err=%v", passkeys, err)
		}
		identity, err := repos.LoadIdentity(ctx, "identity-1")
		if err != nil || identity.ID != "identity-1" || len(identity.Credentials) != 2 {
			t.Fatalf("identity = %#v err=%v", identity, err)
		}
		return repos.DeletePasskeyCredential(ctx, "identity-1", "cred-2")
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	var count int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM passkey_credentials WHERE identity_id = 'identity-1'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d", count)
	}
}

func TestAccountSecurityEmailReplaceAndConflict(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	seedIdentity(t, db, "identity-2", "person-2", "Other", "Person", "Other", "taken@example.org")

	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		taken, err := repos.EmailTaken(ctx, "taken@example.org")
		if err != nil || !taken {
			return fmt.Errorf("taken=%v err=%v", taken, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("email taken check: %v", err)
	}

	err = unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		return repos.ReplaceActiveEmail(ctx, "identity-1", "taken@example.org", "taken@example.org", now)
	})
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("conflict tx err = %v, want ErrEmailTaken", err)
	}

	err = unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		if err := repos.ReplaceActiveEmail(ctx, "identity-1", "fresh@example.org", "fresh@example.org", now); err != nil {
			return err
		}
		email, err := repos.ActiveEmail(ctx, "identity-1")
		if err != nil || email.Normalized() != "fresh@example.org" {
			t.Fatalf("replaced email = %#v err=%v", email, err)
		}
		var verified any
		if err := db.QueryRow(context.Background(), `
			SELECT verified_at FROM identity_emails WHERE identity_id = 'identity-1' AND active
		`).Scan(&verified); err != nil {
			t.Fatal(err)
		}
		if verified != nil {
			t.Fatalf("verified_at = %#v, want nil", verified)
		}
		return repos.RevokeSessionsForIdentity(ctx, "identity-1")
	})
	if err != nil {
		t.Fatalf("replace tx: %v", err)
	}
}

func TestAccountSecurityRegistrationCeremonyRoundTrip(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		if err := repos.StoreAccountRegistrationCeremony(ctx, application.AccountRegistrationCeremony{
			ID: "reg-1", SessionID: 0, IdentityID: "identity-1", Challenge: []byte("challenge"),
			UserHandle: []byte("handle"), ExpiresAt: now.Add(time.Minute),
		}); err != nil {
			return err
		}
		ceremony, err := repos.LockAccountRegistrationCeremony(ctx, "reg-1")
		if err != nil || ceremony.ID != "reg-1" || ceremony.IdentityID != "identity-1" {
			t.Fatalf("ceremony = %#v err=%v", ceremony, err)
		}
		return repos.ConsumeAccountRegistrationCeremony(ctx, "reg-1", now)
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAccountQueriesListsPasskeys(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	if _, err := db.Exec(context.Background(), `
		INSERT INTO passkey_credentials (
			id, identity_id, credential_id, public_key, attestation_type, aaguid, sign_count, created_at
		) VALUES ('cred-1', 'identity-1', '\x01'::bytea, '\x02'::bytea, 'none', 'aaguid', 0, $1)
	`, now); err != nil {
		t.Fatal(err)
	}
	passkeys, err := identitypostgres.NewAccountQueries(db).ListPasskeys(context.Background(), "identity-1")
	if err != nil || len(passkeys) != 1 || passkeys[0].ID != "cred-1" {
		t.Fatalf("passkeys = %#v err=%v", passkeys, err)
	}
}

func TestFixtureSeedsConflictingIdentity(t *testing.T) {
	db := testdb.OpenMigrated(t)
	unit := identitypostgres.NewUnitOfWork(db, clock.System(), session.TestKey)
	err := unit.WithinTestFixtureTx(context.Background(), func(ctx context.Context, repos application.TestFixtureRepositories) error {
		return repos.SeedConflictingIdentity(ctx, "conflict-id", "conflict-person", "taken@example.org")
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	var active bool
	if err := db.QueryRow(context.Background(), `
		SELECT active FROM identity_emails WHERE email_normalized = 'taken@example.org'
	`).Scan(&active); err != nil || !active {
		t.Fatalf("active=%v err=%v", active, err)
	}
}

func TestAccountSecurityNotFoundPaths(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	var sessionID int64
	if err := db.QueryRow(context.Background(), `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at)
		VALUES ('\x21'::bytea, 'csrf', $2, $1, 'identity-1', $1)
		RETURNING id
	`, now, now.Add(time.Hour)).Scan(&sessionID); err != nil {
		t.Fatal(err)
	}
	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		if _, err := repos.LoadSessionStepUp(ctx, 999, "identity-1"); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("LoadSessionStepUp err = %v", err)
		}
		if err := repos.MarkSessionStepUp(ctx, 999, "identity-1", now); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("MarkSessionStepUp err = %v", err)
		}
		if _, err := repos.LockStepUpCeremony(ctx, "missing"); err == nil {
			t.Fatal("expected missing step-up ceremony")
		}
		if err := repos.ConsumeStepUpCeremony(ctx, "missing", now); err == nil {
			t.Fatal("expected consume step-up failure")
		}
		if _, err := repos.LockAccountRegistrationCeremony(ctx, "missing"); err == nil {
			t.Fatal("expected missing registration ceremony")
		}
		if err := repos.ConsumeAccountRegistrationCeremony(ctx, "missing", now); err == nil {
			t.Fatal("expected consume registration failure")
		}
		if err := repos.DeletePasskeyCredential(ctx, "identity-1", "missing"); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("DeletePasskeyCredential err = %v", err)
		}
		if _, err := repos.ActiveEmail(ctx, "missing"); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("ActiveEmail err = %v", err)
		}
		if err := repos.ReplaceActiveEmail(ctx, "missing", "x@example.org", "x@example.org", now); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("ReplaceActiveEmail err = %v", err)
		}
		if err := repos.StoreStepUpCeremony(ctx, application.AssertionCeremony{
			ID: "step-nil-handle", SessionID: sessionID, Challenge: []byte("c"), IdentityID: "identity-1",
			ExpiresAt: now.Add(time.Minute),
		}); err != nil {
			return err
		}
		if err := repos.StoreAccountRegistrationCeremony(ctx, application.AccountRegistrationCeremony{
			ID: "reg-nil-handle", SessionID: sessionID, IdentityID: "identity-1", Challenge: []byte("c"),
			ExpiresAt: now.Add(time.Minute),
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAccountSecurityReplaceSameEmailWhenTakenIsNoop(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")
	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		return repos.ReplaceActiveEmail(ctx, "identity-1", "Ada@Example.org", "ada@example.org", now)
	})
	if err != nil {
		t.Fatalf("same email replace: %v", err)
	}
}

func TestAccountSecurityListPasskeysUnknownIdentity(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	err := unit.WithinAccountSecurityTx(context.Background(), func(ctx context.Context, repos application.AccountSecurityRepositories) error {
		_, err := repos.ListPasskeys(ctx, "missing")
		return err
	})
	if err == nil {
		t.Fatal("expected list passkeys failure")
	}
}

// TestConcurrentPasskeyRemovalCannotLeaveZeroCredentials proves UC-2B's last-passkey
// invariant holds when two removals race with two credentials.
func TestConcurrentPasskeyRemovalCannotLeaveZeroCredentials(t *testing.T) {
	db := testdb.OpenMigrated(t)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Ada", "Admin", "Ada", "ada@example.org")

	var sessionA, sessionB int64
	if err := db.QueryRow(context.Background(), `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at, step_up_at)
		VALUES ('\x31'::bytea, 'csrf-a', $2, $1, 'identity-1', $1, $1)
		RETURNING id
	`, now, now.Add(time.Hour)).Scan(&sessionA); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(context.Background(), `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at, step_up_at)
		VALUES ('\x32'::bytea, 'csrf-b', $2, $1, 'identity-1', $1, $1)
		RETURNING id
	`, now, now.Add(time.Hour)).Scan(&sessionB); err != nil {
		t.Fatal(err)
	}

	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now))
	service := &application.AccountSecurityService{
		UnitOfWork: unit,
		Clock:      clock.NewControllable(now),
		StepUpTTL:  time.Hour,
	}

	// Repeat to catch intermittent TOCTOU wins under the unlocked list-then-delete path.
	const rounds = 40
	for round := 0; round < rounds; round++ {
		if _, err := db.Exec(context.Background(), `
			DELETE FROM passkey_credentials WHERE identity_id = 'identity-1'
		`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(context.Background(), `
			INSERT INTO passkey_credentials (
				id, identity_id, credential_id, public_key, attestation_type, aaguid, sign_count, created_at
			) VALUES
				('cred-1', 'identity-1', '\x01'::bytea, '\x02'::bytea, 'none', 'aaguid', 0, $1),
				('cred-2', 'identity-1', '\x03'::bytea, '\x04'::bytea, 'none', 'aaguid', 0, $1)
		`, now); err != nil {
			t.Fatal(err)
		}

		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			errs[0] = service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
				IdentityID: "identity-1", SessionID: sessionA, CredentialID: "cred-1",
			})
		}()
		go func() {
			defer wg.Done()
			<-start
			errs[1] = service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
				IdentityID: "identity-1", SessionID: sessionB, CredentialID: "cred-2",
			})
		}()
		close(start)
		wg.Wait()

		var count int
		if err := db.QueryRow(context.Background(), `
			SELECT count(*) FROM passkey_credentials WHERE identity_id = 'identity-1'
		`).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count < 1 {
			t.Fatalf("round %d left zero passkeys (lockout); errs=%v %v", round, errs[0], errs[1])
		}
		if count != 1 {
			t.Fatalf("round %d count=%d, want 1; errs=%v %v", round, count, errs[0], errs[1])
		}

		lastPasskeyDenials := 0
		successes := 0
		for i, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, domain.ErrLastPasskey):
				lastPasskeyDenials++
			default:
				t.Fatalf("round %d errs[%d]=%v, want nil or ErrLastPasskey", round, i, err)
			}
		}
		if successes != 1 || lastPasskeyDenials != 1 {
			t.Fatalf("round %d successes=%d denials=%d errs=%v %v", round, successes, lastPasskeyDenials, errs[0], errs[1])
		}
	}
}
