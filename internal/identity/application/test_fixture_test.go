package application_test

import (
	"context"
	"testing"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

func TestFixtureServiceReplacesRoleByNormalizedEmail(t *testing.T) {
	repos := &fakeFixtureRepositories{identityID: "identity-1"}
	service := application.TestFixtureService{UnitOfWork: fakeFixtureUnitOfWork{repos: repos}}

	err := service.SetRole(context.Background(), application.SetFixtureRoleCommand{
		Email: " Family.Manager@Example.org ",
		Role:  "family_manager",
	})
	if err != nil {
		t.Fatalf("SetRole: %v", err)
	}
	if repos.normalizedEmail != "family.manager@example.org" {
		t.Fatalf("normalized email = %q", repos.normalizedEmail)
	}
	if repos.role != domain.RoleFamilyManager || repos.replacedIdentityID != "identity-1" {
		t.Fatalf("role replacement = %q for %q", repos.role, repos.replacedIdentityID)
	}
}

func TestFixtureServiceRevokesSessionsByNormalizedEmail(t *testing.T) {
	repos := &fakeFixtureRepositories{identityID: "identity-1"}
	service := application.TestFixtureService{UnitOfWork: fakeFixtureUnitOfWork{repos: repos}}

	if err := service.RevokeSessions(context.Background(), " Scout@Example.org "); err != nil {
		t.Fatalf("RevokeSessions: %v", err)
	}
	if repos.normalizedEmail != "scout@example.org" || repos.revokedIdentityID != "identity-1" {
		t.Fatalf("revocation = %#v", repos)
	}
}

type fakeFixtureUnitOfWork struct {
	repos application.TestFixtureRepositories
}

func (u fakeFixtureUnitOfWork) WithinTestFixtureTx(ctx context.Context, fn func(context.Context, application.TestFixtureRepositories) error) error {
	return fn(ctx, u.repos)
}

type fakeFixtureRepositories struct {
	identityID         string
	normalizedEmail    string
	replacedIdentityID string
	revokedIdentityID  string
	role               domain.Role
}

func (r *fakeFixtureRepositories) FindIdentityByEmail(_ context.Context, normalized string) (string, error) {
	r.normalizedEmail = normalized
	return r.identityID, nil
}

func (r *fakeFixtureRepositories) ReplaceRoles(_ context.Context, identityID string, roles []domain.Role) error {
	r.replacedIdentityID = identityID
	if len(roles) == 1 {
		r.role = roles[0]
	}
	return nil
}

func (r *fakeFixtureRepositories) RevokeSessionsForIdentity(_ context.Context, identityID string) error {
	r.revokedIdentityID = identityID
	return nil
}

func (r *fakeFixtureRepositories) SeedConflictingIdentity(_ context.Context, _, _, _ string) error {
	return nil
}

func TestFixtureServiceSeedsConflictingIdentity(t *testing.T) {
	repos := &fakeFixtureRepositories{identityID: "identity-1"}
	service := application.TestFixtureService{UnitOfWork: fakeFixtureUnitOfWork{repos: repos}}
	if err := service.SeedConflictingIdentity(context.Background(), "id-2", "person-2", " Taken@Example.org "); err != nil {
		t.Fatalf("SeedConflictingIdentity: %v", err)
	}
}
