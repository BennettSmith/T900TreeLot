package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

func (r *txRepositories) LoadSessionStepUp(ctx context.Context, sessionID int64, identityID string) (*time.Time, error) {
	var stepUpAt *time.Time
	err := r.tx.QueryRow(ctx, `
		SELECT step_up_at
		FROM sessions
		WHERE id = $1 AND identity_id = $2 AND revoked_at IS NULL
	`, sessionID, identityID).Scan(&stepUpAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load session step-up: %w", err)
	}
	return stepUpAt, nil
}

func (r *txRepositories) MarkSessionStepUp(ctx context.Context, sessionID int64, identityID string, at time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE sessions
		SET step_up_at = $3, last_seen_at = $3
		WHERE id = $1 AND identity_id = $2 AND revoked_at IS NULL
	`, sessionID, identityID, at)
	if err != nil {
		return fmt.Errorf("mark session step-up: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return application.ErrAccountNotFound
	}
	return nil
}

func (r *txRepositories) LoadIdentity(ctx context.Context, identityID string) (application.SignInIdentity, error) {
	return r.loadSignInIdentity(ctx, identityID)
}

func (r *txRepositories) ListPasskeys(ctx context.Context, identityID string) ([]application.PasskeyCredential, error) {
	identity, err := r.loadSignInIdentity(ctx, identityID)
	if err != nil {
		return nil, err
	}
	return identity.Credentials, nil
}

func (r *txRepositories) StoreStepUpCeremony(ctx context.Context, ceremony application.AssertionCeremony) error {
	userHandle := ceremony.UserHandle
	if userHandle == nil {
		userHandle = []byte{}
	}
	_, err := r.tx.Exec(ctx, `
		INSERT INTO webauthn_ceremonies (
			id, session_id, purpose, challenge, identity_id, user_handle,
			expires_at, consumed_at, created_at
		)
		VALUES ($1, $2, 'account_step_up', $3, $4, $5, $6, NULL, now())
	`, ceremony.ID, ceremony.SessionID, ceremony.Challenge, ceremony.IdentityID, userHandle, ceremony.ExpiresAt)
	if err != nil {
		return fmt.Errorf("store step-up ceremony: %w", err)
	}
	return nil
}

func (r *txRepositories) LockStepUpCeremony(ctx context.Context, ceremonyID string) (application.AssertionCeremony, error) {
	var ceremony application.AssertionCeremony
	err := r.tx.QueryRow(ctx, `
		SELECT id, COALESCE(session_id, 0), challenge, COALESCE(identity_id, ''),
		       user_handle, expires_at
		FROM webauthn_ceremonies
		WHERE id = $1 AND purpose = 'account_step_up' AND consumed_at IS NULL
		FOR UPDATE
	`, ceremonyID).Scan(
		&ceremony.ID,
		&ceremony.SessionID,
		&ceremony.Challenge,
		&ceremony.IdentityID,
		&ceremony.UserHandle,
		&ceremony.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.AssertionCeremony{}, fmt.Errorf("step-up ceremony not found")
	}
	if err != nil {
		return application.AssertionCeremony{}, fmt.Errorf("lock step-up ceremony: %w", err)
	}
	return ceremony, nil
}

func (r *txRepositories) ConsumeStepUpCeremony(ctx context.Context, ceremonyID string, consumedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE webauthn_ceremonies
		SET consumed_at = $2
		WHERE id = $1 AND purpose = 'account_step_up' AND consumed_at IS NULL
	`, ceremonyID, consumedAt)
	if err != nil {
		return fmt.Errorf("consume step-up ceremony: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("step-up ceremony not found")
	}
	return nil
}

func (r *txRepositories) StoreAccountRegistrationCeremony(ctx context.Context, ceremony application.AccountRegistrationCeremony) error {
	userHandle := ceremony.UserHandle
	if userHandle == nil {
		userHandle = []byte{}
	}
	_, err := r.tx.Exec(ctx, `
		INSERT INTO webauthn_ceremonies (
			id, session_id, purpose, challenge, identity_id, user_handle,
			expires_at, consumed_at, created_at
		)
		VALUES ($1, NULLIF($2, 0), 'account_registration', $3, $4, $5, $6, NULL, now())
	`, ceremony.ID, ceremony.SessionID, ceremony.Challenge, ceremony.IdentityID, userHandle, ceremony.ExpiresAt)
	if err != nil {
		return fmt.Errorf("store account registration ceremony: %w", err)
	}
	return nil
}

