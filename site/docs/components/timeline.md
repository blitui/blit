# Timeline

Vertical or horizontal event sequence with status icons, scrolling, and selection callbacks. Implements `Component` and `Themed`.

## Construction

```go
func NewTimeline(events []TimelineEvent, opts TimelineOpts) *Timeline
```

## Types

```go
type TimelineEvent struct {
    Time        string // Timestamp label (e.g., "2024-01-15")
    Title       string // Event headline
    Description string // Optional detail text
    Status      string // "done", "active", "pending", or custom
}

type TimelineOpts struct {
    OnSelect   func(event TimelineEvent, index int) // Called on enter
    Horizontal bool                                  // Left-to-right layout
}
```

## Usage

```go
tl := blit.NewTimeline([]blit.TimelineEvent{
    {Time: "Jan 1", Title: "Project Started", Status: "done"},
    {Time: "Feb 15", Title: "Alpha Release", Status: "done"},
    {Time: "Mar 1", Title: "Beta Testing", Status: "active"},
    {Time: "Apr 1", Title: "Launch", Status: "pending"},
}, blit.TimelineOpts{
    OnSelect: func(event blit.TimelineEvent, index int) {
        // show event details
    },
})
```

## Orientation

By default, events display vertically (top to bottom). Set `Horizontal: true` for a left-to-right layout — navigation keys change accordingly.

## Status Icons

| Status | Icon |
|--------|------|
| `done` | Checkmark |
| `active` | Star |
| `pending` | Dot |

Status also determines color: done uses Positive, active uses Accent, pending uses Muted.

## Keybindings

**Vertical (default):**

| Key | Action |
|-----|--------|
| `up` / `k` | Previous event |
| `down` / `j` | Next event |
| `enter` | Select event |

**Horizontal:**

| Key | Action |
|-----|--------|
| `left` / `h` | Previous event |
| `right` / `l` | Next event |
| `enter` | Select event |

## State

| Method | Description |
|--------|-------------|
| `Events()` | Returns current events |
| `SetEvents(events)` | Replace events, reset cursor |
| `CursorIndex()` | Highlighted event index |
