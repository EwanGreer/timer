package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// timerTestOutputEnv is set on the test binary when a spawn test wants the
// re-executed child to report its env and args to a file instead of running
// tests. Only TestMain in this file knows about it.
const timerTestOutputEnv = "TIMER_TEST_CHILD_OUTPUT"

// TestMain intercepts re-executed test binaries: when a spawn test launches
// this binary as a detached child, the child reports what it received and
// exits cleanly instead of running tests.
func TestMain(m *testing.M) {
	if out := os.Getenv(timerTestOutputEnv); out != "" {
		f, err := os.Create(out)
		if err == nil {
			f.WriteString("child_env=" + os.Getenv(detachedChildEnv) + "\n")
			f.WriteString("child_args=" + strings.Join(os.Args[1:], " ") + "\n")
			f.Close()
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// resetTimerFlags clears the package-level flag vars for the duration of a
// test, since Execute parses into them and they persist between tests.
func resetTimerFlags(t *testing.T) {
	t.Helper()

	oldName, oldDetach := timerName, detach
	timerName, detach = "", false
	t.Cleanup(func() { timerName, detach = oldName, oldDetach })
}

// spawnRecord counts calls to spawnDetached.
type spawnRecord struct {
	calls int
}

// stubSpawn replaces the package spawnDetached function for the duration of
// a test.
func stubSpawn(t *testing.T) *spawnRecord {
	t.Helper()

	var rec spawnRecord
	orig := spawnDetached
	spawnDetached = func(args []string) error {
		rec.calls++
		return nil
	}
	t.Cleanup(func() { spawnDetached = orig })

	return &rec
}

func TestDetachedChildArgsStripsDetachFlags(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "short flag",
			in:   []string{"-n", "Tea", "-d", "5m"},
			want: []string{"-n", "Tea", "5m"},
		},
		{
			name: "long flag",
			in:   []string{"--detach", "10s"},
			want: []string{"10s"},
		},
		{
			name: "flag with value",
			in:   []string{"-d=true", "5m"},
			want: []string{"5m"},
		},
		{
			name: "long flag with value",
			in:   []string{"--detach=false", "5m"},
			want: []string{"5m"},
		},
		{
			name: "flag after positional",
			in:   []string{"5m", "--detach"},
			want: []string{"5m"},
		},
		{
			name: "no detach flag",
			in:   []string{"-n", "Tea", "-c", "/tmp/cfg.toml", "5m"},
			want: []string{"-n", "Tea", "-c", "/tmp/cfg.toml", "5m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detachedChildArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("detachedChildArgs(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("detachedChildArgs(%v) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

func TestDetachSpawnsChildAndPrintsNamedConfirmation(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	rec := stubSpawn(t)

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	rootCmd.SetArgs([]string{"-d", "-n", "Tea", "5m"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if rec.calls != 1 {
		t.Fatalf("spawnDetached called %d times, want 1", rec.calls)
	}
	if want := "timer \"Tea\" started — will notify on completion\n"; out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestDetachSpawnsChildAndPrintsUnnamedConfirmation(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	rec := stubSpawn(t)

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	rootCmd.SetArgs([]string{"-d", "5m"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if rec.calls != 1 {
		t.Fatalf("spawnDetached called %d times, want 1", rec.calls)
	}
	if want := "timer 5m started — will notify on completion\n"; out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestDetachedChildEnvRunsHeadlessWithoutSpawn(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(detachedChildEnv, "1")
	rec := stubSpawn(t)

	var gotD time.Duration
	var gotName string
	orig := runDetached
	runDetached = func(d time.Duration, name string) { gotD, gotName = d, name }
	t.Cleanup(func() { runDetached = orig })

	rootCmd.SetArgs([]string{"-n", "Tea", "1ms"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotD != time.Millisecond || gotName != "Tea" {
		t.Fatalf("runDetached called with (%v, %q), want (%v, %q)", gotD, gotName, time.Millisecond, "Tea")
	}
	if rec.calls != 0 {
		t.Fatalf("spawnDetached called %d times, want 0", rec.calls)
	}
}

func TestSpawnDetachedLaunchesChildWithEnvAndArgs(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "child.txt")
	t.Setenv(timerTestOutputEnv, outFile)

	if err := spawnDetached([]string{"-n", "Tea", "5m"}); err != nil {
		t.Fatalf("spawnDetached: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(outFile); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("child did not write %s within 5s", outFile)
		}
		time.Sleep(10 * time.Millisecond)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read child output: %v", err)
	}
	if want := "child_env=1\nchild_args=-n Tea 5m\n"; string(got) != want {
		t.Fatalf("child output = %q, want %q", string(got), want)
	}
}

func TestDetachFlagOnTimerCommandsButNotStopwatch(t *testing.T) {
	for _, c := range []*cobra.Command{rootCmd, startCmd} {
		if c.Flags().Lookup("detach") == nil {
			t.Errorf("%q has no --detach flag", c.Name())
		}
	}
	if stopwatchCmd.Flags().Lookup("detach") != nil {
		t.Errorf("%q has a --detach flag, want none", stopwatchCmd.Name())
	}

	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	rootCmd.SetArgs([]string{"stopwatch", "-d"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "-d") {
		t.Fatalf("stopwatch -d error = %v, want a rejection mentioning -d", err)
	}
}
