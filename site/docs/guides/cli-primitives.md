# CLI Primitives

The `cli` package provides interactive prompts for tools that need more than `fmt.Print` but less than a full TUI. Each primitive runs a minimal Bubble Tea program, captures input, and returns the result.

```go
import "github.com/blitui/blit/cli"
```

## Prompts

### Confirm

```go
proceed := cli.Confirm("Deploy to production?", false)
// Returns bool. Second arg is the default when Enter is pressed.
```

### SelectOne

```go
lang, idx, err := cli.SelectOne("Language:", []string{"Go", "Rust", "Python"})
// Returns selected string, its index, and any error.
// Auto-enables type-to-filter when list has 10+ items.
```

### MultiSelect

```go
selected, indices, err := cli.MultiSelect("Features:", []string{"Auth", "DB", "Cache"})
// Returns selected strings and their indices. Space to toggle, Enter to confirm.
```

### Input

```go
name, err := cli.Input("Project name:", func(s string) error {
    if s == "" {
        return fmt.Errorf("required")
    }
    return nil
})
```

### Password

```go
secret, err := cli.Password("API token:", nil)
// Input is masked. Nil validator accepts anything.
```

## Progress Indicators

### Spinner

```go
s := cli.Spin("Installing...")
// do work
s.Stop()
```

### Progress Bar

```go
bar := cli.NewProgress(100, "Downloading")
bar.Increment(25)
bar.Increment(75)
bar.Done()
```

## Styled Message Helpers

Consistent formatting for CLI output:

```go
cli.Title("Setup Wizard")          // Bold underlined title
cli.Step(1, 3, "Installing deps")  // [1/3] numbered step
cli.Success("Build complete")      // ✓ green
cli.Warning("Deprecated flag")     // ! yellow
cli.Error("Connection failed")     // ✗ red
cli.Info("Using defaults")         // ℹ blue
cli.Separator()                    // ────────────
cli.KeyValue("Version", "1.2.3")   // dimmed key: value
```
