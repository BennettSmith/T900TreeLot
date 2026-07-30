# Troop 900 Tree Lot Shift Scheduler

Go modular monolith for coordinating Troop 900 tree-lot volunteer shifts. See
[`docs/use-cases.md`](docs/use-cases.md), [`docs/architecture.md`](docs/architecture.md),
and [`docs/user-stories/roadmap.md`](docs/user-stories/roadmap.md).

## Prerequisites

- Go (see `go.mod`)
- Node.js 22+ (Tailwind asset build only)
- Docker and Docker Compose (local stack, acceptance, and local `make ci` fallback)

Copy [`.env.example`](.env.example) to `.env` for local overrides.

`make ci` / `make test` need a disposable Postgres database whose name ends with
`_test` (default `treelot_test`). Helpers drop foundation tables there, so the
development `treelot` database is rejected.

Resolution order for `TEST_DATABASE_URL`:

1. Use `TEST_DATABASE_URL` when set and reachable (GitHub Actions does this).
2. Otherwise use `postgres://treelot:treelot@127.0.0.1:5432/treelot_test` when that
   role/database already exists.
3. Otherwise start Compose Postgres on host port **5433** and create `treelot_test`.

A failure like `role "treelot" does not exist` means the Postgres on `:5432` is
not the project test database (for example a Homebrew server). Unset
`TEST_DATABASE_URL`, ensure Docker is running, and re-run `make ci` so Compose
can supply `:5433`.

## Common commands

```sh
make help          # list targets
make doctor        # diagnose local tools, Docker, networking guidance, and ports
make traceability  # validate requirement revisions, evidence, and generated report
make ci            # fast checks: traceability, assets, format, vet, coverage
make test          # unit/component tests
make up            # production image + postgres + migrate + web + worker
make migrate       # apply schema via migrate entry point only
make acceptance    # whole-system foundation specs against the production image
make down          # stop Compose services (DOWN_FLAGS=-v removes volumes)
make logs          # tail postgres/web/worker
make ps            # Compose status
make showcase      # design-system gallery (development)
```

## Local Docker flow

```sh
make up
# open http://localhost:8080/
make down
```

Migrations are applied only by `cmd/migrate`. Web and worker validate schema
compatibility and refuse to start on mismatch.

## Acceptance tests

Foundation acceptance specs live under `acceptance/` (ATDD four-layer layout) and
run only with `-tags=acceptance` against the production Docker image:

```sh
make acceptance
```

`make acceptance` builds the image, starts PostgreSQL via Compose, runs migrate/web/worker
from the production image (host networking), then executes the suite. Set
`ACCEPTANCE_KEEP=1` to leave containers running after the suite.

The acceptance runner removes its own previous containers and then runs a
non-destructive preflight before building. The preflight checks required tools
and ports `5433`, `8080`, `8081`, and `8090`. On macOS it also starts a temporary
Alpine container to verify Docker Desktop host networking. Run the same focused
check directly with:

```sh
make acceptance-preflight
```

The preflight reports conflicts but never stops unrelated processes or
containers. Resolve any named conflict and rerun it.

## Requirements traceability

[`docs/traceability.md`](docs/traceability.md) reports the accepted use-case and
user-story revisions, delivery status, increment, implementation PR, and merged
Git SHA. It is generated from `traceability/manifest.yaml`; do not edit the
report directly.

When requirements or delivery status change, follow
[`docs/traceability-process.md`](docs/traceability-process.md). Acceptance tests
identify exact revisions with `// Trace:` metadata so a superseded executable
example cannot verify a newer requirement revision.

```sh
go run ./cmd/traceability write
make traceability
```
