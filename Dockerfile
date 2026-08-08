# Lean multi-stage build for iomesh-memory-mcp (edge Memory MCP host).
# Image name honesty: ghcr.io/iome-sh/iomesh-memory-mcp (not aion-memory-mcp).
# dual_write OFF · not Memory GA · no aion import.

FROM golang:1.26-bookworm AS build
WORKDIR /src

ENV GOPRIVATE=github.com/iome-sh/* \
    GONOSUMDB=github.com/iome-sh/* \
    CGO_ENABLED=0 \
    GOTOOLCHAIN=auto

# Optional: pass a netrc/token at build time for private kernel while still private.
# docker build --build-arg GH_TOKEN=... 
ARG GH_TOKEN=
RUN if [ -n "$GH_TOKEN" ]; then \
      git config --global url."https://x-access-token:${GH_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/iomesh-memory-mcp ./cmd/iomesh-memory-mcp

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/iomesh-memory-mcp /iomesh-memory-mcp
USER nonroot:nonroot
EXPOSE 8080
ENV PALACE_ROOT=/data/memory-palaces \
    MEMORY_MCP_HTTP_ADDR=:8080 \
    MEMORY_MCP_HTTP_PATH=/mcp
VOLUME ["/data/memory-palaces"]
ENTRYPOINT ["/iomesh-memory-mcp"]
