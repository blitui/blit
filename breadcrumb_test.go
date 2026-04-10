package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func TestBreadcrumb_NewDefaults(t *testing.T) {
	items := []blit.BreadcrumbItem{
		{Label: "Home"},
		{Label: "Users"},
		{Label: "Profile"},
	}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})
	if bc.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2 (last item)", bc.CursorIndex())
	}
	if len(bc.Items()) != 3 {
		t.Fatalf("Items() = %d, want 3", len(bc.Items()))
	}
}

func TestBreadcrumb_View(t *testing.T) {
	items := []blit.BreadcrumbItem{
		{Label: "Home"},
		{Label: "Settings"},
	}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)

	view := bc.View()
	if !strings.Contains(view, "Home") {
		t.Fatalf("view should contain 'Home':\n%s", view)
	}
	if !strings.Contains(view, "Settings") {
		t.Fatalf("view should contain 'Settings':\n%s", view)
	}
	if !strings.Contains(view, ">") {
		t.Fatalf("view should contain separator:\n%s", view)
	}
}

func TestBreadcrumb_CustomSeparator(t *testing.T) {
	items := []blit.BreadcrumbItem{{Label: "A"}, {Label: "B"}}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{Separator: " / "})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)

	view := bc.View()
	if !strings.Contains(view, "/") {
		t.Fatalf("view should contain custom separator:\n%s", view)
	}
}

func TestBreadcrumb_Navigate(t *testing.T) {
	items := []blit.BreadcrumbItem{
		{Label: "A"},
		{Label: "B"},
		{Label: "C"},
	}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)
	bc.SetFocused(true)

	// Starts at C (index 2). Move left.
	updated, _ := bc.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", bc.CursorIndex())
	}

	// Left again.
	updated, _ = bc.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 0 {
		t.Fatalf("CursorIndex() = %d, want 0", bc.CursorIndex())
	}

	// Left at start stays.
	updated, _ = bc.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 0 {
		t.Fatalf("CursorIndex() = %d, want 0 (clamped)", bc.CursorIndex())
	}

	// Right.
	updated, _ = bc.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", bc.CursorIndex())
	}
}

func TestBreadcrumb_ViNavigation(t *testing.T) {
	items := []blit.BreadcrumbItem{{Label: "A"}, {Label: "B"}}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)
	bc.SetFocused(true)

	// h moves left.
	updated, _ := bc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 0 {
		t.Fatalf("h: CursorIndex() = %d, want 0", bc.CursorIndex())
	}

	// l moves right.
	updated, _ = bc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 1 {
		t.Fatalf("l: CursorIndex() = %d, want 1", bc.CursorIndex())
	}
}

func TestBreadcrumb_Select(t *testing.T) {
	var selected blit.BreadcrumbItem
	var selectedIdx int
	items := []blit.BreadcrumbItem{{Label: "Home"}, {Label: "Profile"}}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{
		OnSelect: func(item blit.BreadcrumbItem, idx int) {
			selected = item
			selectedIdx = idx
		},
	})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)
	bc.SetFocused(true)

	// Move to Home and select.
	updated, _ := bc.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	updated, _ = bc.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Breadcrumb)

	if selected.Label != "Home" {
		t.Fatalf("selected = %q, want Home", selected.Label)
	}
	if selectedIdx != 0 {
		t.Fatalf("selectedIdx = %d, want 0", selectedIdx)
	}
}

func TestBreadcrumb_Push(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{{Label: "Home"}}, blit.BreadcrumbOpts{})
	bc.Push(blit.BreadcrumbItem{Label: "Settings"})

	if len(bc.Items()) != 2 {
		t.Fatalf("Items() = %d, want 2", len(bc.Items()))
	}
	if bc.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1 (pushed item)", bc.CursorIndex())
	}
}

func TestBreadcrumb_Pop(t *testing.T) {
	items := []blit.BreadcrumbItem{{Label: "A"}, {Label: "B"}, {Label: "C"}}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})

	popped := bc.Pop()
	if popped.Label != "C" {
		t.Fatalf("Pop() = %q, want C", popped.Label)
	}
	if len(bc.Items()) != 2 {
		t.Fatalf("Items() = %d, want 2", len(bc.Items()))
	}
	if bc.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", bc.CursorIndex())
	}

	// Pop all.
	bc.Pop()
	bc.Pop()
	if len(bc.Items()) != 0 {
		t.Fatalf("Items() = %d, want 0", len(bc.Items()))
	}

	// Pop empty returns zero.
	z := bc.Pop()
	if z.Label != "" {
		t.Fatalf("Pop() on empty should return zero, got %q", z.Label)
	}
}

func TestBreadcrumb_SetItems(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{{Label: "A"}}, blit.BreadcrumbOpts{})
	bc.SetItems([]blit.BreadcrumbItem{{Label: "X"}, {Label: "Y"}, {Label: "Z"}})

	if len(bc.Items()) != 3 {
		t.Fatalf("Items() = %d, want 3", len(bc.Items()))
	}
	if bc.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2 (last)", bc.CursorIndex())
	}
}

func TestBreadcrumb_Empty(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{}, blit.BreadcrumbOpts{})
	bc.SetTheme(blit.DefaultTheme())
	bc.SetSize(80, 1)
	if bc.View() != "" {
		t.Fatal("empty breadcrumb should return empty view")
	}
}

func TestBreadcrumb_UnfocusedIgnoresInput(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{{Label: "A"}, {Label: "B"}}, blit.BreadcrumbOpts{})
	bc.SetFocused(false)

	updated, _ := bc.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	bc = updated.(*blit.Breadcrumb)
	if bc.CursorIndex() != 1 {
		t.Fatal("unfocused breadcrumb should not process input")
	}
}

func TestBreadcrumb_KeyBindings(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{{Label: "A"}}, blit.BreadcrumbOpts{})
	binds := bc.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestBreadcrumb_ComponentInterface(t *testing.T) {
	bc := blit.NewBreadcrumb([]blit.BreadcrumbItem{{Label: "A"}}, blit.BreadcrumbOpts{})
	if cmd := bc.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	if bc.Focused() {
		t.Fatal("should not be focused by default")
	}
	bc.SetFocused(true)
	if !bc.Focused() {
		t.Fatal("should be focused")
	}
}

func TestBreadcrumb_DataField(t *testing.T) {
	items := []blit.BreadcrumbItem{
		{Label: "Home", Data: "/"},
		{Label: "Docs", Data: "/docs"},
	}
	bc := blit.NewBreadcrumb(items, blit.BreadcrumbOpts{})
	if bc.Items()[0].Data.(string) != "/" {
		t.Fatal("Data field should be accessible")
	}
}
