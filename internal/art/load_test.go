package art

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeArtFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadMissingDirUsesBuiltIn(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nope")

	s, warns := Load(dir)

	if len(warns) != 0 {
		t.Fatalf("Load warns for a missing dir: %v", warns)
	}
	if got, want := s.RenderClock("12:34"), Default().RenderClock("12:34"); got != want {
		t.Fatalf("RenderClock mismatch:\n got: %q\nwant: %q", got, want)
	}
	if got, want := s.Done(), Default().Done(); got != want {
		t.Fatalf("Done mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestLoadEmptyDirNameUsesBuiltIn(t *testing.T) {
	// An empty dir must not pick up stray art files from the working
	// directory.
	t.Chdir(t.TempDir())
	writeArtFile(t, ".", "done.txt", "STRAY")

	s, warns := Load("")

	if len(warns) != 0 {
		t.Fatalf("Load warns for an empty dir: %v", warns)
	}
	if got, want := s.Done(), Default().Done(); got != want {
		t.Fatalf("Done = %q, want built-in %q", got, want)
	}
}

func TestLoadOverridesDigit(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "3.txt", "XXX\nYYY\nZZZ\nAAA\nBBB")

	s, warns := Load(dir)

	if len(warns) != 0 {
		t.Fatalf("Load warns for a valid glyph: %v", warns)
	}
	if got, want := s.RenderClock("3"), "XXX    \nYYY    \nZZZ    \nAAA    \nBBB    "; got != want {
		t.Fatalf("RenderClock(3) = %q, want %q", got, want)
	}
	if got, want := s.RenderClock("0"), Default().RenderClock("0"); got != want {
		t.Fatalf("RenderClock(0) = %q, want built-in %q", got, want)
	}
}

func TestLoadOverridesDone(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "done.txt", "CUSTOM\nDONE")

	s, warns := Load(dir)

	if len(warns) != 0 {
		t.Fatalf("Load warns for a valid done art: %v", warns)
	}
	if got, want := s.Done(), "CUSTOM\nDONE"; got != want {
		t.Fatalf("Done = %q, want %q", got, want)
	}
}

func TestLoadInvalidGlyphFallsBackWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "3.txt", "too\nshort")

	s, warns := Load(dir)

	if len(warns) != 1 || !strings.Contains(warns[0], "3.txt") {
		t.Fatalf("warns = %v, want one warning naming 3.txt", warns)
	}
	if got, want := s.RenderClock("3"), Default().RenderClock("3"); got != want {
		t.Fatalf("RenderClock(3) = %q, want built-in %q", got, want)
	}
}

func TestLoadBlankGlyphFallsBackWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "3.txt", "\n\n\n\n\n")

	s, warns := Load(dir)

	if len(warns) != 1 || !strings.Contains(warns[0], "3.txt") {
		t.Fatalf("warns = %v, want one warning naming 3.txt", warns)
	}
	if got, want := s.RenderClock("3"), Default().RenderClock("3"); got != want {
		t.Fatalf("RenderClock(3) = %q, want built-in %q", got, want)
	}
}

func TestLoadUnreadableFileFallsBackWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "3.txt", "XXX\nYYY\nZZZ\nAAA\nBBB")
	path := filepath.Join(dir, "3.txt")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(path, 0o644) })

	s, warns := Load(dir)

	if len(warns) != 1 || !strings.Contains(warns[0], "3.txt") {
		t.Fatalf("warns = %v, want one warning naming 3.txt", warns)
	}
	if got, want := s.RenderClock("3"), Default().RenderClock("3"); got != want {
		t.Fatalf("RenderClock(3) = %q, want built-in %q", got, want)
	}
}

func TestLoadNormalizesRaggedGlyphWidth(t *testing.T) {
	dir := t.TempDir()
	writeArtFile(t, dir, "3.txt", "AA\nBBBBBBB\nC\nDDDD\nEEE")

	s, warns := Load(dir)

	if len(warns) != 0 {
		t.Fatalf("Load warns for a valid glyph: %v", warns)
	}

	gotLines := strings.Split(s.RenderClock("30"), "\n")
	if len(gotLines) != 5 {
		t.Fatalf("RenderClock(30) renders %d lines, want 5", len(gotLines))
	}
	// The ragged glyph (max width 7) widens the grid for every glyph.
	for i, line := range gotLines {
		if width := len([]rune(line)); width != 18 {
			t.Errorf("line %d = %q has width %d, want 18", i, line, width)
		}
	}
}
