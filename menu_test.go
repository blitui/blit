package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMenu_NewDefaults(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "Open"},
			{Label: "Save"},
		},
	})
	if !m.IsActive() {
		t.Fatal("menu should be active after creation")
	}
	if m.CursorIndex() != 0 {
		t.Fatal("cursor should start at 0")
	}
}

func TestMenu_View(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "Cut"},
			{Label: "Copy"},
			{Label: "Paste"},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)

	view := m.View()
	if !strings.Contains(view, "Cut") {
		t.Fatalf("view should contain Cut:\n%s", view)
	}
	if !strings.Contains(view, "Copy") {
		t.Fatalf("view should contain Copy:\n%s", view)
	}
	if !strings.Contains(view, "Paste") {
		t.Fatalf("view should contain Paste:\n%s", view)
	}
}

func TestMenu_ViewWithShortcuts(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "Save", Shortcut: "ctrl+s"},
			{Label: "Quit", Shortcut: "ctrl+q"},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)

	view := m.View()
	if !strings.Contains(view, "ctrl+s") {
		t.Fatalf("view should contain shortcut:\n%s", view)
	}
}

func TestMenu_Navigate(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A"},
			{Label: "B"},
			{Label: "C"},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	// Move down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 1 {
		t.Fatalf("cursor = %d, want 1", m.CursorIndex())
	}

	// Move down again.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 2 {
		t.Fatalf("cursor = %d, want 2", m.CursorIndex())
	}

	// Down at end doesn't wrap.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 2 {
		t.Fatalf("cursor = %d, want 2 (clamped)", m.CursorIndex())
	}

	// Move up.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 1 {
		t.Fatalf("cursor = %d, want 1", m.CursorIndex())
	}
}

func TestMenu_ViNavigation(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{{Label: "A"}, {Label: "B"}},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 1 {
		t.Fatalf("j: cursor = %d, want 1", m.CursorIndex())
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 0 {
		t.Fatalf("k: cursor = %d, want 0", m.CursorIndex())
	}
}

func TestMenu_SkipsSeparators(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A"},
			{Separator: true},
			{Label: "B"},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	// Starts at A (0).
	if m.CursorIndex() != 0 {
		t.Fatalf("cursor = %d, want 0", m.CursorIndex())
	}

	// Down should skip separator and land on B (2).
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 2 {
		t.Fatalf("cursor = %d, want 2 (skipping separator)", m.CursorIndex())
	}

	// Up should skip separator and land on A (0).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 0 {
		t.Fatalf("cursor = %d, want 0 (skipping separator)", m.CursorIndex())
	}
}

func TestMenu_SeparatorAtStart(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Separator: true},
			{Label: "A"},
		},
	})
	// Cursor should skip to A.
	if m.CursorIndex() != 1 {
		t.Fatalf("cursor = %d, want 1 (skip initial separator)", m.CursorIndex())
	}
}

func TestMenu_EnterActivates(t *testing.T) {
	activated := ""
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A", Action: func() { activated = "A" }},
			{Label: "B", Action: func() { activated = "B" }},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	// Move to B and activate.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	m = updated.(*blit.Menu)
	if activated != "B" {
		t.Fatalf("activated = %q, want B", activated)
	}
	if m.IsActive() {
		t.Fatal("menu should close after activation")
	}
}

func TestMenu_SpaceActivates(t *testing.T) {
	activated := false
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A", Action: func() { activated = true }},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	_ = updated.(*blit.Menu)
	if !activated {
		t.Fatal("space should activate item")
	}
}

func TestMenu_DisabledItem(t *testing.T) {
	activated := false
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A", Disabled: true, Action: func() { activated = true }},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	m = updated.(*blit.Menu)
	if activated {
		t.Fatal("disabled item should not activate")
	}
	if !m.IsActive() {
		t.Fatal("menu should stay open when disabled item is activated")
	}
}

func TestMenu_EscCloses(t *testing.T) {
	closed := false
	m := blit.NewMenu(blit.MenuOpts{
		Items:   []blit.MenuItem{{Label: "A"}},
		OnClose: func() { closed = true },
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.IsActive() {
		t.Fatal("menu should close after esc")
	}
	if !closed {
		t.Fatal("OnClose should be called")
	}
}

func TestMenu_CloseCallsOnClose(t *testing.T) {
	closed := false
	m := blit.NewMenu(blit.MenuOpts{
		Items:   []blit.MenuItem{{Label: "A"}},
		OnClose: func() { closed = true },
	})
	m.Close()
	if m.IsActive() {
		t.Fatal("menu should not be active after Close()")
	}
	if !closed {
		t.Fatal("OnClose should be called")
	}
}

func TestMenu_InactiveViewEmpty(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{{Label: "A"}},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.Close()

	if m.View() != "" {
		t.Fatal("inactive menu should return empty view")
	}
}

func TestMenu_EmptyItems(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	if m.View() != "" {
		t.Fatal("menu with no items should return empty view")
	}
}

func TestMenu_UnfocusedIgnoresInput(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{{Label: "A"}, {Label: "B"}},
	})
	m.SetFocused(false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.CursorIndex() != 0 {
		t.Fatal("unfocused menu should not process input")
	}
}

func TestMenu_KeyBindings(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{Items: []blit.MenuItem{{Label: "A"}}})
	binds := m.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestMenu_ComponentInterface(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{Items: []blit.MenuItem{{Label: "A"}}})

	if cmd := m.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	if m.Focused() {
		t.Fatal("should not be focused by default")
	}
	m.SetFocused(true)
	if !m.Focused() {
		t.Fatal("should be focused after SetFocused(true)")
	}
}

func TestMenu_MinWidth(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items:    []blit.MenuItem{{Label: "A"}},
		MinWidth: 30,
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Fatal("View() should not be empty with MinWidth")
	}
}

func TestMenu_SeparatorView(t *testing.T) {
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{
			{Label: "A"},
			{Separator: true},
			{Label: "B"},
		},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)

	view := m.View()
	if !strings.Contains(view, "─") {
		t.Fatalf("view should contain separator line:\n%s", view)
	}
}

func TestMenu_NilAction(t *testing.T) {
	// Item with nil action should still close menu.
	m := blit.NewMenu(blit.MenuOpts{
		Items: []blit.MenuItem{{Label: "Close"}},
	})
	m.SetTheme(blit.DefaultTheme())
	m.SetSize(80, 24)
	m.SetFocused(true)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	m = updated.(*blit.Menu)
	if m.IsActive() {
		t.Fatal("menu should close even with nil action")
	}
}
