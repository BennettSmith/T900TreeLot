package domain_test

import (
	"errors"
	"testing"

	"github.com/troop900/treelot/internal/identity/domain"
)

func TestNormalizeEmailTrimsAndLowercases(t *testing.T) {
	email, err := domain.NewEmail("  FIRST.Admin@Example.ORG  ")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	if email.String() != "FIRST.Admin@Example.ORG" {
		t.Errorf("String = %q", email.String())
	}
	if email.Normalized() != "first.admin@example.org" {
		t.Errorf("Normalized = %q", email.Normalized())
	}
}

func TestValidateProfileRequiresFirstAndLastName(t *testing.T) {
	name, err := domain.ValidateProfile("  Ada  ", "  Lovelace ", " Countess ")
	if err != nil {
		t.Fatalf("ValidateProfile: %v", err)
	}
	if name.FirstName != "Ada" || name.LastName != "Lovelace" || name.PreferredDisplayName != "Countess" {
		t.Fatalf("profile = %#v", name)
	}

	if _, err := domain.ValidateProfile("", "Lovelace", ""); !errors.Is(err, domain.ErrInvalidProfile) {
		t.Fatalf("empty first name error = %v, want ErrInvalidProfile", err)
	}
	if _, err := domain.ValidateProfile("Ada", "", ""); !errors.Is(err, domain.ErrInvalidProfile) {
		t.Fatalf("empty last name error = %v, want ErrInvalidProfile", err)
	}
}

func TestCanBootstrapRequiresNoAdminAndOpenBootstrap(t *testing.T) {
	if err := domain.CanBootstrap(false, false); err != nil {
		t.Fatalf("CanBootstrap open: %v", err)
	}
	if err := domain.CanBootstrap(true, false); !errors.Is(err, domain.ErrBootstrapClosed) {
		t.Fatalf("admin exists error = %v, want ErrBootstrapClosed", err)
	}
	if err := domain.CanBootstrap(false, true); !errors.Is(err, domain.ErrBootstrapClosed) {
		t.Fatalf("closed error = %v, want ErrBootstrapClosed", err)
	}
}
