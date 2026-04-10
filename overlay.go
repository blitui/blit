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

// contains reports whether the overlay is anywhere in the stack.
func (s *overlayStack) contains(o Overlay) bool {
	for _, item := range s.stack {
		if item == o {
			return true
		}
	}
	return false
}

// remove removes a specific overlay from the stack without closing it.
func (s *overlayStack) remove(o Overlay) {
	filtered := s.stack[:0]
	for _, item := range s.stack {
		if item != o {
			filtered = append(filtered, item)
		}
	}
	s.stack = filtered
}

// active returns the top overlay, or nil if the stack is empty.
func (s *overlayStack) active() Overlay {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}
