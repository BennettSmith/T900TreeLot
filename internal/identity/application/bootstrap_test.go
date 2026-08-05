package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

func TestFinishBootstrapCreatesFirstAdminAndRotatesSessionAtomically(t *testing.T) {
	fx := newBootstrapFixture()
	service := fx.service()

	result, err := service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip:127.0.0.1",
		SessionID:         41,
		Email:             " First.Admin@Example.org ",
		FirstName:         " First ",
		LastName:          " Admin ",
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{"ok":true}`),
	})
	if err != nil {
		t.Fatalf("FinishBootstrap: %v", err)
	}
	if result.IdentityID != "id-2" || result.PersonID != "id-1" {
		t.Fatalf("result = %#v", result)
	}
	if result.Session.RawToken != "rotated-token" || result.Session.IdentityID != "id-2" {
		t.Fatalf("session = %#v", result.Session)
	}
	if !fx.store.closed || fx.store.closedBy != "id-2" {
		t.Fatalf("bootstrap not closed by new identity: %#v", fx.store)
	}
	if !fx.store.ceremonyConsumed {
		t.Fatal("registration ceremony was not consumed")
	}
	if got := fx.store.roles["id-2"]; len(got) != 1 || got[0] != domain.RoleAdmin {
		t.Fatalf("roles = %#v", got)
	}
	if fx.store.people["id-1"].FirstName != "First" {
		t.Fatalf("profile not trimmed: %#v", fx.store.people["id-1"])
	}
	if fx.store.emails["first.admin@example.org"] != "id-2" {
		t.Fatalf("email index = %#v", fx.store.emails)
	}
	if len(fx.store.audit) != 1 || fx.store.audit[0].Action != "identity.bootstrap.completed" {
		t.Fatalf("audit = %#v", fx.store.audit)
	}
}

func TestFinishBootstrapRollsBackWhenCredentialPersistenceFails(t *testing.T) {
	fx := newBootstrapFixture()
	fx.store.failStoreCredential = errors.New("disk full")
	service := fx.service()

	_, err := service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip:127.0.0.1",
		SessionID:         41,
		Email:             "first.admin@example.org",
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{"ok":true}`),
	})
	if err == nil {
		t.Fatal("FinishBootstrap succeeded")
	}
	if len(fx.store.people) != 0 || len(fx.store.identities) != 0 || fx.store.closed || fx.store.ceremonyConsumed {
		t.Fatalf("state was not rolled back: %#v", fx.store)
	}
}

func TestStartBootstrapRejectsClosedBootstrapWithoutRevealingDetails(t *testing.T) {
	fx := newBootstrapFixture()
	fx.store.closed = true
	service := fx.service()

	_, err := service.StartBootstrap(context.Background(), application.StartBootstrapCommand{
		Token:        "valid-token",
		RateLimitKey: "ip:127.0.0.1",
	})
	if !errors.Is(err, domain.ErrBootstrapClosed) {
		t.Fatalf("StartBootstrap error = %v, want ErrBootstrapClosed", err)
	}
}

func TestClaimBootstrapProfileAndBeginPasskeyRegistration(t *testing.T) {
	fx := newBootstrapFixture()
	fx.ids.values = []string{"ceremony-1", "user-handle-1"}
	service := fx.service()

	pending, err := service.ClaimBootstrapProfile(context.Background(), application.ClaimBootstrapProfileCommand{
		Token:                "valid-token",
		RateLimitKey:         "ip:127.0.0.1",
		Email:                " First.Admin@Example.org ",
		FirstName:            " First ",
		LastName:             " Admin ",
		PreferredDisplayName: " Tree Lot Admin ",
	})
	if err != nil {
		t.Fatalf("ClaimBootstrapProfile: %v", err)
	}
	if pending.Email.Normalized() != "first.admin@example.org" || pending.Name.DisplayName() != "Tree Lot Admin" {
		t.Fatalf("pending = %#v", pending)
	}

	options, err := service.BeginPasskeyRegistration(context.Background(), application.BeginPasskeyRegistrationCommand{
		Pending:   pending,
		SessionID: 11,
	})
	if err != nil {
		t.Fatalf("BeginPasskeyRegistration: %v", err)
	}
	if options.CeremonyID != "ceremony-1" || string(options.UserHandle) != "user-handle-1" {
		t.Fatalf("options = %#v", options)
	}
	if fx.passkeys.lastStart.DisplayName != "Tree Lot Admin" || fx.passkeys.lastStart.SessionID != 11 {
		t.Fatalf("passkey start = %#v", fx.passkeys.lastStart)
	}
}

