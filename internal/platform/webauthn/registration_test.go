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

func TestRegistrationCeremonyBeginsAndStoresChallenge(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 15, 0, 0, time.UTC))
	ceremony, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewRegistrationCeremony: %v", err)
	}
	email, err := domain.NewEmail("first@example.org")
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}

	options, err := ceremony.BeginRegistration(context.Background(), application.RegistrationStart{
		SessionID:   0,
		CeremonyID:  "ceremony-1",
		Email:       email,
		DisplayName: "First Admin",
		UserHandle:  []byte("user-handle-1"),
		ExpiresAt:   clk.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	if options.CeremonyID != "ceremony-1" || options.PublicKey == nil {
		t.Fatalf("options = %#v", options)
	}

	var storedChallenge []byte
	if err := db.QueryRow(context.Background(), `
		SELECT challenge
		FROM webauthn_ceremonies
		WHERE id = 'ceremony-1'
	`).Scan(&storedChallenge); err != nil {
		t.Fatalf("read ceremony: %v", err)
	}
	if len(storedChallenge) < 16 {
		t.Fatalf("stored challenge too short: %d", len(storedChallenge))
	}
}

func TestRegistrationCeremonyRejectsInvalidConfigAndInvalidVerification(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 15, 0, 0, time.UTC))
	if _, err := platformwebauthn.NewRegistrationCeremony(db, clk, "not a host", []string{"http://localhost:8080"}); err == nil {
		t.Fatal("NewRegistrationCeremony succeeded with invalid RP ID")
	}
	ceremony, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewRegistrationCeremony: %v", err)
	}
	if _, err := ceremony.VerifyRegistration(context.Background(), application.RegistrationVerification{
		Challenge:  []byte("challenge-1234567890"),
		UserHandle: []byte("user-handle"),
		Response:   []byte(`{}`),
	}); err == nil {
		t.Fatal("VerifyRegistration succeeded for invalid response")
	}
}

func TestRegistrationCeremonyVerificationRejectsMalformedResponse(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 23, 15, 0, 0, time.UTC))
	ceremony, err := platformwebauthn.NewRegistrationCeremony(db, clk, "localhost", []string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewRegistrationCeremony: %v", err)
	}
	if _, err := ceremony.VerifyRegistration(context.Background(), application.RegistrationVerification{
		Challenge:  []byte("challenge-1234567890"),
		UserHandle: []byte("user-handle"),
		Response:   []byte(`not-json`),
	}); err == nil {
		t.Fatal("VerifyRegistration succeeded for malformed response")
	}
}
