package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailOverlayOpts configures a DetailOverlay.
type DetailOverlayOpts[T any] struct {
	// Render draws the item detail content. Required.
	Render func(item T, width, height int, theme Theme) string

	// Title is shown at the top of the overlay. Optional.
	Title string

	// KeyBindings are extra bindings shown in help. Optional.
	KeyBindings []KeyBind

	// OnKey handles custom key events. Optional.
	// Return Consumed() if handled, nil to let the App handle it.
	OnKey func(item T, key tea.KeyMsg) tea.Cmd
}

// DetailOverlay is a generic full-screen overlay for displaying item details.
type DetailOverlay[T any] struct {
	opts    DetailOverlayOpts[T]
	item    T
	theme   Theme
	active  bool
	focused bool
	width   int
	height  int
}

// NewDetailOverlay creates a detail overlay with the given options.
func NewDetailOverlay[T any](opts DetailOverlayOpts[T]) *DetailOverlay[T] {
	return &DetailOverlay[T]{opts: opts}
}

// Show opens the overlay with the given item.
func (d *DetailOverlay[T]) Show(item T) {
	d.item = item
	d.active = true
}

// Item returns the currently displayed item.
func (d *DetailOverlay[T]) Item() T {
	return d.item
}

func (d *DetailOverlay[T]) Init() tea.Cmd { return nil }

func (d *DetailOverlay[T]) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if d.opts.OnKey != nil {
			cmd := d.opts.OnKey(d.item, msg)
			if cmd != nil {
				return d, cmd
			}
		}
	}
	return d, nil
}

func (d *DetailOverlay[T]) View() string {
	if !d.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Accent)).
		Bold(true)
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Muted))

	innerW := d.width - 8  // border + padding
	innerH := d.height - 6 // border + padding + title + hint

	var sections []string

	if d.opts.Title != "" {
		sections = append(sections, titleStyle.Render(d.opts.Title))
		sections = append(sections, "")
		innerH -= 2
	}

	if innerH < 1 {
		innerH = 1
	}

	if d.opts.Render != nil {
		content := d.opts.Render(d.item, innerW, innerH, d.theme)
		sections = append(sections, content)
	}

	// Fill to push hint to bottom
	contentLines := strings.Count(strings.Join(sections, "\n"), "\n") + 1
	if contentLines < innerH+2 {
		sections = append(sections, strings.Repeat("\n", innerH+2-contentLines))
	}

	hints := "Esc close"
	for _, kb := range d.opts.KeyBindings {
		hints += "  " + kb.Key + " " + kb.Label
	}
	sections = append(sections, hintStyle.Render(hints))

	body := strings.Join(sections, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(d.theme.Border)).
		Padding(1, 2).
		Width(d.width - 4).
		Height(d.height - 2)

	return lipgloss.Place(d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(body))
}

func (d *DetailOverlay[T]) KeyBindings() []KeyBind {
	bindings := []KeyBind{
		{Key: "esc", Label: "Close", Group: "DETAIL"},
	}
	bindings = append(bindings, d.opts.KeyBindings...)
	return bindings
}

func (d *DetailOverlay[T]) SetSize(w, h int)  { d.width = w; d.height = h }
func (d *DetailOverlay[T]) Focused() bool      { return d.focused }
func (d *DetailOverlay[T]) SetFocused(f bool)  { d.focused = f }
func (d *DetailOverlay[T]) IsActive() bool     { return d.active }
func (d *DetailOverlay[T]) SetActive(v bool)   { d.active = v }
func (d *DetailOverlay[T]) Close()             { d.active = false }

// SetTheme implements the Themed interface.
func (d *DetailOverlay[T]) SetTheme(t Theme) { d.theme = t }
