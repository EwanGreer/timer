package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/EwanGreer/timer/internal/registry"
)

func TestFormatRemaining(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{3*time.Minute + 24*time.Second, "3m24s"},
		{12*time.Minute + 7*time.Second, "12m07s"},
		{1*time.Hour + 2*time.Minute + 10*time.Second, "1h02m10s"},
		{time.Hour, "1h00m00s"},
	}

	for _, tt := range tests {
		if got := formatRemaining(tt.in); got != tt.want {
			t.Errorf("formatRemaining(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRenderTable(t *testing.T) {
	started := time.Date(2026, 8, 23, 14, 32, 0, 0, time.Local)
	timers := []registry.Timer{
		{Pid: 2930, Name: "", StartedAt: started.Add(-time.Minute), Remaining: 12*time.Minute + 7*time.Second},
		{Pid: 2491, Name: "Tea", StartedAt: started, Remaining: 3*time.Minute + 24*time.Second},
	}

	var out bytes.Buffer
	renderTable(&out, timers)

	want := "ID    NAME  REMAINING  STARTED\n" +
		"2930  -     12m07s     14:31\n" +
		"2491  Tea   3m24s      14:32\n"
	if out.String() != want {
		t.Fatalf("table = %q, want %q", out.String(), want)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	var out bytes.Buffer
	renderTable(&out, nil)
	if want := "ID  NAME  REMAINING  STARTED\n"; out.String() != want {
		t.Fatalf("table = %q, want %q", out.String(), want)
	}
}

func TestPsRunPrintsHeaderWithNoTimers(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	rootCmd.SetOut(&out)

	orig := readTimers
	readTimers = func(dir string) ([]registry.Timer, error) {
		return nil, nil
	}
	t.Cleanup(func() { readTimers = orig })

	rootCmd.SetArgs([]string{"ps"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if want := "ID  NAME  REMAINING  STARTED\n"; out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestPsRunPrintsTimers(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	rootCmd.SetOut(&out)

	orig := readTimers
	readTimers = func(dir string) ([]registry.Timer, error) {
		return []registry.Timer{{Pid: 2491, Name: "Tea", Remaining: 3*time.Minute + 24*time.Second}}, nil
	}
	t.Cleanup(func() { readTimers = orig })

	rootCmd.SetArgs([]string{"ps"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(out.String(), "2491  Tea   3m24s") {
		t.Fatalf("output = %q, want a Tea row", out.String())
	}
}

func TestPsRunReturnsErrorOnReadFailure(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	orig := readTimers
	readTimers = func(dir string) ([]registry.Timer, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() { readTimers = orig })

	rootCmd.SetArgs([]string{"ps"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %v, want an error mentioning boom", err)
	}
}
