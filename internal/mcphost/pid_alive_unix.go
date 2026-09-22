//go:build !windows

package mcphost

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// processStartWall reads the kernel start time of pid. Linux only.
// Other unix systems return ok false and the lock stays pid-alive.
func processStartWall(pid int) (time.Time, bool) {
	if pid <= 0 {
		return time.Time{}, false
	}
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return time.Time{}, false
	}
	text := string(b)
	end := strings.LastIndex(text, ")")
	if end < 0 || end+2 >= len(text) {
		return time.Time{}, false
	}
	fields := strings.Fields(text[end+2:])
	// starttime is field 22. After "pid (comm)" the first field is state,
	// so starttime is index 19. CLK_TCK is 100 on Linux.
	if len(fields) < 20 {
		return time.Time{}, false
	}
	ticks, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil || ticks < 0 {
		return time.Time{}, false
	}
	stat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, false
	}
	var btime int64
	for _, line := range strings.Split(string(stat), "\n") {
		if !strings.HasPrefix(line, "btime ") {
			continue
		}
		btime, err = strconv.ParseInt(strings.TrimPrefix(line, "btime "), 10, 64)
		if err != nil {
			return time.Time{}, false
		}
		break
	}
	if btime <= 0 {
		return time.Time{}, false
	}
	const clkTck = 100
	sec := btime + ticks/clkTck
	nsec := (ticks % clkTck) * (int64(time.Second) / clkTck)
	return time.Unix(sec, nsec).UTC(), true
}
