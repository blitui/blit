package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
)

func makeTestKanban() *blit.Kanban {
	cols := []blit.KanbanColumn{
		{Title: "Todo", Cards: []blit.KanbanCard{
			{ID: "1", Title: "Task A"},
			{ID: "2", Title: "Task B", Description: "details"},
		}},
		{Title: "In Progress", Cards: []blit.KanbanCard{
			{ID: "3", Title: "Task C", Tag: "bug"},
		}},
		{Title: "Done", Cards: []blit.KanbanCard{}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(120, 30)
	k.SetFocused(true)
	return k
}

func TestKanban_NewDefaults(t *testing.T) {
	k := makeTestKanban()
	if k.ActiveColumn() != 0 {
		t.Fatalf("ActiveColumn() = %d, want 0", k.ActiveColumn())
	}
	if k.ActiveCard() != 0 {
		t.Fatalf("ActiveCard() = %d, want 0", k.ActiveCard())
	}
}

func TestKanban_View(t *testing.T) {
	k := makeTestKanban()
	view := k.View()
	if view == "" {
		t.Fatal("View() should not be empty")
	}
	if !strings.Contains(view, "Todo") {
		t.Fatalf("view should contain 'Todo':\n%s", view)
	}
	if !strings.Contains(view, "In Progress") {
		t.Fatalf("view should contain 'In Progress':\n%s", view)
	}
	if !strings.Contains(view, "Task A") {
		t.Fatalf("view should contain 'Task A':\n%s", view)
	}
}

func TestKanban_NavigateColumns(t *testing.T) {
	k := makeTestKanban()

	// Right to "In Progress".
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 1 {
		t.Fatalf("ActiveColumn() = %d, want 1", k.ActiveColumn())
	}

	// Right to "Done".
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 2 {
		t.Fatalf("ActiveColumn() = %d, want 2", k.ActiveColumn())
	}

	// Right at end stays.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 2 {
		t.Fatalf("ActiveColumn() = %d, want 2 (clamped)", k.ActiveColumn())
	}

	// Left back.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 1 {
		t.Fatalf("ActiveColumn() = %d, want 1", k.ActiveColumn())
	}
}

func TestKanban_NavigateCards(t *testing.T) {
	k := makeTestKanban()

	// Down to Task B.
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 1 {
		t.Fatalf("ActiveCard() = %d, want 1", k.ActiveCard())
	}

	// Down at end stays.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 1 {
		t.Fatalf("ActiveCard() = %d, want 1 (clamped)", k.ActiveCard())
	}

	// Up back.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 0 {
		t.Fatalf("ActiveCard() = %d, want 0", k.ActiveCard())
	}
}

func TestKanban_ViNavigation(t *testing.T) {
	k := makeTestKanban()

	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 1 {
		t.Fatalf("j: ActiveCard() = %d, want 1", k.ActiveCard())
	}

	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 0 {
		t.Fatalf("k: ActiveCard() = %d, want 0", k.ActiveCard())
	}

	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 1 {
		t.Fatalf("l: ActiveColumn() = %d, want 1", k.ActiveColumn())
	}

	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 0 {
		t.Fatalf("h: ActiveColumn() = %d, want 0", k.ActiveColumn())
	}
}

func TestKanban_MoveCardRight(t *testing.T) {
	var moved blit.KanbanCard
	var fromCol, toCol int
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "X"}, {ID: "2", Title: "Y"}}},
		{Title: "B", Cards: []blit.KanbanCard{}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{
		OnMove: func(card blit.KanbanCard, from, to int) {
			moved = card
			fromCol = from
			toCol = to
		},
	})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	// Move first card right (Shift+Right or L).
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}, blit.Context{})
	k = updated.(*blit.Kanban)

	if moved.ID != "1" {
		t.Fatalf("moved card ID = %q, want 1", moved.ID)
	}
	if fromCol != 0 || toCol != 1 {
		t.Fatalf("moved from %d to %d, want 0 to 1", fromCol, toCol)
	}
	if k.ActiveColumn() != 1 {
		t.Fatalf("ActiveColumn() = %d, want 1 (followed card)", k.ActiveColumn())
	}
	// Source column should have 1 card left.
	if len(k.Columns()[0].Cards) != 1 {
		t.Fatalf("source column cards = %d, want 1", len(k.Columns()[0].Cards))
	}
	// Dest column should have 1 card.
	if len(k.Columns()[1].Cards) != 1 {
		t.Fatalf("dest column cards = %d, want 1", len(k.Columns()[1].Cards))
	}
}

