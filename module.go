package blit

import tea "github.com/charmbracelet/bubbletea"

// Module is the interface for pluggable app extensions that participate in
// the App lifecycle. Modules receive messages, can emit commands, and
// optionally provide debug information and keybindings.
type Module interface {
	// Name returns a unique identifier for this module (e.g., "devConsole", "poller").
	Name() string

	// Init is called once before the first frame, after all modules are registered.
	// Return a tea.Cmd for background initialization work (or nil).
	Init() tea.Cmd

	// Update receives every message the App processes. Return the updated module
	// and an optional command. Modules see messages after overlays but before
	// components in the dispatch order.
	Update(msg tea.Msg, ctx Context) (Module, tea.Cmd)
}

// DebugProvider supplies a named debug section for the DevConsole overlay.
// Providers are queried each frame when the console is visible.
type DebugProvider interface {
	// Name returns the display name for this provider (used as tab label or section header).
	Name() string

	// View renders the provider's content. width and height are the available space.
	View(width, height int, theme Theme) string
}

// DebugDataProvider is an optional extension of DebugProvider that exposes
// structured data for machine consumption (JSON export, alternative renders).
type DebugDataProvider interface {
	DebugProvider
	// Data returns structured key-value data for this provider.
	Data() map[string]any
}

// ModuleWithProviders is an optional extension of Module that contributes
// debug providers to the DevConsole.
type ModuleWithProviders interface {
	Module
	// Providers returns the debug providers this module contributes.
	Providers() []DebugProvider
}

// ModuleWithKeybinds is an optional extension of Module that contributes
// keybindings to the app's registry.
type ModuleWithKeybinds interface {
	Module
	// Keybinds returns the keybindings this module contributes.
	Keybinds() []KeyBind
}

// WithModule registers a module with the App. Modules are initialized in
// registration order during Init and receive all messages during Update.
func WithModule(m Module) Option {
	return func(a *appModel) {
		a.modules = append(a.modules, m)
	}
}

// WithDebugProvider registers a standalone debug provider with the DevConsole.
// For providers that are part of a module, implement ModuleWithProviders instead.
func WithDebugProvider(p DebugProvider) Option {
	return func(a *appModel) {
		if a.devConsole == nil {
			a.devConsole = newDevConsole()
		}
		a.devConsole.providers = append(a.devConsole.providers, p)
	}
}
