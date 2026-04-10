package blit

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerOpts configures a Spinner.
type SpinnerOpts struct {
	// Label is displayed next to the spinner animation.
	Label string

	// Interval is the time between frame advances. Defaults to 80ms.
	Interval time.Duration
}

// Spinner is an animated loading indicator that cycles through glyph frames.
// Implements Component and Themed.
type Spinner struct {
	opts    SpinnerOpts
	frame   int
	theme   Theme
	focused bool
	width   int
	height  int
	active  bool
}

// spinnerTickMsg advances the spinner to the next frame.
type spinnerTickMsg struct{}

// NewSpinner creates a Spinner with the given options.
func NewSpinner(opts SpinnerOpts) *Spinner {
	if opts.Interval <= 0 {
		opts.Interval = 80 * time.Millisecond
	}
	return &Spinner{
		opts:   opts,
		active: true,
	}
}

// Label returns the current label.
func (s *Spinner) Label() string { return s.opts.Label }

// SetLabel updates the spinner label.
func (s *Spinner) SetLabel(label string) { s.opts.Label = label }

// Active returns whether the spinner is animating.
func (s *Spinner) Active() bool { return s.active }

// SetActive starts or stops the animation. When restarted, call Init()
// again to schedule the next tick.
func (s *Spinner) SetActive(active bool) { s.active = active }

// Frame returns the current frame index.
func (s *Spinner) Frame() int { return s.frame }

// Init implements Component. It schedules the first tick.
func (s *Spinner) Init() tea.Cmd {
	if !s.active {
		return nil
	}
	return s.tick()
}

// Update implements Component.
func (s *Spinner) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !s.active {
		return s, nil
	}
	if _, ok := msg.(spinnerTickMsg); ok {
		frames := s.theme.glyphsOrDefault().SpinnerFrames
		if len(frames) > 0 {
			s.frame = (s.frame + 1) % len(frames)
		}
		return s, s.tick()
	}
	return s, nil
}

// View implements Component.
func (s *Spinner) View() string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	g := s.theme.glyphsOrDefault()
	frames := g.SpinnerFrames
	if len(frames) == 0 {
		return s.opts.Label
	}

	accentColor := lipgloss.Color(s.theme.Accent)
	textColor := lipgloss.Color(s.theme.Text)

	frameIdx := s.frame % len(frames)
	icon := lipgloss.NewStyle().Foreground(accentColor).Render(frames[frameIdx])

	if s.opts.Label == "" {
		return icon
	}

	label := lipgloss.NewStyle().Foreground(textColor).Render(s.opts.Label)
	return icon + " " + label
}

// KeyBindings implements Component.
func (s *Spinner) KeyBindings() []KeyBind { return nil }

// SetSize implements Component.
func (s *Spinner) SetSize(w, h int) {
	s.width = w
	s.height = h
}

// Focused implements Component.
func (s *Spinner) Focused() bool { return s.focused }

// SetFocused implements Component.
func (s *Spinner) SetFocused(f bool) { s.focused = f }

// SetTheme implements Themed.
func (s *Spinner) SetTheme(theme Theme) { s.theme = theme }

func (s *Spinner) tick() tea.Cmd {
	return tea.Tick(s.opts.Interval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}
