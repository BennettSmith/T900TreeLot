package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/config"
	"github.com/troop900/treelot/internal/platform/jobs"
	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	os.Exit(run(logger))
}

func run(logger *slog.Logger) int {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	worker := jobs.NewWorker(db, clock.System(), "worker")
	logger.Info("worker starting")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		claimed, err := worker.Tick(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("worker stopped")
				return 0
			}
			logger.Error("worker tick failed", "error", err)
		} else if claimed > 0 {
			logger.Info("worker claimed work", "claimed", claimed)
		}

		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return 0
		case <-ticker.C:
		}
	}
}
