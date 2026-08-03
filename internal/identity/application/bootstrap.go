// Package application orchestrates Identity and Access use cases.
package application

import (
	"context"
	"fmt"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
)

// UnitOfWork provides one transaction for state-changing bootstrap operations.
type UnitOfWork interface {
	WithinTx(context.Context, func(context.Context, Repositories) error) error
}

// Repositories are the persistence ports consumed inside a bootstrap transaction.
type Repositories interface {
	AdminExists(context.Context) (bool, error)
	LockBootstrap(context.Context) (BootstrapState, error)
	EmailTaken(context.Context, string) (bool, error)
	LockRegistrationCeremony(context.Context, string) (RegistrationCeremony, error)
	ConsumeRegistrationCeremony(context.Context, string, time.Time) error
	CreatePersonalProfile(context.Context, PersonalProfile) error
	CreateIdentity(context.Context, IdentityRecord) error
	AddEmail(context.Context, IdentityEmail) error
	GrantRole(context.Context, string, domain.Role) error
	StorePasskeyCredential(context.Context, PasskeyCredential) error
	CloseBootstrap(context.Context, string, time.Time) error
	WriteAudit(context.Context, AuditEvent) error
	RotateForIdentity(context.Context, int64, string, time.Time) (IssuedSession, error)
}

type BootstrapTokenValidator interface {
	ValidateBootstrapToken(context.Context, string, time.Time) error
}

type RateLimiter interface {
	Allow(context.Context, string, int, time.Duration) (bool, error)
}

