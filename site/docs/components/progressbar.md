# ProgressBar

Value-based progress indicator with label and percentage display. Implements `Component` and `Themed`.

## Construction

```go
func NewProgressBar(opts ProgressBarOpts, value float64) *ProgressBar
```

## Types

```go
type ProgressBarOpts struct {
    Label       string // Optional text left of bar
    ShowPercent bool   // Display percentage right of bar
    Width       int    // Override bar width (0 = use component width)
}
```

## Usage

```go
bar := blit.NewProgressBar(blit.ProgressBarOpts{
    Label:       "Uploading",
    ShowPercent: true,
}, 0.0)
```

Update progress programmatically:

```go
bar.SetValue(0.5)    // set to 50%
bar.Increment(0.1)   // advance by 10%
```

Values are clamped to the `[0.0, 1.0]` range.

## Keybindings

None — ProgressBar is a display-only component.

## State

| Method | Description |
|--------|-------------|
| `Value()` | Current progress (0.0–1.0) |
| `SetValue(v)` | Set progress (clamped) |
| `Increment(delta)` | Add to current value |
