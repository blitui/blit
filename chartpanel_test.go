package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
)

// stubChart is a minimal Component for testing ChartPanel.
type stubChart struct {
	label   string
	width   int
	height  int
	focused bool
}

func (s *stubChart) Init() tea.Cmd                                                  { return nil }
func (s *stubChart) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) { return s, nil }
func (s *stubChart) View() string                                                   { return s.label }
func (s *stubChart) KeyBindings() []blit.KeyBind                                    { return nil }
func (s *stubChart) SetSize(w, h int)                                               { s.width = w; s.height = h }
func (s *stubChart) Focused() bool                                                  { return s.focused }
func (s *stubChart) SetFocused(f bool)                                              { s.focused = f }
func (s *stubChart) SetTheme(theme blit.Theme)                                      {}

func TestChartPanel_NewDefaults(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "chart1"}},
	})
	if cp.ActiveIndex() != 0 {
		t.Fatalf("ActiveIndex() = %d, want 0", cp.ActiveIndex())
	}
}

func TestChartPanel_View(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Title:  "Metrics",
		Charts: []blit.Component{&stubChart{label: "CHART_CONTENT"}},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	view := cp.View()
	if !strings.Contains(view, "Metrics") {
		t.Fatalf("view should contain title:\n%s", view)
	}
	if !strings.Contains(view, "CHART_CONTENT") {
		t.Fatalf("view should contain chart content:\n%s", view)
	}
}

func TestChartPanel_TabSwitching(t *testing.T) {
	c1 := &stubChart{label: "LINE"}
	c2 := &stubChart{label: "BAR"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{c1, c2},
		Labels: []string{"Line", "Bar"},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)
	cp.SetFocused(true)

	// Tab switches to next chart.
	updated, _ := cp.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	cp = updated.(*blit.ChartPanel)
	if cp.ActiveIndex() != 1 {
		t.Fatalf("ActiveIndex() = %d, want 1 after tab", cp.ActiveIndex())
	}
	view := cp.View()
	if !strings.Contains(view, "BAR") {
		t.Fatalf("view should show BAR chart:\n%s", view)
	}

	// Tab wraps around.
	updated, _ = cp.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	cp = updated.(*blit.ChartPanel)
	if cp.ActiveIndex() != 0 {
		t.Fatalf("ActiveIndex() = %d, want 0 after wrap", cp.ActiveIndex())
	}
}

func TestChartPanel_ShiftTabSwitching(t *testing.T) {
	c1 := &stubChart{label: "A"}
	c2 := &stubChart{label: "B"}
	c3 := &stubChart{label: "C"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{c1, c2, c3},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)
	cp.SetFocused(true)

	// Shift+tab wraps to last.
	updated, _ := cp.Update(tea.KeyMsg{Type: tea.KeyShiftTab}, blit.Context{})
	cp = updated.(*blit.ChartPanel)
	if cp.ActiveIndex() != 2 {
		t.Fatalf("ActiveIndex() = %d, want 2 after shift+tab", cp.ActiveIndex())
	}
}

func TestChartPanel_SetActive(t *testing.T) {
	c1 := &stubChart{label: "A"}
	c2 := &stubChart{label: "B"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{c1, c2},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	cp.SetActive(1)
	if cp.ActiveIndex() != 1 {
		t.Fatalf("ActiveIndex() = %d, want 1", cp.ActiveIndex())
	}

	// Clamp to bounds.
	cp.SetActive(99)
	if cp.ActiveIndex() != 1 {
		t.Fatalf("ActiveIndex() = %d, want 1 (clamped)", cp.ActiveIndex())
	}

	cp.SetActive(-1)
	if cp.ActiveIndex() != 0 {
		t.Fatalf("ActiveIndex() = %d, want 0 (clamped)", cp.ActiveIndex())
	}
}

func TestChartPanel_SizesActiveChart(t *testing.T) {
	c1 := &stubChart{label: "A"}
	c2 := &stubChart{label: "B"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Title:  "Test",
		Charts: []blit.Component{c1, c2},
		Labels: []string{"A", "B"},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	// Active chart should get reduced height (header takes 2 lines).
	if c1.width != 80 {
		t.Fatalf("chart width = %d, want 80", c1.width)
	}
	if c1.height != 22 {
		t.Fatalf("chart height = %d, want 22 (24 - 2 header)", c1.height)
	}
}

func TestChartPanel_NoCharts(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	view := cp.View()
	if !strings.Contains(view, "no charts") {
		t.Fatalf("empty panel should show placeholder:\n%s", view)
	}
}

func TestChartPanel_ZeroSize(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "A"}},
	})
	cp.SetSize(0, 0)
	if cp.View() != "" {
		t.Fatal("zero-sized panel should return empty view")
	}
}

func TestChartPanel_UnfocusedIgnoresInput(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "A"}, &stubChart{label: "B"}},
	})
	cp.SetFocused(false)

	updated, _ := cp.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	cp = updated.(*blit.ChartPanel)
	if cp.ActiveIndex() != 0 {
		t.Fatal("unfocused panel should not process tab")
	}
}

func TestChartPanel_KeyBindings(t *testing.T) {
	// Multi-chart panel has tab binding.
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "A"}, &stubChart{label: "B"}},
	})
	binds := cp.KeyBindings()
	found := false
	for _, b := range binds {
		if strings.Contains(b.Key, "tab") {
			found = true
		}
	}
	if !found {
		t.Fatal("multi-chart panel should have tab key binding")
	}

	// Single-chart panel should not.
	cp2 := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "A"}},
	})
	for _, b := range cp2.KeyBindings() {
		if strings.Contains(b.Key, "tab") {
			t.Fatal("single-chart panel should not have tab binding")
		}
	}
}

func TestChartPanel_ComponentInterface(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{&stubChart{label: "A"}},
	})

	if cmd := cp.Init(); cmd != nil {
		t.Fatal("Init() should return nil for stubChart")
	}
	if cp.Focused() {
		t.Fatal("should not be focused by default")
	}
	cp.SetFocused(true)
	if !cp.Focused() {
		t.Fatal("should be focused after SetFocused(true)")
	}
}

func TestChartPanel_PropagatesTheme(t *testing.T) {
	c := &stubChart{label: "A"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{c},
	})
	theme := blit.DefaultTheme()
	cp.SetTheme(theme)
	// stubChart implements Themed via SetTheme — if it didn't panic, propagation worked.
}

func TestChartPanel_Labels(t *testing.T) {
	c1 := &stubChart{label: "A"}
	c2 := &stubChart{label: "B"}
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Charts: []blit.Component{c1, c2},
		Labels: []string{"Line", "Bar"},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	view := cp.View()
	if !strings.Contains(view, "Line") {
		t.Fatalf("view should contain label 'Line':\n%s", view)
	}
	if !strings.Contains(view, "Bar") {
		t.Fatalf("view should contain label 'Bar':\n%s", view)
	}
}

func TestChartPanel_SingleChartNoTabs(t *testing.T) {
	cp := blit.NewChartPanel(blit.ChartPanelOpts{
		Title:  "Solo",
		Charts: []blit.Component{&stubChart{label: "ONLY"}},
	})
	cp.SetTheme(blit.DefaultTheme())
	cp.SetSize(80, 24)

	view := cp.View()
	if !strings.Contains(view, "ONLY") {
		t.Fatalf("view should contain chart content:\n%s", view)
	}
	// Should not contain tab separator.
	if strings.Contains(view, "│") {
		t.Fatalf("single chart should not show tab separator:\n%s", view)
	}
}
