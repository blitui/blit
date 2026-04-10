# App Structure

## The Component Interface

Every blit widget implements `Component`:

```go
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg, ctx Context) (Component, tea.Cmd)
    View() string
    KeyBindings() []KeyBind
    SetSize(w, h int)
    Focused() bool
    SetFocused(bool)
}
```

Components that support theming also implement `Themed`:

```go
type Themed interface {
    SetTheme(Theme)
}
```

The App calls `SetTheme` on all registered components automatically when the theme changes.

## Key Dispatch Order

When a key arrives the App dispatches it in this order:

1. Overlay stack (e.g. Help, ConfigEditor, Picker)
2. Built-in globals: `q` quit, `tab` cycle focus, `?` help
3. Pane toggle key (DualPane)
4. Overlay trigger keys registered with `WithOverlay`
5. App-level keybindings registered with `WithKeyBind`
6. Focused component's `Update`

Return `blit.Consumed()` from `Update` to stop further dispatch.

## Slots

The App organises components into named slots:

| Slot | Description |
|------|-------------|
| `SlotMain` | Primary content pane |
| `SlotSidebar` | Side panel (DualPane only) |
| `SlotFooter` | StatusBar |

`WithComponent` fills `SlotMain`. `WithLayout` maps `DualPane.Main` → `SlotMain` and `DualPane.Side` → `SlotSidebar`.

## Registering Keybindings

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

All registered bindings appear in the auto-generated Help overlay.

## Pushing Data from Goroutines

```go
app := blit.NewApp(...)
go func() {
    for data := range stream {
        app.Send(MyDataMsg{data})
    }
}()
app.Run()
```

Unknown message types are forwarded to all components via `Update`.
