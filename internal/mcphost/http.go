package mcphost

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSecretHeader is the optional shared-secret header for streamable HTTP MCP.
// Also accepted: Authorization: Bearer <secret>. GET /healthz stays open.
const MCPSecretHeader = "X-Memory-MCP-Secret"

// HealthzResponse is the JSON body for GET /healthz.
type HealthzResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	DualWrite   string `json:"dual_write"`
	NotMemoryGA bool   `json:"not_memory_ga"`
	// Embeddings: "hash" (default) or "onnx" when MEMORY_ONNX_MODEL_PATH is set at process start.
	Embeddings string `json:"embeddings"`
	// Qdrant: lean host does not wire VectorStore into search — always "off" here (kernel residual only).
	Qdrant  string `json:"qdrant"`
	Version string `json:"version,omitempty"`
	// Tools is the compile-time lean registered count (not a live MCP tools/list stamp).
	Tools int `json:"tools"`
	// ToolNames is the compile-time lean registered names. Residual-honest; optional
	// for probes that only need the count. s1509 TUI attach tools=6 is historical.
	ToolNames []string `json:"tool_names,omitempty"`
}

// HealthzSnapshot is the residual-honest GET /healthz body used by HTTP and CLI -preflight.
// Live EmbeddingMode when host != nil; nil → env snapshot (same as HealthzHandler).
// tools / tool_names are compile-time registration — not a live tools/list stamp, not ingest.
// dual_write OFF · not Memory GA · qdrant off · no hosted palace probe.
// Honesty fields are unchanged by DLP / optional HTTP secret (no dlp/auth fields).
func HealthzSnapshot(host *Host) HealthzResponse {
	emb := "hash"
	if host != nil {
		emb = host.EmbeddingMode()
	} else if strings.TrimSpace(os.Getenv("MEMORY_ONNX_MODEL_PATH")) != "" {
		emb = "onnx" // process env intent; host construction may still fail-open
	}
	names := LeanToolNames()
	return HealthzResponse{
		Status:      "ok",
		Service:     ServerName,
		DualWrite:   "off",
		NotMemoryGA: true,
		Embeddings:  emb,
		Qdrant:      "off",
		Version:     ServerVersion,
		Tools:       len(names),
		ToolNames:   names,
	}
}

// HealthzHandler returns 200 JSON honesty locks for edge probes.
// Optional host argument reports live embedding mode; nil → env snapshot.
// Always unauthenticated (honesty unchanged when an MCP shared secret is set).
func HealthzHandler(hosts ...*Host) http.HandlerFunc {
	var host *Host
	if len(hosts) > 0 {
		host = hosts[0]
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body := HealthzSnapshot(host)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	}
}

// HTTPConfig configures streamable MCP HTTP + healthz.
type HTTPConfig struct {
	Addr string
	Path string
	// Host optional — when set, /healthz reports live EmbeddingMode.
	// tools / tool_names are always compile-time lean registration.
	Host *Host
	// SharedSecret, when non-empty, fail-closes MCP HTTP without a matching
	// X-Memory-MCP-Secret or Authorization: Bearer header. /healthz stays open.
	SharedSecret string
	// AllowNonLoopback permits 0.0.0.0 / :: / non-loopback binds.
	// :port with this false is forced to 127.0.0.1:port.
	AllowNonLoopback bool
}

