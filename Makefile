.PHONY: all build test test-race cover vet fmt fmt-check tidy vuln check ci clean

COVER ?= coverage.out
BIN ?= bin/iomesh-memory-mcp

all: check build

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

check: fmt-check vet test

# Mirrors GitHub Actions required gate (fmt + vet + test + vuln + build).
ci: fmt-check vet test vuln build

clean:
	rm -rf bin/ $(COVER) coverage.html
