//go:build windows

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func psRun(cmd *cobra.Command, args []string) error {
	return errors.New("ps is not supported on Windows")
}
