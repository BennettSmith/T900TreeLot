package application

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
)

// AccountSecurityUnitOfWork provides one transaction for credential/email commands.
type AccountSecurityUnitOfWork interface {
	WithinAccountSecurityTx(context.Context, func(context.Context, AccountSecurityRepositories) error) error
}

// AccountSecurityRepositories are persistence ports for UC-2B self-service.
type AccountSecurityRepositories interface {
	LoadSessionStepUp(context.Context, int64, string) (*time.Time, error)
	MarkSessionStepUp(context.Context, int64, string, time.Time) error
	LoadIdentity(context.Context, string) (SignInIdentity, error)
	ListPasskeys(context.Context, string) ([]PasskeyCredential, error)
	StoreStepUpCeremony(context.Context, AssertionCeremony) error
	LockStepUpCeremony(context.Context, string) (AssertionCeremony, error)
	ConsumeStepUpCeremony(context.Context, string, time.Time) error
	UpdatePasskeyAfterAssertion(context.Context, string, uint32, uint8, time.Time) error
	StoreAccountRegistrationCeremony(context.Context, AccountRegistrationCeremony) error
	LockAccountRegistrationCeremony(context.Context, string) (AccountRegistrationCeremony, error)
	ConsumeAccountRegistrationCeremony(context.Context, string, time.Time) error
	StorePasskeyCredential(context.Context, PasskeyCredential) error
	DeletePasskeyCredential(context.Context, string, string) error
	ActiveEmail(context.Context, string) (domain.Email, error)
	EmailTaken(context.Context, string) (bool, error)
	ReplaceActiveEmail(context.Context, string, string, string, time.Time) error
	RevokeSessionsForIdentity(context.Context, string) error
	WriteAudit(context.Context, AuditEvent) error
}

// AccountRegistrationCeremony is an authenticated additional-passkey ceremony.
type AccountRegistrationCeremony struct {
	ID         string
	SessionID  int64
	IdentityID string
	Challenge  []byte
	UserHandle []byte
	ExpiresAt  time.Time
}

// AccountPasskeyRegistration begins/verifies additional passkey registration.
type AccountPasskeyRegistration interface {
	BeginRegistration(context.Context, AccountRegistrationStart) (RegistrationOptions, error)
	VerifyRegistration(context.Context, RegistrationVerification) (PasskeyCredential, error)
}

// AccountSecurityService implements UC-2B@r1 / US-003@r1 self-service.
type AccountSecurityService struct {
	UnitOfWork   AccountSecurityUnitOfWork
	Passkeys     PasskeyAssertions
	Registration AccountPasskeyRegistration
	Clock        Clock
	IDs          IDGenerator
	StepUpTTL    time.Duration
}

type ChangeEmailCommand struct {
	IdentityID string
	SessionID  int64
	NewEmail   string
}

type RemovePasskeyCommand struct {
	IdentityID   string
	SessionID    int64
	CredentialID string
}

type BeginStepUpCommand struct {
	IdentityID string
	SessionID  int64
}

type FinishStepUpCommand struct {
	IdentityID        string
	SessionID         int64
	PasskeyCeremonyID string
	PasskeyResponse   []byte
}

type BeginAddPasskeyCommand struct {
	IdentityID string
	SessionID  int64
}

type FinishAddPasskeyCommand struct {
	IdentityID        string
	SessionID         int64
	PasskeyCeremonyID string
	PasskeyResponse   []byte
}

type AccountRegistrationStart struct {
	SessionID          int64
	CeremonyID         string
	IdentityID         string
	Email              domain.Email
	DisplayName        string
	UserHandle         []byte
	ExcludeCredentials [][]byte
	ExpiresAt          time.Time
}

type SecurityStatus struct {
	StepUpRequired bool
	Passkeys       []PasskeyCredential
	PrimaryEmail   string
	DisplayName    string
	IdentityID     string
}

func (s *AccountSecurityService) stepUpTTL() time.Duration {
	if s.StepUpTTL > 0 {
		return s.StepUpTTL
	}
	return 5 * time.Minute
}

