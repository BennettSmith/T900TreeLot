//go:build acceptance

package cases_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

// Trace: UC-2B@r1 US-003@r1
func TestAccountSecurityRequiresPasskeyStepUp(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("account-security-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().RequiresStepUpBeforeCredentialChanges(acceptanceBootstrapToken, email)
}
