# iomesh-memory-mcp

[![ci](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/iomesh-memory-mcp/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Lean edge Memory MCP host** — stdio or streamable HTTP tools over the Palace kernel (`github.com/iome-sh/memory`).

| Docs | |
|------|--|
| [LICENSE](LICENSE) (MIT) · [NOTICE](NOTICE) | [SECURITY](SECURITY.md) · [CONTRIBUTING](CONTRIBUTING.md) |
| [SUPPORT](SUPPORT.md) · [CODE_OF_CONDUCT](CODE_OF_CONDUCT.md) | [RELEASING](RELEASING.md) · [CHANGELOG](CHANGELOG.md) |
| [Open-source audit](docs/OPEN_SOURCE_AUDIT.md) | [**M3 edge dogfood**](docs/EDGE_DOGFOOD.md) |

### Honesty locks (read first)

- **Edge host only** — this binary is **not product Memory GA**.
- **Local-primary** — durable memory lives under `PALACE_ROOT` on the operator’s filesystem.
- **dual_write OFF** by default (audit dual_write is a later optional path; not wired in lean v1).
- **Naming honesty** — product edge binary/image = **`iomesh-memory-mcp`** / `ghcr.io/iome-sh/iomesh-memory-mcp`.  
  Do **not** ship product edge as **`aion-memory-mcp`** (private monorepo may thin-wrap only).
- **No aion import** — builds MCP directly on the Palace kernel.
- **aion broker / control plane stays private**.
- **Qdrant / ONNX not required** for the default hybrid path (kernel hash/simple embeddings).
- **Visibility:** this repository is **still private** until a deliberate public flip ([audit](docs/OPEN_SOURCE_AUDIT.md)). M2 lean scaffold (**s1457**) + **M3** edge dogfood surfaces (**s1462**) — residual PASS ≠ public flip / live dogfood green / full platform sidecar parity.
- Program continuum: free eng concurrent **s1462+** after free-floor **s1460** · lag **s1461** · peers TUI s1463 · aion residual s1464 (mention only) · free-floor peer s1465 · free eng after s1467+ · **M4** public flip later.

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
# image: iomesh-memory-mcp:local  — compose PASS ≠ public registry · build PASS ≠ invent GA
curl -fsS http://127.0.0.1:8080/healthz
```

## M3 edge dogfood

Residual-honest offline SSOT for operator dogfood (stdio · HTTP healthz/`/mcp` · local compose · tool honesty). **Not** live dogfood invent · **not** public flip · **not** Memory GA · dual_write **OFF**.

| | |
|--|--|
| Checklist | [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) |
| Offline gate | `make edge-dogfood-gate` → `scripts/edge_dogfood_gate.sh` (file greps only; **no docker daemon**) |
| Serial | **s1462** · peers TUI s1463 · aion residual s1464 (mention only) · **M4** deliberate later |

```bash
make edge-dogfood-gate   # residual PASS ≠ live dogfood green
make help                # lists targets including edge-dogfood-gate
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
make check              # fmt-check · vet · test
make ci                 # + govulncheck · build
make edge-dogfood-gate  # offline M3 residual greps (optional; not required by ci)
```

Required GitHub Actions gate: **ci-success** (lint · test · build · govulncheck).  
`edge-dogfood-gate` is an offline residual; it does **not** need a docker daemon and is not invent live dogfood.

While `memory` is private, CI needs org access or `GO_MODULE_TOKEN` with `repo` read on `iome-sh/memory`.

## Option A (edge OSS)

1. **M1** — private kernel TUI-grade bar (`memory` s1452)  
2. **M2** — this repo lean scaffold/extract (**s1457**)  
3. **M3** — edge dogfood checklist + offline gate (**s1462**; peers TUI s1463 · aion residual s1464)  
4. **M4** — public flip (kernel first, then this host) — **later · deliberate**  
5. **M5** — signing / matrix / extensions  

## License

MIT © IOMesh Technology Ltd. — see [LICENSE](LICENSE).
