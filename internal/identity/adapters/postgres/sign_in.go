package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
)

func (r *txRepositories) FindSignInIdentityByEmail(ctx context.Context, normalized string) (application.SignInIdentity, error) {
	var identityID string
	err := r.tx.QueryRow(ctx, `
		SELECT i.id
		FROM identities i
		JOIN identity_emails e ON e.identity_id = i.id
		WHERE e.email_normalized = $1 AND e.active
	`, normalized).Scan(&identityID)
	if err == pgx.ErrNoRows {
		return application.SignInIdentity{}, application.ErrSignInIdentityNotFound
	}
	if err != nil {
		return application.SignInIdentity{}, fmt.Errorf("find sign-in identity by email: %w", err)
	}
	return r.loadSignInIdentity(ctx, identityID)
}

func (r *txRepositories) FindSignInIdentityByCredential(ctx context.Context, credentialID []byte) (application.SignInIdentity, error) {
	var identityID string
	err := r.tx.QueryRow(ctx, `
		SELECT identity_id
		FROM passkey_credentials
		WHERE credential_id = $1
	`, credentialID).Scan(&identityID)
	if err == pgx.ErrNoRows {
		return application.SignInIdentity{}, application.ErrSignInIdentityNotFound
	}
	if err != nil {
		return application.SignInIdentity{}, fmt.Errorf("find sign-in identity by credential: %w", err)
	}
	return r.loadSignInIdentity(ctx, identityID)
}

func (r *txRepositories) loadSignInIdentity(ctx context.Context, identityID string) (application.SignInIdentity, error) {
	var identity application.SignInIdentity
	err := r.tx.QueryRow(ctx, `
		SELECT id, person_id, COALESCE(webauthn_user_handle, '\x'::bytea)
		FROM identities
		WHERE id = $1
	`, identityID).Scan(&identity.ID, &identity.PersonID, &identity.UserHandle)
	if err != nil {
		return application.SignInIdentity{}, fmt.Errorf("load sign-in identity: %w", err)
	}
	roleRows, err := r.tx.Query(ctx, `
		SELECT role
		FROM identity_roles
		WHERE identity_id = $1
		ORDER BY role
	`, identityID)
	if err != nil {
		return application.SignInIdentity{}, fmt.Errorf("load sign-in roles: %w", err)
	}
	for roleRows.Next() {
		var value string
		if err := roleRows.Scan(&value); err != nil {
			roleRows.Close()
			return application.SignInIdentity{}, fmt.Errorf("scan sign-in role: %w", err)
		}
		role, err := domain.ParseRole(value)
		if err != nil {
			roleRows.Close()
			return application.SignInIdentity{}, err
		}
		identity.Roles = append(identity.Roles, role)
	}
	if err := roleRows.Err(); err != nil {
		roleRows.Close()
		return application.SignInIdentity{}, fmt.Errorf("iterate sign-in roles: %w", err)
	}
	roleRows.Close()

	credentialRows, err := r.tx.Query(ctx, `
		SELECT id, credential_id, public_key, attestation_type, aaguid,
		       sign_count, transports, authenticator_flags, created_at, last_used_at
		FROM passkey_credentials
		WHERE identity_id = $1
		ORDER BY id
	`, identityID)
	if err != nil {
		return application.SignInIdentity{}, fmt.Errorf("load sign-in credentials: %w", err)
	}
	defer credentialRows.Close()
	for credentialRows.Next() {
		var credential application.PasskeyCredential
		var signCount int64
		var authenticatorFlags *int16
		if err := credentialRows.Scan(
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
			return application.SignInIdentity{}, fmt.Errorf("scan sign-in credential: %w", err)
		}
		credential.SignCount = uint32(signCount)
		if authenticatorFlags != nil {
			credential.AuthenticatorFlags = uint8(*authenticatorFlags)
			credential.FlagsKnown = true
		}
		identity.Credentials = append(identity.Credentials, credential)
	}
	if err := credentialRows.Err(); err != nil {
		return application.SignInIdentity{}, fmt.Errorf("iterate sign-in credentials: %w", err)
	}
	return identity, nil
}

func (r *txRepositories) StoreAssertionCeremony(ctx context.Context, ceremony application.AssertionCeremony) error {
	userHandle := ceremony.UserHandle
	if userHandle == nil {
		userHandle = []byte{}
	}
	_, err := r.tx.Exec(ctx, `
		INSERT INTO webauthn_ceremonies (
			id, session_id, purpose, challenge, identity_id, user_handle,
			expires_at, consumed_at, created_at
		)
		VALUES ($1, $2, 'sign_in', $3, NULLIF($4, ''), $5, $6, NULL, now())
	`, ceremony.ID, ceremony.SessionID, ceremony.Challenge, ceremony.IdentityID, userHandle, ceremony.ExpiresAt)
	if err != nil {
		return fmt.Errorf("store assertion ceremony: %w", err)
	}
	return nil
}

func (r *txRepositories) LockAssertionCeremony(ctx context.Context, ceremonyID string) (application.AssertionCeremony, error) {
	var ceremony application.AssertionCeremony
	err := r.tx.QueryRow(ctx, `
		SELECT id, COALESCE(session_id, 0), challenge, COALESCE(identity_id, ''),
		       user_handle, expires_at
		FROM webauthn_ceremonies
		WHERE id = $1 AND purpose = 'sign_in' AND consumed_at IS NULL
		FOR UPDATE
	`, ceremonyID).Scan(
		&ceremony.ID,
		&ceremony.SessionID,
		&ceremony.Challenge,
		&ceremony.IdentityID,
		&ceremony.UserHandle,
		&ceremony.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return application.AssertionCeremony{}, fmt.Errorf("assertion ceremony not found")
	}
	if err != nil {
		return application.AssertionCeremony{}, fmt.Errorf("lock assertion ceremony: %w", err)
	}
	return ceremony, nil
}

func (r *txRepositories) ConsumeAssertionCeremony(ctx context.Context, ceremonyID string, consumedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE webauthn_ceremonies
		SET consumed_at = $2
		WHERE id = $1 AND purpose = 'sign_in' AND consumed_at IS NULL
	`, ceremonyID, consumedAt)
	if err != nil {
		return fmt.Errorf("consume assertion ceremony: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("assertion ceremony not found")
	}
	return nil
}

func (r *txRepositories) UpdatePasskeyAfterAssertion(ctx context.Context, credentialID string, signCount uint32, authenticatorFlags uint8, usedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE passkey_credentials
		SET sign_count = $2, authenticator_flags = $3, last_used_at = $4
		WHERE id = $1
	`, credentialID, int64(signCount), int16(authenticatorFlags), usedAt)
	if err != nil {
		return fmt.Errorf("update passkey assertion state: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("passkey credential not found")
	}
	return nil
}
