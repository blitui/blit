package tuitest

import (
	"fmt"
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
