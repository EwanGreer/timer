package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/EwanGreer/timer/internal/registry"
)

func resetStopFlags(t *testing.T) {
	t.Helper()

	old := stopAll
	stopAll = false
	t.Cleanup(func() { stopAll = old })
}

type stopRecord struct {
	pids []int
}

func stubStop(t *testing.T, err error) *stopRecord {
	t.Helper()

	var rec stopRecord
	orig := stopTimer
	stopTimer = func(dir string, pid int) error {
		rec.pids = append(rec.pids, pid)
		return err
	}
	t.Cleanup(func() { stopTimer = orig })

	return &rec
}

func stubTimers(t *testing.T, timers []registry.Timer) {
	t.Helper()

	orig := readTimers
	readTimers = func(dir string) ([]registry.Timer, error) { return timers, nil }
	t.Cleanup(func() { readTimers = orig })
}

func runStop(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(io.Discard)
	t.Cleanup(func() { rootCmd.SetErr(nil) })
	rootCmd.SetArgs(append([]string{"stop"}, args...))

	err := rootCmd.Execute()

	return out.String(), err
}

func samePids(got []int, want ...int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestStopNamedTimerPrintsNameAndID(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea", Remaining: 3 * time.Minute}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "2491")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 2491) {
		t.Fatalf("stopped pids = %v, want [2491]", rec.pids)
	}
	if want := "stopped timer \"Tea\" (2491)\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopUnnamedTimerPrintsIDOnly(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2930}})
	stubStop(t, nil)

	out, err := runStop(t, "2930")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if want := "stopped timer 2930\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopSeveralIDsStopsEachInOrder(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}, {Pid: 2930}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "2491", "2930")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 2491, 2930) {
		t.Fatalf("stopped pids = %v, want [2491 2930]", rec.pids)
	}
	want := "stopped timer \"Tea\" (2491)\nstopped timer 2930\n"
	if out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopAllStopsEveryTimer(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}, {Pid: 2930}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "--all")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 2491, 2930) {
		t.Fatalf("stopped pids = %v, want [2491 2930]", rec.pids)
	}
	if want := "stopped 2 timers\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopAllWithOneTimerUsesSingular(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	stubStop(t, nil)

	out, err := runStop(t, "--all")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if want := "stopped 1 timer\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopAllWithNoTimersSucceeds(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, nil)
	rec := stubStop(t, nil)

	out, err := runStop(t, "--all")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(rec.pids) != 0 {
		t.Fatalf("stopped pids = %v, want none", rec.pids)
	}
	if want := "no running timers\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopUnknownIDReturnsErrorAndStopsNothing(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	rec := stubStop(t, nil)

	_, err := runStop(t, "999")
	if err == nil || !strings.Contains(err.Error(), "no running timer with ID 999") {
		t.Fatalf("error = %v, want an error mentioning ID 999", err)
	}
	if len(rec.pids) != 0 {
		t.Fatalf("stopped pids = %v, want none", rec.pids)
	}
}

func TestStopStopsKnownIDsAndReportsUnknownOne(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "2491", "999")
	if err == nil || !strings.Contains(err.Error(), "no running timer with ID 999") {
		t.Fatalf("error = %v, want an error mentioning ID 999", err)
	}
	if !samePids(rec.pids, 2491) {
		t.Fatalf("stopped pids = %v, want [2491]", rec.pids)
	}
	if want := "stopped timer \"Tea\" (2491)\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopInvalidIDReturnsError(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	stubStop(t, nil)

	_, err := runStop(t, "abc")
	if err == nil || !strings.Contains(err.Error(), `invalid timer ID "abc"`) {
		t.Fatalf("error = %v, want an error mentioning abc", err)
	}
}

func TestStopReportsSignalFailure(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	stubStop(t, errors.New("operation not permitted"))

	out, err := runStop(t, "2491")
	if err == nil || !strings.Contains(err.Error(), "operation not permitted") {
		t.Fatalf("error = %v, want an error mentioning the failed signal", err)
	}
	if out != "" {
		t.Fatalf("output = %q, want nothing for a timer that did not stop", out)
	}
}

func TestStopWithoutIDsOrAllReturnsError(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	rec := stubStop(t, nil)

	_, err := runStop(t)
	if err == nil || !strings.Contains(err.Error(), "--all") {
		t.Fatalf("error = %v, want an error mentioning --all", err)
	}
	if len(rec.pids) != 0 {
		t.Fatalf("stopped pids = %v, want none", rec.pids)
	}
}

func TestStopPrefixStopsSingleMatchingTimer(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}, {Pid: 3412}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "249")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 2491) {
		t.Fatalf("stopped pids = %v, want [2491]", rec.pids)
	}
	if want := "stopped timer \"Tea\" (2491)\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopExactMatchWinsOverLongerPrefix(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 249, Name: "Short"}, {Pid: 2491, Name: "Tea"}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "249")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 249) {
		t.Fatalf("stopped pids = %v, want [249]", rec.pids)
	}
	if want := "stopped timer \"Short\" (249)\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopAmbiguousPrefixReturnsErrorAndStopsNothing(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}, {Pid: 2492, Name: "Toast"}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "249")
	if err == nil || !strings.Contains(err.Error(), "timer ID 249 matches 2491, 2492") {
		t.Fatalf("error = %v, want an error listing both matches", err)
	}
	if len(rec.pids) != 0 {
		t.Fatalf("stopped pids = %v, want none", rec.pids)
	}
	if out != "" {
		t.Fatalf("output = %q, want nothing for an ambiguous ID", out)
	}
}

func TestStopPrefixIgnoresLeadingZeros(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stubTimers(t, []registry.Timer{{Pid: 2491, Name: "Tea"}})
	rec := stubStop(t, nil)

	out, err := runStop(t, "0249")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !samePids(rec.pids, 2491) {
		t.Fatalf("stopped pids = %v, want [2491]", rec.pids)
	}
	if want := "stopped timer \"Tea\" (2491)\n"; out != want {
		t.Fatalf("output = %q, want %q", out, want)
	}
}

func TestStopAllWithIDsReturnsError(t *testing.T) {
	resetStopFlags(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	rec := stubStop(t, nil)

	_, err := runStop(t, "--all", "2491")
	if err == nil || !strings.Contains(err.Error(), "--all") {
		t.Fatalf("error = %v, want an error mentioning --all", err)
	}
	if len(rec.pids) != 0 {
		t.Fatalf("stopped pids = %v, want none", rec.pids)
	}
}
