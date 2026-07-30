package traceability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAcceptsRevisionMatchedEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/use-cases.md", "### Use Case 0: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/bootstrap/us-001-bootstrap.md", "# US-001: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/roadmap.md", "## INC-01 — Foundation\n## INC-02 — Bootstrap\n")
	writeFixture(t, root, "acceptance/cases/bootstrap_test.go", `package cases

// Trace: UC-0@r1 US-001@r1
func TestBootstrap() {}
`)
	writeFixture(t, root, "traceability/manifest.yaml", `schema_version: 1
baseline:
  requirements_commit: c66a77c
  use_case_document_version: "3.13"
use_cases:
  UC-0:
    current_revision: 1
    revisions:
      - revision: 1
        requirements_status: accepted
stories:
  US-001:
    path: docs/user-stories/bootstrap/us-001-bootstrap.md
    increment: INC-02
    current_revision: 1
    revisions:
      - revision: 1
        requirements_status: accepted
        delivery_status: verified
        source_use_cases: [UC-0@r1]
        implementation_pr: 12
increments:
  INC-01:
    delivery_status: verified
    implementation_pr: 4
  INC-02:
    delivery_status: verified
`)

	manifest, err := Load(filepath.Join(root, "traceability/manifest.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if problems := Validate(root, manifest); len(problems) != 0 {
		t.Fatalf("Validate() problems = %v", problems)
	}
}

func TestValidateRejectsVerifiedStoryWithStaleAcceptanceRevision(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/use-cases.md", "### Use Case 0: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/bootstrap/us-001-bootstrap.md", "# US-001: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/roadmap.md", "## INC-02 — Bootstrap\n")
	writeFixture(t, root, "acceptance/cases/bootstrap_test.go", "// Trace: UC-0@r1 US-001@r1\n")
	writeFixture(t, root, "traceability/manifest.yaml", `schema_version: 1
baseline:
  requirements_commit: c66a77c
  use_case_document_version: "3.13"
use_cases:
  UC-0:
    current_revision: 2
    revisions:
      - revision: 1
        requirements_status: superseded
      - revision: 2
        requirements_status: accepted
stories:
  US-001:
    path: docs/user-stories/bootstrap/us-001-bootstrap.md
    increment: INC-02
    current_revision: 2
    revisions:
      - revision: 1
        requirements_status: superseded
        delivery_status: verified
        source_use_cases: [UC-0@r1]
        implementation_pr: 11
      - revision: 2
        requirements_status: accepted
        delivery_status: verified
        source_use_cases: [UC-0@r2]
        implementation_pr: 12
increments:
  INC-02:
    delivery_status: verified
`)

	manifest, err := Load(filepath.Join(root, "traceability/manifest.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	problems := Validate(root, manifest)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "US-001@r2 is verified but has no matching acceptance trace") {
		t.Fatalf("Validate() problems = %v, want stale revision problem", problems)
	}
}

func TestValidateRejectsUnknownAndMissingArtifacts(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/use-cases.md", "### Use Case 0: Bootstrap\n### Use Case 1: Household\n")
	writeFixture(t, root, "docs/user-stories/bootstrap/us-001-bootstrap.md", "# US-001: Bootstrap\n")
	writeFixture(t, root, "docs/user-stories/bootstrap/us-002-sign-in.md", "# US-002: Sign in\n")
	writeFixture(t, root, "docs/user-stories/roadmap.md", "## INC-02 — Bootstrap\n")
	writeFixture(t, root, "acceptance/cases/bootstrap_test.go", "// Trace: UC-99@r1 US-001@r1\n")
	writeFixture(t, root, "traceability/manifest.yaml", `schema_version: 1
baseline:
  requirements_commit: c66a77c
  use_case_document_version: "3.13"
use_cases:
  UC-0:
    current_revision: 1
    revisions:
      - revision: 1
        requirements_status: accepted
stories:
  US-001:
    path: docs/user-stories/bootstrap/us-001-bootstrap.md
    increment: INC-02
    current_revision: 1
    revisions:
      - revision: 1
        requirements_status: accepted
        delivery_status: planned
        source_use_cases: [UC-0@r1]
increments:
  INC-02:
    delivery_status: planned
`)

	manifest, err := Load(filepath.Join(root, "traceability/manifest.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	joined := strings.Join(Validate(root, manifest), "\n")
	for _, want := range []string{
		"UC-1 exists in docs/use-cases.md but not the manifest",
		"US-002 exists in docs/user-stories but not the manifest",
		"acceptance trace references unknown UC-99@r1",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("Validate() problems = %q, want %q", joined, want)
		}
	}
}

func TestRenderReportResolvesPullRequestToMergeSHA(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 1,
		UseCases: map[string]UseCase{
			"UC-0": {
				CurrentRevision: 1,
				Revisions:       []UseCaseRevision{{Revision: 1, RequirementsStatus: "accepted"}},
			},
		},
		Stories: map[string]Story{
			"US-001": {
				Path:            "docs/user-stories/bootstrap/us-001-bootstrap.md",
				Increment:       "INC-02",
				CurrentRevision: 1,
				Revisions: []StoryRevision{{
					Revision:           1,
					RequirementsStatus: "accepted",
					DeliveryStatus:     "verified",
					SourceUseCases:     []string{"UC-0@r1"},
					ImplementationPR:   12,
				}},
			},
		},
		Increments: map[string]Increment{
			"INC-02": {DeliveryStatus: "verified"},
		},
	}

	report := RenderReport(manifest, map[int]string{12: "abc1234"})
	for _, want := range []string{"UC-0@r1", "US-001@r1", "verified", "#12", "`abc1234`"} {
		if !strings.Contains(report, want) {
			t.Errorf("RenderReport() missing %q:\n%s", want, report)
		}
	}
}

func writeFixture(t *testing.T, root, name, contents string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
