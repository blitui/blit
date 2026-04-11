# Menu

Popup overlay with separator support, disabled items, and shortcut hints. Implements `Component`, `Themed`, and `Overlay`.

## Construction

```go
func NewMenu(opts MenuOpts) *Menu
```

## Types

```go
type MenuItem struct {
    Label     string // Display text
    Shortcut  string // Optional key hint, right-aligned (e.g., "ctrl+s")
    Action    func() // Called when activated (optional)
    Disabled  bool   // Greyed out and non-interactive
    Separator bool   // Renders as horizontal divider (ignores other fields)
}

type MenuOpts struct {
    Items    []MenuItem // Menu entries
    OnClose  func()     // Called when dismissed
    MinWidth int        // Minimum menu width (0 = auto)
}
```

## Usage

```go
menu := blit.NewMenu(blit.MenuOpts{
    Items: []blit.MenuItem{
        {Label: "New File", Shortcut: "ctrl+n", Action: func() { /* create */ }},
        {Label: "Open...", Shortcut: "ctrl+o", Action: func() { /* open */ }},
        {Separator: true},
        {Label: "Save", Shortcut: "ctrl+s", Action: func() { /* save */ }},
        {Label: "Export PDF", Disabled: true},
    },
})
```

## Separators and Disabled Items

Set `Separator: true` to render a horizontal divider. Navigation automatically skips separators and disabled items.

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Move up |
| `down` / `j` | Move down |
| `enter` / `space` | Select item |
| `esc` | Close menu |

## State

| Method | Description |
|--------|-------------|
| `IsActive()` | Whether the menu is showing |
| `Close()` | Dismiss menu and call OnClose |
| `CursorIndex()` | Highlighted item index |
