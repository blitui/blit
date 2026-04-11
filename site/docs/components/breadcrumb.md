# Breadcrumb

Navigable path display with Push/Pop API, custom separators, and selection callbacks. Implements `Component` and `Themed`.

## Construction

```go
func NewBreadcrumb(items []BreadcrumbItem, opts BreadcrumbOpts) *Breadcrumb
```

## Types

```go
type BreadcrumbItem struct {
    Label string // Display text
    Data  any    // Arbitrary payload
}

type BreadcrumbOpts struct {
    Separator string                               // Defaults to " > "
    OnSelect  func(item BreadcrumbItem, index int) // Called on enter
}
```

## Usage

```go
bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{
    {Label: "Home"},
    {Label: "Settings"},
    {Label: "Theme"},
}, blit.BreadcrumbOpts{
    Separator: " / ",
    OnSelect: func(item blit.BreadcrumbItem, index int) {
        // navigate to item
    },
})
```

## Push / Pop

Dynamically modify the path at runtime:

```go
bc.Push(blit.BreadcrumbItem{Label: "Advanced"})
removed := bc.Pop() // returns the removed item
```

`Push` appends an item and moves the cursor to it. `Pop` removes and returns the last item, moving the cursor back.

## Keybindings

| Key | Action |
|-----|--------|
| `left` / `h` | Previous item |
| `right` / `l` | Next item |
| `enter` | Select item |

## State

| Method | Description |
|--------|-------------|
| `Items()` | Returns current items |
| `SetItems(items)` | Replaces all items, resets cursor |
| `CursorIndex()` | Returns highlighted item index |
