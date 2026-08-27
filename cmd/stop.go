package cmd

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EwanGreer/timer/internal/registry"
	"github.com/spf13/cobra"
)

var stopTimer = registry.Stop

var stopAll bool

var stopCmd = &cobra.Command{
	Use:   "stop [id...]",
	Short: "stop running timers",
	Long: `
stop one or more detached timers before they complete.
IDs come from the ID column of "timer ps".
	`,
	Example:      "timer stop 2491",
	Args:         cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE:         stopRun,
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVar(&stopAll, "all", false, "stop every running timer")
}

func validateStopArgs(args []string, all bool) error {
	if all && len(args) > 0 {
		return errors.New("give either timer IDs or --all, not both")
	}
	if !all && len(args) == 0 {
		return errors.New("give at least one timer ID, or --all")
	}
	return nil
}

func stopByID(w io.Writer, dir string, timers []registry.Timer, args []string) error {
	byPid := make(map[int]registry.Timer, len(timers))
	for _, tm := range timers {
		byPid[tm.Pid] = tm
	}

	var errs []error
	for _, a := range args {
		pid, err := strconv.Atoi(a)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid timer ID %q", a))
			continue
		}
		tm, ok := byPid[pid]
		if !ok {
			tm, err = matchByPrefix(timers, pid)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		if err := stopTimer(dir, tm.Pid); err != nil {
			errs = append(errs, fmt.Errorf("could not stop timer %d: %w", tm.Pid, err))
			continue
		}
		fmt.Fprintln(w, stopLine(tm))
	}

	return errors.Join(errs...)
}

func matchByPrefix(timers []registry.Timer, pid int) (registry.Timer, error) {
	prefix := strconv.Itoa(pid)
	var matches []registry.Timer
	for _, tm := range timers {
		if strings.HasPrefix(strconv.Itoa(tm.Pid), prefix) {
			matches = append(matches, tm)
		}
	}
	switch len(matches) {
	case 0:
		return registry.Timer{}, fmt.Errorf("no running timer with ID %d", pid)
	case 1:
		return matches[0], nil
	}
	pids := make([]string, len(matches))
	for i, tm := range matches {
		pids[i] = strconv.Itoa(tm.Pid)
	}
	return registry.Timer{}, fmt.Errorf("timer ID %d matches %s — be more specific", pid, strings.Join(pids, ", "))
}

func stopEvery(w io.Writer, dir string, timers []registry.Timer) error {
	if len(timers) == 0 {
		fmt.Fprintln(w, "no running timers")
		return nil
	}

	var errs []error
	stopped := 0
	for _, tm := range timers {
		if err := stopTimer(dir, tm.Pid); err != nil {
			errs = append(errs, fmt.Errorf("could not stop timer %d: %w", tm.Pid, err))
			continue
		}
		stopped++
	}

	if stopped > 0 {
		noun := "timers"
		if stopped == 1 {
			noun = "timer"
		}
		fmt.Fprintf(w, "stopped %d %s\n", stopped, noun)
	}

	return errors.Join(errs...)
}

func stopLine(tm registry.Timer) string {
	if tm.Name != "" {
		return fmt.Sprintf("stopped timer %q (%d)", tm.Name, tm.Pid)
	}
	return fmt.Sprintf("stopped timer %d", tm.Pid)
}
