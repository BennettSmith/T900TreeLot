package main

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConfigFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("APP_ENV", "development")

	config := configFromEnvironment()

	if config.address != "0.0.0.0:9090" {
		t.Errorf("address = %q, want %q", config.address, "0.0.0.0:9090")
	}
	if !config.development {
		t.Error("development = false, want true")
	}
}

func TestRunReportsListenerOutcome(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	for _, test := range []struct {
		name       string
		listen     func(*http.Server) error
		wantStatus int
	}{
		{name: "clean stop", listen: func(*http.Server) error { return nil }, wantStatus: 0},
		{name: "server closed", listen: func(*http.Server) error { return http.ErrServerClosed }, wantStatus: 0},
		{name: "listener failure", listen: func(*http.Server) error { return errors.New("bind failed") }, wantStatus: 1},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if status := run(config{address: "127.0.0.1:0"}, logger, test.listen); status != test.wantStatus {
				t.Errorf("run status = %d, want %d", status, test.wantStatus)
			}
		})
	}
}

func TestConfigUsesSafeDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")

	config := configFromEnvironment()

	if config.address != "0.0.0.0:8080" {
		t.Errorf("address = %q, want default", config.address)
	}
	if config.development {
		t.Error("development = true outside development environment")
	}
}

func TestNewHTTPServerBuildsProductionAdapter(t *testing.T) {
	t.Parallel()

	server, err := newHTTPServer(config{address: "0.0.0.0:7070"}, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("newHTTPServer: %v", err)
	}

	if server.Addr != "0.0.0.0:7070" {
		t.Errorf("Addr = %q, want configured address", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.IdleTimeout != 60*time.Second {
		t.Error("server timeouts do not match hardened defaults")
	}
	request := httptest.NewRequest(http.MethodGet, "/_dev/components", nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("production gallery status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
