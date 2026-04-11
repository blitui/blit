# Spinner

Animated loading indicator that cycles through glyph frames with a configurable interval. Implements `Component` and `Themed`.

## Construction

```go
func NewSpinner(opts SpinnerOpts) *Spinner
```

## Types

```go
type SpinnerOpts struct {
    Label    string        // Displayed next to the animation
    Interval time.Duration // Time between frame advances (default 80ms)
}
```

## Usage

```go
sp := blit.NewSpinner(blit.SpinnerOpts{
    Label:    "Loading...",
    Interval: 100 * time.Millisecond,
})
```

The spinner uses glyph frames from the active theme. Call `Init()` to start the tick loop.

## Controlling the Spinner

```go
sp.SetActive(true)   // start animating
sp.SetActive(false)  // stop
sp.SetLabel("Done")  // update label text
```

## Keybindings

None — Spinner is a display-only component.

## State

| Method | Description |
|--------|-------------|
| `Label()` | Current label text |
| `SetLabel(label)` | Update label |
| `Active()` | Whether animation is running |
| `SetActive(active)` | Start or stop animation |
| `Frame()` | Current frame index |
