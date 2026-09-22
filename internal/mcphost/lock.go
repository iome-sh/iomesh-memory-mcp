package mcphost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// palaceLockReuseGrace is how far a live pid's start may follow the lock
// timestamp before the pid is treated as reused. The lock is written just
// after the process starts, so the holder is not after this grace.
const palaceLockReuseGrace = 2 * time.Second

// PalaceLockFile is the crash-safety writer lock under PalaceRoot (cloud mode).
// Contains pid + the wall time the lock was written. Start fails only when
// that pid is still the same process. A reused pid (container pid 1 after a
// snapshot restore) starts later than the lock time and is stale. Not flock.
const PalaceLockFile = "palace.lock"

// ErrPalaceLockHeld is returned when palace.lock names a still-alive pid.
var ErrPalaceLockHeld = errors.New("palace.lock: writer pid is alive")

type palaceLockRecord struct {
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

func acquirePalaceLock(root string) (string, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", fmt.Errorf("palace.lock: mkdir: %w", err)
	}
	path := filepath.Join(root, PalaceLockFile)
	self := palaceLockRecord{
		PID:     os.Getpid(),
		Started: time.Now().UTC().Format(time.RFC3339Nano),
	}
	for attempt := 0; attempt < 3; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			encErr := json.NewEncoder(f).Encode(self)
			syncErr := f.Sync()
			closeErr := f.Close()
			if encErr != nil || syncErr != nil || closeErr != nil {
				_ = os.Remove(path)
				if encErr != nil {
					return "", fmt.Errorf("palace.lock: write: %w", encErr)
				}
				if syncErr != nil {
					return "", fmt.Errorf("palace.lock: sync: %w", syncErr)
				}
				return "", fmt.Errorf("palace.lock: close: %w", closeErr)
			}
			return path, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("palace.lock: create: %w", err)
		}
		held, alive, readErr := readPalaceLock(path)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return "", readErr
		}
		if alive {
			return "", fmt.Errorf("%w (pid %d started %s)", ErrPalaceLockHeld, held.PID, held.Started)
		}
		if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			return "", fmt.Errorf("palace.lock: remove stale: %w", rmErr)
		}
	}
	return "", fmt.Errorf("%w: could not acquire", ErrPalaceLockHeld)
}

func readPalaceLock(path string) (palaceLockRecord, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return palaceLockRecord{}, false, err
	}
	var rec palaceLockRecord
	if json.Unmarshal(b, &rec) != nil || rec.PID <= 0 {
		return palaceLockRecord{}, false, nil
	}
	start, ok := processStartWall(rec.PID)
	return rec, palaceLockHeld(rec, pidAlive(rec.PID), start, ok), nil
}

// palaceLockHeld is false when the pid is dead or when that pid started
// after the lock was written. hasStart false keeps the old pid-alive rule.
func palaceLockHeld(rec palaceLockRecord, pidLive bool, procStart time.Time, hasStart bool) bool {
	if rec.PID <= 0 || !pidLive {
		return false
	}
	if !hasStart {
		return true
	}
	started, err := time.Parse(time.RFC3339Nano, rec.Started)
	if err != nil {
		started, err = time.Parse(time.RFC3339, rec.Started)
		if err != nil {
			return true
		}
	}
	if procStart.After(started.Add(palaceLockReuseGrace)) {
		return false
	}
	return true
}

func releasePalaceLock(path string, pid int) error {
	if path == "" || pid <= 0 {
		return nil
	}
	rec, _, err := readPalaceLock(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if rec.PID != 0 && rec.PID != pid {
		return nil
	}
	if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
		return rmErr
	}
	return nil
}
