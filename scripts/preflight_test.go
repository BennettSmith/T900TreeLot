package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptancePreflightReportsOccupiedPort(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\necho 'go version go1.26.4 darwin/arm64'\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\necho 'v22.0.0'\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "docker", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "lsof", `#!/bin/sh
case "$*" in
  *28081*) echo "old-web 1234 user 10u IPv4 TCP *:28081 (LISTEN)"; exit 0 ;;
  *) exit 1 ;;
esac
`)

	command := exec.Command("bash", "preflight.sh", "acceptance")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
		"PREFLIGHT_HOST_NETWORK_ATTEMPTS=1",
		"ACCEPTANCE_PRODUCTION_PORT=28081",
	)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("preflight succeeded with configured production port occupied")
	}
	if !strings.Contains(string(output), "Port 28081 is in use") {
		t.Fatalf("output = %q, want occupied port diagnostic", output)
	}
}

func TestAcceptancePreflightReportsHostNetworkFailure(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\necho 'go version go1.26.4 darwin/arm64'\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\necho 'v22.0.0'\n")
	writeExecutable(t, bin, "lsof", "#!/bin/sh\nexit 1\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 1\n")
	writeExecutable(t, bin, "docker", "#!/bin/sh\nexit 0\n")

	command := exec.Command("bash", "preflight.sh", "acceptance")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
		"PREFLIGHT_HOST_NETWORK_ATTEMPTS=1",
	)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("preflight succeeded when host networking was unreachable")
	}
	if !strings.Contains(string(output), "Enable host networking") {
		t.Fatalf("output = %q, want Docker Desktop host-network guidance", output)
	}
}

func TestAcceptancePreflightPassesWhenLocalEnvironmentIsReady(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\necho 'go version go1.26.4 darwin/arm64'\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\necho 'v22.0.0'\n")
	writeExecutable(t, bin, "lsof", "#!/bin/sh\nexit 1\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "docker", `#!/bin/sh
case "$*" in
  *"busybox:1.36 sh -c"*"exec httpd"*) exit 0 ;;
  *"run -d --rm --network host"*) exit 1 ;;
  *) exit 0 ;;
esac
`)

	command := exec.Command("bash", "preflight.sh", "acceptance")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
		"PREFLIGHT_HOST_NETWORK_ATTEMPTS=1",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("preflight failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Acceptance preflight passed") {
		t.Fatalf("output = %q, want success summary", output)
	}
}

func TestDoctorAllowsProjectPostgresOnPort5433(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\necho 'go version go1.26.4 darwin/arm64'\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\necho 'v22.0.0'\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "lsof", `#!/bin/sh
case "$*" in
  *5433*) echo "com.docker 1234 user 10u IPv4 TCP *:5433 (LISTEN)"; exit 0 ;;
  *) exit 1 ;;
esac
`)
	writeExecutable(t, bin, "docker", `#!/bin/sh
case "$*" in
  "compose ps -q postgres") echo "project-postgres" ;;
  "port project-postgres 5432/tcp") echo "0.0.0.0:5433" ;;
esac
exit 0
`)
	writeGitStub(t, bin, ".githooks")

	command := exec.Command("bash", "preflight.sh", "doctor")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor failed for project Postgres: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "project PostgreSQL") {
		t.Fatalf("output = %q, want project PostgreSQL ownership", output)
	}
	if !strings.Contains(string(output), "Tracked Git hooks are installed") {
		t.Fatalf("output = %q, want installed hooks status", output)
	}
}

func TestDoctorAllowsProjectWebOnPort8080(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "lsof", `#!/bin/sh
case "$*" in
  *8080*) echo "com.docker 1234 user 10u IPv4 TCP *:8080 (LISTEN)"; exit 0 ;;
  *) exit 1 ;;
esac
`)
	writeExecutable(t, bin, "docker", `#!/bin/sh
case "$*" in
  "compose ps -q web") echo "project-web" ;;
  "port project-web 8080/tcp") echo "0.0.0.0:8080" ;;
