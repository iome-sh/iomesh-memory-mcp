# Edge dogfood evidence log (E4 residual)

Contemporaneous **local residual dogfood evidence** for **E4** progress on
`github.com/iome-sh/iomesh-memory-mcp`.

**Serial stamp:** **s1504** · free eng after **s1500** (E3/E4 runbook · E5 support)
· prior M3 offline SSOT **s1462** · Edge Memory GA candidacy residual (aion **s1496**,
mention only).

This file records **observed** local runs only. It does **not** declare product GA.
Parent runbook: [EDGE_DOGFOOD.md](EDGE_DOGFOOD.md).

---

## Honesty locks (read first — non-claims)

| Lock | Meaning |
|------|---------|
| **local residual dogfood evidence** | Logged laptop / CI-adjacent runs for E4 **progress** only. |
| **residual PASS ≠ invent Edge Memory GA declared** | Evidence PASS / green unit / healthz ≠ **Edge Memory GA** declared. |
| **residual PASS ≠ invent forever product green** | One local stamp is not forever-green product status. |
| **dual_write OFF** | Observed payloads and process logs show dual_write **off**. Do not invent ON. |
| **not bare Memory GA** | Edge host only — residual PASS ≠ invent bare product **Memory GA**. |
| **not hosted Memory GA** | residual PASS ≠ invent hosted Memory GA / freemium Palace GA. |
| **unit test path ≠ full MCP client attach dogfood** | `go test ./internal/mcphost/` exercises tool **handlers**; it is **not** a full MCP client attach / JSON-RPC session dogfood. |
| **healthz ≠ tool round-trip over MCP JSON-RPC** | `GET /healthz` proves HTTP process + honesty fields; it is **not** an ingest→retrieve→list→as-of→status tool round-trip over MCP. Tool path is proven in unit tests (separate row). |
| **offline gate ≠ this evidence** | `make edge-dogfood-gate` remains offline file greps only; it does not re-run these live steps. |

Open boxes stay open until deliberate product APPLY (full client attach dogfood · E10 founder/GTM · sales matrix flip).

---

## Evidence stamp

| Field | Value |
|-------|--------|
| **Date (UTC)** | **2026-08-09T06:06:22Z** |
| **Git tip** | **`f46afe2462ebaa94890b30296b1a19d03d6853da`** (`main` at stamp) |
| **Short SHA** | `f46afe2` |
| **Repo** | `github.com/iome-sh/iomesh-memory-mcp` |
| **Scope** | Local residual E4 progress — **not** Edge Memory GA declared |

---

## 1. Tool path (unit)

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

## 2. HTTP healthz (local binary)

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

## What this does / does not claim

| Claim | Status |
|-------|--------|
| Local residual E4 progress evidence logged | **Yes** (this file, s1504) |
| E4 operator runbook present | **Yes** ([EDGE_DOGFOOD.md](EDGE_DOGFOOD.md), s1500) |
| Tool handlers unit-ok at tip `f46afe2` | **Yes** (row 1) |
| HTTP healthz ok at tip `f46afe2` | **Yes** (row 2) |
| Full MCP client attach live dogfood (TUI/etc.) | **Not claimed** on this serial |
| MCP JSON-RPC tool round-trip over `/mcp` | **Not claimed** here (handlers via unit) |
| **Edge Memory GA declared** | **No** — residual PASS ≠ invent Edge Memory GA |
| Forever product green | **No** — residual PASS ≠ invent forever product green |
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

## Audit one-liner (s1504)

**Local residual E4 dogfood evidence at 2026-08-09T06:06:22Z · tip f46afe2 · unit tools ok · healthz ok (dual_write off · not_memory_ga) · residual PASS ≠ invent Edge Memory GA declared · residual PASS ≠ invent forever product green · dual_write OFF · not bare Memory GA · not hosted Memory GA · unit ≠ full MCP client attach · healthz ≠ MCP JSON-RPC tool round-trip.**
