package config_test

import (
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/platform/config"
)

func TestLoadRequiresValidatedSettings(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://treelot:treelot@localhost:5432/treelot?sslmode=disable")
	t.Setenv("TREE_LOT_TIME_ZONE", "America/Los_Angeles")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.troop900livermore.org")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("GROUPS_IO_ENABLED", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppEnv != config.EnvProduction {
		t.Errorf("AppEnv = %q, want production", cfg.AppEnv)
	}
	if cfg.ListenAddress != "0.0.0.0:8080" {
		t.Errorf("ListenAddress = %q", cfg.ListenAddress)
	}
	if cfg.TimeZone.String() != "America/Los_Angeles" {
		t.Errorf("TimeZone = %q", cfg.TimeZone)
	}
	if cfg.PublicBaseURL.String() != "https://treelot.troop900livermore.org" {
		t.Errorf("PublicBaseURL = %q", cfg.PublicBaseURL)
	}
	if cfg.GroupsIOEnabled {
		t.Error("GroupsIOEnabled = true, want false")
	}
	if !cfg.SecureCookies {
		t.Error("SecureCookies = false in production")
	}
}

func TestAcceptanceDisablesSecureCookiesForHTTPDrivers(t *testing.T) {
	t.Setenv("APP_ENV", "acceptance")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.test")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("TEST_CONTROL_KEY", "test-control-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SecureCookies {
		t.Error("SecureCookies = true in acceptance; HTTP drivers cannot use Secure cookies")
	}
	if !cfg.TestControlEnabled {
		t.Error("TestControlEnabled = false in acceptance")
	}
}

func TestLoadRejectsMissingRequiredValues(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("TREE_LOT_TIME_ZONE", "America/Los_Angeles")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.troop900livermore.org")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load succeeded with empty DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("error = %v, want DATABASE_URL mention", err)
	}
}

func TestLoadRejectsInvalidTimeZone(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "Not/A_Zone")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load succeeded with invalid time zone")
	}
}

func TestLoadRequiresHTTPSOutsideDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "acceptance")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("TEST_CONTROL_KEY", "test-control-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load succeeded with http PUBLIC_BASE_URL outside development")
	}
}

func TestDevelopmentAllowsHTTPAndDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("GROUPS_IO_ENABLED", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ListenAddress != "0.0.0.0:8080" {
		t.Errorf("ListenAddress = %q", cfg.ListenAddress)
	}
	if cfg.SecureCookies {
		t.Error("SecureCookies = true in development")
	}
	if cfg.TestControlEnabled {
		t.Error("TestControlEnabled = true in development")
	}
}

func TestAcceptanceRequiresTestControlKey(t *testing.T) {
	t.Setenv("APP_ENV", "acceptance")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.test")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("TEST_CONTROL_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load succeeded without TEST_CONTROL_KEY in acceptance")
	}
}