func (s *AccountSecurityService) now() time.Time {
	if s.Clock != nil {
		return s.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *AccountSecurityService) requireStepUp(ctx context.Context, repos AccountSecurityRepositories, sessionID int64, identityID string) error {
	stepUpAt, err := repos.LoadSessionStepUp(ctx, sessionID, identityID)
	if err != nil {
		return err
	}
	return domain.RequireRecentStepUp(stepUpAt, s.now(), s.stepUpTTL())
}

func (s *AccountSecurityService) BeginStepUp(ctx context.Context, command BeginStepUpCommand) (BeginSignInResult, error) {
	if command.IdentityID == "" || command.SessionID <= 0 {
		return BeginSignInResult{}, ErrAccountNotFound
	}
	ceremonyID, err := s.IDs.NewID()
	if err != nil {
		return BeginSignInResult{}, fmt.Errorf("create step-up ceremony id: %w", err)
	}
	var result BeginSignInResult
	err = s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		identity, err := repos.LoadIdentity(txCtx, command.IdentityID)
		if err != nil {
			return err
		}
		options, err := s.Passkeys.BeginAssertion(txCtx, &identity)
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		ceremony := AssertionCeremony{
			ID:         ceremonyID,
			SessionID:  command.SessionID,
			Challenge:  options.Challenge,
			IdentityID: identity.ID,
			UserHandle: identity.UserHandle,
			ExpiresAt:  s.now().Add(5 * time.Minute),
		}
		if err := repos.StoreStepUpCeremony(txCtx, ceremony); err != nil {
			return err
		}
		result = BeginSignInResult{CeremonyID: ceremonyID, PublicKey: options.PublicKey}
		return nil
	})
	return result, err
}

func (s *AccountSecurityService) FinishStepUp(ctx context.Context, command FinishStepUpCommand) error {
	if command.IdentityID == "" || command.SessionID <= 0 || command.PasskeyCeremonyID == "" {
		return domain.ErrCeremonyFailed
	}
	return s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		ceremony, err := repos.LockStepUpCeremony(txCtx, command.PasskeyCeremonyID)
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		if ceremony.SessionID != command.SessionID || ceremony.IdentityID != command.IdentityID || s.now().After(ceremony.ExpiresAt) {
			return domain.ErrCeremonyFailed
		}
		identity, err := repos.LoadIdentity(txCtx, command.IdentityID)
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		verified, err := s.Passkeys.VerifyAssertion(txCtx, AssertionVerification{
			Challenge:    ceremony.Challenge,
			Identity:     identity,
			Discoverable: false,
			Response:     command.PasskeyResponse,
		})
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		credential, ok := findCredential(identity.Credentials, verified.CredentialID)
		if !ok {
			return domain.ErrCeremonyFailed
		}
		if err := repos.UpdatePasskeyAfterAssertion(txCtx, credential.ID, verified.SignCount, verified.AuthenticatorFlags, s.now()); err != nil {
			return err
		}
		if err := repos.ConsumeStepUpCeremony(txCtx, ceremony.ID, s.now()); err != nil {
			return err
		}
		if err := repos.MarkSessionStepUp(txCtx, command.SessionID, command.IdentityID, s.now()); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, AuditEvent{
			ActorID:    command.IdentityID,
			Action:     "identity.step_up.completed",
			TargetType: "session",
			TargetID:   fmt.Sprintf("%d", command.SessionID),
			Payload:    map[string]any{},
			CreatedAt:  s.now(),
		})
	})
}

func (s *AccountSecurityService) BeginAddPasskey(ctx context.Context, command BeginAddPasskeyCommand) (RegistrationOptions, error) {
	if command.IdentityID == "" || command.SessionID <= 0 {
		return RegistrationOptions{}, ErrAccountNotFound
	}
	ceremonyID, err := s.IDs.NewID()
	if err != nil {
		return RegistrationOptions{}, fmt.Errorf("create registration ceremony id: %w", err)
	}
	var result RegistrationOptions
	err = s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		if err := s.requireStepUp(txCtx, repos, command.SessionID, command.IdentityID); err != nil {
			return err
		}
		identity, err := repos.LoadIdentity(txCtx, command.IdentityID)
		if err != nil {
			return err
		}
		email, err := repos.ActiveEmail(txCtx, command.IdentityID)
		if err != nil {
			return err
		}
		exclude := make([][]byte, 0, len(identity.Credentials))
		for _, credential := range identity.Credentials {
			exclude = append(exclude, credential.CredentialID)
		}
		options, err := s.Registration.BeginRegistration(txCtx, AccountRegistrationStart{
			SessionID:          command.SessionID,
			CeremonyID:         ceremonyID,
			IdentityID:         command.IdentityID,
			Email:              email,
			DisplayName:        email.String(),
			UserHandle:         identity.UserHandle,
			ExcludeCredentials: exclude,
			ExpiresAt:          s.now().Add(15 * time.Minute),
		})
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		if err := repos.StoreAccountRegistrationCeremony(txCtx, AccountRegistrationCeremony{
			ID:         ceremonyID,
			SessionID:  command.SessionID,
			IdentityID: command.IdentityID,
			Challenge:  options.Challenge,
			UserHandle: identity.UserHandle,
			ExpiresAt:  options.ExpiresAt,
		}); err != nil {
			return err
		}
		result = options
		return nil
	})
	return result, err
}

