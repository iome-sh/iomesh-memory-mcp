package mcphost

import (
	"context"
	"strings"
	"testing"
)

// Fake paste shapes for tests only — not live credentials.
const (
	fakeGitHubPAT = "ghp_TESTFAKE000000000000000000000000"
	fakeSKToken   = "sk-testfake00000000000000000000000"
)

func TestRedactSecretsGitHubAndSK(t *testing.T) {
	in := "paste " + fakeGitHubPAT + " and " + fakeSKToken + " please"
	got := RedactSecrets(in)
	if strings.Contains(got, fakeGitHubPAT) || strings.Contains(got, "ghp_") {
		t.Fatalf("ghp_ still present: %q", got)
	}
	if strings.Contains(got, fakeSKToken) || strings.Contains(got, "sk-test") {
		t.Fatalf("sk- still present: %q", got)
	}
	if !strings.Contains(got, RedactPlaceholder) {
		t.Fatalf("expected %q in %q", RedactPlaceholder, got)
	}
	if !SecretsWereRedacted(in) {
		t.Fatal("expected SecretsWereRedacted")
	}
}

func TestRedactSecretsLeavesProse(t *testing.T) {
	in := "ask-me later about the zircon-lantern-4829 notes"
	if got := RedactSecrets(in); got != in {
		t.Fatalf("prose changed: %q → %q", in, got)
	}
	if SecretsWereRedacted(in) {
		t.Fatal("prose must not count as redacted")
	}
	if RedactSecrets("") != "" {
		t.Fatal("empty must stay empty")
	}
}

func TestIngestTurnRedactsSecretsBeforeWrite(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	content := "operator demo token " + fakeGitHubPAT + " openai " + fakeSKToken
	if _, _, err := h.handleIngestTurn(ctx, nil, ingestTurnInput{
		Tenant:    "dogfood",
		SessionID: "sess-dlp",
		Role:      "user",
		Content:   content,
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	_, listed, err := h.handleList(ctx, nil, listInput{Tenant: "dogfood", Limit: 50})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed.Entries) == 0 {
		t.Fatal("expected stored turn")
	}
	for _, e := range listed.Entries {
		blob := e.Summary + " " + e.Full
		if strings.Contains(blob, fakeGitHubPAT) || strings.Contains(blob, fakeSKToken) {
			t.Fatalf("palace stored raw secret: %+v", e)
		}
		if strings.Contains(blob, "ghp_") || strings.Contains(blob, "sk-test") {
			t.Fatalf("palace stored secret prefix: %+v", e)
		}
	}
	found := false
	for _, e := range listed.Entries {
		if strings.Contains(e.Full, RedactPlaceholder) || strings.Contains(e.Summary, RedactPlaceholder) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q in stored content: %+v", RedactPlaceholder, listed.Entries)
	}
}

func TestWriteRedactsSecretsBeforeWrite(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if _, _, err := h.handleWrite(ctx, nil, writeInput{
		Tenant:  "dogfood",
		Summary: "token " + fakeGitHubPAT,
		Full:    "full " + fakeSKToken,
	}); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, listed, err := h.handleList(ctx, nil, listInput{Tenant: "dogfood", Query: "token", Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed.Entries) == 0 {
		t.Fatal("expected stored fact")
	}
	for _, e := range listed.Entries {
		blob := e.Summary + " " + e.Full
		if strings.Contains(blob, fakeGitHubPAT) || strings.Contains(blob, fakeSKToken) {
			t.Fatalf("write stored raw secret: %+v", e)
		}
	}
}
