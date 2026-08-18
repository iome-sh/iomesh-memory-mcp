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
	defTenant := strings.TrimSpace(cfg.DefaultTenant)
	if defTenant != "" {
		if err := validateTenantSegment(defTenant); err != nil {
			return nil, fmt.Errorf("mcphost: default tenant: %w", err)
		}
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
			DefaultTenant: defTenant,
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

// validateTenantSegment allows one path segment so TenantDir stays under palace
// root. Rejects ".", "..", and separators. Path-based isolation only — not
// hosted multi-tenant security.
func validateTenantSegment(tenant string) error {
	if tenant == "" || tenant == "." || tenant == ".." {
		return fmt.Errorf("tenant %q must be a single path segment", tenant)
	}
	if strings.ContainsAny(tenant, `/\`) || strings.IndexByte(tenant, 0) >= 0 {
		return fmt.Errorf("tenant %q must be a single path segment", tenant)
	}
	if !filepath.IsLocal(tenant) || filepath.Base(tenant) != tenant {
		return fmt.Errorf("tenant %q must be a single path segment", tenant)
	}
	return nil
}

// ResolveTenant returns a non-empty tenant id (input, else default, else "default").
// A provided tenant must be a single path segment (no separators, not "." or "..").
// Empty/omitted uses the configured default (already validated in New).
func (h *Host) ResolveTenant(tenant string) (string, error) {
	tenant = strings.TrimSpace(tenant)
	if tenant == "" {
		if h != nil && h.cfg.DefaultTenant != "" {
			tenant = h.cfg.DefaultTenant
		} else {
			tenant = "default"
		}
	}
	if err := validateTenantSegment(tenant); err != nil {
		return "", err
	}
	return tenant, nil
}

// TenantDir returns filepath.Join(palaceRoot, tenant) for path-based isolation.
// Invalid tenant returns "" (do not join traversal / multi-segment values).
func (h *Host) TenantDir(tenant string) string {
	key, err := h.ResolveTenant(tenant)
	if err != nil {
		return ""
	}
	return filepath.Join(h.cfg.PalaceRoot, key)
}

// Store returns (creating if needed) the PalaceStore for tenant.
// Multi-tenant isolation is path-based only; this is one process residual-honest.
// Invalid tenant (not a single path segment) returns nil.
// Embeddings: hash default · optional ONNX when host was constructed with MEMORY_ONNX_MODEL_PATH.
// Qdrant is not attached here (lean FS hybrid + EmbeddingFunc re-rank only).
func (h *Host) Store(tenant string) *palace.PalaceStore {
	key, err := h.ResolveTenant(tenant)
	if err != nil {
		return nil
	}
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

// resolveStore resolves tenant then returns the palace store. Invalid tenant
// is an error (fail closed; do not join outside palace root).
func (h *Host) resolveStore(tenant string) (string, *palace.PalaceStore, error) {
	key, err := h.ResolveTenant(tenant)
	if err != nil {
		return "", nil, err
	}
	return key, h.Store(key), nil
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
		Description: "Ingest a conversation turn into the local tenant palace FS (role=user|assistant|tool). dual_write OFF · not Memory GA",
	}, h.handleIngestTurn)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_write",
		Description: "Write a durable fact to the local palace FS via kernel Write (optional WriteAndSupersede when entity_key set). dual_write OFF · not Memory GA",
	}, h.handleWrite)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_retrieve",
		Description: "Read/search the local palace FS (SearchMemoryWithOptions; keyword + optional vector re-rank). Does not ingest. dual_write OFF · not Memory GA",
	}, h.handleRetrieve)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_search_semantic",
		Description: "Read/search local tier-4 semantic facts (hybrid; no Qdrant). Does not ingest. dual_write OFF · not Memory GA",
	}, h.handleSearchSemantic)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_list",
		Description: "List local palace FS entries by event time. Read/list only; does not ingest. dual_write OFF · not Memory GA",
	}, h.handleList)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_compact_status",
		Description: "Local palace FS tier counts (GetStats). Does not ingest. dual_write OFF · not Memory GA",
	}, h.handleCompactStatus)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_facts_as_of",
		Description: "List local facts valid at as_of (bi-temporal lite; not full dual-clock KG). Does not ingest. not Memory GA",
	}, h.handleFactsAsOf)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_related",
		Description: "Multi-hop lite retrieve on local FS (entity BFS). Does not ingest. dual_write OFF · not Memory GA",
	}, h.handleRelated)

	mcp.AddTool(sdkServer, &mcp.Tool{
		Name:        "memory_supersede_entity",
		Description: "Close open facts for an entity key on the local palace FS (SupersedeEntityFacts). Mutating; HITL stays at the client. dual_write OFF · not Memory GA",
	}, h.handleSupersedeEntity)

	log.Printf("mcphost: registered tools=%d server=%s version=%s dual_write=off not_memory_ga=true",
		len(leanToolNames), ServerName, ServerVersion)
}
