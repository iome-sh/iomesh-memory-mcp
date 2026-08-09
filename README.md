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

- **stdio or HTTP** — default stdio for local clients; optional streamable HTTP + `GET /healthz`
- **Local-first** — durable data under `PALACE_ROOT` on disk
- **Thin host** — tools map to the public `github.com/iome-sh/memory` API
- **Tenant paths** — one process, filesystem isolation by tenant subdirectory
- **Releases** — multi-platform binaries via GoReleaser (SBOM + keyless cosign on checksums)

## Install

### From source

```bash
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@latest
```

### Build from a clone

```bash
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make build   # → bin/iomesh-memory-mcp
```

Requires the Go version in [`go.mod`](go.mod). The kernel dependency is public: `github.com/iome-sh/memory`.

### Tagged releases

Push an annotated `v*` tag to run [`.github/workflows/release.yml`](.github/workflows/release.yml). See [RELEASING.md](RELEASING.md) for the release checklist and signing matrix.

Local dry-run (needs `goreleaser` + `syft` on `PATH`):

```bash
make release-snapshot
```

## Quick start

### Stdio (local MCP client)

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
./bin/iomesh-memory-mcp -palace-root "$PALACE_ROOT" -tenant "$MEMORY_TENANT"
```

### HTTP (streamable MCP + health)

```bash
./bin/iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp

curl -fsS http://127.0.0.1:8080/healthz
```

### Client config example

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

### Docker Compose

```bash
docker compose up --build
curl -fsS http://127.0.0.1:8080/healthz
# expect dual_write=off · not_memory_ga=true · embeddings=hash|onnx · qdrant=off
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

**Honesty:** ONNX improves embeddings · dual_write **OFF** · **not** Memory GA · Qdrant still **off** for lean host search · optional path ≠ invent platform GPU palace.

## Configuration

| Flag | Environment | Default | Notes |
|------|-------------|---------|--------|
| `-palace-root` | `PALACE_ROOT` | `./data/memory-palaces` (or `/data/memory-palaces` in image) | Base directory for tenants |
| `-tenant` | `MEMORY_TENANT` | `default` when empty | Tenant subdirectory |
| `-http-addr` | `MEMORY_MCP_HTTP_ADDR` | empty = **stdio** | e.g. `:8080` |
| `-http-path` | `MEMORY_MCP_HTTP_PATH` | `/mcp` | Streamable MCP path (`/healthz` is fixed) |
| (env only) | `MEMORY_ONNX_MODEL_PATH` | empty = **hash** embeddings | Optional ONNX model dir/file for stronger semantic retrieve · see [memory](https://github.com/iome-sh/memory) README |
| (env only) | `MEMORY_EMBEDDING_STRICT` | unset | When `true`, ONNX errors do not fall back to hash (kernel) |
| (env only) | `MEMORY_HUGOT_BACKEND` | `go` | Kernel hugot backend (`go` / `ort` / `auto`) |

**Embeddings:** default is **hash** (no extra deps). Set `MEMORY_ONNX_MODEL_PATH` to maximize semantic `/memory semantic` and hybrid retrieve quality in clients such as `iomesh-tui`.

**Qdrant:** **not required** and **not wired** into this lean host’s search path (`healthz.qdrant=off`). The [memory](https://github.com/iome-sh/memory) kernel has an optional VectorStore API for custom Go; running Qdrant does not change lean host behavior today.

## MCP tools

| Tool | Kernel API |
|------|------------|
| `memory_ingest_turn` | `IngestTurn` |
| `memory_retrieve` | `SearchMemoryWithOptions` |
| `memory_search_semantic` | Hybrid search on semantic tier |
| `memory_list` | `ListMemoryWithOptions` |
| `memory_compact_status` | `GetStats` |
| `memory_facts_as_of` | `ListFactsAsOf` |

Server name: **`iomesh-memory-mcp`**. Default version stamp: **`v0.1.0`** (overridden by `make build` / GoReleaser ldflags).

## Tenant layout

```text
$PALACE_ROOT/
  <tenant>/
    tier-1-working/
    tier-2-contextual/
    …
```

Isolation is path-based within a single process.

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
| [RELEASING.md](RELEASING.md) | Tags, GoReleaser, SBOM, cosign · support / version policy (E5) |
| [SECURITY.md](SECURITY.md) | Security policy |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributor guide |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
| [SUPPORT.md](SUPPORT.md) | Issues, security, support scope |
| [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) | E3 install matrix · E4 operator dogfood runbook |
| [docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md) | M4 public-flip readiness residual |
| [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md) | OSS process checklist |

## Related projects

| Repository | Role |
|------------|------|
| [memory](https://github.com/iome-sh/memory) | Go memory kernel library |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Agent TUI/CLI (MCP client) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |

## License

[MIT](LICENSE) · [NOTICE](NOTICE)
