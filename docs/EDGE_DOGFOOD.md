# M3 edge dogfood (offline SSOT)

Operator checklist for residual-honest **edge dogfood** of
`github.com/iome-sh/iomesh-memory-mcp` with TUI (or any MCP client) over
**stdio** or **streamable HTTP**, plus optional local Docker Compose.

**Serial stamp:** **s1462** · free eng concurrent **s1462+** after free-floor
**s1460** · lag **s1461** · peers TUI **s1463** · aion residual **s1464**
(mention only) · free-floor peer **s1465** · free eng after **s1467+** ·
next **M4** public flip is deliberate and later (not this serial).

---

## Honesty locks (read first — non-claims)

| Lock | Meaning |
|------|---------|
| **dual_write OFF** | Lean host does not publish audit dual_write. Default residual; do not invent ON. |
| **not Memory GA** | This binary is the **edge host only**, not product Memory GA. |
| **public** | Repo is public MIT · dual_write OFF · not Memory GA · still private was pre-flip residual language retired. |
| **residual PASS ≠ live dogfood** | Offline gate / checklist PASS does **not** invent a live green dogfood run. |
| **residual PASS ≠ public flip** | Process bar + docs do not flip GH visibility or invent GHCR publish green. |
| **residual PASS ≠ full platform sidecar parity** | Lean extract is not the private aion `aion-memory-mcp` sidecar feature set. |
| **no aion import** | Builds on `github.com/iome-sh/memory` + MCP SDK only. |
| **naming honesty** | Product edge = **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`). |
| **rates ~$88 / ~$119** | Mesh base footprint ~$88 · Memory Ops Pack ~$119/ws — mesh optional for durable pull/retain; local FS path needs neither. |
| **open boxes stay open** | Still-human APPLY / product gates remain open; do not close by residual alone. |
| **Palace sunset** | Hosted Palace path remains sunset / residual; local-primary FS is the edge dogfood path. |
| **mesh optional for pull** | Mesh credentials + platform endpoint are optional; offline edge dogfood uses local Palace FS only. |
| **compose PASS ≠ public registry** | `docker compose up --build` uses **local image** `iomesh-memory-mcp:local` only. |
| **build PASS ≠ invent GA** | Binary or image build success is not Memory GA. |
| **gate does not need docker daemon** | `make edge-dogfood-gate` is offline file greps only (no docker, no long-running server, no gcloud). |

---

## 0. Offline residual gate (CI-friendly)

No docker daemon, no listening server, no gcloud:

```bash
make edge-dogfood-gate
# → scripts/edge_dogfood_gate.sh  (exit 0 PASS / 1 FAIL)
```

This only asserts docs + compose/Dockerfile/README/Makefile needles + host layout.
**PASS here ≠ live dogfood green.**

---

## 1. Build binary

While `github.com/iome-sh/memory` is private, module fetch needs org access:

```bash
export GOPRIVATE=github.com/iome-sh/*
export GONOSUMDB=github.com/iome-sh/*
# optional: GH_TOKEN / netrc with repo read on iome-sh/memory

make build   # → bin/iomesh-memory-mcp
# or: go build -o bin/iomesh-memory-mcp ./cmd/iomesh-memory-mcp
```

Honesty: **build PASS ≠ invent GA** · dual_write remains OFF · not Memory GA.

---

## 2. Stdio attach path (TUI / MCP client)

Default transport is **stdio** when `-http-addr` / `MEMORY_MCP_HTTP_ADDR` is empty.

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
./bin/iomesh-memory-mcp \
  -palace-root "$PALACE_ROOT" \
  -tenant "$MEMORY_TENANT"
# log line should include: dual_write=off not_memory_ga=true mode=stdio
```

Illustrative TUI / client config:

```toml
[[mcp.servers]]
name = "iomesh-memory-mcp"
command = "/path/to/iomesh-memory-mcp"
args = ["-palace-root", "/path/to/memory-palaces", "-tenant", "default"]
```

Peers: TUI product tip serial **s1463** (mention only — not implemented in this repo).

---

## 3. HTTP healthz + `/mcp` path

```bash
./bin/iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp

curl -fsS http://127.0.0.1:8080/healthz
# expect JSON with:
#   "status":"ok"
#   "service":"iomesh-memory-mcp"
#   "dual_write":"off"
#   "not_memory_ga":true
```

Streamable MCP endpoint (client-dependent):

```text
http://127.0.0.1:8080/mcp
```

Env equivalents: `MEMORY_MCP_HTTP_ADDR=:8080` · `MEMORY_MCP_HTTP_PATH=/mcp`.
Prefer these over deprecated `AION_MEMORY_MCP_*` aliases.

---

## 4. Docker Compose (local image only)

```bash
# While kernel is private, pass a token with read on iome-sh/memory:
export GH_TOKEN=ghp_...   # or equivalent; never commit

docker compose up --build
# image: iomesh-memory-mcp:local  (NOT a public registry publish claim)

curl -fsS http://127.0.0.1:8080/healthz
```

| Claim | Truth |
|-------|--------|
| Local image tag | `iomesh-memory-mcp:local` |
| Public GHCR green | **Not claimed** on this serial |
| compose PASS | ≠ public registry · ≠ invent GA · ≠ dual_write ON |
| Docker daemon | **Required for this optional path only** — **not** required for `make edge-dogfood-gate` |

---

## 5. Round-trip tool honesty (dual_write OFF)

Lean tool surface (kernel FS only; audited always false / dual_write off in outputs):

| Tool | Intent |
|------|--------|
| `memory_ingest_turn` | Ingest turn → Palace FS |
| `memory_retrieve` | Search / retrieve |
| `memory_search_semantic` | Hybrid semantic (+ residual) |
| `memory_list` | List with options |
| `memory_compact_status` | Stats (`GetStats`) |
| `memory_facts_as_of` | Facts as-of |

Operator expectations when dogfooding against a client:

1. Ingest a user/assistant turn under a tenant.
2. Retrieve / list and confirm FS-backed hits under `$PALACE_ROOT/<tenant>/…`.
3. Call `memory_compact_status` — payload should reflect dual_write **off** / not Memory GA residual.
4. Do **not** expect aion audit dual_write publish, mesh pull, or platform sidecar parity.

**residual PASS ≠ live dogfood** — this section is the honesty SSOT for a human or client round-trip; CI does not run a live MCP session.

---

## 6. Product continuum (Option A)

| Milestone | Status on s1462 |
|-----------|-----------------|
| **M1** kernel TUI-grade OSS process bar (`memory`) | Prior |
| **M2** lean host scaffold (this repo, s1457) | Shipped private |
| **M3** edge dogfood surfaces (this doc + gate) | **This serial** — offline SSOT |
| **M4** public flip (kernel first, then this host) | **Later · deliberate** |
| **M5** signing / matrix / extensions | Later |

Peers (mention only): TUI s1463 dogfood tip · aion residual s1464 · free-floor peer s1465 · free eng after s1467+.

---

## 7. Rates / mesh / Palace residual (narrative only)

- **Local edge path:** TUI + this MCP host + LLM — local Palace FS; **no** Qdrant / Cloud Run palace required for lean dogfood.
- **Mesh-backed path (optional):** local stack + mesh credentials + platform endpoint + **Memory Ops Pack (~$119/ws)** for durable pull/retain; mesh base footprint separate (**~$88**). Not required for offline edge dogfood.
- **Hosted Palace sunset:** do not invent always-on hosted Palace GA; local-primary remains the edge default.
- **Open boxes stay open** until deliberate product APPLY.

---

## 8. Related files

| Path | Role |
|------|------|
| [README.md](../README.md) | Quick start + M3 link |
| [docker-compose.yml](../docker-compose.yml) | Local HTTP dogfood image |
| [Dockerfile](../Dockerfile) | Multi-stage build → `iomesh-memory-mcp` |
| [Makefile](../Makefile) | `edge-dogfood-gate` · `check` · `ci` |
| [scripts/edge_dogfood_gate.sh](../scripts/edge_dogfood_gate.sh) | Offline residual greps |
| [docs/OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md) | Visibility / OSS process bar |
| [CHANGELOG.md](../CHANGELOG.md) | s1462 entry |

---

## M4 public-flip readiness (pointer)

M4 readiness residual lives in [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md)
(`make public-flip-readiness-gate`, serial **s1468**). Kernel public first, then this host.
**still private** on dogfood and readiness residuals · residual PASS ≠ public flip.

---

## Audit one-liner (s1462)

**M3 edge dogfood docs + offline gate shipped; dual_write OFF · not Memory GA · still private · residual PASS ≠ live dogfood / public flip / platform sidecar parity · M4 public-flip readiness → s1468.**
