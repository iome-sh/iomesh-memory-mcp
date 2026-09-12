# Changelog

All notable changes to this project (lean edge Memory MCP host) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Public-docs hygiene:** `docs/PUBLIC_FLIP_READINESS.md` and `docs/OPEN_SOURCE_AUDIT.md` marked maintainer residuals (flip complete; not operator how-tos). README / CONTRIBUTING qualify those files. EDGE_DOGFOOD serials remain historical engineering pins, not a product ledger. dual_write OFF · not Memory GA · catalog ≠ Connected · public MIT ≠ Memory GA.
- **Public-docs hygiene:** `docs/EDGE_DOGFOOD_EVIDENCE.md` marked a maintainer residual (local evidence log; not operator how-to; not a product claim). residual PASS ≠ live dogfood. dual_write OFF · not Memory GA · catalog ≠ Connected.

## [0.4.1] — 2026-09-12

Kernel pin **v1.5.11** + optional ONNX `PersistEmbeddings` (#72). dual_write OFF · not Memory GA · persist default off · hash never stored.

### Added
- **Optional ONNX `PersistEmbeddings` (#72):** env `MEMORY_PERSIST_EMBEDDINGS` (`1`/`true`/`on`/`yes`, case-insensitive) default **off**. `PalaceConfig.PersistEmbeddings` is true only when embeddings are **onnx** and the env is on. Hash never persists (kernel #45). `GET /healthz` reports `persist_embeddings` (`off`|`on`). Does not require Qdrant/usearch. Default write path unchanged. dual_write OFF · not Memory GA · catalog ≠ Connected.

### Changed
- **Kernel pin (#72):** `github.com/iome-sh/memory` annotated **`v1.5.10`** → annotated **`v1.5.11`** (optional ONNX vector persist, default off). dual_write OFF · not Memory GA · catalog ≠ Connected.

## [0.4.0] — 2026-09-12

Optional HITL `memory_extract_facts` (#70). Compatible with [iomesh-tui **v1.3.4+**](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.4) `/memory extract`. dual_write OFF · not Memory GA · catalog ≠ connected.

### Added
- **`memory_extract_facts` (#70):** optional HITL extract-after-persist. Required `tenant` (omit fail-closes) and `memory_id` (parent turn already on disk); optional `facts` (HITL strings; else `palace.ExtractAtomicFacts` on a copy). Writes `turn_fact` children (`TierSemantic`, inherited tags, `provenance.source_step=mcp_memory_extract_facts`, inherited `source_hint` / private — never invent mesh, `parent_ids=[parent]`, `valid_from` stamp). Does **not** rewrite or delete the parent. Not called from `handleIngestTurn` (ingest already writes caller `ExtractedFacts` / kernel auto-extract children; extract is not a PalaceStore write-gate and never fails `memory_ingest_turn`). dual_write always off. Structural extract, not NLP / not Memory GA. Advertised in `leanToolNames` / `GET /healthz` `tool_names`. Ingest schema unchanged; TUI **v1.3.4** ignores unknown tools.

### Changed
- **Release hygiene (#69):** README install pin **v0.3.2** (was leftover v0.3.0); kernel pin language **v1.5.10** to match `go.mod` (was leftover v1.5.8). Default `ServerVersion` **v0.3.2**. CHANGELOG records already-cut **v0.3.1** / **v0.3.2** instead of leaving those waves under Unreleased.
- **Single-writer contract (#69):** README + SECURITY.md: one host process per palace root. Multi-process writers on a shared root remain unsupported (product contract). HTTP loopback default, optional shared secret, unauth-if-unset still documented.

## [0.3.2] — 2026-09-10

Cite-both companion to [iomesh-tui v1.3.3](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.3). dual_write OFF · not Memory GA · catalog ≠ connected.

### Fixed
- **`ops_digest_export` receipt selection (#66 / #67):** default receipts no longer drop in-window `source_hint=mesh` turns when newer private RCA fills newest-`event_time` (TUI sticky limit). Scan the window past the receipt cap, merge `source_hint:mesh` tagged entries, and prefer source-class diversity (mesh+private) when both exist — never invent mesh. Payload keeps honest `since`/`as_of` and adds `receipt_selection` (scan/class flags). Companion primary UX is [iomesh-tui#419](https://github.com/iome-sh/iomesh-tui/issues/419).
- **`ops_digest_export` receipt class wire (#66 residual / #68):** receipts still labeled `source_hint=palace_timeline` after #67, so TUI `ClassifyDigestReceipt` could not see palace mesh. Each receipt now copies palace `provenance.source_hint` / `source_step` and tags (`source_hint:mesh` / TemporalTags included), and sets receipt `source_hint=mesh` when the entry is mesh-sourced — never invented. Sticky default limit still keeps ≥1 mesh + ≥1 private when both exist in-window.

## [0.3.1] — 2026-09-10

Kernel v1.5.10 + optional ingest class. Compatible with TUI durable pull `source_hint=mesh` (iomesh-tui#418). dual_write OFF · not Memory GA · catalog ≠ connected.

### Added
- **`memory_ingest_turn` optional `source_hint` (#63 / #65):** callers (durable mesh pull) can pass `mesh`, `private`, or a kernel-classifiable alias. When non-empty, the host stamps `provenance.source_hint` and tag `source_hint:<hint>` before `IngestTurn` (`FormatSourceHintTag`). When omitted or blank, kernel `ensurePrivateIngestSource` keeps today’s private default — session ids such as `dept.*.events.*` do not invent mesh. `ops_digest_export` honors a stamped mesh/private class on the entry (never invents mesh).

### Changed
- **Kernel pin (#58 / #60):** `github.com/iome-sh/memory` annotated **`v1.5.9`** → annotated **`v1.5.10`** (memory #90 / PR #91: `IngestTurn` stamps observable `provenance.source_hint=private` and tag `source_hint:private` when the caller does not already supply a classifiable mesh or private source). Host process labels (`mcp_memory_ingest_turn`, `source:iomesh-memory-mcp`) are not a cite-both class.

## [0.3.0] — 2026-09-10

### Added
- **`ops_digest_export` (#55):** lean MCP tool for TUI `/memory digest` when sync `POST /v1|/v5/memory/ops_digest` is unavailable. Args: `window` (day|week, default day), `horizon` (ops|knowledge|analytical|all, default ops), `limit` (default 20, cap 50), optional `tenant` / `as_of`. Returns TUI `MemoryOpsDigestResult` JSON: window/horizon/as_of/since, honesty (`dual_write_default=off`, `never_invent_ga`, knowledge/analytical Beta, book_demo off), empty `patterns` (insufficient-signal OK — do not invent GA), local palace `receipts` with `source_hint=palace_timeline` (TUI-classifiable private; mesh* only when the entry is mesh-sourced — never invented), empty `decision_stub`. Advertised in `leanToolNames` / `GET /healthz` `tool_names` / `-preflight`. dual_write OFF · not Memory GA · catalog ≠ connected.

## [0.2.1] — 2026-09-10

### Changed
- **Kernel pin:** `github.com/iome-sh/memory` annotated **`v1.5.8`** → annotated **`v1.5.9`** (memory #88: palace mode bits / writeMu / retrieve skip archival; hugot 0.7.8). `go` line follows the kernel (`1.26.6` → `1.27.0`). dual_write OFF · not Memory GA.

## [0.2.0] — 2026-09-10

P0 host polish (TTFH): ingest DLP, HTTP loopback + optional secret, memory
`v1.5.8` pin. dual_write OFF · not Memory GA · Catalog ≠ Connected.

### Added
- **Host-side DLP on ingest (#48):** `memory_ingest_turn` and `memory_write` redact common secret-shaped tokens (`ghp_` / `sk-` / AWS `AKIA` / Slack `xox*` / PEM / JWT-shaped) to `[REDACTED]` before palace write. Residual heuristics — not commercial DLP, not hardware-bound keys, not default envelope encryption. `GET /healthz` honesty unchanged. dual_write OFF · not Memory GA.
- **HTTP loopback default + optional shared secret (#49):** `:8080` is forced to `127.0.0.1:8080`. `0.0.0.0` / non-loopback requires `-allow-non-loopback` / `MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK`. Optional `MEMORY_MCP_HTTP_SECRET` (`X-Memory-MCP-Secret` or `Authorization: Bearer`) fail-closes MCP HTTP when set. `/healthz` stays open. stdio remains default when `-http-addr` is empty. Compose/image set the allow flag so host publish `127.0.0.1:8080` can reach the container. dual_write OFF · not Memory GA.

### Changed
- **Kernel pin (#50):** `github.com/iome-sh/memory` `v1.5.8-0.20260816062432-e1ffb9db873e` → annotated **`v1.5.8`**. Current kernel main tip `f834699` is docs-only after the tag; kernel #85/#86 (mode bits / writeMu) were still open so this hop does not wait on them. dual_write OFF · not Memory GA.

## [0.1.1] — 2026-09-06

Patch: leftover product-plane env alias reads (#45) and public OSS narrative
scrub (#46). Not a new MCP tool surface. dual_write OFF · not Memory GA.

### Changed
- **Public OSS env scrub:** drop leftover legacy product-plane env alias reads and deprecation helpers (`firstEnvPrefer`, `warnDeprecatedEnvAliases`). Host reads `MEMORY_MCP_*` / `PALACE_ROOT` / `MEMORY_TENANT` only. Operator docs (`.env.example`, EDGE_DOGFOOD) no longer list those aliases. Does not import private control-plane/broker packages. Naming remains **iomesh-memory-mcp**. dual_write OFF · not Memory GA.
- **Public OSS narrative scrub:** replace leftover product-codename language in docs, comments, templates, and honesty-gate needles with **private control plane / broker** wording. Gates require the new wording and forbid the old codename (they no longer require the old token to appear). Naming remains **iomesh-memory-mcp**. dual_write OFF · not Memory GA.

## [0.1.0] — 2026-09-04

First annotated `v*` GitHub Release of this local palace MCP host. Omitted
`tenant` fail-closes (no fallback to process `-tenant` / `MEMORY_TENANT` /
`"default"`). Path isolation `PALACE_ROOT/<tenant>/` ≠ cloud multi-tenant.
dual_write OFF · not Memory GA.

### Changed
- **Public copy hygiene:** operator-facing README, RELEASING, SECURITY, OPEN_SOURCE_AUDIT, and public-flip notes drop internal serials and private-plane names. Path isolation `PALACE_ROOT/<tenant>/` ≠ cloud multi-tenant. dual_write OFF · not Memory GA.
- **Omitted tenant fail-closed (#40):** tool/HTTP calls that omit `tenant` return an error (`IsError` / tenant required). They do not fall back to `-tenant` / `MEMORY_TENANT` or `"default"`, so two callers on one MCP HTTP process cannot mix in `PALACE_ROOT/default`. Invalid tenant still fail-closed. Path isolation `PALACE_ROOT/<tenant>/` unchanged. `GET /healthz` stays honest (no tenant/org leak; `dual_write=off`; `not_memory_ga=true`). dual_write OFF · not Memory GA.
- **govulncheck:** pin indirect `golang.org/x/crypto` `v0.54.0` → `v0.56.0` (GO-2026-6354 / GO-2026-6355 via optional ONNX/`ssh.Dial` residual). Kernel pin unchanged. dual_write OFF · not Memory GA.

### Added
- **CLI `-preflight` (#28):** constructs the host and prints the same honesty JSON as `GET /healthz` (`status`, `service`, `dual_write=off`, `not_memory_ga`, `embeddings`, `qdrant=off`, `version`, `tools`, `tool_names`), then exits without listening or running stdio MCP. Registration ≠ `tools/list` ≠ ingest. No hosted palace probe. dual_write OFF · not Memory GA.
- **Tenant single path segment (#27):** when `tenant` is provided (tool input or `-tenant` / `MEMORY_TENANT`), it must be a single path segment (reject `.`, `..`, separators). Invalid tool tenant returns an `IsError` result; invalid default tenant fails process start. (#40 later fail-closes omitted/empty tool tenant.) Same-process path isolation only · dual_write OFF · not Memory GA.

### Fixed
- **Docs close-token (#38):** README and EDGE_DOGFOOD no longer use the internal workflow close-token. Operator language is human product close / still-open product gates. `GET /healthz` 200 is not Connected. dual_write OFF · not Memory GA.
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
  - Honesty: tip ≠ invent successful public tag release shipped · residual PASS ≠ invent forever-green signed releases · dual_write OFF · not Memory GA · naming **iomesh-memory-mcp** · kernel public prerequisite met · no auto-tag · private control plane / broker stays out of this tree
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

- dual_write **OFF** · not product Memory GA · host + kernel public · residual PASS ≠ live dogfood / invent forever-green signed releases · residual PASS ≠ invent Edge Memory GA · readiness ≠ invent flip · tip ≠ invent tag release shipped · does not import private control-plane/broker packages · naming **iomesh-memory-mcp** · kernel public prerequisite met · M5 packaging residual (s1492) ≠ invent M5 complete · s1500 E3–E5 docs ≠ invent Edge Memory GA declared · s1504 local evidence ≠ invent Edge Memory GA / forever product green · s1509 client attach ≠ invent Edge Memory GA / forever green full product dogfood · healthz `tools` / `tool_names` = compile-time lean surface ≠ invent live TUI attach restamp / forever-green `tools=N` · no auto-tag

## [0.1.0-s1457] — 2026-08-08

### Added

- Lean edge MCP host scaffold (**s1457** / Option A M2):
  - Binary **`iomesh-memory-mcp`**
  - stdio + streamable HTTP (`MEMORY_MCP_HTTP_ADDR`) with `GET /healthz`
  - Tools: `memory_ingest_turn`, `memory_retrieve`, `memory_search_semantic`,
    `memory_list`, `memory_compact_status`, `memory_facts_as_of`
  - Path-based tenant layout: `filepath.Join(palaceRoot, tenant)`
  - Depends on `github.com/iome-sh/memory` + `modelcontextprotocol/go-sdk` only (does not import private control-plane/broker packages)
  - TUI-grade OSS process bar: LICENSE, NOTICE, SECURITY, community docs,
    RELEASING, CHANGELOG, OPEN_SOURCE_AUDIT, Makefile, CI, Dependabot, Dockerfile, compose
  - **Repository remains private** until a deliberate visibility flip
  - dual_write **OFF** · not product Memory GA · private control plane / broker stays out of this tree · no default Qdrant/ONNX requirement

[Unreleased]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/iome-sh/iomesh-memory-mcp/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.1.0
