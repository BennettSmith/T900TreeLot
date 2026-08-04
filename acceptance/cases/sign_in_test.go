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
