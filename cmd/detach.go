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

const detachedChildEnv = "TIMER_DETACHED_CHILD"

var (
	spawnDetached = spawnDetachedImpl
	headlessRun   = commands.RunDetached
	runDetached   = runDetachedTimer
)

func runDetachedTimer(d time.Duration, name string) {
	logPath, err := getLogPath()
	if err == nil {
		if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			defer func() { _ = f.Close() }()
			slog.SetDefault(slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})))
		}
	}

	dir, err := getRunningDir()
	if err != nil {
		slog.Error("could not determine running dir", "err", err, "pid", os.Getpid())
		if err := headlessRun(d, name); err != nil {
			slog.Error("completion notification failed", "err", err, "pid", os.Getpid())
		}
		return
	}
	if err := registry.Write(dir, registry.Entry{Name: name, Duration: d, StartedAt: time.Now()}); err != nil {
		slog.Error("could not write registry entry", "err", err, "pid", os.Getpid())
	}
	if err := headlessRun(d, name); err != nil {
		slog.Error("completion notification failed", "err", err, "pid", os.Getpid())
	}
	if err := registry.Remove(dir, os.Getpid()); err != nil {
		slog.Error("could not remove registry entry", "err", err, "pid", os.Getpid())
	}
}

// detachedChildArgs strips the detach flag so the child does not detach again.
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

func confirmationLine(input, name string) string {
	if name != "" {
		return fmt.Sprintf("timer %q started — will notify on completion", name)
	}
	return fmt.Sprintf("timer %s started — will notify on completion", input)
}
