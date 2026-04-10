package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TooltipOpts configures a Tooltip.
type TooltipOpts struct {
	// Text is the tooltip content.
	Text string

	// MaxWidth limits the tooltip width. Defaults to 40.
	MaxWidth int
}

// Tooltip is a floating text hint that composites over background content.
// Implements FloatingOverlay and Themed.
type Tooltip struct {
	opts    TooltipOpts
	theme   Theme
	focused bool
	active  bool
	width   int
	height  int
	anchorX int
	anchorY int
}

// NewTooltip creates a Tooltip with the given options.
func NewTooltip(opts TooltipOpts) *Tooltip {
	if opts.MaxWidth <= 0 {
		opts.MaxWidth = 40
	}
	return &Tooltip{
		opts: opts,
	}
}

// Text returns the current tooltip text.
func (t *Tooltip) Text() string { return t.opts.Text }

// SetText updates the tooltip content.
func (t *Tooltip) SetText(text string) { t.opts.Text = text }

// SetAnchor sets the position where the tooltip should appear.
func (t *Tooltip) SetAnchor(x, y int) {
	t.anchorX = x
	t.anchorY = y
}

// Show activates the tooltip.
func (t *Tooltip) Show() { t.active = true }

// Init implements Component.
func (t *Tooltip) Init() tea.Cmd { return nil }

// Update implements Component.
func (t *Tooltip) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !t.active {
		return t, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return t, nil
	}
	if km.String() == "esc" {
		t.active = false
		return t, Consumed()
	}
	return t, nil
}

// View implements Component. Returns the tooltip box for standalone use.
func (t *Tooltip) View() string {
	if !t.active || t.opts.Text == "" {
		return ""
	}
	return t.renderBox()
}

// FloatView implements FloatingOverlay. Composites the tooltip over background.
func (t *Tooltip) FloatView(background string) string {
	if !t.active || t.opts.Text == "" {
		return background
	}

	box := t.renderBox()
	return t.composite(background, box, t.anchorX, t.anchorY)
}

// IsActive implements Overlay.
func (t *Tooltip) IsActive() bool { return t.active }

// Close implements Overlay.
func (t *Tooltip) Close() { t.active = false }

// Inline returns false — tooltip is a floating overlay.
func (t *Tooltip) Inline() bool { return false }

// KeyBindings implements Component.
func (t *Tooltip) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "esc", Label: "Dismiss", Group: "TOOLTIP"},
	}
}

// SetSize implements Component.
func (t *Tooltip) SetSize(w, h int) {
	t.width = w
	t.height = h
}

// Focused implements Component.
func (t *Tooltip) Focused() bool { return t.focused }

// SetFocused implements Component.
func (t *Tooltip) SetFocused(f bool) { t.focused = f }

// SetTheme implements Themed.
func (t *Tooltip) SetTheme(theme Theme) { t.theme = theme }

func (t *Tooltip) renderBox() string {
	borderColor := lipgloss.Color(t.theme.Border)
	textColor := lipgloss.Color(t.theme.Text)
	bgColor := lipgloss.Color(t.theme.Cursor)

	borders := DefaultBorders()

	style := lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top:         borders.Rounded.Top,
			Bottom:      borders.Rounded.Bottom,
			Left:        borders.Rounded.Left,
			Right:       borders.Rounded.Right,
			TopLeft:     borders.Rounded.TopLeft,
			TopRight:    borders.Rounded.TopRight,
			BottomLeft:  borders.Rounded.BottomLeft,
			BottomRight: borders.Rounded.BottomRight,
		}).
		BorderForeground(borderColor).
		Foreground(textColor).
		Background(bgColor).
		MaxWidth(t.opts.MaxWidth).
		Padding(0, 1)

	return style.Render(t.opts.Text)
}

func (t *Tooltip) composite(bg, overlay string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	ovLines := strings.Split(overlay, "\n")

	// Clamp position.
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	// Ensure enough background lines.
	for len(bgLines) < y+len(ovLines) {
		bgLines = append(bgLines, "")
	}

	for i, ovLine := range ovLines {
		row := y + i
		if row >= len(bgLines) {
			break
		}

		bgLine := bgLines[row]
		bgRunes := []rune(bgLine)

		// Pad if needed.
		for len(bgRunes) < x {
			bgRunes = append(bgRunes, ' ')
		}

		ovRunes := []rune(ovLine)
		result := make([]rune, 0, len(bgRunes)+len(ovRunes))
		result = append(result, bgRunes[:x]...)
		result = append(result, ovRunes...)
		if x+len(ovRunes) < len(bgRunes) {
			result = append(result, bgRunes[x+len(ovRunes):]...)
		}

		bgLines[row] = string(result)
	}

	return strings.Join(bgLines, "\n")
}
