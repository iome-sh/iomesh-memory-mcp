package mcphost

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
}

// RunHTTP serves streamable MCP at Path with GET /healthz and graceful shutdown.
func RunHTTP(ctx context.Context, sdk *mcp.Server, cfg HTTPConfig) error {
	path := NormalizeMCPPath(cfg.Path)
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return sdk
	}, &mcp.StreamableHTTPOptions{
		JSONResponse: true,
		// go-sdk v1.7.0 serves protocol 2026-07-28 on HTTP only when Stateless.
		// Keep true so new clients can discover; legacy initialize still works.
		Stateless: true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HealthzHandler(cfg.Host))
	if path == "/" {
		mux.Handle("/", handler)
	} else {
		mux.Handle(path, handler)
		mux.Handle(path+"/", handler)
	}

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("%s mode=http addr=%s path=%s healthz=/healthz tools=%d dual_write=off not_memory_ga=true version=%s (stateless+json)",
			ServerName, cfg.Addr, path, len(leanToolNames), ServerVersion)
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
