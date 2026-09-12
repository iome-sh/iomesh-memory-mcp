package mcphost

import (
	"context"
	"fmt"
	"sort"
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
	// opsDigestScanCap is how many in-window entries we list before selecting
	// receipts. Kernel ListMemoryWithOptions defaults Limit<=0 to 50 and sorts
	// newest event_time first, so a receipt Limit of 20 can drop older mesh
	// turns when newer private RCA fills the window (#66). Scan wider, then
	// prefer source-class diversity. Never invent mesh.
	opsDigestScanCap = 200
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

// opsDigestReceiptProvenance is optional palace provenance on a receipt.
// TUI ClassifyDigestReceipt reads source_hint here when the export-origin
// receipt.source_hint is palace_timeline. Never invent mesh.
type opsDigestReceiptProvenance struct {
	SourceHint string `json:"source_hint,omitempty"`
	SourceStep string `json:"source_step,omitempty"`
}

// opsDigestReceipt is one local palace timeline receipt.
// Field names match TUI MemoryOpsDigestReceipt (including tags / provenance).
type opsDigestReceipt struct {
	ID          string                     `json:"id,omitempty"`
	EventTime   string                     `json:"event_time,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	SourceHint  string                     `json:"source_hint,omitempty"`
	Pointer     string                     `json:"pointer,omitempty"`
	AccountHash string                     `json:"account_hash,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
	Provenance  opsDigestReceiptProvenance `json:"provenance,omitempty"`
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

// opsDigestReceiptSelection documents how the receipt set was chosen.
// TUI MemoryOpsDigestResult ignores unknown JSON keys; since/as_of remain
// the printed window bounds. Do not treat this as cite-both / Connected.
type opsDigestReceiptSelection struct {
	Mode              string `json:"mode"`
	Limit             int    `json:"limit"`
	Scanned           int    `json:"scanned"`
	ScanCap           int    `json:"scan_cap"`
	ScanTruncated     bool   `json:"scan_truncated"`
	MeshInWindow      bool   `json:"mesh_in_window"`
	PrivateInWindow   bool   `json:"private_in_window"`
	MeshInReceipts    bool   `json:"mesh_in_receipts"`
	PrivateInReceipts bool   `json:"private_in_receipts"`
	Note              string `json:"note,omitempty"`
}

// opsDigestExportOutput matches TUI MemoryOpsDigestResult JSON.
type opsDigestExportOutput struct {
	Window           string                    `json:"window"`
	Horizon          string                    `json:"horizon"`
	AsOf             string                    `json:"as_of"`
	Since            string                    `json:"since,omitempty"`
	Honesty          opsDigestHonesty          `json:"honesty"`
	Patterns         []opsDigestPattern        `json:"patterns"`
	Receipts         []opsDigestReceipt        `json:"receipts"`
	DecisionStub     opsDigestDecisionStub     `json:"decision_stub"`
	ReceiptSelection opsDigestReceiptSelection `json:"receipt_selection"`
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

	windowOpts := palace.ListMemoryOptions{
		TimeFrom:  &since,
		TimeTo:    &asOf,
		Ascending: false,
	}
	scanOpts := windowOpts
	scanOpts.Limit = opsDigestScanCap
	scanned := ps.ListMemoryWithOptions(scanOpts)

	// Supplemental mesh-tag list so older source_hint:mesh turns survive when
	// newer private RCA fills the newest-N scan (#66). Do not invent mesh.
	meshOpts := windowOpts
	meshOpts.Limit = opsDigestScanCap
	if tag := palace.FormatSourceHintTag("mesh"); tag != "" {
		meshOpts.Tag = tag
	} else {
		meshOpts.Tag = "source_hint:mesh"
	}
	meshTagged := ps.ListMemoryWithOptions(meshOpts)

	merged := mergeMemoryEntries(scanned, meshTagged)
	receipts, sel := selectDigestReceipts(merged, limit)
	sel.Scanned = len(scanned)
	sel.ScanCap = opsDigestScanCap
	sel.ScanTruncated = len(scanned) >= opsDigestScanCap
	sel.Note = digestReceiptSelectionNote(sel)

	out := opsDigestExportOutput{
		Window:           window,
		Horizon:          horizon,
		AsOf:             asOf.Format(time.RFC3339),
		Since:            since.Format(time.RFC3339),
		Honesty:          leanOpsDigestHonesty(horizon),
		Patterns:         []opsDigestPattern{},
		Receipts:         receipts,
		DecisionStub:     opsDigestDecisionStub{},
		ReceiptSelection: sel,
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
		Note:             "Local palace listing only (memory_list window). Receipt selection prefers source-class diversity when mesh and private both exist in-window — never invents mesh. Patterns empty — insufficient-signal OK. No mesh bind required.",
	}
	if horizon == "knowledge" || horizon == "analytical" {
		h.Note = "Horizon " + horizon + " is Beta. Local palace listing only; receipt selection prefers source-class diversity when both classes exist in-window — never invents mesh. Patterns empty — insufficient-signal OK."
	}
	return h
}

// digestSourceHint returns a TUI-classifiable hint.
// Local palace entries default to palace_timeline (private).
// mesh* is used only when the entry itself is mesh-sourced — never invented.
func digestSourceHint(e palace.MemoryEntry) string {
	cands := digestHintCandidates(e)
	for _, raw := range cands {
		if hint := classifyDigestTag(raw); hint == "mesh" {
			return "mesh"
		}
	}
	for _, raw := range cands {
		if hint := classifyDigestTag(raw); hint == "private" {
			return palaceTimelineHint
		}
	}
	return palaceTimelineHint
}

// digestHintCandidates is the palace class wire TUI ClassifyDigestReceipt
// reads: provenance.source_hint, tags (Content + Temporal), source_step.
func digestHintCandidates(e palace.MemoryEntry) []string {
	out := make([]string, 0, 2+len(e.Content.Tags)+len(e.TemporalTags))
	if hint := strings.TrimSpace(e.Provenance.SourceHint); hint != "" {
		out = append(out, hint)
	}
	out = append(out, e.Content.Tags...)
	out = append(out, e.TemporalTags...)
	if step := strings.TrimSpace(e.Provenance.SourceStep); step != "" {
		out = append(out, step)
	}
	return out
}

// digestReceiptTags copies palace tags TUI can classify (never invented).
func digestReceiptTags(e palace.MemoryEntry) []string {
	seen := make(map[string]struct{}, len(e.Content.Tags)+len(e.TemporalTags))
	out := make([]string, 0, len(e.Content.Tags)+len(e.TemporalTags))
	for _, raw := range append(append([]string{}, e.Content.Tags...), e.TemporalTags...) {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func digestReceiptProvenance(e palace.MemoryEntry) opsDigestReceiptProvenance {
	return opsDigestReceiptProvenance{
		SourceHint: strings.TrimSpace(e.Provenance.SourceHint),
		SourceStep: strings.TrimSpace(e.Provenance.SourceStep),
	}
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

func mergeMemoryEntries(groups ...[]palace.MemoryEntry) []palace.MemoryEntry {
	seen := make(map[string]struct{})
	out := make([]palace.MemoryEntry, 0)
	for _, g := range groups {
		for _, e := range g {
			id := strings.TrimSpace(e.ID)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}

type digestCandidate struct {
	receipt opsDigestReceipt
	class   string // mesh | private
	t       time.Time
}

func digestEntryTime(e palace.MemoryEntry) time.Time {
	ts := e.Timestamp
	if ts.IsZero() {
		ts = e.CreatedAt
	}
	return ts.UTC()
}

func digestCandidateFromEntry(e palace.MemoryEntry) (digestCandidate, bool) {
	id := strings.TrimSpace(e.ID)
	if id == "" {
		return digestCandidate{}, false
	}
	hit := hitFromEntry(e)
	summary := strings.TrimSpace(hit.Summary)
	if summary == "" {
		summary = strings.TrimSpace(hit.Full)
	}
	if len(summary) > 240 {
		summary = summary[:240] + "…"
	}
	hint := digestSourceHint(e)
	class := classifyDigestTag(hint)
	if class == "" {
		class = "private"
	}
	return digestCandidate{
		receipt: opsDigestReceipt{
			ID:         id,
			EventTime:  hit.Timestamp,
			Summary:    summary,
			SourceHint: hint,
			Pointer:    id,
			Tags:       digestReceiptTags(e),
			Provenance: digestReceiptProvenance(e),
		},
		class: class,
		t:     digestEntryTime(e),
	}, true
}

func sortDigestCandidatesNewest(cands []digestCandidate) {
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].t.Equal(cands[j].t) {
			return cands[i].receipt.ID > cands[j].receipt.ID
		}
		return cands[i].t.After(cands[j].t)
	})
}

// selectDigestReceipts picks up to limit receipts. When both mesh and private
// exist in the candidate set and limit >= 2, newest mesh and newest private
// are reserved first; remaining slots fill newest-event_time. Never invents
// a class that is not present.
func selectDigestReceipts(entries []palace.MemoryEntry, limit int) ([]opsDigestReceipt, opsDigestReceiptSelection) {
	if limit <= 0 {
		limit = opsDigestLimitDef
	}
	cands := make([]digestCandidate, 0, len(entries))
	for _, e := range entries {
		c, ok := digestCandidateFromEntry(e)
		if ok {
			cands = append(cands, c)
		}
	}
	sortDigestCandidatesNewest(cands)

	var mesh, priv []digestCandidate
	for _, c := range cands {
		if c.class == "mesh" {
			mesh = append(mesh, c)
		} else {
			priv = append(priv, c)
		}
	}

	picked := make([]digestCandidate, 0, limit)
	used := make(map[string]struct{}, limit)
	pick := func(c digestCandidate) {
		if len(picked) >= limit {
			return
		}
		id := c.receipt.ID
		if _, ok := used[id]; ok {
			return
		}
		used[id] = struct{}{}
		picked = append(picked, c)
	}
	if limit >= 2 && len(mesh) > 0 && len(priv) > 0 {
		pick(mesh[0])
		pick(priv[0])
	}
	for _, c := range cands {
		if len(picked) >= limit {
			break
		}
		pick(c)
	}
	sortDigestCandidatesNewest(picked)

	receipts := make([]opsDigestReceipt, 0, len(picked))
	meshInReceipts := false
	privInReceipts := false
	for _, c := range picked {
		receipts = append(receipts, c.receipt)
		if c.class == "mesh" {
			meshInReceipts = true
		} else {
			privInReceipts = true
		}
	}

	sel := opsDigestReceiptSelection{
		Mode:              "source_class_diversity",
		Limit:             limit,
		MeshInWindow:      len(mesh) > 0,
		PrivateInWindow:   len(priv) > 0,
		MeshInReceipts:    meshInReceipts,
		PrivateInReceipts: privInReceipts,
	}
	return receipts, sel
}

func digestReceiptSelectionNote(sel opsDigestReceiptSelection) string {
	switch {
	case sel.MeshInWindow && sel.PrivateInWindow && sel.MeshInReceipts && sel.PrivateInReceipts:
		return "In-window mesh and private both represented. Window bounds are since/as_of. Never invents mesh."
	case sel.MeshInWindow && !sel.MeshInReceipts:
		return "Mesh exists in the scanned window but is not in this receipt set (limit too small or scan truncated). Window bounds are since/as_of. Never invents mesh."
	case sel.ScanTruncated && !sel.MeshInWindow:
		return "Newest-event_time scan hit cap; older in-window entries may exist. Window bounds are since/as_of. Never invents mesh."
	default:
		return "Receipts from in-window palace listing. Window bounds are since/as_of. Never invents mesh."
	}
}
