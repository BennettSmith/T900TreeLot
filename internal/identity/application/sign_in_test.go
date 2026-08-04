package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

func TestBeginSignInStoresDiscoverableCeremonyAfterRateLimit(t *testing.T) {
	repos := &fakeSignInRepositories{}
	passkeys := &fakeAssertions{}
	limiter := &fakeRateLimiter{allowed: true}
	service := application.SignInService{
		UnitOfWork:          fakeSignInUnitOfWork{repos: repos},
		RateLimiter:         limiter,
		Passkeys:            passkeys,
		Clock:               fakeClock{now: signInNow},
		IDs:                 &fakeIDs{values: []string{"ceremony-1"}},
		AuthRateLimitMax:    5,
		AuthRateLimitWindow: time.Minute,
	}

	result, err := service.BeginSignIn(context.Background(), application.BeginSignInCommand{
		SessionID:    41,
		RateLimitKey: "sign-in:127.0.0.1",
	})
	if err != nil {
		t.Fatalf("BeginSignIn: %v", err)
	}
	if result.CeremonyID != "ceremony-1" || string(result.PublicKey.([]byte)) != "options" {
		t.Fatalf("result = %#v", result)
	}
	if passkeys.beginIdentity != nil {
		t.Fatalf("discoverable begin received identity %#v", passkeys.beginIdentity)
	}
	if repos.ceremony.SessionID != 41 || repos.ceremony.IdentityID != "" || string(repos.ceremony.Challenge) != "challenge" {
		t.Fatalf("ceremony = %#v", repos.ceremony)
	}
	if limiter.lastKey != "sign-in:127.0.0.1" || limiter.lastMax != 5 || limiter.lastWindow != time.Minute {
		t.Fatalf("rate limit = %#v", limiter)
	}
}

func TestBeginSignInUsesDecoyCredentialForUnknownEmailHint(t *testing.T) {
	repos := &fakeSignInRepositories{findByEmailErr: application.ErrSignInIdentityNotFound}
	passkeys := &fakeAssertions{}
	service := application.SignInService{
		UnitOfWork: fakeSignInUnitOfWork{repos: repos},
		RateLimiter: &fakeRateLimiter{
			allowed: true,
		},
		Passkeys: passkeys,
		Clock:    fakeClock{now: signInNow},
		IDs:      &fakeIDs{values: []string{"ceremony-1", "decoy-handle", "decoy-credential"}},
	}

	if _, err := service.BeginSignIn(context.Background(), application.BeginSignInCommand{
		SessionID:    41,
		RateLimitKey: "sign-in:127.0.0.1",
		EmailHint:    "missing@example.org",
	}); err != nil {
		t.Fatalf("BeginSignIn: %v", err)
	}
	if passkeys.beginIdentity == nil || passkeys.beginIdentity.ID != "" || len(passkeys.beginIdentity.Credentials) != 1 {
		t.Fatalf("decoy identity = %#v", passkeys.beginIdentity)
	}
	if repos.ceremony.IdentityID != "" {
		t.Fatalf("decoy ceremony exposed identity = %#v", repos.ceremony)
	}
}

func TestBeginSignInRejectsRateLimitedAttemptBeforeCeremony(t *testing.T) {
	repos := &fakeSignInRepositories{}
	service := application.SignInService{
		UnitOfWork: fakeSignInUnitOfWork{repos: repos},
		RateLimiter: &fakeRateLimiter{
			allowed: false,
		},
		Passkeys: &fakeAssertions{},
		IDs:      &fakeIDs{values: []string{"unused"}},
	}

	_, err := service.BeginSignIn(context.Background(), application.BeginSignInCommand{RateLimitKey: "sign-in:127.0.0.1"})
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("error = %v, want ErrRateLimited", err)
	}
	if repos.storedCeremonyCalled {
		t.Fatal("rate-limited attempt stored a ceremony")
	}
}

