package mcphost

import (
	"context"
	"fmt"
	"strings"
	"time"

	palace "github.com/iome-sh/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	opsDigestWindowDay  = "day"
	opsDigestWindowWeek = "week"
	opsDigestLimitDef   = 20
	opsDigestLimitMax   = 50
	// palaceTimelineHint is a TUI-classifiable private source_hint
	// (private / palace / palace_timeline). Never invent mesh.
	palaceTimelineHint = "palace_timeline"
)

// --- ops_digest_export ---

type opsDigestExportInput struct {
	Tenant  string `json:"tenant,omitempty" jsonschema:"optional tenant subdirectory under palace root; omit fail-closes"`
	Window  string `json:"window,omitempty" jsonschema:"day|week (default day)"`
	Horizon string `json:"horizon,omitempty" jsonschema:"ops|knowledge|analytical|all (default ops)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"max receipts (default 20, cap 50)"`
	AsOf    string `json:"as_of,omitempty" jsonschema:"optional RFC3339 upper bound (default now UTC)"`
}

// opsDigestHonesty is residual-honest framing on the lean digest export.
// Field names match TUI MemoryOpsDigestResult.Honesty.
type opsDigestHonesty struct {
	OpsPulse         string `json:"ops_pulse"`
	Knowledge        string `json:"knowledge"`
	Analytical       string `json:"analytical"`
	NeverInventGA    bool   `json:"never_invent_ga"`
	DualWriteDefault string `json:"dual_write_default"`
	BookDemo         string `json:"book_demo"`
	Note             string `json:"note,omitempty"`
}

// opsDigestReceipt is one local palace timeline receipt.
// Field names match TUI MemoryOpsDigestReceipt.
type opsDigestReceipt struct {
	ID          string `json:"id,omitempty"`
	EventTime   string `json:"event_time,omitempty"`
	Summary     string `json:"summary,omitempty"`
	SourceHint  string `json:"source_hint,omitempty"`
	Pointer     string `json:"pointer,omitempty"`
	AccountHash string `json:"account_hash,omitempty"`
}

// opsDigestDecisionStub is a human-owned scaffold (not auto-apply).
// Lean host leaves this empty — do not invent hypotheses.
type opsDigestDecisionStub struct {
	Pattern                string   `json:"pattern,omitempty"`
	ReceiptsRef            []string `json:"receipts_ref,omitempty"`
	ProductOrGTMHypothesis string   `json:"product_or_gtm_hypothesis,omitempty"`
}

