package mcphost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	palace "github.com/iome-sh/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestOpsDigestExportRegistered(t *testing.T) {
	names := LeanToolNames()
	found := false
	for _, n := range names {
		if n == "ops_digest_export" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("leanToolNames missing ops_digest_export: %v", names)
	}
	if len(names) < 11 {
		t.Fatalf("lean tools=%d want >= 11 (ops_digest_export + memory_extract_facts): %v", len(names), names)
	}

	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if sdk := h.NewSDKServer(); sdk == nil {
		t.Fatal("nil sdk server")
	}
	snap := HealthzSnapshot(h)
	have := false
	for _, n := range snap.ToolNames {
		if n == "ops_digest_export" {
			have = true
			break
		}
	}
	if !have {
		t.Fatalf("healthz/preflight tool_names missing ops_digest_export: %v", snap.ToolNames)
	}
	if snap.Tools != len(snap.ToolNames) || snap.Tools < 11 {
		t.Fatalf("healthz tools=%d names=%d", snap.Tools, len(snap.ToolNames))
	}
}

func TestOpsDigestExportEmptyPalaceHonesty(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	res, out, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood"})
	if err != nil {
		t.Fatalf("empty palace digest: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("empty palace must be a success result, got %+v", res)
	}
	if out.Window != "day" {
		t.Fatalf("default window: %q", out.Window)
	}
	if out.Horizon != "ops" {
		t.Fatalf("default horizon: %q", out.Horizon)
	}
	if out.AsOf == "" || out.Since == "" {
		t.Fatalf("as_of/since required: as_of=%q since=%q", out.AsOf, out.Since)
	}
	if out.Patterns == nil {
		t.Fatal("patterns must be [] not null")
	}
	if len(out.Patterns) != 0 {
		t.Fatalf("empty palace must not invent patterns: %+v", out.Patterns)
	}
	if out.Receipts == nil {
		t.Fatal("receipts must be [] not null")
	}
	if len(out.Receipts) != 0 {
		t.Fatalf("empty palace must not invent receipts: %+v", out.Receipts)
	}
	assertOpsDigestHonesty(t, out.Honesty)
	if out.Honesty.DualWriteDefault != "off" {
		t.Fatalf("dual_write_default: %q", out.Honesty.DualWriteDefault)
	}
	if out.DecisionStub.Pattern != "" || out.DecisionStub.ProductOrGTMHypothesis != "" {
		t.Fatalf("decision_stub must stay empty: %+v", out.DecisionStub)
	}

	raw := toolJSON(out)
	if raw == nil || raw.IsError || len(raw.Content) == 0 {
		t.Fatalf("tool JSON: %+v", raw)
	}
	tc, ok := raw.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type %T", raw.Content[0])
	}
	var decoded opsDigestExportOutput
	if err := json.Unmarshal([]byte(tc.Text), &decoded); err != nil {
		t.Fatalf("round-trip json: %v body=%s", err, tc.Text)
	}
	if decoded.Honesty.DualWriteDefault != "off" || !decoded.Honesty.NeverInventGA {
		t.Fatalf("marshaled honesty: %+v", decoded.Honesty)
	}
	if decoded.Patterns == nil || decoded.Receipts == nil {
		t.Fatalf("marshaled slices must be arrays: patterns=%v receipts=%v", decoded.Patterns, decoded.Receipts)
	}
	if decoded.ReceiptSelection.MeshInWindow || decoded.ReceiptSelection.MeshInReceipts {
		t.Fatalf("empty palace must not invent mesh: %+v", decoded.ReceiptSelection)
	}
	if decoded.ReceiptSelection.Mode != "source_class_diversity" {
		t.Fatalf("receipt_selection.mode: %q", decoded.ReceiptSelection.Mode)
	}
}

