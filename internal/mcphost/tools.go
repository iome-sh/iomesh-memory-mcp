package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	palace "github.com/iome-sh/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- memory_ingest_turn ---

type ingestTurnInput struct {
	Tenant     string `json:"tenant,omitempty" jsonschema:"tenant subdirectory under palace root"`
	SessionID  string `json:"session_id" jsonschema:"conversation session id"`
	Role       string `json:"role" jsonschema:"user|assistant|tool"`
	Content    string `json:"content" jsonschema:"turn text content"`
	Timestamp  string `json:"timestamp,omitempty" jsonschema:"optional RFC3339 timestamp (legacy)"`
	EventTime  string `json:"event_time,omitempty" jsonschema:"optional RFC3339 source event time"`
	SessionSeq int    `json:"session_seq,omitempty" jsonschema:"optional monotonic order within session"`
	TurnID     string `json:"turn_id,omitempty"`
	MemoryID   string `json:"memory_id,omitempty"`
	Tier       int    `json:"tier,omitempty" jsonschema:"optional MemoryTier 1..4 (default working=1)"`
}

type ingestTurnOutput struct {
	MemoryID  string `json:"memory_id"`
	Tier      int    `json:"tier"`
	Tenant    string `json:"tenant"`
	Audited   bool   `json:"audited"` // always false in lean v1 (dual_write OFF)
	DualWrite string `json:"dual_write"`
}

func (h *Host) handleIngestTurn(_ context.Context, _ *mcp.CallToolRequest, in ingestTurnInput) (*mcp.CallToolResult, ingestTurnOutput, error) {
	content := strings.TrimSpace(in.Content)
	if content == "" {
		err := fmt.Errorf("content required")
		return toolError(err), ingestTurnOutput{}, err
	}
	role := strings.ToLower(strings.TrimSpace(in.Role))
	if role == "" {
		role = "user"
	}
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)

	ts := parseTimeOrNow(firstNonEmpty(in.EventTime, in.Timestamp))
	id := strings.TrimSpace(in.MemoryID)
	if id == "" {
		id = palace.GenerateMemoryID()
	}
	tier := palace.MemoryTier(in.Tier)
	if tier == 0 {
		tier = palace.TierWorking
	}

	tags := []string{"role:" + role, "source:iomesh-memory-mcp"}
	if in.SessionSeq > 0 {
		tags = append(tags, fmt.Sprintf("session_seq:%d", in.SessionSeq))
	}

	entry := palace.MemoryEntry{
		ID:           id,
		Type:         "turn",
		Tier:         tier,
		Version:      1,
		CreatedAt:    ts,
		UpdatedAt:    ts,
		Timestamp:    ts,
		TurnID:       strings.TrimSpace(in.TurnID),
		SessionID:    strings.TrimSpace(in.SessionID),
		OriginalText: content,
		Content: palace.MemoryContent{
			Summary: truncate(content, 280),
			Full:    content,
			Tags:    tags,
		},
		Provenance: palace.MemoryProvenance{
			SourceStep: "mcp_memory_ingest_turn",
		},
		Metrics: palace.MemoryMetrics{
			UsageCount: 1,
		},
	}

	if err := ps.IngestTurn(entry); err != nil {
		return toolError(err), ingestTurnOutput{}, err
	}

	out := ingestTurnOutput{
		MemoryID:  id,
		Tier:      int(tier),
		Tenant:    tenant,
		Audited:   false,
		DualWrite: "off",
	}
	return toolJSON(out), out, nil
}

// --- memory_write (durable fact; kernel Write / WriteAndSupersede) ---

type writeInput struct {
	Tenant    string   `json:"tenant,omitempty" jsonschema:"tenant subdirectory under palace root"`
	Summary   string   `json:"summary,omitempty" jsonschema:"short fact summary"`
	Full      string   `json:"full,omitempty" jsonschema:"full fact text (defaults to summary)"`
	Tags      []string `json:"tags,omitempty"`
	Tier      int      `json:"tier,omitempty" jsonschema:"optional MemoryTier 1..4 (default contextual=2)"`
	EntityKey string   `json:"entity_key,omitempty" jsonschema:"optional entity key; stamps entity: tag"`
	MemoryID  string   `json:"memory_id,omitempty"`
	// Supersede, when true (default if entity_key set), calls WriteAndSupersede.
	Supersede *bool `json:"supersede,omitempty" jsonschema:"when entity_key set, default true → WriteAndSupersede"`
}

