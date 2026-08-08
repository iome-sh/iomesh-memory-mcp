# Changelog

All notable changes to this project (lean edge Memory MCP host) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **s1474 / final private→public flip audit closeout (TUI binary parity)** (still private · residual PASS ≠ public flip):
  - CONTRIBUTING expanded to TUI parity: development setup (GOPRIVATE residual), coding standards, tests, security-sensitive changes, Issues, **Public repository policy**, PR + CI table + branch protection `ci-success`, MIT contribution clause
  - [`.goreleaser.yaml`](.goreleaser.yaml) + [`.github/workflows/release.yml`](.github/workflows/release.yml) (multi-arch · SBOM · keyless cosign) · `make release-snapshot`
  - RELEASING expanded (GoReleaser · cosign verify · kernel public prerequisite · honesty locks)
  - OPEN_SOURCE_AUDIT + PUBLIC_FLIP_READINESS final s1474 checklist/verdict
  - ISSUE_TEMPLATE docs contact_link · CI comments (IOMESH_CI_PAT while kernel private; optional after)
  - `ServerVersion` → linkable `var` default `v0.1.0` (ldflags from GoReleaser / make build)
  - **Does not** flip visibility · invent GHCR publish green · Memory GA · dual_write ON · full platform sidecar parity · live dogfood green
- **s1468 / Option A M4 public-flip readiness** (residual-honest offline SSOT):
  - [docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md) — flip order (kernel first, then this host) · pre-flight · post-flip residual steps
  - `scripts/public_flip_readiness_gate.sh` + `make public-flip-readiness-gate` — offline file greps only (no visibility flip / docker / gcloud)
  - OPEN_SOURCE_AUDIT + README continuum stamp **s1468**; EDGE_DOGFOOD M4 pointer
  - **Does not** flip visibility · invent GHCR publish green · Memory GA · dual_write ON · full platform sidecar parity · live dogfood green
- **s1462 / Option A M3 edge dogfood** (residual-honest offline SSOT):
  - [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — operator checklist (build · stdio · HTTP healthz/`/mcp` · local compose · tool honesty)
  - `scripts/edge_dogfood_gate.sh` + `make edge-dogfood-gate` — offline file greps only (no docker daemon / server / gcloud)
  - README M3 section + continuum stamp **s1462**; `make help`; compose comments for local image honesty
  - **Does not** invent live dogfood green · public flip · GHCR publish · Memory GA · dual_write ON · full platform sidecar parity

### Honesty

- dual_write **OFF** · not product Memory GA · still private · residual PASS ≠ live dogfood / public flip · readiness ≠ invent flip · no aion import · naming **iomesh-memory-mcp** · kernel public first · M4 flip deliberate later (s1474 final audit closeout only)

## [0.1.0-s1457] — 2026-08-08

### Added

- Lean edge MCP host scaffold (**s1457** / Option A M2):
  - Binary **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`)
  - stdio + streamable HTTP (`MEMORY_MCP_HTTP_ADDR`) with `GET /healthz`
  - Tools: `memory_ingest_turn`, `memory_retrieve`, `memory_search_semantic`,
    `memory_list`, `memory_compact_status`, `memory_facts_as_of`
  - Path-based tenant layout: `filepath.Join(palaceRoot, tenant)`
  - Depends on `github.com/iome-sh/memory` + `modelcontextprotocol/go-sdk` only (no aion)
  - TUI-grade OSS process bar: LICENSE, NOTICE, SECURITY, community docs,
    RELEASING, CHANGELOG, OPEN_SOURCE_AUDIT, Makefile, CI, Dependabot, Dockerfile, compose
  - **Repository remains private** until a deliberate visibility flip
  - dual_write **OFF** · not product Memory GA · aion broker stays private · no default Qdrant/ONNX requirement
