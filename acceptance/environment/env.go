//go:build acceptance

package environment

import (
	"os"
	"time"
)

// Config describes how acceptance drivers reach the system under test.
type Config struct {
	BaseURL        string
	TestControlKey string
	DatabaseURL    string
	ReadyTimeout   time.Duration
}

// Load reads acceptance driver configuration from the environment.
func Load() Config {
	baseURL := os.Getenv("ACCEPTANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	return Config{
		BaseURL:        baseURL,
		TestControlKey: os.Getenv("ACCEPTANCE_TEST_CONTROL_KEY"),
		DatabaseURL:    os.Getenv("ACCEPTANCE_DATABASE_URL"),
		ReadyTimeout:   60 * time.Second,
	}
}
