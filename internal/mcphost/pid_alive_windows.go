//go:build windows

package mcphost

import (
	"syscall"
	"time"
)

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = syscall.CloseHandle(h)
	return true
}

func processStartWall(pid int) (time.Time, bool) {
	return time.Time{}, false
}