func TestOpsDigestExportPrivateTaggedReceipt(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const summary = "private palace RCA note for digest cite"
	_, wrote, err := h.handleWrite(ctx, nil, writeInput{
		Tenant:  "dogfood",
		Summary: summary,
		Full:    summary,
		Tags:    []string{"private", "source:local_palace"},
	})
	if err != nil {
		t.Fatalf("write private: %v", err)
	}
	if wrote.DualWrite != "off" {
		t.Fatalf("write honesty: %+v", wrote)
	}

	_, out, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{
		Tenant:  "dogfood",
		Window:  "day",
		Horizon: "ops",
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	assertOpsDigestHonesty(t, out.Honesty)
	if len(out.Patterns) != 0 {
		t.Fatalf("must not invent patterns: %+v", out.Patterns)
	}
	if len(out.Receipts) == 0 {
		t.Fatal("expected private-tagged palace receipt")
	}
	found := false
	for _, r := range out.Receipts {
		if r.ID != wrote.MemoryID && !strings.Contains(r.Summary, "private palace RCA") {
			continue
		}
		found = true
		if classifyTestSourceHint(r.SourceHint) != "private" {
			t.Fatalf("source_hint %q not classifiable as private (TUI private/palace/palace_timeline): %+v", r.SourceHint, r)
		}
		if strings.HasPrefix(strings.ToLower(r.SourceHint), "mesh") {
			t.Fatalf("must not invent mesh source_hint: %+v", r)
		}
		if r.EventTime == "" {
			t.Fatalf("receipt event_time required: %+v", r)
		}
		if r.Pointer == "" {
			t.Fatalf("receipt pointer required: %+v", r)
		}
	}
	if !found {
		t.Fatalf("private-tagged entry missing from receipts: %+v", out.Receipts)
	}
}

func TestOpsDigestExportHonorsMeshSourceHint(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	_, wrote, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:     "dogfood",
		SessionID:  "dept.engineering.events.github",
		Role:       "user",
		Content:    "durable mesh consume for digest cite",
		SourceHint: "mesh",
	})
	if err != nil {
		t.Fatalf("ingest mesh: %v", err)
	}
	if wrote.DualWrite != "off" || wrote.Audited {
		t.Fatalf("dual_write must be off; do not invent Connected/Memory GA: %+v", wrote)
	}

	_, out, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood"})
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	assertOpsDigestHonesty(t, out.Honesty)
	if len(out.Patterns) != 0 {
		t.Fatalf("must not invent patterns: %+v", out.Patterns)
	}
	found := false
	for _, r := range out.Receipts {
		if r.ID != wrote.MemoryID && !strings.Contains(r.Summary, "durable mesh consume") {
			continue
		}
		found = true
		assertReceiptCarriesMeshClass(t, r)
	}
	if !found {
		t.Fatalf("mesh-hinted entry missing from receipts: %+v", out.Receipts)
	}

	raw := toolJSON(out)
	if raw == nil || raw.IsError || len(raw.Content) == 0 {
		t.Fatalf("tool JSON: %+v", raw)
	}
	tc, ok := raw.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type %T", raw.Content[0])
	}
	if !strings.Contains(tc.Text, `"provenance"`) {
		t.Fatalf("marshaled receipt must include provenance object: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, `"source_hint": "mesh"`) {
		t.Fatalf("marshaled receipt source_hint must be mesh when palace is mesh-sourced: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, "source_hint:mesh") {
		t.Fatalf("marshaled receipt must include tag source_hint:mesh: %s", tc.Text)
	}
	var decoded opsDigestExportOutput
	if err := json.Unmarshal([]byte(tc.Text), &decoded); err != nil {
		t.Fatalf("round-trip json: %v body=%s", err, tc.Text)
	}
	wired := false
	for _, r := range decoded.Receipts {
		if r.ID != wrote.MemoryID && !strings.Contains(r.Summary, "durable mesh consume") {
			continue
		}
		wired = true
		assertReceiptCarriesMeshClass(t, r)
	}
	if !wired {
		t.Fatalf("round-trip missed mesh receipt: %+v", decoded.Receipts)
	}
}

func TestOpsDigestExportMeshAndPrivateWhenBothExist(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	asOf := time.Now().UTC().Add(time.Second)
	// Older mesh event_time (GitHub created ~morning) vs newer private RCA.
	meshAt := asOf.Add(-10 * time.Hour).Format(time.RFC3339)
	_, meshWrote, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:     "dogfood",
		SessionID:  "dept.engineering.events.github",
		Role:       "user",
		Content:    "durable mesh consume morning pull for digest diversity",
		EventTime:  meshAt,
		SourceHint: "mesh",
	})
	if err != nil {
		t.Fatalf("ingest mesh: %v", err)
	}
	if meshWrote.DualWrite != "off" || meshWrote.Audited {
		t.Fatalf("dual_write must be off; do not invent Connected/Memory GA: %+v", meshWrote)
	}

	const privates = 25
	for i := 0; i < privates; i++ {
		_, wrote, err := h.handleWrite(ctx, nil, writeInput{
			Tenant:  "dogfood",
			Summary: fmt.Sprintf("private RCA note %02d newer than mesh pull", i),
			Tags:    []string{"private", "private_rca"},
		})
		if err != nil {
			t.Fatalf("write private %d: %v", i, err)
		}
		if wrote.DualWrite != "off" {
			t.Fatalf("write honesty: %+v", wrote)
		}
	}

	_, out, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{
		Tenant: "dogfood",
		Window: "day",
		AsOf:   asOf.Format(time.RFC3339),
		Limit:  5, // TUI sticky default is 20; small limit must still keep both classes
	})
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	assertOpsDigestHonesty(t, out.Honesty)
	if len(out.Patterns) != 0 {
		t.Fatalf("must not invent patterns: %+v", out.Patterns)
	}
	if len(out.Receipts) == 0 || len(out.Receipts) > 5 {
		t.Fatalf("receipt count: %d want 1..5", len(out.Receipts))
	}

	haveMesh, havePrivate := false, false
	foundMeshID := false
	for _, r := range out.Receipts {
		switch classifyTestReceipt(r) {
		case "mesh":
			haveMesh = true
			if r.ID == meshWrote.MemoryID || strings.Contains(r.Summary, "durable mesh consume") {
				foundMeshID = true
				assertReceiptCarriesMeshClass(t, r)
			}
		case "private":
			havePrivate = true
		default:
			t.Fatalf("unclassifiable receipt (source_hint/provenance/tags): %+v", r)
		}
	}
	if !haveMesh || !havePrivate {
		t.Fatalf("default receipt set must include mesh+private when both exist: mesh=%v private=%v receipts=%+v sel=%+v",
			haveMesh, havePrivate, out.Receipts, out.ReceiptSelection)
	}
	if !foundMeshID {
		t.Fatalf("mesh-stamped turn missing from receipts: %+v", out.Receipts)
	}
	if !out.ReceiptSelection.MeshInWindow || !out.ReceiptSelection.PrivateInWindow {
		t.Fatalf("window class flags: %+v", out.ReceiptSelection)
	}
	if !out.ReceiptSelection.MeshInReceipts || !out.ReceiptSelection.PrivateInReceipts {
		t.Fatalf("receipt class flags: %+v", out.ReceiptSelection)
	}
	if out.Since == "" || out.AsOf == "" {
		t.Fatalf("honest window bounds required: since=%q as_of=%q", out.Since, out.AsOf)
	}
	if strings.Contains(strings.ToLower(out.Honesty.Note), "cite-both") {
		t.Fatalf("must not invent cite-both: %q", out.Honesty.Note)
	}

	// Sticky TUI default is 20 — mesh-visible fields must still survive.
	_, sticky, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{
		Tenant: "dogfood",
		Window: "day",
		AsOf:   asOf.Format(time.RFC3339),
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("sticky digest: %v", err)
	}
	stickyMesh, stickyPriv := false, false
	for _, r := range sticky.Receipts {
		switch classifyTestReceipt(r) {
		case "mesh":
			stickyMesh = true
			assertReceiptCarriesMeshClass(t, r)
		case "private":
			stickyPriv = true
		}
	}
	if !stickyMesh || !stickyPriv {
		t.Fatalf("sticky limit=20 must keep mesh+private with wired fields: mesh=%v private=%v receipts=%+v",
			stickyMesh, stickyPriv, sticky.Receipts)
	}
}

