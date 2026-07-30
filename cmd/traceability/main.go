package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/troop900/treelot/internal/traceability"
)

const reportPath = "docs/traceability.md"

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "check" && os.Args[1] != "write") {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/traceability <check|write>")
		os.Exit(2)
	}

	manifest, err := traceability.Load("traceability/manifest.yaml")
	if err != nil {
		fail(err)
	}
	if problems := traceability.Validate(".", manifest); len(problems) != 0 {
		for _, problem := range problems {
			fmt.Fprintln(os.Stderr, "traceability:", problem)
		}
		os.Exit(1)
	}

	logOutput, err := exec.Command("git", "log", "--format=%h%x09%s").Output()
	if err != nil {
		fail(fmt.Errorf("read Git history: %w", err))
	}
	report := []byte(traceability.RenderReport(manifest, traceability.ParseMergeLog(string(logOutput))))

	switch os.Args[1] {
	case "write":
		if err := os.WriteFile(reportPath, report, 0o644); err != nil {
			fail(err)
		}
		fmt.Printf("wrote %s\n", reportPath)
	case "check":
		current, err := os.ReadFile(filepath.Clean(reportPath))
		if err != nil {
			fail(err)
		}
		if !bytes.Equal(current, report) {
			fmt.Fprintf(os.Stderr, "%s is stale; run 'go run ./cmd/traceability write'.\n", reportPath)
			os.Exit(1)
		}
		fmt.Println("traceability manifest and report are valid")
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "traceability:", err)
	os.Exit(1)
}
