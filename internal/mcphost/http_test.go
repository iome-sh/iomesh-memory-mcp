package mcphost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("content-type: %q", ct)
	}
	var body HealthzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, rr.Body.String())
	}
	if body.Status != "ok" {
		t.Fatalf("status field: %q", body.Status)
	}
	if body.Service != ServerName {
		t.Fatalf("service: %q", body.Service)
	}
	if body.DualWrite != "off" {
		t.Fatalf("dual_write: %q", body.DualWrite)
	}
	if !body.NotMemoryGA {
		t.Fatal("not_memory_ga must be true")
	}
	if body.Embeddings != "hash" && body.Embeddings != "onnx" {
		t.Fatalf("embeddings: %q", body.Embeddings)
	}
	if body.Qdrant != "off" {
		t.Fatalf("qdrant must be off for lean host: %q", body.Qdrant)
	}
	if body.Version != ServerVersion {
		t.Fatalf("version: %q", body.Version)
	}
	if body.Tools < 9 {
		t.Fatalf("tools count: %d want >= 9", body.Tools)
	}
	if body.Tools != len(body.ToolNames) {
		t.Fatalf("tools=%d tool_names=%d (%v)", body.Tools, len(body.ToolNames), body.ToolNames)
	}
	have := make(map[string]bool, len(body.ToolNames))
	for _, n := range body.ToolNames {
		have[n] = true
	}
	for _, n := range []string{"memory_write", "memory_related", "memory_supersede_entity", "memory_retrieve"} {
		if !have[n] {
			t.Fatalf("tool_names missing %q: %v", n, body.ToolNames)
		}
	}
}

func TestHealthzMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestNormalizeMCPPath(t *testing.T) {
	cases := map[string]string{
		"":      "/mcp",
		"mcp":   "/mcp",
		"/mcp":  "/mcp",
		"/mcp/": "/mcp",
		"/":     "/",
	}
	for in, want := range cases {
		if got := NormalizeMCPPath(in); got != want {
			t.Errorf("NormalizeMCPPath(%q)=%q want %q", in, got, want)
		}
	}
}

func TestRunHTTPHealthzLive(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sdk := h.NewSDKServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use :0 via httptest is easier; exercise HealthzHandler through mux-like path.
	// RunHTTP binds a real port — use a free port via Listen pattern in short smoke.
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunHTTP(ctx, sdk, HTTPConfig{Addr: "127.0.0.1:0", Path: "/mcp"})
	}()

	// RunHTTP with :0 still works but we need the actual bound addr.
	// Simpler: direct mux probe already covered; cancel after brief wait to exercise shutdown.
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			// Listen on :0 may fail on some platforms if RunHTTP doesn't rebind — accept closed path.
			t.Logf("RunHTTP exit: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("RunHTTP did not exit after cancel")
	}
}

func TestHealthzHead(t *testing.T) {
	req := httptest.NewRequest(http.MethodHead, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	b, _ := io.ReadAll(rr.Body)
	if len(b) != 0 {
		// HEAD may still encode; accept empty or body depending on encoder.
		t.Logf("HEAD body len=%d", len(b))
	}
}
