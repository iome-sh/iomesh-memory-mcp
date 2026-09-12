# iomesh-memory-mcp

[![ci](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml)
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

## Features

- **stdio or HTTP** — default stdio for local clients; optional streamable HTTP + `GET /healthz` (HTTP defaults to loopback; optional shared secret)
- **Local-first** — durable data under `PALACE_ROOT` on disk
- **Thin host** — tools map to the public `github.com/iome-sh/memory` API
- **Tenant paths** — one process, filesystem isolation by tenant subdirectory
- **Releases** — multi-platform binaries via GoReleaser (SBOM + keyless cosign on checksums)

## TTFH walking skeleton

Kernel operator page: [memory `docs/TTFH.md`](https://github.com/iome-sh/memory/blob/main/docs/TTFH.md)
(three RCA-shaped turns → retrieve in-process → facts-as-of → print `source_hint`).
Cost-max: **hash embedder**, **no Qdrant**, **no cloud palace**. This host pin
(`v0.4.1`) and companion TUI pin (`v1.3.6`) are not **E-G1**. **E-G1 is not closed.**
Cite-both is a TUI session rule (`/memory digest --require-sources mesh,private`);
honest miss is success; catalog/grant ≠ cite-both; never invent mesh.
dual_write OFF · **not** Memory GA.

## Install

### From source

Pin the latest annotated `v*` GitHub Release:
[`v0.4.1`](https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.4.1).
`@latest` / floating `main` are not production pins. Default `ServerVersion` is
`v0.4.1` (GoReleaser ldflags override on tagged assets). **Not** Memory GA.
Path isolation `PALACE_ROOT/<tenant>/` ≠ cloud multi-tenant. `X-IOMesh-Org` is
a mesh-client header; this host does not implement it.

```bash
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@v0.4.1
```

### Build from a clone

```bash
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make build   # → bin/iomesh-memory-mcp
```

Requires the Go version in [`go.mod`](go.mod). The kernel dependency is public
`github.com/iome-sh/memory` **v1.5.11** (annotated tag; `go.mod` pin). Current
companion TUI pin is [iomesh-tui **v1.3.6**](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.6).
Historical: ingest + digest since [TUI **v1.3.3**](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.3)
(optional `source_hint` on ingest · `ops_digest_export` mesh+private receipts);
[`/memory extract`](https://github.com/iome-sh/iomesh-tui/releases/tag/v1.3.4) since TUI **v1.3.4**.
dual_write OFF · **not** Memory GA.

### Tagged releases

[`v0.1.0`](https://github.com/iome-sh/iomesh-memory-mcp/releases/tag/v0.1.0) is
the first annotated `v*` GitHub Release. Later annotated `v*` tags run
[`.github/workflows/release.yml`](.github/workflows/release.yml). See
[RELEASING.md](RELEASING.md) for the checklist and signing matrix.

Local dry-run (needs `goreleaser` + `syft` on `PATH`):

```bash
make release-snapshot
```

## Quick start

### Stdio (local MCP client)

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
./bin/iomesh-memory-mcp -preflight   # same honesty JSON as GET /healthz; no listen · dual_write=off · not Memory GA
./bin/iomesh-memory-mcp -palace-root "$PALACE_ROOT" -tenant "$MEMORY_TENANT"
```

### HTTP (streamable MCP + health)

```bash
./bin/iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp
# :8080 is forced to 127.0.0.1:8080. 0.0.0.0 requires -allow-non-loopback.

curl -fsS http://127.0.0.1:8080/healthz
# expect dual_write=off · not_memory_ga=true · persist_embeddings=off (default) · qdrant=off · tools>=11 (compile-time)
# healthz stays open even if MEMORY_MCP_HTTP_SECRET is set
```

### Client config example (TUI)

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

### Other MCP clients (Cursor, Claude Desktop, generic)

No TUI, mesh, or Memory Ops Pack required. Point any MCP client at the same binary
or HTTP URL. **Not** a partnership claim. **Not** Memory GA.

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
`claude_desktop_config.json`, or equivalent). `command` can be an absolute path
if `iomesh-memory-mcp` is not on `PATH`.

**HTTP** (streamable MCP). Start the host with `-http-addr :8080 -http-path /mcp`
(`:8080` binds `127.0.0.1:8080`). Optional `MEMORY_MCP_HTTP_SECRET` fail-closes
the MCP path when set (`X-Memory-MCP-Secret` or `Authorization: Bearer`). Then:

```json
{
  "mcpServers": {
    "iomesh-memory-mcp": {
      "url": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

Probe honesty (`GET /healthz` 200 is not Connected):

```bash
curl -fsS http://127.0.0.1:8080/healthz
# expect dual_write=off · not_memory_ga=true · embeddings=hash|onnx · persist_embeddings=off (default)
#   · qdrant=off + residual-honest "tools" (compile-time lean count, >=11) and "tool_names"
#   healthz.tools is compile-time registration, not a live MCP tools/list stamp
#   persist_embeddings is on only for ONNX + MEMORY_PERSIST_EMBEDDINGS; hash never persists
```

Tools exposed after `tools/list` (lean kernel maps; dual_write OFF):
`memory_ingest_turn`, `memory_extract_facts`, `memory_write`, `memory_retrieve`, `memory_search_semantic`,
`memory_list`, `memory_compact_status`, `memory_facts_as_of`, `memory_related`,
`memory_supersede_entity`, `ops_digest_export`.

### Docker Compose

```bash
docker compose up --build
curl -fsS http://127.0.0.1:8080/healthz
# expect dual_write=off · not_memory_ga=true · embeddings=hash|onnx · persist_embeddings=off (default) · qdrant=off · tools>=11
```

### Advanced: better semantic recall (optional ONNX)

Default path needs **no** Qdrant and **no** ONNX. To maximize hybrid/semantic quality:

```bash
# From a checkout of github.com/iome-sh/memory (public):
go run ./scripts/download_onnx_model.go
# then point the host at the model directory/file:
export MEMORY_ONNX_MODEL_PATH=/path/to/model   # hugot model dir or .onnx file
iomesh-memory-mcp -palace-root ./data/memory-palaces -tenant default -http-addr :8080
curl -fsS http://127.0.0.1:8080/healthz   # embeddings should report "onnx" when load succeeds
```

Compose (optional env passthrough already works if you set the variable on the host):

```bash
MEMORY_ONNX_MODEL_PATH=/absolute/path/to/model docker compose up --build
```

**Honesty:** ONNX improves embeddings · dual_write **OFF** · **not** Memory GA · Qdrant still **off** for lean host search · `persist_embeddings` default **off** (ONNX-only opt-in) · optional path ≠ invent platform GPU palace.

## Configuration

| Flag | Environment | Default | Notes |
|------|-------------|---------|--------|
| `-palace-root` | `PALACE_ROOT` | `./data/memory-palaces` (or `/data/memory-palaces` in image) | Base directory for tenants |
| `-tenant` | `MEMORY_TENANT` | empty | Process label only (validated if set). Tool `tenant` is required; omit fail-closes (does not write `PALACE_ROOT/default`) |
| `-http-addr` | `MEMORY_MCP_HTTP_ADDR` | empty = **stdio** | e.g. `:8080` (forced to `127.0.0.1:8080`) |
| `-http-path` | `MEMORY_MCP_HTTP_PATH` | `/mcp` | Streamable MCP path (`/healthz` is fixed) |
| `-allow-non-loopback` | `MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK` | false | Required to bind `0.0.0.0` / `::` / LAN. Compose/image set this so the published `127.0.0.1:8080` can reach the container. |
| `-http-secret` | `MEMORY_MCP_HTTP_SECRET` | empty = **off** | Optional shared secret for MCP HTTP. Fail-closed when set. `/healthz` stays open. stdio unchanged. |
| `-preflight` | — | false | Print the same honesty JSON as `GET /healthz` and exit (no listen, no stdio MCP; `tool_names` = registration, not ingest) |
| (env only) | `MEMORY_ONNX_MODEL_PATH` | empty = **hash** embeddings | Optional ONNX model dir/file for stronger semantic retrieve · see [memory](https://github.com/iome-sh/memory) README |
| (env only) | `MEMORY_PERSIST_EMBEDDINGS` | unset = **off** | Opt-in ONNX vector persist on palace JSON (`1`/`true`/`on`/`yes`, case-insensitive). Ignored on hash (never persist hash; kernel #45). Does not require Qdrant/usearch. Default path unchanged. |
| (env only) | `MEMORY_EMBEDDING_STRICT` | unset | When `true`, ONNX errors do not fall back to hash (kernel) |
| (env only) | `MEMORY_HUGOT_BACKEND` | `go` | Kernel hugot backend (`go` / `ort` / `auto`) |

**Embeddings:** default is **hash** (no extra deps). Set `MEMORY_ONNX_MODEL_PATH` to maximize semantic `/memory semantic` and hybrid retrieve quality in clients such as `iomesh-tui`. `MEMORY_PERSIST_EMBEDDINGS` is **off** unless you opt in **and** embeddings are ONNX. Hash embeddings are **never** persisted.

**Qdrant:** **not required** and **not wired** into this lean host’s search path (`healthz.qdrant=off`). The [memory](https://github.com/iome-sh/memory) kernel has an optional VectorStore API for custom Go; running Qdrant does not change lean host behavior today.

## MCP tools

Local palace FS on the operator machine. `tools/list` and `healthz.tool_names` are discovery / compile-time registration — they are **not** ingest. dual_write **OFF** · **not** Memory GA.

| Tool | Kernel API | Surface |
|------|------------|---------|
| `memory_ingest_turn` | `IngestTurn` | Write local FS (conversation turn). Optional `source_hint` (`mesh` / `private` or kernel-classifiable alias) stamps provenance + `source_hint:<hint>` tag; omit keeps private — do not invent mesh from session id. Host DLP redacts common secret shapes (`ghp_` / `sk-` / …) before write — residual heuristics, not commercial DLP. |
| `memory_extract_facts` | `ExtractAtomicFacts` + `Write` (`turn_fact` children) | Optional HITL extract-after-persist. Required `tenant` (omit fail-closes) + `memory_id` (parent already on disk). Optional `facts`; omit runs kernel `ExtractAtomicFacts` on a copy. Writes semantic `turn_fact` children; does **not** rewrite/delete the parent and is **not** called from ingest (extract is not a PalaceStore write-gate). dual_write OFF · not NLP · not Memory GA. TUI v1.3.4 ignores unknown tools. |
| `memory_write` | `Write` / `WriteAndSupersede` (durable facts; not a conversation turn) | Write local FS (same host DLP as ingest) |
| `memory_retrieve` | `SearchMemoryWithOptions` | Read/search local FS; does not ingest |
| `memory_search_semantic` | Hybrid search on semantic tier | Read local FS; does not ingest |
| `memory_list` | `ListMemoryWithOptions` | List local FS; does not ingest |
| `memory_compact_status` | `GetStats` | Local FS stats; does not ingest |
| `memory_facts_as_of` | `ListFactsAsOf` | List local FS; does not ingest |
| `memory_related` | `MultiHopRetrieve` (entity BFS lite; not full graph RAG) | Read local FS; does not ingest |
| `memory_supersede_entity` | `SupersedeEntityFacts` (mutating; HITL stays at the client) | Write local FS (close facts) |
| `ops_digest_export` | Local `ListMemoryWithOptions` window → receipts (TUI `/memory digest` MCP fallback) | Read/list local FS; does not ingest. Patterns stay empty (insufficient-signal OK). Receipt selection prefers mesh+private diversity when both exist in-window (not newest-`event_time` only); `source_hint=palace_timeline` for local/private; mesh only when the entry is mesh-sourced. Receipts also carry palace `provenance.source_hint` + tags so TUI can classify mesh — never invented. dual_write OFF · not Memory GA · catalog ≠ connected |

Server name: **`iomesh-memory-mcp`**. Default version stamp: **`v0.4.1`** (overridden by `make build` / GoReleaser ldflags).

## Tenant layout

```text
$PALACE_ROOT/
  <tenant>/
    tier-1-working/
    tier-2-contextual/
    …
```

Isolation is path-based within a single process (`PALACE_ROOT/<tenant>/`). Tool and HTTP calls must pass `tenant`; omit fail-closes and does not write `PALACE_ROOT/default`. Invalid segments (`.`, `..`, separators) stay fail-closed. Path isolation ≠ cloud multi-tenant security. Organization isolation for the I/O Mesh broker is a separate HTTP header (`X-IOMesh-Org`) on mesh clients; this host does not implement that.

**Supported topology:** **one host process per palace root.** Multi-process writers on a shared root remain unsupported (product contract, not a hidden defect). In-process kernel `writeMu` serializes the two shared files; two processes are last-write-wins. HTTP defaults to loopback; optional `MEMORY_MCP_HTTP_SECRET`; unauthenticated HTTP remains the residual when the secret is unset.

dual_write **OFF** · **not** Memory GA.

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
| [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) | E3 install matrix · E4 operator dogfood runbook (optional TTFH-shaped path; **E-G1 not closed**) |
| [memory docs/TTFH.md](https://github.com/iome-sh/memory/blob/main/docs/TTFH.md) | Kernel TTFH walking skeleton · cost-max hash / no Qdrant / no cloud palace · **E-G1 not closed** |
| [docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md) | Maintainer residual (flip complete; not operator how-to; public MIT ≠ Memory GA) |
| [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md) | Maintainer OSS process residual (not a product claim; public MIT ≠ Memory GA) |

## Related projects

| Repository | Role |
|------------|------|
| [memory](https://github.com/iome-sh/memory) | Go memory kernel library |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Agent TUI/CLI (MCP client) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |
| [iomesh-client-sdk-python](https://github.com/iome-sh/iomesh-client-sdk-python) | Official Python client for I/O Mesh (**Beta** / pre-1.0 — not invent 1.0 / live PyPI GA) |

## License

[MIT](LICENSE) · [NOTICE](NOTICE)
