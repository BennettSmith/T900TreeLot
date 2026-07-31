package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/ratelimit"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestBucketsLimitAndResetByInjectedClock(t *testing.T) {
	db := testdb.OpenMigrated(t)
	clk := clock.NewControllable(time.Date(2026, 7, 31, 22, 0, 0, 0, time.UTC))
	buckets := ratelimit.NewBuckets(db, clk)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		allowed, err := buckets.Allow(ctx, "bootstrap:ip:127.0.0.1", 2, time.Minute)
		if err != nil {
			t.Fatalf("Allow #%d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("Allow #%d = false, want true", i+1)
		}
	}
	allowed, err := buckets.Allow(ctx, "bootstrap:ip:127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatalf("Allow #3: %v", err)
	}
	if allowed {
		t.Fatal("Allow #3 = true, want false")
	}

	clk.Advance(time.Minute)
	allowed, err = buckets.Allow(ctx, "bootstrap:ip:127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatalf("Allow after reset: %v", err)
	}
	if !allowed {
		t.Fatal("Allow after reset = false, want true")
	}
}

func TestBucketsValidateInputs(t *testing.T) {
	db := testdb.OpenMigrated(t)
	buckets := ratelimit.NewBuckets(db, nil)
	ctx := context.Background()

	if _, err := buckets.Allow(ctx, "", 1, time.Minute); err == nil {
		t.Fatal("Allow succeeded with empty key")
	}
	if _, err := buckets.Allow(ctx, "key", 0, time.Minute); err == nil {
		t.Fatal("Allow succeeded with non-positive max")
	}
	if _, err := buckets.Allow(ctx, "key", 1, 0); err == nil {
		t.Fatal("Allow succeeded with non-positive window")
	}
}
