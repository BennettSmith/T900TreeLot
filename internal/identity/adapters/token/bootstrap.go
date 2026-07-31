// Package token implements configured enrollment token validators.
package token

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
)

type BootstrapValidator struct {
	hash      [32]byte
	expiresAt time.Time
}

func NewBootstrapValidator(token string, ttl time.Duration, clk clock.Clock) *BootstrapValidator {
	if ttl <= 0 {
		ttl = 72 * time.Hour
	}
	if clk == nil {
		clk = clock.System()
	}
	return &BootstrapValidator{
		hash:      sha256.Sum256([]byte(token)),
		expiresAt: clk.Now().Add(ttl).UTC(),
	}
}

func (v *BootstrapValidator) ValidateBootstrapToken(_ context.Context, token string, now time.Time) error {
	hash := sha256.Sum256([]byte(token))
	if now.After(v.expiresAt) || subtle.ConstantTimeCompare(hash[:], v.hash[:]) != 1 {
		return domain.ErrInvalidToken
	}
	return nil
}
