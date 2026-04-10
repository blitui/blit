package blit

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
	}
	if tbl.cursor != 2 {
		t.Errorf("cursor should clamp to 2, got %d", tbl.cursor)
	}

	for i := 0; i < 10; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyUp}, Context{})
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

	tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
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
	tbl.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown}, Context{})
	if tbl.cursor != 1 {
		t.Errorf("cursor should be 1 after scroll down, got %d", tbl.cursor)
	}

	// Scroll up
	tbl.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp}, Context{})
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
	tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{}) // move to row 1
	tbl.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})
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

func TestTableRowStylerBackgroundCoversContent(t *testing.T) {
	// Force truecolor so lipgloss emits ANSI sequences in test.
	prev := lipgloss.DefaultRenderer().ColorProfile()
	lipgloss.DefaultRenderer().SetColorProfile(termenv.TrueColor)
	defer lipgloss.DefaultRenderer().SetColorProfile(prev)

	cols := []Column{
		{Title: "Name", Width: 20},
	}
	rows := []Row{{"Alice"}}

	// RowStyler that sets a background color on every row.
	bgColor := lipgloss.Color("#ff0000")
	tbl := NewTable(cols, rows, TableOpts{
		RowStyler: func(row Row, idx int, isCursor bool, theme Theme) *lipgloss.Style {
			s := lipgloss.NewStyle().Background(bgColor)
			return &s
		},
	})
	tbl.SetSize(80, 10)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	view := tbl.View()

	// The ANSI escape for our background color (48;2;255;0;0 for #ff0000 truecolor)
	// should appear before the cell content "Alice", not just around padding.
	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		if !strings.Contains(line, "Alice") {
			continue
		}
		found = true
		// Find position of "Alice" in the line
		aliceIdx := strings.Index(line, "Alice")
		if aliceIdx < 0 {
			t.Fatal("could not find Alice in line")
		}
		// The substring leading up to (and including) Alice should contain
		// a background escape sequence. 48;2; is the SGR truecolor background prefix.
		prefix := line[:aliceIdx+len("Alice")]
		if !strings.Contains(prefix, "48;2;") {
			t.Errorf("RowStyler background should be applied to cell content; got line: %q", line)
		}
	}
	if !found {
		t.Fatal("could not find any line containing 'Alice' in table view")
	}
}

// B3: cursor tween tests

func TestTable_CursorTweenStartsOnMove(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"Alice"}, {"Bob"}, {"Carol"}}
	tb := NewTable(cols, rows, TableOpts{})
	tb.SetSize(80, 10)
	tb.SetFocused(true)

	// Tween should not be running before any movement
	if tb.cursorTween.Running() {
		t.Fatal("tween should not be running before cursor moves")
	}

	// Move cursor down — tween must start
	tb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, Context{})
	if !tb.cursorTween.Running() {
		t.Error("tween should be running after cursor move")
	}

	// Verify tween finishes within its 120ms window (Progress drains running state)
	time.Sleep(130 * time.Millisecond)
	prog := tb.cursorTween.Progress(time.Now())
	if prog != 1.0 {
		t.Errorf("tween progress should be 1.0 after 130ms, got %f", prog)
	}
	if tb.cursorTween.Running() {
		t.Error("tween should have finished after 130ms")
	}
}

func TestTable_CursorTweenSnapOnNoAnim(t *testing.T) {
	animDisabled = true
	defer func() { animDisabled = false }()

	cols := []Column{{Title: "Name", Width: 20}}
	rows := []Row{{"Alice"}, {"Bob"}}
	tb := NewTable(cols, rows, TableOpts{})
	tb.SetSize(80, 10)
	tb.SetFocused(true)

	tb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, Context{})

	// With animDisabled, Progress() returns 1.0 immediately regardless of elapsed time
	tval := tb.cursorTween.Progress(time.Now())
	if tval != 1.0 {
		t.Errorf("expected tween progress 1.0 when animDisabled, got %f", tval)
	}
}

// ---------------------------------------------------------------------------
// Track D — Virtualized table
// ---------------------------------------------------------------------------

// intProvider is a synthetic TableRowProvider that lazily renders row i as
// ["row-<i>", "<i*10>"]. It tracks the last fetched window so tests can
// assert that only the visible slice was materialized.
type intProvider struct {
	total       int
	lastOffset  int
	lastLimit   int
	totalFetchd int
}

func (p *intProvider) Len() int { return p.total }

func (p *intProvider) Rows(offset, limit int) []Row {
	p.lastOffset = offset
	p.lastLimit = limit
	p.totalFetchd += limit
	if offset >= p.total {
		return nil
	}
	end := offset + limit
	if end > p.total {
		end = p.total
	}
	out := make([]Row, 0, end-offset)
	for i := offset; i < end; i++ {
		out = append(out, Row{"row-" + strconv.Itoa(i), strconv.Itoa(i * 10)})
	}
	return out
}

