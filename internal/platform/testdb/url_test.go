package testdb_test

import (
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestValidateTestDatabaseURLAcceptsTestNames(t *testing.T) {
	t.Parallel()

	for _, databaseURL := range []string{
		"postgres://treelot:treelot@localhost:5432/treelot_test?sslmode=disable",
		"postgres://localhost/unit_test",
		"postgresql://user:pass@db:5432/treelot_ci_test",
	} {
		if err := testdb.ValidateTestDatabaseURL(databaseURL); err != nil {
			t.Fatalf("ValidateTestDatabaseURL(%q): %v", databaseURL, err)
		}
	}
}

func TestValidateTestDatabaseURLRejectsDevelopmentDatabase(t *testing.T) {
	t.Parallel()

	err := testdb.ValidateTestDatabaseURL("postgres://treelot:treelot@localhost:5432/treelot?sslmode=disable")
	if err == nil {
		t.Fatal("expected rejection of development database name treelot")
	}
	if !strings.Contains(err.Error(), "_test") {
		t.Fatalf("error = %v, want _test guidance", err)
	}
}

func TestValidateTestDatabaseURLRejectsMissingName(t *testing.T) {
	t.Parallel()

	if err := testdb.ValidateTestDatabaseURL("postgres://localhost/"); err == nil {
		t.Fatal("expected rejection of empty database name")
	}
}
