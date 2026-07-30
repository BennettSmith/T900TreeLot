// Command provider-stubs serves controllable substitutes for external providers.
package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	stub := newGroupsIOStub()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = response.Write([]byte("ok\n"))
	})
	mux.HandleFunc("POST /api/v1/groups/{group}/messages", stub.postMessage)
	mux.HandleFunc("GET /_stub/groupsio/messages", stub.listMessages)
	mux.HandleFunc("POST /_stub/groupsio/reset", stub.reset)

	server := &http.Server{
		Addr:              "0.0.0.0:" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("provider stubs starting", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("provider stubs stopped", "error", err)
		os.Exit(1)
	}
}

type groupsIOStub struct {
	mu       sync.Mutex
	messages []storedMessage
}

type storedMessage struct {
	Group   string            `json:"group"`
	Body    json.RawMessage   `json:"body"`
	Headers map[string]string `json:"headers"`
}

func newGroupsIOStub() *groupsIOStub {
	return &groupsIOStub{}
}

func (s *groupsIOStub) postMessage(response http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Authorization") == "" {
		http.Error(response, "missing authorization", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, 1<<20))
	if err != nil {
		http.Error(response, "invalid body", http.StatusBadRequest)
		return
	}
	group := request.PathValue("group")
	s.mu.Lock()
	s.messages = append(s.messages, storedMessage{
		Group: group,
		Body:  json.RawMessage(append([]byte(nil), body...)),
		Headers: map[string]string{
			"Authorization": request.Header.Get("Authorization"),
			"Content-Type":  request.Header.Get("Content-Type"),
		},
	})
	s.mu.Unlock()

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(response).Encode(map[string]any{
		"id":      len(s.messages),
		"group":   group,
		"status":  "accepted",
		"stubbed": true,
	})
}

func (s *groupsIOStub) listMessages(response http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(s.messages)
}

func (s *groupsIOStub) reset(response http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	s.messages = nil
	s.mu.Unlock()
	response.WriteHeader(http.StatusNoContent)
}
