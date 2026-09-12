#!/usr/bin/env bash
# edge_dogfood_gate.sh — offline residual gate for M3 edge dogfood (s1462).
#
# File greps only. No docker daemon, no long-running server, no gcloud.
# residual PASS ≠ live dogfood · ≠ public flip
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

# file must exist and must NOT contain pattern
forbid_needle() {
  local f="$1"
  local pat="$2"
  local label="${3:-$pat}"
  if [[ ! -f "$f" ]]; then
    fail_msg "forbid skip (missing file): $f ($label)"
    return
  fi
  if grep -E -q -- "$pat" "$f"; then
    fail_msg "forbid hit: $f ← $label"
  else
    pass "forbid: $f ↛ $label"
  fi
}

echo "== edge_dogfood_gate (s1462 M3) offline — $ROOT =="
echo "   public · PALACE_ROOT · no docker daemon required"
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
echo "-- docs/EDGE_DOGFOOD.md checklist --"
need_needle "docs/EDGE_DOGFOOD.md" "s1462" "serial s1462"
need_needle "docs/EDGE_DOGFOOD.md" "PALACE_ROOT" "PALACE_ROOT"
need_needle "docs/EDGE_DOGFOOD.md" "dual_write" "dual_write healthz field"
need_needle "docs/EDGE_DOGFOOD.md" "public|still private" "public (or historical still private)"
need_needle "docs/EDGE_DOGFOOD.md" "residual PASS ≠ live dogfood|residual PASS != live dogfood" "residual ≠ live dogfood"
need_needle "docs/EDGE_DOGFOOD.md" "residual PASS ≠ public flip|residual PASS != public flip" "residual ≠ public flip"
need_needle "docs/EDGE_DOGFOOD.md" "full platform sidecar parity|platform sidecar parity" "no full platform sidecar parity"
need_needle "docs/EDGE_DOGFOOD.md" "does not import private control-plane/broker packages|private control-plane / broker" "no private control-plane/broker import"
need_needle "docs/EDGE_DOGFOOD.md" "iomesh-memory-mcp" "naming iomesh-memory-mcp"
forbid_needle "docs/EDGE_DOGFOOD.md" 'Memory Ops Pack' "no Memory Ops Pack SKU"
forbid_needle "docs/EDGE_DOGFOOD.md" '\$88|~\$88' "no ~\$88 mesh rate"
forbid_needle "docs/EDGE_DOGFOOD.md" '\$119|~\$119' "no ~\$119 pack rate"
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
need_needle "docs/EDGE_DOGFOOD.md" "memory_ingest_turn|ingest" "ingest tool"
need_needle "docs/EDGE_DOGFOOD.md" "memory_retrieve|retrieve" "retrieve tool"
need_needle "docs/EDGE_DOGFOOD.md" "memory_list|list" "list tool"
need_needle "docs/EDGE_DOGFOOD.md" "memory_compact_status|compact_status" "compact_status"
need_needle "docs/EDGE_DOGFOOD.md" "M4" "M4 later"
need_needle "docs/EDGE_DOGFOOD.md" "s1463|TUI" "peer TUI s1463 mention"
need_needle "docs/EDGE_DOGFOOD.md" "s1464" "peer private-plane residual s1464"
need_needle "docs/EDGE_DOGFOOD.md" "edge-dogfood-gate|edge_dogfood_gate" "gate target"

echo
echo "-- README / Makefile / CHANGELOG --"
# Public README is consumer-facing; continuum lives in docs/EDGE_DOGFOOD.md
need_needle "README.md" "EDGE_DOGFOOD|edge-dogfood-gate|edge dogfood" "README edge dogfood docs pointer"
need_needle "README.md" "iomesh-memory-mcp" "README naming"
need_needle "README.md" "PALACE_ROOT" "README PALACE_ROOT"
need_needle "docs/EDGE_DOGFOOD.md" "s1462" "EDGE_DOGFOOD continuum s1462"
need_needle "docs/EDGE_DOGFOOD.md" "dual_write" "EDGE_DOGFOOD dual_write field"
need_needle "Makefile" "edge-dogfood-gate" "Makefile edge-dogfood-gate"
need_needle "Makefile" "edge_dogfood_gate\\.sh" "Makefile script path"
need_needle "CHANGELOG.md" "s1462" "CHANGELOG s1462"
need_needle "CHANGELOG.md" "edge dogfood|EDGE_DOGFOOD|M3" "CHANGELOG M3 edge dogfood"

echo
echo "-- docker-compose / Dockerfile --"
need_needle "docker-compose.yml" "iomesh-memory-mcp:local" "compose local image"
need_needle "docker-compose.yml" "PALACE_ROOT" "compose PALACE_ROOT"
need_needle "docker-compose.yml" "M3|edge dogfood|dogfood" "compose M3 / dogfood comment"
need_needle "docker-compose.yml" "healthz|/healthz" "compose healthz note"
need_needle "Dockerfile" "iomesh-memory-mcp" "Dockerfile binary name"

echo
echo "-- host layout --"
need_needle "cmd/iomesh-memory-mcp/main.go" "PALACE_ROOT" "main PALACE_ROOT"
need_needle "internal/mcphost/host.go" "iomesh-memory-mcp" "ServerName iomesh-memory-mcp"
need_needle "internal/mcphost/http.go" "healthz|/healthz" "HTTP healthz"
need_needle "internal/mcphost/http.go" "dual_write" "HTTP dual_write field"
need_needle "internal/mcphost/tools.go" "dual_write" "tools dual_write field"

echo
echo "-- product-codename residual forbid --"
# Character-class pattern so this script is not itself a \\b token hit.
codename_pat='[Aa][Ii][Oo][Nn]'
user_facing=(
  CONTRIBUTING.md
  NOTICE
  docs/EDGE_DOGFOOD.md
  docs/EDGE_DOGFOOD_EVIDENCE.md
  CHANGELOG.md
  Dockerfile
  docker-compose.yml
  .github/pull_request_template.md
  .github/ISSUE_TEMPLATE/feature_request.yml
  .github/workflows/ci.yml
  cmd/iomesh-memory-mcp/main.go
  internal/mcphost/host.go
)
for f in "${user_facing[@]}"; do
  forbid_needle "$f" "$codename_pat" "no product-codename residual"
done
# Tree-wide: no leftover product-codename token (pattern is split above).
# Exclude tokenizer/vocab fixtures: WordPiece lists contain coincidental
# letter runs and are not product-plane names.
if git grep -I -n -E -- "$codename_pat" -- . \
    ':!**/testdata/**' ':!testdata' ':!**/tokenizer.json' ':!**/vocab.txt' >/dev/null; then
  git grep -I -n -E -- "$codename_pat" -- . \
    ':!**/testdata/**' ':!testdata' ':!**/tokenizer.json' ':!**/vocab.txt' || true
  fail_msg "product-codename residual (git grep)"
else
  pass "tree has no product-codename residual"
fi

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
echo "  residual PASS ≠ live dogfood · ≠ public flip"
exit 0