type writeOutput struct {
	MemoryID   string `json:"memory_id"`
	Tier       int    `json:"tier"`
	Tenant     string `json:"tenant"`
	Superseded bool   `json:"superseded"`
	Audited    bool   `json:"audited"`
	DualWrite  string `json:"dual_write"`
}

func (h *Host) handleWrite(_ context.Context, _ *mcp.CallToolRequest, in writeInput) (*mcp.CallToolResult, writeOutput, error) {
	summary := strings.TrimSpace(in.Summary)
	full := strings.TrimSpace(in.Full)
	if summary == "" && full == "" {
		err := fmt.Errorf("summary or full required")
		return toolError(err), writeOutput{}, err
	}
	if summary == "" {
		summary = truncate(full, 280)
	}
	if full == "" {
		full = summary
	}

	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)
	ts := time.Now().UTC()
	id := strings.TrimSpace(in.MemoryID)
	if id == "" {
		id = palace.GenerateMemoryID()
	}
	tier := palace.MemoryTier(in.Tier)
	if tier == 0 {
		tier = palace.TierContextual
	}

	tags := append([]string{"source:iomesh-memory-mcp", "type:fact"}, in.Tags...)
	var temporal []string
	entityKey := strings.TrimSpace(in.EntityKey)
	if tag := entityTag(entityKey); tag != "" {
		tags = append(tags, tag)
		temporal = append(temporal, tag)
	}

	entry := palace.MemoryEntry{
		ID:           id,
		Type:         "fact",
		Tier:         tier,
		Version:      1,
		CreatedAt:    ts,
		UpdatedAt:    ts,
		Timestamp:    ts,
		TemporalTags: temporal,
		OriginalText: full,
		Content: palace.MemoryContent{
			Summary: summary,
			Full:    full,
			Tags:    tags,
		},
		Provenance: palace.MemoryProvenance{
			SourceStep: "mcp_memory_write",
		},
		Metrics: palace.MemoryMetrics{
			UsageCount: 1,
		},
	}

	doSuper := entityKey != ""
	if in.Supersede != nil {
		doSuper = *in.Supersede && entityKey != ""
	}

	var err error
	if doSuper {
		err = ps.WriteAndSupersede(entry, []string{entityKey})
	} else {
		err = ps.Write(entry)
	}
	if err != nil {
		return toolError(err), writeOutput{}, err
	}

	out := writeOutput{
		MemoryID:   id,
		Tier:       int(tier),
		Tenant:     tenant,
		Superseded: doSuper,
		Audited:    false,
		DualWrite:  "off",
	}
	return toolJSON(out), out, nil
}

func entityTag(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(key), "entity:") {
		return key
	}
	return "entity:" + key
}

// --- memory_retrieve ---

type retrieveInput struct {
	Tenant    string `json:"tenant,omitempty" jsonschema:"tenant subdirectory under palace root"`
	Query     string `json:"query" jsonschema:"recall query text"`
	Limit     int    `json:"limit,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Since     string `json:"since,omitempty" jsonschema:"optional RFC3339 inclusive lower bound"`
	Until     string `json:"until,omitempty" jsonschema:"optional RFC3339 inclusive upper bound"`
}

type memoryHit struct {
	ID        string   `json:"id"`
	Tier      int      `json:"tier"`
	SessionID string   `json:"session_id,omitempty"`
	TurnID    string   `json:"turn_id,omitempty"`
	Summary   string   `json:"summary"`
	Full      string   `json:"full,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
}

type retrieveOutput struct {
	Memories []memoryHit `json:"memories"`
	Tenant   string      `json:"tenant"`
	Mode     string      `json:"mode"`
}

func (h *Host) handleRetrieve(_ context.Context, _ *mcp.CallToolRequest, in retrieveInput) (*mcp.CallToolResult, retrieveOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		err := fmt.Errorf("query required")
		return toolError(err), retrieveOutput{}, err
	}
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)

	opts := palace.SearchMemoryOptions{
		SessionID: strings.TrimSpace(in.SessionID),
		Limit:     in.Limit,
	}
	if t, ok := parseOptionalTime(in.Since); ok {
		opts.TimeFrom = &t
	}
	if t, ok := parseOptionalTime(in.Until); ok {
		opts.TimeTo = &t
	}

	// Hash embeddings are random unit vectors (SHA-256 seed). Injecting them as
	// QueryVec used to skip the kernel keyword path (#21 / kernel #45).
	// Only pass a query vector when a real embedder (ONNX) is loaded.
	opts.QueryVec = h.searchQueryVec(ps, query)

	entries := ps.SearchMemoryWithOptions(query, opts)
	hits := make([]memoryHit, 0, len(entries))
	for _, e := range entries {
		hits = append(hits, hitFromEntry(e))
	}
	out := retrieveOutput{
		Memories: hits,
		Tenant:   tenant,
		Mode:     "hybrid_search_memory_with_options",
	}
	return toolJSON(out), out, nil
}

