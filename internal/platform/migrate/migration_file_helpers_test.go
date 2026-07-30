package migrate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readMigrationsFile(t *testing.T) (string, error) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "migrations", "000001_foundation.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func containsAntiPattern(contents string) bool {
	return strings.Contains(contents, "DEFAULT (now() AT TIME ZONE") ||
		strings.Contains(contents, "DEFAULT now() AT TIME ZONE")
}
