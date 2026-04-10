package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func TestDialog_NewDefaults(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Body: "Hello",
	})
	if !d.IsActive() {
		t.Fatal("dialog should be active after creation")
	}
	if d.CursorIndex() != 0 {
		t.Fatal("cursor should start at 0")
	}
}

func TestDialog_View(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Title: "Confirm",
		Body:  "Are you sure?",
		Buttons: []blit.DialogButton{
			{Label: "Yes"},
			{Label: "No"},
		},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	view := d.View()
	if !strings.Contains(view, "Are you sure?") {
		t.Fatalf("view should contain body:\n%s", view)
	}
	if !strings.Contains(view, "Yes") {
		t.Fatalf("view should contain Yes button:\n%s", view)
	}
	if !strings.Contains(view, "No") {
		t.Fatalf("view should contain No button:\n%s", view)
	}
}

func TestDialog_ViewWithTitle(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Title: "Warning",
		Body:  "Something happened",
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)

	view := d.View()
	if !strings.Contains(view, "Warning") {
		t.Fatalf("view should contain title:\n%s", view)
	}
}

func TestDialog_NavigateButtons(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Buttons: []blit.DialogButton{
			{Label: "A"},
			{Label: "B"},
			{Label: "C"},
		},
		Body: "pick",
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	// Move right.
	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 1 {
		t.Fatalf("cursor = %d, want 1", d.CursorIndex())
	}

	// Move right again.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 2 {
		t.Fatalf("cursor = %d, want 2", d.CursorIndex())
	}

	// Right at end doesn't go past.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 2 {
		t.Fatalf("cursor = %d, want 2 (clamped)", d.CursorIndex())
	}

	// Move left.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 1 {
		t.Fatalf("cursor = %d, want 1", d.CursorIndex())
	}

	// Left at start doesn't go negative.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	d = updated.(*blit.Dialog)
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 0 {
		t.Fatalf("cursor = %d, want 0 (clamped)", d.CursorIndex())
	}
}

func TestDialog_TabNavigation(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Buttons: []blit.DialogButton{{Label: "A"}, {Label: "B"}},
		Body:    "tab",
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	// Tab moves right.
	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 1 {
		t.Fatalf("tab: cursor = %d, want 1", d.CursorIndex())
	}

	// Shift+tab moves left.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyShiftTab}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 0 {
		t.Fatalf("shift+tab: cursor = %d, want 0", d.CursorIndex())
	}
}

func TestDialog_EnterActivatesButton(t *testing.T) {
	activated := false
	d := blit.NewDialog(blit.DialogOpts{
		Body: "confirm",
		Buttons: []blit.DialogButton{
			{Label: "OK", Action: func() { activated = true }},
		},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	d = updated.(*blit.Dialog)
	if !activated {
		t.Fatal("button action should have been called")
	}
	if d.IsActive() {
		t.Fatal("dialog should be closed after button activation")
	}
}

func TestDialog_SpaceActivatesButton(t *testing.T) {
	activated := false
	d := blit.NewDialog(blit.DialogOpts{
		Body: "confirm",
		Buttons: []blit.DialogButton{
			{Label: "OK", Action: func() { activated = true }},
		},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	_ = updated.(*blit.Dialog)
	if !activated {
		t.Fatal("space should activate button")
	}
}

func TestDialog_EscCloses(t *testing.T) {
	closed := false
	d := blit.NewDialog(blit.DialogOpts{
		Body:    "esc test",
		OnClose: func() { closed = true },
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyEsc}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.IsActive() {
		t.Fatal("dialog should be closed after esc")
	}
	if !closed {
		t.Fatal("OnClose should have been called")
	}
}

func TestDialog_CloseCallsOnClose(t *testing.T) {
	closed := false
	d := blit.NewDialog(blit.DialogOpts{
		Body:    "close test",
		OnClose: func() { closed = true },
	})
	d.Close()
	if d.IsActive() {
		t.Fatal("dialog should not be active after Close()")
	}
	if !closed {
		t.Fatal("OnClose should have been called")
	}
}

func TestDialog_InactiveViewEmpty(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{Body: "test"})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.Close()

	if d.View() != "" {
		t.Fatal("inactive dialog should return empty view")
	}
}

func TestDialog_UnfocusedIgnoresInput(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Body:    "test",
		Buttons: []blit.DialogButton{{Label: "A"}, {Label: "B"}},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(false)

	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 0 {
		t.Fatal("unfocused dialog should not process input")
	}
}

func TestDialog_DefaultOKButton(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{Body: "default"})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)

	view := d.View()
	if !strings.Contains(view, "OK") {
		t.Fatalf("dialog with no buttons should show OK:\n%s", view)
	}
}

func TestDialog_KeyBindings(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{Body: "test"})
	binds := d.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestDialog_ComponentInterface(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{Body: "test"})

	if cmd := d.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}

	if d.Focused() {
		t.Fatal("should not be focused by default")
	}
	d.SetFocused(true)
	if !d.Focused() {
		t.Fatal("should be focused after SetFocused(true)")
	}
}

func TestDialog_ViNavigation(t *testing.T) {
	d := blit.NewDialog(blit.DialogOpts{
		Body:    "vi",
		Buttons: []blit.DialogButton{{Label: "A"}, {Label: "B"}},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	// l moves right.
	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 1 {
		t.Fatalf("l: cursor = %d, want 1", d.CursorIndex())
	}

	// h moves left.
	updated, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.CursorIndex() != 0 {
		t.Fatalf("h: cursor = %d, want 0", d.CursorIndex())
	}
}

func TestDialog_ButtonWithNoAction(t *testing.T) {
	// Button with nil Action should still close the dialog.
	d := blit.NewDialog(blit.DialogOpts{
		Body:    "test",
		Buttons: []blit.DialogButton{{Label: "Close"}},
	})
	d.SetTheme(blit.DefaultTheme())
	d.SetSize(80, 24)
	d.SetFocused(true)

	updated, _ := d.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	d = updated.(*blit.Dialog)
	if d.IsActive() {
		t.Fatal("dialog should close even with nil action")
	}
}
