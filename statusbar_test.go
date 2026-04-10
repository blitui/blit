package blit

import (
	"strings"
	"testing"
)

func TestStatusBarView(t *testing.T) {
	bar := NewStatusBar(StatusBarOpts{
		Left:  func() string { return "LEFT" },
		Right: func() string { return "RIGHT" },
	})
	bar.SetSize(40, 1)
	bar.SetTheme(DefaultTheme())
	view := bar.View()
	if !strings.Contains(view, "LEFT") {
		t.Error("view should contain LEFT")
	}
	if !strings.Contains(view, "RIGHT") {
		t.Error("view should contain RIGHT")
	}
}

func TestStatusBarOverflow(t *testing.T) {
	bar := NewStatusBar(StatusBarOpts{
		Left:  func() string { return "? help  / search  p panel  q quit  •  42 pairs" },
		Right: func() string { return "  1/42  BTC 67,432.10  •  15:04:05 ● connected " },
	})
	bar.SetSize(60, 1) // too narrow for both
	bar.SetTheme(DefaultTheme())
	view := bar.View()
	lines := strings.Split(view, "\n")
	if len(lines) != 1 {
		t.Errorf("status bar should always be 1 line, got %d", len(lines))
	}
}

func TestStatusBarNilFuncs(t *testing.T) {
	bar := NewStatusBar(StatusBarOpts{})
	bar.SetSize(40, 1)
	bar.SetTheme(DefaultTheme())
	view := bar.View()
	if view == "" {
		t.Error("view should not be empty even with nil funcs")
	}
}
