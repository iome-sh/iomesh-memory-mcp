package mcphost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PalaceLockFile is the crash-safety writer lock under PalaceRoot (cloud mode).
// Contains pid + start time. Start fails if that pid is still alive. Not flock.
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
	return rec, pidAlive(rec.PID), nil
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
