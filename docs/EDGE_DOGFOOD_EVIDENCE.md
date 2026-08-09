# Edge dogfood evidence log (E4 residual)

Contemporaneous **local residual dogfood evidence** for **E4** progress on
`github.com/iome-sh/iomesh-memory-mcp`.

**Serial stamp:** **s1509** · free eng after **s1504** (local E4 unit + healthz evidence) ·
**s1500** (E3/E4 runbook · E5 support) · prior M3 offline SSOT **s1462** · Edge Memory GA
candidacy residual (aion **s1496**, mention only).

This file records **observed** local runs only. It does **not** declare product GA.
Parent runbook: [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md).

---

## Honesty locks (read first — non-claims)

| Lock | Meaning |
|------|---------|
| **local residual dogfood evidence** | Logged laptop / CI-adjacent runs for E4 **progress** only. |
| **residual PASS ≠ invent Edge Memory GA declared** | Evidence PASS / green unit / healthz / client attach ≠ **Edge Memory GA** declared. |
| **residual PASS ≠ invent forever product green** | One local stamp is not forever-green product status. |
| **dual_write OFF** | Observed payloads and process logs show dual_write **off**. Do not invent ON. |
| **not bare Memory GA** | Edge host only — residual PASS ≠ invent bare product **Memory GA**. |
| **not hosted Memory GA** | residual PASS ≠ invent hosted Memory GA / freemium Palace GA. |
| **unit test path ≠ full MCP client attach dogfood** | `go test ./internal/mcphost/` exercises tool **handlers**; it is **not** a full MCP client attach / JSON-RPC session dogfood. |
| **healthz ≠ tool round-trip over MCP JSON-RPC** | `GET /healthz` proves HTTP process + honesty fields; it is **not** an ingest→retrieve→list→as-of→status tool round-trip over MCP. Tool path is proven in unit tests (separate row). |
| **attach + tools/list ≠ invent Edge Memory GA** | TUI/`iomesh mcp --connect` with `connected=1` and tools/list count proves **client attach + tool discovery** only — not Edge Memory GA declared. |
| **attach + tools/list ≠ invent forever green full product dogfood** | Client attach + listing six tools is **not** a forever-green full product dogfood (ingest→retrieve→list→as-of→status live RT over the client session is a separate claim; not asserted as forever green here). |
| **offline gate ≠ this evidence** | `make edge-dogfood-gate` remains offline file greps only; it does not re-run these live steps. |

Open boxes stay open until deliberate product APPLY (full E4 tool RT dogfood closeout · E10 founder/GTM · sales matrix flip).

---

## Evidence stamp (s1504 — unit + healthz)

| Field | Value |
|-------|--------|
| **Date (UTC)** | **2026-08-09T06:06:22Z** |
| **Git tip** | **`f46afe2462ebaa94890b30296b1a19d03d6853da`** (`main` at stamp) |
| **Short SHA** | `f46afe2` |
| **Repo** | `github.com/iome-sh/iomesh-memory-mcp` |
| **Scope** | Local residual E4 progress — **not** Edge Memory GA declared |

---

## 1. Tool path (unit) — s1504

**Command:**

```bash
go test ./internal/mcphost/ -count=1 -timeout 120s
```

**Result:** **ok** (~0.533s)

**What this exercises:** ingest / retrieve / list / facts / compact tool handlers
(see `internal/mcphost/tools_test.go`).

**Honesty:**

- Unit path proves handler wiring against the lean host surface.
- **unit test path ≠ full MCP client attach dogfood** — no TUI/Claude/Cursor attach,
  no streamable MCP JSON-RPC client session in this row.
- Tool round-trip **order** for operators remains the E4 runbook in
  [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) (§ E4.3).

---

## 2. HTTP healthz (local binary) — s1504

**Build:** `bin/iomesh-memory-mcp` (from git tip above).

**Run:**

```bash
./bin/iomesh-memory-mcp \
  -palace-root <tmp>/palace \
  -tenant e4dogfood \
  -http-addr 127.0.0.1:18080 \
  -http-path /mcp

curl -fsS http://127.0.0.1:18080/healthz
```

**Response (observed):**

```json
{"status":"ok","service":"iomesh-memory-mcp","dual_write":"off","not_memory_ga":true,"version":"f46afe2"}
```

**Logs (observed):** `dual_write=off` · `not_memory_ga=true` · `mode=http`

**Honesty:**

- healthz confirms process start, service name, dual_write **off**, not_memory_ga residual, version stamp.
- **healthz ≠ tool round-trip over MCP JSON-RPC** — no `memory_ingest_turn` /
  `memory_retrieve` / `memory_list` / `memory_facts_as_of` / `memory_compact_status`
  over the MCP endpoint in this row (those handlers are covered by the unit path above).
