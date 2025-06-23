/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/EwanGreer/timer-cli/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer",
	Long:  `Starts an on screen timer`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		durationStr := args[0]
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			panic(err)
		}

		if _, err := tea.NewProgram(commands.Model{Remaining: duration}, tea.WithAltScreen()).Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
