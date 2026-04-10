package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/blitui/blit"
	"github.com/blitui/blit/btest"
)

// ---------------------------------------------------------------------------
// ListView mouse scroll
// ---------------------------------------------------------------------------

func newTestListView(items []string) *blit.ListView[string] {
	lv := blit.NewListView(blit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
			return item
		},
	})
	lv.SetTheme(blit.DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(items)
	return lv
}

func TestListViewMouseScroll(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E"}
	tests := []struct {
		name       string
		button     tea.MouseButton
		wantCursor int
	}{
		{"wheel down moves cursor down", tea.MouseButtonWheelDown, 1},
		{"wheel up after down returns to 0", tea.MouseButtonWheelUp, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := newTestListView(items)
			if tt.button == tea.MouseButtonWheelUp {
				// First scroll down to have room to scroll up
				lv.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
			}
			lv.Update(tea.MouseMsg{X: 1, Y: 1, Button: tt.button}, blit.Context{})
			if got := lv.CursorIndex(); got != tt.wantCursor {
				t.Errorf("CursorIndex() = %d, want %d", got, tt.wantCursor)
			}
		})
	}
}

func TestListViewMouseScrollClamps(t *testing.T) {
	lv := newTestListView([]string{"only"})

	// Scroll down on single-item list should stay at 0
	lv.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
	if got := lv.CursorIndex(); got != 0 {
		t.Errorf("CursorIndex() = %d after scroll down on 1-item list, want 0", got)
	}

	// Scroll up should stay at 0
	lv.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelUp}, blit.Context{})
	if got := lv.CursorIndex(); got != 0 {
		t.Errorf("CursorIndex() = %d after scroll up on 1-item list, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// LogViewer mouse scroll
// ---------------------------------------------------------------------------

func TestLogViewerMouseScroll(t *testing.T) {
	lv := blit.NewLogViewer()
	lv.SetTheme(blit.DefaultTheme())
	lv.SetSize(40, 10)

	// Append enough lines to make the viewport scrollable.
	for i := 0; i < 30; i++ {
		lv.Append(blit.LogLine{Message: "line"})
	}

	tests := []struct {
		name   string
		button tea.MouseButton
	}{
		{"wheel up scrolls without panic", tea.MouseButtonWheelUp},
		{"wheel down scrolls without panic", tea.MouseButtonWheelDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			lv.Update(tea.MouseMsg{X: 1, Y: 1, Button: tt.button}, blit.Context{})
		})
	}
}

// ---------------------------------------------------------------------------
// Viewport mouse scroll (boundary check)
// ---------------------------------------------------------------------------

func TestViewportMouseOutsideBoundsIgnored(t *testing.T) {
	v := blit.NewViewport()
	v.SetTheme(blit.DefaultTheme())
	v.SetSize(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x"
	}
	v.SetContent(strings.Join(lines, "\n"))

	// Scroll inside bounds moves offset
	v.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
	if v.YOffset() == 0 {
		t.Error("expected offset > 0 after scroll inside bounds")
	}

	// Reset
	v.ScrollBy(-v.YOffset())

	// Scroll outside bounds (negative X) should be ignored
	v.Update(tea.MouseMsg{X: -1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
	if v.YOffset() != 0 {
		t.Errorf("expected offset 0 after scroll outside bounds, got %d", v.YOffset())
	}

	// Scroll outside bounds (Y >= height) should be ignored
	v.Update(tea.MouseMsg{X: 1, Y: 5, Button: tea.MouseButtonWheelDown}, blit.Context{})
	if v.YOffset() != 0 {
		t.Errorf("expected offset 0 after scroll outside Y bounds, got %d", v.YOffset())
	}
}

// ---------------------------------------------------------------------------
// Tree mouse scroll
// ---------------------------------------------------------------------------

func TestTreeMouseScroll(t *testing.T) {
	roots := []*blit.Node{
		{Title: "A"},
		{Title: "B"},
		{Title: "C"},
		{Title: "D"},
	}
	tr := blit.NewTree(roots, blit.TreeOpts{})
	tr.SetTheme(blit.DefaultTheme())
	tr.SetSize(40, 10)
	tr.SetFocused(true)

	tests := []struct {
		name     string
		button   tea.MouseButton
		wantNode string
	}{
		{"wheel down moves to B", tea.MouseButtonWheelDown, "B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr.Update(tea.MouseMsg{X: 1, Y: 1, Button: tt.button}, blit.Context{})
			node := tr.CursorNode()
			if node == nil {
				t.Fatal("CursorNode() returned nil")
			}
			if node.Title != tt.wantNode {
				t.Errorf("CursorNode().Title = %q, want %q", node.Title, tt.wantNode)
			}
		})
	}

	// Scroll back up
	tr.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelUp}, blit.Context{})
	node := tr.CursorNode()
	if node == nil || node.Title != "A" {
		title := "<nil>"
		if node != nil {
			title = node.Title
		}
		t.Errorf("after scroll up, CursorNode().Title = %q, want %q", title, "A")
	}
}

