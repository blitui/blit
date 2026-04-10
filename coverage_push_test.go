package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Table tests for 0% and low-coverage functions ---

func TestTable_CycleSort(t *testing.T) {
	cols := []Column{
		{Title: "Name", Sortable: true},
		{Title: "Age", Sortable: true},
	}
	rows := []Row{
		Row{"Alice", "30"},
		Row{"Bob", "25"},
	}
	tbl := NewTable(cols, rows, TableOpts{Sortable: true})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	tbl.cycleSort()
	if tbl.sortCol != 0 {
		t.Errorf("expected sortCol 0, got %d", tbl.sortCol)
	}

	tbl.cycleSort()
	if tbl.sortCol != 1 {
		t.Errorf("expected sortCol 1 after second cycle, got %d", tbl.sortCol)
	}
}

func TestTable_RenderDetail(t *testing.T) {
	cols := []Column{{Title: "Name"}}
	rows := []Row{Row{"Alice"}}
	tbl := NewTable(cols, rows, TableOpts{
		DetailFunc: func(row Row, idx, width int, theme Theme) string {
			return "detail: " + row[0]
		},
		DetailHeight: 3,
	})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	detail := tbl.renderDetail()
	if detail == "" {
		t.Error("expected non-empty detail")
	}
}

func TestTable_HandleKeyFilter(t *testing.T) {
	cols := []Column{{Title: "Name"}}
	rows := []Row{
		Row{"Alice"},
		Row{"Bob"},
	}
	tbl := NewTable(cols, rows, TableOpts{Filterable: true})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	// Enter filter mode
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !tbl.filtering {
		t.Error("expected filtering mode")
	}

	// Type a character
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Backspace
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	// Escape
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if tbl.filtering {
		t.Error("expected filtering mode off after esc")
	}
}

func TestTable_HandleKeyNavigation(t *testing.T) {
	cols := []Column{{Title: "Name"}}
	rows := []Row{
		Row{"Alice"},
		Row{"Bob"},
		Row{"Charlie"},
	}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	// Move down
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if tbl.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", tbl.cursor)
	}

	// Move up
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if tbl.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", tbl.cursor)
	}

	// Home
	tbl.cursor = 2
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if tbl.cursor != 0 {
		t.Errorf("expected cursor 0 after 'g', got %d", tbl.cursor)
	}

	// End
	tbl.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if tbl.cursor != 2 {
		t.Errorf("expected cursor 2 after 'G', got %d", tbl.cursor)
	}
}

func TestTable_HandleMouse(t *testing.T) {
	cols := []Column{{Title: "Name"}}
	rows := []Row{
		Row{"Alice"},
		Row{"Bob"},
	}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	// Scroll down
	tbl.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelDown, X: 5, Y: 5})
	if tbl.cursor != 1 {
		t.Errorf("expected cursor 1 after scroll down, got %d", tbl.cursor)
	}

	// Scroll up
	tbl.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp, X: 5, Y: 5})
	if tbl.cursor != 0 {
		t.Errorf("expected cursor 0 after scroll up, got %d", tbl.cursor)
	}

	// Click on a row
	tbl.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      5, Y: 2, // row 1 (Y=0 is header-ish, Y=1 is first data, Y=2 is second)
	})
}

func TestTable_HandleMouseOutOfBounds(t *testing.T) {
	cols := []Column{{Title: "Name"}}
	rows := []Row{Row{"Alice"}}
	tbl := NewTable(cols, rows, TableOpts{})
	tbl.SetTheme(DefaultTheme())
	tbl.SetSize(80, 24)

	// Out of bounds
	tbl.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, X: -1, Y: 5})
}

func TestTableRowProviderFunc_LenRows(t *testing.T) {
	p := TableRowProviderFunc{
		Total: 2,
		Fetch: func(offset, limit int) []Row {
			return []Row{Row{"a"}, Row{"b"}}
		},
	}
	if p.Len() != 2 {
		t.Errorf("expected 2, got %d", p.Len())
	}
	rows := p.Rows(0, 2)
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestTableRowProviderFunc_NilFetch(t *testing.T) {
	p := TableRowProviderFunc{Total: 0}
	rows := p.Rows(0, 10)
	if rows != nil {
		t.Error("expected nil rows with nil Fetch")
	}
}

// --- Split Update (25%) ---

func TestSplit_Update_KeyMsg(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Horizontal, 0.5, a, b)
	s.SetTheme(DefaultTheme())
	s.SetSize(80, 24)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}

	// Tab to switch focus
	s.Update(tea.KeyMsg{Type: tea.KeyTab}, ctx)
	if s.focusA {
		t.Error("expected focusA=false after tab")
	}

	// Regular key delegated to focused child
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, ctx)
}

