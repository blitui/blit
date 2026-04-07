package tuikit

import (
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTable(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Score", Width: 10, Align: Right, Sortable: true},
	}
	rows := []Row{
		{"Alice", "100"},
		{"Bob", "200"},
	}
	tbl := NewTable(cols, rows, TableOpts{})
	if tbl == nil {
		t.Fatal("NewTable should not return nil")
	}
}

func TestTableCursorClamp(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"A"}, {"B"}, {"C"}}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	for i := 0; i < 10; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if tbl.cursor != 2 {
		t.Errorf("cursor should clamp to 2, got %d", tbl.cursor)
	}

	for i := 0; i < 10; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if tbl.cursor != 0 {
		t.Errorf("cursor should clamp to 0, got %d", tbl.cursor)
	}
}

func TestTableSort(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20, Sortable: true},
		{Title: "Score", Width: 10, Sortable: true},
	}
	rows := []Row{
		{"Charlie", "300"},
		{"Alice", "100"},
		{"Bob", "200"},
	}
	tbl := NewTable(cols, rows, TableOpts{Sortable: true})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	tbl.sortCol = 0
	tbl.sortAsc = true
	tbl.rebuildVisible()
	if tbl.visible[0][0] != "Alice" {
		t.Errorf("expected Alice first, got %s", tbl.visible[0][0])
	}
}

func TestTableCustomSort(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Score", Width: 10, Sortable: true},
	}
	rows := []Row{
		{"Alice", "100"},
		{"Bob", "20"},
		{"Charlie", "300"},
	}

	// Numeric sort on Score column
	numSort := func(a, b Row, sortCol int, sortAsc bool) bool {
		va, _ := strconv.Atoi(a[sortCol])
		vb, _ := strconv.Atoi(b[sortCol])
		if sortAsc {
			return va < vb
		}
		return va > vb
	}

	tbl := NewTable(cols, rows, TableOpts{Sortable: true, SortFunc: numSort})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	tbl.sortCol = 1
	tbl.sortAsc = true
	tbl.rebuildVisible()

	// Numeric: 20 < 100 < 300
	if tbl.visible[0][0] != "Bob" {
		t.Errorf("expected Bob first (score 20), got %s", tbl.visible[0][0])
	}
	if tbl.visible[2][0] != "Charlie" {
		t.Errorf("expected Charlie last (score 300), got %s", tbl.visible[2][0])
	}
}

func TestTableFilter(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"Alice"}, {"Bob"}, {"Alicia"}}
	tbl := NewTable(cols, rows, TableOpts{Filterable: true})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	tbl.filterQuery = "ali"
	tbl.rebuildVisible()
	if len(tbl.visible) != 2 {
		t.Errorf("expected 2 filtered rows, got %d", len(tbl.visible))
	}
}

func TestTablePredicateFilter(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 10},
	}
	rows := []Row{
		{"Alice", "online"},
		{"Bob", "offline"},
		{"Charlie", "online"},
		{"Dave", "offline"},
	}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	// Filter to only online users
	tbl.SetFilter(func(row Row) bool {
		return len(row) > 1 && row[1] == "online"
	})

	if len(tbl.visible) != 2 {
		t.Errorf("expected 2 online rows, got %d", len(tbl.visible))
	}
	if tbl.visible[0][0] != "Alice" {
		t.Errorf("expected Alice first, got %s", tbl.visible[0][0])
	}

	// Clear filter
	tbl.SetFilter(nil)
	if len(tbl.visible) != 4 {
		t.Errorf("expected 4 rows after clearing filter, got %d", len(tbl.visible))
	}
}

