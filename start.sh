#!/usr/bin/env bash
set -Eeuo pipefail

FRONTEND_PORT="${FRONTEND_PORT:-8080}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yaml}"

cleanup() {
  echo
  echo "Stopping tunnel and containers..."
  [[ -n "${TUNNEL_PID:-}" ]] && kill "$TUNNEL_PID" 2>/dev/null || true
  docker compose -f "$COMPOSE_FILE" down
}
trap cleanup EXIT INT TERM

command -v docker >/dev/null 2>&1 || {
  echo "docker is not installed"
  exit 1
}

docker compose version >/dev/null 2>&1 || {
  echo "docker compose plugin is not installed"
  exit 1
}

command -v cloudflared >/dev/null 2>&1 || {
  echo "cloudflared is not installed"
  echo "Install: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/"
  exit 1
}

echo "Starting Docker Compose..."
docker compose -f "$COMPOSE_FILE" up -d --build

echo "Waiting for frontend on http://localhost:${FRONTEND_PORT} ..."
for i in {1..60}; do
  if curl -fsS "http://localhost:${FRONTEND_PORT}" >/dev/null 2>&1; then
    break
  fi

  if [[ "$i" -eq 60 ]]; then
    echo "Frontend did not become reachable on port ${FRONTEND_PORT}"
    docker compose -f "$COMPOSE_FILE" ps
    exit 1
  fi

  sleep 2
done

echo "Opening temporary Cloudflare tunnel..."
LOG_FILE="$(mktemp)"

cloudflared tunnel --url "http://localhost:${FRONTEND_PORT}" 2>&1 | tee "$LOG_FILE" &
TUNNEL_PID="$!"

echo "Waiting for temporary URL..."
for i in {1..60}; do
  URL="$(grep -oE 'https://[-a-zA-Z0-9]+\.trycloudflare\.com' "$LOG_FILE" | head -n 1 || true)"

  if [[ -n "$URL" ]]; then
    echo
    echo "Temporary public URL:"
    echo "$URL"
    echo
    echo "Press Ctrl+C to stop."
    wait "$TUNNEL_PID"
    exit 0
  fi

  sleep 1
done

echo "Could not detect Cloudflare URL."
echo "Tunnel logs:"
cat "$LOG_FILE"
exit 1