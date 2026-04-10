# API Reference

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go).

## Core

| Type | Description |
|------|-------------|
| [`App`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#App) | Application container configured via functional options |
| [`Component`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Component) | Interface all components implement (Init, Update, View, KeyBindings, SetSize, Focused, SetFocused) |
| [`Context`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Context) | Per-update context carrying Theme, Size, Focus, Hotkeys, Clock, Logger |
| [`Theme`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Theme) | Semantic color tokens (Positive, Negative, Accent, Muted, etc.) |
| [`Registry`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Registry) | Keybinding registry with conflict detection |
| [`KeyBind`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#KeyBind) | Keybinding definition (key, label, group, handler) |
| [`UpdateConfig`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#UpdateConfig) | Self-update configuration |

## Components

| Type | Description |
|------|-------------|
| [`Table`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Table) | Adaptive table with sorting, filtering, custom rendering, virtualization |
| [`ListView`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#ListView) | Generic scrollable list with cursor navigation |
| [`Tabs`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Tabs) | Tabbed container with horizontal/vertical orientation |
| [`Form`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Form) | Multi-field form with validation and wizard mode |
| [`Picker`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Picker) | Fuzzy-search selection list |
| [`Tree`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Tree) | Expandable tree view |
| [`FilePicker`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#FilePicker) | File system browser with tree navigation and preview |
| [`LogViewer`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#LogViewer) | Streaming log viewer with level filtering |
| [`Viewport`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Viewport) | Scrollable content pane |
| [`Markdown`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Markdown) | Glamour-powered markdown renderer |
| [`StatusBar`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#StatusBar) | Left/right footer driven by closures or signals |
| [`Help`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Help) | Auto-generated keybinding overlay |
| [`Breadcrumbs`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Breadcrumbs) | Navigation breadcrumb trail |
| [`ToastMsg`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#ToastMsg) | Toast notification message |

## Layout

| Type | Description |
|------|-------------|
| [`DualPane`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#DualPane) | Main + collapsible sidebar layout |
| [`HBox`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#HBox) | Horizontal flex container |
| [`VBox`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#VBox) | Vertical flex container |
| [`Split`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go#Split) | Resizable split pane |

## Packages

| Package | Description |
|---------|-------------|
| [`cli`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go/cli) | Interactive CLI prompts (Confirm, Select, Input, Spinner, Progress) |
| [`charts`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go/charts) | Chart components (Bar, Line, Ring, Gauge, Heatmap) |
| [`tuitest`](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go/tuitest) | Virtual terminal testing framework |