func TestSplit_Update_MouseMsg(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Horizontal, 0.5, a, b)
	s.SetTheme(DefaultTheme())
	s.SetSize(80, 24)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}

	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, X: 5, Y: 5}, ctx)
}

func TestSplit_Update_Resize(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Horizontal, 0.5, a, b)
	s.Resizable = true
	s.SetTheme(DefaultTheme())
	s.SetSize(80, 24)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}

	// alt+right to resize
	s.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: true}, ctx)
}

func TestSplit_Update_VerticalResize(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Vertical, 0.5, a, b)
	s.Resizable = true
	s.SetTheme(DefaultTheme())
	s.SetSize(80, 24)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}

	s.Update(tea.KeyMsg{Type: tea.KeyDown, Alt: true}, ctx)
}

// --- openOverlay non-Activatable paths ---

func TestOpenOverlay_ConfigEditor(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.width = 80
	a.height = 24

	ce := NewConfigEditor(nil)
	a.openOverlay(ce)
	if !ce.active {
		t.Error("ConfigEditor should be activated")
	}
}

// --- Form field SetFocused (remaining 0% for Select, MultiSelect, Confirm) ---

func TestSelectField_SetFocused(t *testing.T) {
	f := NewSelectField("c", "Color", []string{"red"})
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestMultiSelectField_SetFocused(t *testing.T) {
	f := NewMultiSelectField("t", "Tags", []string{"a"})
	f.SetFocused(true)
	f.SetFocused(false)
}

// --- logviewer levelName (33%) ---

func TestLogViewer_LevelName(t *testing.T) {
	lv := NewLogViewer()
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogDebug, "debug+"},
		{LogInfo, "info+"},
		{LogWarn, "warn+"},
		{LogError, "error"},
	}
	for _, tt := range tests {
		got := lv.levelName(tt.level)
		if got != tt.want {
			t.Errorf("levelName(%d) = %q, want %q", tt.level, got, tt.want)
		}
	}
}

// --- filepicker Init, Focused ---

// --- toFlexItem (50%) ---

func TestToFlexItem_PlainComponent(t *testing.T) {
	c := &stubComponent{name: "plain"}
	fi := toFlexItem(c)
	if fi.c != c {
		t.Error("expected plain component")
	}
	// Plain components are treated as Flex{Grow:1}
	if fi.grow != 1 {
		t.Errorf("expected grow=1 for plain component, got %d", fi.grow)
	}
}

func TestToFlexItem_Sized(t *testing.T) {
	c := &stubComponent{name: "x"}
	fi := toFlexItem(Sized{W: 15, C: c})
	if fi.fixedSz != 15 {
		t.Errorf("expected fixedSz 15, got %d", fi.fixedSz)
	}
}

func TestToFlexItem_Flex(t *testing.T) {
	c := &stubComponent{name: "x"}
	fi := toFlexItem(Flex{Grow: 3, C: c})
	if fi.grow != 3 {
		t.Errorf("expected grow 3, got %d", fi.grow)
	}
}

func TestToFlexItem_PointerSized(t *testing.T) {
	c := &stubComponent{name: "x"}
	fi := toFlexItem(&Sized{W: 10, C: c})
	if fi.fixedSz != 10 {
		t.Errorf("expected fixedSz 10, got %d", fi.fixedSz)
	}
}

func TestToFlexItem_PointerFlex(t *testing.T) {
	c := &stubComponent{name: "x"}
	fi := toFlexItem(&Flex{Grow: 2, C: c})
	if fi.grow != 2 {
		t.Errorf("expected grow 2, got %d", fi.grow)
	}
}

// --- filepicker ---

func TestFilePicker_Init(t *testing.T) {
	fp := NewFilePicker(FilePickerOpts{Root: "."})
	cmd := fp.Init()
	_ = cmd
}

func TestFilePicker_Focused(t *testing.T) {
	fp := NewFilePicker(FilePickerOpts{Root: "."})
	_ = fp.Focused()
}
