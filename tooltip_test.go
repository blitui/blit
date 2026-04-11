package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
)

func makeTestTooltip() *blit.Tooltip {
	tt := blit.NewTooltip(blit.TooltipOpts{Text: "Hint text"})
	tt.SetTheme(blit.DefaultTheme())
	tt.SetSize(80, 24)
	return tt
}

func TestTooltip_NewDefaults(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{Text: "Hello"})
	if tt.IsActive() {
		t.Fatal("tooltip should not be active by default")
	}
	if tt.Text() != "Hello" {
		t.Fatalf("Text() = %q, want Hello", tt.Text())
	}
}

func TestTooltip_SetText(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{})
	tt.SetText("Updated")
	if tt.Text() != "Updated" {
		t.Fatalf("Text() = %q, want Updated", tt.Text())
	}
}

func TestTooltip_ShowClose(t *testing.T) {
	tt := makeTestTooltip()
	tt.Show()
	if !tt.IsActive() {
		t.Fatal("should be active after Show()")
	}
	tt.Close()
	if tt.IsActive() {
		t.Fatal("should be inactive after Close()")
	}
}

func TestTooltip_ViewInactive(t *testing.T) {
	tt := makeTestTooltip()
	if tt.View() != "" {
		t.Fatal("inactive tooltip should return empty view")
	}
}

func TestTooltip_ViewActive(t *testing.T) {
	tt := makeTestTooltip()
	tt.Show()
	view := tt.View()
	if !strings.Contains(view, "Hint text") {
		t.Fatalf("active tooltip should contain text:\n%s", view)
	}
}

func TestTooltip_ViewEmptyText(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{Text: ""})
	tt.SetTheme(blit.DefaultTheme())
	tt.SetSize(80, 24)
	tt.Show()
	if tt.View() != "" {
		t.Fatal("tooltip with empty text should return empty view")
	}
}

func TestTooltip_EscDismisses(t *testing.T) {
	tt := makeTestTooltip()
	tt.Show()

	updated, _ := tt.Update(tea.KeyMsg{Type: tea.KeyEsc}, blit.Context{})
	tt = updated.(*blit.Tooltip)

	if tt.IsActive() {
		t.Fatal("esc should dismiss tooltip")
	}
}

func TestTooltip_InactiveIgnoresInput(t *testing.T) {
	tt := makeTestTooltip()

	updated, _ := tt.Update(tea.KeyMsg{Type: tea.KeyEsc}, blit.Context{})
	tt = updated.(*blit.Tooltip)

	// Should not panic or error.
	if tt.IsActive() {
		t.Fatal("inactive tooltip should remain inactive")
	}
}

func TestTooltip_FloatView(t *testing.T) {
	tt := makeTestTooltip()
	tt.Show()
	tt.SetAnchor(5, 2)

	bg := "Line 0\nLine 1\nLine 2\nLine 3\nLine 4"
	result := tt.FloatView(bg)

	if !strings.Contains(result, "Hint text") {
		t.Fatalf("FloatView should composite tooltip:\n%s", result)
	}
	// Background lines should still be present.
	if !strings.Contains(result, "Line 0") {
		t.Fatalf("FloatView should preserve background:\n%s", result)
	}
}

func TestTooltip_FloatViewInactive(t *testing.T) {
	tt := makeTestTooltip()
	bg := "Background"
	result := tt.FloatView(bg)
	if result != bg {
		t.Fatal("inactive FloatView should return background unchanged")
	}
}

func TestTooltip_SetAnchor(t *testing.T) {
	tt := makeTestTooltip()
	tt.SetAnchor(10, 5)
	tt.Show()

	bg := strings.Repeat(".\n", 10)
	result := tt.FloatView(bg)

	// The tooltip should be placed starting at row 5.
	lines := strings.Split(result, "\n")
	if len(lines) <= 5 {
		t.Fatal("FloatView should have enough lines for anchor position")
	}
}

func TestTooltip_ComponentInterface(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{})
	if cmd := tt.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	if tt.Focused() {
		t.Fatal("should not be focused by default")
	}
	tt.SetFocused(true)
	if !tt.Focused() {
		t.Fatal("should be focused")
	}
}

func TestTooltip_KeyBindings(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{})
	binds := tt.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestTooltip_CustomMaxWidth(t *testing.T) {
	tt := blit.NewTooltip(blit.TooltipOpts{
		Text:     "Short",
		MaxWidth: 20,
	})
	tt.SetTheme(blit.DefaultTheme())
	tt.SetSize(80, 24)
	tt.Show()
	view := tt.View()
	if !strings.Contains(view, "Short") {
		t.Fatalf("custom maxwidth tooltip should render text:\n%s", view)
	}
}
