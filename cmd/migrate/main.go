package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/troop900/treelot/internal/platform/config"
	"github.com/troop900/treelot/internal/platform/migrate"
	"github.com/troop900/treelot/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	os.Exit(run(logger, os.Args[1:]))
}

func run(logger *slog.Logger, args []string) int {
	if len(args) == 0 || args[0] != "up" {
		fmt.Fprintln(os.Stderr, "usage: migrate up")
		return 2
	}
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

	directory, err := migrate.Directory()
	if err != nil {
		logger.Error("locate migrations", "error", err)
		return 1
	}
	applied, err := migrate.Up(ctx, db, directory)
	if err != nil {
		logger.Error("migrate up", "error", err)
		return 1
	}
	version, err := migrate.CurrentVersion(ctx, db)
	if err != nil {
		logger.Error("read schema version", "error", err)
		return 1
	}
	logger.Info("migrations complete", "applied", applied, "version", version)
	return 0
}
