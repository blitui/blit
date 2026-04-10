# Toasts

Timed notification panels that slide in from the right edge of the screen. Toasts are managed by the App's built-in `ToastManager` — no component registration needed.

## Sending a Toast

Fire a toast from anywhere in the update loop using `ToastCmd`:

```go
blit.ToastCmd(
    blit.SeverityInfo,
    "Deployment started",
    "Pushing to production…",
    4*time.Second,
)
```

Return it as a `tea.Cmd` from your component's `Update`:

```go
func (m MyModel) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
    switch msg.(type) {
    case deployStartedMsg:
        return m, blit.ToastCmd(blit.SeverityInfo, "Deploy", "Pushing…", 4*time.Second)
    }
    return m, nil
}
```

## Severities

| Constant | Icon | Color |
|----------|------|-------|
| `SeverityInfo` | `i` | Accent |
| `SeveritySuccess` | `✓` | Positive |
| `SeverityWarn` | `⚠` | Flash |
| `SeverityError` | `✗` | Negative |

## Toast with Action Buttons

Attach labelled action buttons. Handlers are called when the user clicks them:

```go
blit.ToastCmd(
    blit.SeverityWarn,
    "Update available",
    "v2.1.0 is ready",
    0, // 0 = stays until dismissed
    blit.ToastAction{Label: "Update now", Handler: func() { runUpdate() }},
    blit.ToastAction{Label: "Dismiss",    Handler: func() {}},
)
```

## ToastMsg

You can also build and send the message struct directly:

```go
func() tea.Msg {
    return blit.ToastMsg{
        Severity: blit.SeverityError,
        Title:    "Connection lost",
        Body:     "Reconnecting in 5s…",
        Duration: 5 * time.Second,
    }
}
```

## Simple Notify

For a minimal text banner without severity styling, use `NotifyCmd`:

```go
blit.NotifyCmd("Saved", 2*time.Second)
```

## App Configuration

The ToastManager is enabled by default. Configure it via `WithToastManager`:

```go
blit.WithToastManager(blit.ToastManagerOpts{
    MaxVisible:   3,                    // max simultaneous toasts (default 5)
    AnimDuration: 200*time.Millisecond, // slide-in animation duration
})
```

Set `BLIT_NO_ANIM=1` to disable all toast animations.
