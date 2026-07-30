package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/outbox"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/handlers"
)

type memoryOutbox struct {
	messages map[string]outbox.Message
}

func (m *memoryOutbox) Enqueue(_ context.Context, idempotencyKey, channel string) error {
	if m.messages == nil {
		m.messages = map[string]outbox.Message{}
	}
	m.messages[idempotencyKey] = outbox.Message{
		IdempotencyKey: idempotencyKey,
		Channel:        channel,
		Status:         "pending",
	}
	return nil
}

func (m *memoryOutbox) Status(_ context.Context, idempotencyKey string) (outbox.Message, error) {
	message, ok := m.messages[idempotencyKey]
	if !ok {
		return outbox.Message{}, outbox.ErrNotFound
	}
	return message, nil
}

func TestOutboxTestControlUnavailableWithoutStore(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		Sessions:           session.NewMemoryStore(clock.System(), time.Hour),
	})
	response := request(t, server, http.MethodPost, "/_test/outbox", `{"idempotency_key":"k"}`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
}

func TestOutboxTestControlRequiresKeyAndEnqueues(t *testing.T) {
	t.Parallel()

	control := &memoryOutbox{}
	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		Sessions:           session.NewMemoryStore(clock.System(), time.Hour),
		Outbox:             control,
	})

	unauthorized := request(t, server, http.MethodPost, "/_test/outbox", `{"idempotency_key":"k1"}`, map[string]string{
		"Content-Type": "application/json",
	}, nil)
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	created := request(t, server, http.MethodPost, "/_test/outbox", `{"idempotency_key":"k1","channel":"groupsio"}`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", created.Code, created.Body.String())
	}

	statusResponse := request(t, server, http.MethodGet, "/_test/outbox?idempotency_key=k1", "", map[string]string{
		"X-Test-Control-Key": "secret",
	}, nil)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("status code = %d", statusResponse.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(statusResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["status"] != "pending" {
		t.Fatalf("payload = %#v", payload)
	}

	missing := request(t, server, http.MethodGet, "/_test/outbox?idempotency_key=missing", "", map[string]string{
		"X-Test-Control-Key": "secret",
	}, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", missing.Code)
	}

	badBody := request(t, server, http.MethodPost, "/_test/outbox", `{`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if badBody.Code != http.StatusBadRequest {
		t.Fatalf("bad body status = %d", badBody.Code)
	}

	defaultChannel := request(t, server, http.MethodPost, "/_test/outbox", `{"idempotency_key":"k2"}`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if defaultChannel.Code != http.StatusCreated {
		t.Fatalf("default channel status = %d", defaultChannel.Code)
	}
}