func TestFinishSignInVerifiesAssertionUpdatesCredentialAuditsAndRotatesSession(t *testing.T) {
	repos := &fakeSignInRepositories{
		ceremony: application.AssertionCeremony{
			SessionID: 41,
			Challenge: []byte("challenge"),
			ExpiresAt: signInNow.Add(time.Minute),
		},
		identity: application.SignInIdentity{
			ID:         "identity-1",
			PersonID:   "person-1",
			UserHandle: []byte("user-handle"),
			Roles:      []domain.Role{domain.RoleFamilyManager},
			Credentials: []application.PasskeyCredential{{
				ID:           "credential-1",
				CredentialID: []byte("credential-id"),
				PublicKey:    []byte("public-key"),
				SignCount:    3,
			}},
		},
	}
	passkeys := &fakeAssertions{
		credentialID: []byte("credential-id"),
		verified: application.VerifiedAssertion{
			CredentialID: []byte("credential-id"),
			SignCount:    4,
		},
	}
	service := application.SignInService{
		UnitOfWork: fakeSignInUnitOfWork{repos: repos},
		Passkeys:   passkeys,
		Clock:      fakeClock{now: signInNow},
	}

	result, err := service.FinishSignIn(context.Background(), application.FinishSignInCommand{
		SessionID:         41,
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{"credential":"response"}`),
	})
	if err != nil {
		t.Fatalf("FinishSignIn: %v", err)
	}
	if result.IdentityID != "identity-1" || result.RedirectTo != "/family" || result.Session.RawToken != "rotated-token" {
		t.Fatalf("result = %#v", result)
	}
	if !repos.consumed || repos.updatedCredentialID != "credential-1" || repos.updatedSignCount != 4 {
		t.Fatalf("credential state = %#v", repos)
	}
	if len(repos.audit) != 1 || repos.audit[0].Action != "identity.sign_in.completed" || len(repos.audit[0].Payload) != 0 {
		t.Fatalf("audit = %#v", repos.audit)
	}
	if passkeys.verification.Discoverable != true || passkeys.verification.Identity.ID != "identity-1" {
		t.Fatalf("verification = %#v", passkeys.verification)
	}
}

func TestFinishSignInRejectsCeremonyFromAnotherBrowserSession(t *testing.T) {
	repos := &fakeSignInRepositories{ceremony: application.AssertionCeremony{
		SessionID: 99,
		ExpiresAt: signInNow.Add(time.Minute),
	}}
	service := application.SignInService{
		UnitOfWork: fakeSignInUnitOfWork{repos: repos},
		Passkeys:   &fakeAssertions{},
		Clock:      fakeClock{now: signInNow},
	}

	_, err := service.FinishSignIn(context.Background(), application.FinishSignInCommand{
		SessionID:         41,
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("error = %v, want ErrCeremonyFailed", err)
	}
	if repos.consumed || repos.rotations != 0 {
		t.Fatalf("failed ceremony changed state: %#v", repos)
	}
}

var signInNow = time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC)

type fakeSignInUnitOfWork struct {
	repos application.SignInRepositories
}

func (u fakeSignInUnitOfWork) WithinSignInTx(ctx context.Context, fn func(context.Context, application.SignInRepositories) error) error {
	return fn(ctx, u.repos)
}

type fakeSignInRepositories struct {
	ceremony             application.AssertionCeremony
	identity             application.SignInIdentity
	consumed             bool
	updatedCredentialID  string
	updatedSignCount     uint32
	audit                []application.AuditEvent
	rotations            int
	findByEmailErr       error
	findByCredentialErr  error
	storedCeremonyCalled bool
}

func (r *fakeSignInRepositories) FindSignInIdentityByEmail(context.Context, string) (application.SignInIdentity, error) {
	return r.identity, r.findByEmailErr
}

func (r *fakeSignInRepositories) FindSignInIdentityByCredential(context.Context, []byte) (application.SignInIdentity, error) {
	return r.identity, r.findByCredentialErr
}

func (r *fakeSignInRepositories) StoreAssertionCeremony(_ context.Context, ceremony application.AssertionCeremony) error {
	r.ceremony = ceremony
	r.storedCeremonyCalled = true
	return nil
}

func (r *fakeSignInRepositories) LockAssertionCeremony(context.Context, string) (application.AssertionCeremony, error) {
	return r.ceremony, nil
}

func (r *fakeSignInRepositories) ConsumeAssertionCeremony(context.Context, string, time.Time) error {
	r.consumed = true
	return nil
}

func (r *fakeSignInRepositories) UpdatePasskeyAfterAssertion(_ context.Context, credentialID string, signCount uint32, _ time.Time) error {
	r.updatedCredentialID = credentialID
	r.updatedSignCount = signCount
	return nil
}

func (r *fakeSignInRepositories) WriteAudit(_ context.Context, event application.AuditEvent) error {
	r.audit = append(r.audit, event)
	return nil
}

func (r *fakeSignInRepositories) RotateForIdentity(_ context.Context, _ int64, identityID string, now time.Time) (application.IssuedSession, error) {
	r.rotations++
	return application.IssuedSession{
		ID:              42,
		IdentityID:      identityID,
		RawToken:        "rotated-token",
		AuthenticatedAt: now,
		ExpiresAt:       now.Add(time.Hour),
	}, nil
}

type fakeAssertions struct {
	beginIdentity *application.SignInIdentity
	credentialID  []byte
	verified      application.VerifiedAssertion
	verification  application.AssertionVerification
}

func (a *fakeAssertions) BeginAssertion(_ context.Context, identity *application.SignInIdentity) (application.AssertionOptions, error) {
	a.beginIdentity = identity
	return application.AssertionOptions{PublicKey: []byte("options"), Challenge: []byte("challenge")}, nil
}

func (a *fakeAssertions) CredentialID([]byte) ([]byte, error) {
	return a.credentialID, nil
}

func (a *fakeAssertions) VerifyAssertion(_ context.Context, verification application.AssertionVerification) (application.VerifiedAssertion, error) {
	a.verification = verification
	return a.verified, nil
}
