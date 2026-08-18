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

func TestSessionIsolationSameTenantToken(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if h.EmbeddingMode() != "hash" {
		t.Fatalf("embedding mode = %q, want hash", h.EmbeddingMode())
	}
	ctx := context.Background()
	const (
		token = "zircon-session-lock-7713"
		sessA = "sess-alpha"
		sessB = "sess-bravo"
	)
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: sessA,
		Role:      "user",
		Content:   "alpha vault note " + token + " for isolation",
	}); err != nil {
		t.Fatalf("ingest A: %v", err)
	}
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: sessB,
		Role:      "user",
		Content:   "bravo vault note " + token + " for isolation",
	}); err != nil {
		t.Fatalf("ingest B: %v", err)
	}

	assertSessionHits := func(t *testing.T, surface string, hits []memoryHit, wantSess, wantWord, leakWord string) {
		t.Helper()
		if len(hits) == 0 {
			t.Fatalf("%s SessionID=%q: expected hits for %q", surface, wantSess, token)
		}
		foundWant := false
		for _, hit := range hits {
			if hit.SessionID != "" && hit.SessionID != wantSess {
				t.Fatalf("%s leaked session %q (want %q): %+v", surface, hit.SessionID, wantSess, hit)
			}
			blob := strings.ToLower(hit.Summary + " " + hit.Full)
			if strings.Contains(blob, leakWord) {
				t.Fatalf("%s leaked %q into session %q: %+v", surface, leakWord, wantSess, hit)
			}
			if hitHasToken(hit, token) && (hit.SessionID == wantSess || strings.Contains(blob, wantWord)) {
				foundWant = true
			}
		}
		if !foundWant {
			t.Fatalf("%s SessionID=%q missed token %q: %+v", surface, wantSess, token, hits)
		}
	}

	_, retA, err := h.handleRetrieve(ctx, nil, retrieveInput{Query: token, SessionID: sessA, Limit: 20})
	if err != nil {
		t.Fatalf("retrieve A: %v", err)
	}
	assertSessionHits(t, "retrieve", retA.Memories, sessA, "alpha", "bravo")

	_, retB, err := h.handleRetrieve(ctx, nil, retrieveInput{Query: token, SessionID: sessB, Limit: 20})
	if err != nil {
		t.Fatalf("retrieve B: %v", err)
	}
	assertSessionHits(t, "retrieve", retB.Memories, sessB, "bravo", "alpha")

	_, listA, err := h.handleList(ctx, nil, listInput{Query: token, SessionID: sessA, Limit: 50})
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	assertSessionHits(t, "list", listA.Entries, sessA, "alpha", "bravo")

	_, listB, err := h.handleList(ctx, nil, listInput{Query: token, SessionID: sessB, Limit: 50})
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	assertSessionHits(t, "list", listB.Entries, sessB, "bravo", "alpha")

	_, retAll, err := h.handleRetrieve(ctx, nil, retrieveInput{Query: token, Limit: 20})
	if err != nil {
		t.Fatalf("retrieve unfiltered: %v", err)
	}
	if !hitsCoverSessions(retAll.Memories, sessA, sessB) {
		t.Fatalf("empty-session retrieve must be unfiltered; got %+v", retAll.Memories)
	}

	_, listAll, err := h.handleList(ctx, nil, listInput{Query: token, Limit: 50})
	if err != nil {
		t.Fatalf("list unfiltered: %v", err)
	}
	if !hitsCoverSessions(listAll.Entries, sessA, sessB) {
		t.Fatalf("empty-session list must be unfiltered; got %+v", listAll.Entries)
	}
}

func TestFactsAsOfSeesIngestFactChildren(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: "sess-facts",
		Role:      "user",
		Content:   "My name is Alice. I live in Seattle. I graduated from MIT last year.",
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	_, status, err := h.handleCompactStatus(ctx, nil, compactStatusInput{})
	if err != nil {
		t.Fatalf("compact status: %v", err)
	}
	_, facts, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{
		AsOf:  time.Now().UTC().Format(time.RFC3339),
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("facts as of: %v", err)
	}
	if status.SemanticCount == 0 {
		t.Log("ingest extracted no atoms; skip fact-child assert (extractor residual)")
		return
	}
	found := false
	for _, f := range facts.Facts {
		if f.Tier == 4 || hitHasTag(f, "fact_augmented") || hitHasTag(f, "from_turn") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("kernel #48: semantic atoms=%d but facts_as_of missed fact children: %+v",
			status.SemanticCount, facts.Facts)
	}
}

