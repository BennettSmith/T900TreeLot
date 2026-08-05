//go:build acceptance

package cases_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

// Trace: INC-02 UC-2B@r1 US-003@r1
func TestAccountSecurityRequiresPasskeyStepUp(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("account-security-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().RequiresStepUpBeforeCredentialChanges(acceptanceBootstrapToken, email)
}

// Trace: INC-02 UC-2B@r1 US-003@r1
func TestAccountSecurityManagesPasskeysAfterStepUp(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("passkey-manage-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().CompletesStepUpAndManagesPasskeys(acceptanceBootstrapToken, email)
}

// Trace: INC-02 UC-2B@r1 US-003@r1
func TestAccountSecurityChangesEmailAndRevokesSessions(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("email-change-%d@example.org", time.Now().UTC().UnixNano())
	newEmail := fmt.Sprintf("email-changed-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().ChangesAccountEmailRevokesSessionsAndPreservesIdentity(acceptanceBootstrapToken, email, newEmail)
}

// Trace: INC-02 UC-2B@r1 US-003@r1
func TestAccountSecurityEmailChangeRevokesOtherSessions(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("email-other-session-%d@example.org", time.Now().UTC().UnixNano())
	newEmail := fmt.Sprintf("email-other-changed-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().EmailChangeRevokesOtherActiveSessions(acceptanceBootstrapToken, email, newEmail)
}

// Trace: INC-02 UC-2B@r1 US-003@r1
func TestAccountSecurityRejectsTakenEmailWithoutEnumeration(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("email-owner-%d@example.org", time.Now().UTC().UnixNano())
	taken := fmt.Sprintf("email-taken-%d@example.org", time.Now().UTC().UnixNano())
	lot.AccountSecurity().RejectsTakenEmailWithoutRevealingOtherIdentity(acceptanceBootstrapToken, email, taken)
}
