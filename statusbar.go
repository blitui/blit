package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBarOpts configures a StatusBar.
type StatusBarOpts struct {
	Left  func() string // Dynamic left-aligned content
	Right func() string // Dynamic right-aligned content
}

// StatusBar renders a footer bar with left and right aligned sections.
type StatusBar struct {
	opts    StatusBarOpts
	width   int
	height  int
	theme   Theme
	focused bool
}

// NewStatusBar creates a new StatusBar with the given options.
func NewStatusBar(opts StatusBarOpts) *StatusBar {
	return &StatusBar{opts: opts}
}

func (s *StatusBar) Init() tea.Cmd { return nil }

func (s *StatusBar) Update(msg tea.Msg) (Component, tea.Cmd) {
	return s, nil
}

func (s *StatusBar) View() string {
	left := ""
	right := ""
	if s.opts.Left != nil {
		left = s.opts.Left()
	}
	if s.opts.Right != nil {
		right = s.opts.Right()
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Muted)).
		Width(s.width)

	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
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
