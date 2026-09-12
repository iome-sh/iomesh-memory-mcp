# Support

## How to get help

| Need | Where |
|------|--------|
| Usage questions / bugs | [GitHub Issues](https://github.com/iome-sh/iomesh-memory-mcp/issues) (use the templates when open to contributors) |
| Security vulnerability | Private [Security Advisory](https://github.com/iome-sh/iomesh-memory-mcp/security/advisories/new) or **security@iome.sh** — see [SECURITY.md](SECURITY.md) |
| Kernel API / library | [github.com/iome-sh/memory](https://github.com/iome-sh/memory) (related **memory kernel**) |
| Release / version policy | [RELEASING.md](RELEASING.md) — **Support / version policy** |
| Edge dogfood / install matrix | [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — E3 matrix · E4 operator runbook |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## What we maintain

- **Supported line:** latest **GitHub Release** tag on the main line (deliberate annotated `v*` tags from `main`)
- **Best effort** on `main` tip and the latest published `v0.x` (or later major) family
- Security fixes on the default branch when feasible
- **No cloud Memory SLA** — local-primary edge host, not hosted Palace

Packaging: GoReleaser + SBOM + keyless cosign on tag releases (see [RELEASING.md](RELEASING.md)).  
**Pin versions for production**; **snapshot ≠ production release**.

## What we do not provide here

- Hosted Palace / multitenant cloud Memory onboarding  
- Cloud Memory install SLAs (this host is local-primary edge)  
- Mesh control-plane / broker support via this binary  
- Default mesh audit publish from this host  

## Before filing an issue

1. Run `make check` or note CI failures  
2. Redact API keys, palace contents, and private paths from logs  
3. Include binary version (`iomesh-memory-mcp` / `v0.1.0` or `git describe`) or commit SHA and OS  
4. Confirm the report is about the **edge MCP host**  
5. For install/attach questions, cite which matrix row you used (stdio · HTTP · Compose · TUI) from [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md)  
