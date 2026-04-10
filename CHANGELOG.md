# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.12.0] - 2026-04-09

### Added
- SSH serve via Charm Wish — host any tuikit app over SSH
- Cosign ed25519 signature verification for self-update
- Delta binary patching for smaller update downloads
- MkDocs Material documentation site
- Starter project template with GoReleaser and CI wiring

## [0.11.0] - 2026-04-09

### Added
- tuitest snapshot review UI
- VHS tape integration for automated GIF generation
- Screen diff viewer for visual test comparison
- tuitest CLI with vitest-style reporter, JUnit/HTML output, watch mode

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
- tuitest virtual terminal testing framework

[0.12.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/moneycaringcoder/tuikit-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/moneycaringcoder/tuikit-go/releases/tag/v0.1.0
