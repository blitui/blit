package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBarOpts configures a StatusBar.
//
// Left and Right accept any value that can be converted into a StringSource:
// a plain `func() string` closure (legacy), a `*Signal[string]` (reactive,
// v0.10+), or an explicit StringSource. Nil means "empty". The overload is
// intentional so existing call sites keep working while new code can hand
// in a signal for per-frame reactivity without plumbing a getter closure
// that captures the signal.
type StatusBarOpts struct {
	Left  any // func() string, *Signal[string], or StringSource
	Right any
}

// StatusBar renders a footer bar with left and right aligned sections.
type StatusBar struct {
	left    StringSource
	right   StringSource
	width   int
	height  int
	theme   Theme
	focused bool
}

// NewStatusBar creates a new StatusBar with the given options.
//
// Left and Right in opts accept either a `func() string` closure or a
// `*Signal[string]`. See StatusBarOpts for details.
func NewStatusBar(opts StatusBarOpts) *StatusBar {
	return &StatusBar{
		left:  toStringSource(opts.Left),
		right: toStringSource(opts.Right),
	}
}

func (s *StatusBar) Init() tea.Cmd { return nil }

func (s *StatusBar) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	return s, nil
}

func (s *StatusBar) View() string {
	left := ""
	right := ""
	if s.left != nil {
		left = s.left.Value()
	}
	if s.right != nil {
		right = s.right.Value()
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Muted)).
		Width(s.width).
		MaxHeight(1)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := s.width - leftW - rightW
	if gap < 0 {
		gap = 0
		maxLeft := s.width - rightW
		if maxLeft < 0 {
			maxLeft = 0
		}
		left = lipgloss.NewStyle().MaxWidth(maxLeft).Render(left)
	}

	content := left + strings.Repeat(" ", gap) + right
	return style.Render(content)
}

func (s *StatusBar) KeyBindings() []KeyBind { return nil }
func (s *StatusBar) SetSize(w, h int)       { s.width = w; s.height = h }
func (s *StatusBar) Focused() bool           { return s.focused }
func (s *StatusBar) SetFocused(f bool)       { s.focused = f }
// SetTheme implements the Themed interface.
func (s *StatusBar) SetTheme(t Theme)        { s.theme = t }
