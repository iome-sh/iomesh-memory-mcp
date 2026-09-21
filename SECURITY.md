# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v0.4.x` (incl. `v0.4.2`) | ✅ security fixes |
| `v0.3.x` | best-effort |
| `v0.2.x` | best-effort |
| `v0.1.x` | best-effort |
| `main` | development tip |
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
It exposes tools over **stdio** or **streamable HTTP**.

| Trust boundary | Posture |
|----------------|---------|
| Local Palace filesystem (`PALACE_ROOT` / `-palace-root`) | **User data** — the process can read/write all entries under that root. Treat as confidential. |
| Tenant subdirectory (`MEMORY_TENANT` / tool `tenant`) | **Path-based isolation only** (local-dev) — same process residual. Cloud mode (`-cloud` / `MEMORY_CLOUD`): one store, required process tenant, second tool tenant is 400. Omitted tool tenant fail-closes (does not write `$PALACE_ROOT/default`). `.`, `..`, and separators fail closed. |
| Shared palace root | **Single-writer contract** — one host process per `PALACE_ROOT`. Cloud: `palace.lock` (pid + start time) fails start if that pid is alive (crash-safety, not flock). Multi-process writers are unsupported (last-write-wins). Path tenancy is not cloud isolation. |
| Streamable HTTP (`MEMORY_MCP_HTTP_ADDR`) | Network-exposed MCP. Default bind is loopback (`:8080` → `127.0.0.1:8080`). `0.0.0.0` / non-loopback requires `-allow-non-loopback` **and** `MEMORY_MCP_HTTP_SECRET` (fatal before listen if the secret is missing — not a request 401 after bind). When set, the secret (`X-Memory-MCP-Secret` or `Authorization: Bearer`) fail-closes MCP. `GET /healthz` stays open. Unauthenticated HTTP remains residual on loopback when the secret is unset. |
| Mesh audit publish | **Off by default**. Lean v1 does not enable mesh audit publish. Optional `dual_write` is residual / later. |
| Kernel embeddings | Default hash/simple embedding path; optional ONNX residual via kernel — not required for dogfood. |

### Residual risks

- **HTTP unauthenticated on loopback** when the secret is unset — do not expose to untrusted networks. Loopback is the default bind; `0.0.0.0` is refused without `-allow-non-loopback` **and** `MEMORY_MCP_HTTP_SECRET` (fatal before listen). `/healthz` stays open.  
- **Host-side DLP is residual heuristics** — ingest/write redacts common paste shapes (`ghp_` / `sk-` / AWS `AKIA` / Slack `xox*` / PEM / JWT-shaped). Not commercial DLP, not hardware-bound keys, not default envelope encryption.  
- **Path-based tenancy ≠ multi-tenant cloud isolation** — one process, shared code. Omitted and invalid tenant fail closed so list/write stay under `$PALACE_ROOT/<tenant>/` and never share `$PALACE_ROOT/default`. `GET /healthz` does not leak tenant or org. Organization isolation for the I/O Mesh broker is a separate HTTP header (`X-IOMesh-Org`) on mesh clients; this host does not implement that.  
- **Single-writer contract** — one host process per palace root. Cloud mode adds `palace.lock` (pid + start time; start fails if that pid is alive). That is crash-safety, not flock. Multi-process writers on a shared root remain unsupported. In-process kernel `writeMu` is not flock and not a tenancy boundary.  
- **Palace FS is user data** — encryption at rest, backup, and OS permissions are operator responsibilities.  
- **No hosted Palace SLA.**  
- **No default mesh audit side effects.**

### What this is *not*

- Not multitenant hosted Memory  
- Not a freemium cloud Palace  
- Not a mesh control plane  
- Not automatic mesh audit publish / billing / plan gates  

## Hardening checklist for operators

1. Point `PALACE_ROOT` at a directory with appropriate OS permissions  
2. Prefer **stdio** for local TUI attach; if HTTP, keep the loopback default. Non-loopback **requires** `MEMORY_MCP_HTTP_SECRET` (process will not listen without it). `/healthz` stays unauthenticated.  
3. Do not commit palace contents, `.env`, or API keys  
4. Keep mesh audit publish off unless you deliberately wire a future audit adapter  
5. Run `make vuln` / CI govulncheck on PRs  

## Dependency security

```bash
make vuln   # govulncheck ./...
make test
```

## Disclosure preference

Coordinated disclosure: we prefer to ship a fix (or mitigating docs) before public write-ups when the issue is exploitable in default configurations.
