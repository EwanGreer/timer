package commands

import "github.com/spf13/cobra"

func Stop() CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) {}
}
