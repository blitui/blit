package tuikit

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// stubComponent is a minimal Component for testing.
type stubComponent struct {
	name     string
	focused  bool
	width    int
	height   int
	bindings []KeyBind
	lastKey  string
	lastMsg  tea.Msg
}

func (s *stubComponent) Init() tea.Cmd { return nil }
func (s *stubComponent) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	s.lastMsg = msg
	if km, ok := msg.(tea.KeyMsg); ok {
		s.lastKey = km.String()
		return s, Consumed()
	}
	return s, nil
}
func (s *stubComponent) View() string {
	if s.height > 0 {
		lines := make([]string, s.height)
		for i := range lines {
			lines[i] = s.name
		}
		return strings.Join(lines, "\n")
	}
	return s.name
}
func (s *stubComponent) KeyBindings() []KeyBind { return s.bindings }
func (s *stubComponent) SetSize(w, h int)      { s.width = w; s.height = h }
func (s *stubComponent) Focused() bool          { return s.focused }
func (s *stubComponent) SetFocused(f bool)      { s.focused = f }

func TestAppFocusCycle(t *testing.T) {
	c1 := &stubComponent{name: "one"}
	c2 := &stubComponent{name: "two"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("one", c1),
		WithComponent("two", c2),
	)

	if !c1.focused {
		t.Error("first component should be focused initially")
	}
	if c2.focused {
		t.Error("second component should not be focused initially")
	}

	a.Update(tea.KeyMsg{Type: tea.KeyTab})
	if c1.focused {
		t.Error("first component should lose focus after Tab")
	}
	if !c2.focused {
		t.Error("second component should gain focus after Tab")
	}
}

func TestAppKeyDispatchToFocused(t *testing.T) {
	c1 := &stubComponent{name: "one"}
	c2 := &stubComponent{name: "two"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("one", c1),
		WithComponent("two", c2),
	)

	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if c1.lastKey != "x" {
		t.Errorf("focused component should receive key, got '%s'", c1.lastKey)
	}
	if c2.lastKey != "" {
		t.Error("unfocused component should not receive key")
	}
}

func TestAppOverlayPriority(t *testing.T) {
	c := &stubComponent{name: "main"}
	o := &stubOverlay{name: "overlay", active: true}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithOverlay("test", "o", o),
	)
	a.overlays.push(o)

	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if c.lastKey != "" {
		t.Error("component should not receive key when overlay is active")
	}
}

func TestAppQuit(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("'q' should produce a quit command")
	}
}

func TestAppTickForwarding(t *testing.T) {
	c := &stubComponent{name: "main"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithTickInterval(100*time.Millisecond),
	)

	tick := TickMsg{Time: time.Now()}
	a.Update(tick)

	if _, ok := c.lastMsg.(TickMsg); !ok {
		t.Errorf("component should receive TickMsg, got %T", c.lastMsg)
	}
}

func TestAppTickCmd(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithTickInterval(100*time.Millisecond),
	)

	cmd := a.tickCmd()
	if cmd == nil {
		t.Error("tickCmd should return a command when interval is set")
	}
}

func TestAppNoTickWithoutInterval(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))

	cmd := a.tickCmd()
	if cmd != nil {
		t.Error("tickCmd should return nil when no interval is set")
	}
}

func TestAppMouseForwarding(t *testing.T) {
	c := &stubComponent{name: "main"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithMouseSupport(),
	)

	mouseMsg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	a.Update(mouseMsg)

	if _, ok := c.lastMsg.(tea.MouseMsg); !ok {
		t.Errorf("component should receive MouseMsg, got %T", c.lastMsg)
	}
}

func TestAppKeyBindHandler(t *testing.T) {
	called := false
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithKeyBind(KeyBind{
			Key:   "f",
			Label: "Do thing",
			Group: "OTHER",
			Handler: func() {
				called = true
			},
		}),
	)

	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if !called {
		t.Error("keybind handler should have been called")
	}
}

func TestAppOverlayTriggerKey(t *testing.T) {
	o := &stubOverlay{name: "config"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithOverlay("config", "c", o),
	)

	// Press 'c' — should open the overlay
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if a.overlays.active() == nil {
		t.Error("overlay should be active after pressing trigger key")
	}
}

func TestAppUnknownMessageForwarding(t *testing.T) {
	c := &stubComponent{name: "main"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
	)

	// Custom message type
	type customMsg struct{ data string }
	a.Update(customMsg{data: "hello"})

	if msg, ok := c.lastMsg.(customMsg); !ok {
		t.Errorf("component should receive custom msg, got %T", c.lastMsg)
	} else if msg.data != "hello" {
		t.Errorf("expected data 'hello', got '%s'", msg.data)
	}
}

func TestAppNotifySetAndExpire(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithTickInterval(100*time.Millisecond),
	)
	a.width = 80
	a.height = 24
	a.resize()

	// Send a NotifyMsg
	a.Update(NotifyMsg{Text: "hello", Duration: 50 * time.Millisecond})

	if a.notifyMsg != "hello" {
		t.Errorf("expected notify msg 'hello', got '%s'", a.notifyMsg)
	}

	// Simulate expiry
	time.Sleep(60 * time.Millisecond)
	a.Update(TickMsg{Time: time.Now()})

	if a.notifyMsg != "" {
		t.Errorf("notify msg should be cleared after expiry, got '%s'", a.notifyMsg)
	}
}

