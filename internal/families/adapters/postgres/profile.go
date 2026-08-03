// Package postgres implements Families persistence adapters.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	families "github.com/troop900/treelot/internal/families/application"
	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
)

type ProfileCreator struct {
	exec executor
}

func NewProfileCreator(db *platformpostgres.DB) *ProfileCreator {
	return &ProfileCreator{exec: db}
}

func NewTxProfileCreator(tx pgx.Tx) *ProfileCreator {
	return &ProfileCreator{exec: tx}
}

func (c *ProfileCreator) CreatePersonalProfile(ctx context.Context, profile families.PersonalProfile) error {
	_, err := c.exec.Exec(ctx, `
		INSERT INTO people (id, first_name, last_name, preferred_display_name, created_at, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
	`, profile.ID, profile.FirstName, profile.LastName, profile.PreferredDisplayName, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create personal profile: %w", err)
	}
	return nil
}

type executor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}
