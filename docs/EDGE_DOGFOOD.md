# Edge dogfood + Edge Memory GA candidacy (E3 / E4)

Operator checklist for residual-honest **edge dogfood** of
`github.com/iome-sh/iomesh-memory-mcp` with TUI (or any MCP client) over
**stdio** or **streamable HTTP**, plus optional local Docker Compose.

This doc advances **Edge Memory GA candidacy** exit criteria on the public binary host:

| Exit | Scope here |
|------|------------|
| **E3** | Install matrix clarity (stdio · HTTP · Docker Compose · TUI attach example) |
| **E4** | Operator dogfood **runbook** (ingest → retrieve → list → as-of → status) |

**Serial stamp:** **s1500** · free eng **s1504** (local E4 unit + healthz evidence) · free eng **s1509** (TUI client attach evidence) · free eng after free-floor **s1499+** · prior M3 offline SSOT **s1462** · peers TUI **s1463** · aion residual **s1464** (mention only) · free-floor peer **s1465** · free eng after **s1467+** · M4 public flip residual **s1468+** · M5 signing **s1492** · Edge Memory GA candidacy residual (aion **s1496**, mention only).

**Modules are public:** host + kernel (`github.com/iome-sh/memory`) are public MIT. Historical “still private” language on pre-flip residuals is **retired** for install paths (no `GOPRIVATE` / PAT for consumers).

---

## Honesty locks (read first — non-claims)

| Lock | Meaning |
|------|---------|
| **Edge Memory GA candidacy** | Local-primary stack candidacy only. Offline docs / gates do **not** declare **Edge Memory GA**. |
| **residual PASS ≠ invent Edge Memory GA** | Gate PASS / runbook presence ≠ Edge Memory GA declared. |
| **dual_write OFF** | Lean host does not publish audit dual_write. Default residual; do not invent ON. |
| **not Memory GA** | This binary is the **edge host only**, not bare product **Memory GA**. residual PASS ≠ invent bare Memory GA. |
| **not hosted Memory GA** | residual PASS ≠ invent hosted Memory GA / freemium Palace GA. |
| **public** | Repo is public MIT · dual_write OFF · not Memory GA · still private was pre-flip residual language retired. |
| **residual PASS ≠ live dogfood** | Offline gate / checklist PASS does **not** invent a live green dogfood run. **PASS ≠ live dogfood green.** |
| **residual PASS ≠ public flip** | Process bar + docs do not flip GH visibility or invent GHCR publish green (flip already deliberate elsewhere). |
| **residual PASS ≠ full platform sidecar parity** | Lean extract is not the private aion `aion-memory-mcp` sidecar feature set. |
| **no aion import** | Builds on `github.com/iome-sh/memory` + MCP SDK only. |
| **naming honesty** | Product edge = **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`). |
| **rates ~$88 / ~$119** | Mesh base footprint ~$88 · Memory Ops Pack ~$119/ws — mesh optional for durable pull/retain; local FS path needs neither. |
| **open boxes stay open** | Still-human APPLY / product gates (E4 live evidence · E10 founder/GTM) remain open; do not close by residual alone. |
| **Palace sunset** | Hosted Palace path remains sunset / residual; local-primary FS is the edge dogfood path. |
| **mesh optional for pull** | Mesh credentials + platform endpoint are optional; offline edge dogfood uses local Palace FS only. |
| **compose PASS ≠ public registry** | `docker compose up --build` uses **local image** `iomesh-memory-mcp:local` only. |
| **build PASS ≠ invent GA** | Binary or image build success is not Memory GA · not Edge Memory GA. |
| **gate does not need docker daemon** | `make edge-dogfood-gate` is offline file greps only (no docker, no long-running server, no gcloud). |

---

## 0. Offline residual gate (CI-friendly)

No docker daemon, no listening server, no gcloud:

```bash
make edge-dogfood-gate
# → scripts/edge_dogfood_gate.sh  (exit 0 PASS / 1 FAIL)
```

This only asserts docs + compose/Dockerfile/README/Makefile needles + host layout.
**PASS here ≠ live dogfood green.** residual PASS ≠ live dogfood · residual PASS ≠ invent Edge Memory GA.

---

## E3 — Install matrix

Supported install / attach surfaces for the **local-primary** edge host. All paths use public modules (no private PAT).

| Mode | How | When to use |
|------|-----|-------------|
| **stdio** | `./bin/iomesh-memory-mcp -palace-root … -tenant …` (default when `-http-addr` empty) | Local MCP clients (TUI, Claude Desktop, Cursor, etc.) |
| **HTTP** | `-http-addr :8080 -http-path /mcp` + `GET /healthz` | Streamable MCP over URL; health probes |
| **Docker Compose** | `docker compose up --build` → image **`iomesh-memory-mcp:local`** | Reproducible local HTTP dogfood (daemon required for this path only) |
| **TUI attach** | MCP client config `command` + `args` pointing at the binary | Product tip path with [iomesh-tui](https://github.com/iome-sh/iomesh-tui) (peer serial **s1463**, mention only) |

### Install options

```bash
# go install (public modules)
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@latest

