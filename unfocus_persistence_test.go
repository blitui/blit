package blit

import (
	"testing"
	"time"
)

// TestTablePersistenceOnUnfocus verifies that table cursor position and sort
// state survive an unfocus/refocus cycle.
func TestTablePersistenceOnUnfocus(t *testing.T) {
	cols := []Column{
		{Title: "Name", Width: 20, Sortable: true},
		{Title: "Score", Width: 10, Sortable: true},
	}
	rows := []Row{
		{"Alice", "100"},
		{"Bob", "200"},
		{"Charlie", "300"},
	}

	tests := []struct {
		name      string
		mutate    func(tbl *Table)
		checkPre  func(t *testing.T, tbl *Table) // verify state before unfocus
		checkPost func(t *testing.T, tbl *Table) // verify state after refocus
	}{
		{
			name: "cursor position preserved",
			mutate: func(tbl *Table) {
				tbl.SetCursor(2)
			},
			checkPre: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.CursorIndex() != 2 {
					t.Errorf("pre-unfocus: cursor = %d, want 2", tbl.CursorIndex())
				}
			},
			checkPost: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.CursorIndex() != 2 {
					t.Errorf("post-refocus: cursor = %d, want 2", tbl.CursorIndex())
				}
			},
		},
		{
			name: "sort state preserved",
			mutate: func(tbl *Table) {
				tbl.SetSort(1, false)
			},
			checkPre: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.SortCol() != 1 || tbl.SortAsc() != false {
					t.Errorf("pre-unfocus: sortCol=%d sortAsc=%v, want 1/false", tbl.SortCol(), tbl.SortAsc())
				}
			},
			checkPost: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.SortCol() != 1 || tbl.SortAsc() != false {
					t.Errorf("post-refocus: sortCol=%d sortAsc=%v, want 1/false", tbl.SortCol(), tbl.SortAsc())
				}
			},
		},
		{
			name: "cursor and sort combined",
			mutate: func(tbl *Table) {
				tbl.SetSort(0, true)
				tbl.SetCursor(1) // After sort, "Bob" is at index 1
			},
			checkPre: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.CursorIndex() != 1 {
					t.Errorf("pre-unfocus: cursor = %d, want 1", tbl.CursorIndex())
				}
				if tbl.SortCol() != 0 {
					t.Errorf("pre-unfocus: sortCol = %d, want 0", tbl.SortCol())
				}
			},
			checkPost: func(t *testing.T, tbl *Table) {
				t.Helper()
				if tbl.CursorIndex() != 1 {
					t.Errorf("post-refocus: cursor = %d, want 1", tbl.CursorIndex())
				}
				if tbl.SortCol() != 0 {
					t.Errorf("post-refocus: sortCol = %d, want 0", tbl.SortCol())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := NewTable(cols, rows, TableOpts{Sortable: true})
			tbl.SetSize(80, 24)
			tbl.SetFocused(true)

			tt.mutate(tbl)
			tt.checkPre(t, tbl)

			tbl.SetFocused(false)
			tt.checkPre(t, tbl) // state must survive unfocus

			tbl.SetFocused(true)
			tt.checkPost(t, tbl) // state must survive refocus
		})
	}
}

