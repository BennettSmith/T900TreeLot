package testdb_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestOpenEmptyRejectsDevelopmentDatabaseName(t *testing.T) {
	if os.Getenv("TEST_OPEN_EMPTY_UNSAFE") == "1" {
		testdb.OpenEmpty(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestOpenEmptyRejectsDevelopmentDatabaseName$", "-test.v")
	cmd.Env = append(os.Environ(),
		"TEST_OPEN_EMPTY_UNSAFE=1",
		"TEST_DATABASE_URL=postgres://treelot:treelot@127.0.0.1:5432/treelot?sslmode=disable",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected OpenEmpty to fail for development database; output:\n%s", out)
	}
	if !strings.Contains(string(out), "unsafe TEST_DATABASE_URL") {
		t.Fatalf("output missing safety failure; got:\n%s", out)
	}
	if !strings.Contains(string(out), "_test") {
		t.Fatalf("output missing _test guidance; got:\n%s", out)
	}
}
