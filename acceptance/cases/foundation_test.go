//go:build acceptance

package cases_test

import (
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

// Trace: INC-01
func TestAnonymousVisitorSeesDeployableFoundation(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Presence().SeesHealthyFoundation()
}

// Trace: INC-01
func TestAnonymousVisitorCompletesCSRFProtectedSmokeCheck(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Presence().SubmitsSmokeCheck("Foundation smoke confirmed")
}

// Trace: INC-01
func TestAcceptanceClockCanBeAdvancedAndRejectsBadCredentials(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Clock().IsUnavailableWithoutKey()
	lot.Clock().CanBeAdvanced(15 * time.Minute)
}

// Trace: INC-01
func TestPollingHelperObservesReadinessWithoutFixedSleepOnly(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.WaitUntilReady()
}

// Trace: INC-01
func TestUnmigratedDatabaseIsRejectedWithoutSchemaMutation(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Platform().RejectsUnmigratedDatabaseWithoutSchemaChange()
}

// Trace: INC-01
func TestTestControlIsAbsentAndCookiesAreSecureOutsideAcceptance(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Platform().TestControlIsAbsentOutsideAcceptance()
	lot.Platform().IssuesSecureSessionCookiesInProduction()
}

// Trace: INC-01
func TestProductionRedirectsNoncanonicalHostsWithoutBreakingHealth(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Platform().RedirectsNoncanonicalHostsToCanonicalOrigin()
	lot.Platform().ServesHealthChecksOnNoncanonicalHosts()
}

// Trace: INC-01
func TestDeployedWorkerDeliversOutboxMessage(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Worker().DeliversEnqueuedOutboxMessage("acceptance-outbox-" + time.Now().UTC().Format("20060102T150405.000000000"))
}

// Trace: INC-01
func TestGroupsIOStubIsProtocolFaithful(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Providers().GroupsIOStubAcceptsAuthorizedTraffic()
}