func TestTableVirtualBasic(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20},
		{Title: "Score", Width: 10, Align: Right},
	}
	p := &intProvider{total: 1_000_000}
	tbl := NewTable(cols, nil, TableOpts{Virtual: true, RowProvider: p})
	tbl.SetSize(80, 10)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	if tbl.RowCount() != 1_000_000 {
		t.Errorf("RowCount = %d, want 1M", tbl.RowCount())
	}
	if tbl.VisibleRowCount() != 1_000_000 {
		t.Errorf("VisibleRowCount = %d, want 1M", tbl.VisibleRowCount())
	}

	view := tbl.View()
	if !strings.Contains(view, "row-0") {
		t.Error("virtual view should contain row-0 at top")
	}

	// Visible window: 10 rows - header (1) - (no detail) = 9 rows, but
	// visibleRows math in the table subtracts 2 (header + filter line).
	if p.lastLimit > 10 {
		t.Errorf("provider should only be asked for ~visible rows, got limit=%d", p.lastLimit)
	}
}

func TestTableVirtualScrollFetchesWindow(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	p := &intProvider{total: 1_000_000}
	tbl := NewTable(cols, nil, TableOpts{Virtual: true, RowProvider: p})
	tbl.SetSize(80, 12)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	// Jump to end
	tbl.Update(tea.KeyMsg{Type: tea.KeyEnd}, Context{})
	if tbl.CursorIndex() != 999_999 {
		t.Errorf("cursor should be 999999, got %d", tbl.CursorIndex())
	}

	view := tbl.View()
	if !strings.Contains(view, "row-999999") {
		t.Error("expected row-999999 in view after End")
	}

	// Provider must not have been asked for anywhere close to 1M rows in a
	// single frame — scrolling is windowed.
	if p.lastLimit > 20 {
		t.Errorf("provider limit=%d exceeds expected visible window", p.lastLimit)
	}
}

func TestTableVirtualCursorNavigation(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	p := &intProvider{total: 500}
	tbl := NewTable(cols, nil, TableOpts{Virtual: true, RowProvider: p})
	tbl.SetSize(80, 12)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	for i := 0; i < 50; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
	}
	if tbl.CursorIndex() != 50 {
		t.Errorf("cursor = %d, want 50", tbl.CursorIndex())
	}
	row := tbl.CursorRow()
	if row == nil || row[0] != "row-50" {
		t.Errorf("CursorRow = %v, want row-50", row)
	}
}

func TestTableVirtualOnRowClick(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	p := &intProvider{total: 1000}
	clicked := -1
	tbl := NewTable(cols, nil, TableOpts{
		Virtual:     true,
		RowProvider: p,
		OnRowClick: func(row Row, idx int) {
			clicked = idx
		},
	})
	tbl.SetSize(80, 12)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
	tbl.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
	tbl.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})
	if clicked != 2 {
		t.Errorf("clicked = %d, want 2", clicked)
	}
}

func TestTableVirtualSetRowsIsNoop(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	p := &intProvider{total: 100}
	tbl := NewTable(cols, nil, TableOpts{Virtual: true, RowProvider: p})
	tbl.SetRows([]Row{{"X"}, {"Y"}}) // should not affect the provider
	if tbl.RowCount() != 100 {
		t.Errorf("RowCount = %d, want 100 (SetRows should be no-op in virtual mode)", tbl.RowCount())
	}
}

func TestTableVirtualFilterCallback(t *testing.T) {
	cols := []Column{{Title: "Name", Width: 20}}
	p := &intProvider{total: 100}
	var lastQuery string
	tbl := NewTable(cols, nil, TableOpts{
		Filterable:     true,
		Virtual:        true,
		RowProvider:    p,
		OnFilterChange: func(q string) { lastQuery = q },
	})
	tbl.SetSize(80, 12)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}, Context{})
	tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r', 'o', 'w'}}, Context{})
	if lastQuery != "row" {
		t.Errorf("OnFilterChange last query = %q, want %q", lastQuery, "row")
	}
	if tbl.FilterQuery() != "row" {
		t.Errorf("FilterQuery = %q, want %q", tbl.FilterQuery(), "row")
	}
}

// BenchmarkTableVirtualRender1M renders a single frame of a virtualized table
// holding 1,000,000 logically-addressable rows. The provider is a trivial
// closure that lazy-formats rows on demand, so the only work measured is the
// table's render pipeline over the visible window — not data generation.
//
// Target: < 2 ms/op on reference hardware (see Track D in
// .omc/plans/v0.9-component-expansion.md). Recent local runs on a Ryzen-class
// laptop land well under 200 us/op because only ~20 rows are rendered per
// frame regardless of total row count — that is the point of virtualization.
func BenchmarkTableVirtualRender1M(b *testing.B) {
	cols := []Column{
		{Title: "ID", Width: 10, Align: Right},
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 10},
		{Title: "Score", Width: 10, Align: Right},
	}
	p := &intProvider{total: 1_000_000}
	tbl := NewTable(cols, nil, TableOpts{Virtual: true, RowProvider: p})
	tbl.SetSize(120, 30)
	tbl.SetTheme(DefaultTheme())
	tbl.SetFocused(true)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate continuous scrolling so the table actually refetches a
		// different window each frame — a worst-case for the virtual path.
		tbl.offset = i % (1_000_000 - 30)
		_ = tbl.View()
	}
}
