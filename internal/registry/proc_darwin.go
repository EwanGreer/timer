//go:build darwin

package registry

import (
	"time"

	"golang.org/x/sys/unix"
)

func defaultProcAlive(pid int) bool {
	return unix.Kill(pid, 0) == nil
}

func init() { procChecksSupported = true }

func defaultTerminate(pid int) error {
	return unix.Kill(pid, unix.SIGTERM)
}

func defaultProcStartedAt(pid int) (time.Time, error) {
	kp, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(kp.Proc.P_starttime.Sec, int64(kp.Proc.P_starttime.Usec)*1000), nil
}