func TestTablePredicateAndTextFilter(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 10},
	}
	rows := []Row{
		{"Alice", "online"},
		{"Alicia", "offline"},
		{"Bob", "online"},
	}
	tbl := NewTable(cols, rows, TableOpts{Filterable: true})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	// Both filters active: predicate (online) + text (ali)
	tbl.SetFilter(func(row Row) bool {
		return len(row) > 1 && row[1] == "online"
	})
	tbl.filterQuery = "ali"
	tbl.rebuildVisible()

	// Only Alice matches both (online AND contains "ali")
	if len(tbl.visible) != 1 {
		t.Errorf("expected 1 row matching both filters, got %d", len(tbl.visible))
	}
}

func TestTableCellRenderer(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Score", Width: 10},
	}
	rows := []Row{
		{"Alice", "100"},
	}

	rendered := false
	renderer := func(row Row, colIdx int, isCursor bool, theme Theme) string {
		rendered = true
		if colIdx < len(row) {
			return "[" + row[colIdx] + "]"
		}
		return ""
	}

	tbl := NewTable(cols, rows, TableOpts{CellRenderer: renderer})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	view := tbl.View()
	if !rendered {
		t.Error("CellRenderer should have been called")
	}
	if !strings.Contains(view, "[Alice]") {
		t.Error("view should contain custom rendered '[Alice]'")
	}
}

func TestTableResponsiveColumns(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Extra", Width: 20, MinWidth: 80},
	}
	rows := []Row{{"Alice", "data"}}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetTheme(DefaultTheme())

	tbl.SetSize(120, 24)
	view := tbl.View()
	if !strings.Contains(view, "Extra") {
		t.Error("Extra column should be visible at width 120")
	}

	tbl.SetSize(60, 24)
	view = tbl.View()
	if strings.Contains(view, "Extra") {
		t.Error("Extra column should be hidden at width 60")
	}
}

func TestTableSetRows(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	tbl := NewTable(cols, []Row{{"A"}}, TableOpts{})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	tbl.SetRows([]Row{{"X"}, {"Y"}, {"Z"}})
	if len(tbl.visible) != 3 {
		t.Errorf("expected 3 rows after SetRows, got %d", len(tbl.visible))
	}
}

func TestTableCursorRowAccess(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"Alice"}, {"Bob"}}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetSize(80, 24)

	row := tbl.CursorRow()
	if row[0] != "Alice" {
		t.Errorf("expected cursor row 'Alice', got '%s'", row[0])
	}

	tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	row = tbl.CursorRow()
	if row[0] != "Bob" {
		t.Errorf("expected cursor row 'Bob', got '%s'", row[0])
	}
}

func TestTableMouseScroll(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"A"}, {"B"}, {"C"}, {"D"}}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	// Scroll down
	tbl.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if tbl.cursor != 1 {
		t.Errorf("cursor should be 1 after scroll down, got %d", tbl.cursor)
	}

	// Scroll up
	tbl.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if tbl.cursor != 0 {
		t.Errorf("cursor should be 0 after scroll up, got %d", tbl.cursor)
	}
}

func TestTableRowClick(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"A"}, {"B"}, {"C"}}

	clickedRow := -1
	tbl := NewTable(cols, rows, TableOpts{
		OnRowClick: func(row Row, idx int) {
			clickedRow = idx
		},
	})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	// Enter key triggers OnRowClick
	tbl.Update(tea.KeyMsg{Type: tea.KeyDown}) // move to row 1
	tbl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if clickedRow != 1 {
		t.Errorf("expected click on row 1, got %d", clickedRow)
	}
}

func TestTableSortIndicator(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20, Sortable: true},
	}
	rows := []Row{{"A"}, {"B"}}
	tbl := NewTable(cols, rows, TableOpts{Sortable: true})
	tbl.SetSize(80, 24)
	tbl.SetTheme(DefaultTheme())

	tbl.sortCol = 0
	tbl.sortAsc = true
	view := tbl.View()
	if !strings.Contains(view, "▲") {
		t.Error("should show ascending sort indicator")
	}

	tbl.sortAsc = false
	view = tbl.View()
	if !strings.Contains(view, "▼") {
		t.Error("should show descending sort indicator")
	}
}
