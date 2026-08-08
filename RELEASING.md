# Releasing

Ship from `main` via PR; cut annotated semver tags for binary/image consumers.

## When to bump and tag

| Trigger | Bump | Examples |
|---------|------|----------|
| New MCP tools / CLI flags | **minor** within `v0.x` while pre-1.0 | `memory_list`, HTTP path |
| Breaking tool schemas | **minor** (pre-1.0 honesty) or document migration | Rename tool input fields |
| Docs-only / OSS process bar | usually **no** tag | SECURITY, OPEN_SOURCE_AUDIT |
| Security fix | **patch** | CVE follow-up |

Checklist items that must move with the tag:

- [CHANGELOG.md](CHANGELOG.md) `[Unreleased]` → `## [X.Y.Z]`  
- [README.md](README.md) version / install line  
- `ServerVersion` constant in `internal/mcphost` if stamped in binary  

## Checklist before a tag

1. [ ] `make ci` green locally  
2. [ ] GitHub Actions **ci-success** green on the release commit  
3. [ ] [CHANGELOG.md](CHANGELOG.md) updated  
4. [ ] No secrets or palace data in tree  
5. [ ] Honesty locks intact: dual_write OFF · not Memory GA · no aion import  
6. [ ] Annotated tag `vX.Y.Z` pushed  

## Tag (maintainers)

```bash
git checkout main
git pull origin main
# edit CHANGELOG.md + version stamp if needed
git commit -am "chore: release vX.Y.Z"
git push origin main

git tag -a vX.Y.Z -m "vX.Y.Z — short release title"
git push origin vX.Y.Z
```

## Image name honesty

Product edge image (when published): **`ghcr.io/iome-sh/iomesh-memory-mcp`**  
Do **not** publish product edge as `aion-memory-mcp`.

## Versioning policy

- **0.x** — pre-stability lean host; additive tools preferred  
- Module path: `github.com/iome-sh/iomesh-memory-mcp`  
