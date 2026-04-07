package tuikit

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

func TestStatusBarNilFuncs(t *testing.T) {
	bar := NewStatusBar(StatusBarOpts{})
	bar.SetSize(40, 1)
	bar.SetTheme(DefaultTheme())
	view := bar.View()
	if view == "" {
		t.Error("view should not be empty even with nil funcs")
	}
}
