package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func makeTestStepper() *blit.Stepper {
	steps := []blit.Step{
		{Title: "Setup"},
		{Title: "Configure"},
		{Title: "Deploy"},
	}
	s := blit.NewStepper(steps, blit.StepperOpts{})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(80, 10)
	s.SetFocused(true)
	return s
}

func TestStepper_NewDefaults(t *testing.T) {
	s := makeTestStepper()
	if s.Current() != 0 {
		t.Fatalf("Current() = %d, want 0", s.Current())
	}
	if len(s.Steps()) != 3 {
		t.Fatalf("Steps() = %d, want 3", len(s.Steps()))
	}
}

func TestStepper_View(t *testing.T) {
	s := makeTestStepper()
	view := s.View()
	if !strings.Contains(view, "Setup") {
		t.Fatalf("view should contain 'Setup':\n%s", view)
	}
	if !strings.Contains(view, "Configure") {
		t.Fatalf("view should contain 'Configure':\n%s", view)
	}
	if !strings.Contains(view, "Deploy") {
		t.Fatalf("view should contain 'Deploy':\n%s", view)
	}
}

func TestStepper_Next(t *testing.T) {
	s := makeTestStepper()
	s.Next()
	if s.Current() != 1 {
		t.Fatalf("Current() = %d, want 1", s.Current())
	}
	s.Next()
	if s.Current() != 2 {
		t.Fatalf("Current() = %d, want 2", s.Current())
	}
}

func TestStepper_NextAtEnd(t *testing.T) {
	completed := false
	steps := []blit.Step{{Title: "Only"}}
	s := blit.NewStepper(steps, blit.StepperOpts{
		OnComplete: func() { completed = true },
	})
	s.Next()
	if !completed {
		t.Fatal("OnComplete should be called at last step")
	}
	if s.Current() != 0 {
		t.Fatalf("Current() = %d, want 0 (should not advance past end)", s.Current())
	}
}

func TestStepper_Prev(t *testing.T) {
	s := makeTestStepper()
	s.Next()
	s.Next()
	s.Prev()
	if s.Current() != 1 {
		t.Fatalf("Current() = %d, want 1", s.Current())
	}
}

func TestStepper_PrevAtStart(t *testing.T) {
	s := makeTestStepper()
	s.Prev()
	if s.Current() != 0 {
		t.Fatalf("Current() = %d, want 0 (clamped at start)", s.Current())
	}
}

func TestStepper_SetCurrent(t *testing.T) {
	s := makeTestStepper()
	s.SetCurrent(2)
	if s.Current() != 2 {
		t.Fatalf("Current() = %d, want 2", s.Current())
	}

	// Clamp above.
	s.SetCurrent(100)
	if s.Current() != 2 {
		t.Fatalf("Current() = %d, want 2 (clamped)", s.Current())
	}

	// Clamp below.
	s.SetCurrent(-5)
	if s.Current() != 0 {
		t.Fatalf("Current() = %d, want 0 (clamped)", s.Current())
	}
}

func TestStepper_Status(t *testing.T) {
	s := makeTestStepper()
	s.SetCurrent(1)

	if s.Status(0) != blit.StepDone {
		t.Fatal("step 0 should be Done")
	}
	if s.Status(1) != blit.StepActive {
		t.Fatal("step 1 should be Active")
	}
	if s.Status(2) != blit.StepPending {
		t.Fatal("step 2 should be Pending")
	}
}

func TestStepper_KeyNavRight(t *testing.T) {
	s := makeTestStepper()

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 1 {
		t.Fatalf("Current() = %d, want 1", s.Current())
	}
}

func TestStepper_KeyNavLeft(t *testing.T) {
	s := makeTestStepper()
	s.SetCurrent(2)

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 1 {
		t.Fatalf("Current() = %d, want 1", s.Current())
	}
}

func TestStepper_KeyNavTab(t *testing.T) {
	s := makeTestStepper()

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 1 {
		t.Fatalf("tab: Current() = %d, want 1", s.Current())
	}
}

func TestStepper_KeyNavShiftTab(t *testing.T) {
	s := makeTestStepper()
	s.SetCurrent(1)

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyShiftTab}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 0 {
		t.Fatalf("shift+tab: Current() = %d, want 0", s.Current())
	}
}

func TestStepper_ViNavigation(t *testing.T) {
	s := makeTestStepper()

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 1 {
		t.Fatalf("l: Current() = %d, want 1", s.Current())
	}

	updated, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 0 {
		t.Fatalf("h: Current() = %d, want 0", s.Current())
	}
}

func TestStepper_OnChange(t *testing.T) {
	var changedTo int
	steps := []blit.Step{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	s := blit.NewStepper(steps, blit.StepperOpts{
		OnChange: func(step int) { changedTo = step },
	})
	s.Next()
	if changedTo != 1 {
		t.Fatalf("OnChange step = %d, want 1", changedTo)
	}
}

func TestStepper_Description(t *testing.T) {
	steps := []blit.Step{
		{Title: "Setup", Description: "Install dependencies"},
	}
	s := blit.NewStepper(steps, blit.StepperOpts{})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(80, 10)
	view := s.View()
	if !strings.Contains(view, "Install dependencies") {
		t.Fatalf("view should contain description:\n%s", view)
	}
}

func TestStepper_Empty(t *testing.T) {
	s := blit.NewStepper([]blit.Step{}, blit.StepperOpts{})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(80, 10)
	if s.View() != "" {
		t.Fatal("empty stepper should return empty view")
	}
}

func TestStepper_ZeroSize(t *testing.T) {
	s := makeTestStepper()
	s.SetSize(0, 0)
	if s.View() != "" {
		t.Fatal("zero-sized stepper should return empty view")
	}
}

func TestStepper_UnfocusedIgnoresInput(t *testing.T) {
	s := makeTestStepper()
	s.SetFocused(false)

	updated, _ := s.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	s = updated.(*blit.Stepper)
	if s.Current() != 0 {
		t.Fatal("unfocused stepper should not process input")
	}
}

func TestStepper_ComponentInterface(t *testing.T) {
	s := blit.NewStepper([]blit.Step{{Title: "A"}}, blit.StepperOpts{})
	if cmd := s.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	if s.Focused() {
		t.Fatal("should not be focused by default")
	}
	s.SetFocused(true)
	if !s.Focused() {
		t.Fatal("should be focused")
	}
}

func TestStepper_KeyBindings(t *testing.T) {
	s := blit.NewStepper([]blit.Step{{Title: "A"}}, blit.StepperOpts{})
	binds := s.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}