// RunHTTP serves streamable MCP at Path with GET /healthz and graceful shutdown.
// Listen address is normalized (loopback default). Optional shared secret wraps MCP only.
func RunHTTP(ctx context.Context, sdk *mcp.Server, cfg HTTPConfig) error {
	addr, err := NormalizeListenAddr(cfg.Addr, cfg.AllowNonLoopback)
	if err != nil {
		return err
	}
	path := NormalizeMCPPath(cfg.Path)
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return sdk
	}, &mcp.StreamableHTTPOptions{
		JSONResponse: true,
		// go-sdk v1.7.0 serves protocol 2026-07-28 on HTTP only when Stateless.
		// Keep true so new clients can discover; legacy initialize still works.
		Stateless: true,
	})
	handler = WithOptionalSharedSecret(cfg.SharedSecret, handler)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HealthzHandler(cfg.Host))
	if path == "/" {
		mux.Handle("/", handler)
	} else {
		mux.Handle(path, handler)
		mux.Handle(path+"/", handler)
	}

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	secretState := "off"
	if strings.TrimSpace(cfg.SharedSecret) != "" {
		secretState = "on"
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("%s mode=http addr=%s path=%s healthz=/healthz secret=%s allow_non_loopback=%v tools=%d dual_write=off not_memory_ga=true version=%s (stateless+json)",
			ServerName, addr, path, secretState, cfg.AllowNonLoopback, len(leanToolNames), ServerVersion)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// NormalizeMCPPath ensures a leading slash and no trailing slash (except root).
func NormalizeMCPPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/mcp"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if p != "/" {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// NormalizeListenAddr applies the loopback default for HTTP MCP.
//
//   - empty host (`:8080`) without AllowNonLoopback → 127.0.0.1:8080
//   - 0.0.0.0 / :: / * without AllowNonLoopback → error
//   - loopback (127.0.0.1, ::1, localhost) always allowed
//   - other hosts require AllowNonLoopback
//
// stdio is selected by an empty -http-addr before this is called.
func NormalizeListenAddr(addr string, allowNonLoopback bool) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", fmt.Errorf("http listen address required")
	}
	host, port, err := splitListenAddr(addr)
	if err != nil {
		return "", err
	}
	loopback, unspecifiedAll, unspecifiedEmpty := classifyListenHost(host)
	switch {
	case unspecifiedEmpty && !allowNonLoopback:
		return net.JoinHostPort("127.0.0.1", port), nil
	case unspecifiedEmpty && allowNonLoopback:
		return ":" + port, nil
	case unspecifiedAll && !allowNonLoopback:
		return "", fmt.Errorf("http listen %q is not loopback; set -allow-non-loopback or MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK", addr)
	case loopback:
		return net.JoinHostPort(host, port), nil
	case allowNonLoopback:
		return net.JoinHostPort(host, port), nil
	default:
		return "", fmt.Errorf("http listen %q is not loopback; set -allow-non-loopback or MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK", addr)
	}
}

func splitListenAddr(addr string) (host, port string, err error) {
	if strings.HasPrefix(addr, ":") && !strings.HasPrefix(addr, "[") {
		port = strings.TrimPrefix(addr, ":")
		if port == "" || strings.Contains(port, ":") {
			return "", "", fmt.Errorf("invalid http listen address %q", addr)
		}
		return "", port, nil
	}
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return "", "", fmt.Errorf("invalid http listen address %q", addr)
	}
	return host, port, nil
}

func classifyListenHost(host string) (loopback, unspecifiedAll, unspecifiedEmpty bool) {
	h := strings.Trim(host, "[]")
	if h == "" {
		return false, false, true
	}
	if h == "0.0.0.0" || h == "::" || h == "*" {
		return false, true, false
	}
	if strings.EqualFold(h, "localhost") {
		return true, false, false
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback(), ip.IsUnspecified(), false
	}
	return false, false, false
}

// WithOptionalSharedSecret wraps MCP HTTP. Empty secret is a no-op.
// When set, missing/wrong secret → 401 (fail-closed). Does not wrap /healthz.
func WithOptionalSharedSecret(secret string, next http.Handler) http.Handler {
	want := strings.TrimSpace(secret)
	if want == "" || next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !mcpSecretMatch(r, want) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mcpSecretMatch(r *http.Request, want string) bool {
	if r == nil {
		return false
	}
	got := strings.TrimSpace(r.Header.Get(MCPSecretHeader))
	if got == "" {
		got = bearerToken(r.Header.Get("Authorization"))
	}
	return secretEqual(got, want)
}

func bearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	const prefix = "bearer "
	if len(auth) < len(prefix) {
		return ""
	}
	if !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

func secretEqual(got, want string) bool {
	gh := sha256.Sum256([]byte(got))
	wh := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gh[:], wh[:]) == 1
}
