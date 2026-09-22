# Detached Timer Robustness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `timer -d` honest: the confirmation prints only after the
child proves it started, the child runs exactly the duration the parent
computed, and the notification fires on wall-clock time.

**Architecture:** The parent computes the duration once and passes it to
the child through environment variables, so the child never re-parses
user input and the parent builds the child argv itself (no syntactic flag
stripping). Spawn returns the child pid and the parent polls the registry
for `<pid>.json` before confirming. The child waits in short wall-clock
ticks. The registry gains nanosecond precision with a seconds fallback
for old files. Root and `start` share one Run body under the correct
subcommand name.

**Tech Stack:** Go 1.24.3 (per go.mod), cobra, slog, beeep, stdlib
`encoding/json`, `os/exec`, `syscall`.

**Spec:** `docs/superpowers/specs/2026-09-21-detached-timer-robustness.md`

**Note:** This plan applies on top of the current working tree, which
renames `detachedChildEnv` to `detachedChildEnvKey` and
`spawnDetachedImpl` to `spawnDetachedTimer`. Commit that WIP (or leave it
staged) before starting Task 1; the task code below uses the new names.

## Global Constraints

- Child state travels only through the environment
  (`TIMER_DETACHED_DURATION_NS`, `TIMER_DETACHED_NAME`); the child never
  re-parses the user's duration or deadline.
- `TIMER_DETACHED_CHILD` counts as set only when it equals exactly `1`.
  All `TIMER_DETACHED_*` keys are stripped from the inherited environment
  before the payload is appended.
- Registry files stay JSON named `<pid>.json`. `duration_ns` (integer
  nanoseconds) is added; `duration_seconds` (truncated) is still written
  so older binaries keep reading new files.
- Test seams are package vars (`spawnDetached`, `runDetached`,
  `headlessRun`, `now`, `maxTick`, `handshakeTimeout`), not interfaces.
- The confirmation prints only after the registry entry exists; otherwise
  the parent prints a failure line naming the timer log.
- The child polls the wall clock at most once per second while waiting.
- Commits use `feat:` / `test:` prefixes and end with
  `Co-Authored-By: Claude Code <noreply@anthropic.com>`.
- Run tests with `go test ./...` (mise has no test task).

---

### Task 1: Name the start subcommand correctly and share one Run body

**Files:**
- Modify: `cmd/start.go` — `Use: "start [duration|deadline]"`, move the
  shared body into `runStart`
- Modify: `cmd/root.go` — `Run: runStart`, drop the duplicated body and
  its imports
- Test: `cmd/start_test.go` (create)

**Interfaces:**
- Consumes: `detachedChildEnvKey`, `spawnDetached`, `detachedChildArgs`,
  `runDetached`, `confirmationLine`, `loadArt` — all already in package
  `cmd`.
- Produces:
  - `runStart(cmd *cobra.Command, args []string)` — the single timer Run
    body used by both `rootCmd` and `startCmd`; Tasks 3 and 4 modify this
    one function.
  - `startCmd` with cobra name `start`, so `timer start 5m` resolves.

- [ ] **Step 1: Write the failing tests**

Create `cmd/start_test.go`:

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestStartCommandName(t *testing.T) {
	if got := startCmd.Name(); got != "start" {
		t.Fatalf("startCmd.Name() = %q, want %q", got, "start")
	}
}

