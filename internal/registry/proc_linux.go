//go:build linux

package registry

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func defaultProcAlive(pid int) bool {
	return unix.Kill(pid, 0) == nil
}

// defaultProcStartedAt returns the process start time from
// /proc/<pid>/stat field 22 (starttime, clock ticks since boot) plus the
// boot time from /proc/stat.
func defaultProcStartedAt(pid int) (time.Time, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return time.Time{}, err
	}
	s := string(data)
	idx := strings.LastIndexByte(s, ')')
	if idx < 0 {
		return time.Time{}, fmt.Errorf("malformed /proc/%d/stat", pid)
	}
	// fields[0] is field 3 (state); starttime is field 22 overall.
	fields := strings.Fields(s[idx+2:])
	if len(fields) < 20 {
		return time.Time{}, fmt.Errorf("malformed /proc/%d/stat: %d fields", pid, len(fields))
	}
	startTicks, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse starttime: %w", err)
	}
	ticksPerSec := clkTck()
	btime, err := readBtime()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(btime+startTicks/ticksPerSec, 0), nil
}

// clkTck returns the kernel's clock ticks per second, taken from the
// AT_CLKTCK entry of the ELF auxiliary vector — the same value libc's
// sysconf(_SC_CLK_TCK) reports. It falls back to 100 when unavailable.
func clkTck() int64 {
	const atClkTck = 17
	auxv, err := unix.Auxv()
	if err != nil {
		return 100
	}
	for _, kv := range auxv {
		if kv[0] == atClkTck {
			return int64(kv[1])
		}
	}
	return 100
}

func readBtime() (int64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "btime ") {
			return strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "btime ")), 10, 64)
		}
	}
	return 0, fmt.Errorf("btime not found in /proc/stat")
}
