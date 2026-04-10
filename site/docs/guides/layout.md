# Layout

## Single Pane

Register one component as the main pane:

```go
blit.WithComponent("main", myComponent)
```

## Dual Pane

Side-by-side layout with a collapsible sidebar:

```go
blit.WithLayout(&blit.DualPane{
    Main:         table,
    Side:         panel,
    SideWidth:    30,        // character columns
    MinMainWidth: 60,        // sidebar auto-hides below this terminal width
    SideRight:    true,      // sidebar on the right (false = left)
    ToggleKey:    "p",       // key to collapse/expand
})
```

`DualPane.Main` maps to `SlotMain`; `DualPane.Side` maps to `SlotSidebar`. Focus cycles between the two panes with `Tab`.

## StatusBar

Attach a footer with left and right text sections:

```go
blit.WithStatusBar(
    func() string { return " ? help  q quit" },
    func() string { return fmt.Sprintf(" %d rows", count) },
)
```

For reactive content driven by signals (e.g. background polling):

```go
leftSig  := blit.NewSignal("")
rightSig := blit.NewSignal("")

blit.WithStatusBarSignal(leftSig, rightSig)

// From any goroutine:
leftSig.Set("connected")
```

Signal updates are coalesced into one notification per frame via a dirty-bit mechanism.

## Tick Interval

Register a periodic tick for polling or animation:

```go
blit.WithTickInterval(100 * time.Millisecond)
```

Components receive `blit.TickMsg` in their `Update` method.