func TestKanban_MoveCardLeft(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{}},
		{Title: "B", Cards: []blit.KanbanCard{{ID: "1", Title: "X"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	// Move to column B.
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)

	// Move card left.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("H")}, blit.Context{})
	k = updated.(*blit.Kanban)

	if k.ActiveColumn() != 0 {
		t.Fatalf("ActiveColumn() = %d, want 0", k.ActiveColumn())
	}
	if len(k.Columns()[0].Cards) != 1 {
		t.Fatalf("column A cards = %d, want 1", len(k.Columns()[0].Cards))
	}
	if len(k.Columns()[1].Cards) != 0 {
		t.Fatalf("column B cards = %d, want 0", len(k.Columns()[1].Cards))
	}
}

func TestKanban_MoveCardBoundary(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "X"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	// Try moving right when there's no column to the right — should be no-op.
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}, blit.Context{})
	k = updated.(*blit.Kanban)
	if len(k.Columns()[0].Cards) != 1 {
		t.Fatal("card should not move when no destination column")
	}
}

func TestKanban_Select(t *testing.T) {
	var selected blit.KanbanCard
	var selectedCol int
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "X"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{
		OnSelect: func(card blit.KanbanCard, col int) {
			selected = card
			selectedCol = col
		},
	})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Kanban)
	if selected.ID != "1" {
		t.Fatalf("selected card ID = %q, want 1", selected.ID)
	}
	if selectedCol != 0 {
		t.Fatalf("selected col = %d, want 0", selectedCol)
	}
}

func TestKanban_EmptyColumns(t *testing.T) {
	k := blit.NewKanban([]blit.KanbanColumn{}, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	if k.View() != "" {
		t.Fatal("empty kanban should return empty view")
	}
}

func TestKanban_EmptyColumnView(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "Empty", Cards: []blit.KanbanCard{}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)

	view := k.View()
	if !strings.Contains(view, "empty") {
		t.Fatalf("empty column should show placeholder:\n%s", view)
	}
}

func TestKanban_ClampCardOnColumnSwitch(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "X"}, {ID: "2", Title: "Y"}, {ID: "3", Title: "Z"}}},
		{Title: "B", Cards: []blit.KanbanCard{{ID: "4", Title: "W"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	// Move to card Z (index 2).
	k.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 2 {
		t.Fatalf("ActiveCard() = %d, want 2", k.ActiveCard())
	}

	// Switch to column B (only 1 card) — card index should clamp.
	updated, _ = k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveCard() != 0 {
		t.Fatalf("ActiveCard() = %d, want 0 (clamped)", k.ActiveCard())
	}
}

func TestKanban_ZeroSize(t *testing.T) {
	k := makeTestKanban()
	k.SetSize(0, 0)
	if k.View() != "" {
		t.Fatal("zero-sized kanban should return empty view")
	}
}

func TestKanban_UnfocusedIgnoresInput(t *testing.T) {
	k := makeTestKanban()
	k.SetFocused(false)

	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	k = updated.(*blit.Kanban)
	if k.ActiveColumn() != 0 {
		t.Fatal("unfocused kanban should not process input")
	}
}

func TestKanban_KeyBindings(t *testing.T) {
	k := makeTestKanban()
	binds := k.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestKanban_ComponentInterface(t *testing.T) {
	k := makeTestKanban()
	if cmd := k.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	k.SetFocused(false)
	if k.Focused() {
		t.Fatal("should not be focused")
	}
	k.SetFocused(true)
	if !k.Focused() {
		t.Fatal("should be focused")
	}
}

func TestKanban_CardWithTag(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "Fix crash", Tag: "bug"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)

	view := k.View()
	if !strings.Contains(view, "bug") {
		t.Fatalf("view should contain tag:\n%s", view)
	}
}

func TestKanban_CardWithDescription(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{{ID: "1", Title: "Task", Description: "some details"}}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)

	view := k.View()
	if !strings.Contains(view, "some details") {
		t.Fatalf("view should contain description:\n%s", view)
	}
}

func TestKanban_MoveFromEmptyColumn(t *testing.T) {
	cols := []blit.KanbanColumn{
		{Title: "A", Cards: []blit.KanbanCard{}},
		{Title: "B", Cards: []blit.KanbanCard{}},
	}
	k := blit.NewKanban(cols, blit.KanbanOpts{})
	k.SetTheme(blit.DefaultTheme())
	k.SetSize(80, 20)
	k.SetFocused(true)

	// Moving from empty column should be a no-op.
	updated, _ := k.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}, blit.Context{})
	_ = updated.(*blit.Kanban)
	// No panic is sufficient.
}
