package mcphost

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestIngestRetrieveListRoundTrip(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	_, out, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID:  "sess-1",
		Role:       "user",
		Content:    "alpha project notes for memory kernel",
		EventTime:  time.Now().UTC().Format(time.RFC3339),
		SessionSeq: 1,
	})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if out.MemoryID == "" {
		t.Fatal("expected memory_id")
	}
	if out.DualWrite != "off" || out.Audited {
		t.Fatalf("dual_write must be off: %+v", out)
	}
	if out.Tenant != "dogfood" {
		t.Fatalf("tenant: got %q", out.Tenant)
	}

	// IngestTurn also writes semantic facts from extracted atoms — give FS a moment not needed for sync FS.
	_, ret, err := h.handleRetrieve(ctx, nil, retrieveInput{
		Query: "alpha project",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(ret.Memories) == 0 {
		t.Fatal("hash retrieve must keyword-hit ingested turn (do not inject hash QueryVec)")
	}

	_, status, err := h.handleCompactStatus(ctx, nil, compactStatusInput{})
	if err != nil {
		t.Fatalf("compact status: %v", err)
	}
	if status.TotalEntries < 1 {
		t.Fatalf("expected total_entries >= 1, got %+v", status)
	}
	if status.DualWrite != "off" || !status.NotMemoryGA {
		t.Fatalf("honesty locks: %+v", status)
	}

	_, sem, err := h.handleSearchSemantic(ctx, nil, searchSemanticInput{Query: "alpha", Limit: 10})
	if err != nil {
		t.Fatalf("semantic: %v", err)
	}
	if !strings.Contains(sem.Note, "not Memory GA") {
		t.Fatalf("semantic note honesty: %q", sem.Note)
	}

	_, facts, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{
		AsOf:  time.Now().UTC().Format(time.RFC3339),
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("facts as of: %v", err)
	}
	if facts.AsOf == "" {
		t.Fatal("expected as_of")
	}
}

func TestRetrieveHashKeepsHyphenNeedle(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if h.EmbeddingMode() != "hash" {
		t.Fatalf("embedding mode = %q, want hash", h.EmbeddingMode())
	}
	ctx := context.Background()
	const needle = "zircon-lantern-4829"
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: "sess-needle",
		Role:      "user",
		Content:   "lab note " + needle + " for retrieve recall",
	}); err != nil {
		t.Fatalf("ingest needle: %v", err)
	}
	for i := 0; i < 8; i++ {
		if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
			SessionID: "sess-needle",
			Role:      "user",
			Content:   "unrelated distractor checklist item " + string(rune('a'+i)),
		}); err != nil {
			t.Fatalf("ingest distractor: %v", err)
		}
	}
	_, ret, err := h.handleRetrieve(ctx, nil, retrieveInput{Query: needle, Limit: 5})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	found := false
	for _, m := range ret.Memories {
		if strings.Contains(m.Summary, needle) || strings.Contains(m.Full, needle) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("hash retrieve missed exact needle %q in top %d (ids=%d)", needle, 5, len(ret.Memories))
	}
	if len(ret.Memories) > 5 {
		t.Fatalf("Limit 5 not applied; n=%d", len(ret.Memories))
	}

	_, listed, err := h.handleList(ctx, nil, listInput{Query: needle, Limit: 5})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed.Entries) == 0 {
		t.Fatalf("list Query=%q returned no entries", needle)
	}
	top := listed.Entries[0]
	if !strings.Contains(top.Summary, needle) && !strings.Contains(top.Full, needle) {
		t.Fatalf("list rank 1 missed hyphen needle %q: %+v", needle, top)
	}
	if len(listed.Entries) > 5 {
		t.Fatalf("list Limit 5 not applied; n=%d", len(listed.Entries))
	}
}

func TestWriteFactAndSupersede(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const key = "lattice:status"

	_, first, err := h.handleWrite(ctx, nil, writeInput{
		Summary:   "lattice open",
		Full:      "human-gate lattice is open",
		EntityKey: key,
		Tier:      2,
	})
	if err != nil {
		t.Fatalf("write 1: %v", err)
	}
	if first.DualWrite != "off" || first.Audited || !first.Superseded {
		t.Fatalf("honesty/supersede: %+v", first)
	}
	if first.Tier != 2 {
		t.Fatalf("default/set tier: %+v", first)
	}

	_, second, err := h.handleWrite(ctx, nil, writeInput{
		Summary:   "lattice closed",
		Full:      "human-gate lattice is closed",
		EntityKey: key,
	})
	if err != nil {
		t.Fatalf("write 2: %v", err)
	}
	if !second.Superseded {
		t.Fatal("second write should WriteAndSupersede")
	}

	_, facts, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{
		AsOf:   time.Now().UTC().Format(time.RFC3339),
		Entity: key,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("facts as of: %v", err)
	}
	foundSecond := false
	for _, f := range facts.Facts {
		if f.ID == first.MemoryID {
			t.Fatalf("superseded fact still valid: %+v", f)
		}
		if f.ID == second.MemoryID {
			foundSecond = true
		}
	}
	if !foundSecond {
		t.Fatalf("new fact missing from as-of; got %+v", facts.Facts)
	}

	off := false
	_, plain, err := h.handleWrite(ctx, nil, writeInput{
		Summary:   "catalog honesty pin",
		Tags:      []string{"honesty:catalog"},
		EntityKey: "catalog:honesty",
		Supersede: &off,
	})
	if err != nil {
		t.Fatalf("write no-supersede: %v", err)
	}
	if plain.Superseded || plain.DualWrite != "off" {
		t.Fatalf("plain write: %+v", plain)
	}

	_, empty, err := h.handleWrite(ctx, nil, writeInput{})
	if err == nil {
		t.Fatal("expected summary or full required")
	}
	if empty.MemoryID != "" {
		t.Fatalf("empty write leaked id: %+v", empty)
	}
}

