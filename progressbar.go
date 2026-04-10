package blit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBarOpts configures a ProgressBar.
type ProgressBarOpts struct {
	// Label is optional text shown to the left of the bar.
	Label string

	// ShowPercent displays the percentage to the right of the bar.
	ShowPercent bool

	// Width overrides the bar width. If 0, the component uses its SetSize width.
	Width int
}

// ProgressBar is a horizontal bar that fills proportionally to a value in [0,1].
// It implements Component and Themed.
type ProgressBar struct {
	opts    ProgressBarOpts
	value   float64 // 0.0–1.0
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewProgressBar creates a ProgressBar with the given options and initial value.
// The value is clamped to [0, 1].
func NewProgressBar(opts ProgressBarOpts, value float64) *ProgressBar {
	return &ProgressBar{
		opts:  opts,
		value: clampFloat(value, 0, 1),
	}
}

// Value returns the current progress value in [0, 1].
func (p *ProgressBar) Value() float64 { return p.value }

// SetValue updates the progress. The value is clamped to [0, 1].
func (p *ProgressBar) SetValue(v float64) {
	p.value = clampFloat(v, 0, 1)
}

// Increment adds delta to the current value (clamped to [0, 1]).
func (p *ProgressBar) Increment(delta float64) {
	p.SetValue(p.value + delta)
}

// Init implements Component.
func (p *ProgressBar) Init() tea.Cmd { return nil }

// Update implements Component. ProgressBar is display-only and does not
// consume any messages.
func (p *ProgressBar) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	return p, nil
}

// View implements Component.
func (p *ProgressBar) View() string {
	g := p.theme.glyphsOrDefault()

	barWidth := p.opts.Width
	if barWidth == 0 {
		barWidth = p.width
	}

	label := p.opts.Label
	var suffix string
	if p.opts.ShowPercent {
		suffix = fmt.Sprintf(" %3.0f%%", p.value*100)
	}

	// Subtract label and suffix from available width.
	if label != "" {
		label += " "
	}
	available := barWidth - lipgloss.Width(label) - lipgloss.Width(suffix)
	if available < 1 {
		available = 1
	}

	filled := int(p.value * float64(available))
	if filled > available {
		filled = available
	}
	empty := available - filled

	filledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Positive))
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Muted))
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Text))
	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Accent))

	bar := filledStyle.Render(strings.Repeat(g.BarFilled, filled)) +
		emptyStyle.Render(strings.Repeat(g.BarEmpty, empty))

	var sb strings.Builder
	if label != "" {
		sb.WriteString(labelStyle.Render(label))
	}
	sb.WriteString(bar)
	if suffix != "" {
		sb.WriteString(percentStyle.Render(suffix))
	}
	return sb.String()
}

// KeyBindings implements Component. ProgressBar has no key bindings.
func (p *ProgressBar) KeyBindings() []KeyBind { return nil }

// SetSize implements Component.
func (p *ProgressBar) SetSize(w, h int) {
	p.width = w
	p.height = h
}

// Focused implements Component.
func (p *ProgressBar) Focused() bool { return p.focused }

// SetFocused implements Component.
func (p *ProgressBar) SetFocused(f bool) { p.focused = f }

// SetTheme implements Themed.
func (p *ProgressBar) SetTheme(theme Theme) { p.theme = theme }

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
