package cmd

import (
	"fmt"
	"strings"

	"github.com/EwanGreer/timer/internal/commands"
)

// detachedChildEnv marks a re-executed process as the detached child, so it
// runs headless instead of starting the TUI.
const detachedChildEnv = "TIMER_DETACHED_CHILD"

// spawnDetached starts a copy of this binary as a detached child. It is a
// variable so tests can stub it.
var spawnDetached = spawnDetachedImpl

// runDetached is the headless timer body. It is a variable so tests can stub
// it.
var runDetached = commands.RunDetached

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
