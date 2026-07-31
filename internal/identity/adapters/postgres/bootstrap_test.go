package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestBootstrapUnitOfWorkPersistsFirstAdmin(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 0, 0, 0, time.UTC))
	sessionStore := session.NewStore(db, clk, time.Hour)
	oldSession, oldToken, err := sessionStore.Create(context.Background())
	if err != nil {
		t.Fatalf("create old session: %v", err)
	}

	service := &application.BootstrapService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, clk),
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
	if err := db.QueryRow(context.Background(), `
		SELECT b.closed_by_identity_id, e.email_normalized, r.role
		FROM bootstrap_state b
		JOIN identity_emails e ON e.identity_id = b.closed_by_identity_id
		JOIN identity_roles r ON r.identity_id = b.closed_by_identity_id
	`).Scan(&closedBy, &emailNormalized, &role); err != nil {
		t.Fatalf("read bootstrap state: %v", err)
	}
	if closedBy != "identity-1" || emailNormalized != "first.admin@example.org" || role != string(domain.RoleAdmin) {
		t.Fatalf("persisted identity = %q/%q/%q", closedBy, emailNormalized, role)
	}
}

func TestBootstrapUnitOfWorkMapsDuplicateActiveEmail(t *testing.T) {
	db := testdb.OpenMigrated(t)
	uow := identitypostgres.NewUnitOfWork(db, clock.System())
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

func (fakePasskeys) FinishRegistration(context.Context, application.RegistrationFinish) (application.PasskeyCredential, error) {
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
