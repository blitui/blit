package charts

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// Line renders a multi-series line chart using unicode box-drawing characters.
// Each series is drawn as a connected path; when Smooth is true the renderer
// uses diagonal connectors (╱ ╲) between points.
type Line struct {
	// Series holds one data slice per line series.
	Series [][]float64
	// Colors maps series index → color hex string ("#rrggbb").
	// Cycles if shorter than Series; falls back to theme colors.
	Colors []string
	// Smooth enables diagonal connectors between adjacent points.
	Smooth bool

	theme   tuikit.Theme
	width   int
	height  int
	focused bool
}

// NewLine creates a Line chart.
func NewLine(series [][]float64, colors []string, smooth bool) *Line {
	return &Line{
		Series: series,
		Colors: colors,
		Smooth: smooth,
		theme:  tuikit.DefaultTheme(),
	}
}

func (l *Line) Init() tea.Cmd                             { return nil }
func (l *Line) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return l, nil }
func (l *Line) KeyBindings() []tuikit.KeyBind             { return nil }
func (l *Line) SetSize(w, h int)                          { l.width = w; l.height = h }
func (l *Line) Focused() bool                             { return l.focused }
func (l *Line) SetFocused(f bool)                         { l.focused = f }
func (l *Line) SetTheme(t tuikit.Theme)                   { l.theme = t }

// seriesColor returns the lipgloss color for series i.
func (l *Line) seriesColor(i int) lipgloss.Color {
	if i < len(l.Colors) && l.Colors[i] != "" {
		return lipgloss.Color(l.Colors[i])
	}
	defaults := []lipgloss.Color{
		l.theme.Accent,
		l.theme.Positive,
		l.theme.Cursor,
		l.theme.Flash,
		l.theme.Negative,
	}
	return defaults[i%len(defaults)]
}

func (l *Line) View() string {
	if l.width < 2 || l.height < 2 || len(l.Series) == 0 {
		return ""
	}

	// Global min/max across all series for consistent Y scale
	gMin, gMax := math.MaxFloat64, -math.MaxFloat64
	for _, s := range l.Series {
		for _, v := range s {
			if v < gMin {
				gMin = v
			}
			if v > gMax {
				gMax = v
			}
		}
	}
	if gMax == gMin {
		gMax = gMin + 1
	}

	// cell grid: row 0 = top, row height-1 = bottom
	// Each cell holds a rune and a color
	type cell struct {
		ch    rune
		color lipgloss.Color
	}
	grid := make([][]cell, l.height)
	for i := range grid {
		grid[i] = make([]cell, l.width)
		for j := range grid[i] {
			grid[i][j] = cell{' ', ""}
		}
	}

	yToRow := func(v float64) int {
		t := (v - gMin) / (gMax - gMin)
		row := l.height - 1 - int(math.Round(t*float64(l.height-1)))
		if row < 0 {
			row = 0
		}
		if row >= l.height {
			row = l.height - 1
		}
		return row
	}

	for si, series := range l.Series {
		if len(series) == 0 {
			continue
		}
		color := l.seriesColor(si)

		// Map data points to x columns
		n := len(series)
		pts := make([]struct{ col, row int }, n)
		for i, v := range series {
			col := 0
			if n > 1 {
				col = int(math.Round(float64(i) / float64(n-1) * float64(l.width-1)))
			}
			pts[i] = struct{ col, row int }{col, yToRow(v)}
		}

		for i, p := range pts {
			// Plot the point
			grid[p.row][p.col] = cell{'●', color}

			// Draw connector to next point
			if i+1 >= len(pts) {
				continue
			}
			np := pts[i+1]

			if p.row == np.row {
				// Horizontal segment
				for c := p.col + 1; c < np.col; c++ {
					if grid[p.row][c].ch == ' ' {
						grid[p.row][c] = cell{'─', color}
					}
				}
			} else if l.Smooth {
				// Diagonal segments
				dr := np.row - p.row
				dc := np.col - p.col
				steps := dc
				if steps < 1 {
					steps = 1
				}
				for s := 1; s < steps; s++ {
					t := float64(s) / float64(steps)
					row := p.row + int(math.Round(t*float64(dr)))
					col := p.col + s
					if row < 0 || row >= l.height || col < 0 || col >= l.width {
						continue
					}
					var ch rune
					if dr > 0 {
						ch = '╲'
					} else {
						ch = '╱'
					}
					if grid[row][col].ch == ' ' {
						grid[row][col] = cell{ch, color}
					}
				}
			} else {
				// Vertical step at midpoint then horizontal
				midCol := (p.col + np.col) / 2
				// Horizontal to midCol
				for c := p.col + 1; c < midCol && c < l.width; c++ {
					if grid[p.row][c].ch == ' ' {
						grid[p.row][c] = cell{'─', color}
					}
				}
				// Vertical segment
				rMin, rMax := p.row, np.row
				if rMin > rMax {
					rMin, rMax = rMax, rMin
				}
				for r := rMin; r <= rMax; r++ {
					if midCol < l.width && grid[r][midCol].ch == ' ' {
						grid[r][midCol] = cell{'│', color}
					}
				}
				// Horizontal from midCol to next point
				for c := midCol + 1; c < np.col && c < l.width; c++ {
					if grid[np.row][c].ch == ' ' {
						grid[np.row][c] = cell{'─', color}
					}
				}
			}
		}
	}

	// Render grid to string
	var sb strings.Builder
	for r, row := range grid {
		for _, c := range row {
			if c.ch == ' ' || c.color == "" {
				sb.WriteRune(' ')
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(c.color).Render(string(c.ch)))
			}
		}
		if r < len(grid)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
