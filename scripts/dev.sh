#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  local exit_code=$?
  trap - EXIT INT TERM

  for pid in "$BACKEND_PID" "$FRONTEND_PID"; do
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done

  for pid in "$BACKEND_PID" "$FRONTEND_PID"; do
    if [[ -n "$pid" ]]; then
      wait "$pid" 2>/dev/null || true
    fi
  done
  exit "$exit_code"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

(
  cd "$ROOT_DIR/backend"
  exec go run ./cmd/server
) &
BACKEND_PID=$!

(
  cd "$ROOT_DIR"
  exec node frontend/dev-server.mjs
) &
FRONTEND_PID=$!

printf 'Backend:  http://localhost:8080\n'
printf 'Frontend: http://localhost:3000\n'
printf 'Tekan Ctrl+C untuk menghentikan kedua server.\n'

set +e
wait -n "$BACKEND_PID" "$FRONTEND_PID"
server_status=$?
set -e

if (( server_status == 0 )); then
  printf 'Salah satu server berhenti.\n' >&2
else
  printf 'Salah satu server berhenti dengan status %d.\n' "$server_status" >&2
fi
exit "$server_status"
