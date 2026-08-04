package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
)

func TestSignOutRevokesCurrentSessionAndAuditsIdentity(t *testing.T) {
	repos := &fakeSignOutRepositories{}
	service := application.SignOutService{
		UnitOfWork: fakeSignOutUnitOfWork{repos: repos},
		Clock:      fakeClock{now: signInNow},
	}

	err := service.SignOut(context.Background(), application.SignOutCommand{
		IdentityID: "identity-1",
		SessionID:  41,
	})
	if err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if repos.sessionID != 41 || repos.identityID != "identity-1" || !repos.revokedAt.Equal(signInNow) {
		t.Fatalf("revocation = %#v", repos)
	}
	if len(repos.audit) != 1 || repos.audit[0].Action != "identity.session.signed_out" ||
		repos.audit[0].ActorID != "identity-1" || len(repos.audit[0].Payload) != 0 {
		t.Fatalf("audit = %#v", repos.audit)
	}
}

func TestSignOutRejectsMissingAuthenticatedSession(t *testing.T) {
	service := application.SignOutService{}
	if err := service.SignOut(context.Background(), application.SignOutCommand{}); err != application.ErrAccountNotFound {
		t.Fatalf("error = %v, want ErrAccountNotFound", err)
	}
}

type fakeSignOutUnitOfWork struct {
	repos application.SignOutRepositories
}

func (u fakeSignOutUnitOfWork) WithinSignOutTx(ctx context.Context, fn func(context.Context, application.SignOutRepositories) error) error {
	return fn(ctx, u.repos)
}

type fakeSignOutRepositories struct {
	sessionID  int64
	identityID string
	revokedAt  time.Time
	audit      []application.AuditEvent
}

func (r *fakeSignOutRepositories) RevokeCurrentSession(_ context.Context, sessionID int64, identityID string, revokedAt time.Time) error {
	r.sessionID = sessionID
	r.identityID = identityID
	r.revokedAt = revokedAt
	return nil
}

func (r *fakeSignOutRepositories) WriteAudit(_ context.Context, event application.AuditEvent) error {
	r.audit = append(r.audit, event)
	return nil
}
