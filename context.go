package blit

import (
	"log/slog"
)

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

	// Flags exposes the app-wide feature flag registry. Components can
	// check flags to conditionally enable or disable behavior at runtime.
	// Nil means no flags are configured — callers should treat this as
	// "all features at their defaults".
	Flags *FeatureFlag

	// TraceID identifies the current message dispatch cycle. Every call
	// to App.Update generates a unique trace ID so that components can
	// correlate log entries and metrics across the dispatch chain.
	// Components can log it or pass it to downstream systems for
	// end-to-end message tracing.
	TraceID uint64
}

// LogRedact wraps a value so that slog emits "[REDACTED]" instead of the
// actual content. Use it when logging sensitive data (tokens, keys, paths)
// that should not appear in log output.
func LogRedact(v any) slog.LogValuer {
	return redacted{v: v}
}

type redacted struct{ v any }

func (r redacted) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

var _ slog.LogValuer = redacted{}

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
