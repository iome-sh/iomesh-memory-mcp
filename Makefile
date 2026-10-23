.PHONY: all build test test-race cover vet fmt fmt-check tidy vuln check ci clean help release-snapshot

COVER ?= coverage.out
BIN ?= bin/iomesh-memory-mcp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.1.0)
LDFLAGS ?= -s -w -X github.com/iome-sh/iomesh-memory-mcp/internal/mcphost.ServerVersion=$(VERSION)

all: check build

help:
	@echo "Targets:"
	@echo "  build                       Build binary → \$$(BIN) (default bin/iomesh-memory-mcp)"
	@echo "  test / test-race            go test ./..."
	@echo "  cover                       Coverage profile"
	@echo "  vet / fmt / fmt-check / tidy / vuln"
	@echo "  check                       fmt-check · vet · test"
	@echo "  ci                          fmt-check · vet · test · vuln · build (GH Actions mirror)"
	@echo "  release-snapshot            Local GoReleaser snapshot (no publish; needs goreleaser + syft)"
	@echo "  clean                       Remove bin/ dist/ and coverage artifacts"
	@echo ""
	@echo "Honesty: dual_write OFF · not Memory GA · public · residual PASS ≠ live dogfood / public flip"

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=$(COVER) -covermode=atomic
	go tool cover -func=$(COVER) | tail -20

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi

tidy:
	go mod tidy
	go mod verify

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

build:
	go build ./...
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/iomesh-memory-mcp

# Local GoReleaser snapshot (no GitHub publish). Requires goreleaser (+ syft for SBOM).
# Skips cosign (no OIDC on laptop). Tag releases sign checksums via GitHub Actions.
#   go install github.com/goreleaser/goreleaser/v2@latest
release-snapshot:
	goreleaser release --snapshot --clean --skip=sign

check: fmt-check vet test

# Mirrors GitHub Actions required gate (fmt + vet + test + vuln + build).
ci: fmt-check vet test vuln build

clean:
	rm -rf bin/ dist/ $(COVER) coverage.html
