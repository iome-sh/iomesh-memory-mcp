# Contributing

Thanks for helping improve **iomesh-memory-mcp** (lean edge Memory MCP host).

## What this repo is

- **Edge MCP host** — stdio or streamable HTTP tools over `github.com/iome-sh/memory` Palace FS  
- **Naming honesty** — binary/image **`iomesh-memory-mcp`** (not product-name `aion-memory-mcp`)  
- **dual_write OFF** by default · **not product Memory GA**  
- **Does not import** `github.com/iome-sh/aion/**`  
- Private aion broker / CP / INSTALL_STORE / billing stay out of this tree  

This repository may remain **private** until a deliberate visibility flip. Do not assume public contribution workflows until that flip lands.

## Development setup

```bash
# Go version: see go.mod (CI uses that exact toolchain via GOTOOLCHAIN=auto)
export GOPRIVATE=github.com/iome-sh/*
export GONOSUMDB=github.com/iome-sh/*

git clone git@github.com:iome-sh/iomesh-memory-mcp.git
cd iomesh-memory-mcp
make test
make vet
make build
```

Local gate (mirrors CI spirit):

```bash
make ci   # fmt-check · vet · test · vuln · build
```

## Architecture (lean v1)

```text
cmd/iomesh-memory-mcp
  → internal/mcphost (stdio | HTTP)
  → github.com/iome-sh/memory PalaceStore
```

Tenant layout: `filepath.Join(palaceRoot, tenant)` as Palace `BaseDir`.

## PR expectations

1. Branch off `main`  
2. Keep tools lean; no aion imports  
3. Honesty locks intact in docs/responses (`dual_write=off`, `not_memory_ga`)  
4. Unit tests for tool handlers with temp palace dirs  
5. `make check` green  

## Out of scope here

- Enabling dual_write / aion audit by default  
- Inventing Memory GA  
- Requiring Qdrant/ONNX for default path  
- Flipping the repository public (separate deliberate act)  
