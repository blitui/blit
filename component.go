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
	// ctx carries the ambient Theme, Size, Focus, Hotkeys, Clock, and Logger
	// for the current dispatch.
	Update(msg tea.Msg, ctx Context) (Component, tea.Cmd)

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

// InlineOverlay is an overlay that renders as a single line at the bottom
// of the screen instead of replacing all content. The App renders the normal
// content behind it with the overlay's View appended at the bottom.
type InlineOverlay interface {
	Overlay
	Inline() bool
}

// FloatingOverlay is an overlay that composites on top of the main content
// rather than replacing it. The App renders the normal content first, then
// calls FloatView with the rendered content so the overlay can composite
// its panel over it (e.g., a dev console, tooltip, or pop-over).
type FloatingOverlay interface {
	Overlay
	// FloatView receives the fully-rendered background content and returns
	// the composited result with the floating panel drawn on top.
	FloatView(background string) string
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
