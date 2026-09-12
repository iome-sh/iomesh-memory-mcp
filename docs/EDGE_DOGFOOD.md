# Edge dogfood (E3 / E4)

Operator checklist for **edge dogfood** of
`github.com/iome-sh/iomesh-memory-mcp` with TUI (or any MCP client) over
**stdio** or **streamable HTTP**, plus optional local Docker Compose.

| Exit | Scope here |
|------|------------|
| **E3** | Install matrix clarity (stdio · HTTP · Docker Compose · TUI attach example) |
| **E4** | Operator dogfood **runbook** (ingest → retrieve → list → as-of → status) |

**Serial stamp:** **s1500** · free eng **s1504** (local E4 unit + healthz evidence) · free eng **s1509** (TUI client attach evidence) · free eng after free-floor **s1499+** · prior M3 offline SSOT **s1462** · peers TUI **s1463** · private control-plane residual **s1464** (mention only) · free-floor peer **s1465** · free eng after **s1467+** · M4 public flip residual **s1468+** · M5 signing **s1492**.

Serials below are historical engineering pins, not a product ledger.

**Modules are public:** host + kernel (`github.com/iome-sh/memory`) are public MIT. Historical “still private” language on pre-flip residuals is **retired** for install paths (no `GOPRIVATE` / PAT for consumers).

---

## Scope (read first)

| Topic | Meaning |
|------|---------|
| **Local-primary** | This binary is the **edge host** over local Palace FS under `PALACE_ROOT`. |
| **residual PASS ≠ live dogfood** | Offline gate / checklist PASS does **not** stand in for a live green dogfood run. **PASS ≠ live dogfood green.** |
| **residual PASS ≠ public flip** | Process bar + docs do not flip GH visibility or invent GHCR publish green (flip already deliberate elsewhere). |
| **residual PASS ≠ full platform sidecar parity** | Lean extract is not the private control-plane / broker sidecar feature set. |
| **does not import private control-plane/broker packages** | Builds on `github.com/iome-sh/memory` + MCP SDK only. |
| **naming** | Product edge = **`iomesh-memory-mcp`**. |
| **open boxes stay open** | Still-human product close / still-open product gates (E4 live evidence · E10 founder/GTM) remain open; do not close by residual alone. |
| **Palace sunset** | Hosted Palace path remains sunset / residual; local-primary FS is the edge dogfood path. |
| **mesh optional for pull** | Mesh credentials + platform endpoint are optional for durable pull/retain. Local FS path needs neither mesh nor a priced add-on. Do not invent a priced add-on SKU or a mesh base rate here. |
| **compose PASS ≠ public registry** | `docker compose up --build` uses **local image** `iomesh-memory-mcp:local` only. |
| **build PASS ≠ product close** | Binary or image build success is not a product close. |
| **gate does not need docker daemon** | `make edge-dogfood-gate` is offline file greps only (no docker, no long-running server, no gcloud). |

---

## 0. Offline residual gate (CI-friendly)

No docker daemon, no listening server, no gcloud:

```bash
make edge-dogfood-gate
# → scripts/edge_dogfood_gate.sh  (exit 0 PASS / 1 FAIL)
```

This only asserts docs + compose/Dockerfile/README/Makefile needles + host layout.
**PASS here ≠ live dogfood green.** residual PASS ≠ live dogfood.

---

## E3 — Install matrix

Supported install / attach surfaces for the **local-primary** edge host. All paths use public modules (no private PAT).

