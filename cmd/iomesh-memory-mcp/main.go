// Command iomesh-memory-mcp is the lean edge Memory MCP host (Option A M2 / s1457).
//
// Default transport is stdio; set -http-addr (or MEMORY_MCP_HTTP_ADDR) for
// streamable HTTP. Does not import aion. dual_write OFF · not Memory GA.
package main

import (
	"context"
	"flag"
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
	defaultPalace := envOr("PALACE_ROOT", defaultPalaceRoot())
	palaceRoot := flag.String("palace-root", defaultPalace, "tenant palace root base directory")
	tenant := flag.String("tenant", envOr("MEMORY_TENANT", ""), "default tenant subdirectory under palace root")
	httpAddr := flag.String("http-addr", firstEnvPrefer(
		"MEMORY_MCP_HTTP_ADDR",
		"AION_MEMORY_MCP_HTTP_ADDR",
	), "listen address for streamable HTTP (e.g. :8080); empty = stdio mode")
	httpPath := flag.String("http-path", envOr("MEMORY_MCP_HTTP_PATH", "/mcp"),
		"URL path for the MCP streamable HTTP endpoint (healthz always at /healthz)")
	flag.Parse()

	warnDeprecatedEnvAliases()

	host, err := mcphost.New(mcphost.Config{
		PalaceRoot:    *palaceRoot,
		DefaultTenant: *tenant,
	})
	if err != nil {
		log.Fatalf("mcphost: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sdk := host.NewSDKServer()
	addr := strings.TrimSpace(*httpAddr)
	if addr != "" {
		if err := mcphost.RunHTTP(ctx, sdk, mcphost.HTTPConfig{
			Addr: addr,
			Path: *httpPath,
		}); err != nil {
			log.Fatalf("http: %v", err)
		}
		return
	}

	log.Printf("%s mode=stdio palace=%s tenant_default=%q dual_write=off not_memory_ga=true version=%s",
		mcphost.ServerName, *palaceRoot, host.ResolveTenant(""), mcphost.ServerVersion)
	if err := sdk.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		log.Fatalf("mcp server: %v", err)
	}
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
