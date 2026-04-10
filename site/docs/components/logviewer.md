# LogViewer

Append-only, auto-scrolling log display with level filtering, substring search, pause/resume, and goroutine-safe `Append`. Implements `Component` and `Themed`.

## Construction

```go
lv := blit.NewLogViewer()
```

## LogLine

```go
type LogLine struct {
    Level     blit.LogLevel // LogDebug, LogInfo, LogWarn, LogError
    Timestamp time.Time
    Message   string
    Source    string // Optional — shown in accent color between chip and message
}
```

## Log Levels

| Constant | Chip | Color |
|----------|------|-------|
| `LogDebug` | `DBG` | Muted |
| `LogInfo` | `INF` | Accent |
| `LogWarn` | `WRN` | Flash |
| `LogError` | `ERR` | Negative |

## Appending Lines

`Append` is safe to call from any goroutine:

```go
lv.Append(blit.LogLine{
    Level:     blit.LogInfo,
    Timestamp: time.Now(),
    Source:    "api",
    Message:   "request completed in 42ms",
})
```

### Via Bubble Tea Command

To append from within a Bubble Tea command pipeline:

```go
func fetchData() tea.Cmd {
    return blit.LogAppendCmd(blit.LogLine{
        Level:   blit.LogInfo,
        Message: "data fetched",
    })
}
```

The `LogAppendMsg` message type is handled automatically by `LogViewer.Update`.

## Auto-Scroll

The LogViewer auto-scrolls to the bottom as new lines arrive. Scrolling up pauses auto-scroll; pressing `end` or `p` resumes it. The status bar shows `▶ LIVE` or `⏸ PAUSED`.

## Level Filtering

Press `l` to cycle the minimum log level:

```
debug+ → info+ → warn+ → error → (back to debug+)
```

Only lines at or above the selected level are shown.

## Substring Search

Press `/` to enter filter mode. The text input is shown at the bottom. The filter matches against both `Message` and `Source`. Press `Enter` or `Esc` to exit filter mode.

The status bar displays the active filter query alongside the line count (`filtered/total lines`).

## State Methods

```go
lv.Append(line blit.LogLine) // Add a line (goroutine-safe)
lv.Clear()                     // Remove all lines (goroutine-safe)
lv.Lines() []blit.LogLine    // Snapshot of all stored lines (goroutine-safe)
```

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Scroll up (pauses auto-scroll) |
| `down` / `j` | Scroll down |
| `pgup` | Half-page up |
| `pgdown` | Half-page down |
| `end` | Jump to latest (resumes auto-scroll) |
| `p` | Toggle pause / resume |
| `c` | Clear all lines |
| `/` | Open substring filter input |
| `l` | Cycle minimum log level |

## Example: Background Appender

```go
lv := blit.NewLogViewer()

app := blit.NewApp(
    blit.WithComponent("logs", lv),
)

go func() {
    for line := range logStream {
        lv.Append(blit.LogLine{
            Level:     blit.LogInfo,
            Timestamp: time.Now(),
            Message:   line,
        })
    }
}()

app.Run()
```
