package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetCfgFile(t *testing.T) {
	t.Helper()

	old := cfgFile
	cfgFile = ""
	t.Cleanup(func() { cfgFile = old })
}

func TestGetArtDirFromXDG(t *testing.T) {
	resetCfgFile(t)
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	got, err := getArtDir()
	if err != nil {
		t.Fatalf("getArtDir: %v", err)
	}
	if want := filepath.Join(cfgHome, "timer", "art"); got != want {
		t.Fatalf("getArtDir = %q, want %q", got, want)
	}
}

func TestGetArtDirWithoutXDGUsesDotConfig(t *testing.T) {
	resetCfgFile(t)
	t.Setenv("XDG_CONFIG_HOME", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	got, err := getArtDir()
	if err != nil {
		t.Fatalf("getArtDir: %v", err)
	}
	if want := filepath.Join(home, ".config", "timer", "art"); got != want {
		t.Fatalf("getArtDir = %q, want %q", got, want)
	}
}

func TestGetArtDirFollowsConfigFlag(t *testing.T) {
	cfgFile = filepath.Join(t.TempDir(), "custom.toml")
	t.Cleanup(func() { cfgFile = "" })

	got, err := getArtDir()
	if err != nil {
		t.Fatalf("getArtDir: %v", err)
	}
	if want := filepath.Join(filepath.Dir(cfgFile), "art"); got != want {
		t.Fatalf("getArtDir = %q, want %q", got, want)
	}
}

func TestLoadArtUsesArtDir(t *testing.T) {
	resetCfgFile(t)
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	artDir := filepath.Join(cfgHome, "timer", "art")
	if err := os.MkdirAll(artDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artDir, "3.txt"), []byte("AAA\nBBB\nCCC\nDDD\nEEE"), 0o644); err != nil {
		t.Fatalf("write 3.txt: %v", err)
	}

	s := loadArt()
	if got := s.RenderClock("3"); !strings.Contains(got, "AAA") {
		t.Fatalf("RenderClock(3) = %q, want custom glyph", got)
	}
}
