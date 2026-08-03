#!/usr/bin/env bash
# Resolve a disposable TEST_DATABASE_URL for make ci / go test.
# Prints only the URL on stdout. Prefers an already-reachable URL, otherwise
# starts Compose Postgres (host port 5433) and ensures treelot_test exists.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DEFAULT_URL="postgres://treelot:treelot@127.0.0.1:5432/treelot_test?sslmode=disable"
COMPOSE_URL="postgres://treelot:treelot@127.0.0.1:5433/treelot_test?sslmode=disable"

DOCKER="${DOCKER:-docker}"
if ! $DOCKER info >/dev/null 2>&1; then
  DOCKER="sudo docker"
fi
COMPOSE="$DOCKER compose"

# Compose interpolates the shared application environment even when only the
# database service is selected. This expired sentinel permits that parse; this
# helper never starts an application service.
export BOOTSTRAP_TOKEN_EXPIRES_AT="${BOOTSTRAP_TOKEN_EXPIRES_AT:-1970-01-01T00:00:00Z}"

ping_db() {
  local url="$1"
  go run ./scripts/dbping "$url" >/dev/null 2>&1
}

emit() {
  local url="$1"
  echo "Using TEST_DATABASE_URL=${url}" >&2
  printf '%s\n' "$url"
}

if [[ -n "${TEST_DATABASE_URL:-}" ]]; then
  if ping_db "$TEST_DATABASE_URL"; then
    emit "$TEST_DATABASE_URL"
    exit 0
  fi
  echo "TEST_DATABASE_URL is set but not reachable:" >&2
  echo "  ${TEST_DATABASE_URL}" >&2
  echo "Create role/database treelot/treelot_test, or unset TEST_DATABASE_URL to use Compose on port 5433." >&2
  exit 1
fi

if ping_db "$DEFAULT_URL"; then
  emit "$DEFAULT_URL"
  exit 0
fi

echo "No treelot_test on localhost:5432; starting Compose Postgres on :5433..." >&2
$COMPOSE up -d postgres >/dev/null

deadline=$((SECONDS + 90))
until $COMPOSE exec -T postgres pg_isready -U treelot -d treelot >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "Compose PostgreSQL was not ready in time." >&2
    $COMPOSE logs postgres >&2 || true
    exit 1
  fi
  sleep 1
done

$COMPOSE exec -T postgres psql -U treelot -d postgres -v ON_ERROR_STOP=1 <<'SQL' >/dev/null
SELECT 'CREATE DATABASE treelot_test OWNER treelot'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'treelot_test')\gexec
SQL

deadline=$((SECONDS + 30))
until ping_db "$COMPOSE_URL"; do
  if (( SECONDS >= deadline )); then
    echo "Compose treelot_test on localhost:5433 was not reachable in time." >&2
    exit 1
  fi
  sleep 1
done

emit "$COMPOSE_URL"
