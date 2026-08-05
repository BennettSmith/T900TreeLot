package domain_test

import (
	"errors"
	"testing"

	"github.com/troop900/treelot/internal/identity/domain"
)

func TestCanRemovePasskeyRequiresAnotherRemaining(t *testing.T) {
	if err := domain.CanRemovePasskey(1); !errors.Is(err, domain.ErrLastPasskey) {
		t.Fatalf("error = %v, want ErrLastPasskey", err)
	}
	if err := domain.CanRemovePasskey(2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := domain.CanRemovePasskey(0); !errors.Is(err, domain.ErrLastPasskey) {
		t.Fatalf("error = %v, want ErrLastPasskey", err)
	}
}
