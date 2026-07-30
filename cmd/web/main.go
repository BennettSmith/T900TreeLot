package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/views"
)

type config struct {
	address     string
	development bool
}

func configFromEnvironment() config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return config{
		address:     "0.0.0.0:" + port,
		development: os.Getenv("APP_ENV") == "development",
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	os.Exit(run(configFromEnvironment(), logger, func(server *http.Server) error {
		return server.ListenAndServe()
	}))
}

func run(configuration config, logger *slog.Logger, listen func(*http.Server) error) int {
	server, err := newHTTPServer(configuration, logger)
	if err != nil {
		logger.Error("initialize web server", "error", err)
		return 1
	}

	logger.Info("web server starting", "address", server.Addr)
	if err := listen(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("web server stopped", "error", err)
		return 1
	}
	return 0
}

func newHTTPServer(configuration config, logger *slog.Logger) (*http.Server, error) {
	renderer, err := views.NewRenderer()
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:              configuration.address,
		Handler:           handlers.New(renderer, handlers.Options{Development: configuration.development, Logger: logger}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}
