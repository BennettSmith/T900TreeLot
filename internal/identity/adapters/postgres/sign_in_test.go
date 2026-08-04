package postgres_test

import (
	"context"
	"testing"
	"time"

	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestSignInTransactionPersistsAssertionAndRotatesSession(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()
	now := time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Family", "Manager", "", "manager@example.org")
	if _, err := db.Exec(ctx, `UPDATE identities SET webauthn_user_handle = '\x75736572'::bytea WHERE id = 'identity-1'`); err != nil {
		t.Fatalf("seed user handle: %v", err)
	}
	for _, statement := range []string{
		`INSERT INTO identity_roles (identity_id, role, created_at) VALUES ('identity-1', 'family_manager', now())`,
		`INSERT INTO passkey_credentials (id, identity_id, credential_id, public_key, attestation_type, aaguid, sign_count, transports, created_at) VALUES ('credential-1', 'identity-1', '\x63726564'::bytea, '\x707562'::bytea, 'none', '', 0, '{}', now())`,
		`INSERT INTO sessions (id, token_hash, csrf_token, expires_at, last_seen_at) VALUES (41, '\x01'::bytea, 'csrf', now() + interval '1 hour', now())`,
	} {
		if _, err := db.Exec(ctx, statement); err != nil {
			t.Fatalf("seed sign-in state: %v", err)
		}
	}

	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now))
	err := unit.WithinSignInTx(ctx, func(txCtx context.Context, repos application.SignInRepositories) error {
		identity, err := repos.FindSignInIdentityByEmail(txCtx, "manager@example.org")
		if err != nil {
			return err
		}
		if identity.ID != "identity-1" || string(identity.UserHandle) != "user" || len(identity.Credentials) != 1 {
			t.Fatalf("identity = %#v", identity)
		}
		if err := repos.StoreAssertionCeremony(txCtx, application.AssertionCeremony{
			ID:         "assertion-1",
			SessionID:  41,
			Challenge:  []byte("challenge"),
			IdentityID: identity.ID,
			UserHandle: identity.UserHandle,
			ExpiresAt:  now.Add(time.Minute),
		}); err != nil {
			return err
		}
		if _, err := repos.LockAssertionCeremony(txCtx, "assertion-1"); err != nil {
			return err
		}
		loaded, err := repos.FindSignInIdentityByCredential(txCtx, []byte("cred"))
		if err != nil || loaded.ID != identity.ID {
			t.Fatalf("credential identity = %#v err=%v", loaded, err)
		}
		if err := repos.UpdatePasskeyAfterAssertion(txCtx, "credential-1", 1, 12, now); err != nil {
			return err
		}
		if err := repos.ConsumeAssertionCeremony(txCtx, "assertion-1", now); err != nil {
			return err
		}
		if err := repos.WriteAudit(txCtx, application.AuditEvent{
			ActorID: "identity-1", Action: "identity.sign_in.completed", TargetType: "identity",
			TargetID: "identity-1", CorrelationID: "credential-1", Payload: map[string]any{}, CreatedAt: now,
		}); err != nil {
			return err
		}
		_, err = repos.RotateForIdentity(txCtx, 41, "identity-1", now)
		return err
	})
	if err != nil {
		t.Fatalf("sign-in transaction: %v", err)
	}

	var signCount, authenticatorFlags int
	if err := db.QueryRow(ctx, `SELECT sign_count, authenticator_flags FROM passkey_credentials WHERE id = 'credential-1'`).Scan(&signCount, &authenticatorFlags); err != nil {
		t.Fatal(err)
	}
	if signCount != 1 || authenticatorFlags != 12 {
		t.Fatalf("sign count/flags = %d/%d", signCount, authenticatorFlags)
	}
	var consumedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT consumed_at FROM webauthn_ceremonies WHERE id = 'assertion-1'`).Scan(&consumedAt); err != nil {
		t.Fatal(err)
	}
	if consumedAt == nil {
		t.Fatal("assertion ceremony was not consumed")
	}
}
