# Support

## How to get help

| Need | Where |
|------|--------|
| Usage questions / bugs | [GitHub Issues](https://github.com/iome-sh/iomesh-memory-mcp/issues) (use the templates when open to contributors) |
| Security vulnerability | Private [Security Advisory](https://github.com/iome-sh/iomesh-memory-mcp/security/advisories/new) or **security@iome.sh** — see [SECURITY.md](SECURITY.md) |
| Kernel API | [github.com/iome-sh/memory](https://github.com/iome-sh/memory) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## What we maintain

- **Best effort** on `main` and the latest `v0.1.x` line  
- Security fixes on the default branch when feasible  
- **No cloud Memory SLA** — local-primary edge host, not hosted Palace  

## What we do not provide here

- Hosted Palace / multitenant cloud Memory onboarding  
- Product Memory GA install guarantees  
- Private monorepo (`aion`) broker / control-plane support via this binary  
- Default dual_write / audit mesh side effects  

## Before filing an issue

1. Run `make check` or note CI failures  
2. Redact API keys, palace contents, and private paths from logs  
3. Include binary version (`iomesh-memory-mcp` / `v0.1.0` or `git describe`) or commit SHA and OS  
4. Confirm the report is about the **edge MCP host** — not a request to invent hosted Memory GA  
