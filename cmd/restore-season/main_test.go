package main

import "testing"

func TestRunRefusesToMutateSchema(t *testing.T) {
	t.Parallel()
	if status := run([]string{"archive.age"}); status != 2 {
		t.Fatalf("status = %d, want 2", status)
	}
}
