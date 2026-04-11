// Package main demonstrates all blit chart types in a single dashboard.
//
// Layout: a 2×3 tile grid showing Bar (vertical), Bar (horizontal),
// Line, Ring, Gauge, Heatmap, and the built-in Sparkline.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/charts"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// dashboard is the root component that tiles all charts.
type dashboard struct {
	theme   blit.Theme
	width   int
	height  int
	focused bool

	barV    *charts.Bar
	barH    *charts.Bar
	line    *charts.Line
	ring    *charts.Ring
	gauge   *charts.Gauge
	heatmap *charts.Heatmap

	sparkData []float64
}

func newDashboard() *dashboard {
	// Bar vertical
	barV := charts.NewBar(
		[]float64{4, 7, 13, 8, 5, 11, 3, 9, 6, 12},
		[]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct"},
		false,
	)

	// Bar horizontal
	barH := charts.NewBar(
		[]float64{82, 64, 91, 47, 73},
		[]string{"Zorp", "Turbo", "Sally", "Buzz", "Knight"},
		true,
	)

	// Line — two series
	s1 := make([]float64, 20)
	s2 := make([]float64, 20)
	for i := range s1 {
		s1[i] = math.Sin(float64(i)*0.4) * 5
		s2[i] = math.Cos(float64(i)*0.4) * 4
	}
	line := charts.NewLine([][]float64{s1, s2}, nil, true)

	// Ring
	ring := charts.NewRing(73, 100, "CPU")

	// Gauge
	gauge := charts.NewGauge(68, 100, []float64{60, 80}, "Load")

	// Heatmap
	grid := make([][]float64, 5)
	for r := range grid {
		grid[r] = make([]float64, 7)
		for c := range grid[r] {
			grid[r][c] = math.Sin(float64(r+1)*float64(c+1)*0.3) * 50
		}
	}
	hm := charts.NewHeatmap(grid, charts.PaletteDivergent)
	hm.RowLabels = []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	hm.Labels = []string{"S", "M", "T", "W", "T", "F", "S"}

	// Sparkline data
	spark := make([]float64, 40)
	for i := range spark {
		spark[i] = math.Sin(float64(i)*0.3)*10 + 10
	}

	return &dashboard{
		theme:     blit.DefaultTheme(),
		barV:      barV,
		barH:      barH,
		line:      line,
		ring:      ring,
		gauge:     gauge,
		heatmap:   hm,
		sparkData: spark,
	}
}

func (d *dashboard) Init() tea.Cmd { return nil }
func (d *dashboard) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	return d, nil
}
func (d *dashboard) KeyBindings() []blit.KeyBind { return nil }
func (d *dashboard) SetSize(w, h int) {
	d.width = w
	d.height = h
	d.layout()
}
func (d *dashboard) Focused() bool     { return d.focused }
func (d *dashboard) SetFocused(f bool) { d.focused = f }
func (d *dashboard) SetTheme(t blit.Theme) {
	d.theme = t
	d.barV.SetTheme(t)
	d.barH.SetTheme(t)
	d.line.SetTheme(t)
	d.ring.SetTheme(t)
	d.gauge.SetTheme(t)
	d.heatmap.SetTheme(t)
}

func (d *dashboard) layout() {
	// 2-column layout: left and right halves
	// 3 rows per column
	colW := d.width / 2
	rowH := (d.height - 2) / 3 // -2 for title + sparkline row

	if colW < 4 {
		colW = 4
	}
	if rowH < 3 {
		rowH = 3
	}

	d.barV.SetSize(colW-2, rowH)
	d.barH.SetSize(colW-2, rowH)
	d.line.SetSize(colW-2, rowH)
	d.ring.SetSize(colW-2, rowH)
	d.gauge.SetSize(colW-2, rowH)
	d.heatmap.SetSize(colW-2, rowH)
}

func (d *dashboard) View() string {
	if d.width < 8 || d.height < 6 {
		return "terminal too small"
	}

	colW := d.width / 2
	rowH := (d.height - 2) / 3

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Accent)).
		Bold(true)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(d.theme.Border)).
		Width(colW - 2).
		Height(rowH)
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Muted)).
		Bold(true)

	tile := func(label, content string) string {
		inner := labelStyle.Render(label) + "\n" + content
		return box.Render(inner)
	}

	// Row 1
	r1left := tile("Bar Chart (vertical)", d.barV.View())
	r1right := tile("Bar Chart (horizontal)", d.barH.View())
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, r1left, r1right)

	// Row 2
	r2left := tile("Line Chart (smooth)", d.line.View())
	r2right := tile("Ring Chart — CPU 73%", d.ring.View())
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, r2left, r2right)

	// Row 3
	r3left := tile("Gauge — Load 68%", d.gauge.View())
	r3right := tile("Heatmap (divergent)", d.heatmap.View())
	row3 := lipgloss.JoinHorizontal(lipgloss.Top, r3left, r3right)

	// Sparkline footer
	spark, _ := blit.Sparkline(d.sparkData, d.width-4, nil)
	sparkLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Accent)).
		Render("  Sparkline: " + spark)

	header := title.Render("  blit charts dashboard — q to quit")

	return strings.Join([]string{header, row1, row2, row3, sparkLine}, "\n")
}

func main() {
	d := newDashboard()

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("charts", d),
		blit.WithStatusBar(
			func() string { return " q quit  ctrl+t cycle theme" },
			func() string { return fmt.Sprintf("  %s ", time.Now().Format("15:04:05")) },
		),
		blit.WithTickInterval(time.Second),
	)

	if _, err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
