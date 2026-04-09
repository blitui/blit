package tuikit

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CursorStyle controls how the cursor is rendered in a table.
type CursorStyle int

const (
	// CursorRow highlights the entire row.
	CursorRow CursorStyle = iota
)

// Column defines a table column with responsive visibility.
type Column struct {
	Title      string    // Column header text
	Width      int       // Proportional width weight
	MaxWidth   int       // Cap column at this many characters (0 = no cap)
	MinWidth   int       // Hide column when terminal width is below this (0 = always show)
	Align      Alignment // Text alignment within the column
	Sortable   bool      // Whether this column supports sorting
	NoRowStyle bool      // Exempt this column from row-level background styling
}

// Row is a slice of cell values, one per column.
type Row []string

// CellRenderer renders a single cell with full styling control.
// Parameters: row data, column index, whether this row has cursor, theme.
// Return a styled string (use lipgloss). If nil, the table uses plain text.
type CellRenderer func(row Row, colIdx int, isCursor bool, theme Theme) string

// SortFunc compares two rows for custom sorting.
// sortCol is the current sort column index, sortAsc is the sort direction.
// Return true if row a should come before row b.
type SortFunc func(a, b Row, sortCol int, sortAsc bool) bool

// FilterFunc returns true if a row should be visible.
// Used for predicate-based filtering (e.g., show only gainers, filter by type).
type FilterFunc func(row Row) bool

// RowClickHandler is called when a row is clicked or Enter is pressed.
type RowClickHandler func(row Row, rowIdx int)

// RowStyler returns an optional lipgloss.Style to apply to an entire row.
// Use this for full-row highlights (cursor, flash effects, alerts) when
// a CellRenderer is also in use. Return nil for no row-level styling.
type RowStyler func(row Row, idx int, isCursor bool, theme Theme) *lipgloss.Style

// DetailRenderer renders a detail bar for the selected row.
// Parameters: row data, row index, available width, theme.
// Return a styled multi-line string. Shown below the table rows when focused.
type DetailRenderer func(row Row, rowIdx int, width int, theme Theme) string

// CursorChangeHandler is called when the cursor moves to a different row.
type CursorChangeHandler func(row Row, rowIdx int)

// TableOpts configures a Table component.
type TableOpts struct {
	Sortable       bool              // Enable sort cycling with 's'
	Filterable     bool              // Enable '/' search mode
	CursorStyle    CursorStyle       // How the cursor is rendered
	HeaderStyle    lipgloss.Style    // Override header style (zero value = use theme)
	CellRenderer   CellRenderer      // Custom cell renderer (nil = plain text)
	RowStyler      RowStyler         // Optional full-row style (applied after cell rendering)
	SortFunc       SortFunc          // Custom sort function (nil = lexicographic)
	OnRowClick     RowClickHandler   // Called on Enter or mouse click on a row
	DetailFunc     DetailRenderer    // Renders inline detail bar for cursor row (nil = no detail bar)
	DetailHeight   int               // Lines reserved for detail bar (default 3 when DetailFunc set)
	OnCursorChange CursorChangeHandler // Called when cursor moves to a different row
}

// Table is an adaptive table component with sorting, filtering, and responsive columns.
type Table struct {
	columns     []Column
	rows        []Row
	visible     []Row // filtered/sorted view of rows
	opts        TableOpts
	theme       Theme
	focused     bool
	width       int
	height      int
	cursor      int
	offset      int // scroll offset
	sortCol     int // -1 = no sort
	sortAsc     bool
	filtering   bool       // in search mode
	filterQuery string     // current filter text
	filterFunc  FilterFunc // predicate filter
}

// NewTable creates a new Table component.
func NewTable(columns []Column, rows []Row, opts TableOpts) *Table {
	if opts.DetailHeight == 0 && opts.DetailFunc != nil {
		opts.DetailHeight = 3
	}
	t := &Table{
		columns: columns,
		rows:    rows,
		opts:    opts,
		sortCol: -1,
		sortAsc: true,
	}
	t.rebuildVisible()
	return t
}

// SetRows replaces the table data and rebuilds the view.
func (t *Table) SetRows(rows []Row) {
	t.rows = rows
	t.rebuildVisible()
	t.clampCursor()
}