func TestAppNotifyDefaultDuration(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))

	a.Update(NotifyMsg{Text: "test"})

	if a.notifyMsg != "test" {
		t.Errorf("expected 'test', got '%s'", a.notifyMsg)
	}
	if a.notifyExpiry.IsZero() {
		t.Error("expiry should be set even with zero duration (default 2s)")
	}
}

func TestAppNotifyReplace(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))

	a.Update(NotifyMsg{Text: "first", Duration: 5 * time.Second})
	a.Update(NotifyMsg{Text: "second", Duration: 5 * time.Second})

	if a.notifyMsg != "second" {
		t.Errorf("expected 'second', got '%s'", a.notifyMsg)
	}
}

func TestAppNotifyRendersInView(t *testing.T) {
	c := &stubComponent{name: "main"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
	)
	a.width = 80
	a.height = 24
	a.resize()

	a.Update(NotifyMsg{Text: "notification!", Duration: 5 * time.Second})

	view := a.View()
	if !strings.Contains(view, "notification!") {
		t.Error("view should contain the notification text")
	}
}

func TestNotifyCmd(t *testing.T) {
	cmd := NotifyCmd("test", 2*time.Second)
	msg := cmd()
	nm, ok := msg.(NotifyMsg)
	if !ok {
		t.Fatalf("expected NotifyMsg, got %T", msg)
	}
	if nm.Text != "test" {
		t.Errorf("expected 'test', got '%s'", nm.Text)
	}
	if nm.Duration != 2*time.Second {
		t.Errorf("expected 2s, got %v", nm.Duration)
	}
}

func TestAppAutoPushActiveOverlay(t *testing.T) {
	c := &stubComponent{name: "main"}
	o := &stubOverlay{name: "detail"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithOverlay("detail", "", o), // no trigger key
	)

	// Overlay not on stack yet
	if a.overlays.active() != nil {
		t.Error("no overlay should be active initially")
	}

	// Simulate component activating the overlay
	o.active = true

	// Broadcast a message to trigger the auto-push check
	a.Update(TickMsg{Time: time.Now()})

	if a.overlays.active() != o {
		t.Error("overlay should be auto-pushed after becoming active")
	}
}

func TestDualPaneHeightConsistency(t *testing.T) {
	main := &stubComponent{name: "M"}
	side := &stubComponent{name: "S"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithLayout(&DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    20,
			MinMainWidth: 40,
			SideRight:    true,
			ToggleKey:    "p",
		}),
		WithStatusBar(
			func() string { return "left" },
			func() string { return "right" },
		),
	)

	totalHeight := 24
	// Width wide enough for sidebar to show: need >= MinMainWidth + SideWidth + 3 = 63
	wideWidth := 80

	a.width = wideWidth
	a.height = totalHeight
	a.resize()

	view := a.View()
	lines := strings.Split(view, "\n")

	if len(lines) != totalHeight {
		t.Errorf("DualPane wide view: expected %d lines, got %d", totalHeight, len(lines))
	}

	// Badge line should exist (2 focusable components)
	_, _, vis := a.dualPane.compute(wideWidth, 0)
	if !vis {
		t.Error("sidebar should be visible at wide width")
	}

	// Now narrow the terminal so sidebar auto-hides
	narrowWidth := 50 // < 40 + 20 + 3 = 63
	a.width = narrowWidth
	a.resize()

	view = a.View()
	lines = strings.Split(view, "\n")

	if len(lines) != totalHeight {
		t.Errorf("DualPane narrow view: expected %d lines, got %d", totalHeight, len(lines))
	}

	_, _, vis = a.dualPane.compute(narrowWidth, 0)
	if vis {
		t.Error("sidebar should be hidden at narrow width")
	}

	// Test the transition: go wide again
	a.width = wideWidth
	a.resize()

	view = a.View()
	lines = strings.Split(view, "\n")

	if len(lines) != totalHeight {
		t.Errorf("DualPane re-widened view: expected %d lines, got %d", totalHeight, len(lines))
	}
}

func TestDualPaneBadgesMatchVisibility(t *testing.T) {
	main := &stubComponent{name: "M"}
	side := &stubComponent{name: "S"}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithLayout(&DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    20,
			MinMainWidth: 40,
			SideRight:    true,
		}),
	)

	// Wide: sidebar visible, badges should show
	a.width = 80
	a.height = 24
	a.resize()

	if !a.showBadges() {
		t.Error("badges should show when sidebar is visible (2 focusable)")
	}

	// Narrow: sidebar hidden, badges should NOT show (only 1 focusable)
	a.width = 50
	a.resize()

	if a.showBadges() {
		t.Error("badges should not show when sidebar is auto-hidden (1 focusable)")
	}
}
