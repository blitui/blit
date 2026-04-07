package tuikit

import (
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
func (s *stubComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	s.lastMsg = msg
	if km, ok := msg.(tea.KeyMsg); ok {
		s.lastKey = km.String()
		return s, Consumed()
	}
	return s, nil
}
func (s *stubComponent) View() string          { return s.name }
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
