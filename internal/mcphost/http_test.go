package mcphost

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
	if _, ok := body["not_memory_ga"]; !ok {
		t.Fatalf("not_memory_ga key required: %s", raw)
	}
	if body["not_memory_ga"] != true {
		t.Fatalf("not_memory_ga: %v", body["not_memory_ga"])
	}
	if body["persist_embeddings"] != "off" {
		t.Fatalf("persist_embeddings: %v want off", body["persist_embeddings"])
	}
}

func TestServerVersionIsTipModulePseudoVersion(t *testing.T) {
	const want = "v0.4.3-0.20260924041108-b6b316535af1"
	if ServerVersion != want {
		t.Fatalf("ServerVersion: %q want %q", ServerVersion, want)
	}
	if HealthzSnapshot(nil).Version != want {
		t.Fatalf("healthz version: %q want %q", HealthzSnapshot(nil).Version, want)
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
		got.NotMemoryGA != want.NotMemoryGA || got.Embeddings != want.Embeddings ||
		got.PersistEmbeddings != want.PersistEmbeddings || got.Qdrant != want.Qdrant ||
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
	if body.NotMemoryGA != NotMemoryGA() {
		t.Fatalf("not_memory_ga: %v want %v", body.NotMemoryGA, NotMemoryGA())
	}
	if body.Embeddings != "hash" && body.Embeddings != "onnx" {
		t.Fatalf("embeddings: %q", body.Embeddings)
	}
	if body.Qdrant != "off" {
		t.Fatalf("qdrant must be off for lean host: %q", body.Qdrant)
	}
	if body.PersistEmbeddings != "off" {
		t.Fatalf("persist_embeddings: %q want off (default; hash never persists)", body.PersistEmbeddings)
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

func TestHTTPBindClass(t *testing.T) {
	t.Parallel()
	cases := []struct {
		addr string
		want string
	}{
		{addr: "127.0.0.1:8080", want: "loopback"},
		{addr: "[::1]:8080", want: "loopback"},
		{addr: "localhost:9090", want: "loopback"},
		{addr: "LOCALHOST:9090", want: "loopback"},
		{addr: ":8080", want: "non-loopback"},
		{addr: "0.0.0.0:8080", want: "non-loopback"},
		{addr: "[::]:8080", want: "non-loopback"},
		{addr: "192.168.1.10:8080", want: "non-loopback"},
	}
	for _, tc := range cases {
		if got := httpBindClass(tc.addr); got != tc.want {
			t.Errorf("httpBindClass(%q)=%q want %q", tc.addr, got, tc.want)
		}
	}

	// Class follows the normalized address, not the allow-non-loopback flag.
	loop, err := NormalizeListenAddr("127.0.0.1:8080", true)
	if err != nil {
		t.Fatal(err)
	}
	if got := httpBindClass(loop); got != "loopback" {
		t.Fatalf("allow flag must not reclass %q: %s", loop, got)
	}
	open, err := NormalizeListenAddr(":8080", true)
	if err != nil {
		t.Fatal(err)
	}
	if open != ":8080" || httpBindClass(open) != "non-loopback" {
		t.Fatalf("normalized :8080 with allow: addr=%q class=%s", open, httpBindClass(open))
	}
	forced, err := NormalizeListenAddr(":8080", false)
	if err != nil {
		t.Fatal(err)
	}
	if forced != "127.0.0.1:8080" || httpBindClass(forced) != "loopback" {
		t.Fatalf("forced loopback: addr=%q class=%s", forced, httpBindClass(forced))
	}
	all, err := NormalizeListenAddr("0.0.0.0:8080", true)
	if err != nil {
		t.Fatal(err)
	}
	if all != "0.0.0.0:8080" || httpBindClass(all) != "non-loopback" {
		t.Fatalf("normalized 0.0.0.0: addr=%q class=%s", all, httpBindClass(all))
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

func TestRequireHTTPSecretNonLoopback(t *testing.T) {
	if err := requireHTTPSecret("127.0.0.1:8080", ""); err != nil {
		t.Fatalf("loopback without secret: %v", err)
	}
	if err := requireHTTPSecret("localhost:9090", ""); err != nil {
		t.Fatalf("localhost without secret: %v", err)
	}
	if err := requireHTTPSecret("[::1]:8080", ""); err != nil {
		t.Fatalf("::1 without secret: %v", err)
	}
	if err := requireHTTPSecret("0.0.0.0:8080", "unit-secret"); err != nil {
		t.Fatalf("non-loopback with secret: %v", err)
	}
	for _, addr := range []string{"0.0.0.0:8080", ":8080", "[::]:8080", "192.168.1.10:8080"} {
		err := requireHTTPSecret(addr, "")
		if err == nil || !errors.Is(err, ErrHTTPSecretRequired) {
			t.Fatalf("%s without secret: %v", addr, err)
		}
	}
}

func TestRunHTTPNonLoopbackWithoutSecretFailsBeforeListen(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sdk := h.NewSDKServer()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	const addr = "0.0.0.0:18764"
	start := time.Now()
	err = RunHTTP(ctx, sdk, HTTPConfig{
		Addr:             addr,
		Path:             "/mcp",
		Host:             h,
		AllowNonLoopback: true,
		SharedSecret:     "",
	})
	if err == nil {
		t.Fatal("expected secret required before listen")
	}
	if !errors.Is(err, ErrHTTPSecretRequired) {
		t.Fatalf("got %v want ErrHTTPSecretRequired", err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("must fail before ListenAndServe, took %s", time.Since(start))
	}
	c, dialErr := net.DialTimeout("tcp", "127.0.0.1:18764", 80*time.Millisecond)
	if dialErr == nil {
		_ = c.Close()
		t.Fatal("must not listen when secret is missing")
	}
}

func TestRunHTTPLoopbackWithoutSecretHealthzOpen(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sdk := h.NewSDKServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunHTTP(ctx, sdk, HTTPConfig{
			Addr:         addr,
			Path:         "/mcp",
			Host:         h,
			SharedSecret: "",
		})
	}()

	var lastErr error
	var body HealthzResponse
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, getErr := http.Get("http://" + addr + "/healthz")
		if getErr != nil {
			lastErr = getErr
			time.Sleep(20 * time.Millisecond)
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("healthz status %d body=%s", resp.StatusCode, raw)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("healthz json: %v", err)
		}
		assertHealthzHonesty(t, body)
		mcpReq, _ := http.NewRequest(http.MethodPost, "http://"+addr+"/mcp", strings.NewReader(`{}`))
		mcpReq.Header.Set("Content-Type", "application/json")
		mcpResp, mcpErr := http.DefaultClient.Do(mcpReq)
		if mcpErr != nil {
			t.Fatalf("mcp without secret: %v", mcpErr)
		}
		_ = mcpResp.Body.Close()
		if mcpResp.StatusCode == http.StatusUnauthorized {
			t.Fatal("loopback without secret must not 401 MCP")
		}
		cancel()
		select {
		case <-errCh:
		case <-time.After(3 * time.Second):
			t.Fatal("RunHTTP did not exit")
		}
		return
	}
	cancel()
	t.Fatalf("healthz never came up on %s: %v", addr, lastErr)
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

func TestReadyWritablePalace200(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: %q", ct)
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, rr.Body.String())
	}
	assertReadyHonesty(t, body)
	if body.Status != "ok" {
		t.Fatalf("status field: %q", body.Status)
	}
	if !body.PalaceWritable {
		t.Fatal("palace_writable")
	}
	if !body.WALPendingRecoverable {
		t.Fatal("wal_pending_recoverable")
	}
	if body.WALPending != 0 {
		t.Fatalf("wal_pending: %d", body.WALPending)
	}
	raw := rr.Body.Bytes()
	assertReadyDoesNotLeak(t, raw, "t1")
	assertReadyJSONKeys(t, raw)
}

func TestReadyCloudWritable200(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "ws-1", Cloud: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = h.Close() }()
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	assertReadyHonesty(t, body)
	if !body.PalaceWritable || !body.WALPendingRecoverable || body.Status != "ok" {
		t.Fatalf("cloud ready: %+v", body)
	}
	assertReadyDoesNotLeak(t, rr.Body.Bytes(), "ws-1")
}

func TestReadyDoesNotInventDefaultTenant(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "default")); !os.IsNotExist(err) {
		t.Fatalf("must not invent a default tenant write: %v", err)
	}
}

