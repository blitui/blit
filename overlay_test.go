package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// stubOverlay is a minimal Overlay for testing.
type stubOverlay struct {
	name    string
	active  bool
	focused bool
	width   int
	height  int
}

func (s *stubOverlay) Init() tea.Cmd                                        { return nil }
func (s *stubOverlay) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) { return s, nil }
func (s *stubOverlay) View() string                                         { return s.name }
func (s *stubOverlay) KeyBindings() []KeyBind                               { return nil }
func (s *stubOverlay) SetSize(w, h int)                                     { s.width = w; s.height = h }
func (s *stubOverlay) Focused() bool                                        { return s.focused }
func (s *stubOverlay) SetFocused(f bool)                                    { s.focused = f }
func (s *stubOverlay) IsActive() bool                                       { return s.active }
func (s *stubOverlay) SetActive(v bool)                                     { s.active = v }
func (s *stubOverlay) Close()                                               { s.active = false }

func TestOverlayStackEmpty(t *testing.T) {
	stack := newOverlayStack()
	if stack.active() != nil {
		t.Error("empty stack should return nil active")
	}
}

func TestOverlayStackPushPop(t *testing.T) {
	stack := newOverlayStack()
	o1 := &stubOverlay{name: "help", active: true}
	o2 := &stubOverlay{name: "config", active: true}

	stack.push(o1)
	if stack.active() != o1 {
		t.Error("active should be o1")
	}

	stack.push(o2)
	if stack.active() != o2 {
		t.Error("active should be o2 after second push")
	}

	stack.pop()
	if stack.active() != o1 {
		t.Error("active should be o1 after pop")
	}

	stack.pop()
	if stack.active() != nil {
		t.Error("active should be nil after popping all")
	}
}

func TestOverlayStackPopClosesOverlay(t *testing.T) {
	stack := newOverlayStack()
	o := &stubOverlay{name: "help", active: true}
	stack.push(o)
	stack.pop()
	if o.active {
		t.Error("pop should call Close(), setting active to false")
	}
}
