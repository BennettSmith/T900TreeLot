package postgres_test

import (
	"context"
	"testing"
	"time"

	familypostgres "github.com/troop900/treelot/internal/families/adapters/postgres"
	families "github.com/troop900/treelot/internal/families/application"
	"github.com/troop900/treelot/internal/platform/testdb"
)

func TestProfileCreatorCreatesPersonalProfile(t *testing.T) {
	db := testdb.OpenMigrated(t)
	creator := familypostgres.NewProfileCreator(db)
	now := time.Date(2026, 7, 31, 22, 45, 0, 0, time.UTC)

	err := creator.CreatePersonalProfile(context.Background(), families.PersonalProfile{
		ID:                   "person-1",
		FirstName:            "First",
		LastName:             "Admin",
		PreferredDisplayName: "First",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		t.Fatalf("CreatePersonalProfile: %v", err)
	}

	var firstName string
	if err := db.QueryRow(context.Background(), `SELECT first_name FROM people WHERE id = 'person-1'`).Scan(&firstName); err != nil {
		t.Fatalf("read person: %v", err)
	}
	if firstName != "First" {
		t.Fatalf("first_name = %q", firstName)
	}
}
