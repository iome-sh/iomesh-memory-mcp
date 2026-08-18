# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v0.1.x` / `main` (lean scaffold) | ✅ security fixes |
| pre-release pseudo-versions | development tip |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Preferred channels (in order):

1. **GitHub Security Advisory** (private) — Security → Advisories → Report a vulnerability on this repository  
2. Email **security@iome.sh**

Include:

- Description of the issue and impact  
- Reproduction steps or proof-of-concept  
- Affected commit / tag if known  

We aim to acknowledge reports within **72 hours** and provide a remediation timeline after triage.

## Threat model (edge Memory MCP host)

`github.com/iome-sh/iomesh-memory-mcp` is a **local-primary MCP host** over the Palace kernel.
It exposes tools over **stdio** or **streamable HTTP**. It is **not product Memory GA**.

| Trust boundary | Posture |
|----------------|---------|
| Local Palace filesystem (`PALACE_ROOT` / `-palace-root`) | **User data** — the process can read/write all entries under that root. Treat as confidential. |
| Tenant subdirectory (`MEMORY_TENANT` / tool `tenant`) | **Path-based isolation only** — same process residual; not cloud multi-tenant security. `.`, `..`, and separators fail closed. |
| Streamable HTTP (`MEMORY_MCP_HTTP_ADDR`) | Network-exposed MCP. Bind to localhost or put behind auth/proxy for non-lab use. No built-in auth in lean v1. |
| dual_write / audit | **OFF by default**. Lean v1 does not enable aion audit publish. Optional dual_write is residual / later. |
| Kernel embeddings | Default hash/simple embedding path; optional ONNX residual via kernel — not required for dogfood. |

### Residual risks (honest)

- **HTTP mode has no auth in lean v1** — do not expose to untrusted networks.  
- **Path-based tenancy ≠ multi-tenant cloud isolation** — one process, shared code. Invalid tenant segments fail closed so list/write stay under `$PALACE_ROOT/<tenant>/`.  
- **Palace FS is user data** — encryption at rest, backup, and OS permissions are operator responsibilities.  
- **Not Memory GA** — no hosted Palace SLA; aion broker stays private.  
- **dual_write OFF** — no default mesh audit side effects.

### What this is *not*

- Not multitenant hosted Memory  
- Not a freemium cloud Palace  
- Not the private aion control plane  
- Not automatic dual_write / billing / plan gates  

## Hardening checklist for operators

1. Point `PALACE_ROOT` at a directory with appropriate OS permissions  
2. Prefer **stdio** for local TUI attach; if HTTP, bind `127.0.0.1` and use a reverse proxy with auth  
3. Do not commit palace contents, `.env`, or API keys  
4. Keep dual_write OFF unless you deliberately wire a future audit adapter  
5. Run `make vuln` / CI govulncheck on PRs  

## Dependency security

```bash
make vuln   # govulncheck ./...
make test
```

## Disclosure preference

Coordinated disclosure: we prefer to ship a fix (or mitigating docs) before public write-ups when the issue is exploitable in default configurations.
