# Support

## How to get help

| Need | Where |
|------|--------|
| Usage questions / bugs | [GitHub Issues](https://github.com/iome-sh/iomesh-memory-mcp/issues) (use the templates when open to contributors) |
| Security vulnerability | Private [Security Advisory](https://github.com/iome-sh/iomesh-memory-mcp/security/advisories/new) or **security@iome.sh** — see [SECURITY.md](SECURITY.md) |
| Kernel API / library | [github.com/iome-sh/memory](https://github.com/iome-sh/memory) (related **memory kernel**) |
| Release / version policy | [RELEASING.md](RELEASING.md) — **Support / version policy (E5)** |
| Edge dogfood / install matrix | [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md) — E3 matrix · E4 operator runbook |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## What we maintain

- **Supported line:** latest **GitHub Release** tag on the main line (deliberate annotated `v*` tags from `main`)
- **Best effort** on `main` tip and the latest published `v0.x` (or later major) family
- Security fixes on the default branch when feasible
- **No cloud Memory SLA** — local-primary edge host, not hosted Palace

Packaging honesty: GoReleaser + SBOM + keyless cosign on tag releases (see [RELEASING.md](RELEASING.md)).  
**Pin versions for production**; **snapshot ≠ production release**.  
residual PASS ≠ invent forever-green signed releases · residual PASS ≠ invent Edge Memory GA.

## What we do not provide here

- Hosted Palace / multitenant cloud Memory onboarding  
- Product **Memory GA** install guarantees (this host is edge-only; **not Memory GA**)  
- **Edge Memory GA** declaration (candidacy docs only — residual PASS ≠ invent Edge Memory GA declared)  
- Private monorepo (`aion`) broker / control-plane support via this binary  
- Default dual_write / audit mesh side effects  

## Before filing an issue

1. Run `make check` or note CI failures  
2. Redact API keys, palace contents, and private paths from logs  
3. Include binary version (`iomesh-memory-mcp` / `v0.1.0` or `git describe`) or commit SHA and OS  
4. Confirm the report is about the **edge MCP host** — not a request to invent hosted Memory GA or Edge Memory GA  
5. For install/attach questions, cite which matrix row you used (stdio · HTTP · Compose · TUI) from [docs/EDGE_DOGFOOD.md](docs/EDGE_DOGFOOD.md)  
