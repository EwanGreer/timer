// Package art renders the timer's ASCII art. Art is read at runtime from an
// art directory so users can customise it, with built-in art embedded at
// build time as a fallback.
package art

import (
	"embed"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

//go:embed embedded/*.txt
var embedded embed.FS

const glyphHeight = 5

// Set holds the art used to render the timer display.
type Set struct {
	done      string
	digits    map[rune]string
	gridWidth int
}

// defaultSet is built from the embedded files and used whenever no art
// directory provides overrides.
var defaultSet = mustParseEmbedded()

// Default returns the built-in art set.
func Default() *Set {
	return defaultSet
}

// Done returns the completion art.
func (s *Set) Done() string {
	return s.done
}

// RenderClock renders a time string like "12:34" using big digits.
// Characters without a glyph render as the '0' glyph.
func (s *Set) RenderClock(timeStr string) string {
	lines := make([]string, glyphHeight)

	for _, ch := range timeStr {
		glyph, ok := s.digits[ch]
		if !ok {
			glyph = s.digits['0']
		}

		glyphLines := strings.Split(glyph, "\n")
		for i := range glyphHeight {
			lines[i] += glyphLines[i] + "  " // space between characters
		}
	}

	return strings.Join(lines, "\n")
}

func mustParseEmbedded() *Set {
	s := &Set{digits: map[rune]string{}}

	s.done = strings.TrimSuffix(readEmbedded("done.txt"), "\n")
	for _, key := range "0123456789:" {
		glyph, ok := parseGlyph(readEmbedded(glyphFile(key)))
		if !ok {
			panic(fmt.Sprintf("timer: invalid embedded art %s", glyphFile(key)))
		}
		s.digits[key] = glyph
	}
	s.normalize()

	return s
}

func readEmbedded(name string) string {
	content, err := embedded.ReadFile("embedded/" + name)
	if err != nil {
		panic(fmt.Sprintf("timer: missing embedded art %s: %v", name, err))
	}
	return string(content)
}

func glyphFile(key rune) string {
	if key == ':' {
		return "colon.txt"
	}
	return string(key) + ".txt"
}

// Load builds a Set from the art files in dir, falling back file by file to
// the built-in art. A missing dir or missing files produce no warnings;
// present but unusable files are reported in the returned warnings.
func Load(dir string) (*Set, []string) {
	s := &Set{
		done:      defaultSet.done,
		digits:    maps.Clone(defaultSet.digits),
		gridWidth: defaultSet.gridWidth,
	}
	if dir == "" {
		return s, nil
	}

	var warns []string
	for _, name := range []string{
		"done.txt", "0.txt", "1.txt", "2.txt", "3.txt", "4.txt", "5.txt",
		"6.txt", "7.txt", "8.txt", "9.txt", "colon.txt",
	} {
		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				warns = append(warns, fmt.Sprintf("%s: cannot read, using built-in", path))
			}
			continue
		}

		if name == "done.txt" {
			done := strings.TrimSuffix(string(content), "\n")
			if strings.TrimSpace(done) == "" {
				warns = append(warns, fmt.Sprintf("%s: empty, using built-in", path))
				continue
			}
			s.done = done
			continue
		}

		glyph, ok := parseGlyph(string(content))
		if !ok {
			warns = append(warns, fmt.Sprintf("%s: invalid glyph, using built-in", path))
			continue
		}
		s.digits[glyphKey(name)] = glyph
	}

	s.normalize()
	return s, warns
}

func glyphKey(name string) rune {
	if name == "colon.txt" {
		return ':'
	}
	return rune(name[0])
}

// parseGlyph validates glyph content: exactly glyphHeight rows, at least one
// of them non-blank. A trailing newline after the last row is allowed.
func parseGlyph(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == glyphHeight+1 && lines[glyphHeight] == "" {
		lines = lines[:glyphHeight] // trailing newline after the last row
	}
	if len(lines) != glyphHeight {
		return "", false
	}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return strings.Join(lines, "\n"), true
		}
	}
	return "", false
}

// normalize pads every glyph to one shared grid width so glyphs of different
// widths stay aligned when rendered side by side.
func (s *Set) normalize() {
	width := 0
	for _, glyph := range s.digits {
		for line := range strings.SplitSeq(glyph, "\n") {
			if w := utf8.RuneCountInString(line); w > width {
				width = w
			}
		}
	}

	s.gridWidth = width
	for key, glyph := range s.digits {
		s.digits[key] = padGlyph(glyph, width)
	}
}

func padGlyph(glyph string, width int) string {
	lines := strings.Split(glyph, "\n")
	for i, line := range lines {
		lines[i] = line + strings.Repeat(" ", width-utf8.RuneCountInString(line))
	}
	return strings.Join(lines, "\n")
}