// searchQueryVec returns a query embedding only for a real embedder (ONNX).
// Hash mode must not set QueryVec: SHA-256 random vectors skip kernel keyword
// matching and can drop an exact token past Limit (issue #21 / memory#45).
func (h *Host) searchQueryVec(ps *palace.PalaceStore, query string) []float32 {
	if h == nil || h.EmbeddingMode() == "hash" {
		return nil
	}
	if ps == nil || ps.Config.EmbeddingFunc == nil {
		return nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	dim := h.embedDim
	if dim <= 0 {
		dim = palace.ResolveEmbeddingDimFromEnv()
	}
	if dim <= 0 {
		dim = palace.DefaultHashEmbeddingDim
	}
	return ps.Config.EmbeddingFunc(query, dim)
}

// --- memory_search_semantic ---

type searchSemanticInput struct {
	Tenant string `json:"tenant,omitempty"`
	Query  string `json:"query" jsonschema:"filter or hybrid query over semantic tier"`
	Limit  int    `json:"limit,omitempty"`
}

type searchSemanticOutput struct {
	Facts  []memoryHit `json:"facts"`
	Tenant string      `json:"tenant"`
	// Residual notes honesty: full product semantic stack residual; kernel hybrid on TierSemantic.
	Note string `json:"note"`
}

func (h *Host) handleSearchSemantic(_ context.Context, _ *mcp.CallToolRequest, in searchSemanticInput) (*mcp.CallToolResult, searchSemanticOutput, error) {
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)
	query := strings.TrimSpace(in.Query)
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	tier := palace.TierSemantic
	opts := palace.SearchMemoryOptions{
		Tier:  &tier,
		Limit: limit,
	}
	if query != "" {
		opts.QueryVec = h.searchQueryVec(ps, query)
	}
	entries := ps.SearchMemoryWithOptions(query, opts)
	// Fallback: if hybrid returns empty and query set, substring filter tier listing.
	if len(entries) == 0 && query != "" {
		all := ps.ListEntriesInTier(palace.TierSemantic)
		q := strings.ToLower(query)
		for _, e := range all {
			text := strings.ToLower(firstNonEmpty(e.Content.Full, e.Content.Summary, e.OriginalText))
			if strings.Contains(text, q) {
				entries = append(entries, e)
				if len(entries) >= limit {
					break
				}
			}
		}
	} else if len(entries) == 0 && query == "" {
		entries = ps.ListEntriesInTier(palace.TierSemantic)
		if len(entries) > limit {
			entries = entries[:limit]
		}
	}

	facts := make([]memoryHit, 0, len(entries))
	for _, e := range entries {
		facts = append(facts, hitFromEntry(e))
	}
	out := searchSemanticOutput{
		Facts:  facts,
		Tenant: tenant,
		Note:   "semantic_tier_hybrid; default hash · optional ONNX via MEMORY_ONNX_MODEL_PATH · Qdrant not wired lean host; not Memory GA",
	}
	return toolJSON(out), out, nil
}

// --- memory_list ---

