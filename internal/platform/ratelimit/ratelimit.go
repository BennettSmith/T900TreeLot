// Package ratelimit implements PostgreSQL-backed fixed-window rate limits.
package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

type Buckets struct {
	db    *postgres.DB
	clock clock.Clock
}

func NewBuckets(db *postgres.DB, clk clock.Clock) *Buckets {
	if clk == nil {
		clk = clock.System()
	}
	return &Buckets{db: db, clock: clk}
}

func (b *Buckets) Allow(ctx context.Context, key string, max int, window time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("rate limit bucket key is required")
	}
	if max <= 0 {
		return false, fmt.Errorf("rate limit max must be positive")
	}
	if window <= 0 {
		return false, fmt.Errorf("rate limit window must be positive")
	}

	now := b.clock.Now().UTC()
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin rate limit: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var started time.Time
	var count int
	err = tx.QueryRow(ctx, `
		SELECT window_started_at, count
		FROM rate_limit_buckets
		WHERE bucket_key = $1
		FOR UPDATE
	`, key).Scan(&started, &count)
	if errors.Is(err, pgx.ErrNoRows) {
		_, insertErr := tx.Exec(ctx, `
			INSERT INTO rate_limit_buckets (bucket_key, window_started_at, count, updated_at)
			VALUES ($1, $2, 1, $2)
		`, key, now)
		if insertErr != nil {
			return false, fmt.Errorf("create rate limit bucket: %w", insertErr)
		}
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("commit rate limit: %w", err)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read rate limit bucket: %w", err)
	}

	if !now.Before(started.Add(window)) {
		started = now
		count = 0
	}
	allowed := count < max
	if allowed {
		count++
	}
	_, err = tx.Exec(ctx, `
		UPDATE rate_limit_buckets
		SET window_started_at = $2, count = $3, updated_at = $4
		WHERE bucket_key = $1
	`, key, started, count, now)
	if err != nil {
		return false, fmt.Errorf("update rate limit bucket: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit rate limit: %w", err)
	}
	return allowed, nil
}
