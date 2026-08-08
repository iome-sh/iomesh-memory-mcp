package mcphost

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
	Version     string `json:"version,omitempty"`
}

// HealthzHandler returns 200 JSON honesty locks for edge probes.
func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body := HealthzResponse{
			Status:      "ok",
			Service:     ServerName,
			DualWrite:   "off",
			NotMemoryGA: true,
			Version:     ServerVersion,
		}
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
}

// RunHTTP serves streamable MCP at Path with GET /healthz and graceful shutdown.
func RunHTTP(ctx context.Context, sdk *mcp.Server, cfg HTTPConfig) error {
	path := NormalizeMCPPath(cfg.Path)
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return sdk
	}, &mcp.StreamableHTTPOptions{
		JSONResponse: true,
		Stateless:    true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HealthzHandler())
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
		log.Printf("%s mode=http addr=%s path=%s healthz=/healthz dual_write=off not_memory_ga=true version=%s (stateless+json)",
			ServerName, cfg.Addr, path, ServerVersion)
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
