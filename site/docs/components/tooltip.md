# Tooltip

Floating overlay with anchor positioning, escape-to-dismiss, and background compositing. Implements `Component`, `Themed`, and `FloatingOverlay`.

## Construction

```go
func NewTooltip(opts TooltipOpts) *Tooltip
```

## Types

```go
type TooltipOpts struct {
    Text     string // Tooltip content
    MaxWidth int    // Maximum width (default 40)
}
```

## Usage

```go
tip := blit.NewTooltip(blit.TooltipOpts{
    Text:     "Press Enter to confirm your selection",
    MaxWidth: 50,
})

// Position and show
tip.SetAnchor(10, 5)
tip.Show()
```

## Floating Overlay

Tooltip implements `FloatingOverlay`, meaning it composites over background content rather than replacing it:

```go
// Render composited over the main view
output := tip.FloatView(backgroundContent)
```

Use `View()` for standalone rendering or `FloatView(background)` to overlay on existing content.

## Keybindings

| Key | Action |
|-----|--------|
| `esc` | Dismiss tooltip |

## State

| Method | Description |
|--------|-------------|
| `Text()` | Current tooltip text |
| `SetText(text)` | Update content |
| `SetAnchor(x, y)` | Set position |
| `Show()` | Display the tooltip |
| `IsActive()` | Whether tooltip is visible |
| `Close()` | Hide the tooltip |
