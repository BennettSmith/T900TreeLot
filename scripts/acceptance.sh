#!/usr/bin/env bash
# Runs foundation acceptance specs against the production image using host networking.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DOCKER="${DOCKER:-docker}"
if ! $DOCKER info >/dev/null 2>&1; then
  DOCKER="sudo docker"
fi
COMPOSE="$DOCKER compose"
IMAGE="${IMAGE:-treelot:local}"

DB_URL="${ACCEPTANCE_DATABASE_URL:-postgres://treelot:treelot@127.0.0.1:5433/treelot?sslmode=disable}"
UNMIGRATED_DB_URL="${ACCEPTANCE_UNMIGRATED_DATABASE_URL:-postgres://treelot:treelot@127.0.0.1:5433/treelot_unmigrated?sslmode=disable}"
BASE_URL="${ACCEPTANCE_BASE_URL:-http://127.0.0.1:8080}"
PRODUCTION_BASE_URL="${ACCEPTANCE_PRODUCTION_BASE_URL:-http://127.0.0.1:8081}"
STUB_BASE_URL="${ACCEPTANCE_STUB_BASE_URL:-http://127.0.0.1:8090}"
TEST_CONTROL_KEY="${ACCEPTANCE_TEST_CONTROL_KEY:-acceptance-test-control-key}"

cleanup() {
  if [[ -z "${ACCEPTANCE_KEEP:-}" ]]; then
    $COMPOSE --profile acceptance down >/dev/null 2>&1 || true
    $DOCKER rm -f \
      treelot-acceptance-web \
      treelot-acceptance-worker \
      treelot-acceptance-stub \
      treelot-production-web \
      >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

$COMPOSE --profile acceptance down >/dev/null 2>&1 || true
$DOCKER rm -f \
  treelot-acceptance-web \
  treelot-acceptance-worker \
  treelot-acceptance-stub \
  treelot-production-web \
  >/dev/null 2>&1 || true

bash ./scripts/preflight.sh acceptance

$DOCKER build -t "$IMAGE" .
$COMPOSE up -d postgres

echo "Waiting for PostgreSQL on host port 5433..."
deadline=$((SECONDS + 90))
until $COMPOSE exec -T postgres pg_isready -U treelot -d treelot >/dev/null 2>&1 \
  && $DOCKER run --rm --network host postgres:16-alpine \
    pg_isready -h 127.0.0.1 -p 5433 -U treelot -d treelot >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "PostgreSQL was not ready in time." >&2
    if [[ "$(uname -s)" == "Darwin" ]]; then
      echo "Verify Docker Desktop host networking is enabled under Settings > Resources > Network." >&2
    fi
    $COMPOSE logs postgres || true
    exit 1
  fi
  sleep 1
done
sleep 2

echo "Preparing unmigrated database..."
$COMPOSE exec -T postgres psql -U treelot -d postgres -v ON_ERROR_STOP=1 <<'SQL'
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'treelot_unmigrated' AND pid <> pg_backend_pid();
DROP DATABASE IF EXISTS treelot_unmigrated;
CREATE DATABASE treelot_unmigrated OWNER treelot;
SQL

common_env=(
  -e TREE_LOT_TIME_ZONE=America/Los_Angeles
  -e SESSION_KEY=0123456789abcdef0123456789abcdef
  -e BOOTSTRAP_ENROLLMENT_TOKEN=acceptance-bootstrap-token-0001
  -e AUTH_RATE_LIMIT_MAX=20
  -e GROUPS_IO_ENABLED=false
)

$DOCKER run --rm --network host --entrypoint /app/migrate \
  "${common_env[@]}" \
  -e APP_ENV=acceptance \
  -e DATABASE_URL="$DB_URL" \
  -e PUBLIC_BASE_URL=https://treelot.test \
  -e TEST_CONTROL_KEY="$TEST_CONTROL_KEY" \
  "$IMAGE" up

$DOCKER run -d --name treelot-acceptance-web --network host --entrypoint /app/web \
  "${common_env[@]}" \
  -e APP_ENV=acceptance \
  -e PORT=8080 \
  -e DATABASE_URL="$DB_URL" \
  -e PUBLIC_BASE_URL=https://treelot.test \
  -e TEST_CONTROL_KEY="$TEST_CONTROL_KEY" \
  "$IMAGE"

$DOCKER run -d --name treelot-acceptance-worker --network host --entrypoint /app/worker \
  "${common_env[@]}" \
  -e APP_ENV=acceptance \
  -e DATABASE_URL="$DB_URL" \
  -e PUBLIC_BASE_URL=https://treelot.test \
  -e TEST_CONTROL_KEY="$TEST_CONTROL_KEY" \
  "$IMAGE"

$DOCKER run -d --name treelot-production-web --network host --entrypoint /app/web \
  "${common_env[@]}" \
  -e APP_ENV=production \
  -e PORT=8081 \
  -e DATABASE_URL="$DB_URL" \
  -e PUBLIC_BASE_URL=https://treelot.troop900livermore.org \
  "$IMAGE"

$DOCKER run -d --name treelot-acceptance-stub --network host --entrypoint /app/provider-stubs \
  -e PORT=8090 \
  "$IMAGE"

echo "Waiting for acceptance and production web readiness..."
deadline=$((SECONDS + 60))
until curl --fail --silent "$BASE_URL/health/ready" >/dev/null 2>&1 \
  && curl --fail --silent "$PRODUCTION_BASE_URL/health/ready" >/dev/null 2>&1 \
  && curl --fail --silent "$STUB_BASE_URL/health/live" >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "acceptance stack was not ready in time." >&2
    $DOCKER logs treelot-acceptance-web || true
    $DOCKER logs treelot-production-web || true
    $DOCKER logs treelot-acceptance-stub || true
    exit 1
  fi
  sleep 1
done

ACCEPTANCE_BASE_URL="$BASE_URL" \
ACCEPTANCE_PRODUCTION_BASE_URL="$PRODUCTION_BASE_URL" \
ACCEPTANCE_STUB_BASE_URL="$STUB_BASE_URL" \
ACCEPTANCE_TEST_CONTROL_KEY="$TEST_CONTROL_KEY" \
ACCEPTANCE_DATABASE_URL="$DB_URL" \
ACCEPTANCE_UNMIGRATED_DATABASE_URL="$UNMIGRATED_DB_URL" \
ACCEPTANCE_IMAGE="$IMAGE" \
ACCEPTANCE_DOCKER="$DOCKER" \
  go test -tags=acceptance ./acceptance/... -count=1 -v
