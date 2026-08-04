package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
)

var ErrSignInIdentityNotFound = errors.New("sign-in identity not found")

type SignInUnitOfWork interface {
	WithinSignInTx(context.Context, func(context.Context, SignInRepositories) error) error
}

type SignInRepositories interface {
	FindSignInIdentityByEmail(context.Context, string) (SignInIdentity, error)
	FindSignInIdentityByCredential(context.Context, []byte) (SignInIdentity, error)
	StoreAssertionCeremony(context.Context, AssertionCeremony) error
	LockAssertionCeremony(context.Context, string) (AssertionCeremony, error)
	ConsumeAssertionCeremony(context.Context, string, time.Time) error
	UpdatePasskeyAfterAssertion(context.Context, string, uint32, time.Time) error
	WriteAudit(context.Context, AuditEvent) error
	RotateForIdentity(context.Context, int64, string, time.Time) (IssuedSession, error)
}

type PasskeyAssertions interface {
	BeginAssertion(context.Context, *SignInIdentity) (AssertionOptions, error)
	CredentialID([]byte) ([]byte, error)
	VerifyAssertion(context.Context, AssertionVerification) (VerifiedAssertion, error)
}

// SignInService implements UC-2@r2 / US-002@r2.
type SignInService struct {
	UnitOfWork          SignInUnitOfWork
	RateLimiter         RateLimiter
	Passkeys            PasskeyAssertions
	Clock               Clock
	IDs                 IDGenerator
	AuthRateLimitMax    int
	AuthRateLimitWindow time.Duration
}

type BeginSignInCommand struct {
	SessionID    int64
	RateLimitKey string
	EmailHint    string
}

type BeginSignInResult struct {
	CeremonyID string
	PublicKey  any
}

type FinishSignInCommand struct {
	SessionID         int64
	PasskeyCeremonyID string
	PasskeyResponse   []byte
}

type SignInResult struct {
	IdentityID string
	RedirectTo string
	Session    IssuedSession
}

type SignInIdentity struct {
	ID          string
	PersonID    string
	UserHandle  []byte
	Roles       []domain.Role
	Credentials []PasskeyCredential
}

type AssertionOptions struct {
	PublicKey any
	Challenge []byte
}

type AssertionCeremony struct {
	ID         string
	SessionID  int64
	Challenge  []byte
	IdentityID string
	UserHandle []byte
	ExpiresAt  time.Time
}

type AssertionVerification struct {
	Challenge    []byte
	Identity     SignInIdentity
	Discoverable bool
	Response     []byte
}

type VerifiedAssertion struct {
	CredentialID []byte
	SignCount    uint32
}

func (s *SignInService) BeginSignIn(ctx context.Context, command BeginSignInCommand) (BeginSignInResult, error) {
	if err := s.allow(ctx, command.RateLimitKey); err != nil {
		return BeginSignInResult{}, err
	}
	ceremonyID, err := s.IDs.NewID()
	if err != nil {
		return BeginSignInResult{}, fmt.Errorf("create assertion ceremony id: %w", err)
	}
	var result BeginSignInResult
	err = s.UnitOfWork.WithinSignInTx(ctx, func(txCtx context.Context, repos SignInRepositories) error {
		var identity *SignInIdentity
		if command.EmailHint != "" {
			email, err := domain.NewEmail(command.EmailHint)
			if err == nil {
				found, findErr := repos.FindSignInIdentityByEmail(txCtx, email.Normalized())
				switch {
				case findErr == nil:
					identity = &found
				case errors.Is(findErr, ErrSignInIdentityNotFound):
					identity = decoyIdentity(email.Normalized())
				default:
					return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, findErr)
				}
			} else {
				identity = decoyIdentity(domain.NormalizeEmail(command.EmailHint))
				err = nil
			}
			if err != nil {
				return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
			}
		}
		options, err := s.Passkeys.BeginAssertion(txCtx, identity)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
		}
		ceremony := AssertionCeremony{
			ID:        ceremonyID,
			SessionID: command.SessionID,
			Challenge: options.Challenge,
			ExpiresAt: s.now().Add(15 * time.Minute),
		}
		if identity != nil {
			ceremony.IdentityID = identity.ID
			ceremony.UserHandle = append([]byte(nil), identity.UserHandle...)
		}
		if err := repos.StoreAssertionCeremony(txCtx, ceremony); err != nil {
			return err
		}
		result = BeginSignInResult{CeremonyID: ceremonyID, PublicKey: options.PublicKey}
		return nil
	})
	return result, err
}

