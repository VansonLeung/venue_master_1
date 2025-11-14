#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
COMPOSE_DIR="$ROOT_DIR/codes"
ENV_FILE="$COMPOSE_DIR/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE. Copy codes/.env.example to codes/.env" >&2
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

cleanup() {
  (cd "$COMPOSE_DIR" && docker compose down -v >/dev/null 2>&1) || true
}
trap cleanup EXIT

(cd "$COMPOSE_DIR" && docker compose up -d --build)

wait_for() {
  local name=$1
  local url=$2
  echo "Waiting for $name ($url)..."
  for _ in {1..60}; do
    if curl -sf "$url" >/dev/null; then
      echo "$name is ready"
      return 0
    fi
    sleep 2
  done
  echo "Timed out waiting for $name" >&2
  exit 1
}

wait_for "api-gateway" "http://localhost:8080/healthz"
wait_for "auth-service" "http://localhost:8081/healthz"
wait_for "booking-service" "http://localhost:8083/healthz"

(cd "$COMPOSE_DIR" && \
  E2E_GATEWAY_URL=${E2E_GATEWAY_URL:-http://localhost:8080} \
  E2E_AUTH_URL=${E2E_AUTH_URL:-http://localhost:8081} \
  go test -tags e2e -v ./test/e2e)
