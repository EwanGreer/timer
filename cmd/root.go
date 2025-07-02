/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/EwanGreer/timer/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "timer [duration|deadline]",
	Short: "start a timer",
	Long: `
start a timer of a set duration of hours, minutes and seconds.
Provide a duration or a deadline
	`,
	Example: "timer start (10h30m10s|10h30m|10h|30m|10s)",
	Args:    cobra.ExactArgs(1),
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.timer.toml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".timer.toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
