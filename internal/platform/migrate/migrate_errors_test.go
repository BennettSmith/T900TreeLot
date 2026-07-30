package migrate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/troop900/treelot/internal/platform/migrate"
)

func TestLoadMigrationsRejectsDuplicatesAndEmpty(t *testing.T) {
	t.Parallel()

	empty := t.TempDir()
	if _, err := migrate.LatestAvailableVersion(empty); err == nil {
		t.Fatal("expected empty directory error")
	}

	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("000001_a.sql", "SELECT 1;")
	write("000001_b.sql", "SELECT 1;")
	if _, err := migrate.LatestAvailableVersion(dir); err == nil {
		t.Fatal("expected duplicate version error")
	}
}
