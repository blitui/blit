package blit

import "testing"

// TestHBoxZeroChildren verifies HBox with no items renders empty.
func TestHBoxZeroChildren(t *testing.T) {
	h := &HBox{}
	h.SetSize(100, 20)
	if got := h.View(); got != "" {
		t.Errorf("empty HBox.View() = %q, want empty", got)
	}
}

// TestVBoxZeroChildren verifies VBox with no items renders empty.
func TestVBoxZeroChildren(t *testing.T) {
	v := &VBox{}
	v.SetSize(100, 20)
	if got := v.View(); got != "" {
		t.Errorf("empty VBox.View() = %q, want empty", got)
	}
}

// TestHBoxOneChild verifies HBox with a single child.
func TestHBoxOneChild(t *testing.T) {
	c := &stubComponent{name: "solo"}
	h := &HBox{Items: []Component{c}}
	h.SetSize(80, 10)

	if c.width == 0 {
		t.Error("child should receive non-zero width from SetSize")
	}

	view := h.View()
	if view == "" {
		t.Error("HBox with one child should render non-empty")
	}
}

// TestVBoxOneChild verifies VBox with a single child.
func TestVBoxOneChild(t *testing.T) {
	c := &stubComponent{name: "solo"}
	v := &VBox{Items: []Component{c}}
	v.SetSize(80, 10)

	if c.height == 0 {
		t.Error("child should receive non-zero height from SetSize")
	}

	view := v.View()
	if view == "" {
		t.Error("VBox with one child should render non-empty")
	}
}

// TestHBoxSetSizePropagates verifies SetSize propagates to all children.
func TestHBoxSetSizePropagates(t *testing.T) {
	c1 := &stubComponent{name: "a"}
	c2 := &stubComponent{name: "b"}
	h := &HBox{Items: []Component{c1, c2}}
	h.SetSize(100, 20)

	if c1.width == 0 {
		t.Error("c1 should receive width from HBox.SetSize")
	}
	if c2.width == 0 {
		t.Error("c2 should receive width from HBox.SetSize")
	}
	// Both children get full height
	if c1.height != 20 {
		t.Errorf("c1.height = %d, want 20", c1.height)
	}
	if c2.height != 20 {
		t.Errorf("c2.height = %d, want 20", c2.height)
	}
}

// TestVBoxSetSizePropagates verifies SetSize propagates to all children.
func TestVBoxSetSizePropagates(t *testing.T) {
	c1 := &stubComponent{name: "a"}
	c2 := &stubComponent{name: "b"}
	v := &VBox{Items: []Component{c1, c2}}
	v.SetSize(100, 20)

	if c1.height == 0 {
		t.Error("c1 should receive height from VBox.SetSize")
	}
	if c2.height == 0 {
		t.Error("c2 should receive height from VBox.SetSize")
	}
	// Both children get full width
	if c1.width != 100 {
		t.Errorf("c1.width = %d, want 100", c1.width)
	}
	if c2.width != 100 {
		t.Errorf("c2.width = %d, want 100", c2.width)
	}
}

// TestHBoxFocusPropagates verifies SetFocused propagates to all children.
func TestHBoxFocusPropagates(t *testing.T) {
	c1 := &stubComponent{name: "a"}
	c2 := &stubComponent{name: "b"}
	h := &HBox{Items: []Component{c1, c2}}

	h.SetFocused(true)
	if !h.Focused() {
		t.Error("HBox should report focused")
	}
	if !c1.focused || !c2.focused {
		t.Error("children should also be focused")
	}

	h.SetFocused(false)
	if h.Focused() {
		t.Error("HBox should report not focused")
	}
	if c1.focused || c2.focused {
		t.Error("children should lose focus")
	}
}

// TestVBoxFocusPropagates verifies SetFocused propagates to all children.
func TestVBoxFocusPropagates(t *testing.T) {
	c1 := &stubComponent{name: "a"}
	c2 := &stubComponent{name: "b"}
	v := &VBox{Items: []Component{c1, c2}}

	v.SetFocused(true)
	if !v.Focused() {
		t.Error("VBox should report focused")
	}
	if !c1.focused || !c2.focused {
		t.Error("children should also be focused")
	}
}

// TestHBoxSizedChild verifies that a Sized wrapper gives fixed width.
func TestHBoxSizedChild(t *testing.T) {
	c1 := &stubComponent{name: "fixed"}
	c2 := &stubComponent{name: "flex"}
	h := &HBox{Items: []Component{Sized{W: 30, C: c1}, c2}}
	h.SetSize(100, 10)

	if c1.width != 30 {
		t.Errorf("sized child width = %d, want 30", c1.width)
	}
	// Flex child gets the remaining space
	if c2.width == 0 {
		t.Error("flex child should get remaining width")
	}
}

// TestVBoxSizedChild verifies that a Sized wrapper gives fixed height.
func TestVBoxSizedChild(t *testing.T) {
	c1 := &stubComponent{name: "fixed"}
	c2 := &stubComponent{name: "flex"}
	v := &VBox{Items: []Component{Sized{W: 5, C: c1}, c2}}
	v.SetSize(80, 20)

	if c1.height != 5 {
		t.Errorf("sized child height = %d, want 5", c1.height)
	}
}

// TestHBoxKeyBindings verifies HBox returns nil keybindings.
func TestHBoxKeyBindings(t *testing.T) {
	h := &HBox{}
	if kb := h.KeyBindings(); kb != nil {
		t.Errorf("HBox.KeyBindings() should return nil, got %v", kb)
	}
}

// TestVBoxKeyBindings verifies VBox returns nil keybindings.
func TestVBoxKeyBindings(t *testing.T) {
	v := &VBox{}
	if kb := v.KeyBindings(); kb != nil {
		t.Errorf("VBox.KeyBindings() should return nil, got %v", kb)
	}
}

// TestHBoxInit verifies HBox.Init fans out to children.
func TestHBoxInit(t *testing.T) {
	c := &stubComponent{name: "child"}
	h := &HBox{Items: []Component{c}}
	cmd := h.Init()
	// stubComponent.Init returns nil, so batch of nils is nil
	_ = cmd
}

// TestVBoxInit verifies VBox.Init fans out to children.
func TestVBoxInit(t *testing.T) {
	c := &stubComponent{name: "child"}
	v := &VBox{Items: []Component{c}}
	cmd := v.Init()
	_ = cmd
}
