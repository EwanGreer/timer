//go:build !darwin && !linux

package registry

import (
	"errors"
	"time"
)

func defaultProcAlive(pid int) bool { return false }

func init() { procChecksSupported = false }

func defaultProcStartedAt(pid int) (time.Time, error) {
	return time.Time{}, errors.New("process start time not supported on this platform")
}
