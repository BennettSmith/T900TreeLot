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
make ci            # fast checks: assets, format, vet, coverage
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
