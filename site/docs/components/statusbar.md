# StatusBar

A single-line footer with left-aligned hints and right-aligned status text. Implements `Component` and `Themed`.

The left and right content sources can be plain `func() string` closures (legacy) or `*Signal[string]` for reactive per-frame updates.

## Construction

```go
sb := blit.NewStatusBar(blit.StatusBarOpts{
    Left:  func() string { return " ? help  q quit" },
    Right: func() string { return fmt.Sprintf(" %d items", count) },
})
```

Or via the App option (most common):

```go
blit.WithStatusBar(
    func() string { return " ? help  q quit" },
    func() string { return fmt.Sprintf(" %d items", count) },
)
```

## Reactive Signals

For content that changes asynchronously (background polling, WebSocket streams), use `*Signal[string]`. Signal updates are coalesced into one notification per frame via dirty-bit logic:

```go
leftSig  := blit.NewSignal(" connected")
rightSig := blit.NewSignal(" 0 items")

blit.WithStatusBarSignal(leftSig, rightSig)

// From any goroutine:
rightSig.Set(fmt.Sprintf(" %d items", count))
```

## Layout

The StatusBar renders the left string flush-left and the right string flush-right within the full terminal width. If both strings together exceed the available width, the left string is truncated. The bar is always exactly one line tall.

## Styling

The bar inherits its foreground color from `theme.Muted`. To produce accent or color-coded segments, embed lipgloss-styled strings in the closure:

```go
blit.WithStatusBar(
    func() string {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.Accent)).
            Render(" ● connected")
    },
    func() string { return " q quit" },
)
```
