package mcphost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTenantSingleSegment(t *testing.T) {
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := h.ResolveTenant("")
	if err != nil || got != "dogfood" {
		t.Fatalf("empty: got %q err=%v want dogfood", got, err)
	}
	got, err = h.ResolveTenant("  ")
	if err != nil || got != "dogfood" {
		t.Fatalf("whitespace: got %q err=%v want dogfood", got, err)
	}
	got, err = h.ResolveTenant("t-a")
	if err != nil || got != "t-a" {
		t.Fatalf("valid: got %q err=%v", got, err)
	}

	for _, bad := range []string{".", "..", "a/b", `a\b`, "foo/../x", "/abs", "../x", "foo/", `foo\bar`} {
		if _, err := h.ResolveTenant(bad); err == nil {
			t.Fatalf("ResolveTenant(%q) should fail", bad)
		}
		if h.TenantDir(bad) != "" {
			t.Fatalf("TenantDir(%q) must not join", bad)
		}
		if h.Store(bad) != nil {
			t.Fatalf("Store(%q) must be nil", bad)
		}
	}
}

func TestNewRejectsBadDefaultTenant(t *testing.T) {
	root := t.TempDir()
	for _, bad := range []string{".", "..", "a/b", `a\b`, "/abs"} {
		if _, err := New(Config{PalaceRoot: root, DefaultTenant: bad}); err == nil {
			t.Fatalf("New default %q should fail", bad)
		}
	}
	h, err := New(Config{PalaceRoot: root})
	if err != nil {
		t.Fatalf("empty default: %v", err)
	}
	got, err := h.ResolveTenant("")
	if err != nil || got != "default" {
		t.Fatalf("implicit default: got %q err=%v", got, err)
	}
}

func TestBadTenantDoesNotEscapePalaceRoot(t *testing.T) {
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if h.Store("..") != nil || h.Store("foo/bar") != nil {
		t.Fatal("Store must reject escaping tenants")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("palace root should stay empty, got %v", names(entries))
	}
	// Valid tenant still lands under palace root.
	ps := h.Store("ok")
	if ps == nil {
		t.Fatal("valid store nil")
	}
	if !strings.HasPrefix(ps.BaseDir, root) || filepath.Base(ps.BaseDir) != "ok" {
		t.Fatalf("store base %q not under %s/ok", ps.BaseDir, root)
	}
}

func names(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}
