# Accordion

Collapsible sections with exclusive mode, toggle callbacks, and vi-style navigation. Implements `Component` and `Themed`.

## Construction

```go
func NewAccordion(sections []AccordionSection, opts AccordionOpts) *Accordion
```

## Types

```go
type AccordionSection struct {
    Title    string // Header text
    Content  string // Body text shown when expanded
    Expanded bool   // Initial open state
}

type AccordionOpts struct {
    Exclusive bool                           // Only one section open at a time
    OnToggle  func(index int, expanded bool) // Called on expand/collapse
}
```

## Usage

```go
acc := blit.NewAccordion([]blit.AccordionSection{
    {Title: "Getting Started", Content: "Welcome to the app..."},
    {Title: "Configuration", Content: "Set your preferences..."},
    {Title: "FAQ", Content: "Common questions..."},
}, blit.AccordionOpts{
    Exclusive: true,
    OnToggle: func(index int, expanded bool) {
        // handle toggle
    },
})
```

## Exclusive Mode

When `Exclusive` is true, expanding one section automatically collapses all others — useful for FAQ-style layouts where only one answer should be visible at a time.

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Previous section |
| `down` / `j` | Next section |
| `enter` / `space` | Toggle section |

## State

| Method | Description |
|--------|-------------|
| `Sections()` | Returns current sections |
| `CursorIndex()` | Returns highlighted section index |
