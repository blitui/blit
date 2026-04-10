package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Sized wrapper (all 0%) ---

func TestSized_Init(t *testing.T) {
	c := &stubComponent{name: "inner"}
	s := Sized{W: 20, C: c}
	cmd := s.Init()
	if cmd != nil {
		t.Error("stubComponent Init returns nil, so Sized.Init should too")
	}
}

func TestSized_Update(t *testing.T) {
	c := &stubComponent{name: "inner"}
	s := Sized{W: 20, C: c}

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}
	result, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, ctx)
	_ = cmd

	sized, ok := result.(Sized)
	if !ok {
		t.Fatalf("expected Sized, got %T", result)
	}
	if sized.W != 20 {
		t.Errorf("expected W=20, got %d", sized.W)
	}
}

func TestSized_View(t *testing.T) {
	c := &stubComponent{name: "hello"}
	s := Sized{W: 20, C: c}
	if s.View() != "hello" {
		t.Errorf("expected 'hello', got %q", s.View())
	}
}

func TestSized_KeyBindings(t *testing.T) {
	c := &stubComponent{name: "x", bindings: []KeyBind{{Key: "a"}}}
	s := Sized{W: 10, C: c}
	if len(s.KeyBindings()) != 1 {
		t.Errorf("expected 1 keybind, got %d", len(s.KeyBindings()))
	}
}

func TestSized_SetSize(t *testing.T) {
	c := &stubComponent{name: "x"}
	s := Sized{W: 10, C: c}
	s.SetSize(50, 30)
	if c.width != 50 || c.height != 30 {
		t.Errorf("expected 50x30, got %dx%d", c.width, c.height)
	}
}

func TestSized_Focus(t *testing.T) {
	c := &stubComponent{name: "x"}
	s := Sized{W: 10, C: c}

	if s.Focused() {
		t.Error("should not be focused initially")
	}
	s.SetFocused(true)
	if !c.focused {
		t.Error("inner component should be focused")
	}
	if !s.Focused() {
		t.Error("Sized should report focused")
	}
}

// --- Flex wrapper (all 0%) ---

func TestFlex_Init(t *testing.T) {
	c := &stubComponent{name: "inner"}
	f := Flex{Grow: 1, C: c}
	cmd := f.Init()
	if cmd != nil {
		t.Error("stubComponent Init returns nil, so Flex.Init should too")
	}
}

func TestFlex_Update(t *testing.T) {
	c := &stubComponent{name: "inner"}
	f := Flex{Grow: 2, C: c}

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}
	result, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}, ctx)

	flex, ok := result.(Flex)
	if !ok {
		t.Fatalf("expected Flex, got %T", result)
	}
	if flex.Grow != 2 {
		t.Errorf("expected Grow=2, got %d", flex.Grow)
	}
}

func TestFlex_View(t *testing.T) {
	c := &stubComponent{name: "world"}
	f := Flex{Grow: 1, C: c}
	if f.View() != "world" {
		t.Errorf("expected 'world', got %q", f.View())
	}
}

func TestFlex_KeyBindings(t *testing.T) {
	c := &stubComponent{name: "x"}
	f := Flex{Grow: 1, C: c}
	if len(f.KeyBindings()) != 0 {
		t.Errorf("expected 0 keybinds, got %d", len(f.KeyBindings()))
	}
}

func TestFlex_SetSize(t *testing.T) {
	c := &stubComponent{name: "x"}
	f := Flex{Grow: 1, C: c}
	f.SetSize(40, 20)
	if c.width != 40 || c.height != 20 {
		t.Errorf("expected 40x20, got %dx%d", c.width, c.height)
	}
}

func TestFlex_Focus(t *testing.T) {
	c := &stubComponent{name: "x"}
	f := Flex{Grow: 1, C: c}
	if f.Focused() {
		t.Error("should not be focused initially")
	}
	f.SetFocused(true)
	if !f.Focused() {
		t.Error("Flex should report focused")
	}
}

func TestFlex_SetTheme(t *testing.T) {
	c := &themedStub{stubComponent: stubComponent{name: "x"}}
	f := Flex{Grow: 1, C: c}
	th := DefaultTheme()
	th.Accent = "#123456"
	f.SetTheme(th)
	if c.lastTheme.Accent != "#123456" {
		t.Error("theme should propagate through Flex")
	}
}

// --- HBox (Update 0%, many low-coverage methods) ---

func TestHBox_Update(t *testing.T) {
	c1 := &stubComponent{name: "a"}
	c2 := &stubComponent{name: "b"}
	h := &HBox{Items: []Component{c1, c2}}

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}
	result, _ := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, ctx)
	if result != h {
		t.Error("HBox.Update should return itself")
	}
}

func TestHBox_ViewEmpty(t *testing.T) {
	h := &HBox{}
	if h.View() != "" {
		t.Error("empty HBox should render empty string")
	}
}

