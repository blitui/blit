package blit

import (
	"github.com/charmbracelet/lipgloss"
)

// CollapsibleSection is a render helper for collapsible panel sections.
// It can be used in two ways:
//
//  1. As a render helper: call Render() directly from a parent component.
//  2. As a Component: embed it in a layout and let the App manage it.
//
// When used as a Component, it handles its own key events (enter/space to
// toggle), focus, size, and theming.
type CollapsibleSection struct {
	Title     string
	Collapsed bool

	// Content is the text rendered below the header when expanded.
	// Set this directly or use Render() with a contentFunc for lazy evaluation.
	Content string

	// Component state (only used when mounted as a Component)
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewCollapsibleSection creates an expanded section with the given title.
func NewCollapsibleSection(title string) *CollapsibleSection {
	return &CollapsibleSection{Title: title}
}

// Toggle flips the collapsed state.
func (s *CollapsibleSection) Toggle() {
	s.Collapsed = !s.Collapsed
}

// Render returns the section header and, if expanded, the content.
// contentFunc is only called when expanded to avoid wasted work.
func (s *CollapsibleSection) Render(theme Theme, contentFunc func() string) string {
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))

	glyphs := theme.glyphsOrDefault()
	if s.Collapsed {
		return arrowStyle.Render(glyphs.CollapsedArrow) + " " + titleStyle.Render(s.Title)
	}

	header := arrowStyle.Render(glyphs.ExpandedArrow) + " " + titleStyle.Render(s.Title)
	content := contentFunc()
	if content == "" {
		return header
	}
	return header + "\n" + content
}

// --- Component implementation --------------------------------------------------

// Init implements Component.
func (s *CollapsibleSection) Init() Cmd { return nil }

// Update implements Component.
func (s *CollapsibleSection) Update(msg Msg, ctx Context) (Component, Cmd) {
	if IsEnter(msg) || IsKey(msg, " ") {
		s.Toggle()
		return s, Consumed()
	}
	return s, nil
}

// View implements Component.
func (s *CollapsibleSection) View() string {
	theme := s.theme
	if theme.Text == "" {
		theme = DefaultTheme()
	}

	arrowStyle := NewStyle().Foreground(theme.Muted)
	titleStyle := NewStyle().Foreground(theme.Text)

	glyphs := theme.glyphsOrDefault()
	if s.Collapsed {
		return arrowStyle.Render(glyphs.CollapsedArrow) + " " + titleStyle.Render(s.Title)
	}

	header := arrowStyle.Render(glyphs.ExpandedArrow) + " " + titleStyle.Render(s.Title)
	if s.Content == "" {
		return header
	}
	return header + "\n" + s.Content
}

// KeyBindings implements Component.
func (s *CollapsibleSection) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "enter/space", Label: "Toggle section", Group: "NAVIGATION"},
	}
}

// SetSize implements Component.
func (s *CollapsibleSection) SetSize(w, h int) {
	s.width = w
	s.height = h
}

// Focused implements Component.
func (s *CollapsibleSection) Focused() bool { return s.focused }

// SetFocused implements Component.
func (s *CollapsibleSection) SetFocused(f bool) { s.focused = f }

// SetTheme implements Themed.
func (s *CollapsibleSection) SetTheme(t Theme) { s.theme = t }
