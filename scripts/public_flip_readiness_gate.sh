#!/usr/bin/env bash
# public_flip_readiness_gate.sh — offline residual gate for M4 public-flip readiness (s1468).
#
# File greps only. No docker daemon, no long-running server, no gcloud, no Settings mutation.
# residual PASS ≠ public flip · ≠ invent Memory GA · ≠ live dogfood green · dual_write OFF · still private
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

echo "== public_flip_readiness_gate (s1468 M4) offline — $ROOT =="
echo "   dual_write OFF · not Memory GA · still private · kernel first · residual PASS ≠ public flip"
echo

# --- required surfaces ---
need_file "docs/PUBLIC_FLIP_READINESS.md"
need_file "docs/OPEN_SOURCE_AUDIT.md"
need_file "docs/EDGE_DOGFOOD.md"
need_file "LICENSE"
need_file "SECURITY.md"
need_file "README.md"
need_file "Makefile"
need_file "CHANGELOG.md"
need_file "scripts/public_flip_readiness_gate.sh"
need_file "scripts/edge_dogfood_gate.sh"

echo
echo "-- docs/PUBLIC_FLIP_READINESS.md honesty + order + serial --"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1468" "serial s1468"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "dual_write OFF|dual_write \*\*OFF\*\*" "dual_write OFF"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "not Memory GA" "not Memory GA"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "still private" "still private"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "residual PASS ≠ public flip|residual PASS != public flip" "residual ≠ public flip"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "full platform sidecar parity|platform sidecar parity" "no full platform sidecar parity"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "no aion import" "no aion import"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "iomesh-memory-mcp" "naming iomesh-memory-mcp"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "kernel first|Kernel first|memory.*public first|wait for.*memory" "kernel first"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "github.com/iome-sh/memory" "kernel module path"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "OPEN_SOURCE_AUDIT|re-audit|re-run" "OPEN_SOURCE_AUDIT re-run"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "Private dep|private dep|GOPRIVATE|module token" "private dep residual"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "CI token|GO_MODULE_TOKEN|module-token" "CI token residual"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "Post-flip|post-flip" "post-flip steps"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "offline dogfood|edge-dogfood-gate|live dogfood" "offline dogfood ≠ live invent"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "compose PASS ≠ public registry|compose PASS != public registry|public registry" "compose ≠ public registry"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "aion stays private|aion.*private" "aion stays private"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1467" "peer memory s1467"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1469" "peer TUI s1469"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1470" "peer aion residual s1470"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1465" "free-floor s1465"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1466" "lag s1466"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1471" "free-floor peer s1471"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "s1473" "free eng s1473+"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "public-flip-readiness-gate|public_flip_readiness_gate" "gate target"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "ghcr.io/iome-sh/iomesh-memory-mcp" "GHCR image name honesty"
need_needle "docs/PUBLIC_FLIP_READINESS.md" "Does not flip|does not flip|not flip|still private" "does not flip on this serial"

echo
echo "-- LICENSE / SECURITY / audit process bar --"
need_needle "LICENSE" "MIT|Permission is hereby granted" "LICENSE MIT-ish"
need_needle "SECURITY.md" "security@|vulnerability|Private vulnerability" "SECURITY reporting"
need_needle "docs/OPEN_SOURCE_AUDIT.md" "Still private|still private" "audit still private"
need_needle "docs/OPEN_SOURCE_AUDIT.md" "dual_write" "audit dual_write"
need_needle "docs/OPEN_SOURCE_AUDIT.md" "not product Memory GA|not Memory GA|Memory GA" "audit not Memory GA"
need_needle "docs/OPEN_SOURCE_AUDIT.md" "kernel first|M4|public flip" "audit M4 / kernel order"
need_needle "docs/OPEN_SOURCE_AUDIT.md" "s1468|PUBLIC_FLIP_READINESS|public-flip-readiness" "audit links s1468 readiness"

echo
echo "-- README / Makefile / CHANGELOG --"
need_needle "README.md" "PUBLIC_FLIP_READINESS|public-flip-readiness|M4 public-flip|M4 public flip" "README M4 readiness"
need_needle "README.md" "public-flip-readiness-gate" "README public-flip-readiness-gate"
need_needle "README.md" "s1468" "README continuum s1468"
need_needle "README.md" "dual_write OFF|dual_write \*\*OFF\*\*" "README dual_write OFF"
need_needle "README.md" "not product Memory GA|not Memory GA" "README not Memory GA"
need_needle "README.md" "still private" "README still private"
need_needle "README.md" "iomesh-memory-mcp" "README naming"
need_needle "Makefile" "public-flip-readiness-gate" "Makefile public-flip-readiness-gate"
need_needle "Makefile" "public_flip_readiness_gate\\.sh" "Makefile script path"
need_needle "CHANGELOG.md" "s1468" "CHANGELOG s1468"
need_needle "CHANGELOG.md" "public flip|PUBLIC_FLIP|M4" "CHANGELOG M4 public flip readiness"

echo
echo "-- EDGE_DOGFOOD M4 pointer (optional residual link) --"
need_needle "docs/EDGE_DOGFOOD.md" "M4|PUBLIC_FLIP|public flip|public-flip-readiness" "EDGE_DOGFOOD M4 link residual"

# Self-check: this gate is offline greps only — no docker/gcloud/gh visibility mutations as commands.
if grep -E -q '^[[:space:]]*(docker|gcloud)[[:space:]]' "$0"; then
  fail_msg "gate script must not invoke docker/gcloud commands"
else
  pass "gate script does not invoke docker/gcloud"
fi
if grep -E -q 'gh[[:space:]]+repo[[:space:]]+edit|visibility.*public|Change visibility' "$0" && ! grep -E -q 'file greps|does not|≠ public flip|residual PASS' "$0"; then
  # Allow documentary mentions in comments; fail only if script looks like it mutates visibility.
  :
fi
# Hard fail if script tries to run gh repo visibility change as a bare command line.
if grep -E -q '^[[:space:]]*gh[[:space:]]+repo[[:space:]]+(edit|create)' "$0"; then
  fail_msg "gate script must not invoke gh repo edit/create"
else
  pass "gate script does not invoke gh repo edit/create"
fi

echo
if [[ "$fail" -ne 0 ]]; then
  echo "RESULT: FAIL (public_flip_readiness_gate s1468)"
  exit 1
fi
echo "RESULT: PASS (public_flip_readiness_gate s1468 — offline residual only)"
echo "  residual PASS ≠ public flip · ≠ Memory GA · ≠ live dogfood invent · dual_write OFF · still private · kernel first"
exit 0