func TestSelectDigestReceiptsPrefersSourceClassDiversity(t *testing.T) {
	now := time.Date(2026, 9, 10, 16, 10, 0, 0, time.UTC)
	meshAt := time.Date(2026, 9, 10, 6, 46, 0, 0, time.UTC)
	entries := make([]palace.MemoryEntry, 0, 13)
	for i := 0; i < 12; i++ {
		entries = append(entries, palace.MemoryEntry{
			ID:        fmt.Sprintf("priv-%02d", i),
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
			Content: palace.MemoryContent{
				Summary: "private RCA " + fmt.Sprintf("%02d", i),
				Tags:    []string{"source_hint:private", "private_rca"},
			},
			Provenance: palace.MemoryProvenance{SourceHint: "private"},
		})
	}
	entries = append(entries, palace.MemoryEntry{
		ID:        "mesh-1",
		Timestamp: meshAt,
		Content: palace.MemoryContent{
			Summary: "mesh pull morning",
			Tags:    []string{"source_hint:mesh"},
		},
		Provenance: palace.MemoryProvenance{SourceHint: "mesh"},
	})

	got, sel := selectDigestReceipts(entries, 5)
	if len(got) != 5 {
		t.Fatalf("len=%d want 5: %+v", len(got), got)
	}
	haveMesh, havePrivate := false, false
	for _, r := range got {
		switch classifyTestReceipt(r) {
		case "mesh":
			haveMesh = true
			if r.ID != "mesh-1" {
				t.Fatalf("unexpected mesh id: %+v", r)
			}
			assertReceiptCarriesMeshClass(t, r)
		case "private":
			havePrivate = true
		default:
			t.Fatalf("unclassifiable: %+v", r)
		}
	}
	if !haveMesh || !havePrivate {
		t.Fatalf("diversity failed: mesh=%v private=%v receipts=%+v", haveMesh, havePrivate, got)
	}
	if !sel.MeshInWindow || !sel.PrivateInWindow || !sel.MeshInReceipts || !sel.PrivateInReceipts {
		t.Fatalf("selection flags: %+v", sel)
	}

	privOnly, privSel := selectDigestReceipts(entries[:12], 5)
	for _, r := range privOnly {
		if classifyTestSourceHint(r.SourceHint) == "mesh" {
			t.Fatalf("must not invent mesh: %+v", r)
		}
	}
	if privSel.MeshInWindow || privSel.MeshInReceipts {
		t.Fatalf("private-only must not claim mesh: %+v", privSel)
	}

	meshOnly, meshSel := selectDigestReceipts(entries[12:], 5)
	if len(meshOnly) != 1 || classifyTestSourceHint(meshOnly[0].SourceHint) != "mesh" {
		t.Fatalf("mesh-only: %+v", meshOnly)
	}
	if meshSel.PrivateInWindow || meshSel.PrivateInReceipts {
		t.Fatalf("mesh-only must not invent private: %+v", meshSel)
	}
}

