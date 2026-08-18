// Command iomesh-memory-mcp is the lean edge Memory MCP host (Option A M2 / s1457).
//
// Default transport is stdio; set -http-addr (or MEMORY_MCP_HTTP_ADDR) for
// streamable HTTP. -preflight prints the same honesty JSON as GET /healthz
// and exits (no listen, no stdio MCP). Does not import aion.
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
	tenant := fs.String("tenant", envOr("MEMORY_TENANT", ""), "default tenant subdirectory under palace root")
	httpAddr := fs.String("http-addr", firstEnvPrefer(
		"MEMORY_MCP_HTTP_ADDR",
		"AION_MEMORY_MCP_HTTP_ADDR",
	), "listen address for streamable HTTP (e.g. :8080); empty = stdio mode")
	httpPath := fs.String("http-path", envOr("MEMORY_MCP_HTTP_PATH", "/mcp"),
		"URL path for the MCP streamable HTTP endpoint (healthz always at /healthz)")
	preflight := fs.Bool("preflight", false,
		"print the same honesty JSON as GET /healthz and exit (no listen, no stdio MCP; not tools/list, not ingest)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return fmt.Errorf("%w: %v", errFlag, err)
	}

	warnDeprecatedEnvAliases()

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
		if err := mcphost.RunHTTP(ctx, sdk, mcphost.HTTPConfig{
			Addr: addr,
			Path: *httpPath,
			Host: host,
		}); err != nil {
			return fmt.Errorf("http: %w", err)
		}
		return nil
	}

	defaultTenant, err := host.ResolveTenant("")
	if err != nil {
		log.Fatalf("mcphost: %v", err)
	}
	log.Printf("%s mode=stdio palace=%s tenant_default=%q embeddings=%s qdrant=off dual_write=off not_memory_ga=true version=%s",
		mcphost.ServerName, *palaceRoot, defaultTenant, host.EmbeddingMode(), mcphost.ServerVersion)
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

// firstEnvPrefer returns the first non-empty among preferred then deprecated keys.
// Logs once when only the deprecated alias is set.
func firstEnvPrefer(prefer, deprecated string) string {
	if v := strings.TrimSpace(os.Getenv(prefer)); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv(deprecated)); v != "" {
		log.Printf("deprecated env %s is set; prefer %s (one-time notice)", deprecated, prefer)
		return v
	}
	return ""
}

func warnDeprecatedEnvAliases() {
	// Cover additional legacy aliases used in private aion installs.
	pairs := [][2]string{
		{"AION_MEMORY_MCP_HTTP_ADDR", "MEMORY_MCP_HTTP_ADDR"},
		{"AION_MEMORY_MCP_HTTP_PATH", "MEMORY_MCP_HTTP_PATH"},
		{"AION_PALACE_ROOT", "PALACE_ROOT"},
	}
	for _, p := range pairs {
		if strings.TrimSpace(os.Getenv(p[0])) != "" && strings.TrimSpace(os.Getenv(p[1])) == "" {
			log.Printf("deprecated env %s is set; prefer %s (one-time notice)", p[0], p[1])
		}
	}
}