type listInput struct {
	Tenant          string `json:"tenant,omitempty"`
	SessionID       string `json:"session_id,omitempty"`
	Query           string `json:"query,omitempty" jsonschema:"optional substring filter"`
	Since           string `json:"since,omitempty" jsonschema:"RFC3339 inclusive lower bound"`
	Until           string `json:"until,omitempty" jsonschema:"RFC3339 inclusive upper bound"`
	Tag             string `json:"tag,omitempty"`
	TagPrefix       string `json:"tag_prefix,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	IncludeArchival bool   `json:"include_archival,omitempty"`
	Ascending       bool   `json:"ascending,omitempty"`
}

type listOutput struct {
	Entries []memoryHit `json:"entries"`
	Tenant  string      `json:"tenant"`
}

func (h *Host) handleList(_ context.Context, _ *mcp.CallToolRequest, in listInput) (*mcp.CallToolResult, listOutput, error) {
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)
	opts := palace.ListMemoryOptions{
		SessionID:       strings.TrimSpace(in.SessionID),
		Query:           strings.TrimSpace(in.Query),
		Tag:             strings.TrimSpace(in.Tag),
		TagPrefix:       strings.TrimSpace(in.TagPrefix),
		Limit:           in.Limit,
		IncludeArchival: in.IncludeArchival,
		Ascending:       in.Ascending,
	}
	if t, ok := parseOptionalTime(in.Since); ok {
		opts.TimeFrom = &t
	}
	if t, ok := parseOptionalTime(in.Until); ok {
		opts.TimeTo = &t
	}
	entries := ps.ListMemoryWithOptions(opts)
	hits := make([]memoryHit, 0, len(entries))
	for _, e := range entries {
		hits = append(hits, hitFromEntry(e))
	}
	out := listOutput{Entries: hits, Tenant: tenant}
	return toolJSON(out), out, nil
}

// --- memory_compact_status ---

type compactStatusInput struct {
	Tenant string `json:"tenant,omitempty"`
}

type compactStatusOutput struct {
	Tenant          string `json:"tenant"`
	WorkingCount    int    `json:"working_count"`
	ContextualCount int    `json:"contextual_count"`
	ArchivalCount   int    `json:"archival_count"`
	SemanticCount   int    `json:"semantic_count"`
	TotalEntries    int    `json:"total_entries"`
	LastCompaction  string `json:"last_compaction,omitempty"`
	DualWrite       string `json:"dual_write"`
	NotMemoryGA     bool   `json:"not_memory_ga"`
}

func (h *Host) handleCompactStatus(_ context.Context, _ *mcp.CallToolRequest, in compactStatusInput) (*mcp.CallToolResult, compactStatusOutput, error) {
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)
	stats := ps.GetStats()
	out := compactStatusOutput{
		Tenant:          tenant,
		WorkingCount:    stats.WorkingCount,
		ContextualCount: stats.ContextualCount,
		ArchivalCount:   stats.ArchivalCount,
		SemanticCount:   stats.SemanticCount,
		TotalEntries:    stats.TotalEntries,
		DualWrite:       "off",
		NotMemoryGA:     true,
	}
	if !stats.LastCompaction.IsZero() {
		out.LastCompaction = stats.LastCompaction.UTC().Format(time.RFC3339)
	}
	return toolJSON(out), out, nil
}

// --- memory_facts_as_of ---

type factsAsOfInput struct {
	Tenant    string `json:"tenant,omitempty"`
	AsOf      string `json:"as_of,omitempty" jsonschema:"RFC3339 validity instant (default now)"`
	Query     string `json:"query,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Entity    string `json:"entity,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type factsAsOfOutput struct {
	Facts  []memoryHit `json:"facts"`
	Tenant string      `json:"tenant"`
	AsOf   string      `json:"as_of"`
	Note   string      `json:"note"`
}

func (h *Host) handleFactsAsOf(_ context.Context, _ *mcp.CallToolRequest, in factsAsOfInput) (*mcp.CallToolResult, factsAsOfOutput, error) {
	tenant := h.ResolveTenant(in.Tenant)
	ps := h.Store(tenant)
	asOf := parseTimeOrNow(in.AsOf)
	opts := palace.FactsAsOfOptions{
		AsOf:      asOf,
		Query:     strings.TrimSpace(in.Query),
		SessionID: strings.TrimSpace(in.SessionID),
		Entity:    strings.TrimSpace(in.Entity),
		Limit:     in.Limit,
	}
	entries := ps.ListFactsAsOf(opts)
	hits := make([]memoryHit, 0, len(entries))
	for _, e := range entries {
		hits = append(hits, hitFromEntry(e))
	}
	out := factsAsOfOutput{
		Facts:  hits,
		Tenant: tenant,
		AsOf:   asOf.UTC().Format(time.RFC3339),
		Note:   "bi-temporal lite (valid_from/valid_until tags); not full Graphiti dual-clock KG; not Memory GA",
	}
	return toolJSON(out), out, nil
}

// --- helpers ---

func hitFromEntry(e palace.MemoryEntry) memoryHit {
	ts := e.Timestamp
	if ts.IsZero() {
		ts = e.CreatedAt
	}
	tsStr := ""
	if !ts.IsZero() {
		tsStr = ts.UTC().Format(time.RFC3339)
	}
	return memoryHit{
		ID:        e.ID,
		Tier:      int(e.Tier),
		SessionID: e.SessionID,
		TurnID:    e.TurnID,
		Summary:   e.Content.Summary,
		Full:      firstNonEmpty(e.Content.Full, e.OriginalText),
		Tags:      e.Content.Tags,
		Timestamp: tsStr,
	}
}

func toolJSON(v any) *mcp.CallToolResult {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("marshal error: %v", err)}},
			IsError: true,
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func parseOptionalTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
	}
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func parseTimeOrNow(s string) time.Time {
	if t, ok := parseOptionalTime(s); ok {
		return t
	}
	return time.Now().UTC()
}
