# Tree

Collapsible tree view for hierarchical data with cursor navigation, expand/collapse, and optional icons. Implements `Component` and `Themed`.

## Construction

```go
tree := tuikit.NewTree([]*tuikit.Node{
    {
        Title: "src",
        Expanded: true,
        Children: []*tuikit.Node{
            {Title: "main.go", Glyph: "📄"},
            {Title: "app.go", Glyph: "📄"},
            {
                Title: "internal",
                Children: []*tuikit.Node{
                    {Title: "fuzzy.go"},
                },
            },
        },
    },
}, tuikit.TreeOpts{
    OnSelect: func(node *tuikit.Node) {
        fmt.Println("Selected:", node.Title)
    },
})
```

## Node

```go
type Node struct {
    Title    string  // Display label
    Glyph    string  // Optional icon prefix
    Children []*Node // Child nodes (non-nil = expandable)
    Data     any     // Arbitrary payload
    Expanded bool    // Whether children are visible
}
```

## TreeOpts

```go
type TreeOpts struct {
    OnSelect func(node *Node) // Called on Enter
    OnToggle func(node *Node) // Called on expand/collapse
}
```

## Keyboard

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `Enter` / `l` | Expand node or trigger OnSelect on leaf |
| `h` | Collapse node |
| `space` | Toggle expand/collapse |

## Example

```bash
go run ./examples/filetree/
```
