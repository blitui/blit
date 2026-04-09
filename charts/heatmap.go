package charts

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// Palette names for Heatmap.
const (
	// PaletteSequential ranges from a near-black to the theme Accent color.
	PaletteSequential = "sequential"
	// PaletteDivergent ranges from Negative (low) through Muted (mid) to Positive (high).
	PaletteDivergent = "divergent"
)

// Heatmap renders a 2-D grid where each cell's background intensity represents
// its float64 value. The color palette is either Sequential or Divergent.
type Heatmap struct {
	// Grid is a row-major 2-D slice of values.
	Grid [][]float64
	// Palette selects the color scheme: PaletteSequential or PaletteDivergent.
	Palette string
	// Labels provides optional column header labels.
	Labels []string
	// RowLabels provides optional row labels rendered on the left.
	RowLabels []string

	theme   tuikit.Theme
	width   int
	height  int
	focused bool
}

// NewHeatmap creates a Heatmap chart.
func NewHeatmap(grid [][]float64, palette string) *Heatmap {
	return &Heatmap{
		Grid:    grid,
		Palette: palette,
		theme:   tuikit.DefaultTheme(),
	}
}

func (h *Heatmap) Init() tea.Cmd                             { return nil }
func (h *Heatmap) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return h, nil }
func (h *Heatmap) KeyBindings() []tuikit.KeyBind             { return nil }
func (h *Heatmap) SetSize(w, hh int)                         { h.width = w; h.height = hh }
func (h *Heatmap) Focused() bool                             { return h.focused }
func (h *Heatmap) SetFocused(f bool)                         { h.focused = f }
func (h *Heatmap) SetTheme(t tuikit.Theme)                   { h.theme = t }

// cellColor returns the background color for a normalized value t ∈ [0, 1].
func (h *Heatmap) cellColor(t float64) lipgloss.Color {
	switch h.Palette {
	case PaletteDivergent:
		if t < 0.5 {
			g := tuikit.Gradient{Start: h.theme.Negative, End: h.theme.Muted}
			return g.RenderAt(t * 2)
		}
		g := tuikit.Gradient{Start: h.theme.Muted, End: h.theme.Positive}
		return g.RenderAt((t - 0.5) * 2)
	default: // sequential
		g := tuikit.Gradient{Start: lipgloss.Color("#111111"), End: h.theme.Accent}
		return g.RenderAt(t)
	}
}

func (h *Heatmap) View() string {
	if h.width < 2 || h.height < 1 || len(h.Grid) == 0 {
		return ""
	}

	rows := len(h.Grid)
	cols := 0
	for _, row := range h.Grid {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return ""
	}

	// Compute global min/max for normalization
	gMin, gMax := math.MaxFloat64, -math.MaxFloat64
	for _, row := range h.Grid {
		for _, v := range row {
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

	// Row label width
	rowLabelW := 0
	for _, rl := range h.RowLabels {
		if len(rl) > rowLabelW {
			rowLabelW = len(rl)
		}
	}
	if rowLabelW > 0 {
		rowLabelW++ // space separator
	}

	// Column label row height
	hasColLabels := len(h.Labels) > 0

	drawW := h.width - rowLabelW
	drawH := h.height
	if hasColLabels {
		drawH--
	}
	if drawH < 1 {
		drawH = 1
	}

	// Cell width: distribute drawW across cols
	cellW := drawW / cols
	if cellW < 1 {
		cellW = 1
	}
	// How many cols actually fit
	fitCols := drawW / cellW
	if fitCols > cols {
		fitCols = cols
	}

	// Cell height: distribute drawH across rows
	cellH := drawH / rows
	if cellH < 1 {
		cellH = 1
	}
	fitRows := drawH / cellH
	if fitRows > rows {
		fitRows = rows
	}

	// Build output lines
	var lines []string

	// Column labels
	if hasColLabels {
		var lblLine strings.Builder
		if rowLabelW > 0 {
			lblLine.WriteString(strings.Repeat(" ", rowLabelW))
		}
		for c := 0; c < fitCols; c++ {
			lbl := ""
			if c < len(h.Labels) {
				lbl = h.Labels[c]
			}
			if len([]rune(lbl)) > cellW {
				lbl = string([]rune(lbl)[:cellW])
			}
			lblLine.WriteString(
				lipgloss.NewStyle().
					Foreground(h.theme.Muted).
					Width(cellW).
					Render(lbl),
			)
		}
		lines = append(lines, lblLine.String())
	}

	// Data rows
	for r := 0; r < fitRows; r++ {
		for line := 0; line < cellH; line++ {
			var rowStr strings.Builder
			// Row label (only on first cell line)
			if rowLabelW > 0 {
				rl := ""
				if line == 0 && r < len(h.RowLabels) {
					rl = h.RowLabels[r]
				}
				rowStr.WriteString(
					lipgloss.NewStyle().
						Foreground(h.theme.Muted).
						Width(rowLabelW).
						Render(rl),
				)
			}
			for c := 0; c < fitCols; c++ {
				v := 0.0
				if r < len(h.Grid) && c < len(h.Grid[r]) {
					v = h.Grid[r][c]
				}
				t := (v - gMin) / (gMax - gMin)
				bg := h.cellColor(t)

				// Cell content: value on first line of cell, spaces otherwise
				content := strings.Repeat(" ", cellW)
				if line == 0 && cellH >= 1 {
					valStr := fmt.Sprintf("%-*s", cellW, fmt.Sprintf("%.0f", v))
					if len([]rune(valStr)) > cellW {
						valStr = strings.Repeat(" ", cellW)
					}
					content = valStr
				}

				rowStr.WriteString(
					lipgloss.NewStyle().
						Background(bg).
						Foreground(h.theme.TextInverse).
						Render(content),
				)
			}
			lines = append(lines, rowStr.String())
		}
	}

	return strings.Join(lines, "\n")
}