- dual_write remains **OFF** · not bare Memory GA · not hosted Memory GA.

---

## Evidence stamp (s1509 — TUI client attach)

| Field | Value |
|-------|--------|
| **Date (UTC)** | **2026-08-09T06:23:34Z** |
| **MCP git tip** | **`f46afe2462ebaa94890b30296b1a19d03d6853da`** |
| **MCP short SHA** | `f46afe2` |
| **TUI git tip** | **`6b3958a90a01d2c8f50ee161c8dc1009637b64f1`** |
| **TUI short SHA** | `6b3958a` |
| **Repos** | `github.com/iome-sh/iomesh-memory-mcp` · peer client `github.com/iome-sh/iomesh-tui` |
| **Scope** | Local residual E4 **client attach** progress — **not** Edge Memory GA declared |

---

## 3. HTTP healthz (client-attach host) — s1509

**Observed on:** `127.0.0.1:18081` (local binary at MCP tip above).

**Result:** **healthz OK** · `dual_write=off` · `not_memory_ga=true`

**Honesty:**

- Confirms the host process used for the TUI attach stamp was healthy with residual-honest fields.
- **healthz ≠ tool round-trip over MCP JSON-RPC** (unchanged).
- dual_write remains **OFF** · not bare Memory GA · not hosted Memory GA.

---

## 4. TUI client attach + tools/list — s1509

**Client:** [iomesh-tui](https://github.com/iome-sh/iomesh-tui) at tip `6b3958a`.

**Command (observed):**

```bash
iomesh mcp --connect
```

**Result (observed):**

| Field | Value |
|-------|--------|
| **connected** | **1** |
| **tools** | **6** |

**Tools listed (exactly):**

1. `memory_ingest_turn`
2. `memory_retrieve`
3. `memory_search_semantic`
4. `memory_list`
5. `memory_compact_status`
6. `memory_facts_as_of`

**What this exercises:** MCP client attach to the local edge host and **tools/list**
discovery of the lean six-tool surface.

**Honesty:**

- **attach + tools/list ≠ invent Edge Memory GA** — connect success and tool count do not declare Edge Memory GA.
- **attach + tools/list ≠ invent forever green full product dogfood** — this stamp does **not** claim a forever-green full product dogfood, nor a complete live E4 ingest→retrieve→list→as-of→status tool round-trip over the client session as product-closed.
- dual_write remains **OFF** · not bare Memory GA · not hosted Memory GA.
- Operator tool RT **order** for a fuller session remains [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) § E4.3.

---

## What this does / does not claim

| Claim | Status |
|-------|--------|
| Local residual E4 progress evidence logged | **Yes** (this file, s1504 + s1509) |
| E4 operator runbook present | **Yes** ([EDGE_DOGFOOD.md](EDGE_DOGFOOD.md), s1500) |
| Tool handlers unit-ok at tip `f46afe2` | **Yes** (row 1, s1504) |
| HTTP healthz ok at tip `f46afe2` | **Yes** (row 2 s1504 · row 3 s1509 on :18081) |
| TUI client attach + tools/list (connected=1, tools=6) | **Yes** (row 4, s1509) — residual attach evidence only |
| Full live E4 tool RT over MCP client session (ingest→…→status) as product-closed | **Not claimed** as forever-green product dogfood |
| MCP JSON-RPC tool round-trip over `/mcp` (unit handlers yes) | Handlers via unit; live client RT not asserted closed here |
| **Edge Memory GA declared** | **No** — residual PASS / attach ≠ invent Edge Memory GA |
| Forever product green | **No** — residual PASS / attach ≠ invent forever product green |
| dual_write ON | **No** — dual_write **OFF** |
| Bare Memory GA / hosted Memory GA | **No** |

---

## Related

| Path | Role |
|------|------|
| [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) | E3 install matrix · E4 runbook · honesty locks |
| [scripts/edge_dogfood_gate.sh](../scripts/edge_dogfood_gate.sh) | Offline residual greps (not this live stamp) |
| `internal/mcphost/tools_test.go` | Unit coverage for tool handlers |
| [CHANGELOG.md](../CHANGELOG.md) | Unreleased / serial notes |

---

## Audit one-liner (s1504 · s1509)

**Local residual E4 dogfood evidence · s1504 2026-08-09T06:06:22Z unit tools + healthz ok · s1509 2026-08-09T06:23:34Z tip MCP f46afe2 · TUI 6b3958a · healthz ok on :18081 (dual_write off · not_memory_ga) · TUI `iomesh mcp --connect` connected=1 tools=6 · residual PASS ≠ invent Edge Memory GA declared · residual PASS ≠ invent forever product green · dual_write OFF · not bare Memory GA · not hosted Memory GA · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool round-trip · attach + tools/list ≠ invent Edge Memory GA / forever green full product dogfood.**
