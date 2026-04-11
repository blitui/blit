# CommandBar

Inline command palette with tab completion, argument support, and aliased commands. Implements `Component`, `Themed`, and `InlineOverlay`.

## Construction

```go
func NewCommandBar(commands []Command) *CommandBar
```

## Types

```go
type Command struct {
    Name    string               // Primary name (e.g., "quit", "sort")
    Aliases []string             // Alternative names (e.g., ["q"] for "quit")
    Args    bool                 // Whether the command accepts an argument
    Hint    string               // Help text shown in completion
    Run     func(string) tea.Cmd // Handler — receives the argument string
}
```

## Usage

```go
bar := blit.NewCommandBar([]blit.Command{
    {
        Name:    "theme",
        Aliases: []string{"t"},
        Args:    true,
        Hint:    "Switch theme by name",
        Run: func(args string) tea.Cmd {
            return setThemeCmd(args)
        },
    },
    {
        Name: "quit",
        Aliases: []string{"q"},
        Hint: "Exit the application",
        Run: func(_ string) tea.Cmd {
            return tea.Quit
        },
    },
})
```

## Inline Overlay

CommandBar renders as a single line at the bottom of the screen (implements `InlineOverlay`). It does not take over the full viewport like modal overlays.

## Tab Completion

Press `tab` to cycle through matching command names based on the current input.

## Keybindings

| Key | Action |
|-----|--------|
| `enter` | Execute command |
| `tab` | Tab complete |
| `esc` | Close command bar |
| `backspace` | Delete character |

## State

| Method | Description |
|--------|-------------|
| `IsActive()` | Whether the command bar is visible |
| `SetActive(v)` | Show or hide the command bar |
| `Close()` | Deactivate and reset input |
