# Table

Adaptive table with responsive columns, sorting, filtering, custom cell rendering, detail bars, and virtual mode for large datasets. Implements `Component` and `Themed`.

## Construction

```go
table := blit.NewTable(columns []blit.Column, rows []blit.Row, opts blit.TableOpts)
```

## Column

```go
type Column struct {
    Title      string         // Column header text
    Width      int            // Proportional width weight
    MaxWidth   int            // Cap at this many characters (0 = no cap)
    MinWidth   int            // Hide when terminal width is below this (0 = always show)
    Align      blit.Alignment // Left (default), Right, Center
    Sortable   bool           // Whether this column participates in sort cycling
    NoRowStyle bool           // Exempt from row-level background styling (e.g. for sparkline columns)
}
```

## TableOpts

```go
type TableOpts struct {
    Sortable       bool               // Enable sort cycling with 's'
    Filterable     bool               // Enable '/' search mode
    CellRenderer   CellRenderer       // Custom per-cell styling (nil = plain text)
    RowStyler      RowStyler          // Full-row background override (cursor flash, alerts)
    SortFunc       SortFunc           // Custom sort (nil = lexicographic)
    OnRowClick     RowClickHandler    // Called on Enter or mouse click
    DetailFunc     DetailRenderer     // Inline detail bar for cursor row
    DetailHeight   int                // Lines reserved for detail bar (default 3 when DetailFunc set)
    OnCursorChange CursorChangeHandler // Called when cursor moves
    Virtual        bool               // Enable virtualized render path
    RowProvider    TableRowProvider   // Data source for virtual mode
    OnFilterChange func(query string) // Called when filter text changes (virtual mode hook)
}
```

## Basic Usage

```go
columns := []blit.Column{
    {Title: "Name",  Width: 20, Sortable: true},
    {Title: "Score", Width: 10, Align: blit.Right, Sortable: true},
    {Title: "Extra", Width: 15, MinWidth: 100}, // hides below 100-column terminals
}

table := blit.NewTable(columns, rows, blit.TableOpts{
    Sortable:   true, // 's' to cycle sort
    Filterable: true, // '/' to search
})
```

## Updating Data

```go
table.SetRows(newRows)           // Replace all rows; rebuilds sorted/filtered view
table.SetFilter(func(row blit.Row) bool {
    return row[1] == "Online"    // Predicate filter (works alongside text search)
})
table.SetFilter(nil)             // Clear predicate filter
table.SetSort(2, true)           // Sort by column 2, ascending
table.SetCursor(5)               // Move cursor to row 5
```

## Custom Cell Rendering

`CellRenderer` gives full control over per-cell styling using lipgloss:

```go
blit.TableOpts{
    CellRenderer: func(row blit.Row, colIdx int, isCursor bool, theme blit.Theme) string {
        val := row[colIdx]
        if colIdx == 1 && val == "Online" {
            return lipgloss.NewStyle().
                Foreground(lipgloss.Color(theme.Positive)).
                Render(val)
        }
        return val
    },
}
```

## Row Styler

Apply a full-row background for cursor, flash effects, or alert rows. Works in combination with `CellRenderer`:

```go
blit.TableOpts{
    RowStyler: func(row blit.Row, idx int, isCursor bool, theme blit.Theme) *lipgloss.Style {
        if row[2] == "error" {
            s := lipgloss.NewStyle().Background(lipgloss.Color(theme.Negative))
            return &s
        }
        return nil // no row-level style
    },
}
```

## Custom Sort

Override the default lexicographic sort with numeric, time-based, or any comparison:

```go
blit.TableOpts{
    SortFunc: func(a, b blit.Row, sortCol int, sortAsc bool) bool {
        va, _ := strconv.ParseFloat(a[sortCol], 64)
        vb, _ := strconv.ParseFloat(b[sortCol], 64)
        if sortAsc {
            return va < vb
        }
        return va > vb
    },
}
```

## Detail Bar

Render a contextual detail panel below the table for the cursor row:

```go
blit.TableOpts{
    DetailFunc: func(row blit.Row, rowIdx int, width int, theme blit.Theme) string {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.Muted)).
            Render(fmt.Sprintf("Details for %s: %s", row[0], row[1]))
    },
    DetailHeight: 3, // lines reserved (default 3)
}
```

## Virtual Mode (Large Datasets)

For millions of rows, enable virtual mode. Only the visible window is fetched per frame:

```go
provider := blit.TableRowProviderFunc{
    Total: 1_000_000,
    Fetch: func(offset, limit int) []blit.Row {
        return myDB.Query(offset, limit)
    },
}

table := blit.NewTable(columns, nil, blit.TableOpts{
    Virtual:     true,
    RowProvider: provider,
    OnFilterChange: func(query string) {
        // Re-query provider with filtered data
    },
})
```

!!! note
    Sorting is disabled in virtual mode. Implement sorting inside your `TableRowProvider`.

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Move cursor up |
| `down` / `j` | Move cursor down |
| `ctrl+u` | Half page up |
| `ctrl+d` | Half page down |
| `home` / `g` | Jump to top |
| `end` / `G` | Jump to bottom |
| `s` | Cycle sort (requires `Sortable: true`) |
| `/` | Enter search mode (requires `Filterable: true`) |
| `enter` | Select row (fires `OnRowClick`) |
| `esc` | Cancel search |

## State Accessors

```go
table.CursorRow()        // Row at cursor, or nil
table.CursorIndex()      // Current cursor position
table.FilterQuery()      // Active text search query
table.SortCol()          // Current sort column (-1 if none)
table.SortAsc()          // Sort direction
table.RowCount()         // Total rows (before filtering)
table.VisibleRowCount()  // Rows after filtering
```
