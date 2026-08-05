package environment

import (
	"os"
	"strings"
	"time"
)

// schemaIncompatibleLogSignal is emitted by web and worker when EnsureCompatible fails.
// The unmigrated-DB probe must observe this signal so config-only exits cannot pass.
const schemaIncompatibleLogSignal = "schema incompatible"

// UnmigratedAppProcessEnv returns the docker -e values needed for web/worker to pass
// config.Load and reach migrate.EnsureCompatible against an unmigrated database.
// Values match scripts/acceptance.sh common_env plus acceptance APP_ENV wiring.
func UnmigratedAppProcessEnv(databaseURL, testControlKey, bootstrapExpiresAt string) []string {
	return []string{
		"APP_ENV=acceptance",
		"PORT=18082",
		"DATABASE_URL=" + databaseURL,
		"TREE_LOT_TIME_ZONE=America/Los_Angeles",
		"PUBLIC_BASE_URL=https://treelot.test",
		"SESSION_KEY=0123456789abcdef0123456789abcdef",
		"BOOTSTRAP_ENROLLMENT_TOKEN=acceptance-bootstrap-token-0001",
		"BOOTSTRAP_TOKEN_EXPIRES_AT=" + bootstrapExpiresAt,
		"AUTH_RATE_LIMIT_MAX=20",
		"GROUPS_IO_ENABLED=false",
		"TEST_CONTROL_KEY=" + testControlKey,
	}
}

// BootstrapTokenExpiresAtForProbe returns the bootstrap expiry used by acceptance
// process probes. Prefer the value exported by scripts/acceptance.sh.
func BootstrapTokenExpiresAtForProbe() string {
	if raw := strings.TrimSpace(os.Getenv("BOOTSTRAP_TOKEN_EXPIRES_AT")); raw != "" {
		return raw
	}
	if raw := strings.TrimSpace(os.Getenv("ACCEPTANCE_BOOTSTRAP_TOKEN_EXPIRES_AT")); raw != "" {
		return raw
	}
	return time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
}

// ReportsSchemaIncompatibility reports whether process output shows a schema
// compatibility failure rather than an earlier config or connectivity error.
func ReportsSchemaIncompatibility(output string) bool {
	return strings.Contains(output, schemaIncompatibleLogSignal)
}
