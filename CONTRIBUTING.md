# Contributing

Thanks for helping improve **iomesh-memory-mcp** (lean edge Memory MCP host).
Please treat quality, security, honesty locks, and tests as first-class.

## What this repo is

- **Edge MCP host** — stdio or streamable HTTP tools over `github.com/iome-sh/memory` Palace FS  
- **Naming honesty** — binary/image **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`)  
- **dual_write OFF** by default · **not product Memory GA**  
- **Does not import** `github.com/iome-sh/aion/**`  
- Private aion broker / CP / INSTALL_STORE / billing stay out of this tree  

This repository remains **private** until a deliberate visibility flip **after** the
kernel (`github.com/iome-sh/memory`) is public. Do not assume public contribution
workflows until that flip lands. Readiness residual PASS ≠ public flip.

## Development setup

```bash
# Go version: see go.mod (CI uses that exact toolchain via GOTOOLCHAIN=auto)
#
# While github.com/iome-sh/memory is still private, set GOPRIVATE so module fetch
# uses git credentials (SSH or a PAT with repo read on iome-sh/memory):
git clone https://github.com/iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make test
make vet
make build
```

**After the kernel is public**, `GOPRIVATE` / `GONOSUMDB` for `iome-sh/*` become
optional for public consumers (module proxy + checksum DB can resolve
`github.com/iome-sh/memory` without a token). Maintainers may still keep
`GOPRIVATE` for other private org modules.

Optional:

```bash
make test-race
make cover
make vuln
make ci                         # full local gate (fmt + vet + test + vuln + build)
make edge-dogfood-gate          # offline M3 residual greps (no docker)
make public-flip-readiness-gate # offline M4 readiness greps (no visibility flip)
make release-snapshot           # local GoReleaser snapshot (needs goreleaser + syft)
```

## Coding standards

- **Lean host** — tools and HTTP surface stay small; prefer kernel APIs over re-implementing Palace  
- **No aion imports** — do not pull `github.com/iome-sh/aion/**` into this tree  
- **Honesty locks** — dual_write OFF by default; do not invent Memory GA; product name is **iomesh-memory-mcp**  
- **Fail closed** on empty palace root and bad paths; tenant isolation is path-based only (document residuals)  
- Prefer small, focused PRs with tests for new tool/HTTP behavior  
- Run `gofmt` (or `make fmt`) before commit  

## Tests

| Package / surface | Focus |
|-------------------|--------|
| `internal/mcphost` tools | Temp palace dirs; ingest / retrieve / list / facts_as_of honesty |
| `internal/mcphost` HTTP | `/healthz` dual_write=off · not_memory_ga · tools count · version stamp |
| CLI / flags | stdio vs HTTP mode selection; env aliases residual |

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
- Docs first: [docs/](docs/) (EDGE_DOGFOOD · PUBLIC_FLIP_READINESS · OPEN_SOURCE_AUDIT)

## Public repository policy

**Forward policy after the deliberate public flip** (repo is still private today).
When public, keep private program material out of the tree and PR surface:

- Do **not** put private monorepo paths (`aion/**` clone/build instructions), internal pending-todos, or unpublished stage URLs in PRs, docs, or CHANGELOG  
- Do **not** invent **Memory GA**, dual_write ON by default, or full platform sidecar parity with private `aion-memory-mcp`  
- After public flip, **strip private ledger serials** (`s###`) from PR titles, commit subjects, and CHANGELOG user-facing notes (internal continuum stamps stay in private process only)  
- Prefer **I/O Mesh / edge Memory MCP** product language over private aion codenames in new docs  
- Binary/image names operators run (**`iomesh-memory-mcp`**, `ghcr.io/iome-sh/iomesh-memory-mcp`) may appear when documenting install/wire-up  
- Do **not** document “clone the private aion monorepo” as the product edge build path  

While still private, serial stamps may appear in readiness residuals; they are not a public claim.

## Pull requests

- Clear description of *what* and *why*  
- Link related issues  
- Ensure CI is green  
- Do not commit API keys, `.env`, or palace data  
- Update [CHANGELOG.md](CHANGELOG.md) **Unreleased** for user-visible changes  
- Keep honesty locks intact (`dual_write=off`, `not_memory_ga`, naming **iomesh-memory-mcp**, no aion import)  
- Follow **Public repository policy** above (no private aion build paths; no invent Memory GA)  

### CI on PR and merge

GitHub Actions workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on:

| Event | When |
|-------|------|
| `pull_request` | opened / synchronize / reopened / ready_for_review → `main` |
| `push` | commits to `main` (after merge) |
| `merge_group` | GitHub merge queue (if enabled) |
| `workflow_dispatch` | manual re-run |

Jobs: **lint** · **test** · **build** · **govulncheck** · **ci-success** (aggregate gate).

While the kernel is private, CI needs `IOMESH_CI_PAT` / `GH_PAT` / `GO_MODULE_TOKEN`
(or org access) with `repo` read on `iome-sh/memory`. **After the kernel is public**,
module fetch is public and that PAT is optional — keep working private-module config
until the kernel flip lands.

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

- Enabling dual_write / aion audit by default  
- Inventing Memory GA  
- Requiring Qdrant/ONNX for default path  
- Importing private aion packages  
- Flipping the repository public (separate deliberate act; **kernel first**)  

## License

By contributing, you agree that your contributions are licensed under the MIT License (see [LICENSE](LICENSE)).
