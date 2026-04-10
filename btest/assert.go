package btest

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// AssertContains fails the test if the screen doesn't contain the text.
func AssertContains(t testing.TB, s *Screen, text string) {
	t.Helper()
	if !s.Contains(text) {
		t.Errorf("screen does not contain %q\nscreen content:\n%s", text, s.String())
	}
}

// AssertContainsAt fails the test if text doesn't appear at (row, col).
func AssertContainsAt(t testing.TB, s *Screen, row, col int, text string) {
	t.Helper()
	if !s.ContainsAt(row, col, text) {
		actual := s.TextAt(row, col, col+len(text))
		t.Errorf("expected %q at (%d, %d), got %q\nscreen content:\n%s",
			text, row, col, actual, s.String())
	}
}

// AssertNotContains fails the test if the screen contains the text.
func AssertNotContains(t testing.TB, s *Screen, text string) {
	t.Helper()
	if s.Contains(text) {
		t.Errorf("screen should not contain %q\nscreen content:\n%s", text, s.String())
	}
}

// AssertCursorAt fails the test if the cursor isn't at (row, col).
func AssertCursorAt(t testing.TB, s *Screen, row, col int) {
	t.Helper()
	r, c := s.CursorPos()
	if r != row || c != col {
		t.Errorf("expected cursor at (%d, %d), got (%d, %d)", row, col, r, c)
	}
}

// AssertStyleAt fails the test if the cell at (row, col) doesn't match the style.
func AssertStyleAt(t testing.TB, s *Screen, row, col int, style CellStyle) {
	t.Helper()
	actual := s.StyleAt(row, col)
	var mismatches []string
	if style.Fg != "" && actual.Fg != style.Fg {
		mismatches = append(mismatches, fmt.Sprintf("fg: want %q, got %q", style.Fg, actual.Fg))
	}
	if style.Bg != "" && actual.Bg != style.Bg {
		mismatches = append(mismatches, fmt.Sprintf("bg: want %q, got %q", style.Bg, actual.Bg))
	}
	if style.Bold != actual.Bold {
		mismatches = append(mismatches, fmt.Sprintf("bold: want %v, got %v", style.Bold, actual.Bold))
	}
	if style.Italic != actual.Italic {
		mismatches = append(mismatches, fmt.Sprintf("italic: want %v, got %v", style.Italic, actual.Italic))
	}
	if style.Underline != actual.Underline {
		mismatches = append(mismatches, fmt.Sprintf("underline: want %v, got %v", style.Underline, actual.Underline))
	}
	if style.Reverse != actual.Reverse {
		mismatches = append(mismatches, fmt.Sprintf("reverse: want %v, got %v", style.Reverse, actual.Reverse))
	}
	if len(mismatches) > 0 {
		t.Errorf("style mismatch at (%d, %d): %s", row, col, strings.Join(mismatches, "; "))
	}
}

// AssertBoldAt fails the test if the cell at (row, col) is not bold.
func AssertBoldAt(t testing.TB, s *Screen, row, col int) {
	t.Helper()
	style := s.StyleAt(row, col)
	if !style.Bold {
		t.Errorf("expected bold at (%d, %d), got non-bold", row, col)
	}
}

// AssertRowContains fails if the given row doesn't contain the text.
func AssertRowContains(t testing.TB, s *Screen, row int, text string) {
	t.Helper()
	rowText := s.Row(row)
	if !strings.Contains(rowText, text) {
		t.Errorf("row %d does not contain %q\nrow content: %q\nscreen content:\n%s",
			row, text, rowText, s.String())
	}
}

// AssertRowEquals fails if the given row doesn't exactly equal text (trimmed).
func AssertRowEquals(t testing.TB, s *Screen, row int, text string) {
	t.Helper()
	rowText := s.Row(row)
	if rowText != text {
		t.Errorf("row %d = %q, want %q\nscreen content:\n%s",
			row, rowText, text, s.String())
	}
}

// AssertRowNotContains fails if the given row contains the text.
func AssertRowNotContains(t testing.TB, s *Screen, row int, text string) {
	t.Helper()
	rowText := s.Row(row)
	if strings.Contains(rowText, text) {
		t.Errorf("row %d should not contain %q\nrow content: %q",
			row, text, rowText)
	}
}

// AssertRowCount fails if the number of non-empty rows doesn't match.
func AssertRowCount(t testing.TB, s *Screen, want int) {
	t.Helper()
	got := s.RowCount()
	if got != want {
		t.Errorf("RowCount() = %d, want %d\nscreen content:\n%s", got, want, s.String())
	}
}

// AssertEmpty fails if the screen has any visible content.
func AssertEmpty(t testing.TB, s *Screen) {
	t.Helper()
	if !s.IsEmpty() {
		t.Errorf("expected empty screen, got:\n%s", s.String())
	}
}

