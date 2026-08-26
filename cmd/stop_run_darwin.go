package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func stopRun(cmd *cobra.Command, args []string) error {
	if err := validateStopArgs(args, stopAll); err != nil {
		return err
	}
	dir, err := getRunningDir()
	if err != nil {
		return fmt.Errorf("could not determine running dir: %w", err)
	}
	timers, err := readTimers(dir)
	if err != nil {
		return fmt.Errorf("could not list timers: %w", err)
	}
	if stopAll {
		return stopEvery(cmd.OutOrStdout(), dir, timers)
	}
	return stopByID(cmd.OutOrStdout(), dir, timers, args)
}
