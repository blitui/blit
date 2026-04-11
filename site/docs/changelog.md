# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

Full release history is also available on
[GitHub Releases](https://github.com/blitui/blit/releases).

## [0.2.21] - 2026-04-10

### Added
- Final docs update for v0.2.20, roadmap and changelog current through stability phase

## [0.2.20] - 2026-04-10

### Added
- Animation/interpolation edge case tests (nil easing, t clamping, unsupported types, invalid colors, parseHexColor edge cases)

## [0.2.19] - 2026-04-10

### Added
- Changelog, roadmap, and API reference documentation updates for v0.2.15–v0.2.18

## [0.2.18] - 2026-04-10

### Added
- Charts Component interface tests (Init, Update, KeyBindings, Focused, SetFocused, SetTheme) for all 5 chart types
- Charts edge case tests (zero max, custom colors, flat data, empty series)
- Charts package coverage improved from 86.1% to 93.4%

## [0.2.17] - 2026-04-10

### Added
- FilePicker enhanced test coverage (hidden files, search navigation, preview pane, zero size, unfocused state, init, default root, directory collapse)

## [0.2.16] - 2026-04-10

### Added
- Theme importer tests using inline data (Gogh, Alacritty, iTerm2) covering partial fields, 0x prefix, cursor section, comments, invalid XML, empty plist
- `clampF` edge case tests

## [0.2.15] - 2026-04-10

### Added
- Changelog, roadmap, and API reference documentation updates for v0.2.0–v0.2.14

## [0.2.14] - 2026-04-10

### Added
- Form field unit tests covering all field types (TextField, PasswordField, SelectField, MultiSelectField, ConfirmField, NumberField)

## [0.2.13] - 2026-04-10

### Added
- Glyphs test coverage (field validation, ASCII-only constraint, spinner frames)
- Theme presets test coverage (registry, required fields, extra colors for all 8 presets)

## [0.2.12] - 2026-04-10

### Added
- `Stepper` component — multi-step progress indicator with done/active/pending states, forward/back navigation, OnChange and OnComplete callbacks

## [0.2.11] - 2026-04-10

### Added
- `Tooltip` component — floating overlay with anchor positioning, esc dismiss, compositing via `FloatingOverlay` interface

## [0.2.10] - 2026-04-10

### Added
- `Spinner` component — animated loading indicator cycling through glyph SpinnerFrames with configurable interval

## [0.2.9] - 2026-04-10

### Added
- `Accordion` component — collapsible sections with exclusive mode, toggle callbacks, vi navigation

## [0.2.8] - 2026-04-10

### Added
- `Breadcrumb` component — navigable path display with Push/Pop API, custom separators, OnSelect callback

## [0.2.7] - 2026-04-10

### Added
- `Timeline` component — vertical and horizontal event sequences with status icons (done/active/pending), scrolling, OnSelect

## [0.2.6] - 2026-04-10

### Added
- `Kanban` component — multi-column board with card movement between columns, tag badges, OnMove and OnSelect callbacks

## [0.2.5] - 2026-04-10

### Added
- `ChartPanel` component — switchable container for chart Components with tab navigation and label header

## [0.2.4] - 2026-04-10

### Added
- `Menu` overlay component — popup menu with separator support, disabled items, shortcut hints

## [0.2.3] - 2026-04-10

### Added
- `Dialog` overlay component — modal dialog with button navigation, centered rendering, OnClose callback

## [0.2.2] - 2026-04-10

### Added
- `ProgressBar` component — value-based progress with label, percentage display, themed bar glyphs

## [0.2.1] - 2026-04-10

### Added
- TreeView lazy loading via `LoadChildren` callback
- TreeView search/filter with fuzzy matching
- TreeView context menu support via `OnContext` callback

## [0.2.0] - 2026-04-10

### Added
- TreeView selection modes (None, Single, Multi) with `SelectedNodes()`, `SelectAll()`, `DeselectAll()`
- TreeView custom node rendering via `NodeRenderFunc`
- TreeView `Detail` field on Node for additional info display

## [0.12.5] - 2026-04-10

### Fixed
- Resolved all golangci-lint warnings (errcheck, singleCaseSwitch, ifElseChain, unused code)
- Corrected Go version references across docs (1.21/1.26 → 1.24)
- Fixed missing imports in cookbook recipes
- Fixed broken hero GIF link and example count in README
- Updated deprecated viewport API calls in LogViewer

### Changed
- Exported `Registry` type (was unexported `registry`) for `Context.Hotkeys` field
- Rewrote README as concise landing page (570 → ~150 lines)
- Deleted ~70 lines of dead code (unused types, functions, variables)

### Added
- Package-level godoc (`doc.go`) for pkg.go.dev
- Component documentation pages (Form, Tree, Split, Charts, FilePicker)
- MIT license, CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md
- GitHub issue/PR templates
- golangci-lint CI workflow with coverage reporting
- Expanded API reference with Registry, FilePicker, Breadcrumbs, Viewport

## [0.12.0] - 2026-04-09

### Added
- SSH serve via Charm Wish — host any blit app over SSH
- Cosign ed25519 signature verification for self-update
- Delta binary patching for smaller update downloads
- MkDocs Material documentation site
- Starter project template with GoReleaser and CI wiring

## [0.11.0] - 2026-04-09

### Added
- blit snapshot review UI
- VHS tape integration for automated GIF generation
- Screen diff viewer for visual test comparison
- blit CLI with vitest-style reporter, JUnit/HTML output, watch mode

## [0.10.0] - 2026-04-09

### Added
- `Context` struct threaded through `Component.Update` (Theme, Size, Focus, Hotkeys, Clock, Logger)
- Dev console overlay
- Theme hot-reload via fsnotify

### Changed
- **Breaking:** `Component.Update` signature now takes `Context` parameter

## [0.9.0] - 2026-04-09

### Added
- Tree component with expand/collapse
- FilePicker component
- LogViewer with streaming and level filtering
- Virtualized Table with `TableRowProvider` for 1M+ rows
- HBox/VBox flex layout
- Breadcrumbs component
- Split pane with draggable divider

## [0.8.0] - 2026-04-09

### Added
- Dark/light theme system with semantic color tokens and `Extra` map
- Animation engine with tween bus and easing functions
- Form component with validators and wizard mode
- Tabs component with horizontal/vertical orientation
- Picker with fzf-style fuzzy search
- Toast notifications with severity levels
- Gradient text rendering
- VHS tape scripts for README GIFs

## [0.7.0] - 2026-04-09

### Added
- Markdown rendering via glamour
- Collapsible sections
- Detail overlay for row inspection

## [0.6.0] - 2026-04-09

### Added
- Self-update system with SHA256 checksum verification
- Skip-version, forced update, and notify modes
- Rollback on verify failure
- Rate-limit backoff for GitHub API
- Homebrew and Scoop install detection

## [0.5.0] - 2026-04-08

### Added
- CLI primitives package (Confirm, SelectOne, MultiSelect, Input, Password, Spinner, Progress)
- Styled message helpers (Success, Warning, Error, Info, Title, Step)
- ConfigEditor overlay
- CommandBar with completion
- Update progress overlay

## [0.4.0] - 2026-04-08

### Added
- Poller for background data with tick-driven refresh
- Mouse support for Table scroll and click

## [0.3.0] - 2026-04-08

### Added
- Dual-pane layout with collapsible sidebar
- Named overlay system with trigger keys

## [0.2.0] - 2026-04-08

### Added
- StatusBar with left/right content
- Help overlay auto-generated from keybindings
- Keybinding registry

## [0.1.0] - 2026-04-08

### Added
- Initial release
- Table component with sorting, filtering, and cursor navigation
- ListView component
- App framework with functional options
- blit virtual terminal testing framework

[0.12.0]: https://github.com/blitui/blit/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/blitui/blit/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/blitui/blit/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/blitui/blit/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/blitui/blit/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/blitui/blit/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/blitui/blit/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/blitui/blit/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/blitui/blit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/blitui/blit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/blitui/blit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/blitui/blit/releases/tag/v0.1.0
