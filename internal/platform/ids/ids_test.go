package ids_test

import (
	"testing"

	"github.com/troop900/treelot/internal/platform/ids"
)

func TestGeneratorReturnsOpaqueUniqueLookingIDs(t *testing.T) {
	generator := ids.NewGenerator()
	first, err := generator.NewID()
	if err != nil {
		t.Fatalf("NewID first: %v", err)
	}
	second, err := generator.NewID()
	if err != nil {
		t.Fatalf("NewID second: %v", err)
	}
	if first == "" || second == "" || first == second {
		t.Fatalf("ids = %q, %q", first, second)
	}
}
