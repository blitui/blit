# blit

[![CI](https://github.com/blitui/blit/actions/workflows/ci.yml/badge.svg)](https://github.com/blitui/blit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/blitui/blit.svg)](https://pkg.go.dev/github.com/blitui/blit)
[![Go Report Card](https://goreportcard.com/badge/github.com/blitui/blit)](https://goreportcard.com/report/github.com/blitui/blit)
[![Latest Release](https://img.shields.io/github/v/release/blitui/blit)](https://github.com/blitui/blit/releases/latest)

The pragmatic TUI toolkit for shipping CLI tools fast. Wraps [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) with reusable components, a layout engine, a keybinding registry, a theme system, and built-in binary self-update.

## Features

- **30+ components** — Table (virtualized, 1M+ rows), Tree, Form, Tabs, Picker, FilePicker, LogViewer, ListView, and more
- **Rich overlays** — Dialog, Menu, Tooltip, Toast notifications, command bar
- **Interactive widgets** — Accordion, Stepper, Spinner, ProgressBar, Timeline, Kanban, Breadcrumb
- **Layout system** — dual-pane, HBox/VBox flex, split panes with draggable dividers
- **Charts** — bar, line, ring, gauge, heatmap
- **11 theme presets** — Dracula, Catppuccin Mocha, Tokyo Night, Nord, Gruvbox Dark, Rose Pine, Kanagawa, One Dark, Solarized Dark, Everforest, Nightfox
- **Theme system** — semantic color tokens, hot-reload, terminal theme importers (iTerm2, Alacritty, Gogh)
- **Keybinding registry** with auto-generated help screen
- **CLI primitives** (confirm, select, input, spinner, progress) for non-TUI workflows
- **btest** virtual terminal testing framework with golden files, snapshot diffing, and a vitest-style CLI runner
- **Self-update** — binary replacement with SHA256/cosign verification, delta patches, rollback, channels, and rate-limit backoff
- **SSH serve** — host any blit app over SSH via Charm Wish

## Install

```bash
go get github.com/blitui/blit
```

**blit CLI** (optional test runner):

```bash
# Homebrew
brew install blitui/tap/blit

# Scoop
scoop bucket add blitui https://github.com/blitui/scoop-bucket
scoop install blit

# Go
go install github.com/blitui/blit/cmd/blit@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    blit "github.com/blitui/blit"
)

func main() {
    table := blit.NewTable(
        []blit.Column{
            {Title: "Name", Width: 20, Sortable: true},
            {Title: "Status", Width: 15},
        },
        []blit.Row{
            {"Alice", "Online"},
            {"Bob", "Away"},
        },
        blit.TableOpts{Sortable: true, Filterable: true},
    )

    app := blit.NewApp(
        blit.WithTheme(blit.DefaultTheme()),
        blit.WithComponent("main", table),
        blit.WithStatusBar(
            func() string { return " ? help  q quit" },
            func() string { return fmt.Sprintf(" %d items", 2) },
        ),
        blit.WithHelp(),
    )

    app.Run()
}
```

More examples in [`examples/`](examples/).

## Documentation

- **[Docs site](https://blitui.github.io/blit/)** — guides, component reference, theming, self-update setup
- **[Examples](examples/)** — 15 runnable demos from minimal to full dashboard
- **[pkg.go.dev](https://pkg.go.dev/github.com/blitui/blit)** — API reference

## Repository Layout

| Directory | Purpose |
|-----------|---------|
| `charts/` | Chart components (bar, line, ring, gauge, heatmap) |
| `cli/` | Interactive CLI prompt primitives (non-TUI) |
| `cmd/` | CLI binaries (`blit` runner) |
| `docs/` | Design docs and generated GIFs |
| `examples/` | Runnable example apps |
| `internal/` | Private packages (fuzzy search, scaffold, tape) |
| `scripts/` | GIF generation and VHS tape scripts |
| `site/` | MkDocs Material documentation site |
| `templates/` | Starter project template |
| `testdata/` | Test fixtures (theme files) |
| `blit/` | Virtual terminal testing framework |
| `updatetest/` | Self-updater test mocks |

## Used By

- [gitstream-tui](https://github.com/moneycaringcoder/gitstream-tui) — GitHub activity dashboard
- [cryptstream-tui](https://github.com/moneycaringcoder/cryptstream-tui) — Live cryptocurrency ticker

## Compatibility

blit follows [semantic versioning](https://semver.org/). Within a major version, the public API is stable — no breaking changes in minor or patch releases. Pre-v1.0 releases (v0.x) may include breaking changes in minor versions, documented in the [changelog](CHANGELOG.md).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
