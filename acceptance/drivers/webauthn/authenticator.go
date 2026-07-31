//go:build acceptance

// Package webauthn provides a cryptographic virtual authenticator for
// browser-less acceptance of passkey registration ceremonies.
package webauthn

import (
	"encoding/json"
	"fmt"

	"github.com/descope/virtualwebauthn"
)

// RelyingParty mirrors the deployed WebAuthn configuration under test.
type RelyingParty struct {
	Name   string
	ID     string
	Origin string
}

// BeginPayload is the JSON returned by POST /bootstrap/passkey/begin.
type BeginPayload struct {
	CeremonyID string
	PublicKey  json.RawMessage
}

// ParseBeginPayload decodes the ceremony begin response.
func ParseBeginPayload(body string) (BeginPayload, error) {
	var begin BeginPayload
	if err := json.Unmarshal([]byte(body), &begin); err != nil {
		return BeginPayload{}, fmt.Errorf("decode begin payload: %w", err)
	}
	if begin.CeremonyID == "" || len(begin.PublicKey) == 0 {
		return BeginPayload{}, fmt.Errorf("begin payload missing ceremonyId or publicKey")
	}
	return begin, nil
}

// CreateAttestationResponse builds a navigator.credentials.create-shaped
// attestation response for the supplied begin payload.
func CreateAttestationResponse(rp RelyingParty, publicKey json.RawMessage) (string, error) {
	options, err := virtualwebauthn.ParseAttestationOptions(string(publicKey))
	if err != nil {
		return "", fmt.Errorf("parse attestation options: %w", err)
	}
	if options.RelyingPartyID != rp.ID {
		return "", fmt.Errorf("relying party id = %q, want %q", options.RelyingPartyID, rp.ID)
	}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
	return virtualwebauthn.CreateAttestationResponse(virtualwebauthn.RelyingParty{
		Name:   rp.Name,
		ID:     rp.ID,
		Origin: rp.Origin,
	}, authenticator, credential, *options), nil
}
