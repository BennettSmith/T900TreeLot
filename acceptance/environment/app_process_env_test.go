package environment

import (
	"strings"
	"testing"
	"time"
)

func TestUnmigratedAppProcessEnvIncludesBootstrapConfig(t *testing.T) {
	expiresAt := time.Date(2099, 1, 2, 3, 4, 5, 0, time.UTC).Format(time.RFC3339)
	env := UnmigratedAppProcessEnv(
		"postgres://treelot:treelot@127.0.0.1:5433/treelot_unmigrated?sslmode=disable",
		"acceptance-test-control-key",
		expiresAt,
	)

	required := []string{
		"BOOTSTRAP_ENROLLMENT_TOKEN=acceptance-bootstrap-token-0001",
		"BOOTSTRAP_TOKEN_EXPIRES_AT=" + expiresAt,
		"AUTH_RATE_LIMIT_MAX=20",
		"SESSION_KEY=0123456789abcdef0123456789abcdef",
		"TREE_LOT_TIME_ZONE=America/Los_Angeles",
		"PUBLIC_BASE_URL=https://treelot.test",
		"GROUPS_IO_ENABLED=false",
		"TEST_CONTROL_KEY=acceptance-test-control-key",
		"DATABASE_URL=postgres://treelot:treelot@127.0.0.1:5433/treelot_unmigrated?sslmode=disable",
		"APP_ENV=acceptance",
	}
	joined := strings.Join(env, "\n")
	for _, want := range required {
		if !containsExactEnv(env, want) {
			t.Fatalf("missing required process env %q in:\n%s", want, joined)
		}
	}
}

func TestReportsSchemaIncompatibilityRejectsConfigOnlyFailures(t *testing.T) {
	if !ReportsSchemaIncompatibility(`{"time":"...","level":"ERROR","msg":"schema incompatible","error":"schema version 0 is incompatible with expected version 1"}`) {
		t.Fatal("expected schema incompatible log to count as schema failure")
	}
	if ReportsSchemaIncompatibility(`{"time":"...","level":"ERROR","msg":"load config","error":"BOOTSTRAP_ENROLLMENT_TOKEN must be at least 24 characters"}`) {
		t.Fatal("config-only failure must not count as schema incompatibility")
	}
	if ReportsSchemaIncompatibility(`{"msg":"open database","error":"connection refused"}`) {
		t.Fatal("database open failure must not count as schema incompatibility")
	}
}

func containsExactEnv(env []string, want string) bool {
	for _, item := range env {
		if item == want {
			return true
		}
	}
	return false
}
