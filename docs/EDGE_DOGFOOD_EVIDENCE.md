# Edge dogfood evidence log (E4 residual)

**Not current operator documentation.** This file is a **maintainer residual** — a **local evidence log**, **not** a product claim. **residual PASS ≠ live dogfood.** Operators follow the E4 runbook in [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md).

Contemporaneous **local residual dogfood evidence** for **E4** progress on
`github.com/iome-sh/iomesh-memory-mcp`.

**Serial stamp:** **s1509** · free eng after **s1504** (local E4 unit + healthz evidence) ·
**s1500** (E3/E4 runbook · E5 support) · prior M3 offline SSOT **s1462**.

This file records **observed** local runs only. It does **not** declare a product close.
Parent runbook: [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md).

---

## Scope (read first)

| Topic | Meaning |
|------|---------|
| **local residual dogfood evidence** | Logged laptop / CI-adjacent runs for E4 **progress** only. |
| **residual PASS ≠ product close** | Evidence PASS / green unit / healthz / client attach ≠ product close. |
| **residual PASS ≠ invent forever product green** | One local stamp is not forever-green product status. |
| **unit test path ≠ full MCP client attach dogfood** | `go test ./internal/mcphost/` exercises tool **handlers**; it is **not** a full MCP client attach / JSON-RPC session dogfood. |
| **healthz ≠ tool round-trip over MCP JSON-RPC** | `GET /healthz` proves HTTP process start; it is **not** an ingest→retrieve→list→as-of→status tool round-trip over MCP. Tool path is proven in unit tests (separate row). |
| **attach + tools/list ≠ product close** | TUI/`iomesh mcp --connect` with `connected=1` and tools/list count proves **client attach + tool discovery** only. |
| **attach + tools/list ≠ invent forever green full product dogfood** | Client attach + listing six tools is **not** a forever-green full product dogfood (ingest→retrieve→list→as-of→status live RT over the client session is a separate claim; not asserted as forever green here). |
| **offline gate ≠ this evidence** | `make edge-dogfood-gate` remains offline file greps only; it does not re-run these live steps. |

Open boxes stay open until a deliberate human product close (full E4 tool RT dogfood closeout · E10 founder/GTM · sales matrix flip).

---

## Evidence stamp (s1504 — unit + healthz)

| Field | Value |
|-------|--------|
| **Date (UTC)** | **2026-08-09T06:06:22Z** |
| **Git tip** | **`f46afe2462ebaa94890b30296b1a19d03d6853da`** (`main` at stamp) |
| **Short SHA** | `f46afe2` |
| **Repo** | `github.com/iome-sh/iomesh-memory-mcp` |
| **Scope** | Local residual E4 progress — **not** a product close |

---

## 1. Tool path (unit) — s1504

**Command:**

```bash
go test ./internal/mcphost/ -count=1 -timeout 120s
```

**Result:** **ok** (~0.533s)

**What this exercises:** ingest / retrieve / list / facts / compact tool handlers
(see `internal/mcphost/tools_test.go`).

**Notes:**

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

**Logs (observed):** `dual_write=off` · `mode=http`

**Notes:**

- healthz confirms process start, service name, `dual_write` **off**, version stamp.
- **healthz ≠ tool round-trip over MCP JSON-RPC** — no `memory_ingest_turn` /
  `memory_retrieve` / `memory_list` / `memory_facts_as_of` / `memory_compact_status`
  over the MCP endpoint in this row (those handlers are covered by the unit path above).

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
| **Scope** | Local residual E4 **client attach** progress — **not** a product close |

---

## 3. HTTP healthz (client-attach host) — s1509

**Observed on:** `127.0.0.1:18081` (local binary at MCP tip above).

**Result:** **healthz OK** · `dual_write=off`

**Notes:**

- Confirms the host process used for the TUI attach stamp was healthy.
- **healthz ≠ tool round-trip over MCP JSON-RPC** (unchanged).

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

**Notes:**

- **attach + tools/list ≠ product close** — connect success and tool count do not close the product.
- **attach + tools/list ≠ invent forever green full product dogfood** — this stamp does **not** claim a forever-green full product dogfood, nor a complete live E4 ingest→retrieve→list→as-of→status tool round-trip over the client session as product-closed.
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
| Forever product green | **No** — residual PASS / attach ≠ invent forever product green |
| Mesh audit publish on | **No** — `dual_write` **off** |

---

## Related

| Path | Role |
|------|------|
| [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) | E3 install matrix · E4 runbook |
| [scripts/edge_dogfood_gate.sh](../scripts/edge_dogfood_gate.sh) | Offline residual greps (not this live stamp) |
| `internal/mcphost/tools_test.go` | Unit coverage for tool handlers |
| [CHANGELOG.md](../CHANGELOG.md) | Unreleased / serial notes |

---

## Audit one-liner (s1504 · s1509)

**Local residual E4 dogfood evidence · s1504 2026-08-09T06:06:22Z unit tools + healthz ok · s1509 2026-08-09T06:23:34Z tip MCP f46afe2 · TUI 6b3958a · healthz ok on :18081 (`dual_write` off) · TUI `iomesh mcp --connect` connected=1 tools=6 · residual PASS ≠ invent forever product green · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool round-trip · attach + tools/list ≠ forever green full product dogfood.**
