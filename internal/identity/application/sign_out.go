package application

import (
	"context"
	"strconv"
	"time"
)

type SignOutUnitOfWork interface {
	WithinSignOutTx(context.Context, func(context.Context, SignOutRepositories) error) error
}

type SignOutRepositories interface {
	RevokeCurrentSession(context.Context, int64, string, time.Time) error
	WriteAudit(context.Context, AuditEvent) error
}

// SignOutService revokes the current authenticated browser session.
type SignOutService struct {
	UnitOfWork SignOutUnitOfWork
	Clock      Clock
}

type SignOutCommand struct {
	IdentityID string
	SessionID  int64
}

func (s *SignOutService) SignOut(ctx context.Context, command SignOutCommand) error {
	if command.IdentityID == "" || command.SessionID <= 0 {
		return ErrAccountNotFound
	}
	now := time.Now().UTC()
	if s.Clock != nil {
		now = s.Clock.Now().UTC()
	}
	return s.UnitOfWork.WithinSignOutTx(ctx, func(txCtx context.Context, repos SignOutRepositories) error {
		if err := repos.RevokeCurrentSession(txCtx, command.SessionID, command.IdentityID, now); err != nil {
			return err
		}
		return repos.WriteAudit(txCtx, AuditEvent{
			ActorID:       command.IdentityID,
			Action:        "identity.session.signed_out",
			TargetType:    "session",
			TargetID:      strconv.FormatInt(command.SessionID, 10),
			CorrelationID: strconv.FormatInt(command.SessionID, 10),
			Payload:       map[string]any{},
			CreatedAt:     now,
		})
	})
}
