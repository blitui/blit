# blit

The pragmatic TUI toolkit for shipping Go CLI tools fast. Wraps [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) with reusable components, a layout engine, a keybinding registry, a theme system, and built-in binary self-update.

```bash
go get github.com/blitui/blit
```

## Features

- **Table** — adaptive columns, sorting, filtering, custom cell rendering, detail bars, virtual mode for millions of rows
- **ListView** — generic scrollable list with cursor, header, detail bar, and flash highlighting
- **Tabs** — horizontal or vertical tab switcher with click-to-focus content panes
- **Picker** — fuzzy-searchable command palette (Cmd-K style) with optional preview pane
- **Form** — multi-field validated form with keyboard navigation and wizard mode
- **Tree** — collapsible tree view for hierarchical data
- **LogViewer** — append-only auto-scrolling log with level filtering and substring search
- **Charts** — bar, line, ring, gauge, heatmap, and sparkline
- **StatusBar** — left/right footer driven by closures or reactive signals
- **Help** — auto-generated keybinding reference from the registry; press `?` to toggle
- **Dual-pane layout** with collapsible sidebar, flex layout (HBox/VBox), and split panes
- **Dark + light themes** with semantic color tokens, hot-reload, and extensible `Extra` color map
- **CLI primitives** — Confirm, SelectOne, MultiSelect, Input, Password, Spinner, Progress
- **blit** — virtual-terminal testing framework with 30+ assertions, golden files, and a vitest-style reporter
- **Self-update** — binary replacement with SHA256/cosign verification, delta patches, rollback, and Homebrew/Scoop detection
- **SSH serve** — host any blit app over SSH via Charm Wish
- **Notifications, overlays, command bar, breadcrumbs** and other compound components
- **Keybinding registry** with conflict detection and auto-generated help screen

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
            {Title: "Name",   Width: 20, Sortable: true},
            {Title: "Status", Width: 15},
        },
        []blit.Row{
            {"Alice", "Online"},
            {"Bob",   "Away"},
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

See [Getting Started](guides/quickstart.md) to build your first app in under 5 minutes.

## Examples

```bash
go run ./examples/minimal/     # Simple ListView in ~30 lines
go run ./examples/dashboard/   # Table + DualPane + ConfigEditor
go run ./examples/monitor/     # Service fleet dashboard
go run ./examples/cli-demo/    # Interactive CLI primitives showcase
```

## License

MIT — [github.com/blitui/blit](https://github.com/blitui/blit)
