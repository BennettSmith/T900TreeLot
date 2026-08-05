package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/descope/virtualwebauthn"
	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/platform/testdb"
	platformwebauthn "github.com/troop900/treelot/internal/platform/webauthn"
)

func TestBootstrapUnitOfWorkPersistsFirstAdmin(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 0, 0, 0, time.UTC))
	sessionStore := session.NewStore(db, clk, time.Hour, session.TestKey)
	oldSession, oldToken, err := sessionStore.Create(context.Background())
	if err != nil {
		t.Fatalf("create old session: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
		INSERT INTO webauthn_ceremonies (
			id, session_id, purpose, challenge, user_handle, expires_at, created_at,
			bootstrap_email, bootstrap_first_name, bootstrap_last_name
		)
		VALUES ('ceremony-1', $1, 'bootstrap_registration', $2, $3, $4, $5, $6, $7, $8)
	`,
		oldSession.ID,
		[]byte("challenge"),
		[]byte("user-handle"),
		clk.Now().Add(15*time.Minute),
		clk.Now(),
		"First.Admin@Example.org",
		"First",
		"Admin",
	); err != nil {
		t.Fatalf("seed registration ceremony: %v", err)
	}

	service := &application.BootstrapService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, clk, session.TestKey),
		Tokens:              validToken{},
		RateLimiter:         allowAll{},
		Passkeys:            fakePasskeys{},
		Clock:               clk,
		IDs:                 &sequenceIDs{values: []string{"person-1", "identity-1", "passkey-1"}},
		AuthRateLimitMax:    10,
		AuthRateLimitWindow: 15 * time.Minute,
	}
	result, err := service.FinishBootstrap(context.Background(), application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip:127.0.0.1",
		SessionID:         oldSession.ID,
		Email:             "First.Admin@Example.org",
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-1",
		PasskeyResponse:   []byte(`{"ok":true}`),
	})
	if err != nil {
		t.Fatalf("FinishBootstrap: %v", err)
	}
	if result.IdentityID != "identity-1" || result.Session.RawToken == "" {
		t.Fatalf("result = %#v", result)
	}
	if _, err := sessionStore.Get(context.Background(), oldToken); err == nil {
		t.Fatal("old session token still loads")
	}

	var closedBy, emailNormalized, role string
	var consumedAt *time.Time
	if err := db.QueryRow(context.Background(), `
		SELECT b.closed_by_identity_id, e.email_normalized, r.role, c.consumed_at
		FROM bootstrap_state b
		JOIN identity_emails e ON e.identity_id = b.closed_by_identity_id
		JOIN identity_roles r ON r.identity_id = b.closed_by_identity_id
		JOIN webauthn_ceremonies c ON c.id = 'ceremony-1'
	`).Scan(&closedBy, &emailNormalized, &role, &consumedAt); err != nil {
		t.Fatalf("read bootstrap state: %v", err)
	}
	if closedBy != "identity-1" || emailNormalized != "first.admin@example.org" || role != string(domain.RoleAdmin) {
		t.Fatalf("persisted identity = %q/%q/%q", closedBy, emailNormalized, role)
	}
	if consumedAt == nil {
		t.Fatal("registration ceremony was not consumed")
	}
}

func TestBootstrapUnitOfWorkRollsBackVerifiedRegistrationWhenSessionRotationFails(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 5, 0, 0, time.UTC))
	passkeys, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewRegistrationCeremony: %v", err)
	}
	sessionStore := session.NewStore(db, clk, time.Hour, session.TestKey)
	rollbackSession, _, err := sessionStore.Create(ctx)
	if err != nil {
		t.Fatalf("create rollback session: %v", err)
	}
	if err := sessionStore.Revoke(ctx, rollbackSession.ID); err != nil {
		t.Fatalf("revoke rollback session: %v", err)
	}
	email, err := domain.NewEmail("first.admin@example.org")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	options, err := passkeys.BeginRegistration(ctx, application.RegistrationStart{
		SessionID:   rollbackSession.ID,
		CeremonyID:  "ceremony-rollback",
		Email:       email,
		FirstName:   "First",
		LastName:    "Admin",
		DisplayName: "First Admin",
		UserHandle:  []byte("user-handle-rollback"),
		ExpiresAt:   clk.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	publicKey, err := json.Marshal(options.PublicKey)
	if err != nil {
		t.Fatalf("marshal registration options: %v", err)
	}
	attestationOptions, err := virtualwebauthn.ParseAttestationOptions(string(publicKey))
	if err != nil {
		t.Fatalf("parse registration options: %v", err)
	}
	response := virtualwebauthn.CreateAttestationResponse(
		virtualwebauthn.RelyingParty{Name: "Troop 900 Tree Lot", ID: "localhost", Origin: "http://localhost:8080"},
		virtualwebauthn.NewAuthenticator(),
		virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2),
		*attestationOptions,
	)

	service := &application.BootstrapService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, clk, session.TestKey),
		Tokens:              validToken{},
		RateLimiter:         allowAll{},
		Passkeys:            passkeys,
		Clock:               clk,
		IDs:                 &sequenceIDs{values: []string{"person-rollback", "identity-rollback", "passkey-rollback"}},
		AuthRateLimitMax:    10,
		AuthRateLimitWindow: 15 * time.Minute,
	}
	_, err = service.FinishBootstrap(ctx, application.FinishBootstrapCommand{
		Token:             "valid-token",
		RateLimitKey:      "ip:127.0.0.1",
		SessionID:         rollbackSession.ID,
		Email:             email.String(),
		FirstName:         "First",
		LastName:          "Admin",
		PasskeyCeremonyID: "ceremony-rollback",
		PasskeyResponse:   []byte(response),
	})
	if err == nil {
		t.Fatal("FinishBootstrap succeeded without an existing session")
	}

	var consumedAt, closedAt *time.Time
	var people, identities, emails, roles, credentials, audits, sessions int
	if err := db.QueryRow(ctx, `
		SELECT c.consumed_at,
		       b.closed_at,
		       (SELECT count(*) FROM people),
		       (SELECT count(*) FROM identities),
		       (SELECT count(*) FROM identity_emails),
		       (SELECT count(*) FROM identity_roles),
		       (SELECT count(*) FROM passkey_credentials),
		       (SELECT count(*) FROM audit_events),
		       (SELECT count(*) FROM sessions)
		FROM webauthn_ceremonies c
		CROSS JOIN bootstrap_state b
		WHERE c.id = 'ceremony-rollback' AND b.id = 1
	`).Scan(&consumedAt, &closedAt, &people, &identities, &emails, &roles, &credentials, &audits, &sessions); err != nil {
		t.Fatalf("read rolled-back bootstrap state: %v", err)
	}
	if consumedAt != nil || closedAt != nil || people != 0 || identities != 0 || emails != 0 ||
		roles != 0 || credentials != 0 || audits != 0 || sessions != 1 {
		t.Fatalf(
			"bootstrap changes persisted: consumed=%v closed=%v people=%d identities=%d emails=%d roles=%d credentials=%d audits=%d sessions=%d",
			consumedAt, closedAt, people, identities, emails, roles, credentials, audits, sessions,
		)
	}
}

func TestBootstrapUnitOfWorkRejectsFinishProfileDifferentFromPersistedCeremony(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 7, 0, 0, time.UTC))
	passkeys, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewRegistrationCeremony: %v", err)
	}
	email, err := domain.NewEmail("first.admin@example.org")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	if _, err := passkeys.BeginRegistration(ctx, application.RegistrationStart{
		CeremonyID:           "ceremony-bound-profile",
		Email:                email,
		FirstName:            "First",
		LastName:             "Admin",
		PreferredDisplayName: "First Admin",
		DisplayName:          "First Admin",
		UserHandle:           []byte("user-handle-bound-profile"),
		ExpiresAt:            clk.Now().Add(15 * time.Minute),
	}); err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}

	service := &application.BootstrapService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, clk, session.TestKey),
		Tokens:              validToken{},
		RateLimiter:         allowAll{},
		Passkeys:            passkeys,
		Clock:               clk,
		IDs:                 &sequenceIDs{values: []string{"unused"}},
		AuthRateLimitMax:    10,
		AuthRateLimitWindow: 15 * time.Minute,
	}
	_, err = service.FinishBootstrap(ctx, application.FinishBootstrapCommand{
		Token:                "valid-token",
		RateLimitKey:         "ip:127.0.0.1",
		Email:                "changed@example.org",
		FirstName:            "Changed",
		LastName:             "Person",
		PreferredDisplayName: "Changed Person",
		PasskeyCeremonyID:    "ceremony-bound-profile",
		PasskeyResponse:      []byte(`{"unused":true}`),
	})
	if !errors.Is(err, domain.ErrCeremonyFailed) {
		t.Fatalf("FinishBootstrap error = %v, want generic ErrCeremonyFailed", err)
	}

	var consumedAt, closedAt *time.Time
	var people, identities, emails, roles, credentials, sessions int
	if err := db.QueryRow(ctx, `
		SELECT c.consumed_at,
		       b.closed_at,
		       (SELECT count(*) FROM people),
		       (SELECT count(*) FROM identities),
		       (SELECT count(*) FROM identity_emails),
		       (SELECT count(*) FROM identity_roles),
		       (SELECT count(*) FROM passkey_credentials),
		       (SELECT count(*) FROM sessions)
		FROM webauthn_ceremonies c
		CROSS JOIN bootstrap_state b
		WHERE c.id = 'ceremony-bound-profile' AND b.id = 1
	`).Scan(&consumedAt, &closedAt, &people, &identities, &emails, &roles, &credentials, &sessions); err != nil {
		t.Fatalf("read bootstrap state: %v", err)
	}
	if consumedAt != nil || closedAt != nil || people != 0 || identities != 0 || emails != 0 ||
		roles != 0 || credentials != 0 || sessions != 0 {
		t.Fatalf(
			"changed finish persisted state: consumed=%v closed=%v people=%d identities=%d emails=%d roles=%d credentials=%d sessions=%d",
			consumedAt, closedAt, people, identities, emails, roles, credentials, sessions,
		)
	}
}

func TestBootstrapUnitOfWorkRejectsConsumedRegistrationCeremonyReplay(t *testing.T) {
	db := testdb.OpenMigrated(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 31, 23, 8, 0, 0, time.UTC)
	if _, err := db.Exec(ctx, `
		INSERT INTO webauthn_ceremonies (
			id, purpose, challenge, user_handle, expires_at, created_at,
			bootstrap_email, bootstrap_first_name, bootstrap_last_name
		)
		VALUES ('ceremony-replay', 'bootstrap_registration', $1, $2, $3, $4, $5, $6, $7)
	`,
		[]byte("challenge"),
		[]byte("user-handle"),
		now.Add(15*time.Minute),
		now,
		"first.admin@example.org",
		"First",
		"Admin",
	); err != nil {
		t.Fatalf("seed registration ceremony: %v", err)
	}
	uow := identitypostgres.NewUnitOfWork(db, clock.NewControllable(now), session.TestKey)
	if err := uow.WithinTx(ctx, func(txCtx context.Context, repos application.Repositories) error {
		if _, err := repos.LockRegistrationCeremony(txCtx, "ceremony-replay"); err != nil {
			return err
		}
		return repos.ConsumeRegistrationCeremony(txCtx, "ceremony-replay", now)
	}); err != nil {
		t.Fatalf("consume registration ceremony: %v", err)
	}

	err := uow.WithinTx(ctx, func(txCtx context.Context, repos application.Repositories) error {
		_, err := repos.LockRegistrationCeremony(txCtx, "ceremony-replay")
		return err
	})
	if err == nil {
		t.Fatal("consumed registration ceremony replay succeeded")
	}
}

func TestBootstrapUnitOfWorkMapsDuplicateActiveEmail(t *testing.T) {
	db := testdb.OpenMigrated(t)
	uow := identitypostgres.NewUnitOfWork(db, clock.System(), session.TestKey)
	ctx := context.Background()
	now := time.Date(2026, 7, 31, 23, 10, 0, 0, time.UTC)

	if _, err := db.Exec(ctx, `
		INSERT INTO people (id, first_name, last_name, created_at, updated_at)
		VALUES ('person-1', 'First', 'Admin', $1, $1),
		       ('person-2', 'Second', 'Admin', $1, $1)
	`, now); err != nil {
		t.Fatalf("seed people: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO identities (id, person_id, created_at, updated_at)
		VALUES ('identity-1', 'person-1', $1, $1),
		       ('identity-2', 'person-2', $1, $1)
	`, now); err != nil {
		t.Fatalf("seed identities: %v", err)
	}

	err := uow.WithinTx(ctx, func(ctx context.Context, repos application.Repositories) error {
		if err := repos.AddEmail(ctx, application.IdentityEmail{
			IdentityID: "identity-1",
			Email:      "First@example.org",
			Normalized: "first@example.org",
			Active:     true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return err
		}
		return repos.AddEmail(ctx, application.IdentityEmail{
			IdentityID: "identity-2",
			Email:      "FIRST@example.org",
			Normalized: "first@example.org",
			Active:     true,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	})
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("duplicate email error = %v, want ErrEmailTaken", err)
	}
}

type validToken struct{}

func (validToken) ValidateBootstrapToken(_ context.Context, token string, _ time.Time) error {
	if token != "valid-token" {
		return domain.ErrInvalidToken
	}
	return nil
}

type allowAll struct{}

func (allowAll) Allow(context.Context, string, int, time.Duration) (bool, error) {
	return true, nil
}

type fakePasskeys struct{}

func (fakePasskeys) BeginRegistration(context.Context, application.RegistrationStart) (application.RegistrationOptions, error) {
	return application.RegistrationOptions{}, nil
}

func (fakePasskeys) VerifyRegistration(context.Context, application.RegistrationVerification) (application.PasskeyCredential, error) {
	return application.PasskeyCredential{
		CredentialID:    []byte("credential-id"),
		PublicKey:       []byte("public-key"),
		AttestationType: "none",
		AAGUID:          "00000000-0000-0000-0000-000000000000",
		SignCount:       1,
		Transports:      []string{"internal"},
	}, nil
}

type sequenceIDs struct {
	values []string
}

func (s *sequenceIDs) NewID() (string, error) {
	next := s.values[0]
	s.values = s.values[1:]
	return next, nil
}
