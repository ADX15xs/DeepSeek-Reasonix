package agent

import (
	"github.com/mattn/go-runewidth"
)

// streamedRows counts how many rows the cursor has descended after raw text
// of length s was printed at the given terminal width. Used by the markdown
// redraw to know how far up to move before clearing. Each \n descends one
// row; lines whose visible width exceeds the terminal width descend an extra
// row per wrap. A line exactly the terminal width does not wrap on its own —
// terminals "lazy-wrap" only when the next visible character lands.
func streamedRows(s string, width int) int {
	if width <= 0 {
		width = 80
	}

	var rows int
	var currentLineWidth int
	var inEscape bool
	var isOSC bool   // ESC] sequence; only BEL/ST terminate, not letters
	var stPending bool // ESC seen inside OSC; next \ completes ST terminator

	for _, r := range s {
		switch {
		case r == '\n':
			// End of current line
			if currentLineWidth > 0 {
				rows += (currentLineWidth - 1) / width
				currentLineWidth = 0
			}
			rows++
		case stPending && r == '\\':
			// ST terminator (ESC \) ends an OSC sequence.
			inEscape = false
			isOSC = false
			stPending = false
		case stPending:
			// ESC inside OSC was not ST; treat it as the start of a new
			// escape sequence and process the current character normally.
			stPending = false
			isOSC = false
			if r == ']' {
				isOSC = true
			} else if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
				inEscape = false
			}
		case r == '\x1b':
			// Start of ANSI escape sequence (CSI, OSC, etc.)
			if isOSC {
				// ESC inside OSC could be start of ST (ESC \)
				stPending = true
			} else {
				inEscape = true
				isOSC = false
			}
		case inEscape && r == ']':
			// ESC] starts an OSC sequence (e.g. ESC]0;title)
			isOSC = true
		case inEscape && r == '\x07':
			// BEL terminates an OSC sequence (ESC]...#BEL).
			inEscape = false
			isOSC = false
		case inEscape && !isOSC && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'):
			// Any ASCII letter terminates a CSI sequence (X3.64).
			// Not applied in OSC mode - OSC duration is delimited by BEL/ST.
			inEscape = false
		default:
			if !inEscape {
				// Only count visible characters
				w := runewidth.RuneWidth(r)
				currentLineWidth += w
			}
		}
	}

	// Add remaining line
	if currentLineWidth > 0 {
		rows += (currentLineWidth - 1) / width
	}

	return rows
}
