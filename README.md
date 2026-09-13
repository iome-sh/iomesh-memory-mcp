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
# :8080 is forced to 127.0.0.1:8080. 0.0.0.0 requires -allow-non-loopback.

curl -fsS http://127.0.0.1:8080/healthz
# persist_embeddings=off (default) · qdrant=off · tools>=11 (compile-time)
# /healthz stays open even if MEMORY_MCP_HTTP_SECRET is set
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
Optional `MEMORY_MCP_HTTP_SECRET` fail-closes the MCP path when set
(`X-Memory-MCP-Secret` or `Authorization: Bearer`). Then:

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
registration, not a live MCP `tools/list`.

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

## Configuration

| Flag | Environment | Default | Notes |
|------|-------------|---------|--------|
| `-palace-root` | `PALACE_ROOT` | `./data/memory-palaces` (or `/data/memory-palaces` in image) | Base directory for tenants |
| `-tenant` | `MEMORY_TENANT` | empty | Process label only (validated if set). Tool `tenant` is required; omit fail-closes |
| `-http-addr` | `MEMORY_MCP_HTTP_ADDR` | empty = **stdio** | e.g. `:8080` (forced to `127.0.0.1:8080`) |
| `-http-path` | `MEMORY_MCP_HTTP_PATH` | `/mcp` | Streamable MCP path (`/healthz` is fixed) |
| `-allow-non-loopback` | `MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK` | false | Required to bind `0.0.0.0` / `::` / LAN. Compose/image set this so host publish `127.0.0.1:8080` can reach the container |
| `-http-secret` | `MEMORY_MCP_HTTP_SECRET` | empty = **off** | Optional shared secret for MCP HTTP. Fail-closed when set. `/healthz` stays open |
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
| `memory_ingest_turn` | `IngestTurn` | Write a conversation turn; optional `source_hint`; host DLP redacts common secret shapes |
| `memory_extract_facts` | `ExtractAtomicFacts` + `Write` | HITL extract-after-persist; writes `turn_fact` children without rewriting the parent |
| `memory_write` | `Write` / `WriteAndSupersede` | Write a durable fact (same DLP as ingest) |
| `memory_retrieve` | `SearchMemoryWithOptions` | Keyword + optional vector re-rank; does not ingest |
| `memory_search_semantic` | Hybrid search on semantic tier | Hybrid semantic search; Qdrant not wired |
| `memory_list` | `ListMemoryWithOptions` | List by event time; does not ingest |
| `memory_compact_status` | `GetStats` | Local palace stats; does not ingest |
| `memory_facts_as_of` | `ListFactsAsOf` | Facts valid at `as_of`; does not ingest |
| `memory_related` | `MultiHopRetrieve` | Entity BFS lite; does not ingest |
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
on a shared root remain unsupported. HTTP defaults to loopback; optional
`MEMORY_MCP_HTTP_SECRET`; unauthenticated HTTP remains the residual when the
secret is unset.

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