type PasskeyCeremony interface {
	BeginRegistration(context.Context, RegistrationStart) (RegistrationOptions, error)
	VerifyRegistration(context.Context, RegistrationVerification) (PasskeyCredential, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

// BootstrapService implements US-001@r2 / UC-0@r2.
type BootstrapService struct {
	UnitOfWork              UnitOfWork
	Tokens                  BootstrapTokenValidator
	RateLimiter             RateLimiter
	Passkeys                PasskeyCeremony
	Clock                   Clock
	IDs                     IDGenerator
	AuthRateLimitMax        int
	AuthRateLimitWindow     time.Duration
	BootstrapTokenExpiresIn time.Duration
}

type StartBootstrapCommand struct {
	Token        string
	RateLimitKey string
}

type StartBootstrapResult struct {
	ExpiresAt time.Time
}

type ClaimBootstrapProfileCommand struct {
	Token                string
	RateLimitKey         string
	Email                string
	FirstName            string
	LastName             string
	PreferredDisplayName string
}

type PendingEnrollment struct {
	Token      string
	Email      domain.Email
	Name       domain.ProfileName
	ValidUntil time.Time
}

type BeginPasskeyRegistrationCommand struct {
	Pending   PendingEnrollment
	SessionID int64
}

type RegistrationStart struct {
	SessionID            int64
	CeremonyID           string
	Email                domain.Email
	FirstName            string
	LastName             string
	PreferredDisplayName string
	DisplayName          string
	UserHandle           []byte
	ExpiresAt            time.Time
}

type RegistrationOptions struct {
	CeremonyID string
	PublicKey  any
	UserHandle []byte
	ExpiresAt  time.Time
}

type FinishBootstrapCommand struct {
	Token                string
	RateLimitKey         string
	SessionID            int64
	Email                string
	FirstName            string
	LastName             string
	PreferredDisplayName string
	PasskeyCeremonyID    string
	PasskeyResponse      []byte
}

type BootstrapResult struct {
	PersonID   string
	IdentityID string
	Session    IssuedSession
}

type BootstrapState struct {
	Closed bool
}

type PersonalProfile struct {
	ID                   string
	FirstName            string
	LastName             string
	PreferredDisplayName string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type IdentityRecord struct {
	ID        string
	PersonID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IdentityEmail struct {
	IdentityID string
	Email      string
	Normalized string
	VerifiedAt *time.Time
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type RegistrationCeremony struct {
	Challenge  []byte
	UserHandle []byte
	ExpiresAt  time.Time
	Email      domain.Email
	Name       domain.ProfileName
}

type RegistrationVerification struct {
	Challenge  []byte
	UserHandle []byte
	Response   []byte
}

type PasskeyCredential struct {
	ID              string
	IdentityID      string
	CredentialID    []byte
	PublicKey       []byte
	AttestationType string
	AAGUID          string
	SignCount       uint32
	Transports      []string
	CreatedAt       time.Time
	LastUsedAt      *time.Time
}

type IssuedSession struct {
	ID              int64
	IdentityID      string
	RawToken        string
	AuthenticatedAt time.Time
	ExpiresAt       time.Time
}

type AuditEvent struct {
	ActorID       string
	Action        string
	TargetType    string
	TargetID      string
	CorrelationID string
	Payload       map[string]any
	CreatedAt     time.Time
}

func (s *BootstrapService) StartBootstrap(ctx context.Context, command StartBootstrapCommand) (StartBootstrapResult, error) {
	if err := s.preflight(ctx, command.Token, command.RateLimitKey); err != nil {
		return StartBootstrapResult{}, err
	}
	var result StartBootstrapResult
	err := s.UnitOfWork.WithinTx(ctx, func(txCtx context.Context, repos Repositories) error {
		if err := checkBootstrapOpen(txCtx, repos); err != nil {
			return err
		}
		result.ExpiresAt = s.now().Add(s.bootstrapTTL())
		return nil
	})
	return result, err
}

func (s *BootstrapService) ClaimBootstrapProfile(ctx context.Context, command ClaimBootstrapProfileCommand) (PendingEnrollment, error) {
	if err := s.preflight(ctx, command.Token, command.RateLimitKey); err != nil {
		return PendingEnrollment{}, err
	}
	email, name, err := validateClaim(command.Email, command.FirstName, command.LastName, command.PreferredDisplayName)
	if err != nil {
		return PendingEnrollment{}, err
	}
	pending := PendingEnrollment{Token: command.Token, Email: email, Name: name, ValidUntil: s.now().Add(s.bootstrapTTL())}
	err = s.UnitOfWork.WithinTx(ctx, func(txCtx context.Context, repos Repositories) error {
		if err := checkBootstrapOpen(txCtx, repos); err != nil {
			return err
		}
		taken, err := repos.EmailTaken(txCtx, email.Normalized())
		if err != nil {
			return err
		}
		if taken {
			return domain.ErrEmailTaken
		}
		return nil
	})
	if err != nil {
		return PendingEnrollment{}, err
	}
	return pending, nil
}

func (s *BootstrapService) BeginPasskeyRegistration(ctx context.Context, command BeginPasskeyRegistrationCommand) (RegistrationOptions, error) {
	if command.Pending.ValidUntil.Before(s.now()) {
		return RegistrationOptions{}, domain.ErrInvalidToken
	}
	ceremonyID, err := s.IDs.NewID()
	if err != nil {
		return RegistrationOptions{}, fmt.Errorf("create ceremony id: %w", err)
	}
	userHandleID, err := s.IDs.NewID()
	if err != nil {
		return RegistrationOptions{}, fmt.Errorf("create user handle: %w", err)
	}
	start := RegistrationStart{
		SessionID:            command.SessionID,
		CeremonyID:           ceremonyID,
		Email:                command.Pending.Email,
		FirstName:            command.Pending.Name.FirstName,
		LastName:             command.Pending.Name.LastName,
		PreferredDisplayName: command.Pending.Name.PreferredDisplayName,
		DisplayName:          command.Pending.Name.DisplayName(),
		UserHandle:           []byte(userHandleID),
		ExpiresAt:            s.now().Add(15 * time.Minute),
	}
	options, err := s.Passkeys.BeginRegistration(ctx, start)
	if err != nil {
		return RegistrationOptions{}, fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
	}
	return options, nil
}

func (s *BootstrapService) FinishBootstrap(ctx context.Context, command FinishBootstrapCommand) (BootstrapResult, error) {
	if err := s.preflight(ctx, command.Token, command.RateLimitKey); err != nil {
		return BootstrapResult{}, err
	}

	var result BootstrapResult
	err := s.UnitOfWork.WithinTx(ctx, func(txCtx context.Context, repos Repositories) error {
		if err := checkBootstrapOpen(txCtx, repos); err != nil {
			return err
		}
		ceremony, err := repos.LockRegistrationCeremony(txCtx, command.PasskeyCeremonyID)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
		}
		submittedEmail, submittedName, err := validateClaim(
			command.Email,
			command.FirstName,
			command.LastName,
			command.PreferredDisplayName,
		)
		if err != nil || !sameRegistrationClaim(ceremony, submittedEmail, submittedName) {
			return fmt.Errorf("%w: registration claim mismatch", domain.ErrCeremonyFailed)
		}
		email, name := ceremony.Email, ceremony.Name

		taken, err := repos.EmailTaken(txCtx, email.Normalized())
		if err != nil {
			return err
		}
		if taken {
			return domain.ErrEmailTaken
		}
		now := s.now()
		if !ceremony.ExpiresAt.After(now) {
			return fmt.Errorf("%w: webauthn ceremony expired", domain.ErrCeremonyFailed)
		}
		credential, err := s.Passkeys.VerifyRegistration(txCtx, RegistrationVerification{
			Challenge:  ceremony.Challenge,
			UserHandle: ceremony.UserHandle,
			Response:   command.PasskeyResponse,
		})
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
		}
		if err := repos.ConsumeRegistrationCeremony(txCtx, command.PasskeyCeremonyID, now); err != nil {
			return fmt.Errorf("%w: %v", domain.ErrCeremonyFailed, err)
		}

		personID, identityID, credentialID, err := s.bootstrapIDs()
		if err != nil {
			return err
		}
		if err := repos.CreatePersonalProfile(txCtx, PersonalProfile{
			ID:                   personID,
			FirstName:            name.FirstName,
			LastName:             name.LastName,
			PreferredDisplayName: name.PreferredDisplayName,
			CreatedAt:            now,
			UpdatedAt:            now,
		}); err != nil {
			return err
		}
		if err := repos.CreateIdentity(txCtx, IdentityRecord{ID: identityID, PersonID: personID, CreatedAt: now, UpdatedAt: now}); err != nil {
			return err
		}
		if err := repos.AddEmail(txCtx, IdentityEmail{IdentityID: identityID, Email: email.String(), Normalized: email.Normalized(), Active: true, CreatedAt: now, UpdatedAt: now}); err != nil {
			return err
		}
		if err := repos.GrantRole(txCtx, identityID, domain.RoleAdmin); err != nil {
			return err
		}
		credential.ID = credentialID
		credential.IdentityID = identityID
		credential.CreatedAt = now
		if err := repos.StorePasskeyCredential(txCtx, credential); err != nil {
			return err
		}
		if err := repos.CloseBootstrap(txCtx, identityID, now); err != nil {
			return err
		}
		if err := repos.WriteAudit(txCtx, AuditEvent{
			ActorID:       identityID,
			Action:        "identity.bootstrap.completed",
			TargetType:    "identity",
			TargetID:      identityID,
			CorrelationID: credentialID,
			Payload:       map[string]any{"email_normalized": email.Normalized()},
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		session, err := repos.RotateForIdentity(txCtx, command.SessionID, identityID, now)
		if err != nil {
			return err
		}
		result = BootstrapResult{PersonID: personID, IdentityID: identityID, Session: session}
		return nil
	})
	return result, err
}

func sameRegistrationClaim(ceremony RegistrationCeremony, email domain.Email, name domain.ProfileName) bool {
	return ceremony.Email.Normalized() == email.Normalized() &&
		ceremony.Name.FirstName == name.FirstName &&
		ceremony.Name.LastName == name.LastName &&
		ceremony.Name.PreferredDisplayName == name.PreferredDisplayName
}

func (s *BootstrapService) preflight(ctx context.Context, token, rateLimitKey string) error {
	allowed, err := s.RateLimiter.Allow(ctx, rateLimitKey, s.rateLimitMax(), s.rateLimitWindow())
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	if err := s.Tokens.ValidateBootstrapToken(ctx, token, s.now()); err != nil {
		return err
	}
	return nil
}

func (s *BootstrapService) bootstrapIDs() (string, string, string, error) {
	personID, err := s.IDs.NewID()
	if err != nil {
		return "", "", "", fmt.Errorf("create person id: %w", err)
	}
	identityID, err := s.IDs.NewID()
	if err != nil {
		return "", "", "", fmt.Errorf("create identity id: %w", err)
	}
	credentialID, err := s.IDs.NewID()
	if err != nil {
		return "", "", "", fmt.Errorf("create passkey id: %w", err)
	}
	return personID, identityID, credentialID, nil
}

func (s *BootstrapService) now() time.Time {
	if s.Clock == nil {
		return time.Now().UTC()
	}
	return s.Clock.Now().UTC()
}

func (s *BootstrapService) bootstrapTTL() time.Duration {
	if s.BootstrapTokenExpiresIn <= 0 {
		return 72 * time.Hour
	}
	return s.BootstrapTokenExpiresIn
}

func (s *BootstrapService) rateLimitMax() int {
	if s.AuthRateLimitMax <= 0 {
		return 10
	}
	return s.AuthRateLimitMax
}

func (s *BootstrapService) rateLimitWindow() time.Duration {
	if s.AuthRateLimitWindow <= 0 {
		return 15 * time.Minute
	}
	return s.AuthRateLimitWindow
}

func checkBootstrapOpen(ctx context.Context, repos Repositories) error {
	state, err := repos.LockBootstrap(ctx)
	if err != nil {
		return err
	}
	adminExists, err := repos.AdminExists(ctx)
	if err != nil {
		return err
	}
	return domain.CanBootstrap(adminExists, state.Closed)
}

func validateClaim(emailRaw, first, last, preferred string) (domain.Email, domain.ProfileName, error) {
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		return domain.Email{}, domain.ProfileName{}, err
	}
	name, err := domain.ValidateProfile(first, last, preferred)
	if err != nil {
		return domain.Email{}, domain.ProfileName{}, err
	}
	return email, name, nil
}
