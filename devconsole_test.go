package blit

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDevConsoleToggle verifies that ctrl+\ toggles the dev console on and off.
func TestDevConsoleToggle(t *testing.T) {
	main := &stubComponent{name: "main"}
	a := newAppModel(WithComponent("main", main))
	a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Console should not exist yet (no WithDevConsole, no env var)
	if a.devConsole != nil && a.devConsole.active {
		t.Fatal("dev console should be inactive before first toggle")
	}

	// Toggle on via ctrl+\
	a.Update(tea.KeyMsg{Type: tea.KeyCtrlBackslash})
	if a.devConsole == nil {
		t.Fatal("devConsole should be created after first toggle")
	}
	if !a.devConsole.active {
		t.Fatal("devConsole should be active after first toggle")
	}

	// Overlay stack should contain the console
	found := false
	for _, o := range a.overlays.stack {
		if o == a.devConsole {
			found = true
		}
	}
	if !found {
		t.Error("devConsole not found in overlay stack after toggle on")
	}

	// Toggle off
	a.Update(tea.KeyMsg{Type: tea.KeyCtrlBackslash})
	if a.devConsole.active {
		t.Fatal("devConsole should be inactive after second toggle")
	}
	for _, o := range a.overlays.stack {
		if o == a.devConsole {
			t.Error("devConsole should not be in overlay stack after toggle off")
		}
	}
}

// TestDevConsoleWithDevConsoleOption verifies WithDevConsole activates on startup.
func TestDevConsoleWithDevConsoleOption(t *testing.T) {
	main := &stubComponent{name: "main"}
	a := newAppModel(
		WithComponent("main", main),
		WithDevConsole(),
	)
	a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if a.devConsole == nil {
		t.Fatal("devConsole should be created by WithDevConsole")
	}
	if !a.devConsole.active {
		t.Fatal("devConsole should be active when created via WithDevConsole")
	}
}

// TestDevConsoleResize verifies alt+arrows resize the console panel.
func TestDevConsoleResize(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)

	origW := dc.w
	origH := dc.h
	if origW == 0 || origH == 0 {
		t.Fatalf("SetSize did not initialise w/h: w=%d h=%d", origW, origH)
	}

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	// alt+right increases width
	dc.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: true}, ctx)
	if dc.w <= origW {
		t.Errorf("expected width to increase after alt+right: got %d, orig %d", dc.w, origW)
	}

	// alt+down increases height
	dc.Update(tea.KeyMsg{Type: tea.KeyDown, Alt: true}, ctx)
	if dc.h <= origH {
		t.Errorf("expected height to increase after alt+down: got %d, orig %d", dc.h, origH)
	}

	// alt+left decreases width
	wBefore := dc.w
	dc.Update(tea.KeyMsg{Type: tea.KeyLeft, Alt: true}, ctx)
	if dc.w >= wBefore {
		t.Errorf("expected width to decrease after alt+left: got %d, before %d", dc.w, wBefore)
	}

	// alt+up decreases height
	hBefore := dc.h
	dc.Update(tea.KeyMsg{Type: tea.KeyUp, Alt: true}, ctx)
	if dc.h >= hBefore {
		t.Errorf("expected height to decrease after alt+up: got %d, before %d", dc.h, hBefore)
	}
}

// TestDevConsoleSignalIntrospection verifies signal values appear in the snapshot.
func TestDevConsoleSignalIntrospection(t *testing.T) {
	sig := NewSignal("hello")
	main := &stubComponent{name: "main"}
	a := newAppModel(
		WithComponent("main", main),
		WithStatusBarSignal(sig, nil),
		WithDevConsole(),
	)
	a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Force a snapshot update
	a.devConsoleUpdateSnapshot()

	if len(a.devConsole.snapshot.signals) == 0 {
		t.Fatal("expected at least one signal in snapshot")
	}
	found := false
	for _, s := range a.devConsole.snapshot.signals {
		if s.value == "hello" {
			found = true
		}
	}
	if !found {
		t.Errorf("signal value 'hello' not found in snapshot: %+v", a.devConsole.snapshot.signals)
	}
}

// TestDevConsoleKeyRecording verifies keypresses are recorded in the ring buffer.
func TestDevConsoleKeyRecording(t *testing.T) {
	main := &stubComponent{name: "main"}
	a := newAppModel(
		WithComponent("main", main),
		WithDevConsole(),
	)
	a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Toggle on first so ctrl+\ is recorded too
	a.Update(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	// Send some keys — they should be recorded even when console is active
	// because handleKey records before dispatching
	a.devConsole.recordKey("j")
	a.devConsole.recordKey("k")
	a.devConsole.recordKey("enter")

	keys := a.devConsole.recentKeys()
	if len(keys) == 0 {
		t.Fatal("expected recorded keys, got none")
	}
	last := keys[len(keys)-1]
	if last != "enter" {
		t.Errorf("expected last key 'enter', got %q", last)
	}
}

// TestDevConsoleFloatView verifies that FloatView composites the panel on top of background content.
func TestDevConsoleFloatView(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(80, 24)

	// Build a simple background
	bg := strings.Repeat(strings.Repeat(".", 80)+"\n", 24)
	bg = strings.TrimRight(bg, "\n")

	result := dc.FloatView(bg)
	if result == bg {
		t.Error("FloatView should modify the background when console is active")
	}
	if result == "" {
		t.Error("FloatView should not return empty string")
	}
}

// TestDevConsoleView verifies View returns empty when inactive.
func TestDevConsoleView(t *testing.T) {
	dc := newDevConsole()
	dc.active = false
	if v := dc.View(); v != "" {
		t.Errorf("View should return empty when inactive, got %q", v)
	}

	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(80, 24)
	if v := dc.View(); v == "" {
		t.Error("View should return non-empty when active")
	}
}

// TestDevConsoleFPSRingBuffer verifies fps() returns a reasonable value.
func TestDevConsoleFPSRingBuffer(t *testing.T) {
	dc := newDevConsole()
	// Single frame — no fps yet
	if dc.fps() != 0 {
		t.Error("fps should be 0 with fewer than 2 frames")
	}
	// Record two frames ~16ms apart
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(16 * time.Millisecond)
	dc.frameTimes[0] = t0
	dc.frameTimes[1] = t1
	dc.frameHead = 2
	dc.frameCount = 2

	fps := dc.fps()
	if fps < 50 || fps > 70 {
		t.Errorf("expected ~62.5 fps, got %.1f", fps)
	}
}

// TestDevConsoleKeyRingBuffer verifies the key ring buffer wraps correctly.
func TestDevConsoleKeyRingBuffer(t *testing.T) {
	dc := newDevConsole()
	// Record 25 keys — buffer holds 20
	for i := 0; i < 25; i++ {
		dc.recordKey(string(rune('a' + i)))
	}
	keys := dc.recentKeys()
	if len(keys) != 20 {
		t.Fatalf("expected 20 keys in ring buffer, got %d", len(keys))
	}
	// The oldest should be key index 5 ('f'), newest index 24 ('y')
	if keys[0] != "f" {
		t.Errorf("expected oldest key 'f', got %q", keys[0])
	}
	if keys[19] != "y" {
		t.Errorf("expected newest key 'y', got %q", keys[19])
	}
}
