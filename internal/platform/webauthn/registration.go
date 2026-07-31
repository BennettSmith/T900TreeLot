// Package webauthn adapts go-webauthn ceremonies to Identity ports.
package webauthn

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

const purposeBootstrapRegistration = "bootstrap_registration"

type RegistrationCeremony struct {
	db       *postgres.DB
	clock    clock.Clock
	webauthn *gowebauthn.WebAuthn
}

func NewRegistrationCeremony(db *postgres.DB, clk clock.Clock, rpID string, origins []string) (*RegistrationCeremony, error) {
	if clk == nil {
		clk = clock.System()
	}
	wa, err := gowebauthn.New(&gowebauthn.Config{
		RPID:          rpID,
		RPDisplayName: "Troop 900 Tree Lot",
		RPOrigins:     origins,
	})
	if err != nil {
		return nil, err
	}
	return &RegistrationCeremony{db: db, clock: clk, webauthn: wa}, nil
}

func (c *RegistrationCeremony) BeginRegistration(ctx context.Context, start application.RegistrationStart) (application.RegistrationOptions, error) {
	if start.CeremonyID == "" || len(start.UserHandle) == 0 {
		return application.RegistrationOptions{}, fmt.Errorf("ceremony id and user handle are required")
	}
	user := registrationUser{
		id:          start.UserHandle,
		name:        start.Email.Normalized(),
		displayName: start.DisplayName,
	}
	creation, sessionData, err := c.webauthn.BeginRegistration(user)
	if err != nil {
		return application.RegistrationOptions{}, err
	}
	expiresAt := start.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = c.clock.Now().Add(15 * time.Minute)
	}
	_, err = c.db.Exec(ctx, `
		INSERT INTO webauthn_ceremonies (
			id, session_id, purpose, challenge, identity_id, user_handle,
			expires_at, consumed_at, created_at
		)
		VALUES ($1, NULLIF($2, 0), $3, $4, NULL, $5, $6, NULL, $7)
	`, start.CeremonyID, start.SessionID, purposeBootstrapRegistration, []byte(sessionData.Challenge), start.UserHandle, expiresAt, c.clock.Now())
	if err != nil {
		return application.RegistrationOptions{}, fmt.Errorf("store webauthn ceremony: %w", err)
	}
	return application.RegistrationOptions{
		CeremonyID: start.CeremonyID,
		PublicKey:  creation,
		UserHandle: start.UserHandle,
		ExpiresAt:  expiresAt,
	}, nil
}

func (c *RegistrationCeremony) FinishRegistration(ctx context.Context, finish application.RegistrationFinish) (application.PasskeyCredential, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return application.PasskeyCredential{}, fmt.Errorf("begin webauthn finish: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var challenge []byte
	var userHandle []byte
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT challenge, user_handle, expires_at
		FROM webauthn_ceremonies
		WHERE id = $1
		  AND purpose = $2
		  AND consumed_at IS NULL
		FOR UPDATE
	`, finish.CeremonyID, purposeBootstrapRegistration).Scan(&challenge, &userHandle, &expiresAt)
	if err == pgx.ErrNoRows {
		return application.PasskeyCredential{}, fmt.Errorf("webauthn ceremony not found")
	}
	if err != nil {
		return application.PasskeyCredential{}, fmt.Errorf("load webauthn ceremony: %w", err)
	}
	if !expiresAt.After(c.clock.Now()) {
		return application.PasskeyCredential{}, fmt.Errorf("webauthn ceremony expired")
	}

	sessionData := gowebauthn.SessionData{
		Challenge:      string(challenge),
		RelyingPartyID: c.webauthn.Config.RPID,
		UserID:         userHandle,
		CredParams:     gowebauthn.CredentialParametersDefault(),
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(finish.Response)
	if err != nil {
		return application.PasskeyCredential{}, err
	}
	credential, err := c.webauthn.CreateCredential(registrationUser{id: userHandle}, sessionData, parsed)
	if err != nil {
		return application.PasskeyCredential{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE webauthn_ceremonies
		SET consumed_at = $2
		WHERE id = $1
	`, finish.CeremonyID, c.clock.Now()); err != nil {
		return application.PasskeyCredential{}, fmt.Errorf("consume webauthn ceremony: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return application.PasskeyCredential{}, fmt.Errorf("commit webauthn finish: %w", err)
	}

	transports := make([]string, 0, len(credential.Transport))
	for _, transport := range credential.Transport {
		transports = append(transports, string(transport))
	}
	return application.PasskeyCredential{
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          hex.EncodeToString(credential.Authenticator.AAGUID),
		SignCount:       credential.Authenticator.SignCount,
		Transports:      transports,
	}, nil
}

type registrationUser struct {
	id          []byte
	name        string
	displayName string
}

func (u registrationUser) WebAuthnID() []byte {
	return u.id
}

func (u registrationUser) WebAuthnName() string {
	return u.name
}

func (u registrationUser) WebAuthnDisplayName() string {
	if u.displayName != "" {
		return u.displayName
	}
	return u.name
}

func (registrationUser) WebAuthnCredentials() []gowebauthn.Credential {
	return nil
}
