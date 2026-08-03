package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/config"
)

func TestLoadRequiresValidatedSettings(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://treelot:treelot@localhost:5432/treelot?sslmode=disable")
	t.Setenv("TREE_LOT_TIME_ZONE", "America/Los_Angeles")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.troop900livermore.org")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
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
	if cfg.BootstrapEnrollmentToken != "bootstrap-enrollment-token-0001" {
		t.Error("BootstrapEnrollmentToken was not loaded")
	}
	wantBootstrapExpiry := time.Date(2026, 8, 6, 16, 0, 0, 0, time.UTC)
	if !cfg.BootstrapTokenExpiresAt.Equal(wantBootstrapExpiry) {
		t.Errorf("BootstrapTokenExpiresAt = %v, want %v", cfg.BootstrapTokenExpiresAt, wantBootstrapExpiry)
	}
	if cfg.WebAuthnRPID != "treelot.troop900livermore.org" {
		t.Errorf("WebAuthnRPID = %q", cfg.WebAuthnRPID)
	}
	if len(cfg.WebAuthnOrigins) != 1 || cfg.WebAuthnOrigins[0] != "https://treelot.troop900livermore.org" {
		t.Errorf("WebAuthnOrigins = %#v", cfg.WebAuthnOrigins)
	}
	if cfg.AuthRateLimitMax != 10 {
		t.Errorf("AuthRateLimitMax = %d, want 10", cfg.AuthRateLimitMax)
	}
	if cfg.AuthRateLimitWindow != 15*60*1_000_000_000 {
		t.Errorf("AuthRateLimitWindow = %v, want 15m", cfg.AuthRateLimitWindow)
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
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
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
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")

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
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")

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
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
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
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
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
	if cfg.WebAuthnRPID != "localhost" {
		t.Errorf("WebAuthnRPID = %q, want localhost", cfg.WebAuthnRPID)
	}
}

func TestAcceptanceRequiresTestControlKey(t *testing.T) {
	t.Setenv("APP_ENV", "acceptance")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "https://treelot.test")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
	t.Setenv("TEST_CONTROL_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load succeeded without TEST_CONTROL_KEY in acceptance")
	}
}

func TestLoadValidatesBootstrapAndAuthSettings(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "short")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "BOOTSTRAP_ENROLLMENT_TOKEN") {
		t.Fatalf("expected BOOTSTRAP_ENROLLMENT_TOKEN error, got %v", err)
	}

	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "BOOTSTRAP_TOKEN_EXPIRES_AT") {
		t.Fatalf("expected required BOOTSTRAP_TOKEN_EXPIRES_AT error, got %v", err)
	}

	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "not-a-timestamp")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "BOOTSTRAP_TOKEN_EXPIRES_AT") {
		t.Fatalf("expected BOOTSTRAP_TOKEN_EXPIRES_AT error, got %v", err)
	}

	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00-07:00")
	t.Setenv("WEBAUTHN_RP_ID", "example.test")
	t.Setenv("AUTH_RATE_LIMIT_MAX", "7")
	t.Setenv("AUTH_RATE_LIMIT_WINDOW", "5m")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantExpiry := time.Date(2026, 8, 6, 23, 0, 0, 0, time.UTC)
	if !cfg.BootstrapTokenExpiresAt.Equal(wantExpiry) {
		t.Errorf("BootstrapTokenExpiresAt = %v, want %v", cfg.BootstrapTokenExpiresAt, wantExpiry)
	}
	if cfg.WebAuthnRPID != "example.test" {
		t.Errorf("WebAuthnRPID = %q, want example.test", cfg.WebAuthnRPID)
	}
	if cfg.AuthRateLimitMax != 7 || cfg.AuthRateLimitWindow != 5*60*1_000_000_000 {
		t.Errorf("rate limit = %d/%v, want 7/5m", cfg.AuthRateLimitMax, cfg.AuthRateLimitWindow)
	}
}
