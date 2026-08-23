package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EwanGreer/timer/internal/art"
)

func loadArtOverride(t *testing.T, name, content string) *art.Set {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	s, warns := art.Load(dir)
	if len(warns) != 0 {
		t.Fatalf("Load warns: %v", warns)
	}
	return s
}

func TestStartModelViewUsesCustomArt(t *testing.T) {
	s := loadArtOverride(t, "done.txt", "CUSTOM DONE")

	m := StartModel{done: true, Art: s}
	if view := m.View(); !strings.Contains(view, "CUSTOM DONE") {
		t.Fatalf("View = %q, want custom done art", view)
	}
}

func TestStartModelViewNilArtUsesBuiltIn(t *testing.T) {
	// lipgloss right-pads styled lines, so assert on the widest line, which
	// renders unchanged.
	firstLine, _, _ := strings.Cut(art.Default().Done(), "\n")

	m := StartModel{done: true}
	if view := m.View(); !strings.Contains(view, firstLine) {
		t.Fatalf("View = %q, want built-in done art", view)
	}
}

func TestStopWatchModelViewUsesCustomArt(t *testing.T) {
	s := loadArtOverride(t, "2.txt", "AAAAA\nBBBBB\nCCCCC\nDDDDD\nEEEEE")

	m := StopWatchModel{StartTime: time.Now().Add(-2 * time.Second), Art: s}
	if view := m.View(); !strings.Contains(view, "AAAAA") {
		t.Fatalf("View = %q, want custom '2' glyph", view)
	}
}
