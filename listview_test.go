package tuikit

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type lvItem struct {
	id   int
	name string
}

func renderLVItem(it lvItem, idx int, isCursor bool, theme Theme) string {
	return it.name
}

func newTestListView(items []lvItem) *ListView[lvItem] {
	lv := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
	})
	lv.SetTheme(DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(items)
	return lv
}

func sampleItems(n int) []lvItem {
	out := make([]lvItem, n)
	for i := 0; i < n; i++ {
		out[i] = lvItem{id: i, name: "row-" + strings.Repeat("x", 1)}
		out[i].name = "row-" + itoa(i)
	}
	return out
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func TestListView_SetItemsAndItemCount(t *testing.T) {
	lv := newTestListView(sampleItems(5))
	if lv.ItemCount() != 5 {
		t.Errorf("ItemCount = %d, want 5", lv.ItemCount())
	}
	lv.SetItems(sampleItems(2))
	if lv.ItemCount() != 2 {
		t.Errorf("after SetItems(2), ItemCount = %d", lv.ItemCount())
	}
	if len(lv.Items()) != 2 {
		t.Errorf("Items len = %d, want 2", len(lv.Items()))
	}
}

func TestListView_CursorClamps(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	lv.SetCursor(100)
	if lv.CursorIndex() != 2 {
		t.Errorf("SetCursor(100) on 3 items, got cursor %d, want 2", lv.CursorIndex())
	}
	lv.SetCursor(-5)
	if lv.CursorIndex() != 0 {
		t.Errorf("SetCursor(-5), got cursor %d, want 0", lv.CursorIndex())
	}
}

func TestListView_CursorItem(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	lv.SetCursor(1)
	item := lv.CursorItem()
	if item == nil || item.id != 1 {
		t.Errorf("CursorItem = %+v, want id=1", item)
	}
	empty := newTestListView(nil)
	if empty.CursorItem() != nil {
		t.Error("empty list should return nil cursor item")
	}
}

func TestListView_HandleKey_Navigation(t *testing.T) {
	lv := newTestListView(sampleItems(5))
	lv.SetFocused(true)

	tests := []struct {
		name       string
		keys       []string
		wantCursor int
	}{
		{"down twice", []string{"down", "down"}, 2},
		{"j twice", []string{"j", "j"}, 2},
		{"end", []string{"end"}, 4},
		{"G", []string{"G"}, 4},
		{"down past end clamps", []string{"end", "down", "down"}, 4},
		{"up at top", []string{"up"}, 0},
		{"home from end", []string{"end", "home"}, 0},
		{"g from end", []string{"end", "g"}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fresh := newTestListView(sampleItems(5))
			fresh.SetFocused(true)
			for _, k := range tc.keys {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
				if special, ok := keyStr(k); ok {
					msg = tea.KeyMsg{Type: special}
				}
				fresh.HandleKey(msg)
			}
			if fresh.CursorIndex() != tc.wantCursor {
				t.Errorf("cursor = %d, want %d", fresh.CursorIndex(), tc.wantCursor)
			}
		})
	}
	_ = lv
}

func keyStr(k string) (tea.KeyType, bool) {
	switch k {
	case "up":
		return tea.KeyUp, true
	case "down":
		return tea.KeyDown, true
	case "home":
		return tea.KeyHome, true
	case "end":
		return tea.KeyEnd, true
	case "enter":
		return tea.KeyEnter, true
	}
	return 0, false
}

func TestListView_OnSelect(t *testing.T) {
	var selected lvItem
	var selectedIdx int
	lv := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
		OnSelect: func(item lvItem, idx int) {
			selected = item
			selectedIdx = idx
		},
	})
	lv.SetTheme(DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(sampleItems(3))
	lv.SetFocused(true)
	lv.SetCursor(2)

	cmd := lv.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return a consumed cmd")
	}
	if selected.id != 2 || selectedIdx != 2 {
		t.Errorf("OnSelect called with %+v idx=%d, want id=2 idx=2", selected, selectedIdx)
	}
}

func TestListView_EnterWithoutOnSelectReturnsNil(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	lv.SetFocused(true)
	cmd := lv.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter without OnSelect should return nil")
	}
}

func TestListView_ScrollToTopBottom(t *testing.T) {
	lv := newTestListView(sampleItems(20))
	lv.ScrollToBottom()
	if lv.CursorIndex() != 19 {
		t.Errorf("ScrollToBottom cursor = %d, want 19", lv.CursorIndex())
	}
	lv.ScrollToTop()
	if lv.CursorIndex() != 0 {
		t.Errorf("ScrollToTop cursor = %d, want 0", lv.CursorIndex())
	}
	if !lv.IsAtTop() {
		t.Error("IsAtTop should be true after ScrollToTop")
	}
}

func TestListView_EmptyState(t *testing.T) {
	lv := newTestListView(nil)
	if lv.ItemCount() != 0 {
		t.Error("empty list should have 0 items")
	}
	lv.SetCursor(0)
	if lv.CursorIndex() != 0 {
		t.Error("empty list cursor should be 0")
	}
	// View on empty should not panic.
	_ = lv.View()
}

