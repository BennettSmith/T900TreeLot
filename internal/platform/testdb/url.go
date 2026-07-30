package testdb

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateTestDatabaseURL rejects URLs that do not clearly name a disposable
// test database. OpenEmpty/OpenMigrated drop foundation tables, so pointing
// TEST_DATABASE_URL at the development treelot database must fail closed.
func ValidateTestDatabaseURL(databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("TEST_DATABASE_URL is not a valid URL: %w", err)
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if name == "" || strings.Contains(name, "/") {
		return fmt.Errorf("TEST_DATABASE_URL must include a single database name path")
	}
	if !strings.HasSuffix(name, "_test") {
		return fmt.Errorf("TEST_DATABASE_URL database %q must end with _test; refusing to reset a non-test database", name)
	}
	return nil
}
