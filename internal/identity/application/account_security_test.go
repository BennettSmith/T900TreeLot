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
	email                  string
	emailTaken             bool
	stepUpAt               *time.Time
	passkeys               []application.PasskeyCredential
	identity               application.SignInIdentity
	ceremony               application.AssertionCeremony
	registrationCeremony   application.AccountRegistrationCeremony
	replacedEmail          string
	replacedNormalized     string
	revokedAll             bool
	deletedCredentialID    string
	markedStepUpSessionID  int64
	markedStepUpAt         *time.Time
	storedCredential       application.PasskeyCredential
	lastAudit              application.AuditEvent
	storedCeremony         application.AssertionCeremony
	loadStepUpErr          error
	markStepUpErr          error
	loadIdentityErr        error
	listPasskeysErr        error
	storeCeremonyErr       error
	lockCeremonyErr        error
	consumeCeremonyErr     error
	updatePasskeyErr       error
	storeRegistrationErr   error
	lockRegistrationErr    error
	consumeRegistrationErr error
	activeEmailErr         error
	storeCredentialErr     error
	deletePasskeyErr       error
	emailTakenErr          error
	replaceEmailErr        error
	revokeErr              error
	writeAuditErr          error
}

func (r *fakeAccountSecurityRepositories) LoadSessionStepUp(_ context.Context, _ int64, _ string) (*time.Time, error) {
	if r.loadStepUpErr != nil {
		return nil, r.loadStepUpErr
	}
	return r.stepUpAt, nil
}

func (r *fakeAccountSecurityRepositories) MarkSessionStepUp(_ context.Context, sessionID int64, identityID string, at time.Time) error {
	if r.markStepUpErr != nil {
		return r.markStepUpErr
	}
	if identityID == "" {
		return errors.New("missing identity")
	}
	r.markedStepUpSessionID = sessionID
	r.markedStepUpAt = &at
	r.stepUpAt = &at
	return nil
}

func (r *fakeAccountSecurityRepositories) LoadIdentity(_ context.Context, identityID string) (application.SignInIdentity, error) {
	if r.loadIdentityErr != nil {
		return application.SignInIdentity{}, r.loadIdentityErr
	}
	if r.identity.ID == "" {
		r.identity.ID = identityID
	}
	if len(r.identity.Credentials) == 0 {
		r.identity.Credentials = r.passkeys
	}
	return r.identity, nil
}

func (r *fakeAccountSecurityRepositories) ListPasskeys(_ context.Context, _ string) ([]application.PasskeyCredential, error) {
	if r.listPasskeysErr != nil {
		return nil, r.listPasskeysErr
	}
	return r.passkeys, nil
}

func (r *fakeAccountSecurityRepositories) StoreStepUpCeremony(_ context.Context, ceremony application.AssertionCeremony) error {
	if r.storeCeremonyErr != nil {
		return r.storeCeremonyErr
	}
	r.storedCeremony = ceremony
	return nil
}

func (r *fakeAccountSecurityRepositories) LockStepUpCeremony(_ context.Context, _ string) (application.AssertionCeremony, error) {
	if r.lockCeremonyErr != nil {
		return application.AssertionCeremony{}, r.lockCeremonyErr
	}
	return r.ceremony, nil
}

func (r *fakeAccountSecurityRepositories) ConsumeStepUpCeremony(_ context.Context, _ string, _ time.Time) error {
	return r.consumeCeremonyErr
}

func (r *fakeAccountSecurityRepositories) UpdatePasskeyAfterAssertion(_ context.Context, _ string, _ uint32, _ uint8, _ time.Time) error {
	return r.updatePasskeyErr
}

func (r *fakeAccountSecurityRepositories) StoreAccountRegistrationCeremony(_ context.Context, ceremony application.AccountRegistrationCeremony) error {
	if r.storeRegistrationErr != nil {
		return r.storeRegistrationErr
	}
	r.registrationCeremony = ceremony
	return nil
}

