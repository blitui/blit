package charts

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

// Gauge renders a half-circle (semicircle) gauge with colored threshold bands.
// The needle sweeps from left (0°) to right (180°) as Value increases toward Max.
// Thresholds partition the arc into colored bands (green → yellow → red by default).
type Gauge struct {
	// Value is the current reading.
	Value float64
	// Max is the full-scale value.
	Max float64
	// Thresholds are the values at which the band color changes.
	// E.g. []float64{0.6, 0.8} with Max=1.0 gives three bands:
	//   [0,0.6) Positive, [0.6,0.8) Flash, [0.8,1.0] Negative
	Thresholds []float64
	// Label is displayed below the gauge needle value.
	Label string

	theme   blit.Theme
	width   int
	height  int
	focused bool
}

// NewGauge creates a Gauge chart.
func NewGauge(value, max float64, thresholds []float64, label string) *Gauge {
	return &Gauge{
		Value:      value,
		Max:        max,
		Thresholds: thresholds,
		Label:      label,
		theme:      blit.DefaultTheme(),
	}
}

// Init initializes the Gauge component.
func (g *Gauge) Init() tea.Cmd { return nil }
// Update handles incoming messages and updates Gauge state.
func (g *Gauge) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	return g, nil
}
// KeyBindings returns the key bindings for the Gauge.
func (g *Gauge) KeyBindings() []blit.KeyBind { return nil }
// SetSize sets the width and height of the Gauge.
func (g *Gauge) SetSize(w, h int)              { g.width = w; g.height = h }
// Focused reports whether the Gauge is focused.
func (g *Gauge) Focused() bool                 { return g.focused }
// SetFocused sets the focus state of the Gauge.
func (g *Gauge) SetFocused(f bool)             { g.focused = f }
// SetTheme updates the theme used by the Gauge.
func (g *Gauge) SetTheme(t blit.Theme)       { g.theme = t }

// bandColor returns the theme color for the band index.
func (g *Gauge) bandColor(band int) lipgloss.Color {
	colors := []lipgloss.Color{
		g.theme.Positive,
		g.theme.Flash,
		g.theme.Negative,
	}
	if band < len(colors) {
		return colors[band]
	}
	return g.theme.Accent
}

// View renders the Gauge as a string.
func (g *Gauge) View() string {
	if g.width < 5 || g.height < 3 {
		return ""
	}

	// The gauge is a semicircle: angles from π (left) to 0 (right), sweeping via top.
	// Terminal aspect: cells are ~2× tall as wide; we use half-block for sub-rows.
	subH := g.height * 2
	subW := g.width

	cx := float64(subW-1) / 2
	cy := float64(subH - 1) // bottom of the sub-pixel grid

	// Radius: fit within width/height
	maxR := cx
	if float64(subH)*0.5 < maxR {
		maxR = float64(subH) * 0.5
	}
	outerR := maxR * 0.95
	innerR := outerR * 0.60

	// Needle angle: π at 0%, 0 at 100% (sweeping from left to right via top)
	frac := 0.0
	if g.Max > 0 {
		frac = g.Value / g.Max
		if frac > 1 {
			frac = 1
		}
		if frac < 0 {
			frac = 0
		}
	}
	needleAngle := math.Pi - frac*math.Pi // π → 0

	// Determine band for each angle
	threshFracs := make([]float64, len(g.Thresholds))
	for i, t := range g.Thresholds {
		if g.Max > 0 {
			threshFracs[i] = t / g.Max
		}
	}

	arcBandAt := func(angle float64) int {
		// angle goes from π (left=0%) to 0 (right=100%)
		f := 1 - angle/math.Pi
		band := 0
		for _, tf := range threshFracs {
			if f >= tf {
				band++
			}
		}
		return band
	}

	// Sub-pixel grid
	type subPix struct {
		inArc  bool
		filled bool // below needle angle (value side)
		band   int
	}
	sub := make([][]subPix, subH)
	for i := range sub {
		sub[i] = make([]subPix, subW)
	}

	for sy := 0; sy < subH; sy++ {
		for sx := 0; sx < subW; sx++ {
			dx := float64(sx) - cx
			dy := cy - float64(sy) // invert y so up is positive
			dist := math.Sqrt(dx*dx + dy*dy)
			if dy < 0 {
				continue // bottom half only for semicircle
			}
			if dist < innerR || dist > outerR {
				continue
			}
			angle := math.Atan2(dy, dx) // 0..π for upper semicircle
			if angle < 0 || angle > math.Pi {
				continue
			}
			sub[sy][sx].inArc = true
			sub[sy][sx].filled = angle >= needleAngle
			sub[sy][sx].band = arcBandAt(angle)
		}
	}

	// Compose terminal cells using half-blocks
	termH := subH / 2
	var rows []string
	for tr := 0; tr < termH; tr++ {
		var line strings.Builder
		for tc := 0; tc < subW; tc++ {
			topSY := tr * 2
			botSY := tr*2 + 1
			var top, bot subPix
			if topSY < subH {
				top = sub[topSY][tc]
			}
			if botSY < subH {
				bot = sub[botSY][tc]
			}

			if !top.inArc && !bot.inArc {
				line.WriteByte(' ')
				continue
			}

			topFill := top.inArc && top.filled
			botFill := bot.inArc && bot.filled
			topIn := top.inArc
			botIn := bot.inArc

			ch := quarterBlocks[[2]bool{topIn, botIn}]
			if topFill || botFill {
				// Use the band color of whichever sub-pixel is filled
				band := top.band
				if botFill {
					band = bot.band
				}
				line.WriteString(lipgloss.NewStyle().Foreground(g.bandColor(band)).Render(string(ch)))
			} else {
				line.WriteString(lipgloss.NewStyle().Foreground(g.theme.Muted).Render(string(ch)))
			}
		}
		rows = append(rows, line.String())
	}

	// Bottom label row: show value
	valStr := fmt.Sprintf("%.1f / %.1f", g.Value, g.Max)
	if g.Label != "" {
		valStr = g.Label + " " + valStr
	}
	pad := (g.width - len([]rune(valStr))) / 2
	if pad < 0 {
		pad = 0
	}
	labelRow := strings.Repeat(" ", pad) +
		lipgloss.NewStyle().Foreground(g.theme.Text).Render(valStr)

	rows = append(rows, labelRow)
	return strings.Join(rows, "\n")
}