func (s *SignInService) FinishSignIn(ctx context.Context, command FinishSignInCommand) (SignInResult, error) {
	var result SignInResult
	err := s.UnitOfWork.WithinSignInTx(ctx, func(txCtx context.Context, repos SignInRepositories) error {
		ceremony, err := repos.LockAssertionCeremony(txCtx, command.PasskeyCeremonyID)
		if err != nil {
			return signInFailure(err)
		}
		now := s.now()
		if ceremony.SessionID != command.SessionID || !ceremony.ExpiresAt.After(now) {
			return signInFailure(errors.New("assertion ceremony mismatch or expiry"))
		}
		credentialID, err := s.Passkeys.CredentialID(command.PasskeyResponse)
		if err != nil || len(credentialID) == 0 {
			return signInFailure(err)
		}
		identity, err := repos.FindSignInIdentityByCredential(txCtx, credentialID)
		if err != nil || identity.ID == "" || len(identity.Roles) == 0 {
			return signInFailure(err)
		}
		if ceremony.IdentityID != "" && ceremony.IdentityID != identity.ID {
			return signInFailure(errors.New("hinted identity mismatch"))
		}
		verified, err := s.Passkeys.VerifyAssertion(txCtx, AssertionVerification{
			Challenge:    ceremony.Challenge,
			Identity:     identity,
			Discoverable: ceremony.IdentityID == "",
			Response:     command.PasskeyResponse,
		})
		if err != nil || !bytes.Equal(verified.CredentialID, credentialID) {
			return signInFailure(err)
		}
		credential, ok := credentialByExternalID(identity.Credentials, credentialID)
		if !ok {
			return signInFailure(errors.New("credential missing from identity"))
		}
		if err := repos.UpdatePasskeyAfterAssertion(txCtx, credential.ID, verified.SignCount, now); err != nil {
			return err
		}
		if err := repos.ConsumeAssertionCeremony(txCtx, command.PasskeyCeremonyID, now); err != nil {
			return signInFailure(err)
		}
		if err := repos.WriteAudit(txCtx, AuditEvent{
			ActorID:       identity.ID,
			Action:        "identity.sign_in.completed",
			TargetType:    "identity",
			TargetID:      identity.ID,
			CorrelationID: credential.ID,
			Payload:       map[string]any{},
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		session, err := repos.RotateForIdentity(txCtx, command.SessionID, identity.ID, now)
		if err != nil {
			return err
		}
		redirectTo, err := landingFor(identity.Roles)
		if err != nil {
			return err
		}
		result = SignInResult{IdentityID: identity.ID, RedirectTo: redirectTo, Session: session}
		return nil
	})
	return result, err
}

func (s *SignInService) allow(ctx context.Context, key string) error {
	allowed, err := s.RateLimiter.Allow(ctx, key, s.rateLimitMax(), s.rateLimitWindow())
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}

func (s *SignInService) now() time.Time {
	if s.Clock == nil {
		return time.Now().UTC()
	}
	return s.Clock.Now().UTC()
}

func (s *SignInService) rateLimitMax() int {
	if s.AuthRateLimitMax <= 0 {
		return 10
	}
	return s.AuthRateLimitMax
}

func (s *SignInService) rateLimitWindow() time.Duration {
	if s.AuthRateLimitWindow <= 0 {
		return 15 * time.Minute
	}
	return s.AuthRateLimitWindow
}

func decoyIdentity(normalizedHint string) *SignInIdentity {
	userHandle := sha256.Sum256([]byte("sign-in-decoy-user:" + normalizedHint))
	credentialID := sha256.Sum256([]byte("sign-in-decoy-credential:" + normalizedHint))
	return &SignInIdentity{
		UserHandle: userHandle[:],
		Credentials: []PasskeyCredential{{
			CredentialID: credentialID[:],
		}},
	}
}

func signInFailure(err error) error {
	if err == nil {
		err = errors.New("sign-in failed")
	}
	return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
}

func credentialByExternalID(credentials []PasskeyCredential, credentialID []byte) (PasskeyCredential, bool) {
	for _, credential := range credentials {
		if bytes.Equal(credential.CredentialID, credentialID) {
			return credential, true
		}
	}
	return PasskeyCredential{}, false
}

func landingFor(roles []domain.Role) (string, error) {
	for _, role := range roles {
		if role == domain.RoleFamilyManager {
			return "/family", nil
		}
	}
	for _, role := range roles {
		if role == domain.RoleYoungAdultScout {
			return "/scout/schedule", nil
		}
	}
	for _, role := range roles {
		if role == domain.RoleAdmin || role == domain.RoleCommittee {
			return "/account", nil
		}
	}
	return "", domain.ErrInvalidRole
}
