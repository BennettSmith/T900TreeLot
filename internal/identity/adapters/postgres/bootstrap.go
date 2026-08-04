// Package postgres implements Identity persistence adapters.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	familypostgres "github.com/troop900/treelot/internal/families/adapters/postgres"
	families "github.com/troop900/treelot/internal/families/application"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/audit"
	"github.com/troop900/treelot/internal/platform/clock"
	platformpostgres "github.com/troop900/treelot/internal/platform/postgres"
	"github.com/troop900/treelot/internal/platform/session"
)

type UnitOfWork struct {
	db       *platformpostgres.DB
	sessions *session.Store
}

func NewUnitOfWork(db *platformpostgres.DB, clk clock.Clock) *UnitOfWork {
	return &UnitOfWork{db: db, sessions: session.NewStore(db, clk, 24*time.Hour)}
}

func (u *UnitOfWork) WithinTx(ctx context.Context, fn func(context.Context, application.Repositories) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin identity transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, &txRepositories{tx: tx, sessions: u.sessions}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit identity transaction: %w", err)
	}
	return nil
}

func (u *UnitOfWork) WithinTestFixtureTx(ctx context.Context, fn func(context.Context, application.TestFixtureRepositories) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin identity fixture transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, &txRepositories{tx: tx, sessions: u.sessions}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit identity fixture transaction: %w", err)
	}
	return nil
}

func (u *UnitOfWork) WithinSignInTx(ctx context.Context, fn func(context.Context, application.SignInRepositories) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin sign-in transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, &txRepositories{tx: tx, sessions: u.sessions}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit sign-in transaction: %w", err)
	}
	return nil
}

type txRepositories struct {
	tx       pgx.Tx
	sessions *session.Store
}

func (r *txRepositories) AdminExists(ctx context.Context) (bool, error) {
	var exists bool
	err := r.tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM identity_roles
			WHERE role = 'admin'
		)
	`).Scan(&exists)
	return exists, err
}

func (r *txRepositories) LockBootstrap(ctx context.Context) (application.BootstrapState, error) {
	var closedAt *time.Time
	err := r.tx.QueryRow(ctx, `
		SELECT closed_at
		FROM bootstrap_state
		WHERE id = 1
		FOR UPDATE
	`).Scan(&closedAt)
	if err != nil {
		return application.BootstrapState{}, fmt.Errorf("lock bootstrap state: %w", err)
	}
	return application.BootstrapState{Closed: closedAt != nil}, nil
}

func (r *txRepositories) EmailTaken(ctx context.Context, normalized string) (bool, error) {
	var exists bool
	err := r.tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM identity_emails
			WHERE email_normalized = $1 AND active
		)
	`, normalized).Scan(&exists)
	return exists, err
}

func (r *txRepositories) FindIdentityByEmail(ctx context.Context, normalized string) (string, error) {
	var identityID string
	err := r.tx.QueryRow(ctx, `
		SELECT identity_id
		FROM identity_emails
		WHERE email_normalized = $1 AND active
	`, normalized).Scan(&identityID)
	if err != nil {
		return "", fmt.Errorf("find identity by email: %w", err)
	}
	return identityID, nil
}

func (r *txRepositories) ReplaceRoles(ctx context.Context, identityID string, roles []domain.Role) error {
	if _, err := r.tx.Exec(ctx, `DELETE FROM identity_roles WHERE identity_id = $1`, identityID); err != nil {
		return fmt.Errorf("clear identity roles: %w", err)
	}
	for _, role := range roles {
		if err := r.GrantRole(ctx, identityID, role); err != nil {
			return err
		}
	}
	return nil
}

func (r *txRepositories) RevokeSessionsForIdentity(ctx context.Context, identityID string) error {
	_, err := r.tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE identity_id = $1 AND revoked_at IS NULL
	`, identityID)
	if err != nil {
		return fmt.Errorf("revoke identity sessions: %w", err)
	}
	return nil
}

func (r *txRepositories) LockRegistrationCeremony(ctx context.Context, ceremonyID string) (application.RegistrationCeremony, error) {
	var ceremony application.RegistrationCeremony
	var email, firstName, lastName, preferredDisplayName string
	err := r.tx.QueryRow(ctx, `
		SELECT COALESCE(session_id, 0), challenge, user_handle, expires_at, bootstrap_email,
		       bootstrap_first_name, bootstrap_last_name,
		       COALESCE(bootstrap_preferred_display_name, '')
		FROM webauthn_ceremonies
		WHERE id = $1
		  AND purpose = 'bootstrap_registration'
		  AND consumed_at IS NULL
		FOR UPDATE
	`, ceremonyID).Scan(
		&ceremony.SessionID,
		&ceremony.Challenge,
		&ceremony.UserHandle,
		&ceremony.ExpiresAt,
		&email,
		&firstName,
		&lastName,
		&preferredDisplayName,
	)
	if err == pgx.ErrNoRows {
		return application.RegistrationCeremony{}, fmt.Errorf("webauthn ceremony not found")
	}
	if err != nil {
		return application.RegistrationCeremony{}, fmt.Errorf("load webauthn ceremony: %w", err)
	}
	ceremony.Email, err = domain.NewEmail(email)
	if err != nil {
		return application.RegistrationCeremony{}, fmt.Errorf("load webauthn ceremony email: %w", err)
	}
	ceremony.Name, err = domain.ValidateProfile(firstName, lastName, preferredDisplayName)
	if err != nil {
		return application.RegistrationCeremony{}, fmt.Errorf("load webauthn ceremony profile: %w", err)
	}
	return ceremony, nil
}

func (r *txRepositories) ConsumeRegistrationCeremony(ctx context.Context, ceremonyID string, consumedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE webauthn_ceremonies
		SET consumed_at = $2
		WHERE id = $1
		  AND purpose = 'bootstrap_registration'
		  AND consumed_at IS NULL
	`, ceremonyID, consumedAt)
	if err != nil {
		return fmt.Errorf("consume webauthn ceremony: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webauthn ceremony not found")
	}
	return nil
}