func (r *fakeAccountSecurityRepositories) LockAccountRegistrationCeremony(_ context.Context, _ string) (application.AccountRegistrationCeremony, error) {
	if r.lockRegistrationErr != nil {
		return application.AccountRegistrationCeremony{}, r.lockRegistrationErr
	}
	return r.registrationCeremony, nil
}

func (r *fakeAccountSecurityRepositories) ConsumeAccountRegistrationCeremony(_ context.Context, _ string, _ time.Time) error {
	return r.consumeRegistrationErr
}

func (r *fakeAccountSecurityRepositories) ActiveEmail(_ context.Context, _ string) (domain.Email, error) {
	if r.activeEmailErr != nil {
		return domain.Email{}, r.activeEmailErr
	}
	return domain.NewEmail(r.email)
}

func (r *fakeAccountSecurityRepositories) StorePasskeyCredential(_ context.Context, credential application.PasskeyCredential) error {
	if r.storeCredentialErr != nil {
		return r.storeCredentialErr
	}
	r.storedCredential = credential
	r.passkeys = append(r.passkeys, credential)
	return nil
}

func (r *fakeAccountSecurityRepositories) DeletePasskeyCredential(_ context.Context, identityID, credentialID string) error {
	if r.deletePasskeyErr != nil {
		return r.deletePasskeyErr
	}
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
	if r.emailTakenErr != nil {
		return false, r.emailTakenErr
	}
	return r.emailTaken, nil
}

func (r *fakeAccountSecurityRepositories) ReplaceActiveEmail(_ context.Context, identityID, email, normalized string, _ time.Time) error {
	if r.replaceEmailErr != nil {
		return r.replaceEmailErr
	}
	if identityID == "" {
		return errors.New("missing identity")
	}
	r.replacedEmail = email
	r.replacedNormalized = normalized
	r.email = email
	return nil
}

func (r *fakeAccountSecurityRepositories) RevokeSessionsForIdentity(_ context.Context, _ string) error {
	if r.revokeErr != nil {
		return r.revokeErr
	}
	r.revokedAll = true
	return nil
}

func (r *fakeAccountSecurityRepositories) WriteAudit(_ context.Context, event application.AuditEvent) error {
	if r.writeAuditErr != nil {
		return r.writeAuditErr
	}
	r.lastAudit = event
	return nil
}

func TestBeginStepUpStoresIdentityBoundCeremony(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{
		identity: application.SignInIdentity{
			ID: "identity-1", UserHandle: []byte("user"),
			Credentials: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1"), PublicKey: []byte("pk")}},
		},
	}
	passkeys := &fakeAssertions{}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Passkeys:   passkeys,
		Clock:      fakeClock{now: accountNow},
		IDs:        &fakeIDs{values: []string{"ceremony-1"}},
		StepUpTTL:  5 * time.Minute,
	}
	result, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{
		IdentityID: "identity-1", SessionID: 9,
	})
	if err != nil {
		t.Fatalf("BeginStepUp: %v", err)
	}
	if result.CeremonyID != "ceremony-1" || repos.storedCeremony.IdentityID != "identity-1" {
		t.Fatalf("result=%#v ceremony=%#v", result, repos.storedCeremony)
	}
}

