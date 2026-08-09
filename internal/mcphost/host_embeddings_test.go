package mcphost

import (
	"path/filepath"
	"testing"
)

func TestNew_DefaultHashEmbeddings(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if h.EmbeddingMode() != "hash" {
		t.Fatalf("mode=%q", h.EmbeddingMode())
	}
	ps := h.Store("t")
	if ps == nil {
		t.Fatal("store nil")
	}
	// Store base under palace root
	if filepath.Base(ps.BaseDir) != "t" {
		t.Fatalf("base=%q", ps.BaseDir)
	}
}
