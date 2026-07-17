#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [[ ! -f .env ]]; then
  echo "Missing $ROOT_DIR/.env" >&2
  exit 1
fi

# Load only application values. Parsing them instead of sourcing the file
# prevents punctuation in secrets from being interpreted by the shell.
while IFS= read -r line || [[ -n "$line" ]]; do
  line="${line%$'\r'}"
  case "$line" in
    "" | \#*) continue ;;
  esac

  key="${line%%=*}"
  value="${line#*=}"
  case "$key" in
    DATABASE_URL | JWT_SECRET | REDIS_PASSWORD | REDIS_URL | S3_ENDPOINT | S3_ACCESS_KEY | S3_SECRET_KEY | S3_BUCKET | S3_REGION | S3_USE_SSL | S3_FORCE_PATH_STYLE) export "$key=$value" ;;
  esac
done < .env

: "${DATABASE_URL:?DATABASE_URL must be set in .env}"
: "${JWT_SECRET:?JWT_SECRET must be set in .env}"
: "${REDIS_PASSWORD:?REDIS_PASSWORD must be set in .env}"
: "${REDIS_URL:?REDIS_URL must be set in .env}"
: "${S3_ACCESS_KEY:?S3_ACCESS_KEY must be set in .env}"
: "${S3_SECRET_KEY:?S3_SECRET_KEY must be set in .env}"
: "${S3_ENDPOINT:?S3_ENDPOINT must be set in .env}"

# The Compose app reaches PostgreSQL as `db`; host-run Air reaches the same
# published database through localhost.
case "$DATABASE_URL" in
  *@db:*) export DATABASE_URL="${DATABASE_URL/@db:/@127.0.0.1:}" ;;
esac
case "$REDIS_URL" in
  *@redis:*) export REDIS_URL="${REDIS_URL/@redis:/@127.0.0.1:}" ;;
esac
case "$S3_ENDPOINT" in
  http://seaweedfs:*) export S3_ENDPOINT="${S3_ENDPOINT/http:\/\/seaweedfs:/http:\/\/127.0.0.1:}" ;;
esac

backend_pid=""
frontend_pid=""

cleanup() {
  status=$?
  trap - EXIT INT TERM

  echo
  echo "Shutting down dev servers..."
  [[ -n "$backend_pid" ]] && kill "$backend_pid" 2>/dev/null || true
  [[ -n "$frontend_pid" ]] && kill "$frontend_pid" 2>/dev/null || true
  [[ -n "$backend_pid" ]] && wait "$backend_pid" 2>/dev/null || true
  [[ -n "$frontend_pid" ]] && wait "$frontend_pid" 2>/dev/null || true

  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "Starting PostgreSQL, Redis, and SeaweedFS..."
docker compose up -d --wait db redis seaweedfs

echo "Starting Fiber API on http://localhost:3000..."
(cd backend && exec air) &
backend_pid=$!

echo "Starting Vite on http://localhost:5173..."
(cd frontend && exec npm run dev) &
frontend_pid=$!

# Bash 3.2 has no `wait -n`, so monitor both processes and stop everything if
# either dev server exits.
status=0
while kill -0 "$backend_pid" 2>/dev/null && kill -0 "$frontend_pid" 2>/dev/null; do
  sleep 1
done

if ! kill -0 "$backend_pid" 2>/dev/null; then
  wait "$backend_pid" || status=$?
else
  wait "$frontend_pid" || status=$?
fi

exit "$status"
