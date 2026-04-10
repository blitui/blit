package btest

import (
	"testing"
)

func TestAssertColumnContains(t *testing.T) {
	s := NewScreen(20, 5)
	s.Render("hello world\r\nfoo bar baz\r\n")
	// Column 0 should contain 'h' and 'f'
	AssertColumnContains(t, s, 0, 0, 2, "h")
	AssertColumnContains(t, s, 0, 0, 2, "f")
}

func TestAssertNoANSI(t *testing.T) {
	s := NewScreen(20, 2)
	s.Render("plain text\r\n")
	AssertNoANSI(t, s)
}

func TestAssertKeybind(t *testing.T) {
	s := NewScreen(40, 2)
	s.Render("q quit  ? help\r\n")
	AssertKeybind(t, s, "q", "quit")
	AssertKeybind(t, s, "?", "help")
}

func TestAssertScreenMatches(t *testing.T) {
	s := NewScreen(20, 2)
	s.Render("version v1.2.3\r\n")
	AssertScreenMatches(t, s, `v\d+\.\d+\.\d+`)
}

func TestAssertCursorRowContains(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("line0\r\nline1\r\nline2")
	// cursor is on last line after render
	row, _ := s.CursorPos()
	if row >= 0 {
		AssertCursorRowContains(t, s, "line")
	}
}

func TestAssertColumnCount(t *testing.T) {
	s := NewScreen(20, 5)
	s.Render("a\r\na\r\nb\r\na\r\n")
	AssertColumnCount(t, s, 0, 0, 4, "a", 3)
	AssertColumnCount(t, s, 0, 0, 4, "b", 1)
}
