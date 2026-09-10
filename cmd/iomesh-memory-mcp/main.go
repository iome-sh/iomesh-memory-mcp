// Command iomesh-memory-mcp is the lean edge Memory MCP host (Option A M2 / s1457).
//
// Default transport is stdio; set -http-addr (or MEMORY_MCP_HTTP_ADDR) for
// streamable HTTP. :port binds 127.0.0.1 unless -allow-non-loopback.
// Optional MEMORY_MCP_HTTP_SECRET fail-closes MCP HTTP when set (/healthz stays open).
// -preflight prints the same honesty JSON as GET /healthz and exits
// (no listen, no stdio MCP). Does not import private control-plane/broker packages.
// dual_write OFF · not Memory GA.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/iome-sh/iomesh-memory-mcp/internal/mcphost"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		if errors.Is(err, errFlag) {
			os.Exit(2)
		}
		log.Fatal(err)
	}
}

// errFlag marks FlagSet parse errors (usage already written to stderr).
var errFlag = errors.New("flag")

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("iomesh-memory-mcp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	defaultPalace := envOr("PALACE_ROOT", defaultPalaceRoot())
	palaceRoot := fs.String("palace-root", defaultPalace, "tenant palace root base directory")
	tenant := fs.String("tenant", envOr("MEMORY_TENANT", ""), "process tenant label (validated if set; tool tenant is required — omit fail-closes)")
	httpAddr := fs.String("http-addr", envOr("MEMORY_MCP_HTTP_ADDR", ""),
		"listen address for streamable HTTP (e.g. :8080 → 127.0.0.1:8080); empty = stdio mode")
	httpPath := fs.String("http-path", envOr("MEMORY_MCP_HTTP_PATH", "/mcp"),
		"URL path for the MCP streamable HTTP endpoint (healthz always at /healthz)")
	allowNonLoopback := fs.Bool("allow-non-loopback", envTruthy("MEMORY_MCP_HTTP_ALLOW_NON_LOOPBACK"),
		"allow HTTP bind on 0.0.0.0 / :: / non-loopback (required for container publish). Default: :port is forced to 127.0.0.1; 0.0.0.0 is refused")
	httpSecret := fs.String("http-secret", envOr("MEMORY_MCP_HTTP_SECRET", ""),
		"optional shared secret for streamable HTTP MCP (X-Memory-MCP-Secret or Authorization: Bearer). Empty = off. Fail-closed when set. /healthz stays open. stdio unchanged")
	preflight := fs.Bool("preflight", false,
		"print the same honesty JSON as GET /healthz and exit (no listen, no stdio MCP; not tools/list, not ingest)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return fmt.Errorf("%w: %v", errFlag, err)
	}

	host, err := mcphost.New(mcphost.Config{
		PalaceRoot:    *palaceRoot,
		DefaultTenant: *tenant,
	})
	if err != nil {
		return fmt.Errorf("mcphost: %w", err)
	}

	if *preflight {
		// Same HealthzResponse fields as GET /healthz. Registration ≠ tools/list ≠ ingest.
		if err := json.NewEncoder(stdout).Encode(mcphost.HealthzSnapshot(host)); err != nil {
			return fmt.Errorf("preflight: %w", err)
		}
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sdk := host.NewSDKServer()
	addr := strings.TrimSpace(*httpAddr)
	if addr != "" {
		resolved, err := mcphost.NormalizeListenAddr(addr, *allowNonLoopback)
		if err != nil {
			return fmt.Errorf("http: %w", err)
		}
		if err := mcphost.RunHTTP(ctx, sdk, mcphost.HTTPConfig{
			Addr:             resolved,
			Path:             *httpPath,
			Host:             host,
			SharedSecret:     strings.TrimSpace(*httpSecret),
			AllowNonLoopback: *allowNonLoopback,
		}); err != nil {
			return fmt.Errorf("http: %w", err)
		}
		return nil
	}

	log.Printf("%s mode=stdio palace=%s tenant_process=%q embeddings=%s qdrant=off dual_write=off not_memory_ga=true version=%s",
		mcphost.ServerName, *palaceRoot, host.ConfiguredTenant(), host.EmbeddingMode(), mcphost.ServerVersion)
	if err := sdk.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}

// defaultPalaceRoot prefers local dogfood path; containers often bind /data.
func defaultPalaceRoot() string {
	if _, err := os.Stat("/data"); err == nil {
		return "/data/memory-palaces"
	}
	// Local relative path for clone/dogfood without root.
	wd, err := os.Getwd()
	if err != nil {
		return "./data/memory-palaces"
	}
	return filepath.Join(wd, "data", "memory-palaces")
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envTruthy(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
