package tuikit

import tea "github.com/charmbracelet/bubbletea"

// Alignment controls text alignment in table columns.
type Alignment int

const (
	// Left aligns text to the left (default).
	Left Alignment = iota
	// Right aligns text to the right.
	Right
	// Center aligns text to the center.
	Center
)

// Component is the core interface for all tuikit UI elements.
// It extends Bubble Tea's Model with size, focus, and keybinding support.
type Component interface {
	// Init returns an initial command, like tea.Model.
	Init() tea.Cmd

	// Update handles a message and returns the updated component and command.
	// Return Consumed() to signal the App that this key was handled.
	Update(msg tea.Msg) (Component, tea.Cmd)

	// View renders the component to a string.
	View() string

	// KeyBindings returns the keybindings this component handles.
	KeyBindings() []KeyBind

	// SetSize sets the available width and height for rendering.
	SetSize(width, height int)

	// Focused returns whether this component currently has focus.
	Focused() bool

	// SetFocused sets the focus state of this component.
	SetFocused(focused bool)
}

// Overlay is a modal view that stacks on top of the main UI.
// The App manages an overlay stack — Esc pops the top overlay.
type Overlay interface {
	Component

	// IsActive returns whether this overlay is currently visible.
	IsActive() bool

	// Close hides this overlay.
	Close()
}

// Themed is an optional interface for components that accept a theme.
// The App automatically calls SetTheme on any Component or Overlay that
// implements this interface. Built-in components (Table, StatusBar, Help,
// ConfigEditor) all implement Themed. Custom components should too if they
// need access to the theme's semantic color tokens.
type Themed interface {
	SetTheme(Theme)
}

// consumedMsg signals that a component handled a key event.
type consumedMsg struct{}

// Consumed returns a tea.Cmd that signals the App that a key was handled.
// When a component's Update returns this, the App stops dispatching the key
// to other components.
func Consumed() tea.Cmd {
	return func() tea.Msg { return consumedMsg{} }
}

// isConsumed checks whether a tea.Cmd will produce a consumedMsg.
func isConsumed(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(consumedMsg)
	return ok
}
