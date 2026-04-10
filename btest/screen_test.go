package btest

import "testing"

func TestNewScreen(t *testing.T) {
	s := NewScreen(80, 24)
	cols, lines := s.Size()
	if cols != 80 || lines != 24 {
		t.Errorf("Size() = (%d,%d), want (80,24)", cols, lines)
	}
	if !s.IsEmpty() {
		t.Error("new screen should be empty")
	}
}

func TestScreenRenderAndRow(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("Hello World\nLine Two")
	if got := s.Row(0); got != "Hello World" {
		t.Errorf("Row(0) = %q, want %q", got, "Hello World")
	}
	if got := s.Row(1); got != "Line Two" {
		t.Errorf("Row(1) = %q, want %q", got, "Line Two")
	}
	if got := s.Row(2); got != "" {
		t.Errorf("Row(2) = %q, want empty", got)
	}
}

func TestScreenTextAt(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("ABCDEFG")

	if got := s.TextAt(0, 0, 3); got != "ABC" {
		t.Errorf("TextAt(0,0,3) = %q, want %q", got, "ABC")
	}
	if got := s.TextAt(0, 2, 5); got != "CDE" {
		t.Errorf("TextAt(0,2,5) = %q, want %q", got, "CDE")
	}
	// Out of bounds
	if got := s.TextAt(-1, 0, 5); got != "" {
		t.Errorf("TextAt(-1,...) = %q, want empty", got)
	}
	if got := s.TextAt(99, 0, 5); got != "" {
		t.Errorf("TextAt(99,...) = %q, want empty", got)
	}
	// startCol >= endCol
	if got := s.TextAt(0, 5, 3); got != "" {
		t.Errorf("TextAt(0,5,3) = %q, want empty", got)
	}
}

func TestScreenContains(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("Hello World")
	if !s.Contains("World") {
		t.Error("Contains(World) should be true")
	}
	if s.Contains("Missing") {
		t.Error("Contains(Missing) should be false")
	}
}

func TestScreenContainsAt(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("ABCDEF")
	if !s.ContainsAt(0, 0, "ABC") {
		t.Error("ContainsAt(0,0,ABC) should be true")
	}
	if s.ContainsAt(0, 0, "XYZ") {
		t.Error("ContainsAt(0,0,XYZ) should be false")
	}
	if !s.ContainsAt(0, 3, "DEF") {
		t.Error("ContainsAt(0,3,DEF) should be true")
	}
}

func TestScreenFindText(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AAA\nBBB\nCCC")
	r, c := s.FindText("BBB")
	if r != 1 || c != 0 {
		t.Errorf("FindText(BBB) = (%d,%d), want (1,0)", r, c)
	}
	r, c = s.FindText("ZZZ")
	if r != -1 || c != -1 {
		t.Errorf("FindText(ZZZ) = (%d,%d), want (-1,-1)", r, c)
	}
}

func TestScreenFindAllText(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("ABAB\nABAB")
	results := s.FindAllText("AB")
	if len(results) != 4 {
		t.Errorf("FindAllText(AB) found %d, want 4", len(results))
	}
}

func TestScreenRowCount(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("A\nB\nC")
	if got := s.RowCount(); got != 3 {
		t.Errorf("RowCount() = %d, want 3", got)
	}
}

func TestScreenAllRows(t *testing.T) {
	s := NewScreen(40, 3)
	s.Render("X\nY")
	rows := s.AllRows()
	if len(rows) != 3 {
		t.Errorf("AllRows() len = %d, want 3", len(rows))
	}
	if rows[0] != "X" {
		t.Errorf("AllRows()[0] = %q, want X", rows[0])
	}
}

func TestScreenNonEmptyRows(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("A\n\nC")
	rows := s.NonEmptyRows()
	if len(rows) != 2 {
		t.Errorf("NonEmptyRows() len = %d, want 2", len(rows))
	}
	if rows[0].Index != 0 || rows[0].Text != "A" {
		t.Errorf("NonEmptyRows()[0] = %+v, want {0, A}", rows[0])
	}
	if rows[1].Index != 2 || rows[1].Text != "C" {
		t.Errorf("NonEmptyRows()[1] = %+v, want {2, C}", rows[1])
	}
}

func TestScreenColumn(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AB\nCD\nEF")
	got := s.Column(0, 0, 3)
	if got != "A\nC\nE" {
		t.Errorf("Column(0, 0, 3) = %q, want %q", got, "A\nC\nE")
	}
	// Out of bounds col
	if got := s.Column(-1, 0, 3); got != "" {
		t.Errorf("Column(-1,...) = %q, want empty", got)
	}
}

func TestScreenCountOccurrences(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AAAA\nAAA")
	if got := s.CountOccurrences("AA"); got != 3 {
		t.Errorf("CountOccurrences(AA) = %d, want 3", got)
	}
	if got := s.CountOccurrences("ZZ"); got != 0 {
		t.Errorf("CountOccurrences(ZZ) = %d, want 0", got)
	}
}

