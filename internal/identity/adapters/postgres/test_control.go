package postgres

import (
	"context"
	"fmt"

	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
)

// TestControl exposes acceptance-only identity reset operations.
type TestControl struct {
	db *platformpostgres.DB
}

func NewTestControl(db *platformpostgres.DB) *TestControl {
	return &TestControl{db: db}
}

func (c *TestControl) ResetBootstrap(ctx context.Context) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin bootstrap reset: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	statements := []string{
		`DELETE FROM passkey_credentials`,
		`DELETE FROM identity_roles`,
		`DELETE FROM identity_emails`,
		`DELETE FROM webauthn_ceremonies`,
		`DELETE FROM rate_limit_buckets`,
		`UPDATE sessions SET revoked_at = COALESCE(revoked_at, now()), identity_id = NULL, authenticated_at = NULL WHERE identity_id IS NOT NULL`,
		`UPDATE bootstrap_state SET closed_at = NULL, closed_by_identity_id = NULL WHERE id = 1`,
		`DELETE FROM identities`,
		`DELETE FROM people`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("reset bootstrap state: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit bootstrap reset: %w", err)
	}
	return nil
}
