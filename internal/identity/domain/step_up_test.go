package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/domain"
)

func TestRequireRecentStepUpRejectsMissingTimestamp(t *testing.T) {
	err := domain.RequireRecentStepUp(nil, time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC), 5*time.Minute)
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("error = %v, want ErrStepUpRequired", err)
	}
}

func TestRequireRecentStepUpRejectsStaleTimestamp(t *testing.T) {
	stepUpAt := time.Date(2026, 8, 5, 11, 54, 0, 0, time.UTC)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	err := domain.RequireRecentStepUp(&stepUpAt, now, 5*time.Minute)
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("error = %v, want ErrStepUpRequired", err)
	}
}

func TestRequireRecentStepUpAcceptsFreshTimestamp(t *testing.T) {
	stepUpAt := time.Date(2026, 8, 5, 11, 56, 0, 0, time.UTC)
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	if err := domain.RequireRecentStepUp(&stepUpAt, now, 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireRecentStepUpRejectsNonPositiveTTL(t *testing.T) {
	stepUpAt := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	now := stepUpAt
	err := domain.RequireRecentStepUp(&stepUpAt, now, 0)
	if !errors.Is(err, domain.ErrStepUpRequired) {
		t.Fatalf("error = %v, want ErrStepUpRequired", err)
	}
}