func TestOpsDigestExportNeverInventsMesh(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:    "dogfood",
		SessionID: "sess-digest",
		Role:      "user",
		Content:   "local ingest for digest mesh-honesty check",
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	_, out, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood"})
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if len(out.Receipts) == 0 {
		t.Fatal("expected local palace receipts")
	}
	for _, r := range out.Receipts {
		if classifyTestReceipt(r) != "private" {
			t.Fatalf("local palace receipt must stay private class: %+v", r)
		}
		if classifyTestSourceHint(r.SourceHint) == "mesh" || classifyDigestTag(r.Provenance.SourceHint) == "mesh" {
			t.Fatalf("must not invent mesh on receipt wire: %+v", r)
		}
		for _, tag := range r.Tags {
			if classifyDigestTag(tag) == "mesh" {
				t.Fatalf("must not invent mesh tag: %+v", r)
			}
		}
	}
	if out.ReceiptSelection.MeshInWindow || out.ReceiptSelection.MeshInReceipts {
		t.Fatalf("must not invent mesh in receipt_selection: %+v", out.ReceiptSelection)
	}
}

func TestOpsDigestExportWindowAndValidation(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	old := time.Now().UTC().Add(-8 * 24 * time.Hour).Format(time.RFC3339)

	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:    "dogfood",
		SessionID: "sess-old",
		Role:      "user",
		Content:   "old palace note outside day window",
		EventTime: old,
	}); err != nil {
		t.Fatalf("ingest old: %v", err)
	}
	_, wrote, err := h.handleWrite(ctx, nil, writeInput{
		Tenant:  "dogfood",
		Summary: "recent private note inside day window",
		Tags:    []string{"private"},
	})
	if err != nil {
		t.Fatalf("write recent: %v", err)
	}
	asOf := time.Now().UTC().Add(time.Second)

	_, day, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{
		Tenant: "dogfood",
		Window: "day",
		AsOf:   asOf.Format(time.RFC3339),
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("day digest: %v", err)
	}
	for _, r := range day.Receipts {
		if strings.Contains(r.Summary, "outside day window") {
			t.Fatalf("day window leaked old entry: %+v", r)
		}
	}
	foundRecent := false
	for _, r := range day.Receipts {
		if r.ID == wrote.MemoryID {
			foundRecent = true
			if classifyTestSourceHint(r.SourceHint) != "private" {
				t.Fatalf("recent private hint: %+v", r)
			}
		}
	}
	if !foundRecent {
		t.Fatalf("day window missed recent write; receipts=%+v", day.Receipts)
	}

	_, week, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{
		Tenant: "dogfood",
		Window: "week",
		AsOf:   asOf.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("week digest: %v", err)
	}
	if week.Window != "week" {
		t.Fatalf("week window echo: %q", week.Window)
	}

	if _, _, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood", Window: "month"}); err == nil {
		t.Fatal("invalid window must error")
	}
	if _, _, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood", Horizon: "ga"}); err == nil {
		t.Fatal("invalid horizon must error")
	}
	if _, _, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood", AsOf: "last-week"}); err == nil {
		t.Fatal("invalid as_of must error")
	}
	res, _, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{})
	if err == nil || !errors.Is(err, ErrTenantRequired) {
		t.Fatalf("omit tenant: err=%v want ErrTenantRequired", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("omit tenant want IsError, got %+v", res)
	}

	_, know, err := h.handleOpsDigestExport(ctx, nil, opsDigestExportInput{Tenant: "dogfood", Horizon: "knowledge"})
	if err != nil {
		t.Fatalf("knowledge: %v", err)
	}
	if know.Horizon != "knowledge" || !strings.Contains(know.Honesty.Note, "Beta") {
		t.Fatalf("knowledge honesty: %+v", know.Honesty)
	}
	if know.Honesty.DualWriteDefault != "off" {
		t.Fatalf("knowledge dual_write: %q", know.Honesty.DualWriteDefault)
	}
}

