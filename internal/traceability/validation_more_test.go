package traceability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsUnknownManifestFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(path, []byte("schema_version: 1\nunexpected: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "field unexpected not found") {
		t.Fatalf("Load() error = %v, want unknown field error", err)
	}
}

func TestParseMergeLogFindsSquashMergePullRequests(t *testing.T) {
	log := "b78cf9e\tfeat: foundation (#4)\n" +
		"not-a-record\n" +
		"abcdef0\tdocs: no pull request\n" +
		"1234567\tfix: later change (#12)\n"

	got := ParseMergeLog(log)
	if got[4] != "b78cf9e" || got[12] != "1234567" || len(got) != 2 {
		t.Fatalf("ParseMergeLog() = %v", got)
	}
}

func TestValidateReportsMalformedRevisionHistoryAndEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/use-cases.md", "### Use Case 0: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/bootstrap/us-001-bootstrap.md", "# US-001: Bootstrap\n\n- **Source use cases:** UC-0\n")
	writeFixture(t, root, "docs/user-stories/roadmap.md", "## INC-02 — Bootstrap\n")
	writeFixture(t, root, "acceptance/cases/bootstrap_test.go", "// Trace: malformed UC-0@r9 INC-99 US-999@r1\n")
	writeFixture(t, root, "traceability/manifest.yaml", `schema_version: 2
baseline: {}
use_cases:
  UC-0:
    current_revision: 3
    revisions:
      - revision: 0
        requirements_status: invalid
      - revision: 1
        requirements_status: accepted
      - revision: 1
        requirements_status: superseded
  UC-9:
    current_revision: 1
    revisions:
      - revision: 1
        requirements_status: superseded
stories:
  US-001:
    path: docs/user-stories/bootstrap/us-001-bootstrap.md
    increment: INC-03
    current_revision: 2
    revisions:
      - revision: 0
        requirements_status: invalid
        delivery_status: invalid
        source_use_cases: []
      - revision: 1
        requirements_status: accepted
        delivery_status: verified
        source_use_cases: [bad-reference, UC-0@r9]
increments:
  INC-02:
    delivery_status: invalid
  INC-03:
    delivery_status: verified
`)

	manifest, err := Load(filepath.Join(root, "traceability/manifest.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(Validate(root, manifest), "\n")
	for _, want := range []string{
		"schema_version must be 1",
		"baseline requirements_commit is required",
		"baseline use_case_document_version is required",
		"UC-0 has duplicate revision 1",
		"UC-0 current revision 3 does not exist",
		"UC-9 exists in the manifest but not docs/use-cases.md",
		"UC-9@r1 is current but superseded",
		"US-001 references unknown increment INC-03",
		"US-001 current revision 2 does not exist",
		"US-001@r1 is verified but has no implementation_pr",
		"US-001@r1 has invalid source use-case reference \"bad-reference\"",
		"US-001@r1 references unknown UC-0@r9",
		"INC-02 has invalid delivery status \"invalid\"",
		"INC-03 exists in the manifest but not the roadmap",
		"acceptance trace has invalid reference \"malformed\"",
		"acceptance trace references unknown UC-0@r9",
		"acceptance trace references unknown INC-99",
		"acceptance trace references unknown US-999@r1",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("Validate() missing %q:\n%s", want, joined)
		}
	}
}

func TestValidateReportsUnreadableRepositoryDocuments(t *testing.T) {
	problems := strings.Join(Validate(t.TempDir(), Manifest{SchemaVersion: 1}), "\n")
	for _, want := range []string{"read use cases:", "read user stories:", "read roadmap:", "read acceptance traces:"} {
		if !strings.Contains(problems, want) {
			t.Errorf("Validate() missing %q:\n%s", want, problems)
		}
	}
}