func (s *AccountSecurityService) FinishAddPasskey(ctx context.Context, command FinishAddPasskeyCommand) error {
	if command.IdentityID == "" || command.SessionID <= 0 || command.PasskeyCeremonyID == "" {
		return domain.ErrCeremonyFailed
	}
	credentialID, err := s.IDs.NewID()
	if err != nil {
		return fmt.Errorf("create passkey id: %w", err)
	}
	return s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		if err := s.requireStepUp(txCtx, repos, command.SessionID, command.IdentityID); err != nil {
			return err
		}
		ceremony, err := repos.LockAccountRegistrationCeremony(txCtx, command.PasskeyCeremonyID)
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		if ceremony.SessionID != command.SessionID || s.now().After(ceremony.ExpiresAt) {
			return domain.ErrCeremonyFailed
		}
		credential, err := s.Registration.VerifyRegistration(txCtx, RegistrationVerification{
			Challenge:  ceremony.Challenge,
			UserHandle: ceremony.UserHandle,
			Response:   command.PasskeyResponse,
		})
		if err != nil {
			return domain.ErrCeremonyFailed
		}
		credential.ID = credentialID
		credential.IdentityID = command.IdentityID
		credential.CreatedAt = s.now()
		if err := repos.StorePasskeyCredential(txCtx, credential); err != nil {
			return err
		}
		if err := repos.ConsumeAccountRegistrationCeremony(txCtx, ceremony.ID, s.now()); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, AuditEvent{
			ActorID:    command.IdentityID,
			Action:     "identity.passkey.added",
			TargetType: "passkey_credential",
			TargetID:   credential.ID,
			Payload:    map[string]any{},
			CreatedAt:  s.now(),
		})
	})
}

func (s *AccountSecurityService) RemovePasskey(ctx context.Context, command RemovePasskeyCommand) error {
	if command.IdentityID == "" || command.SessionID <= 0 || command.CredentialID == "" {
		return ErrAccountNotFound
	}
	return s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		if err := s.requireStepUp(txCtx, repos, command.SessionID, command.IdentityID); err != nil {
			return err
		}
		passkeys, err := repos.ListPasskeys(txCtx, command.IdentityID)
		if err != nil {
			return err
		}
		if err := domain.CanRemovePasskey(len(passkeys)); err != nil {
			return err
		}
		found := false
		for _, passkey := range passkeys {
			if passkey.ID == command.CredentialID {
				found = true
				break
			}
		}
		if !found {
			return ErrAccountNotFound
		}
		if err := repos.DeletePasskeyCredential(txCtx, command.IdentityID, command.CredentialID); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, AuditEvent{
			ActorID:    command.IdentityID,
			Action:     "identity.passkey.removed",
			TargetType: "passkey_credential",
			TargetID:   command.CredentialID,
			Payload:    map[string]any{},
			CreatedAt:  s.now(),
		})
	})
}

func (s *AccountSecurityService) ChangeEmail(ctx context.Context, command ChangeEmailCommand) error {
	if command.IdentityID == "" || command.SessionID <= 0 {
		return ErrAccountNotFound
	}
	email, err := domain.NewEmail(command.NewEmail)
	if err != nil {
		return err
	}
	return s.UnitOfWork.WithinAccountSecurityTx(ctx, func(txCtx context.Context, repos AccountSecurityRepositories) error {
		if err := s.requireStepUp(txCtx, repos, command.SessionID, command.IdentityID); err != nil {
			return err
		}
		current, err := repos.ActiveEmail(txCtx, command.IdentityID)
		if err != nil {
			return err
		}
		if current.Normalized() == email.Normalized() {
			return nil
		}
		taken, err := repos.EmailTaken(txCtx, email.Normalized())
		if err != nil {
			return err
		}
		if taken {
			return domain.ErrEmailTaken
		}
		if err := repos.ReplaceActiveEmail(txCtx, command.IdentityID, email.String(), email.Normalized(), s.now()); err != nil {
			return err
		}
		if err := repos.RevokeSessionsForIdentity(txCtx, command.IdentityID); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, AuditEvent{
			ActorID:    command.IdentityID,
			Action:     "identity.account_email.changed",
			TargetType: "identity",
			TargetID:   command.IdentityID,
			Payload:    map[string]any{},
			CreatedAt:  s.now(),
		})
	})
}

func findCredential(credentials []PasskeyCredential, credentialID []byte) (PasskeyCredential, bool) {
	for _, credential := range credentials {
		if bytes.Equal(credential.CredentialID, credentialID) {
			return credential, true
		}
	}
	return PasskeyCredential{}, false
}
