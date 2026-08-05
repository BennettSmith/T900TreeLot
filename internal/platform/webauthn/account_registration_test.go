package webauthn_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/testdb"
	platformwebauthn "github.com/troop900/treelot/internal/platform/webauthn"
)

func TestAccountRegistrationBeginsWithoutPersistingCeremony(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC))
	inner, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatal(err)
	}
	registration := platformwebauthn.NewAccountRegistration(inner)
	email, err := domain.NewEmail("ada@example.org")
	if err != nil {
		t.Fatal(err)
	}
	options, err := registration.BeginRegistration(context.Background(), application.AccountRegistrationStart{
		SessionID: 9, CeremonyID: "reg-1", IdentityID: "identity-1", Email: email,
		DisplayName: "Ada", UserHandle: []byte("handle"),
		ExcludeCredentials: [][]byte{[]byte("existing")},
		ExpiresAt:          clk.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	if options.CeremonyID != "reg-1" || len(options.Challenge) == 0 || options.PublicKey == nil {
		t.Fatalf("options = %#v", options)
	}
	var count int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM webauthn_ceremonies WHERE id = 'reg-1'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected no persisted ceremony, got %d", count)
	}
}

func TestAccountRegistrationVerifyRejectsInvalidPayload(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC))
	inner, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatal(err)
	}
	registration := platformwebauthn.NewAccountRegistration(inner)
	if _, err := registration.VerifyRegistration(context.Background(), application.RegistrationVerification{
		Challenge: []byte("challenge"), UserHandle: []byte("handle"), Response: []byte(`{}`),
	}); err == nil {
		t.Fatal("expected invalid verification error")
	}
}

func TestAccountRegistrationDefaultsExpiryWhenUnset(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC))
	inner, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatal(err)
	}
	registration := platformwebauthn.NewAccountRegistration(inner)
	email, err := domain.NewEmail("ada@example.org")
	if err != nil {
		t.Fatal(err)
	}
	options, err := registration.BeginRegistration(context.Background(), application.AccountRegistrationStart{
		CeremonyID: "reg-2", IdentityID: "identity-1", Email: email, DisplayName: "Ada", UserHandle: []byte("handle"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !options.ExpiresAt.Equal(clk.Now().Add(15 * time.Minute)) {
		t.Fatalf("expires = %v", options.ExpiresAt)
	}
}
