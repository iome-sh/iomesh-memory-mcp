package mcphost

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_DefaultHashEmbeddings(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	t.Setenv(EnvPersistEmbeddings, "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if h.EmbeddingMode() != "hash" {
		t.Fatalf("mode=%q", h.EmbeddingMode())
	}
	if h.PersistEmbeddingsHonesty() != "off" {
		t.Fatalf("persist_embeddings=%q want off", h.PersistEmbeddingsHonesty())
	}
	ps := h.Store("t")
	if ps == nil {
		t.Fatal("store nil")
	}
	// Store base under palace root
	if filepath.Base(ps.BaseDir) != "t" {
		t.Fatalf("base=%q", ps.BaseDir)
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("default PersistEmbeddings must be false")
	}
}

func TestStore_PersistEmbeddingsDefaultOff(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	t.Setenv(EnvPersistEmbeddings, "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t"})
	if err != nil {
		t.Fatal(err)
	}
	ps := h.Store("t")
	if ps == nil {
		t.Fatal("store nil")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("unset MEMORY_PERSIST_EMBEDDINGS must leave PersistEmbeddings false")
	}
	if ps.Config.EmbeddingModel != "" && !strings.EqualFold(ps.Config.EmbeddingModel, "hash") {
		t.Fatalf("hash EmbeddingModel=%q want empty or hash", ps.Config.EmbeddingModel)
	}
}

func TestStore_PersistEmbeddingsHashEnvOnRemainsFalse(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	for _, v := range []string{"1", "true", "TRUE", "on", "yes"} {
		t.Run("env="+v, func(t *testing.T) {
			t.Setenv(EnvPersistEmbeddings, v)
			h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "t"})
			if err != nil {
				t.Fatal(err)
			}
			if h.EmbeddingMode() != "hash" {
				t.Fatalf("mode=%q", h.EmbeddingMode())
			}
			if h.PersistEmbeddingsHonesty() != "off" {
				t.Fatalf("persist_embeddings=%q want off (hash never persists)", h.PersistEmbeddingsHonesty())
			}
			ps := h.Store("t")
			if ps == nil {
				t.Fatal("store nil")
			}
			if ps.Config.PersistEmbeddings {
				t.Fatal("hash + MEMORY_PERSIST_EMBEDDINGS must leave PersistEmbeddings false")
			}
		})
	}
}

func TestStore_PersistEmbeddingsONNXInjectEnvOff(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	t.Setenv(EnvPersistEmbeddings, "")
	h, err := New(Config{
		PalaceRoot:    t.TempDir(),
		DefaultTenant: "t",
		EmbeddingMode: "onnx",
	})
	if err != nil {
		t.Fatal(err)
	}
	ps := h.Store("t")
	if ps == nil {
		t.Fatal("store nil")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("onnx without MEMORY_PERSIST_EMBEDDINGS must leave PersistEmbeddings false")
	}
	if h.PersistEmbeddingsHonesty() != "off" {
		t.Fatalf("persist_embeddings=%q want off", h.PersistEmbeddingsHonesty())
	}
}

func TestStore_PersistEmbeddingsONNXInjectEnvOn(t *testing.T) {
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	t.Setenv(EnvPersistEmbeddings, "true")
	h, err := New(Config{
		PalaceRoot:    t.TempDir(),
		DefaultTenant: "t",
		EmbeddingMode: "onnx", // inject without loading a real ONNX model
	})
	if err != nil {
		t.Fatal(err)
	}
	if h.EmbeddingMode() != "onnx" {
		t.Fatalf("mode=%q want onnx", h.EmbeddingMode())
	}
	h.embedDim = 384
	if h.PersistEmbeddingsHonesty() != "on" {
		t.Fatalf("persist_embeddings=%q want on", h.PersistEmbeddingsHonesty())
	}
	ps := h.Store("t")
	if ps == nil {
		t.Fatal("store nil")
	}
	if !ps.Config.PersistEmbeddings {
		t.Fatal("onnx + MEMORY_PERSIST_EMBEDDINGS must set PersistEmbeddings true")
	}
	if ps.Config.EmbeddingModel == "" || strings.EqualFold(ps.Config.EmbeddingModel, "hash") {
		t.Fatalf("onnx EmbeddingModel=%q want non-hash id", ps.Config.EmbeddingModel)
	}
	if ps.Config.EmbeddingDim != 384 {
		t.Fatalf("EmbeddingDim=%d want 384", ps.Config.EmbeddingDim)
	}
}
