package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func psRun(cmd *cobra.Command, args []string) error {
	dir, err := getRunningDir()
	if err != nil {
		return fmt.Errorf("could not determine running dir: %w", err)
	}
	timers, err := readTimers(dir)
	if err != nil {
		return fmt.Errorf("could not list timers: %w", err)
	}
	renderTable(cmd.OutOrStdout(), timers)
	return nil
}
