package commands

import "github.com/spf13/cobra"

func Start() CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) {}
}
