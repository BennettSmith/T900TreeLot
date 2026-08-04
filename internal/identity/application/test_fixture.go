package application

import (
	"context"

	"github.com/troop900/treelot/internal/identity/domain"
)

// TestFixtureUnitOfWork is consumed only by production-disabled acceptance setup.
type TestFixtureUnitOfWork interface {
	WithinTestFixtureTx(context.Context, func(context.Context, TestFixtureRepositories) error) error
}

// TestFixtureRepositories provide precondition setup without exposing private tables to acceptance tests.
type TestFixtureRepositories interface {
	FindIdentityByEmail(context.Context, string) (string, error)
	ReplaceRoles(context.Context, string, []domain.Role) error
	RevokeSessionsForIdentity(context.Context, string) error
}

// TestFixtureService establishes accepted-test preconditions through application ports.
type TestFixtureService struct {
	UnitOfWork TestFixtureUnitOfWork
}

type SetFixtureRoleCommand struct {
	Email string
	Role  string
}

func (s TestFixtureService) SetRole(ctx context.Context, command SetFixtureRoleCommand) error {
	email, err := domain.NewEmail(command.Email)
	if err != nil {
		return err
	}
	role, err := domain.ParseRole(command.Role)
	if err != nil {
		return err
	}
	return s.UnitOfWork.WithinTestFixtureTx(ctx, func(txCtx context.Context, repos TestFixtureRepositories) error {
		identityID, err := repos.FindIdentityByEmail(txCtx, email.Normalized())
		if err != nil {
			return err
		}
		return repos.ReplaceRoles(txCtx, identityID, []domain.Role{role})
	})
}

func (s TestFixtureService) RevokeSessions(ctx context.Context, emailRaw string) error {
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		return err
	}
	return s.UnitOfWork.WithinTestFixtureTx(ctx, func(txCtx context.Context, repos TestFixtureRepositories) error {
		identityID, err := repos.FindIdentityByEmail(txCtx, email.Normalized())
		if err != nil {
			return err
		}
		return repos.RevokeSessionsForIdentity(txCtx, identityID)
	})
}
