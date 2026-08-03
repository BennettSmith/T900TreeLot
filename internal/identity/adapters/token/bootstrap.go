// Package token implements configured enrollment token validators.
package token

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
)

type BootstrapValidator struct {
	hash      [32]byte
	expiresAt time.Time
}

func NewBootstrapValidator(token string, expiresAt time.Time) *BootstrapValidator {
	return &BootstrapValidator{
		hash:      sha256.Sum256([]byte(token)),
		expiresAt: expiresAt.UTC(),
	}
}

func (v *BootstrapValidator) ValidateBootstrapToken(_ context.Context, token string, now time.Time) error {
	hash := sha256.Sum256([]byte(token))
	if now.After(v.expiresAt) || subtle.ConstantTimeCompare(hash[:], v.hash[:]) != 1 {
		return domain.ErrInvalidToken
	}
	return nil
}
