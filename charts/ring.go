package charts

import (
	"fmt"
	"math"
	"strings"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Ring renders a circular progress ring using unicode quarter-block characters.
// The ring fills clockwise from the top as Value approaches Max.
// A label is centered inside the ring.
type Ring struct {
	// Value is the current progress value.
	Value float64
	// Max is the maximum value (100% fill).
	Max float64
	// Label is displayed in the center of the ring.
	Label string

	// FillColor overrides the theme accent color for the filled arc.
	FillColor lipgloss.Color
	// TrackColor overrides the theme muted color for the unfilled arc.
	TrackColor lipgloss.Color

	theme   blit.Theme
	width   int
	height  int
	focused bool
}

// NewRing creates a Ring chart.
func NewRing(value, max float64, label string) *Ring {
	return &Ring{
		Value: value,
		Max:   max,
		Label: label,
		theme: blit.DefaultTheme(),
	}
}

// Init initializes the Ring component.
func (r *Ring) Init() tea.Cmd { return nil }

// Update handles incoming messages and updates Ring state.
func (r *Ring) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	return r, nil
}

// KeyBindings returns the key bindings for the Ring.
func (r *Ring) KeyBindings() []blit.KeyBind { return nil }

// SetSize sets the width and height of the Ring.
func (r *Ring) SetSize(w, h int) { r.width = w; r.height = h }

// Focused reports whether the Ring is focused.
func (r *Ring) Focused() bool { return r.focused }

// SetFocused sets the focus state of the Ring.
func (r *Ring) SetFocused(f bool) { r.focused = f }

// SetTheme updates the theme used by the Ring.
func (r *Ring) SetTheme(t blit.Theme) { r.theme = t }

// quarterBlocks maps (topFill, bottomFill) → unicode block character.
// Each cell covers 2 vertical sub-pixels (top and bottom half).
var quarterBlocks = map[[2]bool]rune{
	{false, false}: ' ',
	{true, false}:  '▀',
	{false, true}:  '▄',
	{true, true}:   '█',
}

// View renders the Ring as a string.
func (r *Ring) View() string {
	if r.width < 3 || r.height < 3 {
		return ""
	}

	// Each terminal cell is approximately 2× taller than wide (aspect ~0.5).
	// We use 2 sub-rows per terminal row via half-block characters.
	cellAspect := 0.5
	size := r.width
	if r.height*2 < size {
		size = r.height * 2
	}
	// size is in sub-pixels; actual terminal rows = size/2
	subH := size
	subW := int(float64(size) * cellAspect * 2)
	if subW < 1 {
		subW = 1
	}

	termW := subW
	termH := subH / 2
	if termH < 1 {
		termH = 1
	}

	cx := float64(subW-1) / 2
	cy := float64(subH-1) / 2
	outerR := cy * 0.92
	innerR := outerR * 0.65

	// Progress fraction 0..1
	frac := 0.0
	if r.Max > 0 {
		frac = r.Value / r.Max
		if frac > 1 {
			frac = 1
		}
		if frac < 0 {
			frac = 0
		}
	}
	// Angle in radians: start at top (−π/2), go clockwise
	endAngle := -math.Pi/2 + frac*2*math.Pi

	fillColor := r.theme.Accent
	if r.FillColor != "" {
		fillColor = r.FillColor
	}
	trackColor := r.theme.Muted
	if r.TrackColor != "" {
		trackColor = r.TrackColor
	}
	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	trackStyle := lipgloss.NewStyle().Foreground(trackColor)

	// sub-pixel grid: true = filled arc, false = track arc, nil = empty
	type subPix struct {
		inRing bool
		filled bool
	}
	sub := make([][]subPix, subH)
	for i := range sub {
		sub[i] = make([]subPix, subW)
	}

	for sy := 0; sy < subH; sy++ {
		for sx := 0; sx < subW; sx++ {
			dx := float64(sx) - cx
			dy := float64(sy) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < innerR || dist > outerR {
				continue
			}
			// Angle from top, clockwise
			angle := math.Atan2(dy, dx) + math.Pi/2
			if angle < 0 {
				angle += 2 * math.Pi
			}
			sub[sy][sx].inRing = true
			// Determine if this angular position is within the filled arc
			startAngle := -math.Pi/2 + math.Pi/2 // = 0 after normalization
			_ = startAngle
			// Re-derive: start = top = angle 0 after atan2 adjustment above
			var filledArc float64
			if endAngle >= -math.Pi/2 {
				filledArc = endAngle + math.Pi/2
			} else {
				filledArc = endAngle + math.Pi/2 + 2*math.Pi
			}
			sub[sy][sx].filled = angle <= filledArc
		}
	}

	// Compose terminal cells
	var rows []string
	for tr := 0; tr < termH; tr++ {
		var line strings.Builder
		for tc := 0; tc < termW; tc++ {
			topSY := tr * 2
			botSY := tr*2 + 1
			topPix := subPix{}
			botPix := subPix{}
			if topSY < subH {
				topPix = sub[topSY][tc]
			}
			if botSY < subH {
				botPix = sub[botSY][tc]
			}

			topIn := topPix.inRing
			botIn := botPix.inRing
			topFill := topPix.inRing && topPix.filled
			botFill := botPix.inRing && botPix.filled

			if !topIn && !botIn {
				line.WriteByte(' ')
				continue
			}

			ch := quarterBlocks[[2]bool{topFill || (!topIn && botIn), botFill || (!botIn && topIn)}]
			if ch == ' ' {
				ch = quarterBlocks[[2]bool{topIn, botIn}]
			}

			var style lipgloss.Style
			if topFill || botFill {
				style = fillStyle
			} else {
				style = trackStyle
			}
			line.WriteString(style.Render(string(ch)))
		}
		rows = append(rows, line.String())
	}

	// Overlay label in center
	if r.Label != "" {
		labelLine := termH / 2
		label := r.Label
		pct := 0.0
		if r.Max > 0 {
			pct = (r.Value / r.Max) * 100
		}
		centreText := fmt.Sprintf("%s %.0f%%", label, pct)
		if len([]rune(centreText)) > termW {
			centreText = fmt.Sprintf("%.0f%%", pct)
		}
		pad := (termW - len([]rune(centreText))) / 2
		if pad < 0 {
			pad = 0
		}
		labelStr := strings.Repeat(" ", pad) + lipgloss.NewStyle().
			Foreground(r.theme.Text).Bold(true).
			Render(centreText)
		if labelLine < len(rows) {
			rows[labelLine] = labelStr
		}
	}

	return strings.Join(rows, "\n")
}
