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
	validator := token.NewBootstrapValidator("bootstrap-enrollment-token-0001", time.Hour, clk)

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

func TestBootstrapValidatorDefaultsTTL(t *testing.T) {
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 30, 0, 0, time.UTC))
	validator := token.NewBootstrapValidator("bootstrap-enrollment-token-0001", 0, clk)
	if err := validator.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now().Add(71*time.Hour)); err != nil {
		t.Fatalf("token before default expiry: %v", err)
	}
	if err := validator.ValidateBootstrapToken(context.Background(), "bootstrap-enrollment-token-0001", clk.Now().Add(73*time.Hour)); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("token after default expiry error = %v, want ErrInvalidToken", err)
	}
}