func (r *txRepositories) CreatePersonalProfile(ctx context.Context, profile application.PersonalProfile) error {
	return familypostgres.NewTxProfileCreator(r.tx).CreatePersonalProfile(ctx, families.PersonalProfile{
		ID:                   profile.ID,
		FirstName:            profile.FirstName,
		LastName:             profile.LastName,
		PreferredDisplayName: profile.PreferredDisplayName,
		CreatedAt:            profile.CreatedAt,
		UpdatedAt:            profile.UpdatedAt,
	})
}

func (r *txRepositories) CreateIdentity(ctx context.Context, record application.IdentityRecord) error {
	_, err := r.tx.Exec(ctx, `
		INSERT INTO identities (id, person_id, webauthn_user_handle, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, '\x'::bytea), $4, $5)
	`, record.ID, record.PersonID, record.UserHandle, record.CreatedAt, record.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create identity: %w", err)
	}
	return nil
}

func (r *txRepositories) AddEmail(ctx context.Context, email application.IdentityEmail) error {
	_, err := r.tx.Exec(ctx, `
		INSERT INTO identity_emails (identity_id, email, email_normalized, verified_at, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, email.IdentityID, email.Email, email.Normalized, email.VerifiedAt, email.Active, email.CreatedAt, email.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("add identity email: %w", err)
	}
	return nil
}

func (r *txRepositories) GrantRole(ctx context.Context, identityID string, role domain.Role) error {
	_, err := r.tx.Exec(ctx, `
		INSERT INTO identity_roles (identity_id, role, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (identity_id, role) DO NOTHING
	`, identityID, string(role))
	if err != nil {
		return fmt.Errorf("grant identity role: %w", err)
	}
	return nil
}

func (r *txRepositories) StorePasskeyCredential(ctx context.Context, credential application.PasskeyCredential) error {
	_, err := r.tx.Exec(ctx, `
		INSERT INTO passkey_credentials (
			id, identity_id, credential_id, public_key, attestation_type,
			aaguid, sign_count, transports, created_at, last_used_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, credential.ID, credential.IdentityID, credential.CredentialID, credential.PublicKey, credential.AttestationType,
		credential.AAGUID, int64(credential.SignCount), credential.Transports, credential.CreatedAt, credential.LastUsedAt)
	if err != nil {
		return fmt.Errorf("store passkey credential: %w", err)
	}
	return nil
}

func (r *txRepositories) CloseBootstrap(ctx context.Context, identityID string, closedAt time.Time) error {
	tag, err := r.tx.Exec(ctx, `
		UPDATE bootstrap_state
		SET closed_at = $1, closed_by_identity_id = $2
		WHERE id = 1 AND closed_at IS NULL
	`, closedAt, identityID)
	if err != nil {
		return fmt.Errorf("close bootstrap: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBootstrapClosed
	}
	return nil
}

func (r *txRepositories) WriteAudit(ctx context.Context, event application.AuditEvent) error {
	return audit.NewTxWriter(r.tx).Write(ctx, audit.Event{
		ActorID:       event.ActorID,
		Action:        event.Action,
		TargetType:    event.TargetType,
		TargetID:      event.TargetID,
		CorrelationID: event.CorrelationID,
		Payload:       event.Payload,
		CreatedAt:     event.CreatedAt,
	})
}

func (r *txRepositories) RotateForIdentity(ctx context.Context, oldSessionID int64, identityID string, _ time.Time) (application.IssuedSession, error) {
	created, rawToken, err := r.sessions.RotateForIdentityInTx(ctx, r.tx, oldSessionID, identityID)
	if err != nil {
		return application.IssuedSession{}, err
	}
	authenticatedAt := time.Time{}
	if created.AuthenticatedAt != nil {
		authenticatedAt = *created.AuthenticatedAt
	}
	return application.IssuedSession{
		ID:              created.ID,
		IdentityID:      created.IdentityID,
		RawToken:        rawToken,
		AuthenticatedAt: authenticatedAt,
		ExpiresAt:       created.ExpiresAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
