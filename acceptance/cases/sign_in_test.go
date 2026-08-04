//go:build acceptance

package cases_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

// Trace: UC-2@r2 US-002@r2
func TestFamilyManagerSignsInWithDiscoverablePasskey(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("family-manager-%d@example.org", time.Now().UTC().UnixNano())
	lot.SignIn().FamilyManagerUsesDiscoverablePasskey(acceptanceBootstrapToken, email)
}

// Trace: UC-2@r2 US-002@r2
func TestYoungAdultScoutSignsInToPersonalSchedule(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("young-adult-scout-%d@example.org", time.Now().UTC().UnixNano())
	lot.SignIn().YoungAdultScoutUsesPersonalSchedule(acceptanceBootstrapToken, email)
}

// Trace: UC-2@r2 US-002@r2
func TestEmailHintDoesNotRevealWhetherAccountExists(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("hinted-manager-%d@example.org", time.Now().UTC().UnixNano())
	lot.SignIn().EmailHintDoesNotRevealAccount(acceptanceBootstrapToken, email, "missing-person@example.org")
}

// Trace: UC-2@r2 US-002@r2
func TestSignInRateLimitsRepeatedHints(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.SignIn().RateLimitsEmailHints("missing-person@example.org", 20)
}

// Trace: UC-2@r2 US-002@r2
func TestRevokedSessionNoLongerGrantsAccess(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("revoked-manager-%d@example.org", time.Now().UTC().UnixNano())
	lot.SignIn().RevokedSessionIsRejected(acceptanceBootstrapToken, email)
}

// Trace: UC-2@r2 US-002@r2
func TestFailedAndReplayedPasskeyAssertionsCreateNoExtraSession(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("failed-sign-in-%d@example.org", time.Now().UTC().UnixNano())
	lot.SignIn().RejectsFailedAndReplayedAssertion(acceptanceBootstrapToken, email)
}
