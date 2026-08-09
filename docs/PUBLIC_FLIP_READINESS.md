# Public-flip readiness / post-flip SSOT (iomesh-memory-mcp)

**Status: FLIP COMPLETE (public)** — `github.com/iome-sh/memory` is **public**; this host is **public**.  
CI no longer requires `GOPRIVATE` / `IOMESH_CI_PAT`. **dual_write OFF** · **not Memory GA** · **aion stays private**.

**Serial stamps (historical continuum):** s1468 M4 readiness · s1474 final TUI-parity audit · free eng concurrent **s1467+** after free-floor **s1465** · lag **s1466** · peers memory **s1467** · TUI **s1469** · aion residual **s1470** · free-floor peer **s1471** · free eng **s1473+**.

> **Hard non-claims:** Public OSS ≠ invent Memory GA · dual_write OFF · residual PASS ≠ live dogfood invent · residual PASS ≠ full platform sidecar parity · compose PASS ≠ public registry · offline dogfood tip ≠ invent live dogfood green · readiness residual history ≠ invent GA · **aion stays private**.

## Flip order (Option A — completed)

| Order | Module | Status |
|-------|--------|--------|
| **1. First** | `github.com/iome-sh/memory` | **Public** |
| **2. Then** | `github.com/iome-sh/iomesh-memory-mcp` (this host) | **Public** |
| Stay private | aion broker / CP | **aion stays private** |

Kernel first, then this host — **completed deliberately**.

## Public import (no private env)

```bash
# Do NOT set GOPRIVATE=github.com/iome-sh/* for these modules
go get github.com/iome-sh/memory@main
go get github.com/iome-sh/iomesh-memory-mcp@main
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@main
```

Historical **Private dep residual** / **CI token residual** (`GOPRIVATE` · module token · `IOMESH_CI_PAT` · `GO_MODULE_TOKEN`) applied only while the kernel was private — **resolved** after kernel public.

## Offline gate

```bash
make public-flip-readiness-gate
# scripts/public_flip_readiness_gate.sh
```

Gate is offline greps only · residual PASS ≠ invent Memory GA · does not flip visibility (already public).

## Pre-flight archive (what we required before flip)

1. Re-run OPEN_SOURCE_AUDIT re-audit  
2. CONTRIBUTING **Public repository policy**  
3. GoReleaser + `.github/workflows/release.yml`  
4. Branch protection **ci-success** · Private vulnerability reporting · homepage/topics/delete-branch-on-merge  
5. dual_write OFF · not Memory GA · no aion import · naming **iomesh-memory-mcp**  
6. edge-dogfood-gate offline dogfood SSOT  
7. GHCR `ghcr.io/iome-sh/iomesh-memory-mcp` optional · **not invent green** until publish  

## Post-flip steps (ongoing)

- Keep branch protection: require PR + **ci-success**  
- Private vulnerability reporting ON  
- Optional CodeQL / Dependabot security updates  
- Optional release tags + GHCR publish (not invent green)  
- Ready for deliberate public flip was the pre-flip verdict; **flip is done** · readiness ≠ invent flip was historical honesty  

## Honesty locks

| Lock | Status |
|------|--------|
| dual_write OFF | Pass |
| not Memory GA | Pass |
| public (host + kernel) | Pass |
| residual PASS ≠ public flip (historical residual) | Pass as process honesty |
| no aion import | Pass |
| aion stays private | Pass |
| compose PASS ≠ public registry | Pass |
| offline dogfood ≠ live invent | Pass |
| full platform sidecar parity | Not claimed |

## Related

- [OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md)  
- [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md)  
- aion org checklist `docs/operations/github-private-to-public-checklist.md` (private monorepo)  