// opsDigestPattern is unused on the lean host (insufficient-signal is OK).
// Kept so the JSON key `patterns` is always an array, never invented GA.
type opsDigestPattern struct {
	ID      string `json:"id,omitempty"`
	Kind    string `json:"kind,omitempty"`
	Subject string `json:"subject,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// opsDigestExportOutput matches TUI MemoryOpsDigestResult JSON.
type opsDigestExportOutput struct {
	Window       string                `json:"window"`
	Horizon      string                `json:"horizon"`
	AsOf         string                `json:"as_of"`
	Since        string                `json:"since,omitempty"`
	Honesty      opsDigestHonesty      `json:"honesty"`
	Patterns     []opsDigestPattern    `json:"patterns"`
	Receipts     []opsDigestReceipt    `json:"receipts"`
	DecisionStub opsDigestDecisionStub `json:"decision_stub"`
}

func (h *Host) handleOpsDigestExport(_ context.Context, _ *mcp.CallToolRequest, in opsDigestExportInput) (*mcp.CallToolResult, opsDigestExportOutput, error) {
	window := strings.ToLower(strings.TrimSpace(in.Window))
	if window == "" {
		window = opsDigestWindowDay
	}
	if window != opsDigestWindowDay && window != opsDigestWindowWeek {
		err := fmt.Errorf("window must be day or week")
		return toolError(err), opsDigestExportOutput{}, err
	}
	horizon := strings.ToLower(strings.TrimSpace(in.Horizon))
	if horizon == "" {
		horizon = "ops"
	}
	switch horizon {
	case "ops", "knowledge", "analytical", "all":
	default:
		err := fmt.Errorf("horizon must be ops|knowledge|analytical|all")
		return toolError(err), opsDigestExportOutput{}, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = opsDigestLimitDef
	}
	if limit > opsDigestLimitMax {
		limit = opsDigestLimitMax
	}

	asOf, err := parseTimeOrNow(in.AsOf)
	if err != nil {
		return toolError(err), opsDigestExportOutput{}, err
	}
	asOf = asOf.UTC()
	var since time.Time
	if window == opsDigestWindowWeek {
		since = asOf.Add(-7 * 24 * time.Hour)
	} else {
		since = asOf.Add(-24 * time.Hour)
	}

	_, ps, err := h.resolveStore(in.Tenant)
	if err != nil {
		return toolError(err), opsDigestExportOutput{}, err
	}

	opts := palace.ListMemoryOptions{
		Limit:     limit,
		TimeFrom:  &since,
		TimeTo:    &asOf,
		Ascending: false,
	}
	entries := ps.ListMemoryWithOptions(opts)

	receipts := make([]opsDigestReceipt, 0, len(entries))
	for _, e := range entries {
		id := strings.TrimSpace(e.ID)
		if id == "" {
			continue
		}
		hit := hitFromEntry(e)
		summary := strings.TrimSpace(hit.Summary)
		if summary == "" {
			summary = strings.TrimSpace(hit.Full)
		}
		if len(summary) > 240 {
			summary = summary[:240] + "…"
		}
		receipts = append(receipts, opsDigestReceipt{
			ID:         id,
			EventTime:  hit.Timestamp,
			Summary:    summary,
			SourceHint: digestSourceHint(e),
			Pointer:    id,
		})
	}

	out := opsDigestExportOutput{
		Window:       window,
		Horizon:      horizon,
		AsOf:         asOf.Format(time.RFC3339),
		Since:        since.Format(time.RFC3339),
		Honesty:      leanOpsDigestHonesty(horizon),
		Patterns:     []opsDigestPattern{},
		Receipts:     receipts,
		DecisionStub: opsDigestDecisionStub{},
	}
	return toolJSON(out), out, nil
}

func leanOpsDigestHonesty(horizon string) opsDigestHonesty {
	h := opsDigestHonesty{
		OpsPulse:         "ga_path",
		Knowledge:        "beta",
		Analytical:       "beta",
		NeverInventGA:    true,
		DualWriteDefault: "off",
		BookDemo:         "off",
		Note:             "Local palace listing only (memory_list window). Patterns empty — insufficient-signal OK; do not invent GA or mesh. dual_write OFF · not Memory GA · catalog ≠ connected. No mesh bind required.",
	}
	if horizon == "knowledge" || horizon == "analytical" {
		h.Note = "Horizon " + horizon + " is Beta. Local palace listing only; patterns empty — insufficient-signal OK. dual_write OFF · not Memory GA · catalog ≠ connected."
	}
	return h
}

// digestSourceHint returns a TUI-classifiable hint.
// Local palace entries default to palace_timeline (private).
// mesh* is used only when the entry itself is mesh-sourced — never invented.
func digestSourceHint(e palace.MemoryEntry) string {
	tags := make([]string, 0, 2+len(e.Content.Tags))
	if hint := strings.TrimSpace(e.Provenance.SourceHint); hint != "" {
		tags = append(tags, hint)
	}
	tags = append(tags, e.Content.Tags...)
	if step := strings.TrimSpace(e.Provenance.SourceStep); step != "" {
		tags = append(tags, step)
	}
	for _, raw := range tags {
		if hint := classifyDigestTag(raw); hint == "mesh" {
			return "mesh"
		}
	}
	for _, raw := range tags {
		if hint := classifyDigestTag(raw); hint == "private" {
			return palaceTimelineHint
		}
	}
	return palaceTimelineHint
}

// classifyDigestTag maps one tag/provenance token the same way TUI
// ClassifyDigestSourceHint treats source_hint (mesh vs private only).
// Catalog/grant/external are not emitted by the lean host.
func classifyDigestTag(raw string) string {
	h := strings.ToLower(strings.TrimSpace(raw))
	h = strings.ReplaceAll(h, "-", "_")
	if i := strings.IndexByte(h, ':'); i >= 0 {
		// source:mesh / source:private / role:user — use the value side too.
		head, tail := h[:i], h[i+1:]
		if head == "source" || head == "origin" || head == "source_hint" {
			h = tail
		}
	}
	if h == "" {
		return ""
	}
	switch h {
	case "mesh", "mesh_stream", "mesh_consume", "mesh_pulse", "mesh_incident",
		"mesh_incidents", "broker", "stream", "consume", "ops_pulse":
		return "mesh"
	case "private", "private_overlay", "private_rca", "palace", "palace_timeline",
		"local", "local_palace", "rca", "overlay":
		return "private"
	}
	if strings.HasPrefix(h, "mesh_") {
		return "mesh"
	}
	if strings.HasPrefix(h, "private_") || strings.HasPrefix(h, "palace_") {
		return "private"
	}
	return ""
}
