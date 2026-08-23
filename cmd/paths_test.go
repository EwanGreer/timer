package cmd

import (
	"path/filepath"
	"testing"
)

func TestGetRunningDirFromXDG(t *testing.T) {
	resetCfgFile(t)
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	got, err := getRunningDir()
	if err != nil {
		t.Fatalf("getRunningDir: %v", err)
	}
	if want := filepath.Join(cfgHome, "timer", "running"); got != want {
		t.Fatalf("getRunningDir = %q, want %q", got, want)
	}
}

func TestGetRunningDirFollowsConfigFlag(t *testing.T) {
	cfgFile = filepath.Join(t.TempDir(), "custom.toml")
	t.Cleanup(func() { cfgFile = "" })

	got, err := getRunningDir()
	if err != nil {
		t.Fatalf("getRunningDir: %v", err)
	}
	if want := filepath.Join(filepath.Dir(cfgFile), "running"); got != want {
		t.Fatalf("getRunningDir = %q, want %q", got, want)
	}
}

func TestGetLogPathFromXDG(t *testing.T) {
	resetCfgFile(t)
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	got, err := getLogPath()
	if err != nil {
		t.Fatalf("getLogPath: %v", err)
	}
	if want := filepath.Join(cfgHome, "timer", "timer.log"); got != want {
		t.Fatalf("getLogPath = %q, want %q", got, want)
	}
}