func assertOpsDigestHonesty(t *testing.T, h opsDigestHonesty) {
	t.Helper()
	if h.DualWriteDefault != "off" {
		t.Fatalf("dual_write_default: %q", h.DualWriteDefault)
	}
	if !h.NeverInventGA {
		t.Fatal("never_invent_ga must be true")
	}
	if h.BookDemo != "off" {
		t.Fatalf("book_demo: %q", h.BookDemo)
	}
	if h.Knowledge != "beta" || h.Analytical != "beta" {
		t.Fatalf("knowledge/analytical must stay beta: %+v", h)
	}
	if h.OpsPulse == "" {
		t.Fatal("ops_pulse required")
	}
	note := strings.ToLower(h.Note)
	if strings.Contains(note, "memory ga") && !strings.Contains(note, "not memory ga") {
		t.Fatalf("must not invent Memory GA: %q", h.Note)
	}
}

func TestOpsDigestReceiptWiresPalaceMeshWithoutInventing(t *testing.T) {
	now := time.Date(2026, 9, 10, 16, 10, 0, 0, time.UTC)
	meshAt := time.Date(2026, 9, 10, 6, 46, 0, 0, time.UTC)

	// Live palace shape: provenance.source_hint=mesh + tag source_hint:mesh,
	// no raw "mesh" tag. Receipt must surface class on source_hint AND wire.
	palaceMesh := palace.MemoryEntry{
		ID:        "mesh-palace",
		Timestamp: meshAt,
		Content: palace.MemoryContent{
			Summary: "dept pull morning",
			Tags:    []string{"source_hint:mesh", "role:user"},
		},
		Provenance: palace.MemoryProvenance{
			SourceHint: "mesh",
			SourceStep: "mcp_memory_ingest_turn",
		},
	}
	priv := palace.MemoryEntry{
		ID:        "priv-palace",
		Timestamp: now,
		Content: palace.MemoryContent{
			Summary: "private RCA",
			Tags:    []string{"source_hint:private", "private_rca"},
		},
		Provenance: palace.MemoryProvenance{SourceHint: "private"},
	}
	got, sel := selectDigestReceipts([]palace.MemoryEntry{priv, palaceMesh}, 20)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(got), got)
	}
	if !sel.MeshInWindow || !sel.PrivateInWindow || !sel.MeshInReceipts || !sel.PrivateInReceipts {
		t.Fatalf("diversity flags: %+v", sel)
	}
	var meshR, privR *opsDigestReceipt
	for i := range got {
		switch classifyTestReceipt(got[i]) {
		case "mesh":
			meshR = &got[i]
		case "private":
			privR = &got[i]
		default:
			t.Fatalf("unclassifiable: %+v", got[i])
		}
	}
	if meshR == nil || privR == nil {
		t.Fatalf("sticky default must keep both classes: %+v", got)
	}
	assertReceiptCarriesMeshClass(t, *meshR)
	if classifyTestSourceHint(privR.SourceHint) != "private" {
		t.Fatalf("private receipt source_hint: %+v", privR)
	}
	if classifyDigestTag(privR.Provenance.SourceHint) == "mesh" {
		t.Fatalf("private must not invent mesh provenance: %+v", privR)
	}

	// TemporalTags-only mesh (kernel EntryHasTag path) must still classify.
	temporalOnly := palace.MemoryEntry{
		ID:           "mesh-temporal",
		Timestamp:    meshAt,
		TemporalTags: []string{"source_hint:mesh"},
		Content:      palace.MemoryContent{Summary: "temporal mesh stamp"},
	}
	c, ok := digestCandidateFromEntry(temporalOnly)
	if !ok {
		t.Fatal("temporal mesh candidate")
	}
	assertReceiptCarriesMeshClass(t, c.receipt)
	if c.class != "mesh" {
		t.Fatalf("temporal tags class=%q", c.class)
	}

	// Bare local entry: palace_timeline, no invented mesh wire.
	bare := palace.MemoryEntry{
		ID:        "local-1",
		Timestamp: now,
		Content:   palace.MemoryContent{Summary: "local note"},
	}
	bareC, ok := digestCandidateFromEntry(bare)
	if !ok {
		t.Fatal("bare candidate")
	}
	if classifyTestReceipt(bareC.receipt) != "private" {
		t.Fatalf("bare class: %+v", bareC.receipt)
	}
	if classifyTestSourceHint(bareC.receipt.SourceHint) != "private" {
		t.Fatalf("bare source_hint: %+v", bareC.receipt)
	}
	if classifyDigestTag(bareC.receipt.Provenance.SourceHint) == "mesh" {
		t.Fatalf("bare must not invent mesh provenance: %+v", bareC.receipt)
	}
	for _, tag := range bareC.receipt.Tags {
		if classifyDigestTag(tag) == "mesh" {
			t.Fatalf("bare must not invent mesh tag: %+v", bareC.receipt)
		}
	}
}

