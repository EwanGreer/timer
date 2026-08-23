package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EwanGreer/timer/internal/art"
	"github.com/EwanGreer/timer/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var timerName string
var detach bool

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

		if os.Getenv(detachedChildEnv) != "" {
			runDetached(duration, timerName)
			return
		}

		if detach {
			if err := spawnDetached(detachedChildArgs(os.Args[1:])); err != nil {
				log.Panicf("could not detach timer: %s", err.Error())
			}
			cmd.Println(confirmationLine(inputStr, timerName))
			return
		}

		if _, err := tea.NewProgram(commands.StartModel{Remaining: duration, Name: timerName, Art: loadArt()}, tea.WithAltScreen()).Run(); err != nil {
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

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $XDG_CONFIG_HOME/timer/config.toml)")
	rootCmd.PersistentFlags().StringVarP(&timerName, "name", "n", "", "name for the timer, shown in the completion notification")
	rootCmd.Flags().BoolVarP(&detach, "detach", "d", false, "run the timer in the background, returning the prompt immediately")
}

func getConfigPath() (string, error) {
	if cfgFile != "" {
		return cfgFile, nil
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}

	configDir := filepath.Join(configHome, "timer")
	configPath := filepath.Join(configDir, "config.toml")

	return configPath, nil
}

func ensureConfigExists(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("# Timer CLI Configuration\n# Add your configuration options here\n")
	return err
}

func getArtDir() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(configPath), "art"), nil
}

// loadArt loads the art set from the art directory, logging any warnings
// about unusable art files.
func loadArt() *art.Set {
	dir, err := getArtDir()
	if err != nil {
		log.Printf("Warning: could not determine art directory: %v", err)
		return art.Default()
	}

	s, warns := art.Load(dir)
	for _, w := range warns {
		log.Printf("Warning: %s", w)
	}
	return s
}

func initConfig() {
	configPath, err := getConfigPath()
	if err != nil {
		log.Printf("Warning: could not determine config path: %v", err)
		return
	}

	// Ensure config file exists
	if err := ensureConfigExists(configPath); err != nil {
		log.Printf("Warning: could not create config file: %v", err)
		return
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: could not read config file: %v", err)
	}
}
