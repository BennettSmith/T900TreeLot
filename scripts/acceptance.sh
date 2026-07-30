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
BASE_URL="${ACCEPTANCE_BASE_URL:-http://127.0.0.1:8080}"
TEST_CONTROL_KEY="${ACCEPTANCE_TEST_CONTROL_KEY:-acceptance-test-control-key}"

cleanup() {
  if [[ -z "${ACCEPTANCE_KEEP:-}" ]]; then
    $COMPOSE --profile acceptance down >/dev/null 2>&1 || true
    $DOCKER rm -f treelot-acceptance-web treelot-acceptance-worker treelot-acceptance-stub >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

$COMPOSE --profile acceptance down >/dev/null 2>&1 || true
$DOCKER rm -f treelot-acceptance-web treelot-acceptance-worker treelot-acceptance-stub >/dev/null 2>&1 || true

$DOCKER build -t "$IMAGE" .
$COMPOSE up -d postgres

echo "Waiting for PostgreSQL on host port 5433..."
deadline=$((SECONDS + 90))
until $COMPOSE exec -T postgres pg_isready -U treelot -d treelot >/dev/null 2>&1 \
  && $DOCKER run --rm --network host postgres:16-alpine \
    pg_isready -h 127.0.0.1 -p 5433 -U treelot -d treelot >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "PostgreSQL was not ready in time." >&2
    $COMPOSE logs postgres || true
    exit 1
  fi
  sleep 1
done
# Brief settle after TCP accept to avoid reset-during-startup races.
sleep 2

common_env=(
  -e APP_ENV=acceptance
  -e DATABASE_URL="$DB_URL"
  -e TREE_LOT_TIME_ZONE=America/Los_Angeles
  -e PUBLIC_BASE_URL=https://treelot.test
  -e SESSION_KEY=0123456789abcdef0123456789abcdef
  -e GROUPS_IO_ENABLED=false
  -e TEST_CONTROL_KEY="$TEST_CONTROL_KEY"
)

$DOCKER run --rm --network host --entrypoint /app/migrate "${common_env[@]}" "$IMAGE" up

$DOCKER run -d --name treelot-acceptance-web --network host --entrypoint /app/web \
  "${common_env[@]}" -e PORT=8080 "$IMAGE"
$DOCKER run -d --name treelot-acceptance-worker --network host --entrypoint /app/worker \
  "${common_env[@]}" "$IMAGE"
$DOCKER run -d --name treelot-acceptance-stub --network host --entrypoint /app/web \
  -e APP_ENV=development \
  -e DATABASE_URL="$DB_URL" \
  -e TREE_LOT_TIME_ZONE=UTC \
  -e PUBLIC_BASE_URL=http://127.0.0.1:8090 \
  -e SESSION_KEY=0123456789abcdef0123456789abcdef \
  -e GROUPS_IO_ENABLED=false \
  -e PORT=8090 \
  "$IMAGE"

echo "Waiting for acceptance web readiness..."
deadline=$((SECONDS + 60))
until curl --fail --silent "$BASE_URL/health/ready" >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "acceptance-web was not ready in time." >&2
    $DOCKER logs treelot-acceptance-web || true
    exit 1
  fi
  sleep 1
done

ACCEPTANCE_BASE_URL="$BASE_URL" \
ACCEPTANCE_TEST_CONTROL_KEY="$TEST_CONTROL_KEY" \
ACCEPTANCE_DATABASE_URL="$DB_URL" \
  go test -tags=acceptance ./acceptance/... -count=1 -v
