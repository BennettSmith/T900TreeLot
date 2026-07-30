// Package jobs implements PostgreSQL-backed outbox and background job claiming.
package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

// Worker claims pending outbox rows and jobs. Provider delivery is a later increment;
// claimed outbox rows are marked delivered after a no-op handler so the queue stays healthy.
type Worker struct {
	db    *postgres.DB
	clock clock.Clock
	owner string
	lease time.Duration
	batch int
}

// NewWorker constructs a worker identity.
func NewWorker(db *postgres.DB, clk clock.Clock, owner string) *Worker {
	return &Worker{
		db:    db,
		clock: clk,
		owner: owner,
		lease: 30 * time.Second,
		batch: 10,
	}
}

// Tick claims and processes one batch of outbox messages and jobs.
func (w *Worker) Tick(ctx context.Context) (int, error) {
	outboxClaimed, err := w.claimOutbox(ctx)
	if err != nil {
		return 0, err
	}
	jobsClaimed, err := w.claimJobs(ctx)
	if err != nil {
		return outboxClaimed, err
	}
	return outboxClaimed + jobsClaimed, nil
}

func (w *Worker) claimOutbox(ctx context.Context) (int, error) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin outbox claim: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id
		FROM outbox_messages
		WHERE status = 'pending' AND available_at <= $1
		ORDER BY id
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`, w.clock.Now(), w.batch)
	if err != nil {
		return 0, fmt.Errorf("select outbox: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan outbox: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, id := range ids {
		if _, err := tx.Exec(ctx, `
			UPDATE outbox_messages
			SET status = 'delivered', attempts = attempts + 1, updated_at = $2
			WHERE id = $1
		`, id, w.clock.Now()); err != nil {
			return 0, fmt.Errorf("complete outbox %d: %w", id, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit outbox: %w", err)
	}
	return len(ids), nil
}

func (w *Worker) claimJobs(ctx context.Context) (int, error) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin job claim: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id
		FROM background_jobs
		WHERE status = 'pending' AND available_at <= $1
		ORDER BY id
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`, w.clock.Now(), w.batch)
	if err != nil {
		return 0, fmt.Errorf("select jobs: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan job: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	leaseExpiry := w.clock.Now().Add(w.lease)
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `
			UPDATE background_jobs
			SET status = 'completed',
			    attempts = attempts + 1,
			    lease_owner = $2,
			    lease_expires_at = $3,
			    updated_at = $4
			WHERE id = $1
		`, id, w.owner, leaseExpiry, w.clock.Now()); err != nil {
			return 0, fmt.Errorf("complete job %d: %w", id, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit jobs: %w", err)
	}
	return len(ids), nil
}
