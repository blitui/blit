# ConfigEditor

Settings overlay with grouped fields, inline editing, and validation. Implements `Component`, `Themed`, and `Overlay`.

## Construction

```go
func NewConfigEditor(fields []ConfigField) *ConfigEditor
```

## Types

```go
type ConfigField struct {
    Label    string             // Display label
    Group    string             // Group heading (e.g., "General", "Display")
    Hint     string             // Help text shown below field
    Get      func() string      // Value getter (legacy — use Source instead)
    Source   any                // func() string, *Signal[string], or StringSource
    Set      func(string) error // Sets new value, returns error on failure
    Validate func(string) error // Validates input before calling Set
}
```

## Usage

```go
editor := blit.NewConfigEditor([]blit.ConfigField{
    {
        Label: "Username",
        Group: "Account",
        Hint:  "Your display name",
        Get:   func() string { return cfg.Username },
        Set:   func(v string) error { cfg.Username = v; return nil },
    },
    {
        Label:    "Port",
        Group:    "Network",
        Hint:     "Server port (1024-65535)",
        Get:      func() string { return fmt.Sprint(cfg.Port) },
        Set:      func(v string) error { /* parse and set */ },
        Validate: func(v string) error { /* validate range */ },
    },
})
```

## Field Groups

Fields with the same `Group` value are displayed under a shared heading. Groups appear in the order of the first field in each group.

## Reactive Sources

Instead of the legacy `Get` function, you can use a `*Signal[string]` or any `StringSource` for reactive updates:

```go
username := blit.NewSignal("alice")
blit.ConfigField{
    Label:  "Username",
    Source: username,
    Set:    func(v string) error { username.Set(v); return nil },
}
```

## Keybindings

**Navigation mode:**

| Key | Action |
|-----|--------|
| `up` / `k` | Previous field |
| `down` / `j` | Next field |
| `enter` | Edit field |
| `esc` / `q` | Close editor |

**Edit mode:**

| Key | Action |
|-----|--------|
| `enter` | Confirm edit |
| `esc` | Cancel edit |
| `backspace` | Delete character |

## State

| Method | Description |
|--------|-------------|
| `IsActive()` | Whether the editor is showing |
| `Close()` | Deactivate and reset state |
