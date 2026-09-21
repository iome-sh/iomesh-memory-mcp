package mcphost

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloudModeRequiresTenant(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	_, err := New(Config{PalaceRoot: t.TempDir(), Cloud: true})
	if err == nil {
		t.Fatal("cloud New without tenant must fail")
	}
	if !strings.Contains(err.Error(), "MEMORY_TENANT") && !strings.Contains(err.Error(), "tenant") {
		t.Fatalf("want tenant required, got %v", err)
	}
}

func TestCloudModeEnvRequiresTenant(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "1")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	_, err := New(Config{PalaceRoot: t.TempDir()})
	if err == nil {
		t.Fatal("MEMORY_CLOUD=1 without tenant must fail")
	}
}

func TestCloudModeRejectsSecondTenant(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	root := t.TempDir()
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "ws-1", Cloud: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = h.Close() }()

	if !h.Cloud() {
		t.Fatal("Cloud() false")
	}
	if h.StoreCount() != 1 {
		t.Fatalf("stores map len=%d want 1 at start", h.StoreCount())
	}
	if h.Store("ws-1") == nil {
		t.Fatal("configured tenant store nil")
	}
	if h.StoreCount() != 1 {
		t.Fatalf("stores map grew after matching tenant: %d", h.StoreCount())
	}

	_, err = h.ResolveTenant("ws-2")
	if !errors.Is(err, ErrCloudTenantRejected) {
		t.Fatalf("ResolveTenant second: %v", err)
	}
	if ErrorHTTPStatus(err) != CloudTenantRejectStatus || ErrorHTTPStatus(err) != 400 {
		t.Fatalf("HTTP status: %d", ErrorHTTPStatus(err))
	}
	if h.Store("ws-2") != nil {
		t.Fatal("Store must not open a second tenant")
	}
	if h.TenantDir("ws-2") != "" {
		t.Fatal("TenantDir must not join a second tenant")
	}
	if h.StoreCount() != 1 {
		t.Fatalf("stores map grew: %d", h.StoreCount())
	}

	ctx := context.Background()
	res, _, err := h.handleList(ctx, nil, listInput{Tenant: "ws-2"})
	if err == nil || !errors.Is(err, ErrCloudTenantRejected) {
		t.Fatalf("tool second tenant err=%v", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("tool second tenant want IsError 400, got %+v", res)
	}
	if ErrorHTTPStatus(err) != 400 {
		t.Fatalf("tool second tenant status %d", ErrorHTTPStatus(err))
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("tool error should say 400: %v", err)
	}

	_, listed, err := h.handleList(ctx, nil, listInput{Tenant: "ws-1"})
	if err != nil {
		t.Fatalf("matching tenant list: %v", err)
	}
	if listed.Tenant != "ws-1" {
		t.Fatalf("tenant: %q", listed.Tenant)
	}
	if h.StoreCount() != 1 {
		t.Fatalf("stores map after matching tool: %d", h.StoreCount())
	}
	if h.PersistEmbeddingsHonesty() != "off" {
		t.Fatalf("E-G4 persist_embeddings=%q want off", h.PersistEmbeddingsHonesty())
	}
}

func TestLocalDevAllowsMultipleTenants(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if h.Cloud() {
		t.Fatal("local-dev must not be cloud")
	}
	if h.Store("a") == nil || h.Store("b") == nil {
		t.Fatal("local-dev stores")
	}
	if h.StoreCount() != 2 {
		t.Fatalf("local-dev map len=%d want 2", h.StoreCount())
	}
	if _, err := os.Stat(filepath.Join(h.cfg.PalaceRoot, PalaceLockFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("local-dev must not write palace.lock: %v", err)
	}
}

func TestCloudStoreSetsTransactionalIngest(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "ws-1", Cloud: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = h.Close() }()
	ps := h.Store("ws-1")
	if ps == nil {
		t.Fatal("store")
	}
	if !ps.Config.TransactionalIngest {
		t.Fatal("cloud store must set TransactionalIngest")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("must not set PersistEmbeddings (E-G4)")
	}
}