// SetFilter sets a predicate filter function. Pass nil to clear.
// This works alongside text search — both must pass for a row to be visible.
func (t *Table) SetFilter(fn FilterFunc) {
	t.filterFunc = fn
	t.rebuildVisible()
	t.clampCursor()
}

// SortCol returns the current sort column index (-1 if no sort).
func (t *Table) SortCol() int { return t.sortCol }

// SortAsc returns whether the current sort is ascending.
func (t *Table) SortAsc() bool { return t.sortAsc }

// CursorRow returns the row at the current cursor position, or nil if empty.
func (t *Table) CursorRow() Row {
	if t.cursor >= 0 && t.cursor < len(t.visible) {
		return t.visible[t.cursor]
	}
	return nil
}

// CursorIndex returns the current cursor position.
func (t *Table) CursorIndex() int { return t.cursor }

// SetCursor moves the cursor to the given index and scrolls to keep it visible.
func (t *Table) SetCursor(idx int) {
	t.cursor = idx
	t.clampCursor()
}

// SetSort sets the sort column and direction, then rebuilds the view.
// Pass col=-1 to clear sorting.
func (t *Table) SetSort(col int, asc bool) {
	t.sortCol = col
	t.sortAsc = asc
	t.rebuildVisible()
	t.clampCursor()
}

// RowCount returns the total number of rows (before filtering).
func (t *Table) RowCount() int { return len(t.rows) }

// VisibleRowCount returns the number of rows after filtering.
func (t *Table) VisibleRowCount() int { return len(t.visible) }

func (t *Table) Init() tea.Cmd { return nil }

func (t *Table) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKey(msg)
	case tea.MouseMsg:
		return t.handleMouse(msg)
	}
	return t, nil
}

func (t *Table) handleMouse(msg tea.MouseMsg) (Component, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		t.cursor--
		t.clampCursor()
		return t, Consumed()
	case tea.MouseButtonWheelDown:
		t.cursor++
		t.clampCursor()
		return t, Consumed()
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionRelease {
			return t, nil // Ignore release to prevent double-fire
		}
		// Calculate which row was clicked (Y relative to table content)
		clickedRow := t.offset + msg.Y - 1 // -1 for header
		if clickedRow >= 0 && clickedRow < len(t.visible) {
			t.cursor = clickedRow
			t.clampCursor()
			if t.opts.OnRowClick != nil {
				t.opts.OnRowClick(t.visible[t.cursor], t.cursor)
			}
			return t, Consumed()
		}
	}
	return t, nil
}

func (t *Table) handleKey(msg tea.KeyMsg) (Component, tea.Cmd) {
	if t.filtering {
		switch msg.String() {
		case "esc":
			t.filtering = false
			t.filterQuery = ""
			t.rebuildVisible()
			t.clampCursor()
			return t, Consumed()
		case "backspace":
			if len(t.filterQuery) > 0 {
				t.filterQuery = t.filterQuery[:len(t.filterQuery)-1]
				t.rebuildVisible()
				t.clampCursor()
			}
			return t, Consumed()
		case "enter":
			t.filtering = false
			return t, Consumed()
		default:
			// Accept any printable rune (space, unicode letters, emoji…)
			// instead of the old ASCII 32-126 restriction.
			added := false
			if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
				runes := msg.Runes
				if msg.Type == tea.KeySpace && len(runes) == 0 {
					runes = []rune{' '}
				}
				for _, r := range runes {
					if unicode.IsPrint(r) {
						t.filterQuery += string(r)
						added = true
					}
				}
			}
			if added {
				t.rebuildVisible()
				t.clampCursor()
			}
			return t, Consumed()
		}
	}

	switch msg.String() {
	case "up", "k":
		t.cursor--
		t.clampCursor()
		return t, Consumed()
	case "down", "j":
		t.cursor++
		t.clampCursor()
		return t, Consumed()
	case "home", "g":
		t.cursor = 0
		t.clampCursor()
		return t, Consumed()
	case "end", "G":
		t.cursor = len(t.visible) - 1
		t.clampCursor()
		return t, Consumed()
	case "ctrl+d":
		half := (t.height - 2) / 2
		if half < 1 {
			half = 1
		}
		t.cursor += half
		t.clampCursor()
		return t, Consumed()
	case "ctrl+u":
		half := (t.height - 2) / 2
		if half < 1 {
			half = 1
		}
		t.cursor -= half
		t.clampCursor()
		return t, Consumed()
	case "enter":
		if t.opts.OnRowClick != nil && t.cursor < len(t.visible) {
			t.opts.OnRowClick(t.visible[t.cursor], t.cursor)
			return t, Consumed()
		}
	case "s":
		if t.opts.Sortable {
			t.cycleSort()
			return t, Consumed()
		}
	case "/":
		if t.opts.Filterable {
			t.filtering = true
			t.filterQuery = ""
			return t, Consumed()
		}
	}

	return t, nil
}

