package blit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StepStatus describes the state of a step.
type StepStatus int

const (
	// StepPending is a step that has not yet been reached.
	StepPending StepStatus = iota
	// StepActive is the step currently in progress.
	StepActive
	// StepDone is a completed step.
	StepDone
)

// Step is a single stage in a Stepper.
type Step struct {
	// Title is the step label.
	Title string
	// Description is optional detail text.
	Description string
}

// StepperOpts configures a Stepper.
type StepperOpts struct {
	// OnComplete is called when the user advances past the last step.
	OnComplete func()

	// OnChange is called when the current step changes.
	OnChange func(step int)
}

// Stepper is a multi-step progress indicator with forward/back navigation.
// Implements Component and Themed.
type Stepper struct {
	opts    StepperOpts
	steps   []Step
	current int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewStepper creates a Stepper with the given steps and options.
func NewStepper(steps []Step, opts StepperOpts) *Stepper {
	return &Stepper{
		opts:  opts,
		steps: steps,
	}
}

// Steps returns the current steps.
func (s *Stepper) Steps() []Step { return s.steps }

// Current returns the index of the active step.
func (s *Stepper) Current() int { return s.current }

// SetCurrent sets the active step index, clamped to valid range.
func (s *Stepper) SetCurrent(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.steps) {
		idx = len(s.steps) - 1
	}
	s.current = idx
}

// Next advances to the next step. If already at the last step, calls OnComplete.
func (s *Stepper) Next() {
	if s.current >= len(s.steps)-1 {
		if s.opts.OnComplete != nil {
			s.opts.OnComplete()
		}
		return
	}
	s.current++
	if s.opts.OnChange != nil {
		s.opts.OnChange(s.current)
	}
}

// Prev moves back to the previous step.
func (s *Stepper) Prev() {
	if s.current > 0 {
		s.current--
		if s.opts.OnChange != nil {
			s.opts.OnChange(s.current)
		}
	}
}

// Status returns the status of a step at the given index.
func (s *Stepper) Status(idx int) StepStatus {
	switch {
	case idx < s.current:
		return StepDone
	case idx == s.current:
		return StepActive
	default:
		return StepPending
	}
}

// Init implements Component.
func (s *Stepper) Init() tea.Cmd { return nil }

// Update implements Component.
func (s *Stepper) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !s.focused {
		return s, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}
	switch km.String() {
	case "right", "l", "tab":
		s.Next()
		return s, Consumed()
	case "left", "h", "shift+tab":
		s.Prev()
		return s, Consumed()
	}
	return s, nil
}

// View implements Component.
func (s *Stepper) View() string {
	if s.width == 0 || s.height == 0 || len(s.steps) == 0 {
		return ""
	}

	g := s.theme.glyphsOrDefault()
	textColor := lipgloss.Color(s.theme.Text)
	accentColor := lipgloss.Color(s.theme.Accent)
	mutedColor := lipgloss.Color(s.theme.Muted)
	positiveColor := lipgloss.Color(s.theme.Positive)

	var parts []string
	for i, step := range s.steps {
		status := s.Status(i)

		// Step number/icon.
		var icon string
		var iconColor lipgloss.Color
		switch status {
		case StepDone:
			icon = g.Check
			iconColor = positiveColor
		case StepActive:
			icon = fmt.Sprintf("%d", i+1)
			iconColor = accentColor
		default:
			icon = fmt.Sprintf("%d", i+1)
			iconColor = mutedColor
		}

		iconStyled := lipgloss.NewStyle().Foreground(iconColor).Bold(status == StepActive).Render(icon)

		// Title.
		var titleStyled string
		switch status {
		case StepDone:
			titleStyled = lipgloss.NewStyle().Foreground(positiveColor).Render(step.Title)
		case StepActive:
			titleStyled = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(step.Title)
		default:
			titleStyled = lipgloss.NewStyle().Foreground(mutedColor).Render(step.Title)
		}

		part := iconStyled + " " + titleStyled
		parts = append(parts, part)

		// Connector between steps.
		if i < len(s.steps)-1 {
			connector := lipgloss.NewStyle().Foreground(mutedColor).Render(" ── ")
			if status == StepDone {
				connector = lipgloss.NewStyle().Foreground(positiveColor).Render(" ── ")
			}
			parts = append(parts, connector)
		}
	}

	line := strings.Join(parts, "")

	// Description below if active step has one.
	if s.current < len(s.steps) && s.steps[s.current].Description != "" {
		desc := lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(2).
			Render(s.steps[s.current].Description)
		line += "\n" + desc
	}

	return line
}

// KeyBindings implements Component.
func (s *Stepper) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "right/l/tab", Label: "Next step", Group: "STEPPER"},
		{Key: "left/h/shift+tab", Label: "Previous step", Group: "STEPPER"},
	}
}

// SetSize implements Component.
func (s *Stepper) SetSize(w, h int) {
	s.width = w
	s.height = h
}

// Focused implements Component.
func (s *Stepper) Focused() bool { return s.focused }

// SetFocused implements Component.
func (s *Stepper) SetFocused(f bool) { s.focused = f }

// SetTheme implements Themed.
func (s *Stepper) SetTheme(theme Theme) { s.theme = theme }