func TestCloudEnvSetsTransactionalIngest(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "1")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "ws-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = h.Close() }()
	if !h.Cloud() {
		t.Fatal("MEMORY_CLOUD=1 should enable cloud")
	}
	ps := h.Store("ws-1")
	if ps == nil {
		t.Fatal("store")
	}
	if !ps.Config.TransactionalIngest {
		t.Fatal("MEMORY_CLOUD store must set TransactionalIngest")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("must not set PersistEmbeddings (E-G4)")
	}
}

func TestLocalDevTransactionalIngestDefaultFalse(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "dogfood"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ps := h.Store("dogfood")
	if ps == nil {
		t.Fatal("store")
	}
	if ps.Config.TransactionalIngest {
		t.Fatal("local-dev TransactionalIngest must stay false (partial persist)")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("must not set PersistEmbeddings")
	}
}

func TestCloudPIDLockRejectsLivePID(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	root := t.TempDir()
	h1, err := New(Config{PalaceRoot: root, DefaultTenant: "ws-1", Cloud: true})
	if err != nil {
		t.Fatalf("first New: %v", err)
	}
	defer func() { _ = h1.Close() }()

	lockPath := filepath.Join(root, PalaceLockFile)
	b, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("palace.lock missing: %v", err)
	}
	var rec palaceLockRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("lock json: %v body=%s", err, b)
	}
	if rec.PID != os.Getpid() {
		t.Fatalf("lock pid %d want %d", rec.PID, os.Getpid())
	}
	if strings.TrimSpace(rec.Started) == "" {
		t.Fatal("lock missing start time")
	}

	_, err = New(Config{PalaceRoot: root, DefaultTenant: "ws-1", Cloud: true})
	if err == nil {
		t.Fatal("second New must fail while pid is alive")
	}
	if !errors.Is(err, ErrPalaceLockHeld) {
		t.Fatalf("want ErrPalaceLockHeld, got %v", err)
	}
}

func TestCloudPIDLockAllowsDeadPID(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "")
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("helper exit: %v", err)
	}
	dead := cmd.ProcessState.Pid()
	if dead <= 0 {
		t.Fatal("dead pid")
	}
	if pidAlive(dead) {
		t.Skip("pid reused; skip stale-lock probe")
	}
	rec := palaceLockRecord{PID: dead, Started: "2020-01-01T00:00:00Z"}
	body, _ := json.Marshal(rec)
	if err := os.WriteFile(filepath.Join(root, PalaceLockFile), append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	h, err := New(Config{PalaceRoot: root, DefaultTenant: "ws-1", Cloud: true})
	if err != nil {
		t.Fatalf("stale lock should be replaced: %v", err)
	}
	_ = h.Close()
}

func TestCloudDoesNotSetPersistEmbeddings(t *testing.T) {
	t.Setenv("MEMORY_CLOUD", "1")
	t.Setenv("MEMORY_PERSIST_EMBEDDINGS", "")
	t.Setenv("MEMORY_ONNX_MODEL_PATH", "")
	h, err := New(Config{PalaceRoot: t.TempDir(), DefaultTenant: "ws-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = h.Close() }()
	if !h.Cloud() {
		t.Fatal("MEMORY_CLOUD=1 should enable cloud")
	}
	if h.PersistEmbeddingsHonesty() != "off" {
		t.Fatalf("persist_embeddings=%q", h.PersistEmbeddingsHonesty())
	}
	ps := h.Store("ws-1")
	if ps == nil {
		t.Fatal("store")
	}
	if ps.Config.PersistEmbeddings {
		t.Fatal("must not set PersistEmbeddings (E-G4)")
	}
	if !ps.Config.TransactionalIngest {
		t.Fatal("cloud store must set TransactionalIngest")
	}
}