func TestRelatedAndSupersedeEntity(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const key = "person:alice"

	_, wrote, err := h.handleWrite(ctx, nil, writeInput{
		Summary:   "alice prefers dark mode",
		EntityKey: key,
		Supersede: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	_, rel, err := h.handleRelated(ctx, nil, relatedInput{SeedEntity: key, Limit: 10})
	if err != nil {
		t.Fatalf("related: %v", err)
	}
	found := false
	for _, m := range rel.Memories {
		if m.ID == wrote.MemoryID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("related missed written fact; got %+v", rel.Memories)
	}
	if !strings.Contains(rel.Note, "not Memory GA") {
		t.Fatalf("related note honesty: %q", rel.Note)
	}

	_, empty, err := h.handleRelated(ctx, nil, relatedInput{})
	if err == nil {
		t.Fatal("expected seed required")
	}
	_ = empty

	_, sup, err := h.handleSupersedeEntity(ctx, nil, supersedeEntityInput{EntityKey: key})
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}
	if sup.Updated < 1 {
		t.Fatalf("expected at least one closed fact: %+v", sup)
	}
	if sup.DualWrite != "off" || sup.Audited {
		t.Fatalf("supersede honesty: %+v", sup)
	}

	_, facts, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{
		AsOf:   time.Now().UTC().Format(time.RFC3339),
		Entity: key,
	})
	if err != nil {
		t.Fatalf("facts as of: %v", err)
	}
	for _, f := range facts.Facts {
		if f.ID == wrote.MemoryID {
			t.Fatalf("closed fact still valid: %+v", f)
		}
	}
}

func boolPtr(v bool) *bool { return &v }

func TestTenantPathIsolation(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:    "t-a",
		SessionID: "s",
		Role:      "user",
		Content:   "secret for tenant A only",
	}); err != nil {
		t.Fatalf("ingest A: %v", err)
	}
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:    "t-b",
		SessionID: "s",
		Role:      "user",
		Content:   "content for tenant B",
	}); err != nil {
		t.Fatalf("ingest B: %v", err)
	}

	_, listA, err := h.handleList(ctx, nil, listInput{Tenant: "t-a", Limit: 50})
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	_, listB, err := h.handleList(ctx, nil, listInput{Tenant: "t-b", Limit: 50})
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(listA.Entries) == 0 || len(listB.Entries) == 0 {
		t.Fatalf("expected entries in both tenants: A=%d B=%d", len(listA.Entries), len(listB.Entries))
	}
	for _, e := range listA.Entries {
		if strings.Contains(strings.ToLower(e.Full+e.Summary), "tenant b") {
			t.Fatalf("tenant A saw B content: %+v", e)
		}
	}
	// Path isolation: different base dirs.
	if h.TenantDir("t-a") == h.TenantDir("t-b") {
		t.Fatal("tenant dirs must differ")
	}
	if !strings.HasSuffix(h.TenantDir("t-a"), "t-a") {
		t.Fatalf("tenant dir: %s", h.TenantDir("t-a"))
	}
}

func TestRegisterSDKServer(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sdk := h.NewSDKServer()
	if sdk == nil {
		t.Fatal("nil sdk server")
	}
	names := LeanToolNames()
	if len(names) < 9 {
		t.Fatalf("lean tools=%d want >= 9: %v", len(names), names)
	}
}

func TestToolJSONHelpers(t *testing.T) {
	r := toolJSON(map[string]any{"ok": true})
	if r.IsError {
		t.Fatal("expected success")
	}
	if len(r.Content) != 1 {
		t.Fatalf("content: %d", len(r.Content))
	}
	er := toolError(context.Canceled)
	if !er.IsError {
		t.Fatal("expected error result")
	}
	// Ensure ingest output marshals dual_write honesty.
	b, _ := json.Marshal(ingestTurnOutput{DualWrite: "off", Audited: false})
	if !strings.Contains(string(b), `"dual_write":"off"`) {
		t.Fatalf("marshal: %s", b)
	}
}

func TestParseTimeHelpers(t *testing.T) {
	if _, ok := parseOptionalTime(""); ok {
		t.Fatal("empty should be unset")
	}
	now := time.Now().UTC().Truncate(time.Second)
	got, ok := parseOptionalTime(now.Format(time.RFC3339))
	if !ok || !got.Equal(now) {
		t.Fatalf("parse: ok=%v got=%v want=%v", ok, got, now)
	}
	if parseTimeOrNow("").IsZero() {
		t.Fatal("now fallback")
	}
}
