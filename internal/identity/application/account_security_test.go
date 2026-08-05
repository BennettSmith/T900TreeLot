package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

var accountNow = time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)

func TestChangeEmailRequiresRecentStepUp(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{email: "old@example.org"}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
		IdentityID: "identity-1",
		SessionID:  9,
		NewEmail:   "new@example.org",
	})
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("error = %v, want ErrStepUpRequired", err)
	}
	if repos.replacedEmail != "" {
		t.Fatalf("email replaced without step-up: %q", repos.replacedEmail)
	}
}

func TestChangeEmailReplacesEmailAndRevokesSessions(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		email:    "old@example.org",
		stepUpAt: &stepUp,
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
		IDs:        &fakeIDs{values: []string{"audit-1"}},
	}
	err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
		IdentityID: "identity-1",
		SessionID:  9,
		NewEmail:   " New@Example.org ",
	})
	if err != nil {
		t.Fatalf("ChangeEmail: %v", err)
	}
	if repos.replacedEmail != "New@Example.org" || repos.replacedNormalized != "new@example.org" {
		t.Fatalf("replaced = %q / %q", repos.replacedEmail, repos.replacedNormalized)
	}
	if !repos.revokedAll {
		t.Fatal("expected all sessions revoked")
	}
	if repos.lastAudit.Action != "identity.account_email.changed" {
		t.Fatalf("audit = %#v", repos.lastAudit)
	}
}

func TestChangeEmailRejectsTakenEmailWithoutChanging(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		email:      "old@example.org",
		stepUpAt:   &stepUp,
		emailTaken: true,
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
		IdentityID: "identity-1",
		SessionID:  9,
		NewEmail:   "taken@example.org",
	})
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("error = %v, want ErrEmailTaken", err)
	}
	if repos.replacedEmail != "" || repos.revokedAll {
		t.Fatal("taken email must not mutate identity")
	}
}

func TestRemovePasskeyRejectsLastCredential(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		stepUpAt: &stepUp,
		passkeys: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1")}},
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
		IdentityID:   "identity-1",
		SessionID:    9,
		CredentialID: "cred-1",
	})
	if !errors.Is(err, domain.ErrLastPasskey) {
		t.Fatalf("error = %v, want ErrLastPasskey", err)
	}
}

func TestRemovePasskeyDeletesWhenAnotherRemains(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		stepUpAt: &stepUp,
		passkeys: []application.PasskeyCredential{
			{ID: "cred-1", CredentialID: []byte("c1")},
			{ID: "cred-2", CredentialID: []byte("c2")},
		},
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	if err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
		IdentityID: "identity-1", SessionID: 9, CredentialID: "cred-1",
	}); err != nil {
		t.Fatalf("RemovePasskey: %v", err)
	}
	if repos.deletedCredentialID != "cred-1" {
		t.Fatalf("deleted = %q", repos.deletedCredentialID)
	}
}

func TestFinishStepUpMarksSession(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{
		identity: application.SignInIdentity{
			ID: "identity-1", UserHandle: []byte("user"),
			Credentials: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1"), PublicKey: []byte("pk")}},
		},
		ceremony: application.AssertionCeremony{
			ID: "ceremony-1", SessionID: 9, IdentityID: "identity-1", Challenge: []byte("challenge"),
			ExpiresAt: accountNow.Add(time.Minute),
		},
	}
	passkeys := &fakeAssertions{verified: application.VerifiedAssertion{CredentialID: []byte("c1"), SignCount: 2}}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Passkeys:   passkeys,
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
		IDs:        &fakeIDs{values: []string{}},
	}
	if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
		IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
	}); err != nil {
		t.Fatalf("FinishStepUp: %v", err)
	}
	if repos.markedStepUpSessionID != 9 || repos.markedStepUpAt == nil || !repos.markedStepUpAt.Equal(accountNow) {
		t.Fatalf("step-up mark = %d %#v", repos.markedStepUpSessionID, repos.markedStepUpAt)
	}
}

type fakeAccountSecurityUnitOfWork struct {
	repos *fakeAccountSecurityRepositories
}

func (u fakeAccountSecurityUnitOfWork) WithinAccountSecurityTx(ctx context.Context, fn func(context.Context, application.AccountSecurityRepositories) error) error {
	return fn(ctx, u.repos)
}