func TestHBox_SetSizeAndView(t *testing.T) {
	c1 := &stubComponent{name: "left"}
	c2 := &stubComponent{name: "right"}
	h := &HBox{
		Items: []Component{
			Sized{W: 20, C: c1},
			Flex{Grow: 1, C: c2},
		},
	}
	h.SetSize(80, 24)

	view := h.View()
	if view == "" {
		t.Error("HBox with items should render non-empty")
	}
}

func TestHBox_Focus(t *testing.T) {
	c := &stubComponent{name: "x"}
	h := &HBox{Items: []Component{c}}

	if h.Focused() {
		t.Error("should not be focused initially")
	}
	h.SetFocused(true)
	if !h.Focused() {
		t.Error("should be focused")
	}
	if !c.focused {
		t.Error("child should be focused")
	}
}

func TestHBox_KeyBindings(t *testing.T) {
	h := &HBox{}
	if h.KeyBindings() != nil {
		t.Error("HBox KeyBindings should return nil")
	}
}

func TestHBox_SetTheme(t *testing.T) {
	c := &themedStub{stubComponent: stubComponent{name: "x"}}
	h := &HBox{Items: []Component{c}}
	th := DefaultTheme()
	th.Accent = "#abcdef"
	h.SetTheme(th)
	if c.lastTheme.Accent != "#abcdef" {
		t.Error("theme should propagate through HBox")
	}
}

func TestHBox_Init(t *testing.T) {
	c := &stubComponent{name: "x"}
	h := &HBox{Items: []Component{c}}
	cmd := h.Init()
	_ = cmd
}

// --- VBox (Update 0%) ---

func TestVBox_Update(t *testing.T) {
	c1 := &stubComponent{name: "top"}
	c2 := &stubComponent{name: "bottom"}
	v := &VBox{Items: []Component{c1, c2}}

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}
	result, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}, ctx)
	if result != v {
		t.Error("VBox.Update should return itself")
	}
}

func TestVBox_ViewEmpty(t *testing.T) {
	v := &VBox{}
	if v.View() != "" {
		t.Error("empty VBox should render empty string")
	}
}

func TestVBox_SetSizeAndView(t *testing.T) {
	c1 := &stubComponent{name: "top"}
	c2 := &stubComponent{name: "bot"}
	v := &VBox{
		Items: []Component{
			Sized{W: 10, C: c1},
			Flex{Grow: 1, C: c2},
		},
	}
	v.SetSize(80, 24)

	view := v.View()
	if view == "" {
		t.Error("VBox with items should render non-empty")
	}
}

func TestVBox_Focus(t *testing.T) {
	c := &stubComponent{name: "x"}
	v := &VBox{Items: []Component{c}}
	v.SetFocused(true)
	if !v.Focused() {
		t.Error("should be focused")
	}
}

func TestVBox_Init(t *testing.T) {
	c := &stubComponent{name: "x"}
	v := &VBox{Items: []Component{c}}
	cmd := v.Init()
	_ = cmd
}

// --- rewrapFlexComponent (0%) ---

func TestRewrapFlexComponent(t *testing.T) {
	inner := &stubComponent{name: "orig"}
	updated := &stubComponent{name: "new"}

	tests := []struct {
		name     string
		original Component
	}{
		{"Sized", Sized{W: 10, C: inner}},
		{"*Sized", &Sized{W: 10, C: inner}},
		{"Flex", Flex{Grow: 1, C: inner}},
		{"*Flex", &Flex{Grow: 1, C: inner}},
		{"plain", inner},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewrapFlexComponent(tt.original, updated)
			unwrapped := unwrapFlexComponent(result)
			if unwrapped != updated {
				t.Error("rewrapped component should contain updated inner")
			}
		})
	}
}

// --- alignCrossHBox (29.6%) ---

func TestAlignCrossHBox_Stretch(t *testing.T) {
	result := alignCrossHBox("line1", 10, 5, FlexAlignStretch)
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestAlignCrossHBox_Center(t *testing.T) {
	result := alignCrossHBox("line1", 10, 5, FlexAlignCenter)
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestAlignCrossHBox_End(t *testing.T) {
	result := alignCrossHBox("line1", 10, 5, FlexAlignEnd)
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestAlignCrossHBox_Start(t *testing.T) {
	result := alignCrossHBox("line1", 10, 5, FlexAlignStart)
	if result == "" {
		t.Error("should not be empty")
	}
}

// --- alignCrossVBox (75%) ---

func TestAlignCrossVBox_Center(t *testing.T) {
	result := alignCrossVBox("hello", 20, FlexAlignCenter)
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestAlignCrossVBox_End(t *testing.T) {
	result := alignCrossVBox("hello", 20, FlexAlignEnd)
	if result == "" {
		t.Error("should not be empty")
	}
}

// --- themedStub helper ---

type themedStub struct {
	stubComponent
	lastTheme Theme
}

func (t *themedStub) SetTheme(th Theme) { t.lastTheme = th }
