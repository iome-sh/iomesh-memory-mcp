#!/usr/bin/env bash
# ci_boot_smoke.sh — start the built binary in HTTP mode, check GET /healthz
# and GET /ready, then stop it.
#
# Usage: scripts/ci_boot_smoke.sh <path-to-binary>
#
# Fails if either endpoint is not 200 within BOOT_SMOKE_TIMEOUT seconds, or if
# the /healthz body contains the test tenant value or a tenant/org key.
set -euo pipefail

BIN="${1:?usage: $0 <path-to-binary>}"
PORT="${BOOT_SMOKE_PORT:-18765}"
TIMEOUT="${BOOT_SMOKE_TIMEOUT:-30}"
TENANT="ci-smoke-tenant-q7x2"
BASE="http://127.0.0.1:${PORT}"

work="$(mktemp -d)"
log="${work}/server.log"
pid=""

cleanup() {
  if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
    kill -TERM "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf "$work"
}
trap cleanup EXIT

fail() {
  echo "::error::boot smoke: $1"
  echo "--- server log ---"
  cat "$log" 2>/dev/null || true
  exit 1
}

"$BIN" \
  -http-addr "127.0.0.1:${PORT}" \
  -palace-root "${work}/palaces" \
  -tenant "$TENANT" >"$log" 2>&1 &
pid=$!

# wait_200 <path> <outfile>: poll until 200 or timeout.
wait_200() {
  local path="$1" out="$2" code="" deadline=$((SECONDS + TIMEOUT))
  while (( SECONDS < deadline )); do
    if ! kill -0 "$pid" 2>/dev/null; then
      fail "process exited before ${path} returned 200"
    fi
    code="$(curl -sS -o "$out" -w '%{http_code}' --max-time 2 "${BASE}${path}" 2>/dev/null || true)"
    if [[ "$code" == "200" ]]; then
      echo "PASS  GET ${path} 200"
      return 0
    fi
    sleep 0.5
  done
  fail "GET ${path} not 200 within ${TIMEOUT}s (last status: ${code:-none})"
}

wait_200 /healthz "${work}/healthz.json"
wait_200 /ready "${work}/ready.json"

if grep -qF "$TENANT" "${work}/healthz.json"; then
  fail "/healthz body contains the test tenant value"
fi
if ! jq -e 'type == "object" and (has("tenant") or has("org") or has("organization") | not)' \
  "${work}/healthz.json" >/dev/null; then
  fail "/healthz body is not a JSON object or has a tenant/org key"
fi
echo "PASS  /healthz has no tenant or org data"

kill -TERM "$pid"
if ! wait "$pid"; then
  pid=""
  fail "process did not exit cleanly on SIGTERM"
fi
pid=""
echo "PASS  process stopped"
