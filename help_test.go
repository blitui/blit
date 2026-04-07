package tuikit

import (
	"strings"
	"testing"
)

func TestHelpRender(t *testing.T) {
	reg := newRegistry()
	reg.addBindings("table", []KeyBind{
		{Key: "up", Label: "Move up", Group: "NAVIGATION"},
		{Key: "s", Label: "Sort", Group: "DATA"},
	})
	reg.addBindings("global", []KeyBind{
		{Key: "q", Label: "Quit", Group: "OTHER"},
	})

	h := NewHelp()
	h.setRegistry(reg)
	h.SetTheme(DefaultTheme())
	h.SetSize(80, 24)
	h.active = true

	view := h.View()
	if !strings.Contains(view, "NAVIGATION") {
		t.Error("help should show NAVIGATION group")
	}
	if !strings.Contains(view, "DATA") {
		t.Error("help should show DATA group")
	}
	if !strings.Contains(view, "Move up") {
		t.Error("help should show 'Move up' label")
	}
	if !strings.Contains(view, "Quit") {
		t.Error("help should show 'Quit' label")
	}
}

func TestHelpOverlayInterface(t *testing.T) {
	h := NewHelp()
	h.active = true
	if !h.IsActive() {
		t.Error("should be active")
	}
	h.Close()
	if h.IsActive() {
		t.Error("should be inactive after Close")
	}
}
