package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReleaseNotesOverlay is a scrollable overlay that displays release notes
// text from an UpdateResult. It is meant to be opened from the notify
// banner or the forced update screen via a keybind ("n" by convention).
//
// The overlay is self-contained: it tracks scroll offset, renders a
// bordered box, and closes itself when the user presses q/esc.
//
// When a Theme has been set via SetTheme, release notes are rendered through
// Markdown() so that headings, code spans, and lists are styled. BREAKING and
// SECURITY section headings are additionally highlighted using the theme's
// Negative and Flash color tokens.
type ReleaseNotesOverlay struct {
	Title   string
	Version string
	Lines   []string
	Offset  int
	Width   int
	Height  int
	Closed  bool
	raw     string // original markdown source, preserved for re-rendering on theme change
	theme   Theme
	themed  bool // true once SetTheme has been called
}

// NewReleaseNotesOverlay constructs an overlay for the given release
// notes text. Width/height default to 80x24.
func NewReleaseNotesOverlay(version, notes string) *ReleaseNotesOverlay {
	return &ReleaseNotesOverlay{
		Title:   "Release notes",
		Version: version,
		Lines:   splitNoteLines(notes),
		Width:   80,
		Height:  24,
		raw:     notes,
	}
}

// NewReleaseNotesOverlayThemed is like NewReleaseNotesOverlay but immediately
// applies a theme so the notes are rendered through Markdown from the start.
func NewReleaseNotesOverlayThemed(version, notes string, t Theme) *ReleaseNotesOverlay {
	return &ReleaseNotesOverlay{
		Title:   "Release notes",
		Version: version,
		Lines:   renderNotes(notes, t),
		Width:   80,
		Height:  24,
		raw:     notes,
		theme:   t,
		themed:  true,
	}
}

// SetTheme implements the Themed interface. When set, release notes are
// rendered through Markdown() with BREAKING/SECURITY section highlighting.
func (o *ReleaseNotesOverlay) SetTheme(t Theme) {
	o.theme = t
	o.themed = true
	o.Lines = renderNotes(o.raw, t)
}

func splitNoteLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return []string{"(no release notes)"}
	}
	return strings.Split(s, "\n")
}

// renderNotes renders notes through Markdown (with BREAKING/SECURITY
// highlighting) and splits the result into lines.
func renderNotes(notes string, t Theme) []string {
	if notes == "" {
		return []string{"(no release notes)"}
	}
	rendered := Markdown(notes, t)
	rendered = highlightSections(rendered, t)
	return strings.Split(strings.ReplaceAll(rendered, "\r\n", "\n"), "\n")
}

// Init implements tea.Model.
func (o *ReleaseNotesOverlay) Init() tea.Cmd { return nil }

// VisibleLines returns the number of content rows rendered per frame.
// Two lines are reserved for the header and one for the footer, plus
// the border/padding (4). Minimum of 1.
func (o *ReleaseNotesOverlay) VisibleLines() int {
	n := o.Height - 6
	if n < 1 {
		return 1
	}
	return n
}

// MaxOffset clamps scroll position to the last full page.
func (o *ReleaseNotesOverlay) MaxOffset() int {
	max := len(o.Lines) - o.VisibleLines()
	if max < 0 {
		return 0
	}
	return max
}

// Update implements tea.Model.
func (o *ReleaseNotesOverlay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		o.Width = m.Width
		o.Height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "q", "esc":
			o.Closed = true
		case "up", "k":
			if o.Offset > 0 {
				o.Offset--
			}
		case "down", "j":
			if o.Offset < o.MaxOffset() {
				o.Offset++
			}
		case "pgup":
			o.Offset -= o.VisibleLines()
			if o.Offset < 0 {
				o.Offset = 0
			}
		case "pgdown", " ":
			o.Offset += o.VisibleLines()
			if o.Offset > o.MaxOffset() {
				o.Offset = o.MaxOffset()
			}
		case "home", "g":
			o.Offset = 0
		case "end", "G":
			o.Offset = o.MaxOffset()
		}
	}
	return o, nil
}

// View implements tea.Model.
func (o *ReleaseNotesOverlay) View() string {
	visible := o.VisibleLines()
	end := o.Offset + visible
	if end > len(o.Lines) {
		end = len(o.Lines)
	}
	page := o.Lines[o.Offset:end]

	borderColor := lipgloss.Color("14")
	titleFg := lipgloss.Color("14")
	if o.themed {
		borderColor = o.theme.Accent
		titleFg = o.theme.Accent
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(titleFg).
		Render(o.Title + "  " + o.Version)
	body := strings.Join(page, "\n")

	max := o.MaxOffset()
	pos := "top"
	switch {
	case max == 0:
		pos = "all"
	case o.Offset >= max:
		pos = "end"
	case o.Offset > 0:
		pos = "mid"
	}
	footer := lipgloss.NewStyle().Faint(true).Render(
		"[↑/↓] scroll   [pgup/pgdn] page   [q]uit   " + pos,
	)

	content := title + "\n\n" + body + "\n\n" + footer
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(content)
	return box
}