func TestClassifyDigestTagSourceHintPrefix(t *testing.T) {
	if got := classifyDigestTag("source_hint:mesh"); got != "mesh" {
		t.Fatalf("source_hint:mesh → %q, want mesh", got)
	}
	if got := classifyDigestTag("source_hint:private"); got != "private" {
		t.Fatalf("source_hint:private → %q, want private", got)
	}
	if got := classifyDigestTag("source:iomesh-memory-mcp"); got != "" {
		t.Fatalf("host process label must not be a class: %q", got)
	}
	if got := classifyDigestTag("mcp_memory_ingest_turn"); got != "" {
		t.Fatalf("source_step must not be a class: %q", got)
	}
}

// assertReceiptCarriesMeshClass checks TUI ClassifyDigestReceipt can see mesh
// from source_hint and from provenance.source_hint / tags (the #66 residual).
func assertReceiptCarriesMeshClass(t *testing.T, r opsDigestReceipt) {
	t.Helper()
	if classifyTestReceipt(r) != "mesh" {
		t.Fatalf("TUI-shaped classify missed mesh: %+v", r)
	}
	if classifyTestSourceHint(r.SourceHint) != "mesh" {
		t.Fatalf("receipt source_hint=%q want mesh (not only palace_timeline): %+v", r.SourceHint, r)
	}
	if classifyDigestTag(r.Provenance.SourceHint) != "mesh" && !receiptHasMeshTag(r) {
		t.Fatalf("receipt must wire provenance.source_hint or source_hint:mesh tag: %+v", r)
	}
}

func receiptHasMeshTag(r opsDigestReceipt) bool {
	for _, tag := range r.Tags {
		if classifyDigestTag(tag) == "mesh" {
			return true
		}
	}
	return false
}

// classifyTestReceipt mirrors TUI ClassifyDigestReceipt (mesh|private only):
// provenance.source_hint and tags win over export-origin palace_timeline.
func classifyTestReceipt(r opsDigestReceipt) string {
	cands := make([]string, 0, 3+len(r.Tags))
	cands = append(cands, r.SourceHint, r.Provenance.SourceHint, r.Provenance.SourceStep)
	cands = append(cands, r.Tags...)
	fallback := ""
	for _, raw := range cands {
		switch classifyDigestTag(raw) {
		case "mesh":
			return "mesh"
		case "private":
			if fallback == "" {
				fallback = "private"
			}
		}
	}
	return fallback
}

// classifyTestSourceHint mirrors TUI ClassifyDigestSourceHint (mesh|private only).
func classifyTestSourceHint(hint string) string {
	h := strings.ToLower(strings.TrimSpace(hint))
	h = strings.ReplaceAll(h, "-", "_")
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
