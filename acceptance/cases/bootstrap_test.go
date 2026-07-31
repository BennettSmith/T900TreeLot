//go:build acceptance

package cases_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/troop900/treelot/acceptance/dsl"
)

// Trace: UC-0@r2 US-001@r2
func TestDesignatedAdminBootstrapsExactlyOnce(t *testing.T) {
	lot := dsl.NewLot(t)
	email := fmt.Sprintf("first-admin-%d@example.org", time.Now().UTC().UnixNano())
	lot.Bootstrap().CreatesFirstAdminAndClosesBootstrap("acceptance-bootstrap-token-0001", email)
}
