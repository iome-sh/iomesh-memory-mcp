# Open-source readiness audit

**Maintainer process residual** — not a product spec and not a user guide. Visibility is already **public MIT**. **Public MIT ≠ Memory GA.** Flip is complete. Operators: [README.md](../README.md) · [SECURITY.md](../SECURITY.md) · [CONTRIBUTING.md](../CONTRIBUTING.md).

Checklist for the OSS **process bar** of **github.com/iome-sh/iomesh-memory-mcp**
(lean edge Memory MCP host) vs public **iomesh-tui** (binary product) and public
kernel **memory**. Visibility flip is **complete** (public MIT). Re-run before
each major release.

**Public-flip SSOT:** [docs/PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) ·
`make public-flip-readiness-gate` (offline greps; residual PASS ≠ public flip).

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Public** (MIT · flipped deliberately). Kernel (`github.com/iome-sh/memory`) is public first; this host is public. |
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
| HTTP mode auth residual documented (lean v1: loopback default + optional secret; unauthenticated when secret unset) | Pass |
| dual_write OFF residual documented | Pass |
| govulncheck in CI | Pass |
| Residual: private dep on `github.com/iome-sh/memory` until kernel public flip | **Resolved** — kernel is public; module fetch is public; no PAT required |

## Honesty locks (product narrative)

| Claim | Status |
|-------|--------|
| Edge host · **not product Memory GA** | Pass |
| Local-primary Palace path | Pass |
| dual_write OFF by default | Pass |
| Naming honesty: **iomesh-memory-mcp** | Pass |
| Qdrant/ONNX not required for default path | Pass |
| Path isolation `PALACE_ROOT/<tenant>/` ≠ cloud multi-tenant | Pass |
| Kernel public first (hard flip order) | Pass (documented · both public) |

## Open-source process artifacts (TUI binary parity)

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present |
| NOTICE | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING (+ **Public repository policy**, CI table, branch protection `ci-success`) | Present |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING (GoReleaser + cosign verify + honesty locks) | Present |
| PR template | Present |
| Issue templates + security contact + docs contact_link | Present |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| **release.yml** + **`.goreleaser.yaml`** (multi-arch · SBOM · cosign) | Present |
| Makefile `ci` / `check` / `vuln` / `fmt-check` / `release-snapshot` | Present |
| Dependabot (gomod + actions) | Present |
| Dockerfile + docker-compose | Present |
| README badges + honesty locks + quick start | Present |
| Edge dogfood SSOT + offline gate | Present ([EDGE_DOGFOOD.md](EDGE_DOGFOOD.md)) |
| Public-flip readiness SSOT + offline gate | Present ([PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md)) |

## Residual risks

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility public | **Pass** (deliberate flip complete) | Audit does **not** flip public |
| Private kernel dependency | **Resolved** (kernel public) | **kernel first**, then this host — both public |
| CI token residual while kernel private | **Resolved** | No `IOMESH_CI_PAT` / `GO_MODULE_TOKEN` required for this host; do not invent a token requirement |
| HTTP unauthenticated lean v1 | Residual | Loopback default + optional shared secret; still unauthenticated when secret unset; bind localhost / proxy |
| Path tenancy same-process | Residual | Documented |
| dual_write optional later | Residual | Interface not wired; default OFF |
| GHCR publish green | **Not claimed** | Optional deliberate act; do not invent |

## Maintainer actions (visibility already public)

See [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) post-flip steps. Summary:

1. Kernel `github.com/iome-sh/memory` is **public**  
2. This host is **public** (flip complete)  
3. Enable **Private vulnerability reporting**  
4. Branch protection on `main`: require PR + status check **`ci-success`**  
5. No private module CI secret required  
6. Optional: publish GHCR image as **`ghcr.io/iome-sh/iomesh-memory-mcp`** only (not invent green)  
7. Keep private ledger serials off the public PR/CHANGELOG surface per CONTRIBUTING policy  
8. Do **not** invent Memory GA or default dual_write ON  

## Out of scope for this host

- Multitenant hosted Palace / cloud Memory SLA  
- Mesh `X-IOMesh-Org` HTTP headers (mesh clients, not this host)  
- Billing / plan gates  
- Guarantees about third-party Qdrant / model hubs  
- Inventing GHCR publish green or visibility flip from residual alone  

## Audit verdict

| Dimension | Verdict |
|-----------|---------|
| Process bar vs iomesh-tui (binary product) | **Pass** — CONTRIBUTING public policy · release.yml · goreleaser · cosign · release-snapshot |
| Process bar vs memory | **Pass** (artifacts + CI spirit aligned) |
| Public-flip readiness residual (docs + offline gate) | **Pass** — readiness only |
| Residual private kernel dep | **Resolved** (kernel public) |
| Visibility public flip | **Done** — public MIT · residual PASS ≠ invent Memory GA · readiness ≠ invent flip (historical) |
| Product honesty | **Pass** |
| Lean extract | **Pass** |

**Overall:** **Public** (host + kernel) · dual_write OFF · not Memory GA · no GOPRIVATE required ·
Ready for deliberate public flip was the pre-flip verdict · readiness ≠ invent flip.
residual PASS ≠ public flip · dual_write OFF · not Memory GA · naming iomesh-memory-mcp · kernel first.


## Public import (post-flip)

```bash
# No GOPRIVATE / PAT:
go get github.com/iome-sh/iomesh-memory-mcp@main
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@main
```
