package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/iome-sh/iomesh-memory-mcp/internal/mcphost"
)

func TestPreflightPrintsHealthzAndExits(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	t.Setenv("MEMORY_MCP_HTTP_ADDR", "")
	t.Setenv("AION_MEMORY_MCP_HTTP_ADDR", "")
	t.Setenv("PALACE_ROOT", "")
	t.Setenv("MEMORY_TENANT", "")

	palace := t.TempDir()
	var stdout bytes.Buffer
	done := make(chan error, 1)
	go func() {
		// -http-addr must not bind: preflight exits without listen / stdio MCP.
		done <- run([]string{
			"-preflight",
			"-palace-root", palace,
			"-tenant", "default",
			"-http-addr", "127.0.0.1:0",
		}, &stdout)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("preflight: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("preflight must exit without listening or running stdio MCP")
	}

	raw := stdout.Bytes()
	if len(bytes.TrimSpace(raw)) == 0 {
		t.Fatal("expected honesty JSON on stdout")
	}
	var body mcphost.HealthzResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v out=%s", err, stdout.String())
	}
	if body.Status != "ok" {
		t.Fatalf("status: %q", body.Status)
	}
	if body.Service != mcphost.ServerName {
		t.Fatalf("service: %q", body.Service)
	}
	if body.DualWrite != "off" {
		t.Fatalf("dual_write: %q (must be off)", body.DualWrite)
	}
	if !body.NotMemoryGA {
		t.Fatal("not_memory_ga must be true")
	}
	if body.Embeddings != "hash" && body.Embeddings != "onnx" {
		t.Fatalf("embeddings: %q", body.Embeddings)
	}
	if body.Qdrant != "off" {
		t.Fatalf("qdrant: %q (must be off)", body.Qdrant)
	}
	if body.Version != mcphost.ServerVersion {
		t.Fatalf("version: %q", body.Version)
	}
	if body.Tools < 9 {
		t.Fatalf("tools: %d want >= 9 (compile-time registration, not tools/list)", body.Tools)
	}
	if body.Tools != len(body.ToolNames) {
		t.Fatalf("tools=%d tool_names=%d", body.Tools, len(body.ToolNames))
	}
	// Same snapshot as GET /healthz after constructing the host.
	host, err := mcphost.New(mcphost.Config{PalaceRoot: palace, DefaultTenant: "default"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	want := mcphost.HealthzSnapshot(host)
	if body.Service != want.Service || body.DualWrite != want.DualWrite || body.Qdrant != want.Qdrant ||
		body.Tools != want.Tools || body.Version != want.Version || body.Embeddings != want.Embeddings {
		t.Fatalf("preflight vs HealthzSnapshot: got=%+v want=%+v", body, want)
	}
	if strings.Contains(stdout.String(), "Memory GA") && !body.NotMemoryGA {
		t.Fatal("must not claim Memory GA")
	}
}

func TestUnknownFlagDoesNotStart(t *testing.T) {
	err := run([]string{"-not-a-real-flag"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected flag error")
	}
	if !errors.Is(err, errFlag) {
		t.Fatalf("want errFlag, got %v", err)
	}
}

func TestPreflightRequiresPalaceRoot(t *testing.T) {
	t.Setenv("PALACE_ROOT", "")
	err := run([]string{"-preflight", "-palace-root", ""}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for empty palace root")
	}
	if !strings.Contains(err.Error(), "palace root") {
		t.Fatalf("err: %v", err)
	}
}