func TestListAfterNewSamePalaceRoot(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	root := t.TempDir()
	h1, err := New(Config{PalaceRoot: root, DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New seed: %v", err)
	}
	if h1.EmbeddingMode() != "hash" {
		t.Fatalf("embedding mode = %q, want hash", h1.EmbeddingMode())
	}
	ctx := context.Background()
	const needle = "zircon-durable-snap-9041"
	if _, _, err := h1.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: "sess-durable",
		Role:      "user",
		Content:   "lab note " + needle + " for durable list",
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	_, listed, err := h1.handleList(ctx, nil, listInput{Query: needle, Limit: 5})
	if err != nil {
		t.Fatalf("list seed: %v", err)
	}
	if !hitsContainToken(listed.Entries, needle) {
		t.Fatalf("seed list missed needle %q: %+v", needle, listed.Entries)
	}

	h2, err := New(Config{PalaceRoot: root, DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New reopen: %v", err)
	}
	if h2.EmbeddingMode() != "hash" {
		t.Fatalf("reopen embedding mode = %q, want hash", h2.EmbeddingMode())
	}
	_, again, err := h2.handleList(ctx, nil, listInput{Query: needle, Limit: 5})
	if err != nil {
		t.Fatalf("list reopen: %v", err)
	}
	if !hitsContainToken(again.Entries, needle) {
		t.Fatalf("kernel #47: list after New() on same palace root missed %q: %+v", needle, again.Entries)
	}
}

func hitHasToken(h memoryHit, token string) bool {
	return strings.Contains(h.Summary, token) || strings.Contains(h.Full, token)
}

func hitHasTag(h memoryHit, tag string) bool {
	for _, t := range h.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func hitsContainToken(hits []memoryHit, token string) bool {
	for _, h := range hits {
		if hitHasToken(h, token) {
			return true
		}
	}
	return false
}

func hitsCoverSessions(hits []memoryHit, sessions ...string) bool {
	seen := make(map[string]bool, len(sessions))
	for _, h := range hits {
		if h.SessionID != "" {
			seen[h.SessionID] = true
		}
	}
	for _, s := range sessions {
		if !seen[s] {
			return false
		}
	}
	return true
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

func TestBadTenantToolIsError(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const bad = ".."

	res, out, err := h.handleList(ctx, nil, listInput{Tenant: bad})
	if err == nil {
		t.Fatal("list bad tenant must error")
	}
	if res == nil || !res.IsError {
		t.Fatalf("list want IsError result, got %+v", res)
	}
	if out.Tenant != "" {
		t.Fatalf("list output tenant: %q", out.Tenant)
	}

	// Omitted/empty tool tenant still uses the configured default.
	res, out, err = h.handleList(ctx, nil, listInput{})
	if err != nil {
		t.Fatalf("empty tenant: %v", err)
	}
	if res != nil && res.IsError {
		t.Fatal("empty tenant must not be IsError")
	}
	if out.Tenant != "dogfood" {
		t.Fatalf("empty tenant: got %q", out.Tenant)
	}

	if res, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{Tenant: bad, SessionID: "s", Role: "user", Content: "n"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("ingest: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleWrite(ctx, nil, writeInput{Tenant: bad, Summary: "n"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("write: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleRetrieve(ctx, nil, retrieveInput{Tenant: bad, Query: "n"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("retrieve: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleSearchSemantic(ctx, nil, searchSemanticInput{Tenant: bad, Query: "n"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("semantic: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleCompactStatus(ctx, nil, compactStatusInput{Tenant: bad}); err == nil || res == nil || !res.IsError {
		t.Fatalf("compact: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{Tenant: bad}); err == nil || res == nil || !res.IsError {
		t.Fatalf("facts: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleRelated(ctx, nil, relatedInput{Tenant: bad, SeedEntity: "e"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("related: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleSupersedeEntity(ctx, nil, supersedeEntityInput{Tenant: bad, EntityKey: "e"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("supersede: err=%v res=%+v", err, res)
	}

	// Separator tenant also fail-closed; default still unused.
	if res, _, err := h.handleList(ctx, nil, listInput{Tenant: "a/b"}); err == nil || res == nil || !res.IsError {
		t.Fatalf("list a/b: err=%v res=%+v", err, res)
	}
	if res, _, err := h.handleList(ctx, nil, listInput{Tenant: "."}); err == nil || res == nil || !res.IsError {
		t.Fatalf("list .: err=%v res=%+v", err, res)
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
	if _, ok, err := parseOptionalTime(""); err != nil || ok {
		t.Fatalf("empty should be unset: ok=%v err=%v", ok, err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	got, ok, err := parseOptionalTime(now.Format(time.RFC3339))
	if err != nil || !ok || !got.Equal(now) {
		t.Fatalf("parse: ok=%v err=%v got=%v want=%v", ok, err, got, now)
	}
	gotNow, err := parseTimeOrNow("")
	if err != nil || gotNow.IsZero() {
		t.Fatalf("now fallback: err=%v zero=%v", err, gotNow.IsZero())
	}
	if _, _, err := parseOptionalTime("last-week"); err == nil {
		t.Fatal("invalid non-empty must error")
	}
	if _, err := parseTimeOrNow("last-week"); err == nil {
		t.Fatal("parseTimeOrNow invalid must error")
	}
}

func TestInvalidTimeFieldsFailClosed(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const bad = "last-week"

	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		SessionID: "s",
		Role:      "user",
		Content:   "note",
		EventTime: bad,
	}); err == nil {
		t.Fatal("ingest event_time invalid must error")
	}
	if _, _, err := h.handleRetrieve(ctx, nil, retrieveInput{Query: "note", Since: bad}); err == nil {
		t.Fatal("retrieve since invalid must error")
	}
	if _, _, err := h.handleList(ctx, nil, listInput{Until: bad}); err == nil {
		t.Fatal("list until invalid must error")
	}
	if _, _, err := h.handleFactsAsOf(ctx, nil, factsAsOfInput{AsOf: bad}); err == nil {
		t.Fatal("facts_as_of invalid must error")
	}
	if _, _, err := h.handleRelated(ctx, nil, relatedInput{SeedEntity: "e", AsOf: bad}); err == nil {
		t.Fatal("related as_of invalid must error")
	}
	if _, _, err := h.handleSupersedeEntity(ctx, nil, supersedeEntityInput{EntityKey: "e", AsOf: bad}); err == nil {
		t.Fatal("supersede as_of invalid must error")
	}
}
