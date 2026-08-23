package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/EwanGreer/timer/internal/registry"
	"github.com/spf13/cobra"
)

// readTimers is a variable so tests can stub it.
var readTimers = registry.Read

var psCmd = &cobra.Command{
	Use:     "ps",
	Short:   "list running timers",
	Example: "timer ps",
	Args:    cobra.ExactArgs(0),
	RunE:    psRun,
}

func init() {
	rootCmd.AddCommand(psCmd)
}

func formatRemaining(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	h := int(d / time.Hour)
	m := int(d%time.Hour) / int(time.Minute)
	s := int(d%time.Minute) / int(time.Second)
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}

func renderTable(w io.Writer, timers []registry.Timer) {
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tREMAINING\tSTARTED")
	for _, tm := range timers {
		name := tm.Name
		if name == "" {
			name = "-"
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", tm.Pid, name, formatRemaining(tm.Remaining), tm.StartedAt.Format("15:04"))
	}
	tw.Flush()
}
