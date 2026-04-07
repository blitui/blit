# tuikit-go

A Go toolkit for building terminal UIs fast. Wraps [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) with reusable components, a layout engine, keybinding registry, and theme system. Build a complete TUI app in under 20 lines.

## Install

```bash
go get github.com/moneycaringcoder/tuikit-go
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

## Full Example

See [`examples/dashboard/`](examples/dashboard/) for a complete app (Galactic Pizza Tracker) showing all components working together — table, sidebar, config editor, help screen, status bar.

```bash
go run ./examples/dashboard/
```

## Components

### Table

Adaptive table with responsive columns, sorting, filtering, and cursor navigation.

```go
columns := []tuikit.Column{
    {Title: "Name", Width: 20, Sortable: true},
    {Title: "Score", Width: 10, Align: tuikit.Right, Sortable: true},
    {Title: "Extra", Width: 15, MinWidth: 100}, // hides below 100 cols
}

table := tuikit.NewTable(columns, rows, tuikit.TableOpts{
    Sortable:   true,  // 's' to cycle sort
    Filterable: true,  // '/' to search
})

table.SetRows(newRows) // update data dynamically
```

#### Custom Cell Rendering

Full control over per-cell styling (colors, icons, conditional formatting):

```go
tuikit.TableOpts{
    CellRenderer: func(row tuikit.Row, colIdx int, isCursor bool, theme tuikit.Theme) string {
        val := row[colIdx]
        if colIdx == 2 && val == "Online" {
            return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive)).Render(val)
        }
        return val
    },
}
```

#### Custom Sort

Numeric, time-based, or any custom sort logic:

```go
tuikit.TableOpts{
    SortFunc: func(a, b tuikit.Row, sortCol int, sortAsc bool) bool {
        va, _ := strconv.ParseFloat(a[sortCol], 64)
        vb, _ := strconv.ParseFloat(b[sortCol], 64)
        if sortAsc { return va < vb }
        return va > vb
    },
}
```

#### Predicate Filter

Filter rows programmatically alongside text search:

```go
table.SetFilter(func(row tuikit.Row) bool {
    return row[1] == "online" // only show online users
})
table.SetFilter(nil) // clear filter
```

#### Mouse Support

Scroll wheel and click are handled automatically when mouse is enabled:

```go
tuikit.WithMouseSupport()
```

### Status Bar

Footer with left-aligned hints and right-aligned status.

```go
tuikit.WithStatusBar(
    func() string { return " ? help  q quit" },
    func() string { return " 42 items" },
)
```

### Help Screen

Auto-generated from all registered keybindings. Zero configuration.

```go
tuikit.WithHelp() // press '?' to toggle
```

### Config Editor

Declarative settings overlay with grouped fields and validation.

```go
editor := tuikit.NewConfigEditor([]tuikit.ConfigField{
    {
        Label: "Refresh Interval",
        Group: "General",
        Hint:  "seconds, min 5",
        Get:   func() string { return fmt.Sprint(cfg.Interval) },
        Set:   func(v string) error {
            n, _ := strconv.Atoi(v)
            if n < 5 { return fmt.Errorf("must be >= 5") }
            cfg.Interval = n
            return nil
        },
    },
})

// Register as overlay with trigger key
tuikit.WithOverlay("Settings", "c", editor) // press 'c' to open
```

### Layout

Single pane or dual pane with collapsible sidebar.

```go
tuikit.WithLayout(&tuikit.DualPane{
    Main:         table,
    Side:         panel,
    SideWidth:    30,
    MinMainWidth: 60,  // sidebar auto-hides below this
    SideRight:    true,
    ToggleKey:    "p",
})
```

## Theming

Built-in dark and light themes, or create your own from a color map.

```go
// Built-in
tuikit.DefaultTheme()
tuikit.LightTheme()

// From config (YAML/JSON/TOML — you parse, we color)
tuikit.ThemeFromMap(map[string]string{
    "positive": "#00ff00",
    "negative": "#ff0000",
    "accent":   "#0000ff",
})
```

Semantic tokens: `Positive`, `Negative`, `Accent`, `Muted`, `Text`, `TextInverse`, `Cursor`, `Border`, `Flash`.

## Building Custom Components

Implement the `Component` interface to create your own:

```go
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
    KeyBindings() []tuikit.KeyBind
    SetSize(width, height int)
    Focused() bool
    SetFocused(bool)
}
```

Return `tuikit.Consumed()` from `Update` to signal the App that your component handled a key. The App stops dispatching to other components.

To receive theme updates, implement `Themed`:

```go
type Themed interface {
    SetTheme(tuikit.Theme)
}
```

The App calls `SetTheme` on any component or overlay that implements this interface whenever the theme is set.

For modal overlays, also implement `Overlay`:

```go
type Overlay interface {
    Component
    IsActive() bool
    Close()
}
```

Register overlays with a trigger key:

```go
tuikit.WithOverlay("Help", "?", helpOverlay)
```

## App-Level Keybindings

Register global key handlers that run outside any component:

```go
tuikit.WithKeyBind(tuikit.KeyBind{
    Key:   "f",
    Label: "Cycle filter",
    Group: "DATA",
    Handler: func() {
        filterIdx = (filterIdx + 1) % len(modes)
        table.SetRows(rows) // re-apply filter
    },
})
```

These appear in the help screen automatically.

## Tick / Timer Support

Enable periodic ticks for animations, flash effects, and polling:

```go
tuikit.WithTickInterval(100 * time.Millisecond)
```

Components receive `tuikit.TickMsg` in their `Update` method.

## External Data (Background Goroutines)

Push data into the app from WebSocket streams, API polling, or any goroutine:

```go
app := tuikit.NewApp(...)
go func() {
    for data := range stream {
        app.Send(MyDataMsg{data})
    }
}()
app.Run()
```

Unknown message types are forwarded to all components via `Update`.

## For AI Agents

tuikit follows predictable patterns:

- All components implement `tuikit.Component`
- Use `tuikit.NewApp()` with functional options (`WithTheme`, `WithComponent`, `WithLayout`, etc.)
- Key dispatch: overlay stack → built-in globals (q/tab/?) → pane toggle → overlay triggers → app keybindings → focused component
- See `examples/dashboard/main.go` for a complete reference

## Dependencies

Charm ecosystem only:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — Component primitives