type fakeAccountSecurityRepositories struct {
	email                string
	emailTaken           bool
	stepUpAt             *time.Time
	passkeys             []application.PasskeyCredential
	identity             application.SignInIdentity
	ceremony             application.AssertionCeremony
	registrationCeremony  application.AccountRegistrationCeremony
	replacedEmail         string
	replacedNormalized    string
	revokedAll            bool
	deletedCredentialID   string
	markedStepUpSessionID int64
	markedStepUpAt        *time.Time
	storedCredential      application.PasskeyCredential
	lastAudit             application.AuditEvent
	storedCeremony        application.AssertionCeremony
}

func (r *fakeAccountSecurityRepositories) LoadSessionStepUp(_ context.Context, _ int64, _ string) (*time.Time, error) {
	return r.stepUpAt, nil
}

func (r *fakeAccountSecurityRepositories) MarkSessionStepUp(_ context.Context, sessionID int64, identityID string, at time.Time) error {
	if identityID == "" {
		return errors.New("missing identity")
	}
	r.markedStepUpSessionID = sessionID
	r.markedStepUpAt = &at
	r.stepUpAt = &at
	return nil
}

func (r *fakeAccountSecurityRepositories) LoadIdentity(_ context.Context, identityID string) (application.SignInIdentity, error) {
	if r.identity.ID == "" {
		r.identity.ID = identityID
	}
	if len(r.identity.Credentials) == 0 {
		r.identity.Credentials = r.passkeys
	}
	return r.identity, nil
}

func (r *fakeAccountSecurityRepositories) ListPasskeys(_ context.Context, _ string) ([]application.PasskeyCredential, error) {
	return r.passkeys, nil
}

func (r *fakeAccountSecurityRepositories) StoreStepUpCeremony(_ context.Context, ceremony application.AssertionCeremony) error {
	r.storedCeremony = ceremony
	return nil
}

func (r *fakeAccountSecurityRepositories) LockStepUpCeremony(_ context.Context, _ string) (application.AssertionCeremony, error) {
	return r.ceremony, nil
}

func (r *fakeAccountSecurityRepositories) ConsumeStepUpCeremony(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (r *fakeAccountSecurityRepositories) UpdatePasskeyAfterAssertion(_ context.Context, _ string, _ uint32, _ uint8, _ time.Time) error {
	return nil
}

func (r *fakeAccountSecurityRepositories) StoreAccountRegistrationCeremony(_ context.Context, ceremony application.AccountRegistrationCeremony) error {
	r.registrationCeremony = ceremony
	return nil
}

func (r *fakeAccountSecurityRepositories) LockAccountRegistrationCeremony(_ context.Context, _ string) (application.AccountRegistrationCeremony, error) {
	return r.registrationCeremony, nil
}

func (r *fakeAccountSecurityRepositories) ConsumeAccountRegistrationCeremony(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (r *fakeAccountSecurityRepositories) ActiveEmail(_ context.Context, _ string) (domain.Email, error) {
	return domain.NewEmail(r.email)
}

func (r *fakeAccountSecurityRepositories) StorePasskeyCredential(_ context.Context, credential application.PasskeyCredential) error {
	r.storedCredential = credential
	r.passkeys = append(r.passkeys, credential)
	return nil
}

func (r *fakeAccountSecurityRepositories) DeletePasskeyCredential(_ context.Context, identityID, credentialID string) error {
	if identityID == "" {
		return errors.New("missing identity")
	}
	r.deletedCredentialID = credentialID
	remaining := make([]application.PasskeyCredential, 0, len(r.passkeys))
	for _, passkey := range r.passkeys {
		if passkey.ID != credentialID {
			remaining = append(remaining, passkey)
		}
	}
	r.passkeys = remaining
	return nil
}

func (r *fakeAccountSecurityRepositories) EmailTaken(_ context.Context, _ string) (bool, error) {
	return r.emailTaken, nil
}

func (r *fakeAccountSecurityRepositories) ReplaceActiveEmail(_ context.Context, identityID, email, normalized string, _ time.Time) error {
	if identityID == "" {
		return errors.New("missing identity")
	}
	r.replacedEmail = email
	r.replacedNormalized = normalized
	r.email = email
	return nil
}

func (r *fakeAccountSecurityRepositories) RevokeSessionsForIdentity(_ context.Context, _ string) error {
	r.revokedAll = true
	return nil
}

func (r *fakeAccountSecurityRepositories) WriteAudit(_ context.Context, event application.AuditEvent) error {
	r.lastAudit = event
	return nil
}