func (t *Table) View() string {
	if t.width == 0 || t.height == 0 {
		return ""
	}

	visibleCols, origIdxs := t.visibleColumnsWithIndices()
	colWidths := t.computeWidths(visibleCols)

	// Exactly t.height lines, joined without trailing \n
	lines := make([]string, 0, t.height)

	// Header
	lines = append(lines, t.renderHeader(visibleCols, origIdxs, colWidths))

	// Rows
	visibleRows := t.height - 2 // header + possible filter line
	if t.opts.DetailFunc != nil {
		visibleRows -= t.opts.DetailHeight
	}
	if visibleRows < 0 {
		visibleRows = 0
	}

	end := t.offset + visibleRows
	if end > len(t.visible) {
		end = len(t.visible)
	}

	for i := t.offset; i < end; i++ {
		lines = append(lines, t.renderRow(t.visible[i], i, visibleCols, origIdxs, colWidths))
	}

	if t.filtering && len(lines) > 0 {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Accent))
		lines[len(lines)-1] = filterStyle.Render(fmt.Sprintf(" / %s█", t.filterQuery))
	}

	// Detail bar
	if t.opts.DetailFunc != nil {
		detail := t.renderDetail()
		detailLines := strings.Split(detail, "\n")
		lines = append(lines, detailLines...)
	}

	// Pad to exactly t.height lines so JoinHorizontal with a separator
	// doesn't misalign when there are fewer rows than available space.
	for len(lines) < t.height {
		lines = append(lines, "")
	}
	if len(lines) > t.height {
		lines = lines[:t.height]
	}
	return strings.Join(lines, "\n")
}

func (t *Table) renderHeader(cols []Column, origIdxs []int, widths []int) string {
	headerStyle := t.opts.HeaderStyle
	if headerStyle.Value() == "" {
		headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.theme.Muted)).
			Bold(true)
	}

	var parts []string
	for i, col := range cols {
		title := col.Title
		// Show sort indicator using the column's original index (not a
		// Title-based lookup, which breaks when two columns share a title).
		origIdx := origIdxs[i]
		if origIdx == t.sortCol {
			if t.sortAsc {
				title += " ▲"
			} else {
				title += " ▼"
			}
		}
		cell := t.alignCell(title, widths[i], col.Align)
		parts = append(parts, cell)
	}
	return headerStyle.Render(strings.Join(parts, " "))
}

func (t *Table) renderRow(row Row, idx int, cols []Column, origIdxs []int, widths []int) string {
	isCursor := idx == t.cursor && t.focused

	// Determine row-level style first so separators can be styled too
	var rowStyle *lipgloss.Style
	if t.opts.RowStyler != nil {
		rowStyle = t.opts.RowStyler(row, idx, isCursor, t.theme)
	}
	if rowStyle == nil && isCursor {
		cs := lipgloss.NewStyle().
			Background(lipgloss.Color(t.theme.Cursor)).
			Foreground(lipgloss.Color(t.theme.TextInverse))
		rowStyle = &cs
	}

	var parts []string
	for i, col := range cols {
		origIdx := origIdxs[i]

		var cellContent string
		if t.opts.CellRenderer != nil {
			cellContent = t.opts.CellRenderer(row, origIdx, isCursor, t.theme)
		} else {
			if origIdx < len(row) {
				cellContent = row[origIdx]
			}
		}

		// Exempt columns marked NoRowStyle from row-level background
		cellRowStyle := rowStyle
		if col.NoRowStyle {
			cellRowStyle = nil
		}
		cell := t.alignCellStyled(cellContent, widths[i], col.Align, cellRowStyle)
		parts = append(parts, cell)
	}

	// Style separators and trailing padding with row background
	sep := " "
	if rowStyle != nil {
		sep = rowStyle.Render(" ")
	}
	line := strings.Join(parts, sep)

	// Pad to full table width with row style background
	if rowStyle != nil {
		vis := lipgloss.Width(line)
		if vis < t.width {
			line += rowStyle.Render(strings.Repeat(" ", t.width-vis))
		}
	}

	return line
}

