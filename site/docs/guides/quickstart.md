# Quick Start

## Install

```bash
go get github.com/blitui/blit
```

Requires Go 1.24+.

## Minimal App

The smallest possible blit app registers one component and runs:

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

Run it:

```bash
go run .
```

Keys: `j`/`k` to move, `s` to cycle sort, `/` to search, `?` for help, `q` to quit.

## What `NewApp` wires up

| Option | Effect |
|--------|--------|
| `WithTheme` | Applies semantic color tokens to all components |
| `WithComponent` | Registers a component as the main pane |
| `WithLayout` | Dual-pane layout with sidebar |
| `WithStatusBar` | Footer with left/right text |
| `WithHelp` | `?` toggle overlay — auto-populated from all `KeyBindings()` |
| `WithMouseSupport` | Enables mouse scroll and click |
| `WithAutoUpdate` | Binary self-update on launch |

## Next Steps

- [App Structure](app-structure.md) — component interface, slots, key dispatch
- [Theming](theming.md) — dark/light themes, custom tokens
- [Testing](testing.md) — blit virtual terminal assertions