func TestScreenMatchesRegexp(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("Error: code 42")
	if !s.MatchesRegexp(`code \d+`) {
		t.Error("MatchesRegexp should match 'code \\d+'")
	}
	if s.MatchesRegexp(`^nothing$`) {
		t.Error("MatchesRegexp should not match '^nothing$'")
	}
	// Invalid regex
	if s.MatchesRegexp(`[invalid`) {
		t.Error("MatchesRegexp should return false for invalid regex")
	}
}

func TestScreenFindRegexp(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("foo\nbar 123 baz")
	r, c := s.FindRegexp(`\d+`)
	if r != 1 || c != 4 {
		t.Errorf("FindRegexp(\\d+) = (%d,%d), want (1,4)", r, c)
	}
	r, c = s.FindRegexp(`zzz`)
	if r != -1 || c != -1 {
		t.Errorf("FindRegexp(zzz) = (%d,%d), want (-1,-1)", r, c)
	}
}

func TestScreenString(t *testing.T) {
	s := NewScreen(40, 3)
	s.Render("A\nB")
	got := s.String()
	if got != "A\nB\n" {
		t.Errorf("String() = %q, want %q", got, "A\nB\n")
	}
}

func TestScreenStyleAt(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("Plain text")
	style := s.StyleAt(0, 0)
	// Default text has no special styles
	if style.Bold || style.Italic || style.Underline || style.Reverse {
		t.Error("plain text should have no style attributes")
	}
	// Out of bounds
	style = s.StyleAt(-1, 0)
	if style != (CellStyle{}) {
		t.Error("out-of-bounds StyleAt should return zero CellStyle")
	}
}

// --- Region tests ---

func TestRegionContains(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AABBCC\nDDEEFF\nGGHHII")
	r := s.Region(1, 2, 4, 2) // rows 1-2, cols 2-5
	if !r.Contains("EEFF") {
		t.Error("Region.Contains(EEFF) should be true")
	}
	if r.Contains("AABB") {
		t.Error("Region.Contains(AABB) should be false (outside region)")
	}
}

func TestRegionRow(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("0123456789\nABCDEFGHIJ")
	r := s.Region(0, 3, 4, 2) // cols 3-6, rows 0-1
	if got := r.Row(0); got != "3456" {
		t.Errorf("Region.Row(0) = %q, want %q", got, "3456")
	}
	if got := r.Row(1); got != "DEFG" {
		t.Errorf("Region.Row(1) = %q, want %q", got, "DEFG")
	}
	// Out of bounds
	if got := r.Row(-1); got != "" {
		t.Errorf("Region.Row(-1) = %q, want empty", got)
	}
	if got := r.Row(99); got != "" {
		t.Errorf("Region.Row(99) = %q, want empty", got)
	}
}

func TestRegionRowCount(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AA\n\nCC")
	r := s.Region(0, 0, 10, 3)
	if got := r.RowCount(); got != 2 {
		t.Errorf("Region.RowCount() = %d, want 2", got)
	}
}

func TestRegionIsEmpty(t *testing.T) {
	s := NewScreen(40, 5)
	// Region over empty area
	r := s.Region(3, 0, 10, 2)
	if !r.IsEmpty() {
		t.Error("Region over empty area should be empty")
	}
}

func TestRegionFindText(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AABB\nCCDD\nEEFF")
	r := s.Region(1, 0, 4, 2)
	row, col := r.FindText("DD")
	if row != 0 || col != 2 {
		t.Errorf("Region.FindText(DD) = (%d,%d), want (0,2)", row, col)
	}
	row, col = r.FindText("ZZ")
	if row != -1 || col != -1 {
		t.Errorf("Region.FindText(ZZ) = (%d,%d), want (-1,-1)", row, col)
	}
}

func TestRegionString(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AB\nCD\nEF")
	r := s.Region(0, 0, 2, 3)
	got := r.String()
	if got != "AB\nCD\nEF" {
		t.Errorf("Region.String() = %q, want %q", got, "AB\nCD\nEF")
	}
}

func TestRegionAllRows(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AB\nCD")
	r := s.Region(0, 0, 2, 2)
	rows := r.AllRows()
	if len(rows) != 2 {
		t.Errorf("Region.AllRows() len = %d, want 2", len(rows))
	}
}

func TestRegionCountOccurrences(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("AAAA\nAAAA")
	r := s.Region(0, 0, 4, 2)
	if got := r.CountOccurrences("AA"); got != 4 {
		t.Errorf("Region.CountOccurrences(AA) = %d, want 4", got)
	}
}

func TestRegionStyleAt(t *testing.T) {
	s := NewScreen(40, 5)
	s.Render("Hello")
	r := s.Region(0, 0, 5, 1)
	style := r.StyleAt(0, 0)
	if style.Bold || style.Italic {
		t.Error("plain text region should have no bold/italic")
	}
}

// --- CellStyle tests ---

func TestCellStyleZeroValue(t *testing.T) {
	var cs CellStyle
	if cs.Fg != "" || cs.Bg != "" || cs.Bold || cs.Italic || cs.Underline || cs.Reverse {
		t.Error("zero CellStyle should have all default values")
	}
}
