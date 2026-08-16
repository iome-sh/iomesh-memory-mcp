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
	"os"
	"path/filepath"
	"strings"
	"sync"

	palace "github.com/iome-sh/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// ServerName is the MCP implementation name (product edge honesty).
	ServerName = "iomesh-memory-mcp"
)

// ServerVersion is the default MCP implementation version stamp.
// Overridden at link time by GoReleaser / make build via:
//
//	-X github.com/iome-sh/iomesh-memory-mcp/internal/mcphost.ServerVersion=vX.Y.Z
//
// Default is a clean semver-ish pre-release stamp (no private ledger serial).
var ServerVersion = "v0.1.0"

// Config configures the lean edge host.
type Config struct {
	// PalaceRoot is the base directory; each tenant is a subdirectory.
	PalaceRoot string
	// DefaultTenant used when tool input omits tenant.
	DefaultTenant string
	// EmbeddingMode is residual-honest: "hash" (default) or "onnx" when MEMORY_ONNX_MODEL_PATH loads.
	// Qdrant is NOT wired into lean host search (kernel VectorStore residual only).
	EmbeddingMode string
}

// Host owns per-tenant PalaceStore instances and MCP registration.
type Host struct {
	cfg      Config
	mu       sync.Mutex
	stores   map[string]*palace.PalaceStore
	embedFn  palace.EmbeddingFunc
	batchFn  palace.BatchEmbeddingFunc
	embedDim int
}

// New constructs a Host. PalaceRoot must be non-empty.
// Optional advanced embeddings: set MEMORY_ONNX_MODEL_PATH to an ONNX model dir/file
// (see github.com/iome-sh/memory README). Empty path keeps hash embeddings (default).
// dual_write OFF · not Memory GA · Qdrant not required / not wired for lean search.
func New(cfg Config) (*Host, error) {
	root := strings.TrimSpace(cfg.PalaceRoot)
	if root == "" {
		return nil, fmt.Errorf("mcphost: palace root required")
	}
	mode := "hash"
	var embedFn palace.EmbeddingFunc
	var batchFn palace.BatchEmbeddingFunc
	dim := 0
	if path := strings.TrimSpace(os.Getenv(palace.EnvONNXModelPath)); path != "" {
		emb, err := palace.NewGONNXEmbedder(palace.GONNXOptions{ModelPath: path})
		if err != nil {
			return nil, fmt.Errorf("mcphost: ONNX embeddings (MEMORY_ONNX_MODEL_PATH): %w", err)
		}
		embedFn = emb.Func()
		batchFn = emb.BatchFunc()
		dim = emb.Dimension()
		if dim <= 0 {
			dim = palace.ResolveEmbeddingDim(path)
		}
		mode = "onnx"
		log.Printf("mcphost: embeddings=onnx path=%s dim=%d dual_write=off not_memory_ga=true", path, dim)
	} else {
		// Explicit hash path; NewPalaceStoreWithConfig also defaults EmbeddingFunc.
		embedFn = palace.GenerateSimpleEmbedding
		mode = "hash"
		log.Printf("mcphost: embeddings=hash (set MEMORY_ONNX_MODEL_PATH for ONNX) dual_write=off not_memory_ga=true")
	}
	return &Host{
		cfg: Config{
			PalaceRoot:    root,
			DefaultTenant: strings.TrimSpace(cfg.DefaultTenant),
			EmbeddingMode: mode,
		},
		stores:   make(map[string]*palace.PalaceStore),
		embedFn:  embedFn,
		batchFn:  batchFn,
		embedDim: dim,
	}, nil
}

// EmbeddingMode returns residual-honest embedding backend for healthz / operators.
func (h *Host) EmbeddingMode() string {
	if h == nil || strings.TrimSpace(h.cfg.EmbeddingMode) == "" {
		return "hash"
	}
	return h.cfg.EmbeddingMode
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
// Embeddings: hash default · optional ONNX when host was constructed with MEMORY_ONNX_MODEL_PATH.
// Qdrant is not attached here (lean FS hybrid + EmbeddingFunc re-rank only).
func (h *Host) Store(tenant string) *palace.PalaceStore {
	key := h.ResolveTenant(tenant)
	h.mu.Lock()
	defer h.mu.Unlock()
	if ps, ok := h.stores[key]; ok {
		return ps
	}
	base := filepath.Join(h.cfg.PalaceRoot, key)
	cfg := palace.PalaceConfig{
		BaseDir:            base,
		EmbeddingFunc:      h.embedFn,
		BatchEmbeddingFunc: h.batchFn,
	}
	ps := palace.NewPalaceStoreWithConfig(cfg)
	h.stores[key] = ps
	return ps
}

// leanToolNames is the compile-time registered lean MCP surface.
// GET /healthz reports tools + tool_names from this list. That is residual-honest
// registration, not a live MCP tools/list client stamp (s1509 tools=6 at tip
// f46afe2 stays contemporaneous attach evidence — do not restamp as live forever-green).
var leanToolNames = []string{
	"memory_ingest_turn",
	"memory_write",
	"memory_retrieve",
	"memory_search_semantic",
	"memory_list",
	"memory_compact_status",
	"memory_facts_as_of",
	"memory_related",
	"memory_supersede_entity",
}

// LeanToolNames returns a copy of the compile-time lean registered tool names.
func LeanToolNames() []string {
	out := make([]string, len(leanToolNames))
	copy(out, leanToolNames)
	return out
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
		Name:        "memory_write",
		Description: "Write a durable fact via kernel Write (optional WriteAndSupersede when entity_key set). dual_write OFF · not Memory GA",
	}, h.handleWrite)

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

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_related",
		Description: "Multi-hop lite retrieve (MultiHopRetrieve; entity BFS). dual_write OFF · not Memory GA",
	}, h.handleRelated)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_supersede_entity",
		Description: "Close open facts for an entity key (SupersedeEntityFacts). Mutating; HITL stays at the client. dual_write OFF",
	}, h.handleSupersedeEntity)

	log.Printf("mcphost: registered tools=%d server=%s version=%s dual_write=off not_memory_ga=true",
		len(leanToolNames), ServerName, ServerVersion)
}
