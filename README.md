# iomesh-memory-mcp

[![ci](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Lean edge Memory MCP host** — stdio or streamable HTTP tools over the Palace kernel (`github.com/iome-sh/memory`).

| Docs | |
|------|--|
| [LICENSE](LICENSE) (MIT) · [NOTICE](NOTICE) | [SECURITY](SECURITY.md) · [CONTRIBUTING](CONTRIBUTING.md) |
| [SUPPORT](SUPPORT.md) · [CODE_OF_CONDUCT](CODE_OF_CONDUCT.md) | [RELEASING](RELEASING.md) · [CHANGELOG](CHANGELOG.md) |
| [Open-source audit](docs/OPEN_SOURCE_AUDIT.md) | |

### Honesty locks (read first)

- **Edge host only** — this binary is **not product Memory GA**.
- **Local-primary** — durable memory lives under `PALACE_ROOT` on the operator’s filesystem.
- **dual_write OFF** by default (audit dual_write is a later optional path; not wired in lean v1).
- **Naming honesty** — product edge binary/image = **`iomesh-memory-mcp`** / `ghcr.io/iome-sh/iomesh-memory-mcp`.  
  Do **not** ship product edge as **`aion-memory-mcp`** (private monorepo may thin-wrap only).
- **No aion import** — builds MCP directly on the Palace kernel.
- **aion broker / control plane stays private**.
- **Qdrant / ONNX not required** for the default hybrid path (kernel hash/simple embeddings).
- **Visibility:** this repository is **still private** until a deliberate public flip ([audit](docs/OPEN_SOURCE_AUDIT.md)). s1457 is Option A **M2** lean scaffold + TUI-grade OSS *process* bar only.
- Program continuum: free eng concurrent **s1457+** after free-floor **s1455** · lag **s1456** · peers TUI s1458 · aion residual s1459 (mention only).

```text
iomesh-tui (or other MCP client)
        │
        ▼
iomesh-memory-mcp   ← this repo (stdio | HTTP)
        │
        ▼
github.com/iome-sh/memory  PalaceStore
        │
        ▼
local palace FS   (PALACE_ROOT/<tenant>/…)
```

## Quick start

### Build

```bash
export GOPRIVATE=github.com/iome-sh/*
export GONOSUMDB=github.com/iome-sh/*
make build   # → bin/iomesh-memory-mcp
```

### Stdio (local / TUI attach)

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
./bin/iomesh-memory-mcp -palace-root "$PALACE_ROOT" -tenant "$MEMORY_TENANT"
```

### HTTP (streamable MCP + healthz)

```bash
./bin/iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp

curl -fsS http://127.0.0.1:8080/healthz
# {"status":"ok","service":"iomesh-memory-mcp","dual_write":"off","not_memory_ga":true,"version":"v0.1.0-s1457"}
```

### TUI attach snippet (illustrative)

```toml
# e.g. MCP client config — product name is iomesh-memory-mcp
[[mcp.servers]]
name = "iomesh-memory-mcp"
command = "/path/to/iomesh-memory-mcp"
args = ["-palace-root", "/path/to/memory-palaces", "-tenant", "default"]
```

HTTP attach (when supported by the client):

```text
url = "http://127.0.0.1:8080/mcp"
```

### Docker Compose (local edge)

```bash
# While kernel is private, pass a GitHub token with read access to iome-sh/memory:
export GH_TOKEN=ghp_...
docker compose up --build
curl -fsS http://127.0.0.1:8080/healthz
```

## CLI / env

| Flag | Env | Default | Notes |
|------|-----|---------|--------|
| `-palace-root` | `PALACE_ROOT` | `./data/memory-palaces` or `/data/memory-palaces` | Tenant base dir |
| `-tenant` | `MEMORY_TENANT` | `default` (when empty at resolve) | Default tenant subdir |
| `-http-addr` | `MEMORY_MCP_HTTP_ADDR` | empty = **stdio** | e.g. `:8080` |
| `-http-path` | `MEMORY_MCP_HTTP_PATH` | `/mcp` | Streamable MCP path |

Deprecated env aliases (one log line if set without preferred key):

- `AION_MEMORY_MCP_HTTP_ADDR` → prefer `MEMORY_MCP_HTTP_ADDR`
- `AION_MEMORY_MCP_HTTP_PATH` → prefer `MEMORY_MCP_HTTP_PATH`
- `AION_PALACE_ROOT` → prefer `PALACE_ROOT`

Lean v1 has **no** `-enable-audit` / `-aion-url` (dual_write residual · default OFF).

## MCP tools

| Tool | Kernel call |
|------|-------------|
| `memory_ingest_turn` | `IngestTurn` |
| `memory_retrieve` | `SearchMemoryWithOptions` |
| `memory_search_semantic` | hybrid on `TierSemantic` (+ substring residual) |
| `memory_list` | `ListMemoryWithOptions` |
| `memory_compact_status` | `GetStats` |
| `memory_facts_as_of` | `ListFactsAsOf` |

Server name: **`iomesh-memory-mcp`** · version stamp e.g. **`v0.1.0-s1457`**.

## Tenant layout

```text
$PALACE_ROOT/
  <tenant>/
    tier-1-working/
    tier-2-contextual/
    …
```

Multi-tenant isolation is **path-based** in one process (residual-honest: not cloud multi-tenant).

## Develop / CI

```bash
make check   # fmt-check · vet · test
make ci      # + govulncheck · build
```

Required GitHub Actions gate: **ci-success** (lint · test · build · govulncheck).

While `memory` is private, CI needs org access or `GO_MODULE_TOKEN` with `repo` read on `iome-sh/memory`.

## Option A (edge OSS)

1. **M1** — private kernel TUI-grade bar (`memory` s1452)  
2. **M2** — this repo lean scaffold/extract (**s1457**)  
3. **M3** — TUI/aion dogfood + deprecations (peers)  
4. **M4** — public flip (kernel first, then this host)  
5. **M5** — signing / matrix / extensions  

## License

MIT © IOMesh Technology Ltd. — see [LICENSE](LICENSE).
