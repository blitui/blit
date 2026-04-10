# Picker

A fuzzy-searchable command palette that can be used as either a full-screen overlay (Cmd-K style) or an embedded component. Implements `Component`, `Themed`, and the `Overlay` interface. Built on an internal fuzzy scorer.

## Construction

```go
picker := blit.NewPicker(items []blit.PickerItem, opts blit.PickerOpts)
```

## PickerItem

```go
type PickerItem struct {
    Title    string        // Primary display label (used for fuzzy matching)
    Subtitle string        // Secondary text below title (also matched). Optional.
    Glyph    string        // Optional icon prefix (e.g. "⚡")
    Preview  func() string // Lazy content for the preview pane. Optional.
    Value    any           // Arbitrary payload the caller can attach.
}
```

## PickerOpts

```go
type PickerOpts struct {
    Placeholder string               // Placeholder text in filter input (default "Type to filter...")
    Preview     bool                 // Enable right-side preview pane (60/40 split)
    OnConfirm   func(item PickerItem) // Called when Enter is pressed on an item
    OnCancel    func()               // Called when Esc is pressed without confirming
}
```

## Basic Usage

```go
picker := blit.NewPicker([]blit.PickerItem{
    {Title: "Open file",       Glyph: "📄", Value: "open"},
    {Title: "Run tests",       Glyph: "✓",  Value: "test"},
    {Title: "Deploy to prod",  Glyph: "🚀", Value: "deploy"},
}, blit.PickerOpts{
    Placeholder: "What do you want to do?",
    OnConfirm: func(item blit.PickerItem) {
        dispatch(item.Value.(string))
    },
    OnCancel: func() {
        // picker dismissed without selection
    },
})
```

## Overlay Mode (Cmd-K)

Register the picker as an overlay triggered by a key:

```go
blit.WithOverlay("command", "ctrl+k", picker)
```

Pressing `ctrl+k` opens the picker as a modal over the main content. `Esc` closes it.

## Preview Pane

Enable the optional 40%-width preview pane:

```go
blit.PickerOpts{
    Preview: true,
    OnConfirm: func(item blit.PickerItem) { ... },
}

// Attach a lazy preview to each item:
blit.PickerItem{
    Title:   "server-01",
    Preview: func() string {
        return fetchServerDetails("server-01")
    },
}
```

The preview function is called lazily when the item gains cursor focus and the result is cached until the cursor moves.

## Dynamic Items

Replace the item list and re-run the filter at any time:

```go
picker.SetItems(newItems)
```

## Programmatic Control

```go
picker.Open()            // Activate picker, focus input, reset filter
picker.Close()           // Deactivate without firing OnCancel
picker.CursorItem()      // *PickerItem at cursor, or nil
picker.Items()           // Full unfiltered item list
picker.IsActive() bool   // Whether the picker overlay is currently open
```

## Fuzzy Matching

Items are ranked by a fuzzy score against `Title + " " + Subtitle`. Items with zero score are hidden. Results are sorted by score descending. Press `ctrl+k` inside the picker to clear the filter input.

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `ctrl+p` | Move cursor up |
| `down` / `ctrl+n` | Move cursor down |
| `enter` | Confirm selection (fires `OnConfirm`) |
| `esc` | Cancel (fires `OnCancel`, closes overlay) |
| `ctrl+k` | Clear filter input |
