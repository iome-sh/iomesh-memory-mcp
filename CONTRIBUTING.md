# Contributing

Thanks for helping improve **iomesh-memory-mcp** (lean edge Memory MCP host).
Please treat quality, security, and tests as first-class.

## What this repo is

- **Edge MCP host** — stdio or streamable HTTP tools over `github.com/iome-sh/memory` Palace FS  
- **Naming** — binary/image **`iomesh-memory-mcp`**  
- Mesh audit publish (`dual_write`) stays **off** by default  
- **Does not import** private control-plane / broker packages  
- Private control plane / broker / INSTALL_STORE / billing stay out of this tree  

This repository is **public** (MIT). The kernel (`github.com/iome-sh/memory`) is
also public. `GOPRIVATE` / a GitHub token are **not** required to clone, test, or
`go install` this host. Maintainers may still set `GOPRIVATE` for other private
org modules. Maintainer process residuals
([docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md),
[docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md)) are **not**
operator how-tos — the visibility flip is already complete.

## Development setup

```bash
# Go version: see go.mod (CI uses that exact toolchain via GOTOOLCHAIN=auto)
# Public modules: no GOPRIVATE / PAT required for this host + memory kernel.
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make test
make vet
make build
```

Historical note: while the kernel was private, `GOPRIVATE` / `GONOSUMDB` were
needed for module fetch. That residual is retired for this host + kernel.

Optional:

```bash
make test-race
make cover
make vuln
make ci                         # full local gate (fmt + vet + test + vuln + build)
make edge-dogfood-gate          # offline M3 residual greps (no docker)
make public-flip-readiness-gate # offline M4 residual greps (flip complete; not operator how-to)
make release-snapshot           # local GoReleaser snapshot (needs goreleaser + syft)
```

## Coding standards

- **Lean host** — tools and HTTP surface stay small; prefer kernel APIs over re-implementing Palace  
- **Does not import private control-plane / broker packages** — do not pull private control-plane or broker modules into this tree  
- Product name is **iomesh-memory-mcp**; mesh audit publish stays off by default  
- **Fail closed** on empty palace root, omitted tool tenant, and bad paths; tenant isolation is path-based only (document residuals)  
- Prefer small, focused PRs with tests for new tool/HTTP behavior  
- Run `gofmt` (or `make fmt`) before commit  

## Tests

| Package / surface | Focus |
|-------------------|--------|
| `internal/mcphost` tools | Temp palace dirs; ingest / retrieve / list / facts_as_of |
| `internal/mcphost` HTTP | `/healthz` `dual_write` field · tools count · version stamp · loopback bind · optional shared secret |
| `internal/mcphost` DLP | ingest/write redact `ghp_` / `sk-` before palace write |
| CLI / flags | stdio vs HTTP mode selection; `MEMORY_MCP_*` only (no leftover product-plane env aliases) |

New features should include unit tests. Prefer temp dirs for Palace FS (no live
broker, no Qdrant requirement on the default path).

## Security-sensitive changes

If you touch filesystem roots, tenant path join, HTTP bind, or future auth:

1. Add/adjust tests under `internal/mcphost`  
2. Update [SECURITY.md](SECURITY.md) if the threat model changes  
3. Keep palace data out of the git tree; never commit `.env` secrets  

Report vulnerabilities privately — see [SECURITY.md](SECURITY.md). **Do not open public issues for exploits.**

## Issues & discussions

- Bugs / features: use [issue templates](https://github.com/iome-sh/iomesh-memory-mcp/issues/new/choose)  
- Support channels: [SUPPORT.md](SUPPORT.md)  
- Docs first: [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) (operator dogfood). Flip/audit/evidence files ([docs/PUBLIC_FLIP_READINESS.md](docs/PUBLIC_FLIP_READINESS.md), [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md), [docs/EDGE_DOGFOOD_EVIDENCE.md](docs/EDGE_DOGFOOD_EVIDENCE.md)) are maintainer residuals, not operator how-tos.

## Public repository policy

**This repository is public.** Keep private program material out of the tree and PR surface:

- Do **not** put private monorepo paths (private control-plane clone/build instructions), internal pending-todos, or unpublished stage URLs in PRs, docs, or CHANGELOG  
- Do **not** turn on mesh audit publish by default, or claim full platform sidecar parity with a private control-plane sidecar  
- Do **not** put private ledger serials (`s###`) in PR titles, commit subjects, or CHANGELOG user-facing notes (internal continuum stamps stay in private process only)  
- Prefer **I/O Mesh / edge Memory MCP** product language over private control-plane / broker codenames in new docs  
- Binary/image names operators run (**`iomesh-memory-mcp`**, `ghcr.io/iome-sh/iomesh-memory-mcp`) may appear when documenting install/wire-up  
- Do **not** document “clone the private control-plane monorepo” as the product edge build path  

Historical readiness residuals may still mention serial stamps; they are maintainer residuals, not operator how-tos or a public product claim. The repo is public.

## Pull requests

- Clear description of *what* and *why*  
- Link related issues  
- Ensure CI is green  
- Do not commit API keys, `.env`, or palace data  
- Update [CHANGELOG.md](CHANGELOG.md) **Unreleased** for user-visible changes  
- Keep naming **iomesh-memory-mcp**; do not import private control-plane/broker packages  
- Follow **Public repository policy** above (no private control-plane build paths)  

### CI on PR and merge

GitHub Actions workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on:

| Event | When |
|-------|------|
| `pull_request` | opened / synchronize / reopened / ready_for_review → `main` |
| `push` | commits to `main` (after merge) |
| `merge_group` | GitHub merge queue (if enabled) |
| `workflow_dispatch` | manual re-run |

Jobs: **lint** · **test** · **build** · **govulncheck** · **ci-success** (aggregate gate).

The kernel is public. CI module fetch does **not** need `IOMESH_CI_PAT` / `GH_PAT` /
`GO_MODULE_TOKEN` for this host (see `.github/workflows/ci.yml`). Historical
private-module tokens are retired here. Do not invent a token requirement.

Recommended branch protection on `main`:

1. Require a pull request before merging  
2. Require status checks to pass: **`ci-success`**  
3. Require branches to be up to date before merging  

Local parity:

```bash
make ci
```

Offline residuals (optional; not required by `ci-success`):

```bash
make edge-dogfood-gate
make public-flip-readiness-gate
```

## Architecture (lean v1)

```text
cmd/iomesh-memory-mcp
  → internal/mcphost (stdio | HTTP)
  → github.com/iome-sh/memory PalaceStore
```

Tenant layout: `filepath.Join(palaceRoot, tenant)` as Palace `BaseDir`.

## Out of scope here

- Enabling mesh audit publish / private control-plane audit by default  
- Requiring Qdrant/ONNX for default path  
- Importing private control-plane / broker packages  
- Inventing a GitHub token / `GOPRIVATE` requirement for this public host + kernel  

## License

By contributing, you agree that your contributions are licensed under the MIT License (see [LICENSE](LICENSE)).
