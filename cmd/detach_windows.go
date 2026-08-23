//go:build windows

package cmd

import "errors"

// This stub keeps the module building on Windows.
func spawnDetachedImpl(args []string) error {
	return errors.New("--detach is not supported on Windows")
}
