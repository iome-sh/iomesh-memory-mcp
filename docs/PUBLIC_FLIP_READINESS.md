# M4 public-flip readiness (residual SSOT)

Residual-honest checklist for a **deliberate** GitHub visibility flip of
`github.com/iome-sh/iomesh-memory-mcp` (lean edge Memory MCP host) under Option A.

**This serial does not flip visibility.** It documents pre-flight, post-flip residual
steps, and offline gate needles only.

**Serial stamp:** **s1468** · free eng concurrent **s1467+** after free-floor
**s1465** · lag **s1466** · peers memory **s1467** · TUI **s1469** · aion residual
**s1470** (mention only) · free-floor peer **s1471** · free eng after **s1473+**.

---

## Honesty locks (read first — non-claims)

| Lock | Meaning |
|------|---------|
| **dual_write OFF** | Lean host does not publish audit dual_write. Default residual; do not invent ON. |
| **not Memory GA** | This binary is the **edge host only**, not product Memory GA. |
| **still private** | Repo remains private until a deliberate **M4** visibility flip (not this serial). |
| **residual PASS ≠ public flip** | Offline gate / checklist PASS does **not** flip GH visibility or invent GHCR publish green. |
| **residual PASS ≠ full platform sidecar parity** | Lean extract is not the private aion `aion-memory-mcp` sidecar feature set. |
| **no aion import** | Builds on `github.com/iome-sh/memory` + MCP SDK only. |
| **naming honesty** | Product edge = **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`). |
| **offline dogfood tip ≠ invent live dogfood green** | M3 offline SSOT / `edge-dogfood-gate` PASS is not a live dogfood claim. |
| **compose PASS ≠ public registry** | Local image `iomesh-memory-mcp:local` only until deliberate GHCR publish. |
| **aion stays private** | Broker / CP / monorepo remain private; edge OSS does not open aion. |
| **open boxes stay open** | Still-human APPLY / product gates remain open; do not close by residual alone. |

---

## Flip order (hard lock)

| Step | Repo | Status on s1468 |
|------|------|-----------------|
| **1. Kernel first** | `github.com/iome-sh/memory` | **Wait** — must be public (or explicitly approved to flip same wave) **before** this host |
| **2. Then this host** | `github.com/iome-sh/iomesh-memory-mcp` | **Still private** — readiness residual only |
| aion | private monorepo | **Stays private** (not in M4 open set) |

**Do not** flip this MCP host public while the kernel module remains private: module
consumers cannot `go get` a public host that depends on a private kernel without
tokens, and that breaks the OSS story.

---

## 0. Offline residual gate (CI-friendly)

No docker daemon, no listening server, no gcloud, **no** GitHub Settings mutation:

```bash
make public-flip-readiness-gate
# → scripts/public_flip_readiness_gate.sh  (exit 0 PASS / 1 FAIL)
```

This only asserts docs + LICENSE/SECURITY/audit + honesty needles + still-private
wording + kernel-first order + dual_write OFF + serial **s1468**.

**PASS here ≠ public flip.**

Related offline residual (M3, unchanged):

```bash
make edge-dogfood-gate   # residual PASS ≠ live dogfood green · ≠ public flip
```

---

## 1. Pre-flight (before any visibility change)

Re-run and resolve residuals **before** Settings → Change visibility:

| Check | Residual on s1468 |
|-------|-------------------|
| [docs/OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md) re-run | Process bar present; re-audit immediately before flip |
| LICENSE (MIT) · NOTICE · SECURITY · CONTRIBUTING · CoC · SUPPORT · RELEASING · CHANGELOG | Present (M2 bar) |
| CI (`ci-success`: fmt · test · build · govulncheck) | Present; keep green |
| **Private dep residual** until kernel public | `github.com/iome-sh/memory` still private on this serial → `GOPRIVATE` / module token residual |
| **CI token residual** | While kernel private, Actions need `GO_MODULE_TOKEN` (or org access) with `repo` read on `iome-sh/memory` |
| Honesty locks intact in README / audit / this doc | dual_write OFF · not Memory GA · still private · no aion import · naming **iomesh-memory-mcp** |
| Kernel public first | **Blocked** until `memory` is public (or same-wave deliberate dual flip with kernel leading) |
| No secrets / palace data in tree | Operator discipline |
| GHCR publish green | **Not claimed** — separate deliberate act after (or with) visibility flip |

---

## 2. Post-flip steps (future — residual-honest; not executed on s1468)

When maintainers deliberately flip (after kernel public + re-audit):

1. Confirm `github.com/iome-sh/memory` is **public**.
2. GitHub → Settings → Change repository visibility → **Public**.
3. Enable **Private vulnerability reporting** (SECURITY.md path).
4. Branch protection on `main`: require PR + status check **`ci-success`**.
5. Drop or narrow `GOPRIVATE` / CI module-token residual for the now-public kernel dep.
6. Optional: publish GHCR image as **`ghcr.io/iome-sh/iomesh-memory-mcp` only**  
   (do **not** invent product edge as `aion-memory-mcp`; compose local tag remains for dogfood).
7. Re-run `make public-flip-readiness-gate` + `make edge-dogfood-gate` + `make ci` on the post-flip tip.
8. Update OPEN_SOURCE_AUDIT visibility row to **Public** and note flip date/PR.
9. Do **not** invent Memory GA, dual_write ON, live dogfood green by flip alone, or aion import.

---

## 3. Product continuum (Option A)

| Milestone | Status on s1468 |
|-----------|-----------------|
| **M1** kernel TUI-grade OSS process bar (`memory`) | Prior |
| **M2** lean host scaffold (this repo, s1457) | Shipped private |
| **M3** edge dogfood surfaces (s1462) | Shipped private · offline SSOT |
| **M4** public flip (kernel first, then this host) | **This serial** — readiness residual only; **still private** |
| **M5** signing / matrix / extensions | Later |

Peers (mention only): memory **s1467** · TUI **s1469** · aion residual **s1470** ·
free-floor peer **s1471** · free eng **s1473+**.

---

## 4. Related files

| Path | Role |
|------|------|
| [README.md](../README.md) | Quick start + M4 readiness link |
| [docs/OPEN_SOURCE_AUDIT.md](OPEN_SOURCE_AUDIT.md) | Visibility / OSS process bar |
| [docs/EDGE_DOGFOOD.md](EDGE_DOGFOOD.md) | M3 offline dogfood SSOT |
| [Makefile](../Makefile) | `public-flip-readiness-gate` · `edge-dogfood-gate` · `ci` |
| [scripts/public_flip_readiness_gate.sh](../scripts/public_flip_readiness_gate.sh) | Offline residual greps (this gate) |
| [scripts/edge_dogfood_gate.sh](../scripts/edge_dogfood_gate.sh) | M3 offline greps |
| [CHANGELOG.md](../CHANGELOG.md) | s1468 entry |
| [LICENSE](../LICENSE) · [SECURITY.md](../SECURITY.md) | Process bar artifacts |

---

## Audit one-liner (s1468)

**M4 public-flip readiness docs + offline gate shipped; wait for kernel public first · dual_write OFF · not Memory GA · still private · residual PASS ≠ public flip / platform sidecar parity / live dogfood invent · no aion import · naming iomesh-memory-mcp · compose PASS ≠ public registry.**
