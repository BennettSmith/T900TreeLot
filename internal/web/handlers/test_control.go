package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/troop900/treelot/internal/platform/outbox"
)

// OutboxControl is the acceptance-only port for enqueueing and inspecting outbox rows.
type OutboxControl interface {
	Enqueue(ctx context.Context, idempotencyKey, channel string) error
	Status(ctx context.Context, idempotencyKey string) (outbox.Message, error)
}

func (s *Server) enqueueOutbox(response http.ResponseWriter, request *http.Request) {
	if !s.authorizeTestControl(request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if s.outbox == nil {
		http.Error(response, "outbox control unavailable", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotency_key"`
		Channel        string `json:"channel"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		http.Error(response, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Channel == "" {
		body.Channel = "groupsio"
	}
	if err := s.outbox.Enqueue(request.Context(), body.IdempotencyKey, body.Channel); err != nil {
		http.Error(response, "unable to enqueue", http.StatusBadRequest)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(response).Encode(map[string]string{
		"idempotency_key": body.IdempotencyKey,
		"channel":         body.Channel,
		"status":          "pending",
	})
}

func (s *Server) getOutbox(response http.ResponseWriter, request *http.Request) {
	if !s.authorizeTestControl(request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if s.outbox == nil {
		http.Error(response, "outbox control unavailable", http.StatusServiceUnavailable)
		return
	}
	key := request.URL.Query().Get("idempotency_key")
	message, err := s.outbox.Status(request.Context(), key)
	if errors.Is(err, outbox.ErrNotFound) {
		http.Error(response, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(response, "unable to read outbox", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{
		"idempotency_key": message.IdempotencyKey,
		"channel":         message.Channel,
		"status":          message.Status,
	})
}
