package mcphost

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// walPendingRelDir matches kernel github.com/iome-sh/memory ingest_tx.go
// (unexported walPendingRelDir). Duplicate the relative path; do not import
// unexported kernel symbols.
const walPendingRelDir = "wal/pending"

const readyProbePrefix = ".ready-probe-"

// ReadyResponse is the JSON body for GET /ready (Cloud Memory writer readiness).
// Distinct from HealthzResponse: process-up (/healthz) is not writer-ready (/ready).
// dual_write OFF · not Memory GA · not tools/list · not ingest · not Fetch.
type ReadyResponse struct {
	Status      string `json:"status"` // "ok" | "not_ready"
	Service     string `json:"service"`
	DualWrite   string `json:"dual_write"`
	NotMemoryGA bool   `json:"not_memory_ga"`
	// PalaceWritable is true when a temp file can be created and removed under
	// the probe store base (fail closed).
	PalaceWritable bool `json:"palace_writable"`
	// WALPending is the number of directory entries in <storeBase>/wal/pending.
	WALPending int `json:"wal_pending"`
	// WALPendingRecoverable is true when the pending dir is missing/empty or
	// contains only regular files (kernel recover-on-open leftover is ok).
	WALPendingRecoverable bool `json:"wal_pending_recoverable"`
	// LastIngestUnix is best-effort mtime unix of the latest regular dest under
	// the store (skips wal/pending). 0 if none.
	LastIngestUnix int64 `json:"last_ingest_unix"`
	// RSSBytes is runtime.MemStats.Alloc (heap bytes in use), not OS RSS / Sys.
	RSSBytes uint64 `json:"rss_bytes"`
	// CgroupMemoryBytes is cgroup v2 memory.current or v1 memory.usage_in_bytes.
	// 0 if the cgroup file is unreadable (laptop).
	CgroupMemoryBytes uint64 `json:"cgroup_memory_bytes"`
}

// cloudMemoryGAEnv is true only when MEMORY_CLOUD_GA is exactly "1".
// Unset, empty, "true", "yes", and "0" are not enough (no trim, no case fold).
// OSS local memory and the dogfood image stay not-GA until a process is marked.
func cloudMemoryGAEnv() bool {
	return os.Getenv("MEMORY_CLOUD_GA") == "1"
}

// ReadySnapshot is the GET /ready body. HTTP 200 iff palace writable and
// wal_pending_recoverable; otherwise status=not_ready (handler returns 503).
// Nil host or empty palace root fail closed. Honesty fields match healthz
// (dual_write off) without growing HealthzResponse. not_memory_ga stays true
// unless MEMORY_CLOUD_GA is exactly "1".
func ReadySnapshot(host *Host) ReadyResponse {
	body := ReadyResponse{
		Status:                "not_ready",
		Service:               ServerName,
		DualWrite:             "off",
		NotMemoryGA:           !cloudMemoryGAEnv(),
		PalaceWritable:        false,
		WALPendingRecoverable: false,
		RSSBytes:              rssAllocBytes(),
		CgroupMemoryBytes:     cgroupMemoryBytes(),
	}
	base := readyStoreBase(host)
	if base == "" {
		return body
	}
	writable := palaceWritable(base)
	pending, recoverable := inspectWALPending(base)
	body.PalaceWritable = writable
	body.WALPending = pending
	body.WALPendingRecoverable = recoverable
	body.LastIngestUnix = lastIngestUnix(base)
	if writable && recoverable {
		body.Status = "ok"
	}
	return body
}

// ReadyHandler returns 200 JSON when the writer palace is ready, else 503.
// Always unauthenticated (probes must work without the MCP shared secret).
// Do not wrap with WithOptionalSharedSecret. GET /healthz stays a separate
// process-up probe and does not include these fields.
func ReadyHandler(hosts ...*Host) http.HandlerFunc {
	var host *Host
	if len(hosts) > 0 {
		host = hosts[0]
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body := ReadySnapshot(host)
		code := http.StatusOK
		if body.Status != "ok" {
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if r.Method == http.MethodHead {
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	}
}

// readyStoreBase is the directory probed for writability, wal/pending, and
// last-ingest mtime.
//
// Cloud: the single store BaseDir (PALACE_ROOT/<configured tenant>).
// Local-dev: configured default tenant if set, else PalaceRoot itself
// (do not invent a "default" tenant write).
func readyStoreBase(h *Host) string {
	if h == nil {
		return ""
	}
	root := strings.TrimSpace(h.cfg.PalaceRoot)
	if root == "" {
		return ""
	}
	if t := strings.TrimSpace(h.cfg.DefaultTenant); t != "" {
		return filepath.Join(root, t)
	}
	return root
}

func palaceWritable(dir string) bool {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return false
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false
	}
	f, err := os.CreateTemp(dir, readyProbePrefix+"*")
	if err != nil {
		return false
	}
	name := f.Name()
	_, writeErr := f.Write([]byte{0})
	closeErr := f.Close()
	rmErr := os.Remove(name)
	return writeErr == nil && closeErr == nil && rmErr == nil
}

func inspectWALPending(storeBase string) (count int, recoverable bool) {
	if strings.TrimSpace(storeBase) == "" {
		return 0, false
	}
	dir := filepath.Join(storeBase, filepath.FromSlash(walPendingRelDir))
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, true
		}
		return 0, false
	}
	if len(ents) == 0 {
		return 0, true
	}
	ok := true
	for _, e := range ents {
		count++
		info, infoErr := e.Info()
		if infoErr != nil || !info.Mode().IsRegular() {
			ok = false
		}
	}
	return count, ok
}

func lastIngestUnix(storeBase string) int64 {
	storeBase = strings.TrimSpace(storeBase)
	if storeBase == "" {
		return 0
	}
	pending := filepath.Join(storeBase, filepath.FromSlash(walPendingRelDir))
	var latest int64
	_ = filepath.WalkDir(storeBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == storeBase {
			return nil
		}
		if path == pending || strings.HasPrefix(path, pending+string(os.PathSeparator)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), readyProbePrefix) {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil || !info.Mode().IsRegular() {
			return nil
		}
		if t := info.ModTime().Unix(); t > latest {
			latest = t
		}
		return nil
	})
	return latest
}

func rssAllocBytes() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Alloc
}

func cgroupMemoryBytes() uint64 {
	for _, p := range []string{
		"/sys/fs/cgroup/memory.current",
		"/sys/fs/cgroup/memory/memory.usage_in_bytes",
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		n, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
		if err != nil {
			continue
		}
		return n
	}
	return 0
}
