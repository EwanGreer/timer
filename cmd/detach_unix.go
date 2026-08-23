//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

// spawnDetachedImpl starts a copy of this binary with the given args as a
// detached child: a new session (so it survives the terminal closing) with
// stdio on the null device (so it never touches the terminal again).
func spawnDetachedImpl(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.Command(exe, args...)
	c.Env = append(os.Environ(), detachedChildEnv+"=1")
	// nil stdio connects the child to os.DevNull.
	c.Stdin, c.Stdout, c.Stderr = nil, nil, nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return c.Start()
}
