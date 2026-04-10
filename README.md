# tuikit-go

The pragmatic TUI toolkit for shipping CLI tools fast. Wraps [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) with reusable components, a layout engine, a keybinding registry, a theme system, and built-in binary self-update.

![tuikit-go quick start](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/quickstart.gif)

## Features

- **Table** with sorting, filtering, custom cell rendering, mouse support, and virtualized scrolling (1M+ rows)
- **ListView, Tabs, Picker, Tree, Form, LogViewer** and more out of the box
- **Dual-pane layout** with collapsible sidebar, flex layout, split panes
- **Keybinding registry** with auto-generated help screen
- **Dark/light themes** with semantic color tokens, hot-reload, and terminal theme importers
- **CLI primitives** (confirm, select, input, spinner, progress) for non-TUI workflows
- **tuitest** virtual terminal testing framework with golden files, snapshot diffing, and a vitest-style CLI runner
- **Charts** — bar, line, ring, gauge, heatmap, sparkline
- **Self-update** — binary replacement with SHA256/cosign verification, delta patches, rollback, channels, and rate-limit backoff
- **SSH serve** — host any tuikit app over SSH via Charm Wish
- **Notifications, overlays, command bar, breadcrumbs** and other compound components

## Install

```bash
go get github.com/moneycaringcoder/tuikit-go
```

**tuitest CLI** (optional test runner):

```bash
# Homebrew
brew install moneycaringcoder/tap/tuitest

# Scoop
scoop bucket add moneycaringcoder https://github.com/moneycaringcoder/scoop-bucket
scoop install tuitest

# Go
go install github.com/moneycaringcoder/tuikit-go/cmd/tuitest@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    tuikit "github.com/moneycaringcoder/tuikit-go"
)

func main() {
    table := tuikit.NewTable(
        []tuikit.Column{
            {Title: "Name", Width: 20, Sortable: true},
            {Title: "Status", Width: 15},
        },
        []tuikit.Row{
            {"Alice", "Online"},
            {"Bob", "Away"},
        },
        tuikit.TableOpts{Sortable: true, Filterable: true},
    )

    app := tuikit.NewApp(
        tuikit.WithTheme(tuikit.DefaultTheme()),
        tuikit.WithComponent("main", table),
        tuikit.WithStatusBar(
            func() string { return " ? help  q quit" },
            func() string { return fmt.Sprintf(" %d items", 2) },
        ),
        tuikit.WithHelp(),
    )

    app.Run()
}
```

More examples in [`examples/`](examples/).

## Documentation

- **[Docs site](https://moneycaringcoder.github.io/tuikit-go/)** — guides, component reference, theming, self-update setup
- **[Examples](examples/)** — 16 runnable demos from minimal to full dashboard
- **[pkg.go.dev](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go)** — API reference

## Repository Layout

| Directory | Purpose |
|-----------|---------|
| `charts/` | Chart components (bar, line, ring, gauge, heatmap) |
| `cli/` | Interactive CLI prompt primitives (non-TUI) |
| `cmd/` | CLI binaries (`tuitest` runner) |
| `docs/` | Design docs and generated GIFs |
| `examples/` | Runnable example apps |
| `internal/` | Private packages (fuzzy search, scaffold, tape) |
| `scripts/` | GIF generation and VHS tape scripts |
| `site/` | MkDocs Material documentation site |
| `templates/` | Starter project template |
| `testdata/` | Test fixtures (theme files) |
| `tuitest/` | Virtual terminal testing framework |
| `updatetest/` | Self-updater test mocks |

## Used By

- [gitstream-tui](https://github.com/moneycaringcoder/gitstream-tui) — GitHub activity dashboard
- [cryptstream-tui](https://github.com/moneycaringcoder/cryptstream-tui) — Live cryptocurrency ticker

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