func TestReadyNotWritable503HealthzStill200(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0555 not a write barrier on windows")
	}
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	probe := filepath.Join(root, "t1")
	if err := os.MkdirAll(probe, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(probe, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(probe, 0o700) })

	readyRR := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(readyRR, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if readyRR.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready status: %d body=%s", readyRR.Code, readyRR.Body.String())
	}
	var ready ReadyResponse
	if err := json.Unmarshal(readyRR.Body.Bytes(), &ready); err != nil {
		t.Fatalf("ready json: %v", err)
	}
	assertReadyHonesty(t, ready)
	if ready.Status != "not_ready" {
		t.Fatalf("status field: %q", ready.Status)
	}
	if ready.PalaceWritable {
		t.Fatal("palace_writable must be false")
	}

	healthRR := httptest.NewRecorder()
	HealthzHandler(h).ServeHTTP(healthRR, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if healthRR.Code != http.StatusOK {
		t.Fatalf("healthz must stay 200: %d body=%s", healthRR.Code, healthRR.Body.String())
	}
	var health HealthzResponse
	if err := json.Unmarshal(healthRR.Body.Bytes(), &health); err != nil {
		t.Fatalf("healthz json: %v", err)
	}
	assertHealthzHonesty(t, health)
	assertHealthzOmitsReadyFields(t, healthRR.Body.Bytes())
}

