// Package config loads and validates process configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment names supported by the application.
const (
	EnvDevelopment = "development"
	EnvAcceptance  = "acceptance"
	EnvProduction  = "production"
)

// Config holds validated runtime settings.
type Config struct {
	AppEnv             string
	ListenAddress      string
	DatabaseURL        string
	TimeZone           *time.Location
	PublicBaseURL      *url.URL
	SessionKey         []byte
	GroupsIOEnabled    bool
	SecureCookies      bool
	TestControlEnabled bool
	TestControlKey     string
	ExpectedSchema     int
}

// Load reads configuration from the process environment and validates it.
func Load() (Config, error) {
	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	if appEnv == "" {
		appEnv = EnvProduction
	}
	switch appEnv {
	case EnvDevelopment, EnvAcceptance, EnvProduction:
	default:
		return Config{}, fmt.Errorf("APP_ENV must be development, acceptance, or production")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("PORT must be numeric: %w", err)
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	zoneName := strings.TrimSpace(os.Getenv("TREE_LOT_TIME_ZONE"))
	if zoneName == "" {
		return Config{}, fmt.Errorf("TREE_LOT_TIME_ZONE is required")
	}
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		return Config{}, fmt.Errorf("TREE_LOT_TIME_ZONE is invalid: %w", err)
	}

	baseURLRaw := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	if baseURLRaw == "" {
		return Config{}, fmt.Errorf("PUBLIC_BASE_URL is required")
	}
	baseURL, err := url.Parse(baseURLRaw)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return Config{}, fmt.Errorf("PUBLIC_BASE_URL must be an absolute URL")
	}
	if appEnv != EnvDevelopment && baseURL.Scheme != "https" {
		return Config{}, fmt.Errorf("PUBLIC_BASE_URL must use https outside development")
	}

	sessionKey := strings.TrimSpace(os.Getenv("SESSION_KEY"))
	if len(sessionKey) < 32 {
		return Config{}, fmt.Errorf("SESSION_KEY must be at least 32 characters")
	}

	groupsIOEnabled := false
	if raw := strings.TrimSpace(os.Getenv("GROUPS_IO_ENABLED")); raw != "" {
		groupsIOEnabled, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("GROUPS_IO_ENABLED must be a boolean: %w", err)
		}
	}

	testControlKey := strings.TrimSpace(os.Getenv("TEST_CONTROL_KEY"))
	testControlEnabled := appEnv == EnvAcceptance
	if testControlEnabled && testControlKey == "" {
		return Config{}, fmt.Errorf("TEST_CONTROL_KEY is required in acceptance")
	}

	return Config{
		AppEnv:             appEnv,
		ListenAddress:      "0.0.0.0:" + port,
		DatabaseURL:        databaseURL,
		TimeZone:           location,
		PublicBaseURL:      baseURL,
		SessionKey:         []byte(sessionKey),
		GroupsIOEnabled:    groupsIOEnabled,
		SecureCookies:      appEnv == EnvProduction,
		TestControlEnabled: testControlEnabled,
		TestControlKey:     testControlKey,
		ExpectedSchema:     1,
	}, nil
}
