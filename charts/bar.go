// Package charts provides terminal chart components for tuikit.
// Each chart implements tuikit.Component and renders to its assigned width/height.
package charts

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// Bar renders a bar chart, either vertical or horizontal.
// Data values are normalized to the available draw area.
// Colors cycle through the theme palette; a Gradient may override per-bar fill.
type Bar struct {
	// Data is the list of values to plot.
	Data []float64
	// Labels are optional per-bar labels (may be shorter than Data).
	Labels []string
	// Horizontal renders bars left-to-right instead of bottom-to-top.
	Horizontal bool
	// Gradient, when non-nil, fills each bar using a start→end color gradient
	// (reuses the v0.8 Gradient type from the parent package).
	Gradient *tuikit.Gradient
	// Colors overrides per-bar colors. Cycles if shorter than Data.
	Colors []lipgloss.Color

	theme   tuikit.Theme
	width   int
	height  int
	focused bool
}

// NewBar creates a Bar chart with the given data and labels.
func NewBar(data []float64, labels []string, horizontal bool) *Bar {
	return &Bar{
		Data:       data,
		Labels:     labels,
		Horizontal: horizontal,
		theme:      tuikit.DefaultTheme(),
	}
}

func (b *Bar) Init() tea.Cmd                             { return nil }
func (b *Bar) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return b, nil }
func (b *Bar) KeyBindings() []tuikit.KeyBind             { return nil }
func (b *Bar) SetSize(w, h int)                          { b.width = w; b.height = h }
func (b *Bar) Focused() bool                             { return b.focused }
func (b *Bar) SetFocused(f bool)                         { b.focused = f }
func (b *Bar) SetTheme(t tuikit.Theme)                   { b.theme = t }

// barColor returns the color for bar i.
func (b *Bar) barColor(i int) lipgloss.Color {
	if len(b.Colors) > 0 {
		return b.Colors[i%len(b.Colors)]
	}
	defaults := []lipgloss.Color{
		b.theme.Accent,
		b.theme.Positive,
		b.theme.Cursor,
		b.theme.Flash,
		b.theme.Negative,
	}
	return defaults[i%len(defaults)]
}

func (b *Bar) View() string {
	if b.width < 2 || b.height < 2 || len(b.Data) == 0 {
		return ""
	}
	if b.Horizontal {
		return b.viewHorizontal()
	}
	return b.viewVertical()
}

func (b *Bar) viewVertical() string {
	n := len(b.Data)
	// Reserve 1 line for labels if any are set
	hasLabels := len(b.Labels) > 0
	drawHeight := b.height
	if hasLabels {
		drawHeight--
	}
	if drawHeight < 1 {
		drawHeight = 1
	}

	// Compute column width: distribute evenly, min 1
	colW := b.width / n
	if colW < 1 {
		colW = 1
	}
	// Clamp n to what fits
	if n*colW > b.width {
		n = b.width / colW
	}

	maxVal := maxFloat(b.Data[:n])
	if maxVal == 0 {
		maxVal = 1
	}

	// Build each column as a slice of lines (top → bottom)
	blocks := " ▁▂▃▄▅▆▇█"
	blockRunes := []rune(blocks)

	cols := make([]string, n)
	for i := 0; i < n; i++ {
		v := b.Data[i]
		fillF := (v / maxVal) * float64(drawHeight)
		fullRows := int(fillF)
		frac := fillF - float64(fullRows)

		color := b.barColor(i)
		if b.Gradient != nil {
			t := 0.0
			if n > 1 {
				t = float64(i) / float64(n-1)
			}
			color = b.Gradient.RenderAt(t)
		}
		style := lipgloss.NewStyle().Foreground(color)

		var colLines []string
		// Empty rows above the bar
		emptyRows := drawHeight - fullRows
		if frac > 0 {
			emptyRows--
		}
		for j := 0; j < emptyRows; j++ {
			colLines = append(colLines, strings.Repeat(" ", colW))
		}
		// Partial row
		if frac > 0 {
			fracIdx := int(math.Round(frac * 8))
			if fracIdx < 1 {
				fracIdx = 1
			}
			if fracIdx > 8 {
				fracIdx = 8
			}
			fracChar := string(blockRunes[fracIdx])
			colLines = append(colLines, style.Render(strings.Repeat(fracChar, colW)))
		}
		// Full rows
		for j := 0; j < fullRows; j++ {
			colLines = append(colLines, style.Render(strings.Repeat("█", colW)))
		}

		// Pad height to drawHeight
		for len(colLines) < drawHeight {
			colLines = append([]string{strings.Repeat(" ", colW)}, colLines...)
		}

		cols[i] = strings.Join(colLines, "\n")
	}

	// Join columns side by side using lipgloss
	rendered := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	if hasLabels {
		var labelParts []string
		for i := 0; i < n; i++ {
			lbl := ""
			if i < len(b.Labels) {
				lbl = b.Labels[i]
			}
			if len([]rune(lbl)) > colW {
				lbl = string([]rune(lbl)[:colW])
			}
			lbl = fmt.Sprintf("%-*s", colW, lbl)
			labelParts = append(labelParts, lipgloss.NewStyle().
				Foreground(b.theme.Muted).
				Render(lbl))
		}
		labelRow := strings.Join(labelParts, "")
		rendered = lipgloss.JoinVertical(lipgloss.Left, rendered, labelRow)
	}

	return rendered
}

func (b *Bar) viewHorizontal() string {
	n := len(b.Data)
	// Each bar takes 1 row; clamp to height
	if n > b.height {
		n = b.height
	}

	// Label column width: max label length, capped at 1/4 of width
	labelW := 0
	for i := 0; i < n; i++ {
		if i < len(b.Labels) && len(b.Labels[i]) > labelW {
			labelW = len(b.Labels[i])
		}
	}
	maxLabelW := b.width / 4
	if labelW > maxLabelW {
		labelW = maxLabelW
	}
	barW := b.width - labelW - 1
	if barW < 1 {
		barW = 1
	}

	maxVal := maxFloat(b.Data[:n])
	if maxVal == 0 {
		maxVal = 1
	}

	var lines []string
	for i := 0; i < n; i++ {
		v := b.Data[i]
		fillF := (v / maxVal) * float64(barW)
		fullCols := int(fillF)
		frac := fillF - float64(fullCols)

		color := b.barColor(i)
		if b.Gradient != nil {
			t := 0.0
			if n > 1 {
				t = float64(i) / float64(n-1)
			}
			color = b.Gradient.RenderAt(t)
		}
		style := lipgloss.NewStyle().Foreground(color)

		var bar strings.Builder
		bar.WriteString(style.Render(strings.Repeat("█", fullCols)))
		if frac >= 0.5 && fullCols < barW {
			bar.WriteString(style.Render("▌"))
			fullCols++
		}
		// Pad remainder
		remainder := barW - fullCols
		if remainder > 0 {
			bar.WriteString(strings.Repeat(" ", remainder))
		}

		lbl := ""
		if i < len(b.Labels) {
			lbl = b.Labels[i]
		}
		if labelW > 0 {
			lbl = fmt.Sprintf("%-*s", labelW, lbl)
			lbl = lipgloss.NewStyle().Foreground(b.theme.Muted).Render(lbl) + " "
		}
		lines = append(lines, lbl+bar.String())
	}

	return strings.Join(lines, "\n")
}

func maxFloat(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