// AssertNotEmpty fails if the screen has no visible content.
func AssertNotEmpty(t testing.TB, s *Screen) {
	t.Helper()
	if s.IsEmpty() {
		t.Error("expected non-empty screen, got empty")
	}
}

// AssertMatches fails if the screen content doesn't match the regular expression.
func AssertMatches(t testing.TB, s *Screen, pattern string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regexp %q: %v", pattern, err)
	}
	if !re.MatchString(s.String()) {
		t.Errorf("screen does not match pattern %q\nscreen content:\n%s", pattern, s.String())
	}
}

// AssertRowMatches fails if the given row doesn't match the regular expression.
func AssertRowMatches(t testing.TB, s *Screen, row int, pattern string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regexp %q: %v", pattern, err)
	}
	rowText := s.Row(row)
	if !re.MatchString(rowText) {
		t.Errorf("row %d does not match pattern %q\nrow content: %q\nscreen content:\n%s",
			row, pattern, rowText, s.String())
	}
}

// AssertContainsCount fails if text doesn't appear exactly n times on screen.
func AssertContainsCount(t testing.TB, s *Screen, text string, n int) {
	t.Helper()
	got := s.CountOccurrences(text)
	if got != n {
		t.Errorf("screen contains %q %d times, want %d\nscreen content:\n%s",
			text, got, n, s.String())
	}
}

// AssertFgAt fails if the foreground color at (row, col) doesn't match.
func AssertFgAt(t testing.TB, s *Screen, row, col int, color string) {
	t.Helper()
	style := s.StyleAt(row, col)
	if style.Fg != color {
		t.Errorf("fg at (%d, %d) = %q, want %q", row, col, style.Fg, color)
	}
}

// AssertBgAt fails if the background color at (row, col) doesn't match.
func AssertBgAt(t testing.TB, s *Screen, row, col int, color string) {
	t.Helper()
	style := s.StyleAt(row, col)
	if style.Bg != color {
		t.Errorf("bg at (%d, %d) = %q, want %q", row, col, style.Bg, color)
	}
}

// AssertItalicAt fails if the cell at (row, col) is not italic.
func AssertItalicAt(t testing.TB, s *Screen, row, col int) {
	t.Helper()
	style := s.StyleAt(row, col)
	if !style.Italic {
		t.Errorf("expected italic at (%d, %d), got non-italic", row, col)
	}
}

// AssertUnderlineAt fails if the cell at (row, col) is not underlined.
func AssertUnderlineAt(t testing.TB, s *Screen, row, col int) {
	t.Helper()
	style := s.StyleAt(row, col)
	if !style.Underline {
		t.Errorf("expected underline at (%d, %d), got non-underline", row, col)
	}
}

// AssertReverseAt fails if the cell at (row, col) is not reversed.
func AssertReverseAt(t testing.TB, s *Screen, row, col int) {
	t.Helper()
	style := s.StyleAt(row, col)
	if !style.Reverse {
		t.Errorf("expected reverse at (%d, %d), got non-reverse", row, col)
	}
}

// AssertTextAt fails if the text at (row, col) through (row, col+len) doesn't match.
func AssertTextAt(t testing.TB, s *Screen, row, col int, text string) {
	t.Helper()
	actual := s.TextAt(row, col, col+len(text))
	if actual != text {
		t.Errorf("text at (%d, %d) = %q, want %q\nscreen content:\n%s",
			row, col, actual, text, s.String())
	}
}

// AssertRegionContains fails if the region doesn't contain the text.
func AssertRegionContains(t testing.TB, s *Screen, row, col, width, height int, text string) {
	t.Helper()
	r := s.Region(row, col, width, height)
	if !r.Contains(text) {
		t.Errorf("region (%d,%d %dx%d) does not contain %q\nregion content:\n%s",
			row, col, width, height, text, r.String())
	}
}

// AssertRegionNotContains fails if the region contains the text.
func AssertRegionNotContains(t testing.TB, s *Screen, row, col, width, height int, text string) {
	t.Helper()
	r := s.Region(row, col, width, height)
	if r.Contains(text) {
		t.Errorf("region (%d,%d %dx%d) should not contain %q\nregion content:\n%s",
			row, col, width, height, text, r.String())
	}
}

// AssertScreenEquals fails if the full screen text doesn't exactly match expected.
// On failure it also persists a FailureCapture so `blit diff` can replay it.
func AssertScreenEquals(t testing.TB, s *Screen, expected string) {
	t.Helper()
	actual := s.String()
	if actual != expected {
		SaveFailureCapture(t, FailureCapture{
			Kind:           FailureScreenEqual,
			ExpectedScreen: expected,
			ActualScreen:   actual,
		})
		t.Errorf("screen content mismatch\nwant:\n%s\ngot:\n%s\ndiff:\n%s",
			expected, actual, diffStrings(expected, actual))
	}
}
