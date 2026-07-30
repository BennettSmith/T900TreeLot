//go:build acceptance

package cases_test

import (
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

func TestAnonymousVisitorSeesDeployableFoundation(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Presence().SeesHealthyFoundation()
}

func TestAnonymousVisitorCompletesCSRFProtectedSmokeCheck(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Presence().SubmitsSmokeCheck("Foundation smoke confirmed")
}

func TestAcceptanceClockCanBeAdvancedAndRejectsBadCredentials(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Clock().IsUnavailableWithoutKey()
	lot.Clock().CanBeAdvanced(15 * time.Minute)
}

func TestPollingHelperObservesReadinessWithoutFixedSleepOnly(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.WaitUntilReady()
}

func TestUnmigratedDatabaseIsRejectedWithoutSchemaMutation(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Platform().RejectsUnmigratedDatabaseWithoutSchemaChange()
}

func TestTestControlIsAbsentAndCookiesAreSecureOutsideAcceptance(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Platform().TestControlIsAbsentOutsideAcceptance()
	lot.Platform().IssuesSecureSessionCookiesInProduction()
}

func TestDeployedWorkerDeliversOutboxMessage(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Worker().DeliversEnqueuedOutboxMessage("acceptance-outbox-" + time.Now().UTC().Format("20060102T150405.000000000"))
}

func TestGroupsIOStubIsProtocolFaithful(t *testing.T) {
	lot := dsl.NewLot(t)
	lot.Providers().GroupsIOStubAcceptsAuthorizedTraffic()
}
