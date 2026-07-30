package clock_test

import (
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
)

func TestControllableSet(t *testing.T) {
	t.Parallel()

	controllable := clock.NewControllable(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	next := time.Date(2026, 12, 25, 9, 30, 0, 0, time.UTC)
	controllable.Set(next)
	if !controllable.Now().Equal(next) {
		t.Fatalf("Now = %v, want %v", controllable.Now(), next)
	}
}
