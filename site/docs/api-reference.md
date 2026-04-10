# API Reference

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/blitui/blit).

## Core

| Type | Description |
|------|-------------|
| [`App`](https://pkg.go.dev/github.com/blitui/blit#App) | Application container configured via functional options |
| [`Component`](https://pkg.go.dev/github.com/blitui/blit#Component) | Interface all components implement (Init, Update, View, KeyBindings, SetSize, Focused, SetFocused) |
| [`Context`](https://pkg.go.dev/github.com/blitui/blit#Context) | Per-update context carrying Theme, Size, Focus, Hotkeys, Clock, Logger |
| [`Theme`](https://pkg.go.dev/github.com/blitui/blit#Theme) | Semantic color tokens (Positive, Negative, Accent, Muted, etc.) |
| [`Registry`](https://pkg.go.dev/github.com/blitui/blit#Registry) | Keybinding registry with conflict detection |
| [`KeyBind`](https://pkg.go.dev/github.com/blitui/blit#KeyBind) | Keybinding definition (key, label, group, handler) |
| [`UpdateConfig`](https://pkg.go.dev/github.com/blitui/blit#UpdateConfig) | Self-update configuration |

## Components

| Type | Description |
|------|-------------|
| [`Table`](https://pkg.go.dev/github.com/blitui/blit#Table) | Adaptive table with sorting, filtering, custom rendering, virtualization |
| [`ListView`](https://pkg.go.dev/github.com/blitui/blit#ListView) | Generic scrollable list with cursor navigation |
| [`Tabs`](https://pkg.go.dev/github.com/blitui/blit#Tabs) | Tabbed container with horizontal/vertical orientation |
| [`Form`](https://pkg.go.dev/github.com/blitui/blit#Form) | Multi-field form with validation and wizard mode |
| [`Picker`](https://pkg.go.dev/github.com/blitui/blit#Picker) | Fuzzy-search selection list |
| [`Tree`](https://pkg.go.dev/github.com/blitui/blit#Tree) | Expandable tree view |
| [`FilePicker`](https://pkg.go.dev/github.com/blitui/blit#FilePicker) | File system browser with tree navigation and preview |
| [`LogViewer`](https://pkg.go.dev/github.com/blitui/blit#LogViewer) | Streaming log viewer with level filtering |
| [`Viewport`](https://pkg.go.dev/github.com/blitui/blit#Viewport) | Scrollable content pane |
| [`Markdown`](https://pkg.go.dev/github.com/blitui/blit#Markdown) | Glamour-powered markdown renderer |
| [`StatusBar`](https://pkg.go.dev/github.com/blitui/blit#StatusBar) | Left/right footer driven by closures or signals |
| [`Help`](https://pkg.go.dev/github.com/blitui/blit#Help) | Auto-generated keybinding overlay |
| [`Breadcrumbs`](https://pkg.go.dev/github.com/blitui/blit#Breadcrumbs) | Navigation breadcrumb trail |
| [`ToastMsg`](https://pkg.go.dev/github.com/blitui/blit#ToastMsg) | Toast notification message |

## Layout

| Type | Description |
|------|-------------|
| [`DualPane`](https://pkg.go.dev/github.com/blitui/blit#DualPane) | Main + collapsible sidebar layout |
| [`HBox`](https://pkg.go.dev/github.com/blitui/blit#HBox) | Horizontal flex container |
| [`VBox`](https://pkg.go.dev/github.com/blitui/blit#VBox) | Vertical flex container |
| [`Split`](https://pkg.go.dev/github.com/blitui/blit#Split) | Resizable split pane |

## Packages

| Package | Description |
|---------|-------------|
| [`cli`](https://pkg.go.dev/github.com/blitui/blit/cli) | Interactive CLI prompts (Confirm, Select, Input, Spinner, Progress) |
| [`charts`](https://pkg.go.dev/github.com/blitui/blit/charts) | Chart components (Bar, Line, Ring, Gauge, Heatmap) |
| [`blit`](https://pkg.go.dev/github.com/blitui/blit/blit) | Virtual terminal testing framework |
