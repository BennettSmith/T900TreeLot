package webauthn

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/troop900/treelot/internal/identity/application"
)

// AssertionCeremony adapts WebAuthn authentication without owning persistence.
type AssertionCeremony struct {
	webauthn *gowebauthn.WebAuthn
}

func NewAssertionCeremony(rpID string, origins []string) (*AssertionCeremony, error) {
	wa, err := gowebauthn.New(&gowebauthn.Config{
		RPID:          rpID,
		RPDisplayName: "Troop 900 Tree Lot",
		RPOrigins:     origins,
	})
	if err != nil {
		return nil, err
	}
	return &AssertionCeremony{webauthn: wa}, nil
}

func (c *AssertionCeremony) BeginAssertion(_ context.Context, identity *application.SignInIdentity) (application.AssertionOptions, error) {
	var (
		assertion *protocol.CredentialAssertion
		session   *gowebauthn.SessionData
		err       error
	)
	if identity == nil {
		assertion, session, err = c.webauthn.BeginDiscoverableLogin()
	} else {
		assertion, session, err = c.webauthn.BeginLogin(assertionUserFrom(*identity))
	}
	if err != nil {
		return application.AssertionOptions{}, err
	}
	return application.AssertionOptions{
		PublicKey: assertion,
		Challenge: []byte(session.Challenge),
	}, nil
}

func (*AssertionCeremony) CredentialID(response []byte) ([]byte, error) {
	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), parsed.RawID...), nil
}

func (c *AssertionCeremony) VerifyAssertion(_ context.Context, verification application.AssertionVerification) (application.VerifiedAssertion, error) {
	parsed, err := protocol.ParseCredentialRequestResponseBytes(verification.Response)
	if err != nil {
		return application.VerifiedAssertion{}, err
	}
	user := assertionUserFrom(verification.Identity)
	session := gowebauthn.SessionData{
		Challenge:            string(verification.Challenge),
		RelyingPartyID:       c.webauthn.Config.RPID,
		UserVerification:     c.webauthn.Config.AuthenticatorSelection.UserVerification,
		AllowedCredentialIDs: nil,
	}
	var credential *gowebauthn.Credential
	if verification.Discoverable {
		_, credential, err = c.webauthn.ValidatePasskeyLogin(func(rawID, userHandle []byte) (gowebauthn.User, error) {
			if !bytes.Equal(userHandle, verification.Identity.UserHandle) || !user.hasCredential(rawID) {
				return nil, fmt.Errorf("discoverable credential does not match identity")
			}
			return user, nil
		}, session, parsed)
	} else {
		session.UserID = verification.Identity.UserHandle
		session.AllowedCredentialIDs = credentialIDs(verification.Identity.Credentials)
		credential, err = c.webauthn.ValidateLogin(user, session, parsed)
	}
	if err != nil {
		return application.VerifiedAssertion{}, err
	}
	return application.VerifiedAssertion{
		CredentialID: append([]byte(nil), credential.ID...),
		SignCount:    credential.Authenticator.SignCount,
	}, nil
}

type assertionUser struct {
	id          []byte
	credentials []gowebauthn.Credential
}

func assertionUserFrom(identity application.SignInIdentity) assertionUser {
	credentials := make([]gowebauthn.Credential, 0, len(identity.Credentials))
	for _, stored := range identity.Credentials {
		transports := make([]protocol.AuthenticatorTransport, 0, len(stored.Transports))
		for _, transport := range stored.Transports {
			transports = append(transports, protocol.AuthenticatorTransport(transport))
		}
		credentials = append(credentials, gowebauthn.Credential{
			ID:              append([]byte(nil), stored.CredentialID...),
			PublicKey:       append([]byte(nil), stored.PublicKey...),
			AttestationType: stored.AttestationType,
			Transport:       transports,
			Authenticator: gowebauthn.Authenticator{
				SignCount: stored.SignCount,
			},
		})
	}
	return assertionUser{id: append([]byte(nil), identity.UserHandle...), credentials: credentials}
}

func (u assertionUser) WebAuthnID() []byte                           { return u.id }
func (assertionUser) WebAuthnName() string                           { return "Tree Lot volunteer" }
func (assertionUser) WebAuthnDisplayName() string                    { return "Tree Lot volunteer" }
func (u assertionUser) WebAuthnCredentials() []gowebauthn.Credential { return u.credentials }

func (u assertionUser) hasCredential(rawID []byte) bool {
	for _, credential := range u.credentials {
		if bytes.Equal(credential.ID, rawID) {
			return true
		}
	}
	return false
}

func credentialIDs(credentials []application.PasskeyCredential) [][]byte {
	ids := make([][]byte, 0, len(credentials))
	for _, credential := range credentials {
		ids = append(ids, append([]byte(nil), credential.CredentialID...))
	}
	return ids
}
