package charts_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
	"github.com/blitui/blit/charts"
)

// ---------- Bar ----------

func TestBar_Init(t *testing.T) {
	b := charts.NewBar([]float64{1, 2}, nil, false)
	if cmd := b.Init(); cmd != nil {
		t.Fatal("Bar.Init() should return nil")
	}
}

func TestBar_Update(t *testing.T) {
	b := charts.NewBar([]float64{1, 2}, nil, false)
	b.SetTheme(blit.DefaultTheme())
	updated, cmd := b.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	if cmd != nil {
		t.Fatal("Bar.Update should return nil cmd")
	}
	if updated != b {
		t.Fatal("Bar.Update should return same instance")
	}
}

func TestBar_KeyBindings(t *testing.T) {
	b := charts.NewBar([]float64{1}, nil, false)
	if kb := b.KeyBindings(); kb != nil {
		t.Fatal("Bar.KeyBindings() should return nil")
	}
}

func TestBar_FocusedSetFocused(t *testing.T) {
	b := charts.NewBar([]float64{1}, nil, false)
	if b.Focused() {
		t.Fatal("new Bar should not be focused")
	}
	b.SetFocused(true)
	if !b.Focused() {
		t.Fatal("Bar should be focused after SetFocused(true)")
	}
}

func TestBar_SetTheme(t *testing.T) {
	b := charts.NewBar([]float64{1, 2, 3}, nil, false)
	b.SetSize(20, 10)
	theme := blit.DefaultTheme()
	theme.Accent = lipgloss.Color("#123456")
	b.SetTheme(theme)
	out := b.View()
	if out == "" {
		t.Fatal("should render after SetTheme")
	}
}

func TestBar_HorizontalGradient(t *testing.T) {
	g := &blit.Gradient{
		Start: lipgloss.Color("#ff0000"),
		End:   lipgloss.Color("#0000ff"),
	}
	b := charts.NewBar([]float64{10, 50, 100}, []string{"a", "b", "c"}, true)
	b.Gradient = g
	b.SetSize(30, 5)
	b.SetTheme(blit.DefaultTheme())
	out := b.View()
	if out == "" {
		t.Fatal("horizontal gradient bar should render")
	}
}

func TestBar_AllZeroData(t *testing.T) {
	b := charts.NewBar([]float64{0, 0, 0}, nil, false)
	b.SetSize(20, 10)
	b.SetTheme(blit.DefaultTheme())
	out := b.View()
	if out == "" {
		t.Fatal("all-zero data should still render")
	}
}

func TestBar_SingleValue(t *testing.T) {
	b := charts.NewBar([]float64{42}, []string{"x"}, false)
	b.SetSize(10, 8)
	b.SetTheme(blit.DefaultTheme())
	out := b.View()
	if out == "" {
		t.Fatal("single value bar should render")
	}
}

// ---------- Gauge ----------

func TestGauge_Init(t *testing.T) {
	g := charts.NewGauge(50, 100, nil, "")
	if cmd := g.Init(); cmd != nil {
		t.Fatal("Gauge.Init() should return nil")
	}
}

func TestGauge_Update(t *testing.T) {
	g := charts.NewGauge(50, 100, nil, "")
	g.SetTheme(blit.DefaultTheme())
	updated, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	if cmd != nil {
		t.Fatal("Gauge.Update should return nil cmd")
	}
	if updated != g {
		t.Fatal("Gauge.Update should return same instance")
	}
}

func TestGauge_KeyBindings(t *testing.T) {
	g := charts.NewGauge(50, 100, nil, "")
	if kb := g.KeyBindings(); kb != nil {
		t.Fatal("Gauge.KeyBindings() should return nil")
	}
}

func TestGauge_FocusedSetFocused(t *testing.T) {
	g := charts.NewGauge(50, 100, nil, "")
	if g.Focused() {
		t.Fatal("new Gauge should not be focused")
	}
	g.SetFocused(true)
	if !g.Focused() {
		t.Fatal("Gauge should be focused after SetFocused(true)")
	}
}

