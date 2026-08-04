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

func TestSignOutTransactionRevokesOwnedSessionAndWritesAudit(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()
	now := time.Date(2026, 8, 4, 21, 0, 0, 0, time.UTC)
	seedIdentity(t, db, "identity-1", "person-1", "Signing", "Out", "", "signout@example.org")
	if _, err := db.Exec(ctx, `
		INSERT INTO sessions (id, token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at)
		VALUES (41, '\x01'::bytea, 'csrf', $1, $2, 'identity-1', $2)
	`, now.Add(time.Hour), now); err != nil {
		t.Fatal(err)
	}

	unit := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now))
	err := unit.WithinSignOutTx(ctx, func(txCtx context.Context, repos application.SignOutRepositories) error {
		if err := repos.RevokeCurrentSession(txCtx, 41, "identity-1", now); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, application.AuditEvent{
			ActorID: "identity-1", Action: "identity.session.signed_out",
			TargetType: "session", TargetID: "41", CorrelationID: "41",
			Payload: map[string]any{}, CreatedAt: now,
		})
	})
	if err != nil {
		t.Fatalf("sign-out transaction: %v", err)
	}

	var revokedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT revoked_at FROM sessions WHERE id = 41`).Scan(&revokedAt); err != nil {
		t.Fatal(err)
	}
	if revokedAt == nil || !revokedAt.Equal(now) {
		t.Fatalf("revoked_at = %v", revokedAt)
	}
	var audits int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = 'identity.session.signed_out' AND actor_id = 'identity-1'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 1 {
		t.Fatalf("audit count = %d", audits)
	}
}
