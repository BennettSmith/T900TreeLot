package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/troop900/treelot/internal/platform/postgres"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://treelot:treelot@localhost:5432/treelot_test?sslmode=disable"
	}
	return url
}

func TestOpenPingsDatabase(t *testing.T) {
	db, err := postgres.Open(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenRejectsUnreachableDatabase(t *testing.T) {
	_, err := postgres.Open(context.Background(), "postgres://treelot:treelot@127.0.0.1:1/nope?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("Open succeeded for unreachable database")
	}
}
