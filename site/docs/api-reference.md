# API Reference

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/blitui/blit).

## Core

| Type | Description |
|------|-------------|
| [`App`](https://pkg.go.dev/github.com/blitui/blit#App) | Application container configured via functional options |
| [`Component`](https://pkg.go.dev/github.com/blitui/blit#Component) | Interface all components implement (Init, Update, View, KeyBindings, SetSize, Focused, SetFocused) |
| [`Context`](https://pkg.go.dev/github.com/blitui/blit#Context) | Per-update context carrying Theme, Size, Focus, Hotkeys, Clock, Logger |
| [`Option`](https://pkg.go.dev/github.com/blitui/blit#Option) | Functional option for `NewApp` configuration |
| [`Theme`](https://pkg.go.dev/github.com/blitui/blit#Theme) | Semantic color tokens (Positive, Negative, Accent, Muted, etc.) |
| [`Themed`](https://pkg.go.dev/github.com/blitui/blit#Themed) | Interface for components that accept a theme via `SetTheme` |
| [`Registry`](https://pkg.go.dev/github.com/blitui/blit#Registry) | Keybinding registry with conflict detection |
| [`KeyBind`](https://pkg.go.dev/github.com/blitui/blit#KeyBind) | Keybinding definition (key, label, group, handler) |
| [`KeyGroup`](https://pkg.go.dev/github.com/blitui/blit#KeyGroup) | Named group of keybindings for the help overlay |
| [`SlotName`](https://pkg.go.dev/github.com/blitui/blit#SlotName) | Named layout slot (Main, Sidebar, Footer) |
| [`Size`](https://pkg.go.dev/github.com/blitui/blit#Size) | Width/height pair |
| [`Focus`](https://pkg.go.dev/github.com/blitui/blit#Focus) | Focus state for components |

## Interfaces

| Type | Description |
|------|-------------|
| [`Activatable`](https://pkg.go.dev/github.com/blitui/blit#Activatable) | Components that can be activated/deactivated (overlays) |
| [`Overlay`](https://pkg.go.dev/github.com/blitui/blit#Overlay) | Full-screen overlay (IsActive, Close) |
| [`InlineOverlay`](https://pkg.go.dev/github.com/blitui/blit#InlineOverlay) | Single-line overlay (e.g., CommandBar) |
| [`FloatingOverlay`](https://pkg.go.dev/github.com/blitui/blit#FloatingOverlay) | Positioned overlay composited over background (e.g., Tooltip) |
| [`Layout`](https://pkg.go.dev/github.com/blitui/blit#Layout) | Layout container interface |
| [`Sized`](https://pkg.go.dev/github.com/blitui/blit#Sized) | Components that accept size via `SetSize` |
| [`InputCapture`](https://pkg.go.dev/github.com/blitui/blit#InputCapture) | Components that capture keyboard input |
| [`Clock`](https://pkg.go.dev/github.com/blitui/blit#Clock) | Time provider for animations and ticks |
| [`Module`](https://pkg.go.dev/github.com/blitui/blit#Module) | Lifecycle module interface (Init, Start, Stop, Status) |
| [`ModuleWithKeybinds`](https://pkg.go.dev/github.com/blitui/blit#ModuleWithKeybinds) | Module that registers keybindings |
| [`ModuleWithProviders`](https://pkg.go.dev/github.com/blitui/blit#ModuleWithProviders) | Module that provides debug data |
| [`DebugProvider`](https://pkg.go.dev/github.com/blitui/blit#DebugProvider) | Dev console data provider |
| [`DebugDataProvider`](https://pkg.go.dev/github.com/blitui/blit#DebugDataProvider) | Dev console data source |
| [`StringSource`](https://pkg.go.dev/github.com/blitui/blit#StringSource) | Reactive string value source |
| [`BackoffStrategy`](https://pkg.go.dev/github.com/blitui/blit#BackoffStrategy) | Retry backoff strategy |
| [`TableRowProvider`](https://pkg.go.dev/github.com/blitui/blit#TableRowProvider) | Virtual table data source for large datasets |

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
| [`Accordion`](https://pkg.go.dev/github.com/blitui/blit#Accordion) | Collapsible sections with exclusive mode |
| [`Breadcrumb`](https://pkg.go.dev/github.com/blitui/blit#Breadcrumb) | Navigable path display with Push/Pop API |
| [`Breadcrumbs`](https://pkg.go.dev/github.com/blitui/blit#Breadcrumbs) | Navigation breadcrumb trail |
| [`ChartPanel`](https://pkg.go.dev/github.com/blitui/blit#ChartPanel) | Switchable container for chart components |
| [`CommandBar`](https://pkg.go.dev/github.com/blitui/blit#CommandBar) | Inline command palette with tab completion |
| [`ConfigEditor`](https://pkg.go.dev/github.com/blitui/blit#ConfigEditor) | Settings overlay with grouped fields and validation |
| [`CollapsibleSection`](https://pkg.go.dev/github.com/blitui/blit#CollapsibleSection) | Toggleable content section |
| [`Kanban`](https://pkg.go.dev/github.com/blitui/blit#Kanban) | Multi-column board with card movement |
| [`ProgressBar`](https://pkg.go.dev/github.com/blitui/blit#ProgressBar) | Value-based progress indicator with label and percentage |
| [`Spinner`](https://pkg.go.dev/github.com/blitui/blit#Spinner) | Animated loading indicator cycling through glyph frames |
| [`Stepper`](https://pkg.go.dev/github.com/blitui/blit#Stepper) | Multi-step progress indicator with navigation |
| [`Timeline`](https://pkg.go.dev/github.com/blitui/blit#Timeline) | Vertical/horizontal event sequence with status icons |

## Overlays

| Type | Description |
|------|-------------|
| [`Dialog`](https://pkg.go.dev/github.com/blitui/blit#Dialog) | Modal dialog with button navigation |
| [`Menu`](https://pkg.go.dev/github.com/blitui/blit#Menu) | Popup menu with separators and shortcuts |
| [`Tooltip`](https://pkg.go.dev/github.com/blitui/blit#Tooltip) | Floating hint composited over background |
| [`DetailOverlay`](https://pkg.go.dev/github.com/blitui/blit#DetailOverlay) | Generic detail view overlay |
| [`ReleaseNotesOverlay`](https://pkg.go.dev/github.com/blitui/blit#ReleaseNotesOverlay) | Release notes display overlay |
| [`ForcedUpdateScreen`](https://pkg.go.dev/github.com/blitui/blit#ForcedUpdateScreen) | Forced update prompt screen |
| [`ToastMsg`](https://pkg.go.dev/github.com/blitui/blit#ToastMsg) | Toast notification message |

## Layout

| Type | Description |
|------|-------------|
| [`DualPane`](https://pkg.go.dev/github.com/blitui/blit#DualPane) | Main + collapsible sidebar layout |
| [`SinglePane`](https://pkg.go.dev/github.com/blitui/blit#SinglePane) | Single component layout |
| [`HBox`](https://pkg.go.dev/github.com/blitui/blit#HBox) | Horizontal flex container |
| [`VBox`](https://pkg.go.dev/github.com/blitui/blit#VBox) | Vertical flex container |
| [`Flex`](https://pkg.go.dev/github.com/blitui/blit#Flex) | Flexible layout container |
| [`Split`](https://pkg.go.dev/github.com/blitui/blit#Split) | Resizable split pane |

## Form Fields

| Type | Description |
|------|-------------|
| [`Field`](https://pkg.go.dev/github.com/blitui/blit#Field) | Interface for form field types |
| [`TextField`](https://pkg.go.dev/github.com/blitui/blit#TextField) | Single-line text input |
| [`PasswordField`](https://pkg.go.dev/github.com/blitui/blit#PasswordField) | Masked password input |
| [`NumberField`](https://pkg.go.dev/github.com/blitui/blit#NumberField) | Numeric input with validation |
| [`SelectField`](https://pkg.go.dev/github.com/blitui/blit#SelectField) | Single-select dropdown |
| [`MultiSelectField`](https://pkg.go.dev/github.com/blitui/blit#MultiSelectField) | Multi-select checkbox list |
| [`ConfirmField`](https://pkg.go.dev/github.com/blitui/blit#ConfirmField) | Boolean yes/no toggle |
| [`Validator`](https://pkg.go.dev/github.com/blitui/blit#Validator) | Field validation function |

## Configuration Types

| Type | Description |
|------|-------------|
| [`Column`](https://pkg.go.dev/github.com/blitui/blit#Column) | Table column definition |
| [`Row`](https://pkg.go.dev/github.com/blitui/blit#Row) | Table row (string slice) |
| [`TableOpts`](https://pkg.go.dev/github.com/blitui/blit#TableOpts) | Table configuration options |
| [`TabItem`](https://pkg.go.dev/github.com/blitui/blit#TabItem) | Tab definition with label and content |
| [`TabsOpts`](https://pkg.go.dev/github.com/blitui/blit#TabsOpts) | Tabs configuration options |
| [`FormOpts`](https://pkg.go.dev/github.com/blitui/blit#FormOpts) | Form configuration options |
| [`FormGroup`](https://pkg.go.dev/github.com/blitui/blit#FormGroup) | Form field group |
| [`TreeOpts`](https://pkg.go.dev/github.com/blitui/blit#TreeOpts) | Tree configuration options |
| [`Node`](https://pkg.go.dev/github.com/blitui/blit#Node) | Tree node with children |
| [`PickerItem`](https://pkg.go.dev/github.com/blitui/blit#PickerItem) | Picker entry with label and metadata |
| [`PickerOpts`](https://pkg.go.dev/github.com/blitui/blit#PickerOpts) | Picker configuration options |
| [`FilePickerOpts`](https://pkg.go.dev/github.com/blitui/blit#FilePickerOpts) | FilePicker configuration options |
| [`LogLine`](https://pkg.go.dev/github.com/blitui/blit#LogLine) | Log entry with level, timestamp, and message |
| [`LogLevel`](https://pkg.go.dev/github.com/blitui/blit#LogLevel) | Log severity level (Debug, Info, Warn, Error) |
| [`StatusBarOpts`](https://pkg.go.dev/github.com/blitui/blit#StatusBarOpts) | StatusBar configuration options |
| [`AccordionSection`](https://pkg.go.dev/github.com/blitui/blit#AccordionSection) | Accordion section with title and content |
| [`AccordionOpts`](https://pkg.go.dev/github.com/blitui/blit#AccordionOpts) | Accordion configuration options |
| [`BreadcrumbItem`](https://pkg.go.dev/github.com/blitui/blit#BreadcrumbItem) | Breadcrumb entry with label and data |
| [`BreadcrumbOpts`](https://pkg.go.dev/github.com/blitui/blit#BreadcrumbOpts) | Breadcrumb configuration options |
| [`Command`](https://pkg.go.dev/github.com/blitui/blit#Command) | CommandBar command definition |
| [`ConfigField`](https://pkg.go.dev/github.com/blitui/blit#ConfigField) | ConfigEditor field definition |
| [`DialogButton`](https://pkg.go.dev/github.com/blitui/blit#DialogButton) | Dialog action button |
| [`DialogOpts`](https://pkg.go.dev/github.com/blitui/blit#DialogOpts) | Dialog configuration options |
| [`MenuItem`](https://pkg.go.dev/github.com/blitui/blit#MenuItem) | Menu entry with label, shortcut, and action |
| [`MenuOpts`](https://pkg.go.dev/github.com/blitui/blit#MenuOpts) | Menu configuration options |
| [`KanbanCard`](https://pkg.go.dev/github.com/blitui/blit#KanbanCard) | Kanban card with title, description, and tag |
| [`KanbanColumn`](https://pkg.go.dev/github.com/blitui/blit#KanbanColumn) | Kanban column with cards |
| [`KanbanOpts`](https://pkg.go.dev/github.com/blitui/blit#KanbanOpts) | Kanban configuration options |
| [`ProgressBarOpts`](https://pkg.go.dev/github.com/blitui/blit#ProgressBarOpts) | ProgressBar configuration options |
| [`SpinnerOpts`](https://pkg.go.dev/github.com/blitui/blit#SpinnerOpts) | Spinner configuration options |
| [`Step`](https://pkg.go.dev/github.com/blitui/blit#Step) | Stepper step definition |
| [`StepperOpts`](https://pkg.go.dev/github.com/blitui/blit#StepperOpts) | Stepper configuration options |
| [`StepStatus`](https://pkg.go.dev/github.com/blitui/blit#StepStatus) | Step state (Pending, Active, Done) |
| [`TimelineEvent`](https://pkg.go.dev/github.com/blitui/blit#TimelineEvent) | Timeline event with time, title, and status |
| [`TimelineOpts`](https://pkg.go.dev/github.com/blitui/blit#TimelineOpts) | Timeline configuration options |
| [`TooltipOpts`](https://pkg.go.dev/github.com/blitui/blit#TooltipOpts) | Tooltip configuration options |
| [`ChartPanelOpts`](https://pkg.go.dev/github.com/blitui/blit#ChartPanelOpts) | ChartPanel configuration options |
| [`ToastAction`](https://pkg.go.dev/github.com/blitui/blit#ToastAction) | Toast notification action button |
| [`ToastManagerOpts`](https://pkg.go.dev/github.com/blitui/blit#ToastManagerOpts) | Toast system configuration |
| [`ToastSeverity`](https://pkg.go.dev/github.com/blitui/blit#ToastSeverity) | Toast level (Info, Success, Warning, Error) |
| [`DetailOverlayOpts`](https://pkg.go.dev/github.com/blitui/blit#DetailOverlayOpts) | Detail overlay configuration |
| [`DetailRenderer`](https://pkg.go.dev/github.com/blitui/blit#DetailRenderer) | Custom detail view renderer function |

## Theme & Styling

| Type | Description |
|------|-------------|
| [`Glyphs`](https://pkg.go.dev/github.com/blitui/blit#Glyphs) | Custom cursor/flash/spinner glyphs |
| [`StyleSet`](https://pkg.go.dev/github.com/blitui/blit#StyleSet) | Named style collection on a theme |
| [`BorderSet`](https://pkg.go.dev/github.com/blitui/blit#BorderSet) | Border character set |
| [`Gradient`](https://pkg.go.dev/github.com/blitui/blit#Gradient) | Color gradient for text rendering |
| [`ViewportGlyphs`](https://pkg.go.dev/github.com/blitui/blit#ViewportGlyphs) | Scrollbar track/thumb characters |
| [`ThemeHotReload`](https://pkg.go.dev/github.com/blitui/blit#ThemeHotReload) | File-watching theme reloader |
| [`ThemeHotReloadMsg`](https://pkg.go.dev/github.com/blitui/blit#ThemeHotReloadMsg) | Theme reload success message |
| [`ThemeHotReloadErrMsg`](https://pkg.go.dev/github.com/blitui/blit#ThemeHotReloadErrMsg) | Theme reload error message |

## Animation

| Type | Description |
|------|-------------|
| [`Ease`](https://pkg.go.dev/github.com/blitui/blit#Ease) | Easing function type (`func(float64) float64`) |
| [`Tween`](https://pkg.go.dev/github.com/blitui/blit#Tween) | Time-based animation tween |

## Reactive State

| Type | Description |
|------|-------------|
| [`Signal`](https://pkg.go.dev/github.com/blitui/blit#Signal) | Generic reactive value with subscriber notifications |
| [`AnySignal`](https://pkg.go.dev/github.com/blitui/blit#AnySignal) | Type-erased signal interface |
| [`Unsubscribe`](https://pkg.go.dev/github.com/blitui/blit#Unsubscribe) | Callback to remove a signal subscription |
| [`Config`](https://pkg.go.dev/github.com/blitui/blit#Config) | Generic configuration manager with file persistence |
| [`ConfigOption`](https://pkg.go.dev/github.com/blitui/blit#ConfigOption) | Configuration loading option |

## Self-Update

| Type | Description |
|------|-------------|
| [`UpdateConfig`](https://pkg.go.dev/github.com/blitui/blit#UpdateConfig) | Self-update configuration |
| [`UpdateResult`](https://pkg.go.dev/github.com/blitui/blit#UpdateResult) | Result of checking for updates |
| [`UpdateCache`](https://pkg.go.dev/github.com/blitui/blit#UpdateCache) | Cached update check data |
| [`UpdateMode`](https://pkg.go.dev/github.com/blitui/blit#UpdateMode) | Update behavior mode (Notify, Blocking, Forced, Silent) |
| [`UpdateProgress`](https://pkg.go.dev/github.com/blitui/blit#UpdateProgress) | Download progress display component |
| [`UpdateProgressMsg`](https://pkg.go.dev/github.com/blitui/blit#UpdateProgressMsg) | Progress update message |
| [`Release`](https://pkg.go.dev/github.com/blitui/blit#Release) | GitHub release metadata |
| [`ReleaseAsset`](https://pkg.go.dev/github.com/blitui/blit#ReleaseAsset) | Release binary asset |
| [`Version`](https://pkg.go.dev/github.com/blitui/blit#Version) | Parsed semantic version |
| [`InstallMethod`](https://pkg.go.dev/github.com/blitui/blit#InstallMethod) | Detected installation method (binary, Homebrew, Scoop) |
| [`RateLimitError`](https://pkg.go.dev/github.com/blitui/blit#RateLimitError) | GitHub API rate limit error |

## Data & Polling

| Type | Description |
|------|-------------|
| [`Poller`](https://pkg.go.dev/github.com/blitui/blit#Poller) | Background data fetcher with tick-driven refresh |
| [`PollerOpts`](https://pkg.go.dev/github.com/blitui/blit#PollerOpts) | Poller configuration options |
| [`PollerStats`](https://pkg.go.dev/github.com/blitui/blit#PollerStats) | Poller runtime statistics |
| [`PollerSuccessMsg`](https://pkg.go.dev/github.com/blitui/blit#PollerSuccessMsg) | Successful poll result message |
| [`PollerErrorMsg`](https://pkg.go.dev/github.com/blitui/blit#PollerErrorMsg) | Poll error message |
| [`PollerStartMsg`](https://pkg.go.dev/github.com/blitui/blit#PollerStartMsg) | Poller started message |
| [`PollerRateLimitedMsg`](https://pkg.go.dev/github.com/blitui/blit#PollerRateLimitedMsg) | Rate-limited poll message |
| [`RetryOpts`](https://pkg.go.dev/github.com/blitui/blit#RetryOpts) | Retry command configuration |
| [`RetryErrorMsg`](https://pkg.go.dev/github.com/blitui/blit#RetryErrorMsg) | Retry exhaustion error message |

## Table Helpers

| Type | Description |
|------|-------------|
| [`CellRenderer`](https://pkg.go.dev/github.com/blitui/blit#CellRenderer) | Custom table cell render function |
| [`RowStyler`](https://pkg.go.dev/github.com/blitui/blit#RowStyler) | Custom row background styler |
| [`RowClickHandler`](https://pkg.go.dev/github.com/blitui/blit#RowClickHandler) | Mouse click handler for table rows |
| [`CursorChangeHandler`](https://pkg.go.dev/github.com/blitui/blit#CursorChangeHandler) | Callback for cursor movement |
| [`FilterFunc`](https://pkg.go.dev/github.com/blitui/blit#FilterFunc) | Custom table filter function |
| [`SortFunc`](https://pkg.go.dev/github.com/blitui/blit#SortFunc) | Custom table sort function |
| [`TableRowProviderFunc`](https://pkg.go.dev/github.com/blitui/blit#TableRowProviderFunc) | Function-based virtual table provider |
| [`NodeRenderFunc`](https://pkg.go.dev/github.com/blitui/blit#NodeRenderFunc) | Custom tree node render function |

## Enums

| Type | Description |
|------|-------------|
| [`Alignment`](https://pkg.go.dev/github.com/blitui/blit#Alignment) | Text alignment (Left, Center, Right) |
| [`Orientation`](https://pkg.go.dev/github.com/blitui/blit#Orientation) | Layout orientation (Horizontal, Vertical) |
| [`CursorStyle`](https://pkg.go.dev/github.com/blitui/blit#CursorStyle) | Cursor rendering style |
| [`SelectionMode`](https://pkg.go.dev/github.com/blitui/blit#SelectionMode) | Tree selection mode (None, Single, Multi) |
| [`FlexAlign`](https://pkg.go.dev/github.com/blitui/blit#FlexAlign) | Flex cross-axis alignment |
| [`FlexJustify`](https://pkg.go.dev/github.com/blitui/blit#FlexJustify) | Flex main-axis justification |
| [`ForcedChoice`](https://pkg.go.dev/github.com/blitui/blit#ForcedChoice) | Forced update user choice |

## Messages

| Type | Description |
|------|-------------|
| [`TickMsg`](https://pkg.go.dev/github.com/blitui/blit#TickMsg) | Periodic tick message |
| [`SetThemeMsg`](https://pkg.go.dev/github.com/blitui/blit#SetThemeMsg) | Theme change message |
| [`NotifyMsg`](https://pkg.go.dev/github.com/blitui/blit#NotifyMsg) | User notification message |
| [`LogAppendMsg`](https://pkg.go.dev/github.com/blitui/blit#LogAppendMsg) | LogViewer append message |
| [`FormSubmitMsg`](https://pkg.go.dev/github.com/blitui/blit#FormSubmitMsg) | Form submission message |
| [`CopyToClipboardMsg`](https://pkg.go.dev/github.com/blitui/blit#CopyToClipboardMsg) | Clipboard copy message |

## SSH Serve

| Type | Description |
|------|-------------|
| [`ServeConfig`](https://pkg.go.dev/github.com/blitui/blit#ServeConfig) | SSH server configuration for hosting TUI apps |

## Packages

| Package | Description |
|---------|-------------|
| [`cli`](https://pkg.go.dev/github.com/blitui/blit/cli) | Interactive CLI prompts (Confirm, Select, Input, Spinner, Progress) |
| [`charts`](https://pkg.go.dev/github.com/blitui/blit/charts) | Chart components (Bar, Line, Ring, Gauge, Heatmap) |
| [`btest`](https://pkg.go.dev/github.com/blitui/blit/btest) | Virtual terminal testing framework |
