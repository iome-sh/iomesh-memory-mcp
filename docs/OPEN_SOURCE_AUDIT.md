# Open-source readiness audit

Checklist for bringing **github.com/iome-sh/iomesh-memory-mcp** (lean edge Memory MCP host)
to the same OSS **process bar** as public **iomesh-tui** and private kernel **memory** (s1452).
Re-run before any deliberate visibility flip and before each major release.

**Serial stamp:** free eng concurrent **s1467+** after free-floor **s1465** · lag **s1466** ·
M4 readiness residual **s1468** · peers memory s1467 · TUI s1469 · aion residual s1470
(mention only) · free-floor peer **s1471** · free eng **s1473+**.

**M4 readiness SSOT:** [docs/PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) ·
`make public-flip-readiness-gate` (offline greps; residual PASS ≠ public flip).

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Still private** — do **not** flip public on **s1468**. Deliberate M4 flip only after re-audit and **kernel (`github.com/iome-sh/memory`) public first**. |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |
| public-flip-readiness residual | Pass (docs + offline gate on s1468; **not** a visibility flip) |

## Security

| Check | Status |
|-------|--------|
| No committed API keys / private keys / `.env` secrets | Pass (`.env.example` only) |
| Local Palace FS treated as user data in SECURITY.md | Pass |
| Path-based tenant **not** claimed as cloud multi-tenant isolation | Pass |
| HTTP mode auth residual documented (lean v1: none) | Pass |
| dual_write OFF residual documented | Pass |
| govulncheck in CI | Pass (s1457) |
| Residual: private dep on `github.com/iome-sh/memory` until kernel public flip | **Partial** — CI needs module token / org access while both private |

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

## Open-source process artifacts

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present |
| NOTICE | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING | Present |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING | Present |
| PR template | Present |
| Issue templates + security contact | Present |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| Dependabot (gomod + actions) | Present |
| Makefile `ci` / `check` / `vuln` / `fmt-check` | Present |
| Dockerfile + docker-compose | Present |
| README badges + honesty locks + quick start | Present |
| M3 edge dogfood SSOT + offline gate | Present ([EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) · s1462) |
| M4 public-flip readiness SSOT + offline gate | Present ([PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) · s1468) |

## Residual risks

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility still private | **Pass (intentional)** | s1468 readiness residual does **not** flip public |
| Private kernel dependency | Partial | M4: **kernel first**, then this host; private dep residual until kernel public |
| CI token residual while kernel private | Partial | `GO_MODULE_TOKEN` / org access for module fetch |
| HTTP unauthenticated lean v1 | Residual | Documented; bind localhost / proxy |
| Path tenancy same-process | Residual | Documented |
| dual_write optional later | Residual | Interface not wired; default OFF |

## Maintainer actions **after** going public (future — not this PR / not s1468)

See [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) post-flip steps. Summary:

1. Confirm kernel `github.com/iome-sh/memory` is **public** first  
2. GitHub → Settings → Change visibility → Public (**deliberate**)  
3. Enable **Private vulnerability reporting**  
4. Branch protection on `main`: require PR + status check **`ci-success`**  
5. Publish GHCR image as **`ghcr.io/iome-sh/iomesh-memory-mcp`** only (optional; not invent green on readiness gate)  
6. Do **not** invent Memory GA or default dual_write ON  

## Out of scope for this host

- Multitenant hosted Palace / cloud Memory SLA  
- Private aion broker / control plane / billing / plan gates  
- Embed/recmem workers, rqlite plan gate, Cloud Run residual product path  
- Guarantees about third-party Qdrant / model hubs  

## Audit verdict (s1468 refresh)

| Dimension | Verdict |
|-----------|---------|
| Process bar vs memory s1452 / iomesh-tui | **Pass** (artifacts + CI spirit aligned) |
| M4 readiness residual (docs + offline gate) | **Pass** — readiness only |
| Visibility public-ready flip | **Not done** — still private by design |
| Product honesty | **Pass** |
| Lean extract (no aion) | **Pass** |

**Overall:** Ready for **private** OSS process bar + lean MCP dogfood + **M4 readiness residual**.
Public flip remains a separate, deliberate decision after re-audit and **kernel public first**
(Option A M4). residual PASS ≠ public flip.