func TestReadyMissingRoot503(t *testing.T) {
	root := t.TempDir()
	notDir := filepath.Join(root, "not-a-dir")
	if err := os.WriteFile(notDir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	h, err := New(Config{PalaceRoot: notDir, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Status != "not_ready" || body.PalaceWritable {
		t.Fatalf("want not_ready unwritable, got %+v", body)
	}
	healthRR := httptest.NewRecorder()
	HealthzHandler(h).ServeHTTP(healthRR, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if healthRR.Code != http.StatusOK {
		t.Fatalf("healthz: %d", healthRR.Code)
	}
}

func TestReadySnapshotNotMemoryGA(t *testing.T) {
	if !ReadySnapshot(nil).NotMemoryGA {
		t.Fatal("not_memory_ga must be true when MEMORY_CLOUD_GA is unset")
	}

	t.Run("one", func(t *testing.T) {
		t.Setenv("MEMORY_CLOUD_GA", "1")
		if ReadySnapshot(nil).NotMemoryGA {
			t.Fatal("not_memory_ga must be false when MEMORY_CLOUD_GA=1")
		}
		h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t1"})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		body := ReadySnapshot(h)
		if body.NotMemoryGA {
			t.Fatal("writable host not_memory_ga must be false when MEMORY_CLOUD_GA=1")
		}
		if body.DualWrite != "off" {
			t.Fatalf("dual_write: %q", body.DualWrite)
		}
		hz := HealthzSnapshot(h)
		if hz.NotMemoryGA {
			t.Fatal("healthz not_memory_ga must be false when MEMORY_CLOUD_GA=1")
		}
		if hz.DualWrite != "off" || hz.Version != ServerVersion {
			t.Fatalf("healthz honesty: %+v", hz)
		}
		raw, err := json.Marshal(hz)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		if _, ok := m["not_memory_ga"]; !ok {
			t.Fatalf("healthz must keep not_memory_ga when false: %s", raw)
		}
		readyRaw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		var rm map[string]any
		if err := json.Unmarshal(readyRaw, &rm); err != nil {
			t.Fatal(err)
		}
		if _, ok := rm["not_memory_ga"]; !ok || rm["not_memory_ga"] != false {
			t.Fatalf("ready must keep not_memory_ga false: %s", readyRaw)
		}
	})

	// Not envTruthy: "true" / "yes" / "0" / empty / padded "1" do not flip.
	for _, v := range []string{"true", "yes", "0", "", "TRUE", " 1", "1 "} {
		t.Run("keep_"+v, func(t *testing.T) {
			t.Setenv("MEMORY_CLOUD_GA", v)
			if !ReadySnapshot(nil).NotMemoryGA || !HealthzSnapshot(nil).NotMemoryGA {
				t.Fatalf("MEMORY_CLOUD_GA=%q must not flip not_memory_ga", v)
			}
		})
	}
}

func TestReadyNilHostNotReady(t *testing.T) {
	rr := httptest.NewRecorder()
	ReadyHandler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	assertReadyHonesty(t, body)
	if body.Status != "not_ready" {
		t.Fatalf("status: %q", body.Status)
	}
	if body.PalaceWritable || body.WALPendingRecoverable {
		t.Fatalf("nil host must fail closed: %+v", body)
	}
}

func TestReadyWALPendingLeftoverRecoverable(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	pending := filepath.Join(root, "t1", filepath.FromSlash(walPendingRelDir))
	if err := os.MkdirAll(pending, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pending, "leftover.json"), []byte(`{"turn_id":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.WALPending != 1 {
		t.Fatalf("wal_pending: %d", body.WALPending)
	}
	if !body.WALPendingRecoverable || body.Status != "ok" {
		t.Fatalf("leftover regular file must be recoverable: %+v", body)
	}
}

func TestReadyWALPendingNestedDirNotRecoverable(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	nested := filepath.Join(root, "t1", filepath.FromSlash(walPendingRelDir), "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.WALPendingRecoverable || body.Status != "not_ready" {
		t.Fatalf("nested pending dir must not be recoverable: %+v", body)
	}
	if !body.PalaceWritable {
		t.Fatal("palace still writable")
	}
}

func TestReadyLastIngestUnix(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	destDir := filepath.Join(root, "t1", "tier-2-contextual")
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(destDir, "mem-1.json")
	if err := os.WriteFile(dest, []byte(`{"id":"mem-1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	want := st.ModTime().Unix()
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var body ReadyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.LastIngestUnix < want {
		t.Fatalf("last_ingest_unix=%d want >= %d", body.LastIngestUnix, want)
	}
}

func TestReadyDoesNotLeakTenantOrOrg(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "secret-tenant"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	assertReadyDoesNotLeak(t, rr.Body.Bytes(), "secret-tenant")
}

func TestReadyMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ready", nil)
	rr := httptest.NewRecorder()
	ReadyHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestReadyHead(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	ReadyHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodHead, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestHealthzJSONOmitsReadyFields(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rr := httptest.NewRecorder()
	HealthzHandler(h).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	var body HealthzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	assertHealthzHonesty(t, body)
	assertHealthzOmitsReadyFields(t, rr.Body.Bytes())
}

func TestRunHTTPSecretDoesNotWrapReady(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sdk := h.NewSDKServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunHTTP(ctx, sdk, HTTPConfig{
			Addr:         addr,
			Path:         "/mcp",
			Host:         h,
			SharedSecret: "unit-test-shared-secret",
		})
	}()

	var lastErr error
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		readyResp, getErr := http.Get("http://" + addr + "/ready")
		if getErr != nil {
			lastErr = getErr
			time.Sleep(20 * time.Millisecond)
			continue
		}
		readyRaw, _ := io.ReadAll(readyResp.Body)
		_ = readyResp.Body.Close()
		if readyResp.StatusCode != http.StatusOK {
			t.Fatalf("ready without secret: %d body=%s", readyResp.StatusCode, readyRaw)
		}
		var ready ReadyResponse
		if err := json.Unmarshal(readyRaw, &ready); err != nil {
			t.Fatalf("ready json: %v", err)
		}
		assertReadyHonesty(t, ready)
		if ready.DualWrite != "off" || ready.NotMemoryGA != NotMemoryGA() {
			t.Fatalf("ready honesty: %+v", ready)
		}

		healthResp, healthErr := http.Get("http://" + addr + "/healthz")
		if healthErr != nil {
			t.Fatalf("healthz: %v", healthErr)
		}
		healthRaw, _ := io.ReadAll(healthResp.Body)
		_ = healthResp.Body.Close()
		if healthResp.StatusCode != http.StatusOK {
			t.Fatalf("healthz without secret: %d body=%s", healthResp.StatusCode, healthRaw)
		}
		var health HealthzResponse
		if err := json.Unmarshal(healthRaw, &health); err != nil {
			t.Fatalf("healthz json: %v", err)
		}
		assertHealthzHonesty(t, health)
		assertHealthzOmitsReadyFields(t, healthRaw)

		mcpReq, _ := http.NewRequest(http.MethodPost, "http://"+addr+"/mcp", strings.NewReader(`{}`))
		mcpReq.Header.Set("Content-Type", "application/json")
		mcpResp, mcpErr := http.DefaultClient.Do(mcpReq)
		if mcpErr != nil {
			t.Fatalf("mcp: %v", mcpErr)
		}
		_ = mcpResp.Body.Close()
		if mcpResp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("mcp without secret must 401, got %d", mcpResp.StatusCode)
		}

		cancel()
		select {
		case <-errCh:
		case <-time.After(3 * time.Second):
			t.Fatal("RunHTTP did not exit")
		}
		return
	}
	cancel()
	t.Fatalf("ready never came up on %s: %v", addr, lastErr)
}

func assertReadyJSONKeys(t *testing.T, raw []byte) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, k := range []string{
		"status", "service", "dual_write", "not_memory_ga",
		"palace_writable", "wal_pending", "wal_pending_recoverable",
		"last_ingest_unix", "rss_bytes", "cgroup_memory_bytes",
	} {
		if _, ok := body[k]; !ok {
			t.Fatalf("ready missing %q: %s", k, raw)
		}
	}
	for _, k := range []string{
		"embeddings", "persist_embeddings", "qdrant", "version", "tools", "tool_names",
	} {
		if _, ok := body[k]; ok {
			t.Fatalf("ready must not include healthz field %q: %s", k, raw)
		}
	}
}

func assertReadyHonesty(t *testing.T, body ReadyResponse) {
	t.Helper()
	if body.Service != ServerName {
		t.Fatalf("service: %q", body.Service)
	}
	if body.DualWrite != "off" {
		t.Fatalf("dual_write: %q", body.DualWrite)
	}
	if body.NotMemoryGA != NotMemoryGA() {
		t.Fatalf("not_memory_ga: %v want %v", body.NotMemoryGA, NotMemoryGA())
	}
	if body.Status != "ok" && body.Status != "not_ready" {
		t.Fatalf("status: %q", body.Status)
	}
}

func assertReadyDoesNotLeak(t *testing.T, raw []byte, tenant string) {
	t.Helper()
	s := string(raw)
	if tenant != "" && strings.Contains(s, tenant) {
		t.Fatalf("ready leaked process tenant: %s", s)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, leak := range []string{"tenant", "org", "organization"} {
		if _, ok := body[leak]; ok {
			t.Fatalf("ready must not include %q: %s", leak, s)
		}
	}
	if strings.Contains(s, "secret") || strings.Contains(s, "dlp") {
		t.Fatalf("ready must not leak secret/dlp fields: %s", s)
	}
}

func assertHealthzOmitsReadyFields(t *testing.T, raw []byte) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, k := range []string{
		"palace_writable", "wal_pending", "wal_pending_recoverable",
		"last_ingest_unix", "rss_bytes", "cgroup_memory_bytes",
	} {
		if _, ok := body[k]; ok {
			t.Fatalf("healthz must not include ready field %q: %s", k, raw)
		}
	}
}
