package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	identitypostgres "github.com/troop900/treelot/internal/identity/adapters/postgres"
	"github.com/troop900/treelot/internal/identity/adapters/token"
	identityapp "github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/config"
	"github.com/troop900/treelot/internal/platform/ids"
	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/outbox"
	"github.com/troop900/treelot/internal/platform/postgres"
	"github.com/troop900/treelot/internal/platform/ratelimit"
	"github.com/troop900/treelot/internal/platform/session"
	platformwebauthn "github.com/troop900/treelot/internal/platform/webauthn"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/views"
)

// shutdownTimeout bounds graceful drain so in-flight requests can finish within
// Render's termination window (~30s) without hanging the process.
const shutdownTimeout = 25 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, logger, func(server *http.Server) error {
		return server.ListenAndServe()
	}))
}

func run(ctx context.Context, logger *slog.Logger, listen func(*http.Server) error) int {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		return 1
	}

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
	return serve(ctx, logger, server, listen)
}

func serve(ctx context.Context, logger *slog.Logger, server *http.Server, listen func(*http.Server) error) int {
	errCh := make(chan error, 1)
	go func() {
		errCh <- listen(server)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("web server stopped", "error", err)
			return 1
		}
		return 0
	case <-ctx.Done():
		logger.Info("web server shutting down", "reason", ctx.Err())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("web server shutdown", "error", err)
			return 1
		}
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("web server stopped", "error", err)
			return 1
		}
		logger.Info("web server stopped")
		return 0
	}
}

func newHTTPServer(cfg config.Config, db *postgres.DB, appClock clock.Clock, controllable *clock.Controllable, logger *slog.Logger) (*http.Server, error) {
	renderer, err := views.NewRenderer()
	if err != nil {
		return nil, err
	}
	store := session.NewStore(db, appClock, 24*time.Hour)
	passkeys, err := platformwebauthn.NewRegistrationCeremony(db, appClock, cfg.WebAuthnRPID, cfg.WebAuthnOrigins)
	if err != nil {
		return nil, err
	}
	assertions, err := platformwebauthn.NewAssertionCeremony(cfg.WebAuthnRPID, cfg.WebAuthnOrigins)
	if err != nil {
		return nil, err
	}
	bootstrapService := &identityapp.BootstrapService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, appClock),
		Tokens:              token.NewBootstrapValidator(cfg.BootstrapEnrollmentToken, cfg.BootstrapTokenExpiresAt),
		RateLimiter:         ratelimit.NewBuckets(db, appClock),
		Passkeys:            passkeys,
		Clock:               appClock,
		IDs:                 ids.NewGenerator(),
		AuthRateLimitMax:    cfg.AuthRateLimitMax,
		AuthRateLimitWindow: cfg.AuthRateLimitWindow,
	}
	signInService := &identityapp.SignInService{
		UnitOfWork:          identitypostgres.NewUnitOfWork(db, appClock),
		RateLimiter:         ratelimit.NewBuckets(db, appClock),
		Passkeys:            assertions,
		Clock:               appClock,
		IDs:                 ids.NewGenerator(),
		AuthRateLimitMax:    cfg.AuthRateLimitMax,
		AuthRateLimitWindow: cfg.AuthRateLimitWindow,
	}
	signOutService := &identityapp.SignOutService{
		UnitOfWork: identitypostgres.NewUnitOfWork(db, appClock),
		Clock:      appClock,
	}
	accountSecurityService := &identityapp.AccountSecurityService{
		UnitOfWork:   identitypostgres.NewUnitOfWork(db, appClock),
		Passkeys:     assertions,
		Registration: platformwebauthn.NewAccountRegistration(passkeys),
		Clock:        appClock,
		IDs:          ids.NewGenerator(),
		StepUpTTL:    cfg.StepUpTTL,
	}
	var outboxControl handlers.OutboxControl
	var bootstrapReset handlers.BootstrapResetControl
	var identityFixture handlers.IdentityFixtureControl
	if cfg.TestControlEnabled {
		outboxControl = outbox.NewStore(db, appClock)
		bootstrapReset = identitypostgres.NewTestControl(db)
		identityFixture = &identityapp.TestFixtureService{UnitOfWork: identitypostgres.NewUnitOfWork(db, appClock)}
	}
	accountQueries := identitypostgres.NewAccountQueries(db)
	handler := handlers.New(renderer, handlers.Options{
		Development:           cfg.AppEnv == config.EnvDevelopment,
		Logger:                logger,
		Ready:                 db.Ping,
		Sessions:              store,
		SecureCookies:         cfg.SecureCookies,
		TestControlEnabled:    cfg.TestControlEnabled,
		TestControlKey:        cfg.TestControlKey,
		Clock:                 appClock,
		ControllableClock:     controllable,
		Outbox:                outboxControl,
		Bootstrap:             bootstrapService,
		SignIn:                signInService,
		SignOut:               signOutService,
		Accounts:              accountQueries,
		Landings:              accountQueries,
		AccountSecurity:       accountSecurityService,
		AccountSecurityReader: accountQueries,
		StepUpTTL:             cfg.StepUpTTL,
		BootstrapReset:        bootstrapReset,
		IdentityFixture:       identityFixture,
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
