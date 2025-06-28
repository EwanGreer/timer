package cmd

import (
	"time"

	"github.com/EwanGreer/timer/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var stopwatchCmd = &cobra.Command{
	Use:     "stopwatch",
	Short:   "start a stopwatch",
	Long:    `starts a stopwatch`,
	Example: "timer stopwatch",
	Aliases: []string{"sw"},
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := tea.NewProgram(commands.StopWatchModel{StartTime: time.Now()}, tea.WithAltScreen()).Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopwatchCmd)
}
