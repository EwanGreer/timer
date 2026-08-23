package art

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderClockGolden(t *testing.T) {
	want := []string{
		"  █     ███          ████   █  █",
		" ██        █    ░        █  █  █",
		"  █     ███           ███   █████",
		"  █    █        ░        █     █",
		" ███   █████         ████      █",
	}

	gotLines := strings.Split(Default().RenderClock("12:34"), "\n")
	if len(gotLines) != 5 {
		t.Fatalf("RenderClock renders %d lines, want 5", len(gotLines))
	}

	width := utf8.RuneCountInString(gotLines[0])
	for i, line := range gotLines {
		if utf8.RuneCountInString(line) != width {
			t.Errorf("line %d has width %d, want %d", i, utf8.RuneCountInString(line), width)
		}
		if strings.TrimRight(line, " ") != want[i] {
			t.Errorf("line %d = %q, want %q", i, line, want[i])
		}
	}
}

func TestDoneGolden(t *testing.T) {
	want := `██████   ██████  ███    ██ ███████
██   ██ ██    ██ ████   ██ ██
██   ██ ██    ██ ██ ██  ██ █████
██   ██ ██    ██ ██  ██ ██ ██
██████   ██████  ██   ████ ███████`

	if got := Default().Done(); got != want {
		t.Fatalf("Done mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestRenderClockUnknownCharFallsBackToZero(t *testing.T) {
	if got, want := Default().RenderClock("!"), Default().RenderClock("0"); got != want {
		t.Fatalf("RenderClock(!) = %q, want %q", got, want)
	}
}

func TestDefaultGlyphsRenderFiveLines(t *testing.T) {
	for _, ch := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ":"} {
		got := Default().RenderClock(ch)
		if lines := countLines(got); lines != 5 {
			t.Errorf("RenderClock(%q) renders %d lines, want 5", ch, lines)
		}
	}
}

func countLines(s string) int {
	n := 1
	for _, r := range s {
		if r == '\n' {
			n++
		}
	}
	return n
}