esac
exit 0
`)
	writeGitStub(t, bin, ".githooks")

	command := exec.Command("bash", "preflight.sh", "doctor")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor failed for project web: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Port 8080 is in use by project web") {
		t.Fatalf("output = %q, want project web ownership", output)
	}
}

func TestDoctorReportsMissingTrackedGitHooks(t *testing.T) {
	bin := t.TempDir()
	writeExecutable(t, bin, "go", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "node", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "curl", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "docker", "#!/bin/sh\nexit 0\n")
	writeExecutable(t, bin, "lsof", "#!/bin/sh\nexit 1\n")
	writeGitStub(t, bin, "")

	command := exec.Command("bash", "preflight.sh", "doctor")
	command.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"PREFLIGHT_OS=Darwin",
	)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("doctor passed without tracked Git hooks installed")
	}
	if !strings.Contains(string(output), "run make install-hooks") {
		t.Fatalf("output = %q, want hook installation guidance", output)
	}
}

func TestAcceptanceRunnerCleansBeforePreflight(t *testing.T) {
	contents, err := os.ReadFile("acceptance.sh")
	if err != nil {
		t.Fatalf("read acceptance runner: %v", err)
	}
	script := string(contents)
	cleanup := strings.Index(script, "$DOCKER rm -f")
	preflight := strings.Index(script, "./scripts/preflight.sh acceptance")
	if preflight < 0 {
		t.Fatal("acceptance runner does not invoke preflight")
	}
	if cleanup < 0 || cleanup > preflight {
		t.Fatal("acceptance runner must remove its old containers before checking ports")
	}
}

func TestAcceptanceRunnerUsesConfigurableHighPorts(t *testing.T) {
	contents, err := os.ReadFile("acceptance.sh")
	if err != nil {
		t.Fatalf("read acceptance runner: %v", err)
	}
	script := string(contents)
	for _, contract := range []string{
		`WEB_PORT="${ACCEPTANCE_WEB_PORT:-18080}"`,
		`PRODUCTION_PORT="${ACCEPTANCE_PRODUCTION_PORT:-18081}"`,
		`STUB_PORT="${ACCEPTANCE_STUB_PORT:-18090}"`,
		`-e PORT="$WEB_PORT"`,
		`-e PORT="$PRODUCTION_PORT"`,
		`-e PORT="$STUB_PORT"`,
	} {
		if !strings.Contains(script, contract) {
			t.Errorf("acceptance runner missing configurable-port contract %q", contract)
		}
	}
}

func TestMakefileExposesDoctorAndAcceptancePreflight(t *testing.T) {
	contents, err := os.ReadFile("../Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	makefile := string(contents)
	for _, target := range []string{"doctor:", "acceptance-preflight:"} {
		if !strings.Contains(makefile, target) {
			t.Fatalf("Makefile does not expose %s", target)
		}
	}
}

func TestComposeRequiresOneBootstrapExpiryForEveryApplicationService(t *testing.T) {
	command := exec.Command(
		"docker", "compose", "-f", "../docker-compose.yml",
		"--profile", "dev", "--profile", "acceptance",
		"config", "--format", "json",
	)
	command.Env = environmentWith("BOOTSTRAP_TOKEN_EXPIRES_AT", "")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("Compose config succeeded without BOOTSTRAP_TOKEN_EXPIRES_AT")
	}
	if !strings.Contains(string(output), "BOOTSTRAP_TOKEN_EXPIRES_AT") {
		t.Fatalf("output = %q, want BOOTSTRAP_TOKEN_EXPIRES_AT diagnostic", output)
	}

	const expiry = "2030-01-02T03:04:05Z"
	command = exec.Command(
		"docker", "compose", "-f", "../docker-compose.yml",
		"--profile", "dev", "--profile", "acceptance",
		"config", "--format", "json",
	)
	command.Env = environmentWith("BOOTSTRAP_TOKEN_EXPIRES_AT", expiry)
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("Compose config failed with valid expiry: %v\n%s", err, output)
	}

	var rendered struct {
		Services map[string]struct {
			Environment map[string]string `json:"environment"`
		} `json:"services"`
	}
	if err := json.Unmarshal(output, &rendered); err != nil {
		t.Fatalf("decode Compose config: %v", err)
	}
	for _, service := range []string{"migrate", "web", "worker", "acceptance-web", "acceptance-worker"} {
		if got := rendered.Services[service].Environment["BOOTSTRAP_TOKEN_EXPIRES_AT"]; got != expiry {
			t.Errorf("%s BOOTSTRAP_TOKEN_EXPIRES_AT = %q, want %q", service, got, expiry)
		}
	}
}

func environmentWith(name, value string) []string {
	prefix := name + "="
	environment := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, prefix) {
			environment = append(environment, entry)
		}
	}
	return append(environment, prefix+value)
}

func writeExecutable(t *testing.T, directory, name, contents string) {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func writeGitStub(t *testing.T, directory, hooksPath string) {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	contents := fmt.Sprintf(`#!/bin/sh
case "$*" in
  "config --local --get core.hooksPath")
    printf '%%s\n' %q
    ;;
  "rev-parse --show-toplevel")
    printf '%%s\n' %q
    ;;
esac
`, hooksPath, root)
	writeExecutable(t, directory, "git", contents)
}
