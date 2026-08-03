package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/identity/application"
	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
)

// AccountQueries reads authenticated account summaries for browser pages.
type AccountQueries struct {
	db *platformpostgres.DB
}

func NewAccountQueries(db *platformpostgres.DB) *AccountQueries {
	return &AccountQueries{db: db}
}

func (q *AccountQueries) FindAccountProfile(ctx context.Context, identityID string) (application.AccountProfile, error) {
	var profile application.AccountProfile
	err := q.db.QueryRow(ctx, `
		SELECT i.id,
		       COALESCE(NULLIF(p.preferred_display_name, ''), trim(p.first_name || ' ' || p.last_name)) AS display_name,
		       COALESCE((
		           SELECT e.email
		           FROM identity_emails e
		           WHERE e.identity_id = i.id AND e.active
		           ORDER BY e.created_at
		           LIMIT 1
		       ), '') AS primary_email
		FROM identities i
		JOIN people p ON p.id = i.person_id
		WHERE i.id = $1
	`, identityID).Scan(&profile.IdentityID, &profile.DisplayName, &profile.PrimaryEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.AccountProfile{}, application.ErrAccountNotFound
	}
	if err != nil {
		return application.AccountProfile{}, fmt.Errorf("find account profile: %w", err)
	}
	return profile, nil
}
