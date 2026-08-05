package domain

import (
	"errors"
	"time"
)

// ErrStepUpRequired indicates a security-sensitive action needs a recent passkey assertion.
var ErrStepUpRequired = errors.New("passkey step-up required")

// RequireRecentStepUp enforces that stepUpAt is present and within ttl of now.
func RequireRecentStepUp(stepUpAt *time.Time, now time.Time, ttl time.Duration) error {
	if ttl <= 0 || stepUpAt == nil {
		return ErrStepUpRequired
	}
	if now.Before(stepUpAt.UTC()) {
		return ErrStepUpRequired
	}
	if now.Sub(stepUpAt.UTC()) > ttl {
		return ErrStepUpRequired
	}
	return nil
}
