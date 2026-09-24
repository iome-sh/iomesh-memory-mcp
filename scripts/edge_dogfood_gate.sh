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
echo "-- README / Makefile / CHANGELOG --"
need_needle "README.md" "iomesh-memory-mcp" "README naming"
need_needle "README.md" "PALACE_ROOT" "README PALACE_ROOT"
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
