package tuikit

import "github.com/charmbracelet/lipgloss"

// CollapsibleSection is a render helper for collapsible panel sections.
// It is not a Component — the parent component owns its state and calls Toggle().
type CollapsibleSection struct {
	Title     string
	Collapsed bool
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

	if s.Collapsed {
		return arrowStyle.Render("▸") + " " + titleStyle.Render(s.Title)
	}

	header := arrowStyle.Render("▾") + " " + titleStyle.Render(s.Title)
	content := contentFunc()
	if content == "" {
		return header
	}
	return header + "\n" + content
}
