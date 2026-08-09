# Releasing

Ship from `main` via PR; cut annotated semver tags for binary/image consumers.

## When to bump and tag

**Do not leave feature waves only under `[Unreleased]`.** After merging a coherent
minor/major capability set, cut a release in the same delivery loop (or immediately after):

| Trigger | Bump | Examples |
|---------|------|----------|
| New MCP tools / CLI flags / HTTP surface | **minor** within `v0.x` while pre-1.0 | `memory_list`, HTTP path |
| Breaking tool schemas | **minor** (pre-1.0 honesty) or document migration | Rename tool input fields |
| Docs-only / OSS process bar | usually **no** tag | SECURITY, OPEN_SOURCE_AUDIT, readiness residual |
| Security fix | **patch** | CVE follow-up |

Checklist items that must move with the tag:

- [CHANGELOG.md](CHANGELOG.md) `[Unreleased]` → `## [X.Y.Z]`  
- [README.md](README.md) version / install / healthz example line  
- [SECURITY.md](SECURITY.md) supported-versions table when versions are published  
- Default `ServerVersion` in `internal/mcphost` (ldflags override on release builds)

## Checklist before a public tag

1. [ ] `make ci` green locally  
2. [ ] GitHub Actions **ci-success** green on the release commit  
3. [ ] [CHANGELOG.md](CHANGELOG.md) updated (move Unreleased → version section)  
4. [ ] No secrets or palace data in tree  
5. [ ] Honesty locks intact: dual_write OFF · not Memory GA · no aion import · naming **iomesh-memory-mcp**  
6. [ ] **Kernel public prerequisite met:** `github.com/iome-sh/memory` is public — no `GOPRIVATE` / PAT for consumers or release CI  
7. [ ] Default `ServerVersion` string matches the tag family (GoReleaser ldflags set `v{{.Version}}`)  
8. [ ] Annotated tag `vX.Y.Z` pushed (GoReleaser **release** workflow green; assets on GitHub Release) — **no auto-tag**; maintainers cut tags deliberately

## Tag and publish (maintainers)

```bash
git checkout main
git pull origin main
# edit CHANGELOG.md + ServerVersion default + README/SECURITY if needed
git commit -am "chore: release vX.Y.Z"
git push origin main

# Annotated tag — triggers .github/workflows/release.yml (GoReleaser)
# Do not auto-tag from CI; maintainers push tags deliberately.
git tag -a vX.Y.Z -m "vX.Y.Z — short release title"
git push origin vX.Y.Z
```

`make build` embeds `git describe` (or `VERSION=`) into the binary via
`-X github.com/iome-sh/iomesh-memory-mcp/internal/mcphost.ServerVersion=…`.  
GoReleaser ldflags set `ServerVersion` to `v` + tag version on published assets.

### GoReleaser (binaries)

| Piece | Path |
|-------|------|
| Config | [`.goreleaser.yaml`](.goreleaser.yaml) |
| Workflow | [`.github/workflows/release.yml`](.github/workflows/release.yml) (on `v*` tags) |
| Local dry-run | `make release-snapshot` (needs `goreleaser` + `syft` on PATH for SBOM) |

Cross-builds: linux/darwin/windows × amd64/arm64 (windows/arm64 ignored). Archives include
LICENSE + README + CHANGELOG; `checksums.txt` + per-archive **SPDX SBOM**
(`*.sbom.spdx.json`) attach to the GitHub Release.

**Signing (keyless cosign):** tag releases sign `checksums.txt` with **cosign sign-blob**
via GitHub OIDC / Fulcio (`id-token: write`). No long-lived `COSIGN_*` secrets.
Snapshot / `workflow_dispatch` runs use `--skip=sign`.

**Public modules:** host + kernel (`github.com/iome-sh/memory`) are public. Release CI
fetches modules without `GOPRIVATE` / private PAT (aligned with `ci.yml` after public flip).
Historical `IOMESH_CI_PAT` residual is retired for release builds.

