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
| [`StatusBar`](https://pkg.go.dev/github.com/blitui/blit#StatusBar) | Left/right footer driven by closures or signals |
| [`Help`](https://pkg.go.dev/github.com/blitui/blit#Help) | Auto-generated keybinding overlay |
| [`Breadcrumbs`](https://pkg.go.dev/github.com/blitui/blit#Breadcrumbs) | Navigation breadcrumb trail |
| [`ProgressBar`](https://pkg.go.dev/github.com/blitui/blit#ProgressBar) | Value-based progress indicator with label and percentage |
| [`Spinner`](https://pkg.go.dev/github.com/blitui/blit#Spinner) | Animated loading indicator cycling through glyph frames |
| [`Accordion`](https://pkg.go.dev/github.com/blitui/blit#Accordion) | Collapsible sections with exclusive mode |
| [`Breadcrumb`](https://pkg.go.dev/github.com/blitui/blit#Breadcrumb) | Navigable path display with Push/Pop API |
| [`Timeline`](https://pkg.go.dev/github.com/blitui/blit#Timeline) | Vertical/horizontal event sequence with status icons |
| [`Kanban`](https://pkg.go.dev/github.com/blitui/blit#Kanban) | Multi-column board with card movement |
| [`ChartPanel`](https://pkg.go.dev/github.com/blitui/blit#ChartPanel) | Switchable container for chart components |
| [`Stepper`](https://pkg.go.dev/github.com/blitui/blit#Stepper) | Multi-step progress indicator with navigation |
| [`CommandBar`](https://pkg.go.dev/github.com/blitui/blit#CommandBar) | Inline command palette with tab completion |
| [`ConfigEditor`](https://pkg.go.dev/github.com/blitui/blit#ConfigEditor) | Settings overlay with grouped fields and validation |
| [`ToastMsg`](https://pkg.go.dev/github.com/blitui/blit#ToastMsg) | Toast notification message |

## Overlays

| Type | Description |
|------|-------------|
| [`Dialog`](https://pkg.go.dev/github.com/blitui/blit#Dialog) | Modal dialog with button navigation |
| [`Menu`](https://pkg.go.dev/github.com/blitui/blit#Menu) | Popup menu with separators and shortcuts |
| [`Tooltip`](https://pkg.go.dev/github.com/blitui/blit#Tooltip) | Floating hint composited over background |

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
| [`btest`](https://pkg.go.dev/github.com/blitui/blit/btest) | Virtual terminal testing framework |
