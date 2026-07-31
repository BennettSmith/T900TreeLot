package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/config"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://treelot:treelot@localhost:5432/treelot_test?sslmode=disable"
	}
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", url)
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("GROUPS_IO_ENABLED", "false")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	return cfg
}

func TestNewHTTPServerBuildsProductionAdapter(t *testing.T) {
	db := testdb.OpenMigrated(t)
	cfg := testConfig(t)
	cfg.AppEnv = config.EnvProduction
	cfg.SecureCookies = true

	server, err := newHTTPServer(cfg, db, clock.System(), nil, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("newHTTPServer: %v", err)
	}
	if server.Addr != "0.0.0.0:8080" {
		t.Errorf("Addr = %q", server.Addr)
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

func TestRunReportsListenerOutcome(t *testing.T) {
	_ = testdb.OpenMigrated(t)
	_ = testConfig(t)

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
		t.Run(test.name, func(t *testing.T) {
			if status := run(logger, test.listen); status != test.wantStatus {
				t.Errorf("run status = %d, want %d", status, test.wantStatus)
			}
		})
	}
}

func TestRunRejectsIncompatibleSchema(t *testing.T) {
	db := testdb.OpenMigrated(t)
	_ = testConfig(t)
	_, _ = db.Exec(context.Background(), `DELETE FROM schema_migrations`)
	_, _ = db.Exec(context.Background(), `INSERT INTO schema_migrations (version) VALUES (99)`)

	status := run(slog.New(slog.DiscardHandler), func(*http.Server) error { return nil })
	if status != 1 {
		t.Fatalf("status = %d, want 1", status)
	}
}
