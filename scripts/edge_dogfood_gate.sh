#!/usr/bin/env bash
# edge_dogfood_gate.sh — offline residual gate for M3 edge dogfood (s1462).
#
# File greps only. No docker daemon, no long-running server, no gcloud.
# residual PASS ≠ live dogfood · ≠ public flip · ≠ invent Memory GA · dual_write OFF
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail=0
pass() { printf 'PASS  %s\n' "$1"; }
fail_msg() { printf 'FAIL  %s\n' "$1"; fail=1; }

need_file() {
  local f="$1"
  if [[ -f "$f" ]]; then
    pass "file present: $f"
  else
    fail_msg "missing file: $f"
  fi
}

need_dir() {
  local d="$1"
  if [[ -d "$d" ]]; then
    pass "dir present: $d"
  else
    fail_msg "missing dir: $d"
  fi
}

# needle file pattern  — file must exist and contain pattern (grep -E -q)
need_needle() {
  local f="$1"
  local pat="$2"
  local label="${3:-$pat}"
  if [[ ! -f "$f" ]]; then
    fail_msg "needle skip (missing file): $f ($label)"
    return
  fi
  if grep -E -q -- "$pat" "$f"; then
    pass "needle: $f ← $label"
  else
    fail_msg "needle missing: $f ← $label"
  fi
}

echo "== edge_dogfood_gate (s1462 M3) offline — $ROOT =="
echo "   dual_write OFF · not Memory GA · still private · no docker daemon required"
echo

# --- required surfaces ---
need_file "docs/EDGE_DOGFOOD.md"
need_file "docker-compose.yml"
need_file "Dockerfile"
need_file "README.md"
need_file "Makefile"
need_file "CHANGELOG.md"
need_file "scripts/edge_dogfood_gate.sh"
need_dir  "cmd/iomesh-memory-mcp"
need_dir  "internal/mcphost"
need_file "cmd/iomesh-memory-mcp/main.go"

echo
echo "-- docs/EDGE_DOGFOOD.md honesty + checklist --"
need_needle "docs/EDGE_DOGFOOD.md" "s1462" "serial s1462"
need_needle "docs/EDGE_DOGFOOD.md" "dual_write OFF|dual_write \*\*OFF\*\*|dual_write OFF" "dual_write OFF"
need_needle "docs/EDGE_DOGFOOD.md" "not Memory GA" "not Memory GA"
need_needle "docs/EDGE_DOGFOOD.md" "still private" "still private"
need_needle "docs/EDGE_DOGFOOD.md" "residual PASS ≠ live dogfood|residual PASS != live dogfood" "residual ≠ live dogfood"
need_needle "docs/EDGE_DOGFOOD.md" "residual PASS ≠ public flip|residual PASS != public flip" "residual ≠ public flip"
need_needle "docs/EDGE_DOGFOOD.md" "full platform sidecar parity|platform sidecar parity" "no full platform sidecar parity"
need_needle "docs/EDGE_DOGFOOD.md" "no aion import" "no aion import"
need_needle "docs/EDGE_DOGFOOD.md" "iomesh-memory-mcp" "naming iomesh-memory-mcp"
need_needle "docs/EDGE_DOGFOOD.md" '\$88|~\$88|~\\$88|~\$88' "rates ~\$88"
need_needle "docs/EDGE_DOGFOOD.md" '\$119|~\$119' "rates ~\$119"
need_needle "docs/EDGE_DOGFOOD.md" "Palace sunset|hosted Palace sunset" "Palace sunset"
need_needle "docs/EDGE_DOGFOOD.md" "mesh optional" "mesh optional for pull"
need_needle "docs/EDGE_DOGFOOD.md" "open boxes stay open" "open boxes stay open"
need_needle "docs/EDGE_DOGFOOD.md" "make build|Build binary" "build path"
need_needle "docs/EDGE_DOGFOOD.md" "stdio" "stdio attach"
need_needle "docs/EDGE_DOGFOOD.md" "healthz" "healthz"
need_needle "docs/EDGE_DOGFOOD.md" "/mcp" "/mcp path"
need_needle "docs/EDGE_DOGFOOD.md" "docker compose|Docker Compose" "compose path"
need_needle "docs/EDGE_DOGFOOD.md" "iomesh-memory-mcp:local" "local image only"
need_needle "docs/EDGE_DOGFOOD.md" "compose PASS ≠ public registry|compose PASS != public registry" "compose ≠ public registry"
need_needle "docs/EDGE_DOGFOOD.md" "memory_ingest_turn|ingest" "ingest tool honesty"
need_needle "docs/EDGE_DOGFOOD.md" "memory_retrieve|retrieve" "retrieve tool honesty"
need_needle "docs/EDGE_DOGFOOD.md" "memory_list|list" "list tool honesty"
need_needle "docs/EDGE_DOGFOOD.md" "memory_compact_status|compact_status" "compact_status honesty"
need_needle "docs/EDGE_DOGFOOD.md" "M4" "M4 later"
need_needle "docs/EDGE_DOGFOOD.md" "s1463|TUI" "peer TUI s1463 mention"
need_needle "docs/EDGE_DOGFOOD.md" "s1464|aion residual" "peer aion residual s1464"
need_needle "docs/EDGE_DOGFOOD.md" "edge-dogfood-gate|edge_dogfood_gate" "gate target"

