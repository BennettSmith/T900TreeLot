// Package audit writes durable security and domain audit facts.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/troop900/treelot/internal/platform/postgres"
)

type Event struct {
	ActorID       string
	Action        string
	TargetType    string
	TargetID      string
	CorrelationID string
	Payload       map[string]any
	CreatedAt     time.Time
}

type Writer struct {
	exec executor
}

func NewWriter(db *postgres.DB) *Writer {
	return &Writer{exec: db}
}

func NewTxWriter(tx pgx.Tx) *Writer {
	return &Writer{exec: tx}
}

func (w *Writer) Write(ctx context.Context, event Event) error {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode audit payload: %w", err)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	_, err = w.exec.Exec(ctx, `
		INSERT INTO audit_events (actor_id, action, target_type, target_id, correlation_id, payload, created_at)
		VALUES (NULLIF($1, ''), $2, $3, $4, $5, $6::jsonb, $7)
	`, event.ActorID, event.Action, event.TargetType, event.TargetID, event.CorrelationID, string(encoded), event.CreatedAt)
	if err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

type executor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}