func TestGauge_SetTheme(t *testing.T) {
	g := charts.NewGauge(50, 100, []float64{60, 80}, "Test")
	g.SetSize(30, 10)
	theme := blit.DefaultTheme()
	theme.Positive = lipgloss.Color("#00ff00")
	g.SetTheme(theme)
	out := g.View()
	if out == "" {
		t.Fatal("should render after SetTheme")
	}
}

func TestGauge_OverMax(t *testing.T) {
	g := charts.NewGauge(150, 100, nil, "")
	g.SetSize(30, 10)
	g.SetTheme(blit.DefaultTheme())
	out := g.View()
	if out == "" {
		t.Fatal("over-max gauge should still render")
	}
}

func TestGauge_ZeroMax(t *testing.T) {
	g := charts.NewGauge(50, 0, nil, "")
	g.SetSize(30, 10)
	g.SetTheme(blit.DefaultTheme())
	out := g.View()
	if out == "" {
		t.Fatal("zero-max gauge should still render")
	}
}

// ---------- Heatmap ----------

func TestHeatmap_Init(t *testing.T) {
	h := charts.NewHeatmap([][]float64{{1}}, charts.PaletteSequential)
	if cmd := h.Init(); cmd != nil {
		t.Fatal("Heatmap.Init() should return nil")
	}
}

func TestHeatmap_Update(t *testing.T) {
	h := charts.NewHeatmap([][]float64{{1}}, charts.PaletteSequential)
	h.SetTheme(blit.DefaultTheme())
	updated, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	if cmd != nil {
		t.Fatal("Heatmap.Update should return nil cmd")
	}
	if updated != h {
		t.Fatal("Heatmap.Update should return same instance")
	}
}

func TestHeatmap_KeyBindings(t *testing.T) {
	h := charts.NewHeatmap([][]float64{{1}}, charts.PaletteSequential)
	if kb := h.KeyBindings(); kb != nil {
		t.Fatal("Heatmap.KeyBindings() should return nil")
	}
}

func TestHeatmap_FocusedSetFocused(t *testing.T) {
	h := charts.NewHeatmap([][]float64{{1}}, charts.PaletteSequential)
	if h.Focused() {
		t.Fatal("new Heatmap should not be focused")
	}
	h.SetFocused(true)
	if !h.Focused() {
		t.Fatal("Heatmap should be focused after SetFocused(true)")
	}
}

func TestHeatmap_SetTheme(t *testing.T) {
	h := charts.NewHeatmap([][]float64{{1, 2}, {3, 4}}, charts.PaletteSequential)
	h.SetSize(20, 6)
	theme := blit.DefaultTheme()
	theme.Accent = lipgloss.Color("#abcdef")
	h.SetTheme(theme)
	out := h.View()
	if out == "" {
		t.Fatal("should render after SetTheme")
	}
}

func TestHeatmap_UniformValues(t *testing.T) {
	grid := [][]float64{{5, 5}, {5, 5}}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.SetSize(20, 6)
	h.SetTheme(blit.DefaultTheme())
	out := h.View()
	if out == "" {
		t.Fatal("uniform values should still render")
	}
}

func TestHeatmap_EmptyRows(t *testing.T) {
	grid := [][]float64{{}, {}}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.SetSize(20, 6)
	out := h.View()
	if out != "" {
		t.Fatal("empty row grid should return empty view")
	}
}

// ---------- Line ----------

func TestLine_Init(t *testing.T) {
	l := charts.NewLine([][]float64{{1, 2}}, nil, false)
	if cmd := l.Init(); cmd != nil {
		t.Fatal("Line.Init() should return nil")
	}
}

func TestLine_Update(t *testing.T) {
	l := charts.NewLine([][]float64{{1, 2}}, nil, false)
	l.SetTheme(blit.DefaultTheme())
	updated, cmd := l.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	if cmd != nil {
		t.Fatal("Line.Update should return nil cmd")
	}
	if updated != l {
		t.Fatal("Line.Update should return same instance")
	}
}

func TestLine_KeyBindings(t *testing.T) {
	l := charts.NewLine([][]float64{{1}}, nil, false)
	if kb := l.KeyBindings(); kb != nil {
		t.Fatal("Line.KeyBindings() should return nil")
	}
}

