#!/usr/bin/env bash
# Diagnose local prerequisites and fail fast before acceptance reserves ports.
set -u

MODE="${1:-doctor}"
case "$MODE" in
  doctor|acceptance) ;;
  *)
    echo "Usage: $0 [doctor|acceptance]" >&2
    exit 2
    ;;
esac

OS="${PREFLIGHT_OS:-$(uname -s)}"
FAILURES=0
HOST_CHECK_CONTAINER=""
ACCEPTANCE_WEB_PORT="${ACCEPTANCE_WEB_PORT:-18080}"
ACCEPTANCE_PRODUCTION_PORT="${ACCEPTANCE_PRODUCTION_PORT:-18081}"
ACCEPTANCE_STUB_PORT="${ACCEPTANCE_STUB_PORT:-18090}"

cleanup() {
  if [[ -n "$HOST_CHECK_CONTAINER" ]]; then
    docker rm -f "$HOST_CHECK_CONTAINER" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

pass() {
  printf '✓ %s\n' "$1"
}

fail() {
  printf '✗ %s\n' "$1" >&2
  FAILURES=$((FAILURES + 1))
}

note() {
  printf '• %s\n' "$1"
}

require_command() {
  local command_name="$1"
  local label="$2"
  if command -v "$command_name" >/dev/null 2>&1; then
    pass "$label"
  else
    fail "$label is not installed"
  fi
}

check_git_hooks() {
  local hooks_path
  hooks_path="$(git config --local --get core.hooksPath 2>/dev/null || true)"
  if [[ "$hooks_path" != ".githooks" ]]; then
    fail "Tracked Git hooks are not installed; run make install-hooks"
    return
  fi

  local root
  if ! root="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    fail "Cannot locate the Git repository to inspect tracked hooks"
    return
  fi
  if [[ ! -x "$root/$hooks_path/pre-push" ]]; then
    fail "Tracked pre-push hook is missing or not executable; run make install-hooks"
    return
  fi

  pass "Tracked Git hooks are installed"
}

port_listener() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null
    return
  fi
  if command -v ss >/dev/null 2>&1; then
    ss -H -ltn "sport = :$port" 2>/dev/null
    return
  fi
  return 2
}

project_service_owns_port() {
  local service="$1"
  local container_port="$2"
  local host_port="$3"
  local container
  container="$(docker compose ps -q "$service" 2>/dev/null)"
  if [[ -z "$container" ]]; then
    return 1
  fi
  local published
  published="$(docker port "$container" "$container_port/tcp" 2>/dev/null)"
  [[ "$published" == *":$host_port" ]]
}

check_port() {
  local port="$1"
  local output
  output="$(port_listener "$port")"
  local status=$?
  case "$status" in
    0)
      if [[ "$MODE" == "doctor" && "$port" == "5433" ]] && project_service_owns_port postgres 5432 5433; then
        pass "Port 5433 is in use by project PostgreSQL"
      elif [[ "$MODE" == "doctor" && "$port" == "8080" ]] && project_service_owns_port web 8080 8080; then
        pass "Port 8080 is in use by project web"
      else
        fail "Port $port is in use"
        printf '%s\n' "$output" >&2
      fi
      ;;
    1)
      pass "Port $port is free"
      ;;
    *)
      fail "Cannot inspect port $port; install lsof (macOS) or ss (Linux)"
      ;;
  esac
}

check_docker_desktop_host_network() {
  local port="${PREFLIGHT_HOST_NETWORK_PORT:-18090}"
  local attempts="${PREFLIGHT_HOST_NETWORK_ATTEMPTS:-20}"
  HOST_CHECK_CONTAINER="treelot-host-network-check-$$"

  if ! docker run -d --rm --network host --name "$HOST_CHECK_CONTAINER" \
    busybox:1.36 sh -c \
    "mkdir -p /tmp/preflight && printf ok >/tmp/preflight/index.html && exec httpd -f -p $port -h /tmp/preflight" \
    >/dev/null 2>&1; then
    fail "Docker Desktop could not start a host-networked container. Enable host networking under Settings > Resources > Network."
    HOST_CHECK_CONTAINER=""
    return
  fi

  local attempt
  for ((attempt = 1; attempt <= attempts; attempt++)); do
    if curl --fail --silent --max-time 1 "http://127.0.0.1:$port/" >/dev/null 2>&1; then
      pass "Docker Desktop host networking is operational"
      docker rm -f "$HOST_CHECK_CONTAINER" >/dev/null 2>&1 || true
      HOST_CHECK_CONTAINER=""
      return
    fi
    sleep 0.1
  done

  fail "Docker Desktop host networking is unavailable. Enable host networking under Settings > Resources > Network, then apply and restart."
}

require_command go "Go is available"
require_command node "Node.js is available"
require_command curl "curl is available"
require_command docker "Docker CLI is available"

if [[ "$MODE" == "doctor" ]]; then
  require_command git "Git is available"
  if command -v git >/dev/null 2>&1; then
    check_git_hooks
  fi
fi

if command -v docker >/dev/null 2>&1; then
  if docker info >/dev/null 2>&1; then
    pass "Docker daemon is reachable"
  else
    fail "Docker daemon is not reachable"
  fi
  if docker compose version >/dev/null 2>&1; then
    pass "Docker Compose is available"
  else
    fail "Docker Compose is not available"
  fi
fi

ports=(5433 8080 8081 8090)
if [[ "$MODE" == "acceptance" ]]; then
  ports=(5433 "$ACCEPTANCE_WEB_PORT" "$ACCEPTANCE_PRODUCTION_PORT" "$ACCEPTANCE_STUB_PORT")
fi
for port in "${ports[@]}"; do
  check_port "$port"
done

if [[ "$OS" == "Darwin" && "$MODE" == "acceptance" && "$FAILURES" -eq 0 ]]; then
  check_docker_desktop_host_network
elif [[ "$OS" == "Darwin" ]]; then
  note "Run make acceptance-preflight to verify Docker Desktop host networking."
else
  pass "Native Docker host networking is supported on $OS"
fi

if (( FAILURES > 0 )); then
  printf '\nPreflight found %d problem(s).\n' "$FAILURES" >&2
  exit 1
fi

if [[ "$MODE" == "acceptance" ]]; then
  pass "Acceptance preflight passed"
else
  pass "Local development preflight passed"
fi
