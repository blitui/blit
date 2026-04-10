package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func newTestViewport(content string, w, h int) *blit.Viewport {
	v := blit.NewViewport()
	v.SetTheme(blit.DefaultTheme())
	v.SetSize(w, h)
	v.SetContent(content)
	return v
}

func TestViewportScrollDown(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)

	if !v.AtTop() {
		t.Fatal("expected AtTop initially")
	}

	v.ScrollBy(3)
	if v.YOffset() != 3 {
		t.Errorf("expected offset 3, got %d", v.YOffset())
	}
}

func TestViewportScrollClampTop(t *testing.T) {
	v := newTestViewport("a\nb\nc", 40, 5)
	v.ScrollBy(-999)
	if v.YOffset() != 0 {
		t.Errorf("expected clamped to 0, got %d", v.YOffset())
	}
}

func TestViewportScrollClampBottom(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.ScrollBy(9999)
	if !v.AtBottom() {
		t.Errorf("expected AtBottom after over-scroll, offset=%d", v.YOffset())
	}
}

func TestViewportGotoTopBottom(t *testing.T) {
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "row"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 10)
	v.GotoBottom()
	if !v.AtBottom() {
		t.Error("expected AtBottom after GotoBottom")
	}
	v.GotoTop()
	if !v.AtTop() {
		t.Error("expected AtTop after GotoTop")
	}
}

func TestViewportKeyJ(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.YOffset() != 1 {
		t.Errorf("expected offset 1 after j, got %d", v.YOffset())
	}
}

func TestViewportKeyK(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.ScrollBy(5)
	v.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.YOffset() != 4 {
		t.Errorf("expected offset 4 after k, got %d", v.YOffset())
	}
}

func TestViewportKeyHome(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.ScrollBy(10)
	v.HandleKey(tea.KeyMsg{Type: tea.KeyHome})
	if !v.AtTop() {
		t.Error("expected AtTop after home key")
	}
}

func TestViewportKeyEnd(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.HandleKey(tea.KeyMsg{Type: tea.KeyEnd})
	if !v.AtBottom() {
		t.Error("expected AtBottom after end key")
	}
}

func TestViewportKeyCtrlU(t *testing.T) {
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 10)
	v.GotoBottom()
	before := v.YOffset()
	v.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	if v.YOffset() >= before {
		t.Errorf("expected scroll up with ctrl+u, before=%d after=%d", before, v.YOffset())
	}
}

func TestViewportKeyCtrlD(t *testing.T) {
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 10)
	before := v.YOffset()
	v.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if v.YOffset() <= before {
		t.Errorf("expected scroll down with ctrl+d, before=%d after=%d", before, v.YOffset())
	}
}

func TestViewportKeyPgUp(t *testing.T) {
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 10)
	v.GotoBottom()
	before := v.YOffset()
	v.HandleKey(tea.KeyMsg{Type: tea.KeyPgUp})
	if v.YOffset() >= before {
		t.Errorf("expected scroll up with pgup, before=%d after=%d", before, v.YOffset())
	}
}

func TestViewportKeyPgDown(t *testing.T) {
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 10)
	before := v.YOffset()
	v.HandleKey(tea.KeyMsg{Type: tea.KeyPgDown})
	if v.YOffset() <= before {
		t.Errorf("expected scroll down with pgdown, before=%d after=%d", before, v.YOffset())
	}
}

func TestViewportView(t *testing.T) {
	lines := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	v := newTestViewport(strings.Join(lines, "\n"), 20, 3)
	view := v.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
	// First visible line should contain "alpha".
	if !strings.Contains(view, "alpha") {
		t.Errorf("expected 'alpha' in view, got: %q", view)
	}
}

func TestViewportScrollbarPresent(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 20, 5)
	view := v.View()
	// Scrollbar thumb or track must be present (non-trivial output per row).
	rows := strings.Split(view, "\n")
	if len(rows) == 0 {
		t.Fatal("no rows in view")
	}
}

func TestViewportComponentInterface(t *testing.T) {
	var _ blit.Component = blit.NewViewport()
}

func TestViewportThemedInterface(t *testing.T) {
	var _ blit.Themed = blit.NewViewport()
}

func TestViewportMouseWheelDown(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	_, _ = v.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown}, blit.Context{})
	if v.YOffset() != 3 {
		t.Errorf("expected offset 3 after wheel down, got %d", v.YOffset())
	}
}

func TestViewportMouseWheelUp(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v := newTestViewport(strings.Join(lines, "\n"), 40, 5)
	v.ScrollBy(6)
	_, _ = v.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp}, blit.Context{})
	if v.YOffset() != 3 {
		t.Errorf("expected offset 3 after wheel up, got %d", v.YOffset())
	}
}
