package blit

// overlayStack manages a stack of modal overlays.
// The top overlay receives all key events first. Esc pops it.
type overlayStack struct {
	stack []Overlay
}

func newOverlayStack() *overlayStack {
	return &overlayStack{}
}

// push adds an overlay to the top of the stack.
func (s *overlayStack) push(o Overlay) {
	s.stack = append(s.stack, o)
}

// pop removes and closes the top overlay.
func (s *overlayStack) pop() {
	if len(s.stack) == 0 {
		return
	}
	top := s.stack[len(s.stack)-1]
	top.Close()
	s.stack = s.stack[:len(s.stack)-1]
}

// active returns the top overlay, or nil if the stack is empty.
func (s *overlayStack) active() Overlay {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}
