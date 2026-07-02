package agent

import (
	"strings"
	"testing"
)

// TestStreamedRowsBasic covers the four cursor-position cases the markdown
// redraw uses to decide how far up to move before clearing.
func TestStreamedRowsBasic(t *testing.T) {
	cases := []struct {
		name  string
		input string
		width int
		want  int
	}{
		{"no newline fits", "hello", 80, 0},
		{"newline only", "hello\n", 80, 1},
		{"two lines no trailing", "hello\nworld", 80, 1},
		{"two lines trailing", "hello\nworld\n", 80, 2},
		{"empty", "", 80, 0},
		{"single wrap", "abcdefghij", 5, 1}, // 10 cols / width 5 → 1 wrap
		{"two wraps no nl", "abcdefghijklmno", 5, 2},
		{"line exactly width", strings.Repeat("a", 80), 80, 0}, // lazy wrap, no extra row
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := streamedRows(tc.input, tc.width); got != tc.want {
				t.Errorf("streamedRows(%q, %d) = %d, want %d", tc.input, tc.width, got, tc.want)
			}
		})
	}
}

// TestStreamedRowsCJK proves CJK doubles the column footprint so a Chinese
// line wraps at half the column count.
func TestStreamedRowsCJK(t *testing.T) {
	in := strings.Repeat("中", 6) // 12 cols at width 10 → 1 wrap
	if got := streamedRows(in, 10); got != 1 {
		t.Errorf("streamedRows(6×中, 10) = %d, want 1", got)
	}
}

// TestStreamedRowsIgnoresOSC: OSC sequences (ESC]...#BEL) must not inflate the
// row count, and their content must not be counted as visible characters.
func TestStreamedRowsIgnoresOSC(t *testing.T) {
	// OSC title set with BEL terminator, containing letters in the title.
	in := "\x1b]0;hello world\x07visible"
	// "visible" is 7 chars at width 10, no wrap: rows = 0
	if got := streamedRows(in, 10); got != 0 {
		t.Errorf("OSC set-title in 7-char line at width 10 should be 0 rows, got %d", got)
	}
	// Same OSC with a longer visible suffix that wraps at width 10.
	in2 := "\x1b]0;title\x07abcdefghijk" // 11 chars → 1 wrap
	if got := streamedRows(in2, 10); got != 1 {
		t.Errorf("OSC + 11-char line at width 10 should be 1 row (wrap), got %d", got)
	}
	// Verify that OSC containing the BEL-like characters still terminates.
	in3 := "\x1b]0;\x07hello"
	if got := streamedRows(in3, 80); got != 0 {
		t.Errorf("short OSC with BEL should not add rows, got %d", got)
	}
}

// TestStreamedRowsIgnoresAnsi: ANSI SGR codes must not inflate the row count.
func TestStreamedRowsIgnoresAnsi(t *testing.T) {
	in := "\x1b[1mhello\x1b[0m world"
	if got := streamedRows(in, 80); got != 0 {
		t.Errorf("ANSI in 11-char line at width 80 should be 0 rows, got %d", got)
	}
}
