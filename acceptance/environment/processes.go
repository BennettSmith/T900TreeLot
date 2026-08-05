//go:build acceptance

package environment

import (
	"fmt"
	"os/exec"
	"strings"
)

// ProcessDriver starts production-image entry points for foundation invariants.
type ProcessDriver struct {
	config Config
}

// NewProcessDriver constructs a process driver.
func NewProcessDriver(config Config) *ProcessDriver {
	return &ProcessDriver{config: config}
}

func (p *ProcessDriver) dockerCommand(args ...string) *exec.Cmd {
	parts := strings.Fields(p.config.Docker)
	if len(parts) == 0 {
		parts = []string{"docker"}
	}
	name := parts[0]
	commandArgs := append(append([]string{}, parts[1:]...), args...)
	return exec.Command(name, commandArgs...)
}

// RejectsUnmigratedDatabaseWithoutSchemaChange runs web and worker against an
// empty database and verifies they exit non-zero due to schema incompatibility
// without creating schema_migrations.
func (p *ProcessDriver) RejectsUnmigratedDatabaseWithoutSchemaChange() error {
	envPairs := UnmigratedAppProcessEnv(
		p.config.UnmigratedDBURL,
		p.config.TestControlKey,
		BootstrapTokenExpiresAtForProbe(),
	)
	dockerEnv := make([]string, 0, len(envPairs)*2)
	for _, pair := range envPairs {
		dockerEnv = append(dockerEnv, "-e", pair)
	}

	for _, entrypoint := range []string{"/app/web", "/app/worker"} {
		args := []string{"run", "--rm", "--network", "host", "--entrypoint", entrypoint}
		args = append(args, dockerEnv...)
		args = append(args, p.config.Image)
		cmd := p.dockerCommand(args...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return fmt.Errorf("%s started against unmigrated database; output=%s", entrypoint, output)
		}
		if !ReportsSchemaIncompatibility(string(output)) {
			return fmt.Errorf("%s exited before schema compatibility check; want %q in output=%s",
				entrypoint, schemaIncompatibleLogSignal, output)
		}
	}

	probe := p.dockerCommand("run", "--rm", "--network", "host",
		"postgres:16-alpine",
		"psql", p.config.UnmigratedDBURL, "-tAc",
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)`,
	)
	output, err := probe.CombinedOutput()
	if err != nil {
		return fmt.Errorf("probe unmigrated schema: %w (%s)", err, output)
	}
	if strings.TrimSpace(string(output)) != "f" {
		return fmt.Errorf("unmigrated database was mutated; schema_migrations exists (%s)", output)
	}
	return nil
}