func TestLine_FocusedSetFocused(t *testing.T) {
	l := charts.NewLine([][]float64{{1}}, nil, false)
	if l.Focused() {
		t.Fatal("new Line should not be focused")
	}
	l.SetFocused(true)
	if !l.Focused() {
		t.Fatal("Line should be focused after SetFocused(true)")
	}
}

func TestLine_SetTheme(t *testing.T) {
	l := charts.NewLine([][]float64{{1, 5, 3}}, nil, false)
	l.SetSize(30, 8)
	theme := blit.DefaultTheme()
	theme.Accent = lipgloss.Color("#ff00ff")
	l.SetTheme(theme)
	out := l.View()
	if out == "" {
		t.Fatal("should render after SetTheme")
	}
}

func TestLine_EmptySeries(t *testing.T) {
	l := charts.NewLine([][]float64{{}}, nil, false)
	l.SetSize(30, 8)
	l.SetTheme(blit.DefaultTheme())
	// Empty inner series should not panic.
	_ = l.View()
}

func TestLine_MultiSeriesDefaultColors(t *testing.T) {
	series := [][]float64{
		{1, 2, 3},
		{3, 2, 1},
		{2, 3, 1},
		{1, 1, 3},
		{3, 3, 1},
		{2, 1, 2},
	}
	l := charts.NewLine(series, nil, false)
	l.SetSize(40, 10)
	l.SetTheme(blit.DefaultTheme())
	out := l.View()
	if out == "" {
		t.Fatal("multi-series with default colors should render")
	}
}

func TestLine_ZeroSize(t *testing.T) {
	l := charts.NewLine([][]float64{{1, 2, 3}}, nil, false)
	l.SetSize(1, 1)
	out := l.View()
	if out != "" {
		t.Fatal("too-small line chart should return empty")
	}
}

func TestLine_FlatData(t *testing.T) {
	l := charts.NewLine([][]float64{{5, 5, 5, 5}}, nil, false)
	l.SetSize(30, 8)
	l.SetTheme(blit.DefaultTheme())
	out := l.View()
	if out == "" {
		t.Fatal("flat data line chart should render")
	}
}

// ---------- Ring ----------

func TestRing_Init(t *testing.T) {
	r := charts.NewRing(50, 100, "")
	if cmd := r.Init(); cmd != nil {
		t.Fatal("Ring.Init() should return nil")
	}
}

func TestRing_Update(t *testing.T) {
	r := charts.NewRing(50, 100, "")
	r.SetTheme(blit.DefaultTheme())
	updated, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	if cmd != nil {
		t.Fatal("Ring.Update should return nil cmd")
	}
	if updated != r {
		t.Fatal("Ring.Update should return same instance")
	}
}

func TestRing_KeyBindings(t *testing.T) {
	r := charts.NewRing(50, 100, "")
	if kb := r.KeyBindings(); kb != nil {
		t.Fatal("Ring.KeyBindings() should return nil")
	}
}

func TestRing_FocusedSetFocused(t *testing.T) {
	r := charts.NewRing(50, 100, "")
	if r.Focused() {
		t.Fatal("new Ring should not be focused")
	}
	r.SetFocused(true)
	if !r.Focused() {
		t.Fatal("Ring should be focused after SetFocused(true)")
	}
}

func TestRing_SetTheme(t *testing.T) {
	r := charts.NewRing(50, 100, "CPU")
	r.SetSize(20, 10)
	theme := blit.DefaultTheme()
	theme.Accent = lipgloss.Color("#00ffff")
	r.SetTheme(theme)
	out := r.View()
	if out == "" {
		t.Fatal("should render after SetTheme")
	}
}

func TestRing_CustomColors(t *testing.T) {
	r := charts.NewRing(75, 100, "MEM")
	r.FillColor = lipgloss.Color("#ff0000")
	r.TrackColor = lipgloss.Color("#333333")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	if out == "" {
		t.Fatal("ring with custom colors should render")
	}
}

func TestRing_ZeroMax(t *testing.T) {
	r := charts.NewRing(50, 0, "X")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	if out == "" {
		t.Fatal("zero-max ring should still render")
	}
}

func TestRing_NoLabel(t *testing.T) {
	r := charts.NewRing(50, 100, "")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	if out == "" {
		t.Fatal("ring without label should render")
	}
}
