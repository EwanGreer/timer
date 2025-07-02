/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/EwanGreer/timer/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "timer [duration|deadline]",
	Short: "Start a timer",
	Long: `
start a timer of a set duration of hours, minutes and seconds.
Provide a duration or a deadline
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputStr := args[0]
		var duration time.Duration

		switch {
		case strings.Contains(inputStr, ":"):
			t, err := time.Parse("15:04", inputStr)
			if err != nil {
				log.Panicf("could not parse 24 hour time: %s", err.Error())
			}

			now := time.Now()
			targetTime := time.Date(
				now.Year(), now.Month(), now.Day(),
				t.Hour(), t.Minute(), 0, 0, now.Location(),
			)

			if targetTime.Before(now) {
				targetTime = targetTime.Add(24 * time.Hour)
			}

			duration = time.Until(targetTime)
		default:
			d, err := time.ParseDuration(inputStr)
			if err != nil {
				log.Panicf("could not parse duration time: %s", err.Error())
			}
			duration = d
		}

		if _, err := tea.NewProgram(commands.StartModel{Remaining: duration}, tea.WithAltScreen()).Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
