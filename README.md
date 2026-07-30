# Troop 900 Tree Lot Shift Scheduler

Go modular monolith for coordinating Troop 900 tree-lot volunteer shifts. See
[`docs/use-cases.md`](docs/use-cases.md), [`docs/architecture.md`](docs/architecture.md),
and [`docs/user-stories/roadmap.md`](docs/user-stories/roadmap.md).

## Prerequisites

- Go (see `go.mod`)
- Node.js 22+ (Tailwind asset build only)
- Docker and Docker Compose (local stack and acceptance)

Copy [`.env.example`](.env.example) to `.env` for local overrides.

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
