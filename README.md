# iomesh-memory-mcp

[![ci](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/iome-sh/iomesh-memory-mcp)](https://github.com/iome-sh/iomesh-memory-mcp/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/iome-sh/iomesh-memory-mcp.svg)](https://pkg.go.dev/github.com/iome-sh/iomesh-memory-mcp)

**MCP host for local agent memory** — exposes the [memory](https://github.com/iome-sh/memory) kernel over [Model Context Protocol](https://modelcontextprotocol.io/) (stdio or streamable HTTP).

```text
MCP client (e.g. iomesh-tui)
        │
        ▼
iomesh-memory-mcp     ← this repo (stdio | HTTP)
        │
        ▼
github.com/iome-sh/memory  (PalaceStore)
        │
        ▼
local filesystem under PALACE_ROOT/<tenant>/…
```

This host does not dual-write to a mesh. Persist embeddings default **off**; hash embeddings are never stored.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [TTFH (V1.5 host path)](#ttfh-v15-host-path)
- [Configuration](#configuration)
- [MCP tools](#mcp-tools)
- [Tenant layout](#tenant-layout)
- [Development](#development)
- [Documentation](#documentation)
- [Related projects](#related-projects)
- [License](#license)

## Install

Pin the latest annotated GitHub Release:
**[v0.4.2](https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.4.2)**.
`@latest` / floating `main` are not production pins. Default `ServerVersion` is
`v0.4.2` (GoReleaser ldflags override on tagged assets).

```bash
# Requires Go 1.27+ (same as github.com/iome-sh/memory and iomesh-tui; see go.mod).
export PATH="$(go env GOPATH)/bin:${PATH}"
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@v0.4.2
```

Requires the Go version in [`go.mod`](go.mod). Put `$(go env GOPATH)/bin` on `PATH`.
Kernel: [`github.com/iome-sh/memory` **v1.5.12**](https://github.com/iome-sh/memory/releases/tag/v1.5.12).
Companion TUI: [iomesh-tui **v1.3.7**](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.7).

### Build from a clone

```bash
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make build   # → bin/iomesh-memory-mcp
```

### GitHub Releases

Multi-platform archives (linux/darwin/windows; amd64 + arm64 except Windows arm64)
are on the [v0.4.2](https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.4.2)
release. Checksums are keyless-cosign signed; each archive has an SPDX SBOM.
See [RELEASING.md](RELEASING.md).

Local dry-run (needs `goreleaser` + `syft` on `PATH`):

```bash
make release-snapshot
```

## Quick start

### Stdio + healthz (30 seconds)

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
iomesh-memory-mcp -preflight   # same JSON as GET /healthz; no listen
iomesh-memory-mcp -palace-root "$PALACE_ROOT" -tenant "$MEMORY_TENANT"
```

HTTP (streamable MCP + health):

```bash
iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp
# :8080 is forced to 127.0.0.1:8080. 0.0.0.0 requires -allow-non-loopback and -http-secret.

curl -fsS http://127.0.0.1:8080/healthz
# persist_embeddings=off (default) · qdrant=off · tools>=11 (compile-time)
curl -fsS http://127.0.0.1:8080/ready
# writer ready: palace writable + wal/pending recoverable; 503 if not
# /healthz ≠ /ready. Both stay open even if MEMORY_MCP_HTTP_SECRET is set
```

### Cursor / Claude Desktop

No TUI required. Point any MCP client at the binary (stdio) or HTTP URL.

**stdio** (`command` + `args`). Flags match `PALACE_ROOT` / `MEMORY_TENANT`:

```json
{
  "mcpServers": {
    "iomesh-memory-mcp": {
      "command": "iomesh-memory-mcp",
      "args": [
        "-palace-root", "/path/to/memory-palaces",
        "-tenant", "default"
      ],
      "env": {
        "PALACE_ROOT": "/path/to/memory-palaces",
        "MEMORY_TENANT": "default"
      }
    }
  }
}
```

Put that block in the client’s MCP config (`~/.cursor/mcp.json`, Claude Desktop
`claude_desktop_config.json`, or equivalent). Use an absolute `command` if
`iomesh-memory-mcp` is not on `PATH`.

**HTTP** (streamable MCP). Start the host with `-http-addr :8080 -http-path /mcp`.
Loopback may omit the secret. Non-loopback requires `MEMORY_MCP_HTTP_SECRET`
(fatal before listen). When set, the secret fail-closes the MCP path
(`X-Memory-MCP-Secret` or `Authorization: Bearer`); `/healthz` and `/ready` stay open. Then:

```json
{
  "mcpServers": {
    "iomesh-memory-mcp": {
      "url": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

`GET /healthz` 200 means the process is up. `tools` / `tool_names` are compile-time
registration, not a live MCP `tools/list`. `GET /ready` 200 means the writer palace
is writable and `wal/pending` is recoverable — not ingest. `not_memory_ga` stays on `/healthz` and `/ready` and is false only when `MEMORY_CLOUD_GA` is exactly `1`. `dual_write` stays `off`. `version` is `ServerVersion` (`v0.4.2` unless a release ldflag overrides it).

### iomesh-tui

[iomesh-tui](https://github.com/iome-sh/iomesh-tui) TOML:

```toml
[[mcp.servers]]
name = "iomesh-memory-mcp"
command = "/path/to/iomesh-memory-mcp"
args = ["-palace-root", "/path/to/memory-palaces", "-tenant", "default"]
```

HTTP (when the client supports a URL transport):

```text
url = "http://127.0.0.1:8080/mcp"
```

Cite-both is a TUI session flag (`/memory digest --require-sources mesh,private`).

### Docker Compose

```bash
docker compose up --build
curl -fsS http://127.0.0.1:8080/healthz
```

### Optional ONNX (better semantic recall)

Default path needs **no** Qdrant and **no** ONNX. To maximize hybrid/semantic quality:

```bash
# From a checkout of github.com/iome-sh/memory (public):
go run ./scripts/download_onnx_model.go
export MEMORY_ONNX_MODEL_PATH=/path/to/model   # hugot model dir or .onnx file
iomesh-memory-mcp -palace-root ./data/memory-palaces -tenant default -http-addr :8080
curl -fsS http://127.0.0.1:8080/healthz   # embeddings should report "onnx" when load succeeds
```

Compose passthrough:

```bash
MEMORY_ONNX_MODEL_PATH=/absolute/path/to/model docker compose up --build
```

`persist_embeddings` is **on** only for ONNX + `MEMORY_PERSIST_EMBEDDINGS`; hash never persists.

## TTFH (V1.5 host path)

This host is **local palace MCP**. dual_write **OFF**. PersistEmbeddings default **off**. Hash embeddings are **never** stored. Catalog is not Connected.

Pins (already in [Install](#install)): MCP **[v0.4.2](https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.4.2)** · kernel **[v1.5.12](https://github.com/iome-sh/memory/releases/tag/v1.5.12)** · TUI **[v1.3.7](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.7)**.

This list is the rollout (R1 --live ≠ R3 overlay /dashboard PULSE parked).

- **R0** `iomesh ttfh --unit` — offline smoke (no broker). [`scripts/ttfh-demo.sh`](https://github.com/iome-sh/iomesh-tui/blob/main/scripts/ttfh-demo.sh) in iomesh-tui — unit then optional live
- **R1** Optional `--live` — **EMPTY** unless decoded messages; never invent **PULSE**; not overlay PULSE
- **R2** `/memory ingest` — three RCA-shaped turns (local overlay stays **private**). `/memory digest --require-sources mesh,private` — **cite-both or explicit miss**. `/memory patterns` — ops **Beta** · empty ≠ invent · never APPLY. `/memory facts-as-of` — palace
- **R3** Overlay `/dashboard` PULSE **parked** (required for E-G1)
- **R4** After PULSE: `iomesh memory pull` — dual_write **OFF** · pull ≠ Connected

**Tools used:** `memory_ingest_turn`, `ops_digest_export`, `memory_patterns_list` / `memory_anomalies_list` (suggestive, never APPLY), `memory_facts_as_of`, `memory_retrieve`. This lean host registers ingest / digest / facts-as-of / retrieve; patterns/anomalies apply **when present** (not on this host’s `tools/list`).

**E-G1** is a **real laptop** run: **PULSE + 3 RCA + cite-both-or-miss** — **not** this README. `--unit` / `--live`, a docs PR, and CI green are not E-G1.

Kernel walk: [memory `docs/TTFH.md`](https://github.com/iome-sh/memory/blob/main/docs/TTFH.md). V1.5 tracker: `docs/planning/ttfh-pattern-search-v15-2026.md` (control-plane planning; not this host).

Department overlay kit (V1.6 Wave 1, not E-G1): TUI `examples/dept-rca/support` (ticket-export + policy + macro, `source_hint=private`). Ingest via `iomesh memory ingest-dir`. ingest-dir tags `dept:` / `scenario:` are private overlay, not Connected. `--department support` is Tag `dept:support`, not Connected. `/memory digest --require-sources mesh,private` → miss is success until pull. facts-as-of ticket created `2026-06-15T14:22:00Z`. Prefix `dept.support.events.*` is routing, not Connected. Kernel companion: memory `examples/dept-rca/support`. dual_write **OFF**.

## Configuration

| Flag | Environment | Default | Notes |
|------|-------------|---------|--------|
| `-palace-root` | `PALACE_ROOT` | `./data/memory-palaces` (or `/data/memory-palaces` in image) | Base directory for tenants |
| `-tenant` | `MEMORY_TENANT` | empty | Process label only (validated if set). Tool `tenant` is required; omit fail-closes. **Required** with `-cloud` |
| `-cloud` | `MEMORY_CLOUD` | false | Dedicated-tenant cloud mode: one PalaceStore, second tool tenant is 400, `palace.lock` PID lock, `PalaceConfig.TransactionalIngest=true`. Local-dev map remains when false (TransactionalIngest default false / partial persist) |
| `-http-addr` | `MEMORY_MCP_HTTP_ADDR` | empty = **stdio** | e.g. `:8080` (forced to `127.0.0.1:8080`) |
| `-http-path` | `MEMORY_MCP_HTTP_PATH` | `/mcp` | Streamable MCP path (`/healthz` and `/ready` are fixed) |
| `-allow-non-loopback` | `MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK` | false | Required to bind `0.0.0.0` / `::` / LAN. Compose/image set this so host publish `127.0.0.1:8080` can reach the container. **Requires** `-http-secret` |
| `-http-secret` | `MEMORY_MCP_HTTP_SECRET` | empty = **off** on loopback | Shared secret for MCP HTTP. **Required** for non-loopback (fatal before listen, not a request 401). Fail-closed when set. `/healthz` and `/ready` stay open |
| `-preflight` | — | false | Print the same JSON as `GET /healthz` and exit (no listen, no stdio MCP) |
| (env only) | `MEMORY_ONNX_MODEL_PATH` | empty = **hash** embeddings | Optional ONNX model dir/file · see [memory](https://github.com/iome-sh/memory) |
| (env only) | `MEMORY_PERSIST_EMBEDDINGS` | unset = **off** | Opt-in ONNX vector persist (`1`/`true`/`on`/`yes`, case-insensitive). Ignored on hash (never persist hash) |
| (env only) | `MEMORY_EMBEDDING_STRICT` | unset | When `true`, ONNX errors do not fall back to hash |
| (env only) | `MEMORY_HUGOT_BACKEND` | `go` | Kernel hugot backend (`go` / `ort` / `auto`) |

Default embeddings are **hash** (no extra deps). Set `MEMORY_ONNX_MODEL_PATH` for stronger
semantic retrieve. `MEMORY_PERSIST_EMBEDDINGS` is **off** unless you opt in **and**
embeddings are ONNX. Hash embeddings are **never** persisted.

Qdrant is **not wired** into this host’s search path (`healthz.qdrant=off`). The
[memory](https://github.com/iome-sh/memory) kernel has an optional VectorStore API
for custom Go; running Qdrant does not change lean host behavior.

## MCP tools

Local palace FS on the operator machine. `tools/list` and `healthz.tool_names` are
discovery / compile-time registration — they are **not** ingest.

| Tool | Kernel API | Notes |
|------|------------|-------|
| `memory_ingest_turn` | `IngestTurn` | Write a conversation turn; optional `source_hint`; optional extra `tags` (dept:/scenario: private overlay, not Connected); host DLP redacts common secret shapes |
| `memory_extract_facts` | `ExtractAtomicFacts` + `Write` | HITL extract-after-persist; writes `turn_fact` children without rewriting the parent |
| `memory_write` | `Write` / `WriteAndSupersede` | Write a durable fact (same DLP as ingest) |
| `memory_retrieve` | `SearchMemoryWithOptions` | Keyword + optional vector re-rank; optional `session_ids` any-of (union with `session_id`; empty = no extra filter); optional `tag` / `department` (`--department support` is Tag `dept:support`, not Connected); hits include `source_hint` / `source_step` when stamped; does not ingest |
| `memory_search_semantic` | Hybrid search on semantic tier | Hybrid semantic search; Qdrant not wired |
| `memory_list` | `ListMemoryWithOptions` | List by event time; optional `session_ids` any-of (union with `session_id`; empty = no extra filter); does not ingest |
| `memory_compact_status` | `GetStats` | Local palace stats; does not ingest |
| `memory_facts_as_of` | `ListFactsAsOf` | Facts valid at `as_of`; optional `session_ids` any-of (union with `session_id`; empty = no extra filter); optional `tag` / `department` (`--department support` is Tag `dept:support`, not Connected); does not ingest |
| `memory_related` | `MultiHopRetrieve` | Entity BFS lite; optional `session_ids` any-of (union with `session_id`; empty = no extra filter); does not ingest |
| `memory_supersede_entity` | `SupersedeEntityFacts` | Close open facts for an entity key |
| `ops_digest_export` | `ListMemoryWithOptions` window | Local receipts for TUI `/memory digest`; does not ingest |

Server name: **`iomesh-memory-mcp`**. Default version stamp: **`v0.4.2`**
(overridden by `make build` / GoReleaser ldflags).

## Tenant layout

```text
$PALACE_ROOT/
  <tenant>/
    tier-1-working/
    tier-2-contextual/
    …
```

Isolation is path-based within a single process (`PALACE_ROOT/<tenant>/`). Tool
and HTTP calls must pass `tenant`; omit fail-closes and does not write
`PALACE_ROOT/default`. Invalid segments (`.`, `..`, separators) stay fail-closed.
Path isolation is not cloud multi-tenant security. This host does not implement
the mesh-client `X-IOMesh-Org` header.

**Supported topology:** **one host process per palace root.** Multi-process writers
on a shared root remain unsupported. Cloud mode (`-cloud` / `MEMORY_CLOUD=1`)
enforces that in-process: `Host.stores` length 1, required `MEMORY_TENANT`,
`palace.lock` PID lock, second tool tenant is 400, and
`PalaceConfig.TransactionalIngest=true` (`wal/pending` intent log; not flock).
Local-dev leaves TransactionalIngest **false** (partial persist). HTTP defaults
to loopback; non-loopback requires `MEMORY_MCP_HTTP_SECRET` before listen;
`/healthz` and `/ready` stay open (`/healthz` ≠ `/ready`). Unauthenticated HTTP remains the residual on loopback
when the secret is unset. dual_write OFF · PersistEmbeddings default off.

## Development

```bash
make check   # fmt-check · vet · test
make ci      # + govulncheck · build
make test
```

Optional offline checklists (file greps only; not required for `ci-success`):

```bash
make edge-dogfood-gate
make public-flip-readiness-gate
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## Documentation

| Document | Description |
|----------|-------------|
| [CHANGELOG.md](CHANGELOG.md) | Release notes |
| [RELEASING.md](RELEASING.md) | Tags, GoReleaser, SBOM, cosign · support / version policy |
| [SECURITY.md](SECURITY.md) | Security policy |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributor guide |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
| [SUPPORT.md](SUPPORT.md) | Issues, security, support scope |
| [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) | Install matrix and operator dogfood runbook |

## Related projects

| Repository | Role |
|------------|------|
| [memory](https://github.com/iome-sh/memory) | Go memory kernel library |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Agent TUI/CLI (MCP client) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |
| [iomesh-client-sdk-python](https://github.com/iome-sh/iomesh-client-sdk-python) | Official Python client for I/O Mesh (**Beta** / pre-1.0) |

## License

[MIT](LICENSE) · [NOTICE](NOTICE)

## Maintainer notes

Not operator how-tos. Flip is already public MIT.

| Document | Description |
|----------|-------------|
| [docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md) | Maintainer residual (flip complete) |
| [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md) | Maintainer OSS process residual |