func TestBeginAndFinishAddPasskey(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		email:    "ada@example.org",
		stepUpAt: &stepUp,
		identity: application.SignInIdentity{
			ID: "identity-1", UserHandle: []byte("user"),
			Credentials: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1"), PublicKey: []byte("pk")}},
		},
		passkeys: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1")}},
	}
	registration := &fakeAccountRegistration{
		options: application.RegistrationOptions{
			CeremonyID: "reg-1", PublicKey: []byte("options"), Challenge: []byte("challenge"), ExpiresAt: accountNow.Add(time.Minute),
		},
		credential: application.PasskeyCredential{CredentialID: []byte("c2"), PublicKey: []byte("pk2"), AttestationType: "none"},
	}
	service := application.AccountSecurityService{
		UnitOfWork:   fakeAccountSecurityUnitOfWork{repos: repos},
		Registration: registration,
		Clock:        fakeClock{now: accountNow},
		IDs:          &fakeIDs{values: []string{"reg-1", "cred-2"}},
		StepUpTTL:    5 * time.Minute,
	}
	options, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{
		IdentityID: "identity-1", SessionID: 9,
	})
	if err != nil {
		t.Fatalf("BeginAddPasskey: %v", err)
	}
	if options.CeremonyID != "reg-1" || repos.registrationCeremony.IdentityID != "identity-1" {
		t.Fatalf("options=%#v ceremony=%#v", options, repos.registrationCeremony)
	}
	if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
		IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
	}); err != nil {
		t.Fatalf("FinishAddPasskey: %v", err)
	}
	if repos.storedCredential.ID != "cred-2" || repos.storedCredential.IdentityID != "identity-1" {
		t.Fatalf("stored = %#v", repos.storedCredential)
	}
}

type fakeAccountRegistration struct {
	options    application.RegistrationOptions
	credential application.PasskeyCredential
	beginErr   error
	verifyErr  error
}

func (f *fakeAccountRegistration) BeginRegistration(context.Context, application.AccountRegistrationStart) (application.RegistrationOptions, error) {
	if f.beginErr != nil {
		return application.RegistrationOptions{}, f.beginErr
	}
	return f.options, nil
}

func (f *fakeAccountRegistration) VerifyRegistration(context.Context, application.RegistrationVerification) (application.PasskeyCredential, error) {
	if f.verifyErr != nil {
		return application.PasskeyCredential{}, f.verifyErr
	}
	return f.credential, nil
}

func TestBeginAddPasskeyRequiresStepUp(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{
		email:    "ada@example.org",
		identity: application.SignInIdentity{ID: "identity-1", UserHandle: []byte("user")},
	}
	service := application.AccountSecurityService{
		UnitOfWork:   fakeAccountSecurityUnitOfWork{repos: repos},
		Registration: &fakeAccountRegistration{},
		Clock:        fakeClock{now: accountNow},
		IDs:          &fakeIDs{values: []string{"reg-1"}},
		StepUpTTL:    5 * time.Minute,
	}
	_, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{
		IdentityID: "identity-1", SessionID: 9,
	})
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestChangeEmailSameAddressIsNoop(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{email: "ada@example.org", stepUpAt: &stepUp}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
		IdentityID: "identity-1", SessionID: 9, NewEmail: "Ada@Example.org",
	}); err != nil {
		t.Fatalf("ChangeEmail: %v", err)
	}
	if repos.replacedEmail != "" || repos.revokedAll {
		t.Fatal("same email should not replace or revoke")
	}
}

func TestAccountSecurityDefaultsTTLAndClock(t *testing.T) {
	stepUp := time.Now().UTC().Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{email: "ada@example.org", stepUpAt: &stepUp}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
	}
	if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
		IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
	}); err != nil {
		t.Fatalf("ChangeEmail: %v", err)
	}
}

func TestFinishStepUpRejectsMismatchedCeremony(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{
		identity: application.SignInIdentity{
			ID: "identity-1", UserHandle: []byte("user"),
			Credentials: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1")}},
		},
		ceremony: application.AssertionCeremony{
			ID: "ceremony-1", SessionID: 8, IdentityID: "other", Challenge: []byte("challenge"),
			ExpiresAt: accountNow.Add(time.Minute),
		},
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Passkeys:   &fakeAssertions{verified: application.VerifiedAssertion{CredentialID: []byte("c1")}},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
		IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("err = %v", err)
	}
}

