package tuitest

import (
	"regexp"
	"strings"
	"testing"
)

// AssertRegionBold asserts every non-space cell in a region is bold.
func AssertRegionBold(t testing.TB, s *Screen, row, col, width, height int) {
	t.Helper()
	for r := row; r < row+height; r++ {
		for c := col; c < col+width; c++ {
			cell := s.TextAt(r, c, c+1)
			if cell == "" || cell == " " {
				continue
			}
			if !s.StyleAt(r, c).Bold {
				t.Errorf("AssertRegionBold: cell at (%d,%d)=%q not bold", r, c, cell)
				return
			}
		}
	}
}

// AssertRegionFg asserts every non-space cell in a region has the given fg color.
func AssertRegionFg(t testing.TB, s *Screen, row, col, width, height int, color string) {
	t.Helper()
	for r := row; r < row+height; r++ {
		for c := col; c < col+width; c++ {
			cell := s.TextAt(r, c, c+1)
			if cell == "" || cell == " " {
				continue
			}
			if got := s.StyleAt(r, c).Fg; got != color {
				t.Errorf("AssertRegionFg: cell at (%d,%d) fg=%q want %q", r, c, got, color)
				return
			}
		}
	}
}

// AssertRegionBg asserts every non-space cell in a region has the given bg color.
func AssertRegionBg(t testing.TB, s *Screen, row, col, width, height int, color string) {
	t.Helper()
	for r := row; r < row+height; r++ {
		for c := col; c < col+width; c++ {
			cell := s.TextAt(r, c, c+1)
			if cell == "" {
				continue
			}
			if got := s.StyleAt(r, c).Bg; got != color {
				t.Errorf("AssertRegionBg: cell at (%d,%d) bg=%q want %q", r, c, got, color)
				return
			}
		}
	}
}

// AssertColumnContains asserts any row in column range [startRow,endRow] contains text.
func AssertColumnContains(t testing.TB, s *Screen, col, startRow, endRow int, text string) {
	t.Helper()
	got := s.Column(col, startRow, endRow)
	if !strings.Contains(got, text) {
		t.Errorf("AssertColumnContains(col=%d rows=%d..%d, %q): got %q",
			col, startRow, endRow, text, got)
	}
}

// AssertColumnCount asserts how many non-empty rows contain text in the given column range.
func AssertColumnCount(t testing.TB, s *Screen, col, startRow, endRow int, text string, want int) {
	t.Helper()
	col_ := s.Column(col, startRow, endRow)
	got := strings.Count(col_, text)
	if got != want {
		t.Errorf("AssertColumnCount(col=%d, %q): got %d want %d", col, text, got, want)
	}
}

// AssertCursorRowContains asserts the row under the cursor contains text.
func AssertCursorRowContains(t testing.TB, s *Screen, text string) {
	t.Helper()
	row, _ := s.CursorPos()
	if row < 0 {
		t.Errorf("AssertCursorRowContains: cursor not positioned")
		return
	}
	got := s.Row(row)
	if !strings.Contains(got, text) {
		t.Errorf("AssertCursorRowContains(%q): cursor row %d = %q", text, row, got)
	}
}

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// AssertNoANSI asserts the screen's rendered text contains no raw ANSI escape sequences.
// Raw ANSI in the decoded virtual screen indicates a broken writer that emitted
// literal bytes instead of styled runs.
func AssertNoANSI(t testing.TB, s *Screen) {
	t.Helper()
	if ansiRE.MatchString(s.String()) {
		t.Errorf("AssertNoANSI: screen contains raw ANSI escape sequences:\n%s", s.String())
	}
}

// AssertKeybind asserts the screen's footer/help line contains the given key label.
// It searches the whole screen for a "key  description" pattern anywhere.
func AssertKeybind(t testing.TB, s *Screen, key, description string) {
	t.Helper()
	text := s.String()
	if !strings.Contains(text, key) {
		t.Errorf("AssertKeybind: key label %q not found on screen", key)
		return
	}
	if description != "" && !strings.Contains(text, description) {
		t.Errorf("AssertKeybind: description %q not found for key %q", description, key)
	}
}

// AssertScreenMatches is a regexp variant of AssertScreenEquals for fuzzier matches.
func AssertScreenMatches(t testing.TB, s *Screen, pattern string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Errorf("AssertScreenMatches: bad pattern %q: %v", pattern, err)
		return
	}
	if !re.MatchString(s.String()) {
		t.Errorf("AssertScreenMatches(%q): screen did not match:\n%s", pattern, s.String())
	}
}
