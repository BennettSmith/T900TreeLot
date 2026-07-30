package clock_test

import (
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
)

func TestSystemClockReturnsUTC(t *testing.T) {
	t.Parallel()

	now := clock.System().Now()
	if now.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", now.Location())
	}
}

func TestControllableClockAdvances(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 7, 30, 15, 0, 0, 0, time.UTC)
	controllable := clock.NewControllable(start)
	if !controllable.Now().Equal(start) {
		t.Fatalf("Now = %v, want %v", controllable.Now(), start)
	}

	controllable.Advance(2 * time.Hour)
	want := start.Add(2 * time.Hour)
	if !controllable.Now().Equal(want) {
		t.Fatalf("Now after advance = %v, want %v", controllable.Now(), want)
	}
}
