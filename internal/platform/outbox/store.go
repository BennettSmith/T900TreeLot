// Package outbox persists transactional outbox messages.
package outbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

// ErrNotFound indicates no outbox row exists for the idempotency key.
var ErrNotFound = errors.New("outbox message not found")

// Message is a durable outbound work item.
type Message struct {
	IdempotencyKey string
	Channel        string
	Status         string
}

// Store reads and writes outbox rows.
type Store struct {
	db *postgres.DB
}

// NewStore constructs an outbox store.
func NewStore(db *postgres.DB, _ clock.Clock) *Store {
	return &Store{db: db}
}

// Enqueue inserts a pending outbox message. Duplicate keys are ignored.
func (s *Store) Enqueue(ctx context.Context, idempotencyKey, channel string) error {
	if idempotencyKey == "" || channel == "" {
		return fmt.Errorf("idempotency key and channel are required")
	}
	// available_at uses database now() so worker claim comparisons against the
	// system clock are not blocked when the acceptance controllable clock differs.
	_, err := s.db.Exec(ctx, `
		INSERT INTO outbox_messages (idempotency_key, channel, payload, available_at)
		VALUES ($1, $2, '{}'::jsonb, now())
		ON CONFLICT (idempotency_key) DO NOTHING
	`, idempotencyKey, channel)
	if err != nil {
		return fmt.Errorf("enqueue outbox: %w", err)
	}
	return nil
}

// Status returns the current status for an idempotency key.
func (s *Store) Status(ctx context.Context, idempotencyKey string) (Message, error) {
	var message Message
	err := s.db.QueryRow(ctx, `
		SELECT idempotency_key, channel, status
		FROM outbox_messages
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&message.IdempotencyKey, &message.Channel, &message.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, fmt.Errorf("read outbox status: %w", err)
	}
	return message, nil
}
