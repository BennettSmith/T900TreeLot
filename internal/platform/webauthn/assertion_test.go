package webauthn_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/descope/virtualwebauthn"
	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/testdb"
	platformwebauthn "github.com/troop900/treelot/internal/platform/webauthn"
)

func TestAssertionCeremonyVerifiesDiscoverableCredential(t *testing.T) {
	db := testdb.OpenMigrated(t)
	rp := virtualwebauthn.RelyingParty{
		Name:   "Troop 900 Tree Lot",
		ID:     "treelot.test",
		Origin: "https://treelot.test",
	}
	userHandle := []byte("stable-user-handle")
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	registration, err := platformwebauthn.NewRegistrationCeremony(db, clock.System(), rp.ID, []string{rp.Origin})
	if err != nil {
		t.Fatal(err)
	}
	started, err := registration.BeginRegistration(context.Background(), application.RegistrationStart{
		CeremonyID:  "registration-1",
		Email:       mustEmail(t, "manager@example.org"),
		DisplayName: "Family Manager",
		UserHandle:  userHandle,
		ExpiresAt:   time.Now().Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	creationJSON, err := json.Marshal(started.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	creation, err := virtualwebauthn.ParseAttestationOptions(string(creationJSON))
	if err != nil {
		t.Fatal(err)
	}
	attestation := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *creation)
	stored, err := registration.VerifyRegistration(context.Background(), application.RegistrationVerification{
		Challenge:  []byte(base64.RawURLEncoding.EncodeToString(creation.Challenge)),
		UserHandle: userHandle,
		Response:   []byte(attestation),
	})
	if err != nil {
		t.Fatal(err)
	}

	assertions, err := platformwebauthn.NewAssertionCeremony(rp.ID, []string{rp.Origin})
	if err != nil {
		t.Fatal(err)
	}
	identity := application.SignInIdentity{
		ID:         "identity-1",
		UserHandle: userHandle,
		Credentials: []application.PasskeyCredential{{
			ID:              "credential-1",
			CredentialID:    stored.CredentialID,
			PublicKey:       stored.PublicKey,
			AttestationType: stored.AttestationType,
			AAGUID:          stored.AAGUID,
			SignCount:       stored.SignCount,
			Transports:      stored.Transports,
		}},
	}
	options, err := assertions.BeginAssertion(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	optionsJSON, err := json.Marshal(options.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	request, err := virtualwebauthn.ParseAssertionOptions(string(optionsJSON))
	if err != nil {
		t.Fatal(err)
	}
	authenticator.Options.UserHandle = userHandle
	authenticator.AddCredential(credential)
	credential.Counter++
	response := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *request)

	credentialID, err := assertions.CredentialID([]byte(response))
	if err != nil {
		t.Fatal(err)
	}
	verified, err := assertions.VerifyAssertion(context.Background(), application.AssertionVerification{
		Challenge:    options.Challenge,
		Identity:     identity,
		Discoverable: true,
		Response:     []byte(response),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(credentialID) != string(stored.CredentialID) || verified.SignCount != 1 {
		t.Fatalf("credential id/count = %x/%d", credentialID, verified.SignCount)
	}

	identity.Credentials[0].SignCount = verified.SignCount
	hintedOptions, err := assertions.BeginAssertion(context.Background(), &identity)
	if err != nil {
		t.Fatal(err)
	}
	hintedJSON, err := json.Marshal(hintedOptions.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	hintedRequest, err := virtualwebauthn.ParseAssertionOptions(string(hintedJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(hintedRequest.AllowCredentials) != 1 {
		t.Fatalf("hinted allowed credentials = %#v", hintedRequest.AllowCredentials)
	}
	credential.Counter++
	hintedResponse := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *hintedRequest)
	hintedVerified, err := assertions.VerifyAssertion(context.Background(), application.AssertionVerification{
		Challenge: hintedOptions.Challenge,
		Identity:  identity,
		Response:  []byte(hintedResponse),
	})
	if err != nil {
		t.Fatal(err)
	}
	if hintedVerified.SignCount != 2 {
		t.Fatalf("hinted sign count = %d", hintedVerified.SignCount)
	}
}

func mustEmail(t *testing.T, value string) domain.Email {
	t.Helper()
	email, err := domain.NewEmail(value)
	if err != nil {
		t.Fatal(err)
	}
	return email
}
