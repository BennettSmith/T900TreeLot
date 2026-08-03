//go:build acceptance

package cases_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

const acceptanceBootstrapToken = "acceptance-bootstrap-token-0001"

// Trace: UC-0@r2 US-001@r2
func TestDesignatedAdminBootstrapsExactlyOnce(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("first-admin-%d@example.org", time.Now().UTC().UnixNano())
	lot.Bootstrap().CreatesFirstAdminAndClosesBootstrap(acceptanceBootstrapToken, email)
}

// Trace: UC-0@r2 US-001@r2
func TestBootstrapRejectsInvalidTokenWithoutRevealingAccounts(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Bootstrap().RejectsInvalidTokenWithoutRevealingAccounts("not-the-bootstrap-token-0001", "secret-admin@example.org")
}

// Trace: UC-0@r2 US-001@r2
func TestBootstrapRateLimitsRepeatedInvalidAttempts(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Bootstrap().RateLimitsRepeatedInvalidBootstrapAttempts("not-the-bootstrap-token-0001", 20)
}

// Trace: UC-0@r2 US-001@r2
func TestFailedPasskeyCeremonyCreatesNoAdministrator(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("failed-ceremony-%d@example.org", time.Now().UTC().UnixNano())
	lot.Bootstrap().FailedPasskeyCeremonyLeavesBootstrapOpen(acceptanceBootstrapToken, email)
}

// Trace: UC-0@r2 US-001@r2
func TestBootstrapRejectsProfileChangesAfterPasskeyRegistrationBegins(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("bound-ceremony-%d@example.org", time.Now().UTC().UnixNano())
	lot.Bootstrap().RejectsChangedProfileAfterPasskeyRegistrationBegins(acceptanceBootstrapToken, email)
}

// Trace: UC-0@r2 US-001@r2
func TestOnlyOneConcurrentBootstrapAttemptSucceeds(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Bootstrap().OnlyOneConcurrentBootstrapSucceeds(acceptanceBootstrapToken)
}
