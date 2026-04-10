# Split

Resizable split pane that divides space between two child components with a visible divider. Implements `Component` and `Themed`.

## Construction

```go
split := blit.NewSplit(
    blit.Horizontal,  // left/right split
    0.5,                // 50/50 ratio
    leftComponent,
    rightComponent,
)
split.Resizable = true
```

## Split Fields

```go
type Split struct {
    Orientation Orientation // Horizontal (left/right) or Vertical (top/bottom)
    Ratio       float64     // Fraction of space for pane A (0.0–1.0)
    Resizable   bool        // Enable alt+arrow / mouse drag resizing
    A           Component   // First child (left or top)
    B           Component   // Second child (right or bottom)
}
```

## Orientation

| Value | Layout |
|-------|--------|
| `blit.Horizontal` | A on left, B on right |
| `blit.Vertical` | A on top, B on bottom |

## Keyboard

| Key | Action |
|-----|--------|
| `Alt+←` / `Alt+→` | Resize horizontal split |
| `Alt+↑` / `Alt+↓` | Resize vertical split |
| `Tab` | Switch focus between panes |

## Example

```bash
go run ./examples/splitpane/
```
