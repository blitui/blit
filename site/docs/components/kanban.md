# Kanban

Multi-column board with card movement between columns, tag badges, and selection callbacks. Implements `Component` and `Themed`.

## Construction

```go
func NewKanban(columns []KanbanColumn, opts KanbanOpts) *Kanban
```

## Types

```go
type KanbanCard struct {
    ID          string // Unique identifier
    Title       string // Display text
    Description string // Optional secondary text
    Tag         string // Optional badge label (e.g., "bug", "feat")
}

type KanbanColumn struct {
    Title string       // Column header
    Cards []KanbanCard // Items in this column
}

type KanbanOpts struct {
    OnMove   func(card KanbanCard, from, to int) // Called when a card moves between columns
    OnSelect func(card KanbanCard, col int)      // Called on enter
}
```

## Usage

```go
board := blit.NewKanban([]blit.KanbanColumn{
    {Title: "To Do", Cards: []blit.KanbanCard{
        {ID: "1", Title: "Design API", Tag: "feat"},
        {ID: "2", Title: "Fix login bug", Tag: "bug"},
    }},
    {Title: "In Progress", Cards: []blit.KanbanCard{
        {ID: "3", Title: "Write tests", Tag: "test"},
    }},
    {Title: "Done", Cards: []blit.KanbanCard{}},
}, blit.KanbanOpts{
    OnMove: func(card blit.KanbanCard, from, to int) {
        // persist card movement
    },
    OnSelect: func(card blit.KanbanCard, col int) {
        // show card details
    },
})
```

## Moving Cards

Use `H` / `L` (or `shift+left` / `shift+right`) to move the selected card between columns. The `OnMove` callback fires with the card and the source/destination column indices.

## Keybindings

| Key | Action |
|-----|--------|
| `left` / `h` | Previous column |
| `right` / `l` | Next column |
| `up` / `k` | Previous card |
| `down` / `j` | Next card |
| `enter` | Select card |
| `H` / `shift+left` | Move card left |
| `L` / `shift+right` | Move card right |

## State

| Method | Description |
|--------|-------------|
| `Columns()` | Returns all columns |
| `ActiveColumn()` | Focused column index |
| `ActiveCard()` | Focused card index within active column |