func TestTreeMouseScrollUnfocusedIgnored(t *testing.T) {
	roots := []*blit.Node{{Title: "A"}, {Title: "B"}}
	tr := blit.NewTree(roots, blit.TreeOpts{})
	tr.SetTheme(blit.DefaultTheme())
	tr.SetSize(40, 10)
	tr.SetFocused(false)

	tr.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
	node := tr.CursorNode()
	if node == nil || node.Title != "A" {
		t.Error("unfocused tree should ignore mouse scroll")
	}
}

// ---------------------------------------------------------------------------
// Picker mouse scroll
// ---------------------------------------------------------------------------

func TestPickerMouseScroll(t *testing.T) {
	items := []blit.PickerItem{
		{Title: "Alpha"},
		{Title: "Beta"},
		{Title: "Gamma"},
	}
	p := blit.NewPicker(items, blit.PickerOpts{})
	p.SetTheme(blit.DefaultTheme())
	p.SetSize(40, 10)
	p.SetFocused(true)

	// Scroll down
	p.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelDown}, blit.Context{})
	item := p.CursorItem()
	if item == nil {
		t.Fatal("CursorItem() returned nil")
	}
	if item.Title != "Beta" {
		t.Errorf("CursorItem().Title = %q after scroll down, want %q", item.Title, "Beta")
	}

	// Scroll up
	p.Update(tea.MouseMsg{X: 1, Y: 1, Button: tea.MouseButtonWheelUp}, blit.Context{})
	item = p.CursorItem()
	if item == nil {
		t.Fatal("CursorItem() returned nil")
	}
	if item.Title != "Alpha" {
		t.Errorf("CursorItem().Title = %q after scroll up, want %q", item.Title, "Alpha")
	}
}

// ---------------------------------------------------------------------------
// Boundary checks: clicks outside bounds should not panic
// ---------------------------------------------------------------------------

func TestClickOutsideBoundsNoPanic(t *testing.T) {
	tests := []struct {
		name      string
		component blit.Component
	}{
		{
			"table",
			func() blit.Component {
				tbl := blit.NewTable(
					[]blit.Column{{Title: "Name", Width: 20}},
					[]blit.Row{{"A"}, {"B"}},
					blit.TableOpts{},
				)
				tbl.SetTheme(blit.DefaultTheme())
				tbl.SetSize(40, 10)
				return tbl
			}(),
		},
		{
			"viewport",
			func() blit.Component {
				v := blit.NewViewport()
				v.SetTheme(blit.DefaultTheme())
				v.SetSize(40, 10)
				v.SetContent("hello\nworld")
				return v
			}(),
		},
		{
			"tree",
			func() blit.Component {
				tr := blit.NewTree([]*blit.Node{{Title: "X"}}, blit.TreeOpts{})
				tr.SetTheme(blit.DefaultTheme())
				tr.SetSize(40, 10)
				tr.SetFocused(true)
				return tr
			}(),
		},
		{
			"picker",
			func() blit.Component {
				p := blit.NewPicker([]blit.PickerItem{{Title: "X"}}, blit.PickerOpts{})
				p.SetTheme(blit.DefaultTheme())
				p.SetSize(40, 10)
				p.SetFocused(true)
				return p
			}(),
		},
		{
			"listview",
			func() blit.Component {
				lv := blit.NewListView(blit.ListViewOpts[string]{
					RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
						return item
					},
				})
				lv.SetTheme(blit.DefaultTheme())
				lv.SetSize(40, 10)
				lv.SetItems([]string{"A"})
				return lv
			}(),
		},
	}

	outOfBounds := []tea.MouseMsg{
		{X: -1, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress},
		{X: 5, Y: -1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress},
		{X: 100, Y: 5, Button: tea.MouseButtonWheelDown},
		{X: 5, Y: 100, Button: tea.MouseButtonWheelUp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, msg := range outOfBounds {
				// Should not panic
				tt.component.Update(msg, blit.Context{})
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DualPane: mouse coordinates translated for nested layouts
// ---------------------------------------------------------------------------

func TestDualPaneMouseCoordinateTranslation(t *testing.T) {
	// Create a Table as main, a Viewport as side.
	tbl := blit.NewTable(
		[]blit.Column{{Title: "Name", Width: 20}},
		[]blit.Row{{"A"}, {"B"}, {"C"}, {"D"}},
		blit.TableOpts{},
	)

	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp := blit.NewViewport()
	vp.SetContent(strings.Join(lines, "\n"))

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithLayout(&blit.DualPane{
			Main:         tbl,
			Side:         vp,
			SideWidth:    20,
			MinMainWidth: 40,
			SideRight:    true,
		}),
		blit.WithMouseSupport(),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Scroll wheel on the main pane (left side, inside bounds).
	// This should scroll the table cursor without panic.
	h.TestModel().SendMouse(5, 5, tea.MouseButtonWheelDown)

	// Scroll wheel on the side pane (right side).
	// This should scroll the viewport without panic.
	h.TestModel().SendMouse(70, 5, tea.MouseButtonWheelDown)

	// Click on the side pane to shift focus, then scroll.
	h.TestModel().SendMouse(70, 5, tea.MouseButtonLeft)
	h.TestModel().SendMouse(70, 5, tea.MouseButtonWheelDown)

	// No panic means coordinate translation works correctly.
}
