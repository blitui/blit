package blit_test

import (
	"strings"
	"testing"
	"time"

	blit "github.com/blitui/blit"
)

func makeTestSpinner() *blit.Spinner {
	s := blit.NewSpinner(blit.SpinnerOpts{Label: "Loading..."})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(80, 1)
	return s
}

func TestSpinner_NewDefaults(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	if !s.Active() {
		t.Fatal("spinner should be active by default")
	}
	if s.Frame() != 0 {
		t.Fatalf("Frame() = %d, want 0", s.Frame())
	}
}

func TestSpinner_Label(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{Label: "Working"})
	if s.Label() != "Working" {
		t.Fatalf("Label() = %q, want Working", s.Label())
	}
	s.SetLabel("Done")
	if s.Label() != "Done" {
		t.Fatalf("Label() = %q, want Done", s.Label())
	}
}

func TestSpinner_View(t *testing.T) {
	s := makeTestSpinner()
	view := s.View()
	if !strings.Contains(view, "Loading...") {
		t.Fatalf("view should contain label:\n%s", view)
	}
}

func TestSpinner_ViewNoLabel(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(80, 1)
	view := s.View()
	if view == "" {
		t.Fatal("spinner with no label should still render the icon")
	}
}

func TestSpinner_TickAdvancesFrame(t *testing.T) {
	s := makeTestSpinner()

	// Simulate a spinner tick message by sending the tick msg type.
	// We use the spinnerTickMsg through Update.
	cmd := s.Init()
	if cmd == nil {
		t.Fatal("Init() should return a tick command")
	}

	// Execute the command to get the message, then send it.
	msg := cmd()
	updated, cmd := s.Update(msg, blit.Context{})
	s = updated.(*blit.Spinner)

	if s.Frame() != 1 {
		t.Fatalf("Frame() = %d, want 1 after tick", s.Frame())
	}

	// Another tick.
	if cmd == nil {
		t.Fatal("Update should return another tick command")
	}
	msg = cmd()
	updated, _ = s.Update(msg, blit.Context{})
	s = updated.(*blit.Spinner)

	if s.Frame() != 2 {
		t.Fatalf("Frame() = %d, want 2 after second tick", s.Frame())
	}
}

func TestSpinner_FrameWraps(t *testing.T) {
	s := makeTestSpinner()

	// Default spinner has 10 frames. Tick 10 times to wrap.
	cmd := s.Init()
	for i := 0; i < 10; i++ {
		msg := cmd()
		var updated blit.Component
		updated, cmd = s.Update(msg, blit.Context{})
		s = updated.(*blit.Spinner)
	}

	if s.Frame() != 0 {
		t.Fatalf("Frame() = %d, want 0 after wrapping", s.Frame())
	}
}

func TestSpinner_InactiveNoTick(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	s.SetActive(false)

	cmd := s.Init()
	if cmd != nil {
		t.Fatal("inactive spinner Init() should return nil")
	}
}

func TestSpinner_SetActive(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	if !s.Active() {
		t.Fatal("should be active by default")
	}
	s.SetActive(false)
	if s.Active() {
		t.Fatal("should be inactive after SetActive(false)")
	}
	s.SetActive(true)
	if !s.Active() {
		t.Fatal("should be active after SetActive(true)")
	}
}

func TestSpinner_CustomInterval(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{Interval: 200 * time.Millisecond})
	if !s.Active() {
		t.Fatal("should be active")
	}
	cmd := s.Init()
	if cmd == nil {
		t.Fatal("Init() should return a tick command")
	}
}

func TestSpinner_ZeroSize(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{Label: "Test"})
	s.SetTheme(blit.DefaultTheme())
	s.SetSize(0, 0)
	if s.View() != "" {
		t.Fatal("zero-sized spinner should return empty view")
	}
}

func TestSpinner_ComponentInterface(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	if s.Focused() {
		t.Fatal("should not be focused by default")
	}
	s.SetFocused(true)
	if !s.Focused() {
		t.Fatal("should be focused")
	}
}

func TestSpinner_KeyBindings(t *testing.T) {
	s := blit.NewSpinner(blit.SpinnerOpts{})
	if binds := s.KeyBindings(); binds != nil {
		t.Fatal("spinner should have no key bindings")
	}
}

func TestSpinner_InactiveIgnoresTick(t *testing.T) {
	s := makeTestSpinner()
	cmd := s.Init()
	msg := cmd()

	s.SetActive(false)
	updated, cmd := s.Update(msg, blit.Context{})
	s = updated.(*blit.Spinner)

	if s.Frame() != 0 {
		t.Fatal("inactive spinner should not advance frame")
	}
	if cmd != nil {
		t.Fatal("inactive spinner should not schedule next tick")
	}
}

func TestSpinner_ViewChangesPerFrame(t *testing.T) {
	s := makeTestSpinner()
	view1 := s.View()

	cmd := s.Init()
	msg := cmd()
	updated, _ := s.Update(msg, blit.Context{})
	s = updated.(*blit.Spinner)

	view2 := s.View()
	if view1 == view2 {
		t.Fatal("view should change after frame advance")
	}
}
