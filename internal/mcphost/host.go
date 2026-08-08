// Package mcphost is the lean edge Memory MCP host.
//
// Architecture (Option A M2 / s1457):
//
//	stdio or streamable HTTP → MCP tools → github.com/iome-sh/memory PalaceStore
//
// Honesty locks:
//   - dual_write OFF (no aion audit publish in lean v1)
//   - not product Memory GA
//   - does not import github.com/iome-sh/aion/**
//   - path-based multi-tenant FS isolation only (same process residual)
package mcphost

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	palace "github.com/iome-sh/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// ServerName is the MCP implementation name (product edge honesty).
	ServerName = "iomesh-memory-mcp"
	// ServerVersion stamps the s1457 lean scaffold.
	ServerVersion = "v0.1.0-s1457"
)

// Config configures the lean edge host.
type Config struct {
	// PalaceRoot is the base directory; each tenant is a subdirectory.
	PalaceRoot string
	// DefaultTenant used when tool input omits tenant.
	DefaultTenant string
}

// Host owns per-tenant PalaceStore instances and MCP registration.
type Host struct {
	cfg    Config
	mu     sync.Mutex
	stores map[string]*palace.PalaceStore
}

// New constructs a Host. PalaceRoot must be non-empty.
func New(cfg Config) (*Host, error) {
	root := strings.TrimSpace(cfg.PalaceRoot)
	if root == "" {
		return nil, fmt.Errorf("mcphost: palace root required")
	}
	return &Host{
		cfg: Config{
			PalaceRoot:    root,
			DefaultTenant: strings.TrimSpace(cfg.DefaultTenant),
		},
		stores: make(map[string]*palace.PalaceStore),
	}, nil
}

// ResolveTenant returns a non-empty tenant id (input, else default, else "default").
func (h *Host) ResolveTenant(tenant string) string {
	tenant = strings.TrimSpace(tenant)
	if tenant != "" {
		return tenant
	}
	if h.cfg.DefaultTenant != "" {
		return h.cfg.DefaultTenant
	}
	return "default"
}

// TenantDir returns filepath.Join(palaceRoot, tenant) for path-based isolation.
func (h *Host) TenantDir(tenant string) string {
	return filepath.Join(h.cfg.PalaceRoot, h.ResolveTenant(tenant))
}

// Store returns (creating if needed) the PalaceStore for tenant.
// Multi-tenant isolation is path-based only; this is one process residual-honest.
func (h *Host) Store(tenant string) *palace.PalaceStore {
	key := h.ResolveTenant(tenant)
	h.mu.Lock()
	defer h.mu.Unlock()
	if ps, ok := h.stores[key]; ok {
		return ps
	}
	base := filepath.Join(h.cfg.PalaceRoot, key)
	ps := palace.NewPalaceStore(base)
	h.stores[key] = ps
	return ps
}

// NewSDKServer builds a configured modelcontextprotocol/go-sdk server with tools.
func (h *Host) NewSDKServer() *mcp.Server {
	sdk := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)
	h.Register(sdk)
	return sdk
}

// Register mounts lean MCP tools on sdkServer.
func (h *Host) Register(sdkServer *mcp.Server) {
	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_ingest_turn",
		Description: "Ingest a conversation turn (role=user|assistant|tool) into the tenant Palace (FS; dual_write OFF)",
	}, h.handleIngestTurn)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_retrieve",
		Description: "Hybrid retrieve memories via SearchMemoryWithOptions (keyword + optional vector re-rank residual)",
	}, h.handleRetrieve)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_search_semantic",
		Description: "Search tier-4 semantic facts (hybrid search restricted to Semantic tier; no Qdrant required)",
	}, h.handleSearchSemantic)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_list",
		Description: "List palace entries by event time (ListMemoryWithOptions; optional session/query/time window)",
	}, h.handleList)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_compact_status",
		Description: "Return Palace tier counts (GetStats); dual_write OFF · not Memory GA",
	}, h.handleCompactStatus)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_facts_as_of",
		Description: "List facts valid at as_of (ListFactsAsOf bi-temporal lite; not full dual-clock KG)",
	}, h.handleFactsAsOf)

	log.Printf("mcphost: registered tools server=%s version=%s dual_write=off not_memory_ga=true",
		ServerName, ServerVersion)
}
