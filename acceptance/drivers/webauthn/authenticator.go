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

// Passkey retains virtual authenticator state across registration and sign-in.
type Passkey struct {
	rp            RelyingParty
	authenticator virtualwebauthn.Authenticator
	credential    virtualwebauthn.Credential
}

// NewPasskey creates a virtual discoverable credential for one relying party.
func NewPasskey(rp RelyingParty) *Passkey {
	return &Passkey{
		rp:            rp,
		authenticator: virtualwebauthn.NewAuthenticator(),
		credential:    virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2),
	}
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

// AllowCredentialsShape extracts the public allowCredentials field for
// account-existence oracle checks without decoding credential IDs as secrets.
func AllowCredentialsShape(publicKey json.RawMessage) (json.RawMessage, int, error) {
	var wrapped struct {
		PublicKey json.RawMessage `json:"publicKey"`
	}
	raw := publicKey
	if err := json.Unmarshal(publicKey, &wrapped); err == nil && len(wrapped.PublicKey) > 0 {
		raw = wrapped.PublicKey
	}
	var options struct {
		AllowCredentials json.RawMessage `json:"allowCredentials"`
	}
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil, 0, fmt.Errorf("decode assertion publicKey: %w", err)
	}
	if len(options.AllowCredentials) == 0 || string(options.AllowCredentials) == "null" {
		return json.RawMessage("[]"), 0, nil
	}
	var credentials []json.RawMessage
	if err := json.Unmarshal(options.AllowCredentials, &credentials); err != nil {
		return nil, 0, fmt.Errorf("decode allowCredentials: %w", err)
	}
	return options.AllowCredentials, len(credentials), nil
}

// CreateAttestationResponse builds a navigator.credentials.create-shaped
// attestation response for the supplied begin payload.
func CreateAttestationResponse(rp RelyingParty, publicKey json.RawMessage) (string, error) {
	return NewPasskey(rp).CreateAttestationResponse(publicKey)
}

// CreateAttestationResponse registers this passkey with the supplied options.
func (p *Passkey) CreateAttestationResponse(publicKey json.RawMessage) (string, error) {
	options, err := virtualwebauthn.ParseAttestationOptions(string(publicKey))
	if err != nil {
		return "", fmt.Errorf("parse attestation options: %w", err)
	}
	if options.RelyingPartyID != p.rp.ID {
		return "", fmt.Errorf("relying party id = %q, want %q", options.RelyingPartyID, p.rp.ID)
	}
	p.authenticator.Options.UserHandle = append([]byte(nil), options.UserID...)
	p.authenticator.AddCredential(p.credential)
	return virtualwebauthn.CreateAttestationResponse(p.virtualRP(), p.authenticator, p.credential, *options), nil
}

// CreateAssertionResponse builds a navigator.credentials.get-shaped response.
func (p *Passkey) CreateAssertionResponse(publicKey json.RawMessage) (string, error) {
	options, err := virtualwebauthn.ParseAssertionOptions(string(publicKey))
	if err != nil {
		return "", fmt.Errorf("parse assertion options: %w", err)
	}
	if options.RelyingPartyID != p.rp.ID {
		return "", fmt.Errorf("relying party id = %q, want %q", options.RelyingPartyID, p.rp.ID)
	}
	p.credential.Counter++
	return virtualwebauthn.CreateAssertionResponse(p.virtualRP(), p.authenticator, p.credential, *options), nil
}

func (p *Passkey) virtualRP() virtualwebauthn.RelyingParty {
	return virtualwebauthn.RelyingParty{
		Name:   p.rp.Name,
		ID:     p.rp.ID,
		Origin: p.rp.Origin,
	}
}
