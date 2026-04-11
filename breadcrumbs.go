package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Breadcrumbs is a passive, display-only path trail. It renders a horizontal
// sequence of path segments separated by a configurable separator. When the
// rendered width exceeds MaxWidth the leading segments are replaced with "…"
// until it fits.
//
// For an interactive breadcrumb with cursor navigation and selection, see
// Breadcrumb.
type Breadcrumbs struct {
	// Segments are the individual path parts, e.g. []string{"home", "docs", "api"}.
	Segments []string

	// Separator is placed between segments. Defaults to " / ".
	Separator string

	// MaxWidth is the maximum render width in columns. 0 means unlimited.
	MaxWidth int

	theme   Theme
	focused bool
	width   int
	height  int
}

// NewBreadcrumbs creates a Breadcrumbs component with default separator " / ".
func NewBreadcrumbs(segments []string) *Breadcrumbs {
	return &Breadcrumbs{
		Segments:  segments,
		Separator: " / ",
	}
}

// Init initializes the Breadcrumbs component. It returns nil.
func (b *Breadcrumbs) Init() tea.Cmd { return nil }
// Update handles incoming messages and updates Breadcrumbs state.
func (b *Breadcrumbs) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	return b, nil
}
// KeyBindings returns the key bindings for the Breadcrumbs.
func (b *Breadcrumbs) KeyBindings() []KeyBind { return nil }
// SetSize sets the width and height of the Breadcrumbs.
func (b *Breadcrumbs) SetSize(w, h int)       { b.width = w; b.height = h }
// Focused reports whether the Breadcrumbs is focused.
func (b *Breadcrumbs) Focused() bool          { return b.focused }
// SetFocused sets the focus state of the Breadcrumbs.
func (b *Breadcrumbs) SetFocused(f bool)      { b.focused = f }
// SetTheme updates the theme used by the Breadcrumbs.
func (b *Breadcrumbs) SetTheme(th Theme)      { b.theme = th }

// View renders the Breadcrumbs as a string.
func (b *Breadcrumbs) View() string {
	sep := b.Separator
	if sep == "" {
		sep = " / "
	}

	maxW := b.MaxWidth
	if maxW == 0 && b.width > 0 {
		maxW = b.width
	}

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.theme.Muted))
	segStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.theme.Text))
	lastStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.theme.Accent)).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.theme.Muted))

	render := func(segs []string, ellipsis bool) string {
		var parts []string
		if ellipsis {
			parts = append(parts, mutedStyle.Render("…"))
		}
		for i, seg := range segs {
			if i == len(segs)-1 {
				parts = append(parts, lastStyle.Render(seg))
			} else {
				parts = append(parts, segStyle.Render(seg))
			}
			if i < len(segs)-1 {
				parts = append(parts, sepStyle.Render(sep))
			}
		}
		return strings.Join(parts, "")
	}

	if maxW <= 0 || len(b.Segments) == 0 {
		return render(b.Segments, false)
	}

	// Try to fit all segments; if not, drop from the front.
	segs := b.Segments
	ellipsis := false
	for {
		line := render(segs, ellipsis)
		if lipgloss.Width(line) <= maxW || len(segs) <= 1 {
			return line
		}
		segs = segs[1:]
		ellipsis = true
	}
}