func TestFinishAddPasskeyRequiresStepUpAndValidCeremony(t *testing.T) {
	repos := &fakeAccountSecurityRepositories{email: "ada@example.org"}
	service := application.AccountSecurityService{
		UnitOfWork:   fakeAccountSecurityUnitOfWork{repos: repos},
		Registration: &fakeAccountRegistration{},
		Clock:        fakeClock{now: accountNow},
		IDs:          &fakeIDs{values: []string{"cred-2"}},
		StepUpTTL:    5 * time.Minute,
	}
	err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
		IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestRemovePasskeyRejectsUnknownCredential(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	repos := &fakeAccountSecurityRepositories{
		stepUpAt: &stepUp,
		passkeys: []application.PasskeyCredential{{ID: "cred-1"}, {ID: "cred-2"}},
	}
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
		Clock:      fakeClock{now: accountNow},
		StepUpTTL:  5 * time.Minute,
	}
	err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
		IdentityID: "identity-1", SessionID: 9, CredentialID: "missing",
	})
	if !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestBeginStepUpRequiresIdentity(t *testing.T) {
	service := application.AccountSecurityService{
		UnitOfWork: fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{}},
		IDs:        &fakeIDs{values: []string{"c"}},
	}
	_, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{})
	if !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestAccountSecurityErrorPaths(t *testing.T) {
	stepUp := accountNow.Add(-time.Minute)
	baseIdentity := application.SignInIdentity{
		ID: "identity-1", UserHandle: []byte("user"),
		Credentials: []application.PasskeyCredential{{ID: "cred-1", CredentialID: []byte("c1"), PublicKey: []byte("pk")}},
	}
	validCeremony := application.AssertionCeremony{
		ID: "ceremony-1", SessionID: 9, IdentityID: "identity-1", Challenge: []byte("challenge"),
		ExpiresAt: accountNow.Add(time.Minute),
	}
	boom := errors.New("boom")

	t.Run("begin step-up id failure", func(t *testing.T) {
		service := application.AccountSecurityService{
			UnitOfWork: fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{}},
			IDs:        &fakeIDs{err: boom},
		}
		_, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{IdentityID: "identity-1", SessionID: 9})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("begin step-up load and assertion failures", func(t *testing.T) {
		repos := &fakeAccountSecurityRepositories{loadIdentityErr: boom}
		service := application.AccountSecurityService{
			UnitOfWork: fakeAccountSecurityUnitOfWork{repos: repos},
			Passkeys:   &fakeAssertions{},
			Clock:      fakeClock{now: accountNow},
			IDs:        &fakeIDs{values: []string{"ceremony-1"}},
		}
		if _, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, boom) {
			t.Fatalf("load err = %v", err)
		}
		repos.loadIdentityErr = nil
		repos.identity = baseIdentity
		service.IDs = &fakeIDs{values: []string{"ceremony-2", "ceremony-3"}}
		service.Passkeys = &fakeAssertions{beginErr: boom}
		if _, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("assertion err = %v", err)
		}
		service.Passkeys = &fakeAssertions{}
		repos.storeCeremonyErr = boom
		if _, err := service.BeginStepUp(context.Background(), application.BeginStepUpCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, boom) {
			t.Fatalf("store err = %v", err)
		}
	})

	t.Run("finish step-up validation failures", func(t *testing.T) {
		service := application.AccountSecurityService{
			UnitOfWork: fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{}},
			Clock:      fakeClock{now: accountNow},
		}
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("empty command err = %v", err)
		}
		repos := &fakeAccountSecurityRepositories{lockCeremonyErr: boom}
		service.UnitOfWork = fakeAccountSecurityUnitOfWork{repos: repos}
		service.Passkeys = &fakeAssertions{verified: application.VerifiedAssertion{CredentialID: []byte("c1")}}
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("lock err = %v", err)
		}
		repos.lockCeremonyErr = nil
		repos.ceremony = application.AssertionCeremony{
			ID: "ceremony-1", SessionID: 9, IdentityID: "identity-1", Challenge: []byte("challenge"),
			ExpiresAt: accountNow.Add(-time.Minute),
		}
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("expired err = %v", err)
		}
		repos.ceremony = validCeremony
		repos.loadIdentityErr = boom
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("identity err = %v", err)
		}
		repos.loadIdentityErr = nil
		repos.identity = baseIdentity
		service.Passkeys = &fakeAssertions{verifyErr: boom}
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("verify err = %v", err)
		}
		service.Passkeys = &fakeAssertions{verified: application.VerifiedAssertion{CredentialID: []byte("missing")}}
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("credential err = %v", err)
		}
		service.Passkeys = &fakeAssertions{verified: application.VerifiedAssertion{CredentialID: []byte("c1")}}
		repos.updatePasskeyErr = boom
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("update err = %v", err)
		}
		repos.updatePasskeyErr = nil
		repos.consumeCeremonyErr = boom
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("consume err = %v", err)
		}
		repos.consumeCeremonyErr = nil
		repos.markStepUpErr = boom
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("mark err = %v", err)
		}
		repos.markStepUpErr = nil
		repos.writeAuditErr = boom
		if err := service.FinishStepUp(context.Background(), application.FinishStepUpCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "ceremony-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("audit err = %v", err)
		}
	})

	t.Run("begin add passkey failures", func(t *testing.T) {
		service := application.AccountSecurityService{
			UnitOfWork:   fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{stepUpAt: &stepUp}},
			Registration: &fakeAccountRegistration{},
			Clock:        fakeClock{now: accountNow},
			IDs:          &fakeIDs{err: boom},
			StepUpTTL:    5 * time.Minute,
		}
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{}); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("empty err = %v", err)
		}
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{IdentityID: "identity-1", SessionID: 9}); err == nil {
			t.Fatal("expected id error")
		}
		repos := &fakeAccountSecurityRepositories{stepUpAt: &stepUp, loadIdentityErr: boom, email: "ada@example.org"}
		service.UnitOfWork = fakeAccountSecurityUnitOfWork{repos: repos}
		service.IDs = &fakeIDs{values: []string{"reg-1"}}
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, boom) {
			t.Fatalf("load err = %v", err)
		}
		repos.loadIdentityErr = nil
		repos.identity = baseIdentity
		repos.activeEmailErr = boom
		service.IDs = &fakeIDs{values: []string{"reg-2", "reg-3", "reg-4"}}
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, boom) {
			t.Fatalf("email err = %v", err)
		}
		repos.activeEmailErr = nil
		service.Registration = &fakeAccountRegistration{beginErr: boom}
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("begin reg err = %v", err)
		}
		service.Registration = &fakeAccountRegistration{options: application.RegistrationOptions{
			CeremonyID: "reg-1", PublicKey: []byte("options"), Challenge: []byte("challenge"), ExpiresAt: accountNow.Add(time.Minute),
		}}
		repos.storeRegistrationErr = boom
		if _, err := service.BeginAddPasskey(context.Background(), application.BeginAddPasskeyCommand{IdentityID: "identity-1", SessionID: 9}); !errors.Is(err, boom) {
			t.Fatalf("store reg err = %v", err)
		}
	})

	t.Run("finish add passkey failures", func(t *testing.T) {
		service := application.AccountSecurityService{
			UnitOfWork:   fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{stepUpAt: &stepUp}},
			Registration: &fakeAccountRegistration{},
			Clock:        fakeClock{now: accountNow},
			IDs:          &fakeIDs{err: boom},
			StepUpTTL:    5 * time.Minute,
		}
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("empty err = %v", err)
		}
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); err == nil {
			t.Fatal("expected id error")
		}
		repos := &fakeAccountSecurityRepositories{
			stepUpAt: &stepUp,
			registrationCeremony: application.AccountRegistrationCeremony{
				ID: "reg-1", SessionID: 8, IdentityID: "identity-1", Challenge: []byte("challenge"),
				UserHandle: []byte("user"), ExpiresAt: accountNow.Add(time.Minute),
			},
		}
		service.UnitOfWork = fakeAccountSecurityUnitOfWork{repos: repos}
		service.IDs = &fakeIDs{values: []string{"cred-2", "cred-3", "cred-4", "cred-5", "cred-6", "cred-7"}}
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("session mismatch err = %v", err)
		}
		repos.registrationCeremony.SessionID = 9
		repos.registrationCeremony.ExpiresAt = accountNow.Add(-time.Minute)
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("expired err = %v", err)
		}
		repos.registrationCeremony.ExpiresAt = accountNow.Add(time.Minute)
		repos.lockRegistrationErr = boom
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("lock err = %v", err)
		}
		repos.lockRegistrationErr = nil
		service.Registration = &fakeAccountRegistration{verifyErr: boom}
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, domain.ErrCeremonyFailed) {
			t.Fatalf("verify err = %v", err)
		}
		service.Registration = &fakeAccountRegistration{credential: application.PasskeyCredential{CredentialID: []byte("c2"), PublicKey: []byte("pk2")}}
		repos.storeCredentialErr = boom
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("store err = %v", err)
		}
		repos.storeCredentialErr = nil
		repos.consumeRegistrationErr = boom
		if err := service.FinishAddPasskey(context.Background(), application.FinishAddPasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, PasskeyCeremonyID: "reg-1", PasskeyResponse: []byte(`{}`),
		}); !errors.Is(err, boom) {
			t.Fatalf("consume err = %v", err)
		}
	})

	t.Run("remove and change email failures", func(t *testing.T) {
		service := application.AccountSecurityService{
			UnitOfWork: fakeAccountSecurityUnitOfWork{repos: &fakeAccountSecurityRepositories{stepUpAt: &stepUp}},
			Clock:      fakeClock{now: accountNow},
			StepUpTTL:  5 * time.Minute,
		}
		if err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{}); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("empty remove err = %v", err)
		}
		repos := &fakeAccountSecurityRepositories{stepUpAt: &stepUp, listPasskeysErr: boom}
		service.UnitOfWork = fakeAccountSecurityUnitOfWork{repos: repos}
		if err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, CredentialID: "cred-1",
		}); !errors.Is(err, boom) {
			t.Fatalf("list err = %v", err)
		}
		repos.listPasskeysErr = nil
		repos.passkeys = []application.PasskeyCredential{{ID: "cred-1"}, {ID: "cred-2"}}
		repos.deletePasskeyErr = boom
		if err := service.RemovePasskey(context.Background(), application.RemovePasskeyCommand{
			IdentityID: "identity-1", SessionID: 9, CredentialID: "cred-1",
		}); !errors.Is(err, boom) {
			t.Fatalf("delete err = %v", err)
		}
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{}); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("empty change err = %v", err)
		}
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "not-an-email",
		}); err == nil {
			t.Fatal("expected invalid email")
		}
		repos = &fakeAccountSecurityRepositories{stepUpAt: &stepUp, loadStepUpErr: boom, email: "old@example.org"}
		service.UnitOfWork = fakeAccountSecurityUnitOfWork{repos: repos}
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
		}); !errors.Is(err, boom) {
			t.Fatalf("step-up load err = %v", err)
		}
		repos.loadStepUpErr = nil
		repos.activeEmailErr = boom
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
		}); !errors.Is(err, boom) {
			t.Fatalf("active email err = %v", err)
		}
		repos.activeEmailErr = nil
		repos.emailTakenErr = boom
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
		}); !errors.Is(err, boom) {
			t.Fatalf("taken check err = %v", err)
		}
		repos.emailTakenErr = nil
		repos.replaceEmailErr = boom
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
		}); !errors.Is(err, boom) {
			t.Fatalf("replace err = %v", err)
		}
		repos.replaceEmailErr = nil
		repos.revokeErr = boom
		if err := service.ChangeEmail(context.Background(), application.ChangeEmailCommand{
			IdentityID: "identity-1", SessionID: 9, NewEmail: "new@example.org",
		}); !errors.Is(err, boom) {
			t.Fatalf("revoke err = %v", err)
		}
	})
}