func TestStartSubcommandRunsSameDetachPathAsRoot(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	rec := stubSpawn(t)

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	rootCmd.SetArgs([]string{"start", "-d", "5m"})
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
```

`resetTimerFlags` and `stubSpawn` already exist in `cmd/detach_test.go`
(same package).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/ -run 'TestStart' -v`
Expected: FAIL — `TestStartCommandName`: `startCmd.Name() = "timer"`;
`TestStartSubcommandRunsSameDetachPathAsRoot`: `Execute` returns
`unknown command "start" for "timer"`.

- [ ] **Step 3: Write minimal implementation**

In `cmd/start.go`, replace the `Use` line and the `Run` closure, and add
`runStart`:

```go
var startCmd = &cobra.Command{
	Use:   "start [duration|deadline]",
	Short: "Start a timer",
	Long: `
start a timer of a set duration of hours, minutes and seconds.
Provide a duration or a deadline
	`,
	Args: cobra.ExactArgs(1),
	Run:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&detach, "detach", "d", false, "run the timer in the background, returning the prompt immediately")
}

// runStart parses the duration and runs the timer: headless when the
// process is a detached child, spawned when -d is set, TUI otherwise.
func runStart(cmd *cobra.Command, args []string) {
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

	if os.Getenv(detachedChildEnvKey) != "" {
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
}
```

The full `cmd/start.go` after this step:

```go
package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/EwanGreer/timer/internal/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [duration|deadline]",
	Short: "Start a timer",
	Long: `
start a timer of a set duration of hours, minutes and seconds.
Provide a duration or a deadline
	`,
	Args: cobra.ExactArgs(1),
	Run:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&detach, "detach", "d", false, "run the timer in the background, returning the prompt immediately")
}

// runStart parses the duration and runs the timer: headless when the
// process is a detached child, spawned when -d is set, TUI otherwise.
func runStart(cmd *cobra.Command, args []string) {
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

	if os.Getenv(detachedChildEnvKey) != "" {
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
}
```

In `cmd/root.go`, replace the `Run` closure with `Run: runStart` and
remove the now-unused imports (`strings`, `time`, `tea`,
`internal/commands`). The imports become:

```go
import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/EwanGreer/timer/internal/art"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)
```

(`log`, `os`, `path/filepath`, `art` stay: `loadArt` and `initConfig`
still use them.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/ -v`
Expected: PASS (the two new tests plus the existing detach and ps tests;
the existing detach tests exercise the root path, which now runs
`runStart`).

- [ ] **Step 5: Run the full suite**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet warnings.

- [ ] **Step 6: Commit**

```bash
git add cmd/start.go cmd/root.go cmd/start_test.go
git commit -m "feat: name the start subcommand start and share one run body

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 2: Store nanosecond durations in the registry with a seconds fallback

**Files:**
- Modify: `internal/registry/registry.go` — `record`, `Write`, `Read`
- Test: `internal/registry/registry_test.go` (append)

**Interfaces:**
- Consumes: nothing new (`record`, `Write`, `Read` from the existing
  registry code).
- Produces:
  - `record` gains `DurationNS int64 \`json:"duration_ns"\``
  - `recordDuration(rec record) time.Duration` — prefers `DurationNS`,
    falls back to `DurationS` seconds for files written by older
    binaries.

- [ ] **Step 1: Write the failing tests**

Append to `internal/registry/registry_test.go` (all imports it needs —
`encoding/json`, `fmt`, `os`, `path/filepath`, `testing`, `time` — are
already there):

```go
func TestWriteStoresNanosecondDuration(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, Entry{Name: "Tea", Duration: 1500 * time.Millisecond, StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%d.json", os.Getpid())))
	if err != nil {
		t.Fatal(err)
	}
	var rec record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.DurationNS != int64(1500*time.Millisecond) {
		t.Fatalf("duration_ns = %d, want %d", rec.DurationNS, int64(1500*time.Millisecond))
	}
	if rec.DurationS != 1 {
		t.Fatalf("duration_seconds = %d, want 1 (truncated)", rec.DurationS)
	}
}

func TestReadFallsBackToSecondsOnlyFiles(t *testing.T) {
	dir := t.TempDir()
	started := time.Now().Add(-10 * time.Second)
	b, err := json.Marshal(record{Name: "Old", DurationS: 300, StartedAt: started})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "4242.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	stubProc(t, true, started, nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 1 {
		t.Fatalf("len = %d, want 1", len(timers))
	}
	if got := timers[0].Duration; got != 5*time.Minute {
		t.Fatalf("duration = %v, want 5m", got)
	}
	if timers[0].Remaining > 295*time.Second || timers[0].Remaining < 285*time.Second {
		t.Fatalf("remaining = %v, want about 290s", timers[0].Remaining)
	}
}

func TestReadReturnsSubSecondRemaining(t *testing.T) {
	dir := t.TempDir()
	started := time.Now().Add(-time.Second)
	if err := Write(dir, Entry{Duration: 1500 * time.Millisecond, StartedAt: started}); err != nil {
		t.Fatal(err)
	}
	stubProc(t, true, started, nil)

	timers, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(timers) != 1 {
		t.Fatalf("len = %d, want 1", len(timers))
	}
	got := timers[0].Remaining
	if got > 600*time.Millisecond || got < 400*time.Millisecond {
		t.Fatalf("remaining = %v, want about 500ms", got)
	}
}
```

`stubProc` already exists in `internal/registry/registry_test.go`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/registry/ -run 'TestWriteStores|TestReadFallsBack|TestReadReturnsSubSecond' -v`
Expected: FAIL — `rec.DurationNS undefined`, and `TestReadFallsBackToSecondsOnlyFiles` /
`TestReadReturnsSubSecondRemaining` report the truncated duration
(`5m0s` / `1s` instead of `5m` / about `500ms`).

- [ ] **Step 3: Write minimal implementation**

In `internal/registry/registry.go`, extend `record`:

```go
type record struct {
	Name       string    `json:"name"`
	DurationS  int64     `json:"duration_seconds"`
	DurationNS int64     `json:"duration_ns"`
	StartedAt  time.Time `json:"started_at"`
}
```

In `Write`, set both fields:

```go
	rec := record{Name: e.Name, DurationS: int64(e.Duration / time.Second), DurationNS: int64(e.Duration), StartedAt: e.StartedAt}
```

In `Read`, replace the remaining/Duration computation:

```go
			remaining := rec.StartedAt.Add(time.Duration(rec.DurationS) * time.Second).Sub(now)
			if remaining < 0 {
				remaining = 0
			}
			timers = append(timers, Timer{
				Pid:       pid,
				Name:      rec.Name,
				Duration:  time.Duration(rec.DurationS) * time.Second,
				StartedAt: rec.StartedAt,
				Remaining: remaining,
			})
```

with:

```go
			dur := recordDuration(rec)
			remaining := rec.StartedAt.Add(dur).Sub(now)
			if remaining < 0 {
				remaining = 0
			}
			timers = append(timers, Timer{
				Pid:       pid,
				Name:      rec.Name,
				Duration:  dur,
				StartedAt: rec.StartedAt,
				Remaining: remaining,
			})
```

and add the helper after `readRecord`:

```go
// recordDuration prefers the nanosecond field written by current
// binaries and falls back to duration_seconds for files written by older
// ones.
func recordDuration(rec record) time.Duration {
	if rec.DurationNS != 0 {
		return time.Duration(rec.DurationNS)
	}
	return time.Duration(rec.DurationS) * time.Second
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/registry/ -v`
Expected: PASS. The pre-existing tests keep passing: `record` literals
with only `DurationS` (e.g. `TestReadSortsOldestFirst`) go through the
fallback, and `TestWriteCreatesFileWithRecord` still sees
`DurationS == 300`.

- [ ] **Step 5: Run the full suite**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet warnings.

- [ ] **Step 6: Commit**

```bash
git add internal/registry/registry.go internal/registry/registry_test.go
git commit -m "feat: store nanosecond durations in the registry

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 3: Pass the computed payload through the environment

**Files:**
- Modify: `cmd/detach.go` — env constants, `detachedChildArgv`,
  `detachedEnv`, `isDetachedChild`, `detachedDuration`; delete
  `detachedChildArgs`
- Modify: `cmd/detach_darwin.go` — `spawnDetachedTimer(args, env []string) error`
- Modify: `cmd/start.go` — `runStart` child branch and spawn call
- Test: `cmd/detach_test.go` (rewrite the affected tests, delete the
  stripper test)

**Interfaces:**
- Consumes: `runStart` (Task 1), `runDetached`, `cfgFile`, `timerName`
  package vars.
- Produces:
  - `detachedChildArgv(cfgFile string) []string` — child argv
  - `detachedEnv(parent []string, d time.Duration, name string) []string` — child env
  - `isDetachedChild() bool` — marker equals exactly `1`
  - `detachedDuration() (time.Duration, error)` — payload parse
  - `spawnDetached func(args []string, env []string) error` (package var,
    updated signature)

- [ ] **Step 1: Write the failing tests**

In `cmd/detach_test.go`:

Replace `stubSpawn` and its `spawnRecord` with:

```go
type spawnRecord struct {
	calls int
	args  []string
	env   []string
}

func stubSpawn(t *testing.T) *spawnRecord {
	t.Helper()

	var rec spawnRecord
	orig := spawnDetached
	spawnDetached = func(args []string, env []string) error {
		rec.calls++
		rec.args = args
		rec.env = env
		return nil
	}
	t.Cleanup(func() { spawnDetached = orig })

	return &rec
}
```

Delete `TestDetachedChildArgsStripsDetachFlags` entirely.

Replace the body of `TestDetachSpawnsChildAndPrintsNamedConfirmation`
with (added assertions on child args and env):

```go
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
	if len(rec.args) != 1 || rec.args[0] != "0s" {
		t.Fatalf("child args = %v, want [0s]", rec.args)
	}
	wantEnv := []string{
		detachedChildEnvKey + "=1",
		detachedDurationEnvKey + "=300000000000",
		detachedNameEnvKey + "=Tea",
	}
	for _, want := range wantEnv {
		if !contains(rec.env, want) {
			t.Fatalf("child env missing %q: %v", want, rec.env)
		}
	}
	if want := "timer \"Tea\" started — will notify on completion\n"; out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}
```

Add the `contains` helper at the bottom of the file:

```go
func contains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}
```

Replace the body of `TestDetachedChildEnvRunsHeadlessWithoutSpawn`
with (payload now comes from the environment, not the args):

```go
func TestDetachedChildEnvRunsHeadlessWithoutSpawn(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(detachedChildEnvKey, "1")
	t.Setenv(detachedDurationEnvKey, "1000000")
	t.Setenv(detachedNameEnvKey, "Tea")
	rec := stubSpawn(t)

	var gotD time.Duration
	var gotName string
	orig := runDetached
	runDetached = func(d time.Duration, name string) { gotD, gotName = d, name }
	t.Cleanup(func() { runDetached = orig })

	rootCmd.SetArgs([]string{"0s"})
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
```

Append the new tests:

```go
func TestDetachedChildMarkerRequiresOne(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(detachedChildEnvKey, "0")
	rec := stubSpawn(t)

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	rootCmd.SetArgs([]string{"-d", "5m"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if rec.calls != 1 {
		t.Fatalf("spawnDetached called %d times, want 1 (marker \"0\" is not a child)", rec.calls)
	}
}

func TestDetachedChildWithInvalidPayloadDoesNotRunTimer(t *testing.T) {
	resetTimerFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(detachedChildEnvKey, "1")
	t.Setenv(detachedDurationEnvKey, "not-a-number")
	rec := stubSpawn(t)

	var ran bool
	orig := runDetached
	runDetached = func(d time.Duration, name string) { ran = true }
	t.Cleanup(func() { runDetached = orig })

	rootCmd.SetArgs([]string{"0s"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if ran {
		t.Fatal("runDetached ran, want no timer for an invalid payload")
	}
	if rec.calls != 0 {
		t.Fatalf("spawnDetached called %d times, want 0", rec.calls)
	}
}

func TestDetachedEnvFiltersStaleEntries(t *testing.T) {
	env := detachedEnv([]string{
		"PATH=/usr/bin",
		detachedChildEnvKey + "=0",
		detachedDurationEnvKey + "=999",
	}, time.Minute, "Tea")

	if contains(env, detachedChildEnvKey+"=0") || contains(env, detachedDurationEnvKey+"=999") {
		t.Fatalf("stale TIMER_DETACHED_* entries survived: %v", env)
	}
	want := []string{
		"PATH=/usr/bin",
		detachedChildEnvKey + "=1",
		detachedDurationEnvKey + "=60000000000",
		detachedNameEnvKey + "=Tea",
	}
	for _, w := range want {
		if !contains(env, w) {
			t.Fatalf("env missing %q: %v", w, env)
		}
	}
}

func TestDetachedChildArgv(t *testing.T) {
	if got := detachedChildArgv(""); len(got) != 1 || got[0] != "0s" {
		t.Fatalf("detachedChildArgv(\"\") = %v, want [0s]", got)
	}
	want := []string{"0s", "-c", "/tmp/custom.toml"}
	got := detachedChildArgv("/tmp/custom.toml")
	if len(got) != len(want) {
		t.Fatalf("detachedChildArgv = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("detachedChildArgv = %v, want %v", got, want)
		}
	}
}
```

Update `TestMain` to report the payload keys:

```go
func TestMain(m *testing.M) {
	if out := os.Getenv(timerTestOutputEnv); out != "" {
		f, err := os.Create(out)
		if err == nil {
			f.WriteString("child_env=" + os.Getenv(detachedChildEnvKey) + "\n")
			f.WriteString("child_duration_ns=" + os.Getenv(detachedDurationEnvKey) + "\n")
			f.WriteString("child_name=" + os.Getenv(detachedNameEnvKey) + "\n")
			f.WriteString("child_args=" + strings.Join(os.Args[1:], " ") + "\n")
			f.Close()
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
```

Replace `TestSpawnDetachedLaunchesChildWithEnvAndArgs` with:

```go
func TestSpawnDetachedLaunchesChildWithEnvAndArgs(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "child.txt")
	t.Setenv(timerTestOutputEnv, outFile)

	if err := spawnDetachedTimer(detachedChildArgv(""), detachedEnv(os.Environ(), time.Minute, "Tea")); err != nil {
		t.Fatalf("spawnDetachedTimer: %v", err)
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
	want := "child_env=1\nchild_duration_ns=60000000000\nchild_name=Tea\nchild_args=0s\n"
	if string(got) != want {
		t.Fatalf("child output = %q, want %q", string(got), want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/ -v`
Expected: compile FAIL — `undefined: detachedDurationEnvKey`,
`undefined: detachedChildArgv`, `spawnDetachedTimer` has the wrong
signature, `spawnDetached` var type mismatch.

- [ ] **Step 3: Write minimal implementation**

In `cmd/detach.go`, replace the constant block with:

```go
const detachedEnvPrefix = "TIMER_DETACHED_"

const detachedChildEnvKey = detachedEnvPrefix + "CHILD"
const detachedDurationEnvKey = detachedEnvPrefix + "DURATION_NS"
const detachedNameEnvKey = detachedEnvPrefix + "NAME"
```

Delete `detachedChildArgs` and add in its place:

```go
// detachedChildArgv builds the child command line. The duration and name
// travel through the environment, so the only positional argument is a
// placeholder that satisfies ExactArgs(1); the child ignores it.
func detachedChildArgv(cfgFile string) []string {
	args := []string{"0s"}
	if cfgFile != "" {
		args = append(args, "-c", cfgFile)
	}
	return args
}

// detachedEnv builds the child environment: any TIMER_DETACHED_* entries
// are dropped (a stale marker must not leak in) and the computed payload
// is appended.
func detachedEnv(parent []string, d time.Duration, name string) []string {
	env := make([]string, 0, len(parent)+3)
	for _, kv := range parent {
		if strings.HasPrefix(kv, detachedEnvPrefix) {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		detachedChildEnvKey+"=1",
		detachedDurationEnvKey+"="+strconv.FormatInt(int64(d), 10),
		detachedNameEnvKey+"="+name,
	)
	return env
}

func isDetachedChild() bool {
	return os.Getenv(detachedChildEnvKey) == "1"
}

func detachedDuration() (time.Duration, error) {
	ns, err := strconv.ParseInt(os.Getenv(detachedDurationEnvKey), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", detachedDurationEnvKey, err)
	}
	return time.Duration(ns), nil
}
```

Add `strconv` to `cmd/detach.go` imports (it already imports `fmt`,
`log/slog`, `os`, `strings`, `time`, `internal/commands`,
`internal/registry`).

In `cmd/detach_darwin.go`, replace `spawnDetachedTimer` with:

```go
// Setsid gives the child a new session so it survives the terminal
// closing; nil stdio connects the descriptors to os.DevNull so they
// never touch the terminal again. The caller supplies args and env: the
// child argv and the detached payload built in cmd/detach.go.
func spawnDetachedTimer(args []string, env []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	c := exec.Command(exe, args...)
	c.Env = env
	c.Stdin, c.Stdout, c.Stderr = nil, nil, nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	return c.Start()
}
```

In `cmd/start.go`, `runStart`, replace the child branch and the spawn
call:

```go
	if isDetachedChild() {
		d, err := detachedDuration()
		if err != nil {
			slog.Error("invalid detached child payload", "err", err, "pid", os.Getpid())
			return
		}
		runDetached(d, os.Getenv(detachedNameEnvKey))
		return
	}

	if detach {
		if err := spawnDetached(detachedChildArgv(cfgFile), detachedEnv(os.Environ(), duration, timerName)); err != nil {
			log.Panicf("could not detach timer: %s", err.Error())
		}
		cmd.Println(confirmationLine(inputStr, timerName))
		return
	}
```

Add `log/slog` to the `cmd/start.go` imports.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/ -v`
Expected: PASS. `TestDetachSpawnsChildAndPrintsUnnamedConfirmation` still
passes unchanged (it only counts spawn calls and checks the output line).

- [ ] **Step 5: Run the full suite**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet warnings.

- [ ] **Step 6: Commit**

```bash
git add cmd/detach.go cmd/detach_darwin.go cmd/start.go cmd/detach_test.go
git commit -m "feat: pass the computed duration and name to the detached child via env

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 4: Verify the child started before printing the confirmation

**Files:**
- Modify: `cmd/detach_darwin.go` — `spawnDetachedTimer` returns
  `(int, error)` (child pid)
- Modify: `cmd/detach.go` — `handshakeTimeout`, `waitForRegistryEntry`,
  `logPathHint`, `confirmationAfterHandshake`
- Modify: `cmd/start.go` — `runStart` prints
  `confirmationAfterHandshake` instead of `confirmationLine`
- Test: `cmd/detach_test.go` (update spawn stub and confirmation tests,
  append handshake tests)

**Interfaces:**
- Consumes: `spawnDetached` (Task 3 signature, now returning a pid),
  `getRunningDir`, `getLogPath`, `confirmationLine` (Task 1).
- Produces:
  - `spawnDetached func(args []string, env []string) (int, error)`
  - `waitForRegistryEntry(dir string, pid int, timeout time.Duration) bool`
  - `confirmationAfterHandshake(input, name string, pid int) string`

- [ ] **Step 1: Write the failing tests**

In `cmd/detach_test.go`, replace `spawnRecord` and `stubSpawn` with:

```go
type spawnRecord struct {
	calls int
	pid   int
	args  []string
	env   []string
}

func stubSpawn(t *testing.T) *spawnRecord {
	t.Helper()

	var rec spawnRecord
	orig := spawnDetached
	spawnDetached = func(args []string, env []string) (int, error) {
		rec.calls++
		rec.args = args
		rec.env = env
		return rec.pid, nil
	}
	t.Cleanup(func() { spawnDetached = orig })

	rec.pid = 4242
	return &rec
}
```

Replace `TestDetachSpawnsChildAndPrintsNamedConfirmation` with a version
that plants the registry file the handshake waits for:

```go
func TestDetachSpawnsChildAndPrintsNamedConfirmation(t *testing.T) {
	resetTimerFlags(t)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	runningDir := filepath.Join(configHome, "timer", "running")
	if err := os.MkdirAll(runningDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runningDir, "4242.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
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
	if len(rec.args) != 1 || rec.args[0] != "0s" {
		t.Fatalf("child args = %v, want [0s]", rec.args)
	}
	if want := "timer \"Tea\" started — will notify on completion\n"; out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}
```

Replace `TestDetachSpawnsChildAndPrintsUnnamedConfirmation` the same
way:

```go
func TestDetachSpawnsChildAndPrintsUnnamedConfirmation(t *testing.T) {
	resetTimerFlags(t)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	runningDir := filepath.Join(configHome, "timer", "running")
	if err := os.MkdirAll(runningDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runningDir, "4242.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
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
```

Update `TestSpawnDetachedLaunchesChildWithEnvAndArgs` for the new
signature:

```go
func TestSpawnDetachedLaunchesChildWithEnvAndArgs(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "child.txt")
	t.Setenv(timerTestOutputEnv, outFile)

	pid, err := spawnDetachedTimer(detachedChildArgv(""), detachedEnv(os.Environ(), time.Minute, "Tea"))
	if err != nil {
		t.Fatalf("spawnDetachedTimer: %v", err)
	}
	if pid <= 0 {
		t.Fatalf("pid = %d, want positive", pid)
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
	want := "child_env=1\nchild_duration_ns=60000000000\nchild_name=Tea\nchild_args=0s\n"
	if string(got) != want {
		t.Fatalf("child output = %q, want %q", string(got), want)
	}
}
```

Append the new tests:

```go
func TestDetachReportsFailureWhenNoRegistryEntryAppears(t *testing.T) {
	resetTimerFlags(t)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	rec := stubSpawn(t)

	origTimeout := handshakeTimeout
	handshakeTimeout = 20 * time.Millisecond
	t.Cleanup(func() { handshakeTimeout = origTimeout })

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	rootCmd.SetArgs([]string{"-d", "5m"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if rec.calls != 1 {
		t.Fatalf("spawnDetached called %d times, want 1", rec.calls)
	}
	want := fmt.Sprintf("timer 5m may have failed to start — see %s\n", filepath.Join(configHome, "timer", "timer.log"))
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestWaitForRegistryEntryAppears(t *testing.T) {
	dir := t.TempDir()
	go func() {
		time.Sleep(50 * time.Millisecond)
		os.WriteFile(filepath.Join(dir, "4242.json"), []byte("{}"), 0o644)
	}()

	if !waitForRegistryEntry(dir, 4242, time.Second) {
		t.Fatal("waitForRegistryEntry = false, want true once the file appears")
	}
}

func TestWaitForRegistryEntryTimesOut(t *testing.T) {
	if waitForRegistryEntry(t.TempDir(), 4242, 20*time.Millisecond) {
		t.Fatal("waitForRegistryEntry = true, want false for a missing file")
	}
}
```

(`fmt` is already imported in `cmd/detach_test.go`.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/ -v`
Expected: compile FAIL — `spawnDetached` var type mismatch (now returns
`(int, error)`), `undefined: handshakeTimeout`,
`undefined: waitForRegistryEntry`.

- [ ] **Step 3: Write minimal implementation**

In `cmd/detach_darwin.go`, change `spawnDetachedTimer` to return the
pid:

```go
func spawnDetachedTimer(args []string, env []string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, err
	}

	c := exec.Command(exe, args...)
	c.Env = env
	c.Stdin, c.Stdout, c.Stderr = nil, nil, nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := c.Start(); err != nil {
		return 0, err
	}
	return c.Process.Pid, nil
}
```

In `cmd/detach.go`, append after `confirmationLine`:

```go
// handshakeTimeout is how long the parent waits for the child's registry
// entry before reporting the start as unverified. It is a variable so
// tests can shorten it.
var handshakeTimeout = 2 * time.Second

func waitForRegistryEntry(dir string, pid int, timeout time.Duration) bool {
	target := filepath.Join(dir, fmt.Sprintf("%d.json", pid))
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(target); err == nil {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func logPathHint() string {
	p, err := getLogPath()
	if err != nil {
		return "the timer log next to your config"
	}
	return p
}

// confirmationAfterHandshake verifies that the child registered itself
// before reporting success; a child that dies early is reported as a
// possible failure instead of a started timer.
func confirmationAfterHandshake(input, name string, pid int) string {
	dir, err := getRunningDir()
	if err != nil {
		return fmt.Sprintf("timer %s started — could not verify; see %s", input, logPathHint())
	}
	if waitForRegistryEntry(dir, pid, handshakeTimeout) {
		return confirmationLine(input, name)
	}
	return fmt.Sprintf("timer %s may have failed to start — see %s", input, logPathHint())
}
```

Add `path/filepath` to the `cmd/detach.go` imports.

In `cmd/start.go`, `runStart`, replace the confirmation print:

```go
	if detach {
		pid, err := spawnDetached(detachedChildArgv(cfgFile), detachedEnv(os.Environ(), duration, timerName))
		if err != nil {
			log.Panicf("could not detach timer: %s", err.Error())
		}
		cmd.Println(confirmationAfterHandshake(inputStr, timerName, pid))
		return
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/ -v`
Expected: PASS.

- [ ] **Step 5: Run the full suite**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet warnings.

- [ ] **Step 6: Commit**

```bash
git add cmd/detach.go cmd/detach_darwin.go cmd/start.go cmd/detach_test.go
git commit -m "feat: confirm a detached timer only after its registry entry appears

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 5: Wait on wall-clock ticks so the notification fires after sleep or clock jumps

**Files:**
- Modify: `internal/commands/detach.go` — `RunDetached` tick loop,
  `now` and `maxTick` package vars
- Test: `internal/commands/detach_test.go` (append)

**Interfaces:**
- Consumes: `NotifyComplete` (existing), `notify` stub (existing).
- Produces:
  - `now func() time.Time` (package var, default `time.Now`) — test seam
    for wall-clock jumps
  - `maxTick time.Duration` (package var, default `time.Second`) — the
    longest single sleep; test seam for speed

- [ ] **Step 1: Write the failing tests**

Append to `internal/commands/detach_test.go` (imports `testing`, `time`
already present):

```go
// stubNow makes now() return base on the first call and afterwards the
// base jumped past the timer end — what the wall clock does when the
// system sleeps through the deadline.
func stubNow(t *testing.T, base time.Time, jump time.Duration) {
	t.Helper()
	orig := now
	calls := 0
	now = func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(jump)
	}
	t.Cleanup(func() { now = orig })
}

func TestRunDetachedNotifiesAfterWallClockJump(t *testing.T) {
	rec := stubNotify(t)
	origTick := maxTick
	maxTick = time.Millisecond
	t.Cleanup(func() { maxTick = origTick })

	base := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	stubNow(t, base, 2*time.Minute)

	if err := RunDetached(time.Minute, "Tea"); err != nil {
		t.Fatalf("RunDetached: %v", err)
	}

	if rec.calls != 1 {
		t.Fatalf("notification fired %d times, want 1", rec.calls)
	}
	if rec.message != `Your timer "Tea" is completed!` {
		t.Errorf("message = %q, want %q", rec.message, `Your timer "Tea" is completed!`)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/commands/ -run 'TestRunDetached' -v`
Expected: FAIL — `undefined: now`, `undefined: maxTick`.

- [ ] **Step 3: Write minimal implementation**

Replace `internal/commands/detach.go` with:

```go
package commands

import "time"

// now is a variable so tests can simulate wall-clock jumps (system
// sleep, NTP).
var now = time.Now

// maxTick is the longest single sleep while waiting. A short tick keeps
// the timer responsive to wall-clock jumps: after the system wakes, the
// next tick sees the deadline has passed and notifies immediately. It is
// a variable so tests can shorten it.
var maxTick = time.Second

// RunDetached waits out the timer without a UI and fires the completion
// notification once. It re-checks the wall clock every tick, so a timer
// that outlasts a system sleep or an NTP jump fires promptly when the
// wall clock reaches the deadline.
func RunDetached(remaining time.Duration, name string) error {
	end := now().Add(remaining)
	for {
		left := end.Sub(now())
		if left <= 0 {
			return NotifyComplete(name)
		}
		if left > maxTick {
			left = maxTick
		}
		time.Sleep(left)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands/ -v`
Expected: PASS. The existing `TestRunDetachedNotifiesWithName` and
`TestRunDetachedNotifiesWithoutName` still pass: with a real clock and a
1 ms timer the loop sleeps once, then notifies.

- [ ] **Step 5: Run the full suite**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet warnings.

- [ ] **Step 6: Commit**

```bash
git add internal/commands/detach.go internal/commands/detach_test.go
git commit -m "feat: notify detached timers on wall-clock time after sleep or clock jumps

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Self-Review Notes

- Spec coverage: payload via env, strict marker, filtered env, and
  parent-built argv (Task 3); handshake confirmation naming the log
  (Task 4); wall-clock ticks (Task 5); nanosecond registry with fallback
  (Task 2); start subcommand name and shared Run body (Task 1). Spec
  items 1–7 each map to exactly one task.
- The two existing confirmation tests were deliberately updated in Task
  4 (not Task 3) so the handshake lands in one commit with its tests.
- `stubSpawn` and `spawnRecord` are rewritten in both Task 3 and Task 4
  because `spawnDetached` changes signature in each; an executor working
  task-by-task applies the newer version on top.
- No placeholders: every code block is complete, every test asserts a
  concrete output string or value.
- Type consistency: `detachedChildArgv(cfgFile string) []string`,
  `detachedEnv(parent []string, d time.Duration, name string) []string`,
  `spawnDetached(args []string, env []string) (int, error)` (final form
  in Task 4), `waitForRegistryEntry(dir string, pid int, timeout
  time.Duration) bool` — signatures match between the task that defines
  them and the tasks that consume them.
- `test:` commit prefixes are not used; tests ship inside each `feat:`
  commit, matching the repo's history.
