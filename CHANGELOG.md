# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-04-10

Initial release of blit, migrated from tuikit-go.

### Added
- **Components**: Table (with virtualized scrolling for 1M+ rows), ListView, Tabs, Picker, Tree, Form (with validators and wizard mode), FilePicker, LogViewer, Markdown viewer
- **Layout**: DualPane with collapsible sidebar, HBox/VBox flex layout, Split pane with draggable divider
- **Theme system**: Dark/light themes with semantic color tokens, hot-reload via fsnotify, terminal theme importers
- **Keybinding registry** with auto-generated help overlay
- **Charts**: Bar, line, ring, gauge, heatmap
- **CLI primitives**: Confirm, SelectOne, MultiSelect, Input, Password, Spinner, Progress
- **Self-update**: Binary replacement with SHA256/cosign verification, delta patches, rollback, channels, rate-limit backoff
- **SSH serve**: Host any blit app over SSH via Charm Wish
- **blit CLI**: Vitest-style test runner with watch mode, snapshot diffing, JUnit/HTML reports, VHS tape generation
- **btest**: Virtual terminal testing framework with golden files, session recording/replay
- **Compound components**: Notifications/toasts, overlays, command bar, breadcrumbs, config editor, detail overlay
- **Animation engine** with tween bus and easing functions
- **Dev console** overlay for runtime inspection
- **Signal/slot** reactive state system
- **MkDocs Material** documentation site
- **GoReleaser** with Homebrew tap and Scoop bucket

[Unreleased]: https://github.com/blitui/blit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/blitui/blit/releases/tag/v0.1.0
