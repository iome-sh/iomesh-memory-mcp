# Changelog

All notable changes to this project (lean edge Memory MCP host) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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

- dual_write **OFF** · not product Memory GA · still private · residual PASS ≠ live dogfood / public flip · no aion import · naming **iomesh-memory-mcp** · kernel public first · M4 flip deliberate later (readiness only on s1468)

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
