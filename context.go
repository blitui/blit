package tuikit

import "log/slog"

// Context carries ambient state that every Component.Update call receives.
// It is constructed by the App on each dispatch and passed down through the
// focused-component chain. Prefer reading values from Context over stashing
// copies on the component itself — Context always reflects the latest frame.
type Context struct {
	// Theme is the currently active theme. Components can pull semantic
	// colors from it without needing a separate SetTheme call per frame.
	Theme Theme

	// Size is the viewport size at the time of dispatch (width, height).
	Size Size

	// Focus describes which component currently holds focus. Components
	// can compare against their own identity / index to decide whether to
	// react to key input.
	Focus Focus

	// Hotkeys exposes the app-wide keybinding registry. Components can
	// consult it to resolve chord conflicts or render inline hints.
	Hotkeys *Registry

	// Clock abstracts time.Now so components can be tested with a fake
	// clock. It is nil-safe: if unset, callers should fall back to
	// time.Now directly.
	Clock Clock

	// Logger is an optional structured logger. Components should treat a
	// nil Logger as "logging disabled" rather than panicking.
	Logger *slog.Logger
}

// Size is the width and height available to a component.
type Size struct {
	Width  int
	Height int
}

// Focus identifies the currently focused component in the app's focus chain.
type Focus struct {
	// Index is the position of the focused component in the app's focus
	// order. -1 means no component has focus (e.g., an overlay is active).
	Index int

	// Name is an optional human-readable name for the focused component,
	// sourced from WithComponent / DualPane.MainName / DualPane.SideName.
	Name string
}