// TestListViewPersistenceOnUnfocus verifies that ListView cursor and scroll
// offset survive an unfocus/refocus cycle.
func TestListViewPersistenceOnUnfocus(t *testing.T) {
	type item struct {
		name string
	}
	makeListView := func() *ListView[item] {
		lv := NewListView[item](ListViewOpts[item]{
			RenderItem: func(it item, idx int, isCursor bool, th Theme) string {
				return it.name
			},
		})
		items := make([]item, 20)
		for i := range items {
			items[i] = item{name: "item-" + string(rune('A'+i))}
		}
		lv.SetSize(80, 10)
		lv.SetItems(items)
		lv.SetFocused(true)
		return lv
	}

	tests := []struct {
		name  string
		setup func(lv *ListView[item])
		check func(t *testing.T, lv *ListView[item])
	}{
		{
			name: "cursor position preserved",
			setup: func(lv *ListView[item]) {
				lv.SetCursor(5)
			},
			check: func(t *testing.T, lv *ListView[item]) {
				t.Helper()
				if lv.CursorIndex() != 5 {
					t.Errorf("cursor = %d, want 5", lv.CursorIndex())
				}
			},
		},
		{
			name: "selected item preserved",
			setup: func(lv *ListView[item]) {
				lv.SetCursor(7)
			},
			check: func(t *testing.T, lv *ListView[item]) {
				t.Helper()
				ci := lv.CursorItem()
				if ci == nil {
					t.Fatal("cursor item is nil")
				}
				if ci.name != "item-H" {
					t.Errorf("cursor item = %q, want %q", ci.name, "item-H")
				}
			},
		},
		{
			name: "scroll offset preserved",
			setup: func(lv *ListView[item]) {
				lv.ScrollToBottom()
			},
			check: func(t *testing.T, lv *ListView[item]) {
				t.Helper()
				if lv.CursorIndex() != 19 {
					t.Errorf("cursor = %d, want 19", lv.CursorIndex())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := makeListView()
			tt.setup(lv)
			tt.check(t, lv)

			lv.SetFocused(false)
			tt.check(t, lv) // survives unfocus

			lv.SetFocused(true)
			tt.check(t, lv) // survives refocus
		})
	}
}

// TestPickerPersistenceOnUnfocus verifies that Picker filter text and cursor
// survive an unfocus/refocus cycle.
func TestPickerPersistenceOnUnfocus(t *testing.T) {
	items := []PickerItem{
		{Title: "Alpha"},
		{Title: "Beta"},
		{Title: "Gamma"},
		{Title: "Delta"},
	}

	tests := []struct {
		name  string
		setup func(p *Picker)
		check func(t *testing.T, p *Picker)
	}{
		{
			name: "cursor position preserved",
			setup: func(p *Picker) {
				p.cursor = 2
			},
			check: func(t *testing.T, p *Picker) {
				t.Helper()
				if p.cursor != 2 {
					t.Errorf("cursor = %d, want 2", p.cursor)
				}
			},
		},
		{
			name: "filter text preserved",
			setup: func(p *Picker) {
				p.input.SetValue("alph")
				p.rebuildFiltered()
			},
			check: func(t *testing.T, p *Picker) {
				t.Helper()
				if p.input.Value() != "alph" {
					t.Errorf("filter = %q, want %q", p.input.Value(), "alph")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPicker(items, PickerOpts{})
			p.SetSize(80, 24)
			p.SetFocused(true)

			tt.setup(p)
			tt.check(t, p)

			p.SetFocused(false)
			tt.check(t, p) // survives unfocus

			p.SetFocused(true)
			tt.check(t, p) // survives refocus
		})
	}
}

// TestFormPersistenceOnUnfocus verifies that Form field values and focused field
// index survive an unfocus/refocus cycle.
func TestFormPersistenceOnUnfocus(t *testing.T) {
	tests := []struct {
		name  string
		setup func(f *Form)
		check func(t *testing.T, f *Form)
	}{
		{
			name: "field values preserved",
			setup: func(f *Form) {
				f.allFields[0].SetValue("hello")
				f.allFields[1].SetValue("world")
			},
			check: func(t *testing.T, f *Form) {
				t.Helper()
				vals := f.Values()
				if vals["name"] != "hello" {
					t.Errorf("name = %q, want %q", vals["name"], "hello")
				}
				if vals["email"] != "world" {
					t.Errorf("email = %q, want %q", vals["email"], "world")
				}
			},
		},
		{
			name: "focused field index preserved",
			setup: func(f *Form) {
				f.moveFocusTo(1)
			},
			check: func(t *testing.T, f *Form) {
				t.Helper()
				if f.focusIndex != 1 {
					t.Errorf("focusIndex = %d, want 1", f.focusIndex)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := NewForm(FormOpts{
				Groups: []FormGroup{
					{
						Title: "Info",
						Fields: []Field{
							NewTextField("name", "Name"),
							NewTextField("email", "Email"),
						},
					},
				},
			})
			form.SetSize(80, 24)
			form.SetFocused(true)

			tt.setup(form)
			tt.check(t, form)

			form.SetFocused(false)
			tt.check(t, form) // survives unfocus

			form.SetFocused(true)
			tt.check(t, form) // survives refocus
		})
	}
}

// TestTreePersistenceOnUnfocus verifies that Tree expanded nodes and cursor
// position survive an unfocus/refocus cycle.
func TestTreePersistenceOnUnfocus(t *testing.T) {
	makeRoots := func() []*Node {
		return []*Node{
			{
				Title: "root1",
				Children: []*Node{
					{Title: "child1a"},
					{Title: "child1b"},
				},
			},
			{
				Title: "root2",
				Children: []*Node{
					{Title: "child2a"},
				},
			},
		}
	}

	tests := []struct {
		name  string
		setup func(tr *Tree)
		check func(t *testing.T, tr *Tree)
	}{
		{
			name: "cursor position preserved",
			setup: func(tr *Tree) {
				tr.cursor = 1
			},
			check: func(t *testing.T, tr *Tree) {
				t.Helper()
				if tr.cursor != 1 {
					t.Errorf("cursor = %d, want 1", tr.cursor)
				}
			},
		},
		{
			name: "expanded nodes preserved",
			setup: func(tr *Tree) {
				tr.roots[0].Expanded = true
				tr.rebuild()
			},
			check: func(t *testing.T, tr *Tree) {
				t.Helper()
				if !tr.roots[0].Expanded {
					t.Error("root1 should still be expanded")
				}
				// Flat list should contain root1's children
				found := false
				for _, fn := range tr.flat {
					if fn.node.Title == "child1a" {
						found = true
						break
					}
				}
				if !found {
					t.Error("child1a should be in flat list after expand")
				}
			},
		},
		{
			name: "cursor and expanded combined",
			setup: func(tr *Tree) {
				tr.roots[0].Expanded = true
				tr.rebuild()
				tr.cursor = 2 // child1b
			},
			check: func(t *testing.T, tr *Tree) {
				t.Helper()
				if tr.cursor != 2 {
					t.Errorf("cursor = %d, want 2", tr.cursor)
				}
				node := tr.CursorNode()
				if node == nil || node.Title != "child1b" {
					title := "<nil>"
					if node != nil {
						title = node.Title
					}
					t.Errorf("cursor node = %q, want %q", title, "child1b")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots := makeRoots()
			tr := NewTree(roots, TreeOpts{})
			tr.SetSize(80, 24)
			tr.SetFocused(true)

			tt.setup(tr)
			tt.check(t, tr)

			tr.SetFocused(false)
			tt.check(t, tr) // survives unfocus

			tr.SetFocused(true)
			tt.check(t, tr) // survives refocus
		})
	}
}

// TestLogViewerPersistenceOnUnfocus verifies that LogViewer scroll position
// and pause state survive an unfocus/refocus cycle.
func TestLogViewerPersistenceOnUnfocus(t *testing.T) {
	now := time.Now()
	makeLogViewer := func() *LogViewer {
		lv := NewLogViewer()
		lv.SetSize(80, 10)
		lv.SetFocused(true)
		for i := 0; i < 50; i++ {
			lv.Append(LogLine{
				Level:     LogInfo,
				Timestamp: now.Add(time.Duration(i) * time.Second),
				Message:   "line " + string(rune('0'+i%10)),
				Source:    "test",
			})
		}
		return lv
	}

	tests := []struct {
		name  string
		setup func(lv *LogViewer)
		check func(t *testing.T, lv *LogViewer)
	}{
		{
			name: "paused state preserved",
			setup: func(lv *LogViewer) {
				lv.paused = true
				lv.userScrolled = true
			},
			check: func(t *testing.T, lv *LogViewer) {
				t.Helper()
				if !lv.paused {
					t.Error("expected paused to be true")
				}
			},
		},
		{
			name: "filter level preserved",
			setup: func(lv *LogViewer) {
				lv.filterLevel = LogWarn
				lv.rebuildFiltered()
			},
			check: func(t *testing.T, lv *LogViewer) {
				t.Helper()
				if lv.filterLevel != LogWarn {
					t.Errorf("filterLevel = %d, want %d", lv.filterLevel, LogWarn)
				}
			},
		},
		{
			name: "filter text preserved",
			setup: func(lv *LogViewer) {
				lv.filterText = "test"
				lv.rebuildFiltered()
			},
			check: func(t *testing.T, lv *LogViewer) {
				t.Helper()
				if lv.filterText != "test" {
					t.Errorf("filterText = %q, want %q", lv.filterText, "test")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := makeLogViewer()
			tt.setup(lv)
			tt.check(t, lv)

			lv.SetFocused(false)
			tt.check(t, lv) // survives unfocus

			lv.SetFocused(true)
			tt.check(t, lv) // survives refocus
		})
	}
}

// TestViewportPersistenceOnUnfocus verifies that Viewport scroll offset
// survives an unfocus/refocus cycle.
func TestViewportPersistenceOnUnfocus(t *testing.T) {
	makeContent := func() string {
		lines := ""
		for i := 0; i < 100; i++ {
			if i > 0 {
				lines += "\n"
			}
			lines += "line " + string(rune('0'+i%10))
		}
		return lines
	}

	tests := []struct {
		name  string
		setup func(vp *Viewport)
		check func(t *testing.T, vp *Viewport)
	}{
		{
			name: "scroll offset preserved",
			setup: func(vp *Viewport) {
				vp.ScrollBy(10)
			},
			check: func(t *testing.T, vp *Viewport) {
				t.Helper()
				if vp.YOffset() != 10 {
					t.Errorf("yOffset = %d, want 10", vp.YOffset())
				}
			},
		},
		{
			name: "scroll to bottom preserved",
			setup: func(vp *Viewport) {
				vp.GotoBottom()
			},
			check: func(t *testing.T, vp *Viewport) {
				t.Helper()
				if !vp.AtBottom() {
					t.Error("expected viewport to be at bottom")
				}
				if vp.YOffset() == 0 {
					t.Error("expected non-zero offset at bottom")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := NewViewport()
			vp.SetSize(80, 10)
			vp.SetContent(makeContent())
			vp.SetFocused(true)

			tt.setup(vp)
			tt.check(t, vp)

			vp.SetFocused(false)
			tt.check(t, vp) // survives unfocus

			vp.SetFocused(true)
			tt.check(t, vp) // survives refocus
		})
	}
}

// TestTabsPersistenceOnUnfocus verifies that Tabs active tab index survives
// an unfocus/refocus cycle.
func TestTabsPersistenceOnUnfocus(t *testing.T) {
	tests := []struct {
		name  string
		setup func(tabs *Tabs)
		check func(t *testing.T, tabs *Tabs)
	}{
		{
			name: "active tab preserved",
			setup: func(tabs *Tabs) {
				tabs.SetActive(2)
			},
			check: func(t *testing.T, tabs *Tabs) {
				t.Helper()
				if tabs.ActiveIndex() != 2 {
					t.Errorf("active = %d, want 2", tabs.ActiveIndex())
				}
			},
		},
		{
			name: "first tab stays at zero",
			setup: func(tabs *Tabs) {
				// Don't change, verify default stays
			},
			check: func(t *testing.T, tabs *Tabs) {
				t.Helper()
				if tabs.ActiveIndex() != 0 {
					t.Errorf("active = %d, want 0", tabs.ActiveIndex())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs := NewTabs([]TabItem{
				{Title: "Tab1"},
				{Title: "Tab2"},
				{Title: "Tab3"},
			}, TabsOpts{})
			tabs.SetSize(80, 24)
			tabs.SetFocused(true)

			tt.setup(tabs)
			tt.check(t, tabs)

			tabs.SetFocused(false)
			tt.check(t, tabs) // survives unfocus

			tabs.SetFocused(true)
			tt.check(t, tabs) // survives refocus
		})
	}
}

// TestFilePickerPersistenceOnUnfocus verifies that FilePicker tree cursor
// position survives an unfocus/refocus cycle.
func TestFilePickerPersistenceOnUnfocus(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fp *FilePicker)
		check func(t *testing.T, fp *FilePicker)
	}{
		{
			name: "tree cursor preserved",
			setup: func(fp *FilePicker) {
				// Move the internal tree cursor
				if len(fp.tree.flat) > 2 {
					fp.tree.cursor = 2
				}
			},
			check: func(t *testing.T, fp *FilePicker) {
				t.Helper()
				if len(fp.tree.flat) > 2 && fp.tree.cursor != 2 {
					t.Errorf("tree cursor = %d, want 2", fp.tree.cursor)
				}
			},
		},
		{
			name: "current directory preserved",
			setup: func(fp *FilePicker) {
				// The root stays as configured
			},
			check: func(t *testing.T, fp *FilePicker) {
				t.Helper()
				if fp.opts.Root == "" {
					t.Error("root directory should not be empty")
				}
				roots := fp.tree.Roots()
				if len(roots) == 0 {
					t.Error("tree should have roots")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFilePicker(FilePickerOpts{Root: "."})
			fp.SetSize(80, 24)
			fp.SetFocused(true)

			tt.setup(fp)
			tt.check(t, fp)

			fp.SetFocused(false)
			tt.check(t, fp) // survives unfocus

			fp.SetFocused(true)
			tt.check(t, fp) // survives refocus
		})
	}
}