func (r *txRepositories) LockAccountRegistrationCeremony(ctx context.Context, ceremonyID string) (application.AccountRegistrationCeremony, error) {
	var ceremony application.AccountRegistrationCeremony
	err := r.tx.QueryRow(ctx, `
		SELECT id, COALESCE(session_id, 0), COALESCE(identity_id, ''), challenge, user_handle, expires_at
		FROM webauthn_ceremonies
		WHERE id = $1 AND purpose = 'account_registration' AND consumed_at IS NULL
		FOR UPDATE
	`, ceremonyID).Scan(
		&ceremony.ID,
		&ceremony.SessionID,
		&ceremony.IdentityID,
		&ceremony.Challenge,
		&ceremony.UserHandle,
		&ceremony.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.AccountRegistrationCeremony{}, fmt.Errorf("account registration ceremony not found")
	}
	if err != nil {
		return application.AccountRegistrationCeremony{}, fmt.Errorf("lock account registration ceremony: %w", err)
	}
	return ceremony, nil
}

func (r *txRepositories) ConsumeAccountRegistrationCeremony(ctx context.Context, ceremonyID string, consumedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE webauthn_ceremonies
		SET consumed_at = $2
		WHERE id = $1 AND purpose = 'account_registration' AND consumed_at IS NULL
	`, ceremonyID, consumedAt)
	if err != nil {
		return fmt.Errorf("consume account registration ceremony: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("account registration ceremony not found")
	}
	return nil
}

func (r *txRepositories) DeletePasskeyCredential(ctx context.Context, identityID, credentialID string) error {
	tag, err := r.tx.Exec(ctx, `
		DELETE FROM passkey_credentials
		WHERE id = $1 AND identity_id = $2
	`, credentialID, identityID)
	if err != nil {
		return fmt.Errorf("delete passkey credential: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return application.ErrAccountNotFound
	}
	return nil
}

func (r *txRepositories) ActiveEmail(ctx context.Context, identityID string) (domain.Email, error) {
	var value string
	err := r.tx.QueryRow(ctx, `
		SELECT email
		FROM identity_emails
		WHERE identity_id = $1 AND active
		ORDER BY created_at
		LIMIT 1
	`, identityID).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Email{}, application.ErrAccountNotFound
	}
	if err != nil {
		return domain.Email{}, fmt.Errorf("load active email: %w", err)
	}
	return domain.NewEmail(value)
}

func (r *txRepositories) ReplaceActiveEmail(ctx context.Context, identityID, email, normalized string, at time.Time) error {
	taken, err := r.EmailTaken(ctx, normalized)
	if err != nil {
		return err
	}
	if taken {
		// Another active identity already claims this address.
		var current string
		err := r.tx.QueryRow(ctx, `
			SELECT email_normalized
			FROM identity_emails
			WHERE identity_id = $1 AND active
			ORDER BY created_at
			LIMIT 1
		`, identityID).Scan(&current)
		if err == nil && current == normalized {
			return nil
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("load current email for replace: %w", err)
		}
		return domain.ErrEmailTaken
	}
	tag, err := r.tx.Exec(ctx, `
		UPDATE identity_emails
		SET active = false, updated_at = $2
		WHERE identity_id = $1 AND active
	`, identityID, at)
	if err != nil {
		return fmt.Errorf("deactivate current email: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return application.ErrAccountNotFound
	}
	_, err = r.tx.Exec(ctx, `
		INSERT INTO identity_emails (
			identity_id, email, email_normalized, verified_at, active, created_at, updated_at
		)
		VALUES ($1, $2, $3, NULL, true, $4, $4)
		ON CONFLICT (identity_id, email_normalized) DO UPDATE
		SET email = EXCLUDED.email,
		    verified_at = NULL,
		    active = true,
		    updated_at = EXCLUDED.updated_at
	`, identityID, email, normalized, at)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("replace active email: %w", err)
	}
	return nil
}
