package commands

import "github.com/spf13/cobra"

type CobraCmdFunc = func(cmd *cobra.Command, args []string)

func List() CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) {}
}
