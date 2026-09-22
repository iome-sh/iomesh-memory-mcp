package mcphost

import (
	"os"
	"testing"
	"time"
)

func TestPalaceLockHeldWhenPIDReusedAfterStart(t *testing.T) {
	rec := palaceLockRecord{PID: 1, Started: "2020-01-01T00:00:00Z"}
	proc := time.Date(2026, 9, 22, 20, 32, 0, 0, time.UTC)
	if palaceLockHeld(rec, true, proc, true) {
		t.Fatal("a pid that started after the lock was written must not hold it")
	}
}

func TestPalaceLockHeldWhileSameProcess(t *testing.T) {
	rec := palaceLockRecord{PID: os.Getpid(), Started: time.Now().UTC().Format(time.RFC3339Nano)}
	proc := time.Now().UTC().Add(-time.Second)
	if !palaceLockHeld(rec, true, proc, true) {
		t.Fatal("the process that wrote the lock must still hold it")
	}
}

func TestPalaceLockHeldWithoutStartTimeFallsBack(t *testing.T) {
	rec := palaceLockRecord{PID: 1, Started: "not-a-time"}
	if !palaceLockHeld(rec, true, time.Time{}, false) {
		t.Fatal("an unknown start time stays held while the pid is alive")
	}
	if palaceLockHeld(rec, false, time.Time{}, false) {
		t.Fatal("a dead pid must not hold the lock")
	}
}

func TestProcessStartWallWhenProcExists(t *testing.T) {
	if _, err := os.Stat("/proc/self/stat"); err != nil {
		if _, ok := processStartWall(os.Getpid()); ok {
			t.Fatal("start time without /proc")
		}
		return
	}
	got, ok := processStartWall(os.Getpid())
	if !ok {
		t.Fatal("expected a start time from /proc")
	}
	if got.After(time.Now().Add(2 * time.Second)) {
		t.Fatalf("start time in the future: %s", got)
	}
	if time.Since(got) > 48*time.Hour {
		t.Fatalf("start time too old: %s", got)
	}
}