func TestListView_ViewRendersItems(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	lv.SetFocused(true)
	view := lv.View()
	for _, want := range []string{"row-0", "row-1", "row-2"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q: %s", want, view)
		}
	}
}

func TestListView_HeaderFunc(t *testing.T) {
	lv := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
		HeaderFunc: func(theme Theme) string { return "HEADER-LINE" },
	})
	lv.SetTheme(DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(sampleItems(3))
	view := lv.View()
	if !strings.Contains(view, "HEADER-LINE") {
		t.Errorf("view missing header: %s", view)
	}
}

func TestListView_DetailFuncOnlyShownWhenFocused(t *testing.T) {
	detailCalled := 0
	lv := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
		DetailFunc: func(item lvItem, theme Theme) string {
			detailCalled++
			return "DETAIL-BAR"
		},
	})
	lv.SetTheme(DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(sampleItems(3))

	// Unfocused: DetailFunc not invoked (blank reserved).
	lv.SetFocused(false)
	view := lv.View()
	if strings.Contains(view, "DETAIL-BAR") {
		t.Error("detail should not render when unfocused")
	}

	// Focused: DetailFunc renders.
	lv.SetFocused(true)
	view = lv.View()
	if !strings.Contains(view, "DETAIL-BAR") {
		t.Errorf("focused view missing detail: %s", view)
	}
}

func TestListView_FlashFunc(t *testing.T) {
	calls := 0
	lv := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
		FlashFunc: func(item lvItem, now time.Time) bool {
			calls++
			return item.id == 1
		},
	})
	lv.SetTheme(DefaultTheme())
	lv.SetSize(40, 10)
	lv.SetItems(sampleItems(3))
	_ = lv.View()
	if calls == 0 {
		t.Error("FlashFunc was never called during render")
	}
}

func TestListView_SetFocused(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	if lv.Focused() {
		t.Error("new listview should not be focused")
	}
	lv.SetFocused(true)
	if !lv.Focused() {
		t.Error("SetFocused(true) should set focus")
	}
}

func TestListView_Refresh(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	// Should not panic.
	lv.Refresh()
}

func TestListView_UpdateDelegatesToHandleKey(t *testing.T) {
	lv := newTestListView(sampleItems(5))
	lv.SetFocused(true)
	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyDown}, Context{})
	if cmd == nil {
		t.Error("Update(down) should return consumed cmd")
	}
	if lv.CursorIndex() != 1 {
		t.Errorf("cursor = %d, want 1", lv.CursorIndex())
	}
	// Non-key msg is a no-op.
	_, cmd = lv.Update(tea.WindowSizeMsg{Width: 100, Height: 20}, Context{})
	if cmd != nil {
		t.Error("Update(WindowSizeMsg) should return nil cmd")
	}
}

func TestListView_KeyBindings(t *testing.T) {
	lv := newTestListView(sampleItems(3))
	binds := lv.KeyBindings()
	if len(binds) < 4 {
		t.Errorf("expected navigation keybinds, got %d", len(binds))
	}
	// OnSelect adds enter bind.
	lvSel := NewListView[lvItem](ListViewOpts[lvItem]{
		RenderItem: renderLVItem,
		OnSelect:   func(lvItem, int) {},
	})
	bindsSel := lvSel.KeyBindings()
	foundEnter := false
	for _, b := range bindsSel {
		if b.Key == "enter" {
			foundEnter = true
		}
	}
	if !foundEnter {
		t.Error("OnSelect should add enter keybind")
	}
}

func TestListView_SetItemsClampsCursorDown(t *testing.T) {
	lv := newTestListView(sampleItems(10))
	lv.SetCursor(8)
	lv.SetItems(sampleItems(3))
	if lv.CursorIndex() > 2 {
		t.Errorf("after shrink, cursor = %d, want <=2", lv.CursorIndex())
	}
}

func TestListView_InitReturnsNil(t *testing.T) {
	lv := newTestListView(sampleItems(1))
	if cmd := lv.Init(); cmd != nil {
		t.Error("Init should return nil")
	}
}

// B3: cursor tween tests

func TestListView_CursorTweenStartsOnMove(t *testing.T) {
	lv := newTestListView(sampleItems(5))
	lv.SetFocused(true)

	if lv.cursorTween.Running() {
		t.Fatal("tween should not be running before cursor moves")
	}

	lv.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if !lv.cursorTween.Running() {
		t.Error("tween should be running after cursor move down")
	}

	// Progress() advances the tween state and marks it done when elapsed >= Duration
	time.Sleep(130 * time.Millisecond)
	prog := lv.cursorTween.Progress(time.Now())
	if prog != 1.0 {
		t.Errorf("tween progress should be 1.0 after 130ms, got %f", prog)
	}
	if lv.cursorTween.Running() {
		t.Error("tween should have finished after 130ms")
	}
}

func TestListView_CursorTweenSnapOnNoAnim(t *testing.T) {
	animDisabled = true
	defer func() { animDisabled = false }()

	lv := newTestListView(sampleItems(3))
	lv.SetFocused(true)

	lv.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	if lv.cursorTween.Running() {
		tval := lv.cursorTween.Progress(time.Now())
		if tval != 1.0 {
			t.Errorf("expected tween progress 1.0 when animDisabled, got %f", tval)
		}
	}
}