# from clone
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make build   # → bin/iomesh-memory-mcp

# optional: pin a release tag for production (see RELEASING.md Support / version policy)
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@vX.Y.Z
```

Honesty: **build PASS ≠ invent GA** · dual_write remains OFF · not Memory GA · residual PASS ≠ invent Edge Memory GA.

### TUI / client attach example

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

Peers: TUI product tip serial **s1463** (mention only — not implemented in this repo).

---

## E4 — Operator dogfood runbook

Human (or client-driven) steps for residual-honest edge dogfood.  
**This runbook documents order of operations only — residual PASS ≠ live dogfood green.** Do not treat checklist presence as invent of a recorded green run.

**Local residual evidence (s1504 · s1509):** contemporaneous stamps are logged in
[EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md):

| Serial | Date UTC | What was observed |
|--------|----------|-------------------|
| **s1504** | **2026-08-09T06:06:22Z** | unit `go test ./internal/mcphost/` ok · HTTP healthz ok (tip `f46afe2`) |
| **s1509** | **2026-08-09T06:23:34Z** | healthz ok on `:18081` · TUI `iomesh mcp --connect` → **connected=1** **tools=6** (MCP tip `f46afe2` · TUI tip `6b3958a`) |

Honesty: residual PASS ≠ invent Edge Memory GA declared · residual PASS ≠ invent
forever product green · dual_write OFF · not bare Memory GA · not hosted Memory GA ·
**unit test path ≠ full MCP client attach dogfood** · **healthz ≠ tool round-trip over MCP JSON-RPC** ·
**attach + tools/list ≠ invent Edge Memory GA** · **attach + tools/list ≠ invent forever green full product dogfood**.

### E4.1 Build

```bash
make build   # → bin/iomesh-memory-mcp
# or: go build -o bin/iomesh-memory-mcp ./cmd/iomesh-memory-mcp
```

### E4.2 Stdio health (process start)

Default transport is **stdio** when `-http-addr` / `MEMORY_MCP_HTTP_ADDR` is empty.

```bash
export PALACE_ROOT=./data/memory-palaces
export MEMORY_TENANT=default
./bin/iomesh-memory-mcp \
  -palace-root "$PALACE_ROOT" \
  -tenant "$MEMORY_TENANT"
# log line should include: dual_write=off not_memory_ga=true mode=stdio
```

Attach via your MCP client (TUI example above). Process exit / client disconnect is normal for one-shot tool sessions depending on the client.

### E4.3 Tool round-trip order (required sequence)

Lean tool surface (kernel FS only; audited always false / dual_write off in outputs).  
Run in this order against a live client session:

| Step | Tool | Intent |
|------|------|--------|
| 1 | `memory_ingest_turn` | Ingest a user/assistant turn → Palace FS |
| 2 | `memory_retrieve` | Search / retrieve; confirm FS-backed hits under `$PALACE_ROOT/<tenant>/…` |
| 3 | `memory_list` | List with options; confirm listing reflects ingested content |
| 4 | `memory_facts_as_of` | Facts as-of; confirm as-of query surface responds |
| 5 | `memory_compact_status` | Stats (`GetStats`); payload should reflect dual_write **off** / not Memory GA residual |

Optional additional surface (not required for the E4 sequence):

| Tool | Intent |
|------|--------|
| `memory_search_semantic` | Hybrid semantic (+ residual) |
| `memory_write` | Durable fact `Write` / optional `WriteAndSupersede` (#20) · dual_write OFF |

Operator expectations when dogfooding against a client:

1. Ingest a user/assistant turn under a tenant.
2. Retrieve then list and confirm FS-backed hits under `$PALACE_ROOT/<tenant>/…`.
3. Call `memory_facts_as_of`, then `memory_compact_status` — status payload should reflect dual_write **off** / not Memory GA residual.
4. Do **not** expect aion audit dual_write publish, mesh pull, or platform sidecar parity.

**residual PASS ≠ live dogfood** — this section is the honesty SSOT for a human or client round-trip; CI does not run a live MCP session. **PASS ≠ live dogfood green.**

### E4.4 Optional HTTP /healthz

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

### E4.5 Optional Docker Compose (local image only)

```bash
docker compose up --build
# image: iomesh-memory-mcp:local  (NOT a public registry publish claim)

