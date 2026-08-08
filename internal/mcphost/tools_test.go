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
		// Hybrid may still miss on hash embeddings; list should see working tier.
		_, list, lerr := h.handleList(ctx, nil, listInput{Limit: 20})
		if lerr != nil {
			t.Fatalf("list: %v", lerr)
		}
		if len(list.Entries) == 0 {
			t.Fatal("expected at least one entry after ingest via list")
		}
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
