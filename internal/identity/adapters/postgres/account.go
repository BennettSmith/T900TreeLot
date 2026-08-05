package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
)

// AccountQueries reads authenticated account summaries for browser pages.
type AccountQueries struct {
	db *platformpostgres.DB
}

func (q *AccountQueries) FindLandingProfile(ctx context.Context, identityID string) (application.LandingProfile, error) {
	var profile application.LandingProfile
	err := q.db.QueryRow(ctx, `
		SELECT i.id,
		       COALESCE(NULLIF(p.preferred_display_name, ''), trim(p.first_name || ' ' || p.last_name))
		FROM identities i
		JOIN people p ON p.id = i.person_id
		WHERE i.id = $1
	`, identityID).Scan(&profile.IdentityID, &profile.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.LandingProfile{}, application.ErrAccountNotFound
	}
	if err != nil {
		return application.LandingProfile{}, fmt.Errorf("find landing profile: %w", err)
	}
	rows, err := q.db.Query(ctx, `SELECT role FROM identity_roles WHERE identity_id = $1 ORDER BY role`, identityID)
	if err != nil {
		return application.LandingProfile{}, fmt.Errorf("find landing roles: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return application.LandingProfile{}, fmt.Errorf("scan landing role: %w", err)
		}
		role, err := domain.ParseRole(value)
		if err != nil {
			return application.LandingProfile{}, err
		}
		profile.Roles = append(profile.Roles, role)
	}
	if err := rows.Err(); err != nil {
		return application.LandingProfile{}, fmt.Errorf("iterate landing roles: %w", err)
	}
	return profile, nil
}

func NewAccountQueries(db *platformpostgres.DB) *AccountQueries {
	return &AccountQueries{db: db}
}

func (q *AccountQueries) ListPasskeys(ctx context.Context, identityID string) ([]application.PasskeyCredential, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, credential_id, public_key, attestation_type, aaguid,
		       sign_count, transports, authenticator_flags, created_at, last_used_at
		FROM passkey_credentials
		WHERE identity_id = $1
		ORDER BY created_at, id
	`, identityID)
	if err != nil {
		return nil, fmt.Errorf("list passkeys: %w", err)
	}
	defer rows.Close()
	var credentials []application.PasskeyCredential
	for rows.Next() {
		var credential application.PasskeyCredential
		var signCount int64
		var authenticatorFlags *int16
		if err := rows.Scan(
			&credential.ID,
			&credential.CredentialID,
			&credential.PublicKey,
			&credential.AttestationType,
			&credential.AAGUID,
			&signCount,
			&credential.Transports,
			&authenticatorFlags,
			&credential.CreatedAt,
			&credential.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("scan passkey: %w", err)
		}
		credential.IdentityID = identityID
		credential.SignCount = uint32(signCount)
		if authenticatorFlags != nil {
			credential.AuthenticatorFlags = uint8(*authenticatorFlags)
			credential.FlagsKnown = true
		}
		credentials = append(credentials, credential)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate passkeys: %w", err)
	}
	return credentials, nil
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
