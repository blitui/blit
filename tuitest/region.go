package tuitest

import "strings"

// Region is a rectangular sub-area of a Screen for scoped assertions.
type Region struct {
	screen *Screen
	row    int
	col    int
	width  int
	height int
}

// Contains reports whether the region contains the given text on any row.
func (r *Region) Contains(text string) bool {
	for i := 0; i < r.height; i++ {
		if strings.Contains(r.Row(i), text) {
			return true
		}
	}
	return false
}

// Row returns the text content of a row within the region (0-indexed relative
// to the region's top-left corner), trimmed of trailing spaces.
func (r *Region) Row(row int) string {
	absRow := r.row + row
	if row < 0 || row >= r.height {
		return ""
	}
	return strings.TrimRight(r.screen.TextAt(absRow, r.col, r.col+r.width), " ")
}

// String returns a plain text representation of the region.
func (r *Region) String() string {
	var b strings.Builder
	for i := 0; i < r.height; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(r.Row(i))
	}
	return b.String()
}