func TestFinishBootstrapRejectsProfileChangedAfterRegistrationBegan(t *testing.T) {
	fx := newBootstrapFixture()
	fx.ids.values = []string{"ceremony-1", "user-handle-1", "person-1", "identity-1", "passkey-1"}
	service := fx.service()

	pending, err := service.ClaimBootstrapProfile(context.Background(), application.ClaimBootstrapProfileCommand{
		Token:                "valid-token",
		RateLimitKey:         "ip:127.0.0.1",
		Email:                "first.admin@example.org",
		FirstName:            "First",
		LastName:             "Admin",
		PreferredDisplayName: "First Admin",
	})
	if err != nil {
		t.Fatalf("ClaimBootstrapProfile: %v", err)
	}
	if _, err := service.BeginPasskeyRegistration(context.Background(), application.BeginPasskeyRegistrationCommand{
		Pending:   pending,
		SessionID: 41,
	}); err != nil {
		t.Fatalf("BeginPasskeyRegistration: %v", err)
	}

	_, err = service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:                "valid-token",
		RateLimitKey:         "ip:127.0.0.1",
		SessionID:            41,
		Email:                "attacker@example.org",
		FirstName:            "Changed",
		LastName:             "Person",
		PreferredDisplayName: "Changed Person",
		PasskeyCeremonyID:    "ceremony-1",
		PasskeyResponse:      []byte(`{"ok":true}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("FinishBootstrap error = %v, want generic ErrCeremonyFailed", err)
	}
	if len(fx.store.people) != 0 ||
		len(fx.store.identities) != 0 ||
		len(fx.store.emails) != 0 ||
		len(fx.store.roles) != 0 ||
		len(fx.store.credentials) != 0 ||
		fx.store.sessionRotations != 0 ||
		fx.store.closed ||
		fx.store.ceremonyConsumed {
		t.Fatalf("mismatched finish modified state: %#v", fx.store)
	}

	result, err := service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:                "valid-token",
		RateLimitKey:         "ip:127.0.0.1",
		SessionID:            41,
		Email:                "FIRST.ADMIN@example.org",
		FirstName:            "First",
		LastName:             "Admin",
		PreferredDisplayName: "First Admin",
		PasskeyCeremonyID:    "ceremony-1",
		PasskeyResponse:      []byte(`{"ok":true}`),
	})
	if err != nil {
		t.Fatalf("FinishBootstrap retry with original claim: %v", err)
	}
	if result.IdentityID != "identity-1" || fx.store.people["person-1"].PreferredDisplayName != "First Admin" {
		t.Fatalf("retry did not use ceremony-bound profile: result=%#v store=%#v", result, fx.store)
	}
}

func TestFinishBootstrapRejectsCeremonyFromAnotherSessionWithoutConsumingIt(t *testing.T) {
	fx := newBootstrapFixture()
	fx.ids.values = []string{
		"ceremony-session-bound",
		"user-handle-session-bound",
		"person-session-bound",
		"identity-session-bound",
		"passkey-session-bound",
	}
	service := fx.service()

	pending, err := service.ClaimBootstrapProfile(context.Background(), application.ClaimBootstrapProfileCommand{
		Token:        "valid-token",
		RateLimitKey: "ip:127.0.0.1",
		Email:        "first.admin@example.org",
		FirstName:    "First",
		LastName:     "Admin",
	})
	if err != nil {
		t.Fatalf("ClaimBootstrapProfile: %v", err)
	}
	const sessionA int64 = 41
	if _, err := service.BeginPasskeyRegistration(context.Background(), application.BeginPasskeyRegistrationCommand{
		Pending:   pending,
		SessionID: sessionA,
	}); err != nil {
		t.Fatalf("BeginPasskeyRegistration: %v", err)
	}

	finish := application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip:127.0.0.1",
		Email:             "first.admin@example.org",
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-session-bound",
		PasskeyResponse:   []byte(`{"ok":true}`),
	}
	finish.SessionID = 42
	if _, err := service.FinishBootstrap(context.Background(), finish); !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("session B FinishBootstrap error = %v, want generic ErrCeremonyFailed", err)
	}
	if fx.store.ceremonyConsumed {
		t.Fatal("session B consumed session A's registration ceremony")
	}

	finish.SessionID = sessionA
	result, err := service.FinishBootstrap(context.Background(), finish)
	if err != nil {
		t.Fatalf("session A FinishBootstrap: %v", err)
	}
	if result.IdentityID != "identity-session-bound" || !fx.store.ceremonyConsumed {
		t.Fatalf("session A did not finish its ceremony: result=%#v store=%#v", result, fx.store)
	}
}

func TestBootstrapPreflightRejectsRateLimitAndInvalidToken(t *testing.T) {
	fx := newBootstrapFixture()
	fx.rateLimits.allowed = false
	service := fx.service()
	_, err := service.StartBootstrap(context.Background(), application.StartBootstrapCommand{Token: "valid-token", RateLimitKey: "ip"})
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("rate limited error = %v, want ErrRateLimited", err)
	}

	fx = newBootstrapFixture()
	service = fx.service()
	_, err = service.StartBootstrap(context.Background(), application.StartBootstrapCommand{Token: "wrong-token", RateLimitKey: "ip"})
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("invalid token error = %v, want ErrInvalidToken", err)
	}
}

func TestBeginPasskeyRegistrationRejectsExpiredPendingEnrollment(t *testing.T) {
	fx := newBootstrapFixture()
	service := fx.service()
	email, err := domain.NewEmail("first@example.org")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	name, err := domain.ValidateProfile("First", "Admin", "")
	if err != nil {
		t.Fatalf("ValidateProfile: %v", err)
	}
	_, err = service.BeginPasskeyRegistration(context.Background(), application.BeginPasskeyRegistrationCommand{
		Pending: application.PendingEnrollment{
			Token:      "valid-token",
			Email:      email,
			Name:       name,
			ValidUntil: time.Date(2026, 7, 31, 21, 59, 0, 0, time.UTC),
		},
		SessionID: 11,
	})
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expired pending error = %v, want ErrInvalidToken", err)
	}
}

func TestClaimBootstrapProfileRejectsTakenEmail(t *testing.T) {
	fx := newBootstrapFixture()
	fx.store.emails["first.admin@example.org"] = "identity-1"
	service := fx.service()

	_, err := service.ClaimBootstrapProfile(context.Background(), application.ClaimBootstrapProfileCommand{
		Token:        "valid-token",
		RateLimitKey: "ip",
		Email:        "first.admin@example.org",
		FirstName:    "First",
		LastName:     "Admin",
	})
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("taken email error = %v, want ErrEmailTaken", err)
	}
}

func TestFinishBootstrapMapsPasskeyFailure(t *testing.T) {
	fx := newBootstrapFixture()
	fx.passkeys.finishErr = errors.New("bad attestation")
	service := fx.service()

	_, err := service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip",
		SessionID:         41,
		Email:             "first.admin@example.org",
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("passkey failure error = %v, want ErrCeremonyFailed", err)
	}
}

func TestFinishBootstrapRejectsExpiredRegistrationCeremony(t *testing.T) {
	fx := newBootstrapFixture()
	fx.store.ceremonyExpiresAt = time.Date(2026, 7, 31, 21, 59, 0, 0, time.UTC)

	_, err := fx.service().FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip",
		SessionID:         41,
		Email:             "first.admin@example.org",
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("expired ceremony error = %v, want ErrCeremonyFailed", err)
	}
	if fx.store.ceremonyConsumed {
		t.Fatal("expired registration ceremony was consumed")
	}
}

func TestStartBootstrapUsesDefaultTimingAndRateLimitSettings(t *testing.T) {
	fx := newBootstrapFixture()
	service := fx.service()
	service.Clock = nil
	service.AuthRateLimitMax = 0
	service.AuthRateLimitWindow = 0
	service.BootstrapTokenExpiresIn = 0

	result, err := service.StartBootstrap(context.Background(), application.StartBootstrapCommand{Token: "valid-token", RateLimitKey: "ip"})
	if err != nil {
		t.Fatalf("StartBootstrap: %v", err)
	}
	if fx.rateLimits.lastMax != 10 || fx.rateLimits.lastWindow != 15*time.Minute {
		t.Fatalf("rate limit defaults = %d/%v", fx.rateLimits.lastMax, fx.rateLimits.lastWindow)
	}
	if !result.ExpiresAt.After(time.Now().UTC().Add(71 * time.Hour)) {
		t.Fatalf("ExpiresAt = %v, want about 72h from now", result.ExpiresAt)
	}
}

type bootstrapFixture struct {
	store      *fakeStore
	tokens     *fakeTokenValidator
	rateLimits *fakeRateLimiter
	passkeys   *fakePasskeys
	ids        *fakeIDs
}

func newBootstrapFixture() *bootstrapFixture {
	store := newFakeStore()
	return &bootstrapFixture{
		store:      store,
		tokens:     &fakeTokenValidator{valid: "valid-token"},
		rateLimits: &fakeRateLimiter{allowed: true},
		passkeys:   &fakePasskeys{store: store},
		ids:        &fakeIDs{values: []string{"id-1", "id-2", "id-3"}},
	}
}

func (f *bootstrapFixture) service() *application.BootstrapService {
	return &application.BootstrapService{
		UnitOfWork:              f.store,
		Tokens:                  f.tokens,
		RateLimiter:             f.rateLimits,
		Passkeys:                f.passkeys,
		Clock:                   fakeClock{now: time.Date(2026, 7, 31, 22, 0, 0, 0, time.UTC)},
		IDs:                     f.ids,
		AuthRateLimitMax:        10,
		AuthRateLimitWindow:     15 * time.Minute,
		BootstrapTokenExpiresIn: 72 * time.Hour,
	}
}

type fakeStore struct {
	closed              bool
	adminExists         bool
	ceremonyConsumed    bool
	ceremonySessionID   int64
	ceremonyExpiresAt   time.Time
	ceremonyEmail       domain.Email
	ceremonyName        domain.ProfileName
	closedBy            string
	failStoreCredential error
	people              map[string]application.PersonalProfile
	identities          map[string]application.IdentityRecord
	emails              map[string]string
	roles               map[string][]domain.Role
	credentials         map[string]application.PasskeyCredential
	audit               []application.AuditEvent
	sessionRotations    int
}

func newFakeStore() *fakeStore {
	email, _ := domain.NewEmail("first.admin@example.org")
	name, _ := domain.ValidateProfile("First", "Admin", "")
	return &fakeStore{
		people:            map[string]application.PersonalProfile{},
		identities:        map[string]application.IdentityRecord{},
		emails:            map[string]string{},
		roles:             map[string][]domain.Role{},
		credentials:       map[string]application.PasskeyCredential{},
		ceremonySessionID: 41,
		ceremonyExpiresAt: time.Date(2026, 7, 31, 22, 15, 0, 0, time.UTC),
		ceremonyEmail:     email,
		ceremonyName:      name,
	}
}

func (s *fakeStore) WithinTx(ctx context.Context, fn func(context.Context, application.Repositories) error) error {
	snapshot := *s
	snapshot.people = cloneMap(s.people)
	snapshot.identities = cloneMap(s.identities)
	snapshot.emails = cloneMap(s.emails)
	snapshot.roles = cloneRoleMap(s.roles)
	snapshot.credentials = cloneMap(s.credentials)
	snapshot.audit = append([]application.AuditEvent(nil), s.audit...)
	if err := fn(ctx, s); err != nil {
		*s = snapshot
		return err
	}
	return nil
}

func (s *fakeStore) AdminExists(context.Context) (bool, error) {
	return s.adminExists, nil
}

func (s *fakeStore) LockBootstrap(context.Context) (application.BootstrapState, error) {
	return application.BootstrapState{Closed: s.closed}, nil
}

func (s *fakeStore) EmailTaken(_ context.Context, normalized string) (bool, error) {
	_, ok := s.emails[normalized]
	return ok, nil
}

func (s *fakeStore) LockRegistrationCeremony(context.Context, string) (application.RegistrationCeremony, error) {
	if s.ceremonyConsumed {
		return application.RegistrationCeremony{}, errors.New("webauthn ceremony not found")
	}
	return application.RegistrationCeremony{
		SessionID:  s.ceremonySessionID,
		Challenge:  []byte("challenge"),
		UserHandle: []byte("user-handle"),
		ExpiresAt:  s.ceremonyExpiresAt,
		Email:      s.ceremonyEmail,
		Name:       s.ceremonyName,
	}, nil
}

func (s *fakeStore) ConsumeRegistrationCeremony(context.Context, string, time.Time) error {
	s.ceremonyConsumed = true
	return nil
}

func (s *fakeStore) CreatePersonalProfile(_ context.Context, profile application.PersonalProfile) error {
	s.people[profile.ID] = profile
	return nil
}

func (s *fakeStore) CreateIdentity(_ context.Context, record application.IdentityRecord) error {
	s.identities[record.ID] = record
	return nil
}

func (s *fakeStore) AddEmail(_ context.Context, email application.IdentityEmail) error {
	s.emails[email.Normalized] = email.IdentityID
	return nil
}

func (s *fakeStore) GrantRole(_ context.Context, identityID string, role domain.Role) error {
	s.roles[identityID] = append(s.roles[identityID], role)
	return nil
}

func (s *fakeStore) StorePasskeyCredential(_ context.Context, credential application.PasskeyCredential) error {
	if s.failStoreCredential != nil {
		return s.failStoreCredential
	}
	s.credentials[credential.ID] = credential
	return nil
}

func (s *fakeStore) CloseBootstrap(_ context.Context, identityID string, _ time.Time) error {
	s.closed = true
	s.closedBy = identityID
	return nil
}

func (s *fakeStore) WriteAudit(_ context.Context, event application.AuditEvent) error {
	s.audit = append(s.audit, event)
	return nil
}

func (s *fakeStore) RotateForIdentity(_ context.Context, oldSessionID int64, identityID string, now time.Time) (application.IssuedSession, error) {
	s.sessionRotations++
	return application.IssuedSession{ID: oldSessionID + 1, IdentityID: identityID, RawToken: "rotated-token", AuthenticatedAt: now}, nil
}

type fakeTokenValidator struct {
	valid string
}

func (v *fakeTokenValidator) ValidateBootstrapToken(_ context.Context, token string, _ time.Time) error {
	if token != v.valid {
		return domain.ErrInvalidToken
	}
	return nil
}

type fakeRateLimiter struct {
	allowed    bool
	lastKey    string
	lastMax    int
	lastWindow time.Duration
}

func (r *fakeRateLimiter) Allow(_ context.Context, key string, max int, window time.Duration) (bool, error) {
	r.lastKey = key
	r.lastMax = max
	r.lastWindow = window
	return r.allowed, nil
}

type fakePasskeys struct {
	lastStart application.RegistrationStart
	finishErr error
	store     *fakeStore
}

func (p *fakePasskeys) BeginRegistration(_ context.Context, start application.RegistrationStart) (application.RegistrationOptions, error) {
	p.lastStart = start
	if p.store != nil {
		p.store.ceremonySessionID = start.SessionID
		p.store.ceremonyEmail = start.Email
		p.store.ceremonyName = domain.ProfileName{
			FirstName:            start.FirstName,
			LastName:             start.LastName,
			PreferredDisplayName: start.PreferredDisplayName,
		}
		p.store.ceremonyExpiresAt = start.ExpiresAt
	}
	return application.RegistrationOptions{CeremonyID: start.CeremonyID, UserHandle: start.UserHandle, ExpiresAt: start.ExpiresAt}, nil
}

func (p *fakePasskeys) VerifyRegistration(context.Context, application.RegistrationVerification) (application.PasskeyCredential, error) {
	if p.finishErr != nil {
		return application.PasskeyCredential{}, p.finishErr
	}
	return application.PasskeyCredential{
		CredentialID:    []byte("credential-id"),
		PublicKey:       []byte("public-key"),
		AttestationType: "none",
		AAGUID:          "00000000-0000-0000-0000-000000000000",
		SignCount:       1,
		Transports:      []string{"internal"},
	}, nil
}

type fakeClock struct {
	now time.Time
}

func (c fakeClock) Now() time.Time {
	return c.now
}

type fakeIDs struct {
	values []string
	err    error
}

func (g *fakeIDs) NewID() (string, error) {
	if g.err != nil {
		return "", g.err
	}
	next := g.values[0]
	g.values = g.values[1:]
	return next, nil
}

func cloneMap[K comparable, V any](in map[K]V) map[K]V {
	out := make(map[K]V, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneRoleMap(in map[string][]domain.Role) map[string][]domain.Role {
	out := make(map[string][]domain.Role, len(in))
	for key, value := range in {
		out[key] = append([]domain.Role(nil), value...)
	}
	return out
}
