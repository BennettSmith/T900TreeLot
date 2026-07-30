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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, logger, 0))
}

// run starts the worker. If maxTicks > 0, it processes at most that many ticks (for tests).
func run(ctx context.Context, logger *slog.Logger, maxTicks int) int {
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

	worker := jobs.NewWorker(db, clock.System(), "worker")
	logger.Info("worker starting")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ticks := 0
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
		ticks++
		if maxTicks > 0 && ticks >= maxTicks {
			logger.Info("worker stopped after max ticks", "ticks", ticks)
			return 0
		}

		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return 0
		case <-ticker.C:
		}
	}
}
