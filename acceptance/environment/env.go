//go:build acceptance

package environment

import (
	"os"
	"time"
)

// Config describes how acceptance drivers reach the system under test.
type Config struct {
	BaseURL           string
	ProductionBaseURL string
	StubBaseURL       string
	TestControlKey    string
	DatabaseURL       string
	UnmigratedDBURL   string
	Image             string
	Docker            string
	ReadyTimeout      time.Duration
}

// Load reads acceptance driver configuration from the environment.
func Load() Config {
	baseURL := os.Getenv("ACCEPTANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:18080"
	}
	productionURL := os.Getenv("ACCEPTANCE_PRODUCTION_BASE_URL")
	if productionURL == "" {
		productionURL = "http://127.0.0.1:18081"
	}
	stubURL := os.Getenv("ACCEPTANCE_STUB_BASE_URL")
	if stubURL == "" {
		stubURL = "http://127.0.0.1:18090"
	}
	databaseURL := os.Getenv("ACCEPTANCE_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://treelot:treelot@127.0.0.1:5433/treelot?sslmode=disable"
	}
	unmigratedURL := os.Getenv("ACCEPTANCE_UNMIGRATED_DATABASE_URL")
	if unmigratedURL == "" {
		unmigratedURL = "postgres://treelot:treelot@127.0.0.1:5433/treelot_unmigrated?sslmode=disable"
	}
	image := os.Getenv("ACCEPTANCE_IMAGE")
	if image == "" {
		image = "treelot:local"
	}
	docker := os.Getenv("ACCEPTANCE_DOCKER")
	if docker == "" {
		docker = "docker"
	}
	return Config{
		BaseURL:           baseURL,
		ProductionBaseURL: productionURL,
		StubBaseURL:       stubURL,
		TestControlKey:    os.Getenv("ACCEPTANCE_TEST_CONTROL_KEY"),
		DatabaseURL:       databaseURL,
		UnmigratedDBURL:   unmigratedURL,
		Image:             image,
		Docker:            docker,
		ReadyTimeout:      60 * time.Second,
	}
}
