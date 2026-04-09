package tuikit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// splitStub is a minimal Component for split tests.
type splitStub struct {
	view    string
	focused bool
	width   int
	height  int
}

func (s *splitStub) Init() tea.Cmd { return nil }
func (s *splitStub) Update(msg tea.Msg, ctx tuikit.Context) (tuikit.Component, tea.Cmd) {
	return s, nil
}
func (s *splitStub) View() string                  { return s.view }
func (s *splitStub) KeyBindings() []tuikit.KeyBind { return nil }
func (s *splitStub) SetSize(w, h int)              { s.width = w; s.height = h }
func (s *splitStub) Focused() bool                 { return s.focused }
func (s *splitStub) SetFocused(f bool)             { s.focused = f }

func newSplitStub(view string) *splitStub { return &splitStub{view: view} }

func TestSplitComponentInterface(t *testing.T) {
	var _ tuikit.Component = tuikit.NewSplit(tuikit.Horizontal, 0.5, newSplitStub("a"), newSplitStub("b"))
}

func TestSplitHorizontalSizeDistribution(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, a, b)
	s.SetSize(40, 10)

	// With width=40 and ratio=0.5: a gets ~19, divider=1, b gets ~20.
	if a.width+b.width+1 != 40 {
		t.Errorf("expected widths to sum to 39+1=40, got a=%d b=%d", a.width, b.width)
	}
	if a.height != 10 || b.height != 10 {
		t.Errorf("expected height=10 for both, got a=%d b=%d", a.height, b.height)
	}
}

func TestSplitVerticalSizeDistribution(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Vertical, 0.5, a, b)
	s.SetSize(40, 20)

	// With height=20 and ratio=0.5: a gets ~9, divider=1, b gets ~10.
	if a.height+b.height+1 != 20 {
		t.Errorf("expected heights to sum to 19+1=20, got a=%d b=%d", a.height, b.height)
	}
	if a.width != 40 || b.width != 40 {
		t.Errorf("expected width=40 for both, got a=%d b=%d", a.width, b.width)
	}
}

func TestSplitHorizontalViewContainsDivider(t *testing.T) {
	a := newSplitStub(strings.Repeat("X", 5))
	b := newSplitStub(strings.Repeat("Y", 5))
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, a, b)
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(20, 1)
	view := s.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
	// The divider character │ should appear.
	if !strings.Contains(view, "│") {
		t.Errorf("expected vertical divider │ in horizontal split, got: %q", view)
	}
}

func TestSplitVerticalViewContainsDivider(t *testing.T) {
	a := newSplitStub("top")
	b := newSplitStub("bottom")
	s := tuikit.NewSplit(tuikit.Vertical, 0.5, a, b)
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(20, 5)
	view := s.View()
	if !strings.Contains(view, "─") {
		t.Errorf("expected horizontal divider ─ in vertical split, got: %q", view)
	}
}

func TestSplitResizableRatioDecrease(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, a, b)
	s.Resizable = true
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 10)

	initialRatio := s.Ratio
	s.Update(tea.KeyMsg{Alt: true, Type: tea.KeyLeft}, tuikit.Context{})
	if s.Ratio >= initialRatio {
		t.Errorf("expected ratio to decrease after alt+left, before=%.2f after=%.2f", initialRatio, s.Ratio)
	}
}

func TestSplitResizableRatioIncrease(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, a, b)
	s.Resizable = true
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 10)

	initialRatio := s.Ratio
	s.Update(tea.KeyMsg{Alt: true, Type: tea.KeyRight}, tuikit.Context{})
	if s.Ratio <= initialRatio {
		t.Errorf("expected ratio to increase after alt+right, before=%.2f after=%.2f", initialRatio, s.Ratio)
	}
}

func TestSplitRatioClampMin(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.15, a, b)
	s.Resizable = true
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 10)

	for i := 0; i < 20; i++ {
		s.Update(tea.KeyMsg{Alt: true, Type: tea.KeyLeft}, tuikit.Context{})
	}
	if s.Ratio < 0.1 {
		t.Errorf("ratio should not go below 0.1, got %.2f", s.Ratio)
	}
}

func TestSplitRatioClampMax(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.85, a, b)
	s.Resizable = true
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 10)

	for i := 0; i < 20; i++ {
		s.Update(tea.KeyMsg{Alt: true, Type: tea.KeyRight}, tuikit.Context{})
	}
	if s.Ratio > 0.9 {
		t.Errorf("ratio should not go above 0.9, got %.2f", s.Ratio)
	}
}

func TestSplitTabSwitchesFocus(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, a, b)
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 10)
	s.SetFocused(true)

	// Initially focusA should be true → a is focused.
	if !a.focused {
		t.Error("expected pane A to be focused initially")
	}

	s.Update(tea.KeyMsg{Type: tea.KeyTab}, tuikit.Context{})
	if a.focused {
		t.Error("expected pane A to lose focus after tab")
	}
	if !b.focused {
		t.Error("expected pane B to gain focus after tab")
	}
}

func TestSplitSetThemeNoPanic(t *testing.T) {
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, newSplitStub("A"), newSplitStub("B"))
	// Should not panic even when children don't implement Themed.
	s.SetTheme(tuikit.DefaultTheme())
}

func TestSplitKeyBindingsResizable(t *testing.T) {
	s := tuikit.NewSplit(tuikit.Horizontal, 0.5, newSplitStub("A"), newSplitStub("B"))
	s.Resizable = true
	binds := s.KeyBindings()
	if len(binds) == 0 {
		t.Error("expected keybindings for resizable split")
	}
}

func TestSplitDefaultRatioInvalid(t *testing.T) {
	s := tuikit.NewSplit(tuikit.Horizontal, 0.0, newSplitStub("A"), newSplitStub("B"))
	if s.Ratio != 0.5 {
		t.Errorf("expected default ratio 0.5 for invalid input, got %.2f", s.Ratio)
	}
}

func TestSplitVerticalResizable(t *testing.T) {
	a := newSplitStub("A")
	b := newSplitStub("B")
	s := tuikit.NewSplit(tuikit.Vertical, 0.5, a, b)
	s.Resizable = true
	s.SetTheme(tuikit.DefaultTheme())
	s.SetSize(40, 20)

	initialRatio := s.Ratio
	s.Update(tea.KeyMsg{Alt: true, Type: tea.KeyDown}, tuikit.Context{})
	if s.Ratio <= initialRatio {
		t.Errorf("expected ratio to increase after alt+down, before=%.2f after=%.2f", initialRatio, s.Ratio)
	}
}
