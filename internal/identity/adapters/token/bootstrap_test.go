package token_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/adapters/token"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
)

func TestBootstrapValidatorAcceptsConfiguredTokenUntilExpiry(t *testing.T) {
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 30, 0, 0, time.UTC))
	validator := token.NewBootstrapValidator("bootstrap-enrollment-token-0001", clk.Now().Add(time.Hour))

	if err := validator.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now()); err != nil {
		t.Fatalf("ValidateBootstrapToken: %v", err)
	}
	if err := validator.ValidateBootstrapToken(context.Background(), "wrong-token", clk.Now()); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("wrong token error = %v, want ErrInvalidToken", err)
	}
	if err := validator.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now().Add(2*time.Hour)); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expired token error = %v, want ErrInvalidToken", err)
	}
}

func TestBootstrapValidatorRemainsExpiredAfterReconstruction(t *testing.T) {
	expiresAt := time.Date(2026, 8, 1, 0, 30, 0, 0, time.UTC)
	clk := clock.NewControllable(expiresAt.Add(-time.Minute))
	validator := token.NewBootstrapValidator("bootstrap-enrollment-token-0001", expiresAt)

	if err := validator.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now()); err != nil {
		t.Fatalf("token before expiry: %v", err)
	}

	clk.Advance(2 * time.Minute)
	reconstructed := token.NewBootstrapValidator("bootstrap-enrollment-token-0001", expiresAt)
	if err := reconstructed.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now()); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("reconstructed validator error = %v, want ErrInvalidToken", err)
	}
}
