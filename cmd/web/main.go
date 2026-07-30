package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/config"
	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/outbox"
	"github.com/troop900/treelot/internal/platform/postgres"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/views"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	os.Exit(run(logger, func(server *http.Server) error {
		return server.ListenAndServe()
	}))
}

func run(logger *slog.Logger, listen func(*http.Server) error) int {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		return 1
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		return 1
	}
	defer db.Close()

	if err := migrate.EnsureCompatible(ctx, db, cfg.ExpectedSchema); err != nil {
		logger.Error("schema incompatible", "error", err)
		return 1
	}

	var appClock clock.Clock = clock.System()
	var controllable *clock.Controllable
	if cfg.TestControlEnabled {
		controllable = clock.NewControllable(time.Now().UTC())
		appClock = controllable
	}

	server, err := newHTTPServer(cfg, db, appClock, controllable, logger)
	if err != nil {
		logger.Error("initialize web server", "error", err)
		return 1
	}

	logger.Info("web server starting", "address", server.Addr, "env", cfg.AppEnv)
	if err := listen(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("web server stopped", "error", err)
		return 1
	}
	return 0
}

func newHTTPServer(cfg config.Config, db *postgres.DB, appClock clock.Clock, controllable *clock.Controllable, logger *slog.Logger) (*http.Server, error) {
	renderer, err := views.NewRenderer()
	if err != nil {
		return nil, err
	}
	store := session.NewStore(db, appClock, 24*time.Hour)
	var outboxControl handlers.OutboxControl
	if cfg.TestControlEnabled {
		outboxControl = outbox.NewStore(db, appClock)
	}
	handler := handlers.New(renderer, handlers.Options{
		Development:        cfg.AppEnv == config.EnvDevelopment,
		Logger:             logger,
		Ready:              db.Ping,
		Sessions:           store,
		SecureCookies:      cfg.SecureCookies,
		TestControlEnabled: cfg.TestControlEnabled,
		TestControlKey:     cfg.TestControlKey,
		Clock:              appClock,
		ControllableClock:  controllable,
		Outbox:             outboxControl,
	})
	return &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}
