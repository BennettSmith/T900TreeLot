package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitMessageCheckAcceptsConventionalSubjects(t *testing.T) {
	repository, base := newCommitMessageRepository(t)
	for _, subject := range []string{
		"feat: add scheduling command",
		"fix(identity): reject an expired token",
		"refactor!: remove the legacy API",
		"docs(traceability): record implementation evidence",
	} {
		gitCommit(t, repository, subject)
	}

	output, err := runCommitMessageCheck(t, repository, base)
	if err != nil {
		t.Fatalf("check failed for conventional subjects: %v\n%s", err, output)
	}
}

func TestCommitMessageCheckRejectsNonConventionalSubject(t *testing.T) {
	repository, base := newCommitMessageRepository(t)
	gitCommit(t, repository, "Fix the expired bootstrap token")

	output, err := runCommitMessageCheck(t, repository, base)
	if err == nil {
		t.Fatal("check passed for a non-conventional subject")
	}
	if !strings.Contains(output, "Fix the expired bootstrap token") {
		t.Fatalf("output = %q, want rejected subject", output)
	}
	if !strings.Contains(output, "Allowed types:") {
		t.Fatalf("output = %q, want format guidance", output)
	}
}

func TestCommitMessageCheckRejectsUnknownType(t *testing.T) {
	repository, base := newCommitMessageRepository(t)
	gitCommit(t, repository, "change: update bootstrap flow")

	output, err := runCommitMessageCheck(t, repository, base)
	if err == nil {
		t.Fatal("check passed for an unknown conventional-commit type")
	}
	if !strings.Contains(output, "change: update bootstrap flow") {
		t.Fatalf("output = %q, want rejected subject", output)
	}
}

func TestCommitMessageGateIsWiredIntoPrePushAndCI(t *testing.T) {
	assertFileContains(t, "../.githooks/pre-push", "check-commit-messages.sh")
	assertFileContains(t, "../Makefile", "commit-messages:")
	assertFileContains(t, "../Makefile", "install-hooks:")
	assertFileContains(t, "../.github/workflows/ci.yml", "make commit-messages")
	assertFileContains(t, "../.github/workflows/ci.yml", "github.event.pull_request.base.sha")
	assertFileContains(t, "../.github/workflows/ci.yml", "github.event.before")
}

func newCommitMessageRepository(t *testing.T) (string, string) {
	t.Helper()
	repository := t.TempDir()
	runGit(t, repository, "init", "--quiet")
	gitCommit(t, repository, "chore: initialize fixture")
	return repository, strings.TrimSpace(runGit(t, repository, "rev-parse", "HEAD"))
}

func gitCommit(t *testing.T, repository, subject string) {
	t.Helper()
	command := exec.Command("git", "commit", "--quiet", "--allow-empty", "-m", subject)
	command.Dir = repository
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Commit Test",
		"GIT_AUTHOR_EMAIL=commit-test@example.org",
		"GIT_COMMITTER_NAME=Commit Test",
		"GIT_COMMITTER_EMAIL=commit-test@example.org",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git commit %q: %v\n%s", subject, err, output)
	}
}

func runGit(t *testing.T, repository string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = repository
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func runCommitMessageCheck(t *testing.T, repository, base string) (string, error) {
	t.Helper()
	checker, err := filepath.Abs("check-commit-messages.sh")
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(runGit(t, repository, "rev-parse", "HEAD"))
	command := exec.Command("bash", checker, base, head)
	command.Dir = repository
	output, err := command.CombinedOutput()
	return string(output), err
}

func assertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(contents), expected) {
		t.Fatalf("%s does not contain %q", path, expected)
	}
}