Verify:

```bash
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp 'https://github.com/iome-sh/iomesh-memory-mcp/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

## M5 signing / matrix (post-public residual)

**Serial stamp:** free eng **s1492** — residual-honest M5 signing/matrix tip for the public
binary host **`iomesh-memory-mcp`**. Packaging and verify docs only; **does not** invent a
successful public tag release already shipped or forever-green signed CI.

### Release matrix

| Stage | What runs |
|-------|-----------|
| **Trigger** | Annotated `v*` tag push (maintainer deliberate — **no auto-tag**) or `workflow_dispatch` snapshot |
| **Workflow** | [`.github/workflows/release.yml`](.github/workflows/release.yml) |
| **Build** | GoReleaser v2 ([`.goreleaser.yaml`](.goreleaser.yaml)): CGO_ENABLED=0, multi-OS/arch |
| **Archives** | `tar.gz` / Windows `zip` + LICENSE + README + CHANGELOG |
| **Checksums** | `checksums.txt` (sha256) |
| **SBOM** | Syft SPDX per archive (`*.sbom.spdx.json`) |
| **Sign** | Keyless cosign on `checksums.txt` (tag path only; OIDC `id-token: write`) |
| **Publish** | GitHub Release assets (tag path); snapshot skips publish + sign |

Platforms: **linux / darwin / windows** × **amd64 / arm64** (windows/arm64 ignored).

### Local dry-run

```bash
# Needs goreleaser (+ syft for SBOM generation). Skips cosign (no laptop OIDC).
make release-snapshot   # → dist/ multi-arch; no GitHub publish
```

Snapshot CI path: Actions **workflow_dispatch** with `snapshot: true` →
`goreleaser release --snapshot --clean --skip=sign`.

### M5 honesty locks (non-claims)

- This tip **≠ invent a successful public tag release already shipped**
- residual PASS **≠ invent forever-green signed releases** · residual PASS **≠ invent M5 complete**
- dual_write **OFF** · **not Memory GA** · no invent GA
- Product binary name **`iomesh-memory-mcp`** (not `aion-memory-mcp`)
- **Kernel public prerequisite met** (`github.com/iome-sh/memory` public; no release PAT)
- **aion stays private** · **no auto-tag releases**

## Image name honesty

Product edge image (when published — **optional**, do not invent green):
**`ghcr.io/iome-sh/iomesh-memory-mcp`**  
Do **not** publish product edge as `aion-memory-mcp`.  
Local dogfood image remains `iomesh-memory-mcp:local` (compose PASS ≠ public registry).

## Versioning policy

- **0.x** — pre-stability lean host; additive tools preferred; breaking changes allowed with CHANGELOG honesty  
- **1.0+** — SemVer; breaking tool/CLI changes require major bump  
- Module path: `github.com/iome-sh/iomesh-memory-mcp`  

## Artifacts

| Path | Notes |
|------|--------|
| GitHub Release assets | GoReleaser on each deliberate `v*` tag (primary binary packaging) |
| `go install …@vX.Y.Z` | Works with a Go toolchain; public kernel — no GOPRIVATE |
| CI `build` job | linux/amd64 smoke on main / PR |
| GHCR image | Optional deliberate publish — not invent green on readiness residual |

```bash
make build                 # → bin/iomesh-memory-mcp (VERSION= from git describe)
make release-snapshot      # → dist/ (local multi-arch, no publish)
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@vX.Y.Z
```

## Honesty locks (non-claims)

- dual_write **OFF** · not product Memory GA · no aion import  
- Product name **iomesh-memory-mcp** (not `aion-memory-mcp`)  
- Kernel public prerequisite **met** · host public after deliberate flip  
- residual PASS ≠ invent signed release forever green · tip ≠ invent tag release shipped  
- M5 packaging residual present ≠ invent M5 complete · release packaging present ≠ invent GHCR green  
- **aion stays private** · **no auto-tag**  
