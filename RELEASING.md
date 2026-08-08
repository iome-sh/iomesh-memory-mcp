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
6. [ ] **Kernel public prerequisite** for public consumers: `github.com/iome-sh/memory` must be public (or consumers need `GOPRIVATE` + token while kernel private)  
7. [ ] Default `ServerVersion` string matches the tag family (GoReleaser ldflags set `v{{.Version}}`)  
8. [ ] Annotated tag `vX.Y.Z` pushed (GoReleaser **release** workflow green; assets on GitHub Release)

## Tag and publish (maintainers)

```bash
git checkout main
git pull origin main
# edit CHANGELOG.md + ServerVersion default + README/SECURITY if needed
git commit -am "chore: release vX.Y.Z"
git push origin main

# Annotated tag — triggers .github/workflows/release.yml (GoReleaser)
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

**Private kernel residual:** while `github.com/iome-sh/memory` is private, the release
workflow needs the same module token residual as CI (`IOMESH_CI_PAT` / org access).
After the kernel is public, module fetch is public and the PAT is optional.

Verify:

```bash
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp 'https://github.com/iome-sh/iomesh-memory-mcp/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

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
| GitHub Release assets | GoReleaser on each `v*` tag (primary binary packaging) |
| `go install …@vX.Y.Z` | Works with a Go toolchain; needs public kernel (or GOPRIVATE) |
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
- Repo visibility flip is a separate deliberate act (**kernel first**)  
- residual PASS ≠ public flip · release packaging present ≠ invent GHCR green  
