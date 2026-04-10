package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func makeTestAccordion() *blit.Accordion {
	sections := []blit.AccordionSection{
		{Title: "Section A", Content: "Content A"},
		{Title: "Section B", Content: "Content B"},
		{Title: "Section C", Content: "Content C"},
	}
	a := blit.NewAccordion(sections, blit.AccordionOpts{})
	a.SetTheme(blit.DefaultTheme())
	a.SetSize(80, 30)
	a.SetFocused(true)
	return a
}

func TestAccordion_NewDefaults(t *testing.T) {
	a := makeTestAccordion()
	if a.CursorIndex() != 0 {
		t.Fatalf("CursorIndex() = %d, want 0", a.CursorIndex())
	}
	if len(a.Sections()) != 3 {
		t.Fatalf("Sections() = %d, want 3", len(a.Sections()))
	}
}

func TestAccordion_View(t *testing.T) {
	a := makeTestAccordion()
	view := a.View()
	if !strings.Contains(view, "Section A") {
		t.Fatalf("view should contain 'Section A':\n%s", view)
	}
	if !strings.Contains(view, "Section B") {
		t.Fatalf("view should contain 'Section B':\n%s", view)
	}
}

func TestAccordion_Navigate(t *testing.T) {
	a := makeTestAccordion()

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", a.CursorIndex())
	}

	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2", a.CursorIndex())
	}

	// Clamp at end.
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2 (clamped)", a.CursorIndex())
	}

	// Up.
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", a.CursorIndex())
	}
}

func TestAccordion_ViNavigation(t *testing.T) {
	a := makeTestAccordion()

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 1 {
		t.Fatalf("j: CursorIndex() = %d, want 1", a.CursorIndex())
	}

	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 0 {
		t.Fatalf("k: CursorIndex() = %d, want 0", a.CursorIndex())
	}
}

func TestAccordion_ToggleEnter(t *testing.T) {
	a := makeTestAccordion()

	// Enter expands.
	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)
	if !a.Sections()[0].Expanded {
		t.Fatal("section A should be expanded after enter")
	}

	view := a.View()
	if !strings.Contains(view, "Content A") {
		t.Fatalf("expanded section should show content:\n%s", view)
	}

	// Enter again collapses.
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.Sections()[0].Expanded {
		t.Fatal("section A should be collapsed after second enter")
	}
}

func TestAccordion_ToggleSpace(t *testing.T) {
	a := makeTestAccordion()

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	a = updated.(*blit.Accordion)
	if !a.Sections()[0].Expanded {
		t.Fatal("space should expand section")
	}
}

func TestAccordion_ExclusiveMode(t *testing.T) {
	sections := []blit.AccordionSection{
		{Title: "A", Content: "Ca"},
		{Title: "B", Content: "Cb"},
		{Title: "C", Content: "Cc"},
	}
	a := blit.NewAccordion(sections, blit.AccordionOpts{Exclusive: true})
	a.SetTheme(blit.DefaultTheme())
	a.SetSize(80, 30)
	a.SetFocused(true)

	// Expand A.
	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)
	if !a.Sections()[0].Expanded {
		t.Fatal("A should be expanded")
	}

	// Move to B and expand — A should collapse.
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)
	if !a.Sections()[1].Expanded {
		t.Fatal("B should be expanded")
	}
	if a.Sections()[0].Expanded {
		t.Fatal("A should be collapsed in exclusive mode")
	}
}

func TestAccordion_OnToggle(t *testing.T) {
	var toggledIdx int
	var toggledExp bool
	sections := []blit.AccordionSection{
		{Title: "A", Content: "Ca"},
	}
	a := blit.NewAccordion(sections, blit.AccordionOpts{
		OnToggle: func(idx int, expanded bool) {
			toggledIdx = idx
			toggledExp = expanded
		},
	})
	a.SetTheme(blit.DefaultTheme())
	a.SetSize(80, 30)
	a.SetFocused(true)

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Accordion)
	if toggledIdx != 0 || !toggledExp {
		t.Fatalf("OnToggle: idx=%d exp=%v, want 0 true", toggledIdx, toggledExp)
	}
}

func TestAccordion_Empty(t *testing.T) {
	a := blit.NewAccordion([]blit.AccordionSection{}, blit.AccordionOpts{})
	a.SetTheme(blit.DefaultTheme())
	a.SetSize(80, 20)
	if a.View() != "" {
		t.Fatal("empty accordion should return empty view")
	}
}

func TestAccordion_ZeroSize(t *testing.T) {
	a := makeTestAccordion()
	a.SetSize(0, 0)
	if a.View() != "" {
		t.Fatal("zero-sized accordion should return empty view")
	}
}

func TestAccordion_UnfocusedIgnoresInput(t *testing.T) {
	a := makeTestAccordion()
	a.SetFocused(false)

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	if a.CursorIndex() != 0 {
		t.Fatal("unfocused accordion should not process input")
	}
}

func TestAccordion_KeyBindings(t *testing.T) {
	a := makeTestAccordion()
	binds := a.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestAccordion_ComponentInterface(t *testing.T) {
	a := makeTestAccordion()
	if cmd := a.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	a.SetFocused(false)
	if a.Focused() {
		t.Fatal("should not be focused")
	}
	a.SetFocused(true)
	if !a.Focused() {
		t.Fatal("should be focused")
	}
}

func TestAccordion_ContentHiddenWhenCollapsed(t *testing.T) {
	a := makeTestAccordion()
	view := a.View()
	if strings.Contains(view, "Content A") {
		t.Fatal("collapsed section should not show content")
	}
}

func TestAccordion_MultipleExpanded(t *testing.T) {
	// Non-exclusive mode allows multiple expanded.
	a := makeTestAccordion()

	// Expand A.
	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)

	// Move to B and expand.
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	a = updated.(*blit.Accordion)
	updated, _ = a.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	a = updated.(*blit.Accordion)

	if !a.Sections()[0].Expanded {
		t.Fatal("A should remain expanded in non-exclusive mode")
	}
	if !a.Sections()[1].Expanded {
		t.Fatal("B should be expanded")
	}
}