echo
echo "-- README / Makefile / CHANGELOG --"
need_needle "README.md" "EDGE_DOGFOOD|edge dogfood|M3 edge dogfood" "README M3 edge dogfood"
need_needle "README.md" "edge-dogfood-gate" "README edge-dogfood-gate"
need_needle "README.md" "s1462" "README continuum s1462"
need_needle "README.md" "dual_write OFF|dual_write \*\*OFF\*\*" "README dual_write OFF"
need_needle "README.md" "not product Memory GA|not Memory GA" "README not Memory GA"
need_needle "README.md" "iomesh-memory-mcp" "README naming"
need_needle "Makefile" "edge-dogfood-gate" "Makefile edge-dogfood-gate"
need_needle "Makefile" "edge_dogfood_gate\\.sh" "Makefile script path"
need_needle "CHANGELOG.md" "s1462" "CHANGELOG s1462"
need_needle "CHANGELOG.md" "edge dogfood|EDGE_DOGFOOD|M3" "CHANGELOG M3 edge dogfood"

echo
echo "-- docker-compose / Dockerfile honesty --"
need_needle "docker-compose.yml" "iomesh-memory-mcp:local" "compose local image"
need_needle "docker-compose.yml" "dual_write OFF|dual_write" "compose dual_write"
need_needle "docker-compose.yml" "not Memory GA|not_memory_ga|Memory GA" "compose not Memory GA"
need_needle "docker-compose.yml" "M3|edge dogfood|dogfood" "compose M3 / dogfood comment"
need_needle "docker-compose.yml" "healthz|/healthz" "compose healthz note"
need_needle "Dockerfile" "iomesh-memory-mcp" "Dockerfile binary name"
need_needle "Dockerfile" "dual_write OFF|dual_write" "Dockerfile dual_write"
need_needle "Dockerfile" "not Memory GA|Memory GA" "Dockerfile not Memory GA"

echo
echo "-- host layout + dual_write residual in code --"
need_needle "cmd/iomesh-memory-mcp/main.go" "dual_write|not_memory_ga|not Memory GA" "main honesty"
need_needle "internal/mcphost/host.go" "iomesh-memory-mcp" "ServerName iomesh-memory-mcp"
need_needle "internal/mcphost/http.go" "healthz|/healthz" "HTTP healthz"
need_needle "internal/mcphost/http.go" "dual_write" "HTTP dual_write field"
need_needle "internal/mcphost/tools.go" "dual_write" "tools dual_write"

# Self-check: this gate is offline greps only — no docker/gcloud invocations as commands.
# (Mentions of "docker" in comments/strings are fine; bare command lines are not.)
if grep -E -q '^[[:space:]]*(docker|gcloud)[[:space:]]' "$0"; then
  fail_msg "gate script must not invoke docker/gcloud commands"
else
  pass "gate script does not invoke docker/gcloud"
fi

echo
if [[ "$fail" -ne 0 ]]; then
  echo "RESULT: FAIL (edge_dogfood_gate s1462)"
  exit 1
fi
echo "RESULT: PASS (edge_dogfood_gate s1462 — offline residual only)"
echo "  residual PASS ≠ live dogfood · ≠ public flip · ≠ Memory GA · dual_write OFF"
exit 0