func (t *Table) alignCell(content string, width int, align Alignment) string {
	return t.alignCellStyled(content, width, align, nil)
}

func (t *Table) alignCellStyled(content string, width int, align Alignment, rs *lipgloss.Style) string {
	contentWidth := lipgloss.Width(content)
	if contentWidth > width {
		// Truncate — but be careful with styled content
		if width > 1 {
			runes := []rune(content)
			if len(runes) > width-1 {
				content = string(runes[:width-1]) + "…"
			}
		} else if width > 0 {
			content = string([]rune(content)[:width])
		}
		contentWidth = lipgloss.Width(content)
	}

	gap := width - contentWidth
	if gap < 0 {
		gap = 0
	}

	pad := strings.Repeat(" ", gap)
	if rs != nil && gap > 0 {
		pad = rs.Render(pad)
	}

	// Apply row-level background to content so flash/cursor backgrounds
	// cover the full cell, not just the padding around styled text.
	if rs != nil {
		if bg := rs.GetBackground(); bg != (lipgloss.NoColor{}) {
			content = lipgloss.NewStyle().Background(bg).Render(content)
		}
	}

	switch align {
	case Right:
		return pad + content
	case Center:
		left := gap / 2
		right := gap - left
		lpad := strings.Repeat(" ", left)
		rpad := strings.Repeat(" ", right)
		if rs != nil {
			if left > 0 {
				lpad = rs.Render(lpad)
			}
			if right > 0 {
				rpad = rs.Render(rpad)
			}
		}
		return lpad + content + rpad
	default:
		return content + pad
	}
}

func (t *Table) visibleColumns() []Column {
	cols, _ := t.visibleColumnsWithIndices()
	return cols
}

// visibleColumnsWithIndices returns the subset of columns that fit in the
// current width alongside their original indices in t.columns. Using the
// indices (rather than looking up by Title) keeps row/header rendering
// correct when two columns share a title.
func (t *Table) visibleColumnsWithIndices() ([]Column, []int) {
	var cols []Column
	var idxs []int
	for i, col := range t.columns {
		if col.MinWidth == 0 || t.width >= col.MinWidth {
			cols = append(cols, col)
			idxs = append(idxs, i)
		}
	}
	return cols, idxs
}

func (t *Table) computeWidths(cols []Column) []int {
	if len(cols) == 0 {
		return nil
	}
	totalWeight := 0
	for _, c := range cols {
		totalWeight += c.Width
	}
	if totalWeight == 0 {
		totalWeight = 1
	}

	available := t.width - (len(cols) - 1) // subtract separators
	widths := make([]int, len(cols))
	used := 0
	for i, c := range cols {
		w := (c.Width * available) / totalWeight
		if c.MaxWidth > 0 {
			if w > c.MaxWidth {
				w = c.MaxWidth
			}
		} else if w < 4 {
			w = 4
		}
		widths[i] = w
		used += w
	}

	// Redistribute excess to the last uncapped column
	if diff := available - used; diff > 0 && len(widths) > 0 {
		for i := len(widths) - 1; i >= 0; i-- {
			if cols[i].MaxWidth == 0 {
				widths[i] += diff
				break
			}
			if i == 0 {
				widths[len(widths)-1] += diff
			}
		}
	}

	return widths
}

