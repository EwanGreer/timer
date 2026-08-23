package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/EwanGreer/timer/internal/commands"
	"github.com/EwanGreer/timer/internal/registry"
)

// detachedChildEnv marks a re-executed process as the detached child, so it
// runs headless instead of starting the TUI.
const detachedChildEnv = "TIMER_DETACHED_CHILD"

// spawnDetached starts a copy of this binary as a detached child. It is a
// variable so tests can stub it.
var spawnDetached = spawnDetachedImpl

// headlessRun is the bare wait-and-notify body. It is a variable so tests
// can stub it.
var headlessRun = commands.RunDetached

// runDetached is the detached-child body. It is a variable so tests can
// stub it.
var runDetached = runDetachedTimer

// runDetachedTimer records the timer in the registry, waits it out,
// notifies, and removes the record. Errors are written to timer.log next
// to the config.
func runDetachedTimer(d time.Duration, name string) {
	logPath, err := getLogPath()
	if err == nil {
		if f, ferr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
			defer f.Close()
			slog.SetDefault(slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})))
		}
	}

	dir, err := getRunningDir()
	if err != nil {
		slog.Error("could not determine running dir", "err", err)
		if nerr := headlessRun(d, name); nerr != nil {
			slog.Error("completion notification failed", "err", nerr)
		}
		return
	}
	if werr := registry.Write(dir, registry.Entry{Name: name, Duration: d, StartedAt: time.Now()}); werr != nil {
		slog.Error("could not write registry entry", "err", werr)
	}
	if nerr := headlessRun(d, name); nerr != nil {
		slog.Error("completion notification failed", "err", nerr)
	}
	if rerr := registry.Remove(dir, os.Getpid()); rerr != nil {
		slog.Error("could not remove registry entry", "err", rerr)
	}
}

// detachedChildArgs returns the args to pass to the detached child: the
// original args without the detach flag, since the child must not detach
// again.
func detachedChildArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == "-d" || a == "--detach" || strings.HasPrefix(a, "-d=") || strings.HasPrefix(a, "--detach=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

// confirmationLine is the line printed before the prompt returns when a
// timer is detached.
func confirmationLine(input, name string) string {
	if name != "" {
		return fmt.Sprintf("timer %q started — will notify on completion", name)
	}
	return fmt.Sprintf("timer %s started — will notify on completion", input)
}
