package migrate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/troop900/treelot/internal/platform/migrate"
)

func TestLatestAvailableVersionAndDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("..", "..", "..", "migrations")
	version, err := migrate.LatestAvailableVersion(dir)
	if err != nil {
		t.Fatalf("LatestAvailableVersion: %v", err)
	}
	if version != 5 {
		t.Fatalf("version = %d, want 5", version)
	}

	found, err := migrate.Directory(dir)
	if err != nil {
		t.Fatalf("Directory: %v", err)
	}
	if found != dir {
		t.Fatalf("Directory = %q", found)
	}
	if _, err := migrate.Directory(filepath.Join(os.TempDir(), "missing-migrations-dir")); err == nil {
		t.Fatal("Directory succeeded for missing path")
	}
}
