.PHONY: all build test test-race cover vet fmt fmt-check tidy vuln check ci clean help edge-dogfood-gate

COVER ?= coverage.out
BIN ?= bin/iomesh-memory-mcp

all: check build

help:
	@echo "Targets:"
	@echo "  build              Build binary → \$$(BIN) (default bin/iomesh-memory-mcp)"
	@echo "  test / test-race   go test ./..."
	@echo "  cover              Coverage profile"
	@echo "  vet / fmt / fmt-check / tidy / vuln"
	@echo "  check              fmt-check · vet · test"
	@echo "  ci                 fmt-check · vet · test · vuln · build (GH Actions mirror)"
	@echo "  edge-dogfood-gate  Offline M3 residual greps (no docker daemon)"
	@echo "  clean              Remove bin/ and coverage artifacts"
	@echo ""
	@echo "Honesty: dual_write OFF · not Memory GA · still private · residual PASS ≠ live dogfood / public flip"

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
	GOPRIVATE=github.com/iome-sh/* GONOSUMDB=github.com/iome-sh/* go mod tidy
	go mod verify

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

build:
	go build ./...
	go build -o $(BIN) ./cmd/iomesh-memory-mcp

# Offline M3 edge dogfood residual (file greps only — no docker / server / gcloud).
edge-dogfood-gate:
	@bash scripts/edge_dogfood_gate.sh

check: fmt-check vet test

# Mirrors GitHub Actions required gate (fmt + vet + test + vuln + build).
# Does not require edge-dogfood-gate (optional offline residual; run explicitly).
ci: fmt-check vet test vuln build

clean:
	rm -rf bin/ $(COVER) coverage.html
