# Dialog

Modal overlay with button navigation, centered rendering, and close callbacks. Implements `Component`, `Themed`, and `Overlay`.

## Construction

```go
func NewDialog(opts DialogOpts) *Dialog
```

## Types

```go
type DialogButton struct {
    Label  string // Button text
    Action func() // Called when activated (optional)
}

type DialogOpts struct {
    Title   string         // Optional text in top border
    Body    string         // Main content
    Buttons []DialogButton // Action buttons (defaults to single "OK")
    OnClose func()         // Called on dismiss
}
```

## Usage

```go
dlg := blit.NewDialog(blit.DialogOpts{
    Title: "Confirm Delete",
    Body:  "Are you sure you want to delete this item?",
    Buttons: []blit.DialogButton{
        {Label: "Cancel", Action: func() { /* dismiss */ }},
        {Label: "Delete", Action: func() { /* delete item */ }},
    },
    OnClose: func() { /* cleanup */ },
})
```

If no buttons are provided, a single "OK" button is created automatically.

## Keybindings

| Key | Action |
|-----|--------|
| `left` / `h` / `shift+tab` | Previous button |
| `right` / `l` / `tab` | Next button |
| `enter` / `space` | Activate button |
| `esc` | Close dialog |

## State

| Method | Description |
|--------|-------------|
| `IsActive()` | Whether the dialog is showing |
| `Close()` | Dismiss dialog and call OnClose |
| `CursorIndex()` | Highlighted button index |
