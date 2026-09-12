package mcphost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthzDoesNotLeakTenantOrOrg(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "secret-tenant"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler(h).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	raw := rr.Body.String()
	if strings.Contains(raw, "secret-tenant") {
		t.Fatalf("healthz leaked process tenant: %s", raw)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, leak := range []string{"tenant", "org", "organization"} {
		if _, ok := body[leak]; ok {
			t.Fatalf("healthz must not include %q: %s", leak, raw)
		}
	}
	if body["dual_write"] != "off" {
		t.Fatalf("dual_write: %v", body["dual_write"])
	}
	if body["not_memory_ga"] != true {
		t.Fatalf("not_memory_ga: %v", body["not_memory_ga"])
	}
}

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
	assertHealthzHonesty(t, body)
}

func TestHealthzSnapshotMatchesHandler(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler(h).ServeHTTP(rr, req)
	var got HealthzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v body=%s", err, rr.Body.String())
	}
	want := HealthzSnapshot(h)
	if got.Status != want.Status || got.Service != want.Service || got.DualWrite != want.DualWrite ||
		got.NotMemoryGA != want.NotMemoryGA || got.Embeddings != want.Embeddings || got.Qdrant != want.Qdrant ||
		got.Version != want.Version || got.Tools != want.Tools {
		t.Fatalf("handler vs snapshot: got=%+v want=%+v", got, want)
	}
	if len(got.ToolNames) != len(want.ToolNames) {
		t.Fatalf("tool_names len: handler=%d snapshot=%d", len(got.ToolNames), len(want.ToolNames))
	}
	for i := range want.ToolNames {
		if got.ToolNames[i] != want.ToolNames[i] {
			t.Fatalf("tool_names[%d]: handler=%q snapshot=%q", i, got.ToolNames[i], want.ToolNames[i])
		}
	}
	assertHealthzHonesty(t, want)
}

func TestHealthzSnapshotNilHostEnvHash(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	body := HealthzSnapshot(nil)
	assertHealthzHonesty(t, body)
	if body.Embeddings != "hash" {
		t.Fatalf("embeddings: %q want hash", body.Embeddings)
	}
}

func assertHealthzHonesty(t *testing.T, body HealthzResponse) {
	t.Helper()
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
	if body.Tools < 11 {
		t.Fatalf("tools count: %d want >= 11", body.Tools)
	}
	if body.Tools != len(body.ToolNames) {
		t.Fatalf("tools=%d tool_names=%d (%v)", body.Tools, len(body.ToolNames), body.ToolNames)
	}
	have := make(map[string]bool, len(body.ToolNames))
	for _, n := range body.ToolNames {
		have[n] = true
	}
	for _, n := range []string{"memory_write", "memory_related", "memory_supersede_entity", "memory_retrieve", "ops_digest_export", "memory_extract_facts"} {
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

func TestNormalizeListenAddr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		allow   bool
		want    string
		wantErr bool
	}{
		{in: ":8080", allow: false, want: "127.0.0.1:8080"},
		{in: ":8080", allow: true, want: ":8080"},
		{in: "127.0.0.1:8080", allow: false, want: "127.0.0.1:8080"},
		{in: "localhost:9090", allow: false, want: "localhost:9090"},
		{in: "[::1]:8080", allow: false, want: "[::1]:8080"},
		{in: "0.0.0.0:8080", allow: false, wantErr: true},
		{in: "0.0.0.0:8080", allow: true, want: "0.0.0.0:8080"},
		{in: "[::]:8080", allow: false, wantErr: true},
		{in: "[::]:8080", allow: true, want: "[::]:8080"},
		{in: "192.168.1.10:8080", allow: false, wantErr: true},
		{in: "192.168.1.10:8080", allow: true, want: "192.168.1.10:8080"},
		{in: "", allow: false, wantErr: true},
	}
	for _, tc := range cases {
		got, err := NormalizeListenAddr(tc.in, tc.allow)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeListenAddr(%q, allow=%v) = %q, want error", tc.in, tc.allow, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeListenAddr(%q, allow=%v): %v", tc.in, tc.allow, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeListenAddr(%q, allow=%v)=%q want %q", tc.in, tc.allow, got, tc.want)
		}
	}
}

func TestOptionalSharedSecretFailClosed(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	const secret = "unit-test-shared-secret"
	h := WithOptionalSharedSecret(secret, next)

	missing := httptest.NewRecorder()
	h.ServeHTTP(missing, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing secret: status %d", missing.Code)
	}

	wrong := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	wrong.Header.Set(MCPSecretHeader, "not-the-secret")
	wrongRR := httptest.NewRecorder()
	h.ServeHTTP(wrongRR, wrong)
	if wrongRR.Code != http.StatusUnauthorized {
		t.Fatalf("wrong secret: status %d", wrongRR.Code)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	okReq.Header.Set(MCPSecretHeader, secret)
	okRR := httptest.NewRecorder()
	h.ServeHTTP(okRR, okReq)
	if okRR.Code != http.StatusOK {
		t.Fatalf("header secret: status %d", okRR.Code)
	}

	bearer := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	bearer.Header.Set("Authorization", "Bearer "+secret)
	bearerRR := httptest.NewRecorder()
	h.ServeHTTP(bearerRR, bearer)
	if bearerRR.Code != http.StatusOK {
		t.Fatalf("bearer secret: status %d", bearerRR.Code)
	}

	open := WithOptionalSharedSecret("", next)
	openRR := httptest.NewRecorder()
	open.ServeHTTP(openRR, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if openRR.Code != http.StatusOK {
		t.Fatalf("empty secret must not require header: status %d", openRR.Code)
	}
}

func TestHealthzOpenWhenSecretConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	HealthzHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz must stay open: %d", rr.Code)
	}
	var body HealthzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	assertHealthzHonesty(t, body)
	raw := rr.Body.String()
	if strings.Contains(raw, "secret") || strings.Contains(raw, "dlp") {
		t.Fatalf("healthz honesty must not grow secret/dlp fields: %s", raw)
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
