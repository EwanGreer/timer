//go:build windows

package cmd

import "errors"

// spawnDetachedImpl reports that detaching is not supported on Windows, so
// the module still builds there.
func spawnDetachedImpl(args []string) error {
	return errors.New("--detach is not supported on Windows")
}
