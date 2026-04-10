package btest

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

// RowCount returns the number of non-empty rows in the region.
func (r *Region) RowCount() int {
	count := 0
	for i := 0; i < r.height; i++ {
		if r.Row(i) != "" {
			count++
		}
	}
	return count
}

// IsEmpty reports whether the region has no visible text.
func (r *Region) IsEmpty() bool {
	return r.RowCount() == 0
}

// FindText returns the (relativeRow, relativeCol) of the first occurrence
// of text within the region. Returns (-1, -1) if not found.
func (r *Region) FindText(text string) (row, col int) {
	for i := 0; i < r.height; i++ {
		rowText := r.Row(i)
		if idx := strings.Index(rowText, text); idx >= 0 {
			return i, idx
		}
	}
	return -1, -1
}

// StyleAt returns the cell style at the given region-relative coordinates.
func (r *Region) StyleAt(row, col int) CellStyle {
	return r.screen.StyleAt(r.row+row, r.col+col)
}

// AllRows returns all row strings in the region.
func (r *Region) AllRows() []string {
	rows := make([]string, r.height)
	for i := 0; i < r.height; i++ {
		rows[i] = r.Row(i)
	}
	return rows
}

// CountOccurrences returns how many times text appears in the region.
func (r *Region) CountOccurrences(text string) int {
	count := 0
	for i := 0; i < r.height; i++ {
		rowText := r.Row(i)
		offset := 0
		for {
			idx := strings.Index(rowText[offset:], text)
			if idx < 0 {
				break
			}
			count++
			offset += idx + len(text)
		}
	}
	return count
}
