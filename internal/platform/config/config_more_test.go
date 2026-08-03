package config_test

import (
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/platform/config"
)

func TestLoadRejectsInvalidAppEnvPortAndSessionKey(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/treelot")
	t.Setenv("TREE_LOT_TIME_ZONE", "UTC")
	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BOOTSTRAP_ENROLLMENT_TOKEN", "bootstrap-enrollment-token-0001")
	t.Setenv("BOOTSTRAP_TOKEN_EXPIRES_AT", "2026-08-06T16:00:00Z")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV error, got %v", err)
	}

	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "abc")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("expected PORT error, got %v", err)
	}

	t.Setenv("PORT", "8080")
	t.Setenv("SESSION_KEY", "short")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "SESSION_KEY") {
		t.Fatalf("expected SESSION_KEY error, got %v", err)
	}

	t.Setenv("SESSION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("PUBLIC_BASE_URL", "not-a-url")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected PUBLIC_BASE_URL error")
	}

	t.Setenv("PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("GROUPS_IO_ENABLED", "maybe")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "GROUPS_IO_ENABLED") {
		t.Fatalf("expected GROUPS_IO_ENABLED error, got %v", err)
	}
}
