# Open-source readiness audit

Checklist for bringing **github.com/iome-sh/iomesh-memory-mcp** (lean edge Memory MCP host)
to the same OSS **process bar** as public **iomesh-tui** (binary product) and private kernel
**memory** (s1452). Re-run before any deliberate visibility flip and before each major release.

**Serial stamp:** free eng **s1474** final private→public flip audit closeout (TUI parity) ·
prior M4 readiness residual **s1468** · free eng concurrent **s1467+** after free-floor
**s1465** · lag **s1466** · peers memory s1467 · TUI s1469 · aion residual s1470
(mention only) · free-floor peer **s1471** · free eng **s1473+**.

**M4 readiness SSOT:** [docs/PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) ·
`make public-flip-readiness-gate` (offline greps; residual PASS ≠ public flip).

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Still private** — do **not** flip public on **s1474**. Deliberate M4 flip only after re-audit and **kernel (`github.com/iome-sh/memory`) public first**. |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |
| public-flip-readiness residual | Pass (docs + offline gate; **not** a visibility flip) |
| residual PASS ≠ public flip | Pass (explicit non-claim) |

## Security

| Check | Status |
|-------|--------|
| No committed API keys / private keys / `.env` secrets | Pass (`.env.example` only) |
| Local Palace FS treated as user data in SECURITY.md | Pass |
| Path-based tenant **not** claimed as cloud multi-tenant isolation | Pass |
| HTTP mode auth residual documented (lean v1: none) | Pass |
| dual_write OFF residual documented | Pass |
| govulncheck in CI | Pass |
| Residual: private dep on `github.com/iome-sh/memory` until kernel public flip | **Partial** — CI needs module token / org access while both private; **after kernel public**, module fetch is public and PAT optional |

## Honesty locks (product narrative)

| Claim | Status |
|-------|--------|
| Edge host · **not product Memory GA** | Pass |
| Local-primary Palace path | Pass |
| dual_write OFF by default | Pass |
| Naming honesty: **iomesh-memory-mcp** (not aion-memory-mcp) | Pass |
| No import of private aion packages | Pass |
| aion broker / CP stays private | Pass |
| Qdrant/ONNX not required for default path | Pass |
| Option A: open edge later; public flip separate deliberate act | Pass |
| Kernel public first (hard flip order) | Pass (documented) |

## Open-source process artifacts (TUI binary parity)

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present |
| NOTICE | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING (+ **Public repository policy**, CI table, branch protection `ci-success`) | Present (s1474 TUI parity) |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING (GoReleaser + cosign verify + honesty locks) | Present (s1474) |
| PR template | Present |
| Issue templates + security contact + docs contact_link | Present (s1474) |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| **release.yml** + **`.goreleaser.yaml`** (multi-arch · SBOM · cosign) | Present (s1474) |
| Makefile `ci` / `check` / `vuln` / `fmt-check` / `release-snapshot` | Present (s1474) |
| Dependabot (gomod + actions) | Present |
| Dockerfile + docker-compose | Present |
| README badges + honesty locks + quick start | Present |
| M3 edge dogfood SSOT + offline gate | Present ([EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) · s1462) |
| M4 public-flip readiness SSOT + offline gate | Present ([PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) · s1468 + s1474 closeout) |

## Residual risks

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility still private | **Pass (intentional)** | s1474 final audit does **not** flip public |
| Private kernel dependency | **Partial** until kernel public | M4: **kernel first**, then this host |
| CI token residual while kernel private | **Partial** | `IOMESH_CI_PAT` / `GO_MODULE_TOKEN` / org access for module fetch; optional after kernel public |
| HTTP unauthenticated lean v1 | Residual | Documented; bind localhost / proxy |
| Path tenancy same-process | Residual | Documented |
| dual_write optional later | Residual | Interface not wired; default OFF |
| GHCR publish green | **Not claimed** | Optional deliberate act; do not invent |

## Maintainer actions **after** going public (future — not this PR / not s1474)

See [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) post-flip steps. Summary:

1. Confirm kernel `github.com/iome-sh/memory` is **public** first  
2. GitHub → Settings → Change visibility → Public (**deliberate**)  
3. Enable **Private vulnerability reporting**  
4. Branch protection on `main`: require PR + status check **`ci-success`**  
5. Remove/relax private module CI secret requirement after kernel public  
6. Optional: publish GHCR image as **`ghcr.io/iome-sh/iomesh-memory-mcp`** only (not invent green)  
7. Strip private ledger serials from public PR/CHANGELOG surface per CONTRIBUTING policy  
8. Do **not** invent Memory GA or default dual_write ON  

## Out of scope for this host

- Multitenant hosted Palace / cloud Memory SLA  
- Private aion broker / control plane / billing / plan gates  
- Embed/recmem workers, rqlite plan gate, Cloud Run residual product path  
- Guarantees about third-party Qdrant / model hubs  
- Inventing GHCR publish green or visibility flip from residual alone  

## Audit verdict (s1474 final TUI-parity closeout)

| Dimension | Verdict |
|-----------|---------|
| Process bar vs iomesh-tui (binary product) | **Pass** — CONTRIBUTING public policy · release.yml · goreleaser · cosign · release-snapshot |
| Process bar vs memory s1452 | **Pass** (artifacts + CI spirit aligned) |
| M4 readiness residual (docs + offline gate) | **Pass** — readiness only |
| Residual private kernel dep | **Partial** until kernel public |
| Visibility public-ready flip | **Not done** — still private by design · readiness ≠ invent flip |
| Product honesty | **Pass** |
| Lean extract (no aion) | **Pass** |

**Overall:** **Ready for deliberate public flip after kernel is public** · still private ·
readiness ≠ invent flip. residual PASS ≠ public flip · dual_write OFF · not Memory GA ·
no aion import · naming iomesh-memory-mcp · kernel first.