| Mode | How | When to use |
|------|-----|-------------|
| **stdio** | `./bin/iomesh-memory-mcp -palace-root … -tenant …` (default when `-http-addr` empty) | Local MCP clients (TUI, Claude Desktop, Cursor, etc.) |
| **HTTP** | `-http-addr :8080 -http-path /mcp` + `GET /healthz` (`:8080` → `127.0.0.1:8080`; optional `MEMORY_MCP_HTTP_SECRET`) | Streamable MCP over URL; health probes |
| **Docker Compose** | `docker compose up --build` → image **`iomesh-memory-mcp:local`** | Reproducible local HTTP dogfood (daemon required for this path only) |
| **TUI attach** | MCP client config `command` + `args` pointing at the binary | Product tip path with [iomesh-tui](https://github.com/iome-sh/iomesh-tui) (peer serial **s1463**, mention only) |

### Install options

```bash
# operator pin — annotated GitHub Release v0.4.1
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@v0.4.1

# from clone
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make build   # → bin/iomesh-memory-mcp (embeds git describe)

# non-pin tip only (not a production pin; @latest is also not a pin)
# go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@main
```

**build PASS ≠ product close.** Local image only.

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

Cursor / Claude Desktop / generic `mcp.json` (stdio + HTTP) lives in the README
**Other MCP clients** section (#18). No TUI required.

---

## E4 — Operator dogfood runbook

Human (or client-driven) steps for edge dogfood.  
**This runbook documents order of operations only — residual PASS ≠ live dogfood green.** Do not treat checklist presence as a recorded green run.

**Local residual evidence (s1504 · s1509):** contemporaneous stamps are logged in
[EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md) (maintainer residual — local evidence log, not operator how-to; residual PASS ≠ live dogfood):

| Serial | Date UTC | What was observed |
|--------|----------|-------------------|
| **s1504** | **2026-08-09T06:06:22Z** | unit `go test ./internal/mcphost/` ok · HTTP healthz ok (tip `f46afe2`) |
| **s1509** | **2026-08-09T06:23:34Z** | healthz ok on `:18081` · TUI `iomesh mcp --connect` → **connected=1** **tools=6** (MCP tip `f46afe2` · TUI tip `6b3958a`) — **historical attach stamp**, not a live forever-green count |

**unit test path ≠ full MCP client attach dogfood** · **healthz ≠ tool round-trip over MCP JSON-RPC** ·
**attach + tools/list ≠ product close**.

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
# log line should include: dual_write=off mode=stdio
```

Attach via your MCP client (TUI example above). Process exit / client disconnect is normal for one-shot tool sessions depending on the client.

### E4.3 Tool round-trip order (required sequence)

Lean tool surface (kernel FS only; audited always false / `dual_write` off in outputs).  
Run in this order against a live client session:

| Step | Tool | Intent |
|------|------|--------|
| 1 | `memory_ingest_turn` | Ingest a user/assistant turn → Palace FS |
| 2 | `memory_retrieve` | Search / retrieve; confirm FS-backed hits under `$PALACE_ROOT/<tenant>/…` |
| 3 | `memory_list` | List with options; confirm listing reflects ingested content |
| 4 | `memory_facts_as_of` | Facts as-of; confirm as-of query surface responds |
| 5 | `memory_compact_status` | Stats (`GetStats`); payload should reflect `dual_write` **off** |

Optional additional surface (not required for the E4 sequence):

| Tool | Intent |
|------|--------|
| `memory_search_semantic` | Hybrid semantic (+ residual) |
| `memory_write` | Durable fact `Write` / optional `WriteAndSupersede` (#20) |
| `memory_related` | `MultiHopRetrieve` (#17) · not full graph RAG |
| `memory_supersede_entity` | `SupersedeEntityFacts` (#17) · HITL at the client |
| `memory_extract_facts` | HITL extract-after-persist (parent already on disk). Optional `facts`; else kernel `ExtractAtomicFacts` on a copy. Writes `turn_fact` children; does **not** rewrite/delete the parent; **not** a PalaceStore write-gate; **not** required for E4. Structural extract, not NLP. |
| `ops_digest_export` | Local palace receipts for TUI `/memory digest` MCP fallback (#55/#66) · empty patterns OK · source-class diversity when mesh+private exist in-window · receipt `source_hint` + `provenance`/`tags` from palace (mesh only when stamped) · never invent mesh |

Operator expectations when dogfooding against a client:

1. Ingest a user/assistant turn under a tenant.
2. Retrieve then list and confirm FS-backed hits under `$PALACE_ROOT/<tenant>/…`.
3. Call `memory_facts_as_of`, then `memory_compact_status` — status payload should reflect `dual_write` **off**.
4. Do **not** expect private control-plane audit publish, mesh pull, or platform sidecar parity.

#### Optional TTFH-shaped path (not E4 required)

Kernel walking skeleton: [memory `docs/TTFH.md`](https://github.com/iome-sh/memory/blob/main/docs/TTFH.md). Cost-max: **hash embedder**, **no Qdrant**, **no cloud palace**.

After the required E4 sequence (or as extra ingest turns), operators may walk the RCA-shaped path:

1. Three RCA-shaped `memory_ingest_turn` calls (PagerDuty page: webhook ingress 5xx; HMAC-verified delivery HTTP 200 is **not** a consume receipt; `CreateConsumer` 500 when `consumers.mode` is NULL). Local overlay stays **private**; omit `source_hint` or pass `private` — **do not invent mesh**. Host process labels (`mcp_memory_ingest_turn`, `source:iomesh-memory-mcp`) are **not** a cite-both class.
2. `memory_retrieve` then `memory_facts_as_of` for the same session; print `source_hint` on hits (expect `private` on this overlay).
3. Cite-both is a **TUI session flag** (`/memory digest --require-sources mesh,private`) over `ops_digest_export` receipts — **a miss is success**. Catalog/grant is not cite-both. This host cannot mint a mesh-class receipt without a stamped mesh source — never invent mesh.

A docs PR, a green gate, or a TUI slash-command name is not a laptop PULSE run.

**residual PASS ≠ live dogfood** — this section is the runbook for a human or client round-trip; CI does not run a live MCP session. **PASS ≠ live dogfood green.**

### E4.4 Optional HTTP /healthz

```bash
./bin/iomesh-memory-mcp \
  -palace-root ./data/memory-palaces \
  -tenant default \
  -http-addr :8080 \
  -http-path /mcp
# :8080 binds 127.0.0.1:8080. 0.0.0.0 needs -allow-non-loopback.

curl -fsS http://127.0.0.1:8080/healthz
# expect JSON with:
#   "status":"ok"
#   "service":"iomesh-memory-mcp"
#   "dual_write":"off"
#   "embeddings":"hash" | "onnx"
#   "persist_embeddings":"off"  (default; on only for ONNX + MEMORY_PERSIST_EMBEDDINGS; hash never persists)
#   "qdrant":"off"
#   "tools": <compile-time lean count, currently >= 11>
#   "tool_names": [..., "memory_write", "memory_related",
#                  "memory_supersede_entity", "memory_retrieve",
#                  "memory_extract_facts", "ops_digest_export", ...]
# healthz.tools is compile-time registration, not a live MCP
#   tools/list stamp. s1509 TUI attach tools=6 at tip f46afe2 is
#   contemporaneous evidence — do not restamp as live forever-green.
```

Streamable MCP endpoint (client-dependent):

```text
http://127.0.0.1:8080/mcp
```

Env equivalents: `MEMORY_MCP_HTTP_ADDR=:8080` · `MEMORY_MCP_HTTP_PATH=/mcp`.
`:8080` is forced to `127.0.0.1:8080` unless `MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK` is set
(compose/image set that for container publish; host mapping stays `127.0.0.1:8080`).
Optional `MEMORY_MCP_HTTP_SECRET` fail-closes MCP HTTP when set; `/healthz` stays open.

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
| compose PASS | ≠ public registry · ≠ mesh audit publish on |
| Docker daemon | **Required for this optional path only** — **not** required for `make edge-dogfood-gate` |

---

## Product continuum (Option A)

| Milestone | Status |
|-----------|--------|
| **M1** kernel TUI-grade OSS process bar (`memory`) | Prior · public |
| **M2** lean host scaffold (this repo, s1457) | Shipped · public |
| **M3** edge dogfood surfaces (this doc + gate, s1462) | Offline SSOT shipped |
| **M4** public flip (kernel first, then this host) | Deliberate · host + kernel public |
| **M5** signing / matrix / extensions (s1492) | Packaging residual (tip ≠ invent forever-green signed releases) |
| **E3** install matrix | **This serial (s1500)** — documented above |
| **E4** operator dogfood runbook | **s1500** — runbook only; residual PASS ≠ live dogfood green |
| **E4** local residual evidence | **s1504** unit + healthz · **s1509** TUI client attach — [EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md) (maintainer residual); unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool RT · attach + tools/list ≠ product close |
| **E5** support / version policy | [RELEASING.md](../RELEASING.md) · [SUPPORT.md](../SUPPORT.md) |

Peers (mention only): TUI s1463 dogfood tip · private control-plane residual s1464 · free-floor peer s1465 · free eng after s1467+.

---

## Rates / mesh / Palace residual (narrative only)

- **Local edge path:** TUI + this MCP host + LLM — local Palace FS; **no** Qdrant / Cloud Run palace required for lean dogfood.
- **Mesh-backed path (optional):** local stack + mesh credentials + platform endpoint for durable pull/retain. Not required for offline edge dogfood. Do not invent a priced add-on SKU or a mesh base rate on this page.
- **Hosted Palace sunset:** hosted Palace is not the edge default; local-primary remains the path.
- **Open boxes stay open** until a deliberate human product close (live E4 evidence · E10 founder/GTM · sales matrix flip).

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
| [EDGE_DOGFOOD_EVIDENCE.md](EDGE_DOGFOOD_EVIDENCE.md) | Maintainer residual (local evidence log; not operator how-to) · **s1504** unit + healthz · **s1509** TUI client attach |
| [memory docs/TTFH.md](https://github.com/iome-sh/memory/blob/main/docs/TTFH.md) | Kernel TTFH walking skeleton (3 RCA + retrieve + facts-as-of + `source_hint`) · cost-max hash / no Qdrant / no cloud palace |
| [docs/OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md) | Maintainer OSS process residual (not operator how-to) |
| [docs/PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md) | M4 public-flip maintainer residual (flip complete; not operator how-to) |
| [CHANGELOG.md](../CHANGELOG.md) | s1462 · s1500 · s1504 · s1509 entries |

---

## M4 public-flip readiness (pointer)

M4 readiness residual lives in [PUBLIC_FLIP_READINESS.md](PUBLIC_FLIP_READINESS.md)
(`make public-flip-readiness-gate`, serial **s1468** / **s1474**). Kernel public first, then this host — both now public MIT.
residual PASS ≠ public flip (process residual; visibility flip is a separate deliberate act).

---

## Audit one-liner (s1500 · s1504 · s1509 · prior s1462)

**E3 install matrix + E4 operator dogfood runbook shipped on public host; s1504 local residual E4 evidence (unit tools + healthz) + s1509 TUI client attach (connected=1 tools=6) in EDGE_DOGFOOD_EVIDENCE; public · residual PASS ≠ live dogfood / public flip / platform sidecar parity · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool RT · attach + tools/list ≠ product close · M4 readiness → PUBLIC_FLIP_READINESS · E5 → RELEASING/SUPPORT.**
