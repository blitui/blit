package blit

// Compile-time interface assertions for all built-in components.
// These ensure every public component type satisfies Component and Themed.
// If a component fails to implement an interface method, this file will
// cause a build error pointing at the offending type.

// --- Component interface ---

var _ Component = (*Table)(nil)
var _ Component = (*Tabs)(nil)
var _ Component = (*Form)(nil)
var _ Component = (*Picker)(nil)
var _ Component = (*FilePicker)(nil)
var _ Component = (*Tree)(nil)
var _ Component = (*LogViewer)(nil)
var _ Component = (*Viewport)(nil)
var _ Component = (*Help)(nil)
var _ Component = (*StatusBar)(nil)
var _ Component = (*ConfigEditor)(nil)
var _ Component = (*CommandBar)(nil)
var _ Component = (*Breadcrumbs)(nil)
var _ Component = (*Split)(nil)
var _ Component = (*HBox)(nil)
var _ Component = (*VBox)(nil)

// --- Themed interface ---

var _ Themed = (*Table)(nil)
var _ Themed = (*Tabs)(nil)
var _ Themed = (*Form)(nil)
var _ Themed = (*Picker)(nil)
var _ Themed = (*FilePicker)(nil)
var _ Themed = (*Tree)(nil)
var _ Themed = (*LogViewer)(nil)
var _ Themed = (*Viewport)(nil)
var _ Themed = (*Help)(nil)
var _ Themed = (*StatusBar)(nil)
var _ Themed = (*ConfigEditor)(nil)
var _ Themed = (*CommandBar)(nil)
var _ Themed = (*Breadcrumbs)(nil)
var _ Themed = (*Split)(nil)

// --- Overlay interfaces ---

var _ Overlay = (*DetailOverlay[any])(nil)
var _ Overlay = (*Help)(nil)
var _ Overlay = (*ConfigEditor)(nil)

var _ InlineOverlay = (*CommandBar)(nil)

var _ FloatingOverlay = (*devConsole)(nil)
