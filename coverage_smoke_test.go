package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestCoverageSmoke exercises boilerplate getters, setters, Init, and
// Close methods across built-in components. These are one-liners that
// add no real assertion value individually, but bringing them under
// test protects against silent regressions (e.g., a field rename
// breaking Focused/SetFocused) and keeps coverage meaningful.
func TestCoverageSmoke(t *testing.T) {
	// CommandBar
	cb := NewCommandBar([]Command{{Name: "quit"}})
	if cb.Init() != nil {
		t.Error("commandbar Init should be nil")
	}
	cb.SetFocused(true)
	if !cb.Focused() {
		t.Error("commandbar Focused")
	}
	cb.SetFocused(false)
	cb.active = true
	if !cb.IsActive() {
		t.Error("commandbar IsActive")
	}
	cb.Close()
	if cb.IsActive() {
		t.Error("commandbar Close should deactivate")
	}
	if !cb.Inline() {
		t.Error("commandbar Inline should be true")
	}
	_ = cb.KeyBindings()

	// Help
	h := NewHelp()
	if h.Init() != nil {
		t.Error("help Init should be nil")
	}
	h.active = true
	if !h.IsActive() {
		t.Error("help IsActive")
	}
	h.SetFocused(true)
	if !h.Focused() {
		t.Error("help Focused")
	}
	_, _ = h.Update(tea.KeyMsg{Type: tea.KeyEsc}, Context{})
	_ = h.KeyBindings()
	h.Close()
	if h.IsActive() {
		t.Error("help Close")
	}

	// StatusBar
	sb := NewStatusBar(StatusBarOpts{
		Left:  func() string { return "left" },
		Right: func() string { return "right" },
	})
	if sb.Init() != nil {
		t.Error("statusbar Init should be nil")
	}
	sb.SetSize(80, 1)
	_, _ = sb.Update(nil, Context{})
	_ = sb.View()
	_ = sb.KeyBindings()
	sb.SetFocused(true)
	if !sb.Focused() {
		t.Error("statusbar Focused")
	}

	// ConfigEditor
	value := "abc"
	ce := NewConfigEditor([]ConfigField{{
		Label: "Name",
		Group: "General",
		Get:   func() string { return value },
		Set:   func(s string) error { value = s; return nil },
	}})
	if ce.Init() != nil {
		t.Error("configeditor Init should be nil")
	}
	ce.active = true
	if !ce.IsActive() {
		t.Error("configeditor IsActive")
	}
	ce.SetFocused(true)
	if !ce.Focused() {
		t.Error("configeditor Focused")
	}
	_ = ce.KeyBindings()
	ce.Close()
	if ce.IsActive() {
		t.Error("configeditor Close")
	}

	// DetailOverlay
	do := NewDetailOverlay(DetailOverlayOpts[string]{
		Title: "Detail",
		Render: func(item string, w, h int, theme Theme) string {
			return item
		},
	})
	if do.Init() != nil {
		t.Error("detailoverlay Init should be nil")
	}
	do.Show("hello")
	if do.Item() != "hello" {
		t.Error("detailoverlay Item")
	}
	if !do.IsActive() {
		t.Error("detailoverlay should be active after Show")
	}
	do.SetFocused(true)
	if !do.Focused() {
		t.Error("detailoverlay Focused")
	}
	_ = do.KeyBindings()
	do.Close()
	if do.IsActive() {
		t.Error("detailoverlay Close")
	}
}

// TestTableGetters covers Table one-liner getters/setters.
func TestTableGetters(t *testing.T) {
	cols := []Column{{Title: "A", Width: 5}, {Title: "B", Width: 5}}
	rows := []Row{{"a1", "b1"}, {"a2", "b2"}, {"a3", "b3"}}
	tbl := NewTable(cols, rows, TableOpts{})
	if tbl.Init() != nil {
		t.Error("table Init should be nil")
	}

	if tbl.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", tbl.RowCount())
	}
	if tbl.VisibleRowCount() != 3 {
		t.Errorf("VisibleRowCount = %d, want 3", tbl.VisibleRowCount())
	}
	if tbl.CursorIndex() != 0 {
		t.Errorf("CursorIndex = %d, want 0", tbl.CursorIndex())
	}

	tbl.SetCursor(2)
	if tbl.CursorIndex() != 2 {
		t.Errorf("after SetCursor(2) CursorIndex = %d, want 2", tbl.CursorIndex())
	}

	tbl.SetSort(0, true)
	if tbl.SortCol() != 0 || !tbl.SortAsc() {
		t.Errorf("SetSort(0,true): col=%d asc=%v", tbl.SortCol(), tbl.SortAsc())
	}
	tbl.SetSort(-1, false)
	if tbl.SortCol() != -1 {
		t.Errorf("clear sort: col=%d", tbl.SortCol())
	}

	_ = tbl.KeyBindings()
	tbl.SetFocused(true)
	if !tbl.Focused() {
		t.Error("table Focused")
	}
}

// TestUtils covers utility helpers.
func TestUtils(t *testing.T) {
	if got := Hyperlink("https://x", "x"); got == "" {
		t.Error("Hyperlink empty")
	}
	if got := Divider(20, DefaultTheme()); got == "" {
		t.Error("Divider empty")
	}
	if got := Truncate("hello world", 5); len(got) == 0 || len([]rune(got)) > 5 {
		t.Errorf("Truncate: %q", got)
	}
	if got := Truncate("abc", 10); got != "abc" {
		t.Errorf("Truncate short: %q", got)
	}
	if got := Badge("NEW", lipgloss.Color("9"), true); got == "" {
		t.Error("Badge empty")
	}
}

// TestListViewIsAtBottom covers the one unexercised ListView getter.
func TestListViewIsAtBottom(t *testing.T) {
	lv := NewListView(ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme Theme) string {
			return item
		},
	})
	lv.SetItems([]string{"a", "b", "c"})
	lv.SetSize(10, 10)
	lv.SetCursor(2)
	_ = lv.IsAtBottom()
}
