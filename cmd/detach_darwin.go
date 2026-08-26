package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

// Setsid gives the child a new session so it survives the terminal
// closing; nil stdio connects it to os.DevNull so it never touches the
// terminal again.
func spawnDetachedImpl(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.Command(exe, args...)
	c.Env = append(os.Environ(), detachedChildEnv+"=1")
	c.Stdin, c.Stdout, c.Stderr = nil, nil, nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return c.Start()
}
