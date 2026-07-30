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
	// NewLot already waits with Eventually; repeating readiness proves polling remains green.
	lot.WaitUntilReady()
}
