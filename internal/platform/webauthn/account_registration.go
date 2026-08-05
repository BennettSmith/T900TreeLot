package webauthn

import (
	"context"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/troop900/treelot/internal/identity/application"
)

// BeginAccountRegistration builds registration options without persisting a ceremony.
func (c *RegistrationCeremony) BeginAccountRegistration(_ context.Context, start application.AccountRegistrationStart) (application.RegistrationOptions, error) {
	user := registrationUser{
		id:          start.UserHandle,
		name:        start.Email.Normalized(),
		displayName: start.DisplayName,
	}
	descriptors := make([]protocol.CredentialDescriptor, 0, len(start.ExcludeCredentials))
	for _, credentialID := range start.ExcludeCredentials {
		descriptors = append(descriptors, protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: credentialID,
		})
	}
	creation, sessionData, err := c.webauthn.BeginRegistration(
		user,
		gowebauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		gowebauthn.WithExclusions(descriptors),
	)
	if err != nil {
		return application.RegistrationOptions{}, err
	}
	expiresAt := start.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = c.clock.Now().Add(15 * time.Minute)
	}
	return application.RegistrationOptions{
		CeremonyID: start.CeremonyID,
		PublicKey:  creation,
		Challenge:  []byte(sessionData.Challenge),
		UserHandle: start.UserHandle,
		ExpiresAt:  expiresAt,
	}, nil
}

// AccountRegistration adapts RegistrationCeremony to AccountPasskeyRegistration.
type AccountRegistration struct {
	inner *RegistrationCeremony
}

func NewAccountRegistration(inner *RegistrationCeremony) *AccountRegistration {
	return &AccountRegistration{inner: inner}
}

func (a *AccountRegistration) BeginRegistration(ctx context.Context, start application.AccountRegistrationStart) (application.RegistrationOptions, error) {
	return a.inner.BeginAccountRegistration(ctx, start)
}

func (a *AccountRegistration) VerifyRegistration(ctx context.Context, verification application.RegistrationVerification) (application.PasskeyCredential, error) {
	return a.inner.VerifyRegistration(ctx, verification)
}
