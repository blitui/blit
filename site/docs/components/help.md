# Help

An overlay that auto-generates a keybinding reference from every registered component and app-level binding. Zero configuration — just add it to your app and press `?`.

Implements `Component`, `Themed`, and the `Overlay` interface.

## Construction

```go
blit.WithHelp()
```

That is the entire API. The `App` wires the `Help` instance to the global `?` key and populates it from the keybinding registry automatically.

## How It Works

Every `Component` exposes its bindings via `KeyBindings() []KeyBind`. When the `App` initialises, it collects bindings from:

1. All registered components (via `WithComponent` or `WithLayout`)
2. App-level bindings added with `WithKeyBind`
3. Built-in globals (`q`, `tab`, `?`)

These are grouped by the `Group` field of each `KeyBind` and displayed in the Help overlay.

## KeyBind

```go
type KeyBind struct {
    Key     string // e.g. "up/k", "ctrl+d", "1-9"
    Label   string // Human-readable description
    Group   string // Section heading in the Help overlay
    Handler func() // Called by the App on keypress (app-level bindings only)
}
```

## Registering App-Level Bindings

```go
blit.WithKeyBind(blit.KeyBind{
    Key:   "r",
    Label: "Refresh",
    Group: "DATA",
    Handler: func() {
        table.SetRows(fetchRows())
    },
})
```

The binding appears in the Help overlay under the `DATA` section heading alongside component bindings.

## Markdown Group Headings

Group names prefixed with `md:` are rendered as inline Markdown. Use this for rich section headings with bold, code spans, or links:

```go
blit.KeyBind{
    Key:   "ctrl+r",
    Label: "Reload config",
    Group: "md:**Configuration** — runtime settings",
}
```

## Closing the Overlay

The Help overlay closes on `?`, `Esc`, or `q`.

## Example Output

```
Keyboard Shortcuts

NAVIGATION
  up/k          Move cursor up
  down/j        Move cursor down
  ctrl+u        Half page up
  ctrl+d        Half page down

DATA
  s             Cycle sort
  /             Search
  r             Refresh

OTHER
  esc           Close help
```
