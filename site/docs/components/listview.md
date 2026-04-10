# ListView

Generic scrollable list with cursor navigation, optional header, optional detail bar, and flash highlighting. Implements `Component` and `Themed`.

`ListView` is generic over any item type `T`. Use it standalone as a registered component, or embed it inside a custom component and delegate via `HandleKey` and `View`.

## Construction

```go
lv := blit.NewListView[T](opts blit.ListViewOpts[T])
lv.SetItems(items []T)
```

## ListViewOpts

```go
type ListViewOpts[T any] struct {
    // RenderItem renders a single item to a string.
    // Parameters: item, index, whether this item has cursor, theme.
    RenderItem func(item T, idx int, isCursor bool, theme blit.Theme) string

    // HeaderFunc renders a header above the list. Optional.
    // Can be multi-line; ListView accounts for its height automatically.
    HeaderFunc func(theme blit.Theme) string

    // DetailFunc renders a detail bar below the list for the selected item. Optional.
    // Only shown when the list is focused and there is a cursor item.
    DetailFunc func(item T, theme blit.Theme) string

    // FlashFunc returns true if an item should show the flash marker. Optional.
    FlashFunc func(item T, now time.Time) bool

    // OnSelect is called when Enter is pressed on an item. Optional.
    OnSelect func(item T, idx int)

    // DetailHeight is lines to reserve for detail bar (default 3 when DetailFunc set).
    DetailHeight int
}
```

## Basic Usage

```go
type Server struct {
    Name   string
    Status string
}

lv := blit.NewListView[Server](blit.ListViewOpts[Server]{
    RenderItem: func(s Server, idx int, isCursor bool, theme blit.Theme) string {
        color := theme.Positive
        if s.Status != "online" {
            color = theme.Negative
        }
        return fmt.Sprintf("%-20s %s",
            s.Name,
            lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(s.Status),
        )
    },
    OnSelect: func(s Server, idx int) {
        openDetail(s)
    },
})

lv.SetItems(servers)
```

## Header

```go
blit.ListViewOpts[Server]{
    HeaderFunc: func(theme blit.Theme) string {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.Muted)).
            Bold(true).
            Render(fmt.Sprintf("  %-20s %s", "NAME", "STATUS"))
    },
}
```

## Detail Bar

The detail bar is shown below the list when the component is focused and a cursor item exists. Space is always reserved to prevent viewport jitter on focus change:

```go
blit.ListViewOpts[Server]{
    DetailFunc: func(s Server, theme blit.Theme) string {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.Accent)).
            Render(fmt.Sprintf("  IP: %s  Region: %s", s.IP, s.Region))
    },
    DetailHeight: 2,
}
```

## Flash Highlighting

Mark items with a flash glyph based on a time-based predicate:

```go
blit.ListViewOpts[Server]{
    FlashFunc: func(s Server, now time.Time) bool {
        return now.Sub(s.LastEvent) < 2*time.Second
    },
}
```

Call `lv.Refresh()` from a tick handler to re-evaluate flash state each frame.

## Cursor Rendering

The cursor row is automatically styled with the theme's `Cursor` background. The cursor marker glyph (default `▶`) is drawn with a 120ms ease-out tween animation on movement. Use `isCursor bool` in `RenderItem` to suppress per-item highlights that would conflict:

```go
RenderItem: func(s Server, idx int, isCursor bool, theme blit.Theme) string {
    // Don't apply extra background — RowStyler handles it
    name := s.Name
    if !isCursor {
        name = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text)).Render(name)
    }
    return name
},
```

## State Accessors

```go
lv.CursorItem()    // *T at cursor, or nil
lv.CursorIndex()   // Current cursor position
lv.Items()         // All items
lv.ItemCount()     // Number of items
lv.IsAtTop()       // Viewport scrolled to top
lv.IsAtBottom()    // Viewport scrolled to bottom
```

## Navigation Methods

```go
lv.SetCursor(idx int)   // Move cursor to index
lv.ScrollToTop()        // Jump to first item
lv.ScrollToBottom()     // Jump to last item
lv.Refresh()            // Re-render content without changing items
```

## Embedding

When embedding `ListView` inside a custom component, delegate key handling explicitly:

```go
func (c *MyComponent) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
    if key, ok := msg.(tea.KeyMsg); ok {
        if cmd := c.list.HandleKey(key); cmd != nil {
            return c, cmd
        }
    }
    return c, nil
}
```

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Move cursor up |
| `down` / `j` | Move cursor down |
| `home` / `g` | Jump to top |
| `end` / `G` | Jump to bottom |
| `enter` | Select item (fires `OnSelect`) |
