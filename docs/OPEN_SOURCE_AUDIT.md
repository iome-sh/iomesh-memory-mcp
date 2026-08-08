# Open-source readiness audit

Checklist for bringing **github.com/iome-sh/iomesh-memory-mcp** (lean edge Memory MCP host)
to the same OSS **process bar** as public **iomesh-tui** and private kernel **memory** (s1452).
Re-run before any deliberate visibility flip and before each major release.

**Serial stamp:** free eng concurrent **s1457+** after free-floor **s1455** · lag **s1456** ·
peers TUI s1458 · aion residual s1459 (mention only) · free-floor peer **s1460**.

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Still private** — do **not** flip public on this serial. Deliberate future flip only after re-audit (after kernel public flip in Option A M4). |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |

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

## Residual risks

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility still private | **Pass (intentional)** | s1457 does not flip public |
| Private kernel dependency | Partial | M4 public flip kernel first, then this host |
| HTTP unauthenticated lean v1 | Residual | Documented; bind localhost / proxy |
| Path tenancy same-process | Residual | Documented |
| dual_write optional later | Residual | Interface not wired; default OFF |

## Maintainer actions **after** going public (future — not this PR)

1. GitHub → Settings → Change visibility → Public (**deliberate**, after kernel public)  
2. Enable **Private vulnerability reporting**  
3. Branch protection on `main`: require PR + status check **`ci-success`**  
4. Publish GHCR image as **`ghcr.io/iome-sh/iomesh-memory-mcp`** only  
5. Do **not** invent Memory GA or default dual_write ON  

## Out of scope for this host

- Multitenant hosted Palace / cloud Memory SLA  
- Private aion broker / control plane / billing / plan gates  
- Embed/recmem workers, rqlite plan gate, Cloud Run residual product path  
- Guarantees about third-party Qdrant / model hubs  

## Audit verdict (s1457)

| Dimension | Verdict |
|-----------|---------|
| Process bar vs memory s1452 / iomesh-tui | **Pass** (artifacts + CI spirit aligned) |
| Visibility public-ready flip | **Not done** — still private by design |
| Product honesty | **Pass** |
| Lean extract (no aion) | **Pass** |

**Overall:** Ready for **private** OSS process bar + lean MCP dogfood. Public flip remains a separate, deliberate decision after re-audit (Option A M4).
