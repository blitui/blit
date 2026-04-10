package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChartPanelOpts configures a ChartPanel.
type ChartPanelOpts struct {
	// Title is optional text shown above the chart.
	Title string

	// Charts are the available chart components. The panel shows one at a time.
	Charts []Component

	// Labels are optional display names for each chart (e.g., "Line", "Bar").
	// If shorter than Charts, indices are used for unlabelled charts.
	Labels []string
}

// ChartPanel is a container that displays one of several chart Components
// with a title bar and tab-style switcher. Implements Component and Themed.
type ChartPanel struct {
	opts    ChartPanelOpts
	active  int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewChartPanel creates a ChartPanel with the given options.
func NewChartPanel(opts ChartPanelOpts) *ChartPanel {
	return &ChartPanel{
		opts: opts,
	}
}

// ActiveIndex returns the index of the currently displayed chart.
func (c *ChartPanel) ActiveIndex() int { return c.active }

// SetActive switches to the chart at the given index (clamped).
func (c *ChartPanel) SetActive(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(c.opts.Charts) {
		idx = len(c.opts.Charts) - 1
	}
	if idx < 0 {
		idx = 0
	}
	c.active = idx
	c.sizeActiveChart()
}

// Init implements Component.
func (c *ChartPanel) Init() tea.Cmd {
	if len(c.opts.Charts) > 0 {
		return c.opts.Charts[c.active].Init()
	}
	return nil
}

// Update implements Component.
func (c *ChartPanel) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !c.focused {
		return c, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "tab":
			if len(c.opts.Charts) > 1 {
				c.active = (c.active + 1) % len(c.opts.Charts)
				c.sizeActiveChart()
			}
			return c, Consumed()
		case "shift+tab":
			if len(c.opts.Charts) > 1 {
				c.active = (c.active - 1 + len(c.opts.Charts)) % len(c.opts.Charts)
				c.sizeActiveChart()
			}
			return c, Consumed()
		}
	}
	// Forward to active chart.
	if len(c.opts.Charts) > 0 {
		updated, cmd := c.opts.Charts[c.active].Update(msg, ctx)
		c.opts.Charts[c.active] = updated
		return c, cmd
	}
	return c, nil
}

// View implements Component.
func (c *ChartPanel) View() string {
	if c.width == 0 || c.height == 0 {
		return ""
	}

	textColor := lipgloss.Color(c.theme.Text)
	accentColor := lipgloss.Color(c.theme.Accent)
	mutedColor := lipgloss.Color(c.theme.Muted)

	var lines []string

	// Title + tab switcher header.
	headerParts := make([]string, 0, 2)
	if c.opts.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
		headerParts = append(headerParts, titleStyle.Render(c.opts.Title))
	}

	if len(c.opts.Charts) > 1 {
		var tabs []string
		for i := range c.opts.Charts {
			label := c.labelFor(i)
			if i == c.active {
				style := lipgloss.NewStyle().
					Foreground(accentColor).
					Bold(true).
					Underline(true)
				tabs = append(tabs, style.Render(label))
			} else {
				style := lipgloss.NewStyle().
					Foreground(mutedColor)
				tabs = append(tabs, style.Render(label))
			}
		}
		headerParts = append(headerParts, strings.Join(tabs, " │ "))
	}

	if len(headerParts) > 0 {
		header := strings.Join(headerParts, "  ")
		lines = append(lines, header)
		// Separator.
		sep := lipgloss.NewStyle().
			Foreground(mutedColor).
			Render(strings.Repeat("─", c.width))
		lines = append(lines, sep)
	}

	// Chart content.
	if len(c.opts.Charts) > 0 {
		chart := c.opts.Charts[c.active]
		lines = append(lines, chart.View())
	} else {
		emptyStyle := lipgloss.NewStyle().Foreground(textColor)
		lines = append(lines, emptyStyle.Render("(no charts)"))
	}

	return strings.Join(lines, "\n")
}

// KeyBindings implements Component.
func (c *ChartPanel) KeyBindings() []KeyBind {
	binds := []KeyBind{}
	if len(c.opts.Charts) > 1 {
		binds = append(binds, KeyBind{Key: "tab/shift+tab", Label: "Switch chart", Group: "CHART"})
	}
	// Include active chart's bindings.
	if len(c.opts.Charts) > 0 {
		binds = append(binds, c.opts.Charts[c.active].KeyBindings()...)
	}
	return binds
}

// SetSize implements Component.
func (c *ChartPanel) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.sizeActiveChart()
}

// Focused implements Component.
func (c *ChartPanel) Focused() bool { return c.focused }

// SetFocused implements Component.
func (c *ChartPanel) SetFocused(f bool) { c.focused = f }

// SetTheme implements Themed.
func (c *ChartPanel) SetTheme(theme Theme) {
	c.theme = theme
	// Propagate to all charts that implement Themed.
	for _, chart := range c.opts.Charts {
		if themed, ok := chart.(Themed); ok {
			themed.SetTheme(theme)
		}
	}
}

func (c *ChartPanel) labelFor(idx int) string {
	if idx < len(c.opts.Labels) {
		return c.opts.Labels[idx]
	}
	return string(rune('1' + idx))
}

func (c *ChartPanel) sizeActiveChart() {
	if len(c.opts.Charts) == 0 || c.width == 0 {
		return
	}
	// Reserve lines for header (title + separator).
	chartHeight := c.height
	hasHeader := c.opts.Title != "" || len(c.opts.Charts) > 1
	if hasHeader {
		chartHeight -= 2 // title line + separator
	}
	if chartHeight < 1 {
		chartHeight = 1
	}
	c.opts.Charts[c.active].SetSize(c.width, chartHeight)
}
