// Package btest provides a virtual terminal screen and assertion helpers
// for testing Bubble Tea TUI applications. It uses go-te to emulate a terminal
// and parse ANSI escape sequences, enabling screen-based assertions on
// rendered View() output.
package btest

import (
	"regexp"
	"strings"

	"github.com/rcarmo/go-te/pkg/te"
)

// Screen wraps go-te to provide a virtual terminal for testing TUI output.
type Screen struct {
	screen *te.Screen
	stream *te.ByteStream
	cols   int
	lines  int
}

// NewScreen creates a virtual terminal screen of the given dimensions.
func NewScreen(cols, lines int) *Screen {
	scr := te.NewScreen(cols, lines)
	bs := te.NewByteStream(scr, false)
	return &Screen{
		screen: scr,
		stream: bs,
		cols:   cols,
		lines:  lines,
	}
}

// Render feeds View() output (with ANSI codes) into the virtual terminal.
// It translates bare \n to \r\n (mimicking the terminal's ONLCR flag) so that
// Bubble Tea View() output renders correctly.
func (s *Screen) Render(output string) {
	s.screen.Reset()
	// Normalize line endings: replace bare \n with \r\n (but don't double up \r\n).
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\n", "\r\n")
	_ = s.stream.Feed([]byte(normalized))
}

// TextAt returns the text content at the given row, from startCol to endCol.
// Row and col are 0-indexed. endCol is exclusive.
func (s *Screen) TextAt(row, startCol, endCol int) string {
	if row < 0 || row >= s.lines {
		return ""
	}
	if startCol < 0 {
		startCol = 0
	}
	if endCol > s.cols {
		endCol = s.cols
	}
	if startCol >= endCol {
		return ""
	}
	buf := s.screen.Buffer
	if row >= len(buf) {
		return ""
	}
	var b strings.Builder
	rowBuf := buf[row]
	for c := startCol; c < endCol && c < len(rowBuf); c++ {
		data := rowBuf[c].Data
		if data == "" {
			b.WriteByte(' ')
		} else {
			b.WriteString(data)
		}
	}
	return b.String()
}

// Row returns the full text content of a row (trimmed of trailing spaces).
func (s *Screen) Row(row int) string {
	return strings.TrimRight(s.TextAt(row, 0, s.cols), " ")
}

// Contains reports whether the screen contains the given text anywhere.
func (s *Screen) Contains(text string) bool {
	content := s.String()
	return strings.Contains(content, text)
}

// ContainsAt reports whether the given text appears starting at (row, col).
func (s *Screen) ContainsAt(row, col int, text string) bool {
	actual := s.TextAt(row, col, col+len(text))
	return actual == text
}

// CursorPos returns the current cursor position.
func (s *Screen) CursorPos() (row, col int) {
	return s.screen.Cursor.Row, s.screen.Cursor.Col
}

// StyleAt returns the style attributes of the cell at (row, col).
func (s *Screen) StyleAt(row, col int) CellStyle {
	if row < 0 || row >= s.lines || col < 0 || col >= s.cols {
		return CellStyle{}
	}
	buf := s.screen.Buffer
	if row >= len(buf) || col >= len(buf[row]) {
		return CellStyle{}
	}
	attr := buf[row][col].Attr
	return CellStyle{
		Fg:        colorToString(attr.Fg),
		Bg:        colorToString(attr.Bg),
		Bold:      attr.Bold,
		Italic:    attr.Italics,
		Underline: attr.Underline,
		Reverse:   attr.Reverse,
	}
}

// colorToString converts a go-te Color to a string representation.
// Returns empty string for default colors.
func colorToString(c te.Color) string {
	if c.Name == "" || c.Name == "default" {
		return ""
	}
	return c.Name
}

// Region returns a sub-screen for bounded assertions.
func (s *Screen) Region(row, col, width, height int) *Region {
	return &Region{
		screen: s,
		row:    row,
		col:    col,
		width:  width,
		height: height,
	}
}

// FindText returns the (row, col) of the first occurrence of text on the screen.
// Returns (-1, -1) if not found.
func (s *Screen) FindText(text string) (row, col int) {
	for r := 0; r < s.lines; r++ {
		rowText := s.TextAt(r, 0, s.cols)
		if idx := strings.Index(rowText, text); idx >= 0 {
			return r, idx
		}
	}
	return -1, -1
}

// FindAllText returns all (row, col) positions where text appears.
func (s *Screen) FindAllText(text string) [][2]int {
	var results [][2]int
	for r := 0; r < s.lines; r++ {
		rowText := s.TextAt(r, 0, s.cols)
		offset := 0
		for {
			idx := strings.Index(rowText[offset:], text)
			if idx < 0 {
				break
			}
			results = append(results, [2]int{r, offset + idx})
			offset += idx + len(text)
		}
	}
	return results
}

// RowCount returns the number of non-empty rows on screen.
func (s *Screen) RowCount() int {
	count := 0
	for r := 0; r < s.lines; r++ {
		if s.Row(r) != "" {
			count++
		}
	}
	return count
}

// AllRows returns all row strings (including empty ones).
func (s *Screen) AllRows() []string {
	rows := make([]string, s.lines)
	for r := 0; r < s.lines; r++ {
		rows[r] = s.Row(r)
	}
	return rows
}

// NonEmptyRows returns only the non-empty rows with their row indices.
func (s *Screen) NonEmptyRows() []IndexedRow {
	var rows []IndexedRow
	for r := 0; r < s.lines; r++ {
		text := s.Row(r)
		if text != "" {
			rows = append(rows, IndexedRow{Index: r, Text: text})
		}
	}
	return rows
}

// IndexedRow pairs a row index with its text content.
type IndexedRow struct {
	Index int
	Text  string
}

// Column extracts a vertical column of text from startRow to endRow (exclusive).
func (s *Screen) Column(col, startRow, endRow int) string {
	if col < 0 || col >= s.cols {
		return ""
	}
	var b strings.Builder
	for r := startRow; r < endRow && r < s.lines; r++ {
		if r > startRow {
			b.WriteByte('\n')
		}
		ch := s.TextAt(r, col, col+1)
		b.WriteString(ch)
	}
	return b.String()
}

// IsEmpty reports whether the screen has no visible text content.
func (s *Screen) IsEmpty() bool {
	return s.RowCount() == 0
}

// Size returns the screen dimensions as (cols, lines).
func (s *Screen) Size() (cols, lines int) {
	return s.cols, s.lines
}

// CountOccurrences returns how many times text appears on the screen.
func (s *Screen) CountOccurrences(text string) int {
	count := 0
	for r := 0; r < s.lines; r++ {
		rowText := s.TextAt(r, 0, s.cols)
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

// MatchesRegexp reports whether the screen content matches the regular expression.
func (s *Screen) MatchesRegexp(pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s.String())
}

// FindRegexp returns the (row, col) of the first regexp match. Returns (-1, -1) if not found.
func (s *Screen) FindRegexp(pattern string) (row, col int) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1, -1
	}
	for r := 0; r < s.lines; r++ {
		rowText := s.TextAt(r, 0, s.cols)
		loc := re.FindStringIndex(rowText)
		if loc != nil {
			return r, loc[0]
		}
	}
	return -1, -1
}

// String returns a plain text representation of the entire screen (for debugging/golden files).
func (s *Screen) String() string {
	var b strings.Builder
	for r := 0; r < s.lines; r++ {
		if r > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s.Row(r))
	}
	return b.String()
}
