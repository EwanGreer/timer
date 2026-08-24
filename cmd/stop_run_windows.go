//go:build windows

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func stopRun(cmd *cobra.Command, args []string) error {
	return errors.New("stop is not supported on Windows")
}
