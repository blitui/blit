# Stepper

Multi-step progress indicator with done/active/pending states, forward/back navigation, and completion callbacks. Implements `Component` and `Themed`.

## Construction

```go
func NewStepper(steps []Step, opts StepperOpts) *Stepper
```

## Types

```go
type Step struct {
    Title       string // Step label
    Description string // Optional detail text
}

type StepperOpts struct {
    OnComplete func()         // Called when advancing past the last step
    OnChange   func(step int) // Called when the current step changes
}

type StepStatus int
const (
    StepPending StepStatus = iota
    StepActive
    StepDone
)
```

## Usage

```go
stepper := blit.NewStepper([]blit.Step{
    {Title: "Account", Description: "Create your account"},
    {Title: "Profile", Description: "Set up your profile"},
    {Title: "Confirm", Description: "Review and submit"},
}, blit.StepperOpts{
    OnComplete: func() { /* submit form */ },
    OnChange:   func(step int) { /* update view */ },
})
```

## Step Status

Each step is automatically assigned a status based on the current position:

- Steps before the current index are **Done**
- The current step is **Active**
- Steps after are **Pending**

Query status with `stepper.Status(index)`.

## Keybindings

| Key | Action |
|-----|--------|
| `right` / `l` / `tab` | Next step |
| `left` / `h` / `shift+tab` | Previous step |

## State

| Method | Description |
|--------|-------------|
| `Steps()` | Returns all steps |
| `Current()` | Active step index |
| `SetCurrent(idx)` | Set active step (clamped to range) |
| `Next()` | Advance one step (fires OnComplete at end) |
| `Prev()` | Go back one step |
| `Status(idx)` | Status of step at index |