func (t *Table) rebuildVisible() {
	t.visible = nil
	for _, row := range t.rows {
		// Predicate filter
		if t.filterFunc != nil && !t.filterFunc(row) {
			continue
		}
		// Text search filter
		if t.filterQuery != "" {
			match := false
			query := strings.ToLower(t.filterQuery)
			for _, cell := range row {
				if strings.Contains(strings.ToLower(cell), query) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		t.visible = append(t.visible, row)
	}

	if t.sortCol >= 0 && t.sortCol < len(t.columns) {
		if t.opts.SortFunc != nil {
			// Custom sort
			sortCol := t.sortCol
			sortAsc := t.sortAsc
			sort.SliceStable(t.visible, func(i, j int) bool {
				return t.opts.SortFunc(t.visible[i], t.visible[j], sortCol, sortAsc)
			})
		} else {
			// Default lexicographic sort
			sort.SliceStable(t.visible, func(i, j int) bool {
				a, b := "", ""
				if t.sortCol < len(t.visible[i]) {
					a = t.visible[i][t.sortCol]
				}
				if t.sortCol < len(t.visible[j]) {
					b = t.visible[j][t.sortCol]
				}
				if t.sortAsc {
					return a < b
				}
				return a > b
			})
		}
	}
}

func (t *Table) cycleSort() {
	start := t.sortCol + 1
	for i := 0; i < len(t.columns); i++ {
		idx := (start + i) % len(t.columns)
		if t.columns[idx].Sortable {
			if idx == t.sortCol {
				t.sortAsc = !t.sortAsc
			} else {
				t.sortCol = idx
				t.sortAsc = true
			}
			t.rebuildVisible()
			t.clampCursor()
			return
		}
	}
}

func (t *Table) clampCursor() {
	prev := t.cursor
	if t.cursor < 0 {
		t.cursor = 0
	}
	maxCursor := len(t.visible) - 1
	if maxCursor < 0 {
		maxCursor = 0
	}
	if t.cursor > maxCursor {
		t.cursor = maxCursor
	}

	if t.cursor != prev && t.opts.OnCursorChange != nil && t.cursor < len(t.visible) {
		t.opts.OnCursorChange(t.visible[t.cursor], t.cursor)
	}

	visibleRows := t.height - 2
	if t.opts.DetailFunc != nil {
		visibleRows -= t.opts.DetailHeight
	}
	if visibleRows < 1 {
		visibleRows = 1
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+visibleRows {
		t.offset = t.cursor - visibleRows + 1
	}
}

func (t *Table) renderDetail() string {
	if t.focused {
		if row := t.CursorRow(); row != nil {
			detail := t.opts.DetailFunc(row, t.cursor, t.width, t.theme)
			if detail != "" {
				return detail
			}
		}
	}
	// Reserve blank space to prevent jitter
	blank := strings.Repeat("\n", t.opts.DetailHeight-1)
	return strings.Repeat(" ", t.width) + blank
}

func (t *Table) KeyBindings() []KeyBind {
	bindings := []KeyBind{
		{Key: "up/k", Label: "Move up", Group: "NAVIGATION"},
		{Key: "down/j", Label: "Move down", Group: "NAVIGATION"},
		{Key: "ctrl+u", Label: "Half page up", Group: "NAVIGATION"},
		{Key: "ctrl+d", Label: "Half page down", Group: "NAVIGATION"},
		{Key: "home/g", Label: "Go to top", Group: "NAVIGATION"},
		{Key: "end/G", Label: "Go to bottom", Group: "NAVIGATION"},
	}
	if t.opts.Sortable {
		bindings = append(bindings, KeyBind{Key: "s", Label: "Cycle sort", Group: "DATA"})
	}
	if t.opts.Filterable {
		if t.filtering {
			bindings = append(bindings,
				KeyBind{Key: "esc", Label: "Cancel search", Group: "SEARCH"},
				KeyBind{Key: "enter", Label: "Accept search", Group: "SEARCH"},
			)
		} else {
			bindings = append(bindings, KeyBind{Key: "/", Label: "Search", Group: "SEARCH"})
		}
	}
	if t.opts.OnRowClick != nil {
		bindings = append(bindings, KeyBind{Key: "enter", Label: "Select row", Group: "NAVIGATION"})
	}
	return bindings
}

func (t *Table) SetSize(w, h int) {
	t.width = w
	t.height = h
	t.clampCursor()
}

func (t *Table) Focused() bool    { return t.focused }
func (t *Table) SetFocused(f bool) { t.focused = f }
// SetTheme implements the Themed interface.
func (t *Table) SetTheme(th Theme) { t.theme = th }
