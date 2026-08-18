# Changelog

All notable changes to this project (lean edge Memory MCP host) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **CLI `-preflight` (#28):** constructs the host and prints the same honesty JSON as `GET /healthz` (`status`, `service`, `dual_write=off`, `not_memory_ga`, `embeddings`, `qdrant=off`, `version`, `tools`, `tool_names`), then exits without listening or running stdio MCP. Registration ≠ `tools/list` ≠ ingest. No hosted palace probe. dual_write OFF · not Memory GA.
- **Tenant single path segment (#27):** when `tenant` is provided (tool input or `-tenant` / `MEMORY_TENANT`), it must be a single path segment (reject `.`, `..`, separators). Invalid tool tenant returns an `IsError` result; invalid default tenant fails process start. Omitted/empty tool tenant still uses the configured default. Same-process path isolation only · dual_write OFF · not Memory GA.

### Fixed
- **EDGE_DOGFOOD rates (#36):** drop priced Memory Ops Pack / ~$119 / ~$88 language. Mesh stays optional for pull/retain without a SKU. Gate forbids those needles. dual_write OFF · not Memory GA.
- **Invalid RFC3339 time fields (#26):** `parseOptionalTime` / `parseTimeOrNow` now return an error on non-empty unparsable input instead of treating it as unset or now. `memory_ingest_turn`, `memory_retrieve`, `memory_list`, `memory_facts_as_of`, `memory_related`, and `memory_supersede_entity` fail closed. Empty still means now / omitted. dual_write OFF · not Memory GA.
- **Post-flip honesty (#30):** CONTRIBUTING / Makefile help / compose no longer say the repo is private. Docker and `.env.example` no longer present a GitHub token as required (kernel + host are public). Compose publishes `127.0.0.1:8080` for local dogfood. `make tidy` matches public CI (no `GOPRIVATE`). dual_write OFF · not Memory GA.

### Changed
- **Tool copy (#29):** list/retrieve/search/facts_as_of/related/compact_status descriptions and the README tool table say local palace FS, read/list only, **does not ingest**. `tools/list` / `healthz.tool_names` remain discovery, not ingest. Write tools stay local FS only. dual_write OFF · not Memory GA.
- **go-sdk v1.7.0:** bump `github.com/modelcontextprotocol/go-sdk` 1.6.1 → 1.7.0 (protocol `2026-07-28` + legacy `2025-11-25` negotiate). Streamable HTTP already sets `Stateless: true`, which is required for the new revision on HTTP; stdio/legacy clients still negotiate down. Not a protocol-only-new-clients cut. dual_write OFF · not Memory GA.

### Changed
- Pin `github.com/iome-sh/memory` to public main tip `e1ffb9d` (`v1.5.8-0.20260816062432-e1ffb9db873e`) so ONNX retrieve gets keyword-first + expanded haystack (memory #46) and list uses durable event-time snapshot (memory #47). Ingest `valid_from` + Write errors (memory #48). Hash still omits `QueryVec`. dual_write OFF · not Memory GA. Do not invent Edge Memory GA / first `v*` tag.
- Go toolchain pin `go 1.26.6` so CI `govulncheck` is clean on stdlib GO-2026-5972 / GO-2026-5026 (fixed in go1.26.6).

### Fixed
- **`memory_retrieve` / `memory_search_semantic` (#21):** do not inject hash `QueryVec`. SHA-256 unit vectors skipped the kernel keyword path and dropped exact tokens past `Limit`. ONNX still passes a query vector. dual_write OFF · not Memory GA.

### Added
- **Host tests (kernel pin lock):** same-tenant session isolation on `memory_retrieve` / `memory_list` (shared token must not leak across `session_id`; empty session unfiltered). After ingest, `memory_facts_as_of` sees fact children when atoms extract (kernel #48 `valid_from`). `memory_list` after `New()` on the same palace root still lists the needle (kernel #47 durable snapshot; hash). Tests-only · go-sdk not bumped · dual_write OFF · not Memory GA.
- **`/healthz` tools surface:** residual-honest `tools` (count) and `tool_names` for the compile-time lean registered tools. Not a live MCP `tools/list` stamp. Historical s1509 TUI attach `tools=6` at tip `f46afe2` stays contemporaneous evidence — do not restamp as live forever-green. dual_write OFF · not Memory GA.
- **`memory_list` hyphen needle rank 1:** after hyphen ingest, `handleList{Query: needle, Limit: 5}` must hit rank 1. `TestRetrieveHashKeepsHyphenNeedle` kept.
- **`memory_write` (#20):** durable fact ingest via kernel `Write`. Optional `entity_key` stamps `entity:` tags and defaults to `WriteAndSupersede`. `dual_write` OFF · `audited=false` · not Memory GA.
- **`memory_related` / `memory_supersede_entity` (#17):** lean maps to kernel `MultiHopRetrieve` and `SupersedeEntityFacts`. Hash `SeedQuery` does not inject `QueryVec`. HITL stays at the client. dual_write OFF · not Memory GA.
- **Other MCP clients (#18):** README stdio `mcp.json` + streamable HTTP URL attach (Cursor / generic). No TUI required. Not Memory GA.
- **Install pin honesty (#19):** document `@main` until the first annotated `v*` GitHub Release. `@latest` is a pseudo-version today. No tag cut in this change. Not Memory GA.

### Changed
- **Public OSS:** host + kernel are public — CI drops `GOPRIVATE` / private PAT requirement; pin `github.com/iome-sh/memory` to public main tip; docs visibility honesty.
- **s1492 / M5 signing matrix residual:** [`.github/workflows/release.yml`](.github/workflows/release.yml) drops `GOPRIVATE` + private module PAT residual; public `github.com/iome-sh/memory` fetch only (aligned with public CI). GoReleaser + Syft SBOM + keyless cosign kept.
- **s1500 / Edge Memory GA candidacy (E3–E5 docs):** [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) aligns honesty with Edge Memory GA candidacy (local-primary; residual PASS ≠ invent Edge Memory GA declared; dual_write OFF; not bare Memory GA; not hosted Memory GA); public modules; retires stale private-module install residual.

### Added

- **Optional ONNX embeddings (s1525)** — when `MEMORY_ONNX_MODEL_PATH` is set, the lean host constructs Palace stores with kernel ONNX embeddings (else hash). `/healthz` reports `embeddings` (`hash`|`onnx`) and `qdrant=off` (Qdrant not wired into lean search). dual_write OFF · not Memory GA · optional path ≠ invent platform GPU palace.


- **s1509 / E4 TUI client attach residual dogfood evidence** (public binary host residual-honest):
  - [docs/EDGE_DOGFOOD_EVIDENCE.md](docs/EDGE_DOGFOOD_EVIDENCE.md) — stamp **2026-08-09T06:23:34Z** · MCP tip `f46afe2` · TUI tip `6b3958a` · healthz ok on `:18081` (`dual_write=off`, `not_memory_ga=true`) · TUI `iomesh mcp --connect` → **connected=1** **tools=6** (`memory_ingest_turn`, `memory_retrieve`, `memory_search_semantic`, `memory_list`, `memory_compact_status`, `memory_facts_as_of`)
  - [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — E4 section peer link to s1509 client attach evidence
  - Honesty: residual PASS ≠ invent Edge Memory GA declared · residual PASS ≠ invent forever product green · dual_write OFF · not bare Memory GA · not hosted Memory GA · **attach + tools/list ≠ invent Edge Memory GA** · **attach + tools/list ≠ invent forever green full product dogfood**
- **s1504 / E4 local residual dogfood evidence** (public binary host residual-honest):
  - [docs/EDGE_DOGFOOD_EVIDENCE.md](docs/EDGE_DOGFOOD_EVIDENCE.md) — contemporaneous stamp **2026-08-09T06:06:22Z** · tip `f46afe2` · unit `go test ./internal/mcphost/` ok · HTTP `/healthz` ok (`dual_write=off`, `not_memory_ga=true`)
  - [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — E4 section + continuum link to evidence log
  - Honesty: residual PASS ≠ invent Edge Memory GA declared · residual PASS ≠ invent forever product green · dual_write OFF · not bare Memory GA · not hosted Memory GA · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool round-trip
- **s1500 / E3 install matrix · E4 operator dogfood runbook · E5 support/version policy** (public binary host residual-honest):
  - [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — E3 install matrix (stdio · HTTP · Docker Compose · TUI attach) · **E4 operator runbook** (build → stdio health → `memory_ingest_turn` → `memory_retrieve` → `memory_list` → `memory_facts_as_of` → `memory_compact_status` → optional HTTP /healthz)
  - [RELEASING.md](RELEASING.md) — **Support / version policy (E5)**: latest GitHub Release tag · GoReleaser + SBOM + keyless cosign · pin for production · snapshot ≠ production release
  - [SUPPORT.md](SUPPORT.md) — issues · security · related memory kernel · E5 pointers
  - README Documentation table pointers only
  - Honesty: residual PASS ≠ live dogfood green · residual PASS ≠ invent forever-green signed releases · residual PASS ≠ invent Edge Memory GA · dual_write OFF · not Memory GA
- **s1492 / Option A M5 signing/matrix tip** (public binary host residual-honest):
  - [RELEASING.md](RELEASING.md) **M5 signing / matrix** section: tag → release.yml → GoReleaser → archives + checksums + SBOM + cosign keyless · `make release-snapshot` dry-run
  - Honesty: tip ≠ invent successful public tag release shipped · residual PASS ≠ invent forever-green signed releases · dual_write OFF · not Memory GA · naming **iomesh-memory-mcp** · kernel public prerequisite met · no auto-tag · aion private
  - Gate needles lightly updated for public release-path honesty
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

- dual_write **OFF** · not product Memory GA · host + kernel public · residual PASS ≠ live dogfood / invent forever-green signed releases · residual PASS ≠ invent Edge Memory GA · readiness ≠ invent flip · tip ≠ invent tag release shipped · no aion import · naming **iomesh-memory-mcp** · kernel public prerequisite met · M5 packaging residual (s1492) ≠ invent M5 complete · s1500 E3–E5 docs ≠ invent Edge Memory GA declared · s1504 local evidence ≠ invent Edge Memory GA / forever product green · s1509 client attach ≠ invent Edge Memory GA / forever green full product dogfood · healthz `tools` / `tool_names` = compile-time lean surface ≠ invent live TUI attach restamp / forever-green `tools=N` · no auto-tag

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
