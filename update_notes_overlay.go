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
type ReleaseNotesOverlay struct {
	Title   string
	Version string
	Lines   []string
	Offset  int
	Width   int
	Height  int
	Closed  bool
}

// NewReleaseNotesOverlay constructs an overlay for the given release
// notes text. Width/height default to 80x24.
func NewReleaseNotesOverlay(version, notes string) *ReleaseNotesOverlay {
	return &ReleaseNotesOverlay{
		Title:   "Release notes",
		Version: version,
		Lines:   splitLines(notes),
		Width:   80,
		Height:  24,
	}
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return []string{"(no release notes)"}
	}
	return strings.Split(s, "\n")
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

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).
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
		BorderForeground(lipgloss.Color("14")).
		Padding(0, 1).
		Render(content)
	return box
}