curl -fsS http://127.0.0.1:8080/healthz
```

| Claim | Truth |
|-------|--------|
| Local image tag | `iomesh-memory-mcp:local` |
| Public GHCR green | **Not claimed** on this serial |
| compose PASS | ≠ public registry · ≠ invent GA · ≠ dual_write ON · ≠ invent Edge Memory GA |
| Docker daemon | **Required for this optional path only** — **not** required for `make edge-dogfood-gate` |

---

## Product continuum (Option A + Edge Memory GA candidacy)

| Milestone | Status |
|-----------|--------|
| **M1** kernel TUI-grade OSS process bar (`memory`) | Prior · public |
| **M2** lean host scaffold (this repo, s1457) | Shipped · public |
| **M3** edge dogfood surfaces (this doc + gate, s1462) | Offline SSOT shipped |
| **M4** public flip (kernel first, then this host) | Deliberate · host + kernel public |
| **M5** signing / matrix / extensions (s1492) | Packaging residual (tip ≠ invent forever-green signed releases) |
| **E3** install matrix | **This serial (s1500)** — documented above |
| **E4** operator dogfood runbook | **s1500** — runbook only; residual PASS ≠ live dogfood green |
| **E4** local residual evidence | **s1504** unit + healthz · **s1509** TUI client attach — [EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md); residual PASS ≠ invent Edge Memory GA · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool RT · attach + tools/list ≠ invent Edge Memory GA / forever green full product dogfood |
| **E5** support / version policy | [RELEASING.md](../RELEASING.md) · [SUPPORT.md](../SUPPORT.md) |
| **Edge Memory GA declared** | **Not this serial** — residual PASS ≠ invent Edge Memory GA |

Peers (mention only): TUI s1463 dogfood tip · aion residual s1464 · free-floor peer s1465 · free eng after s1467+ · Edge Memory GA candidacy residual aion s1496.

---

## Rates / mesh / Palace residual (narrative only)

- **Local edge path:** TUI + this MCP host + LLM — local Palace FS; **no** Qdrant / Cloud Run palace required for lean dogfood.
- **Mesh-backed path (optional):** local stack + mesh credentials + platform endpoint + **Memory Ops Pack (~$119/ws)** for durable pull/retain; mesh base footprint separate (**~$88**). Not required for offline edge dogfood.
- **Hosted Palace sunset:** do not invent always-on hosted Palace GA; local-primary remains the edge default.
- **Open boxes stay open** until deliberate product APPLY (live E4 evidence · E10 founder/GTM · sales matrix flip).

---

## Related files

| Path | Role |
|------|------|
| [README.md](../README.md) | Quick start + install |
| [RELEASING.md](../RELEASING.md) | Tags · M5 matrix · **E5 support / version policy** |
| [SUPPORT.md](../SUPPORT.md) | Issues · security · support scope |
| [docker-compose.yml](../docker-compose.yml) | Local HTTP dogfood image |
| [Dockerfile](../Dockerfile) | Multi-stage build → `iomesh-memory-mcp` |
| [Makefile](../Makefile) | `edge-dogfood-gate` · `check` · `ci` |
| [scripts/edge_dogfood_gate.sh](../scripts/edge_dogfood_gate.sh) | Offline residual greps |
| [EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md) | **s1504** unit + healthz · **s1509** TUI client attach evidence |
| [docs/OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md) | Visibility / OSS process bar |
| [docs/PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) | M4 public-flip readiness residual |
| [CHANGELOG.md](../CHANGELOG.md) | s1462 · s1500 · s1504 · s1509 entries |

---

## M4 public-flip readiness (pointer)

M4 readiness residual lives in [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md)
(`make public-flip-readiness-gate`, serial **s1468** / **s1474**). Kernel public first, then this host — both now public MIT.
residual PASS ≠ public flip (process residual; visibility flip is a separate deliberate act).

---

## Audit one-liner (s1500 · s1504 · s1509 · prior s1462)

**E3 install matrix + E4 operator dogfood runbook shipped on public host; s1504 local residual E4 evidence (unit tools + healthz) + s1509 TUI client attach (connected=1 tools=6) in EDGE_DOGFOOD_EVIDENCE; dual_write OFF · not Memory GA · public · residual PASS ≠ live dogfood / invent Edge Memory GA / public flip invent / platform sidecar parity · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool RT · attach + tools/list ≠ invent Edge Memory GA / forever green full product dogfood · M4 readiness → PUBLIC_FLIP_READINESS · E5 → RELEASING/SUPPORT · residual PASS ≠ invent Edge Memory GA declared.**
