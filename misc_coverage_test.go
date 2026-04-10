package blit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Tree (SetRoots, Init, Focused) ---

func TestTree_SetRoots(t *testing.T) {
	tr := NewTree(nil, TreeOpts{})
	roots := []*Node{
		{Title: "root1"},
		{Title: "root2"},
	}
	tr.SetRoots(roots)
}

func TestTree_Init(t *testing.T) {
	tr := NewTree([]*Node{{Title: "a"}}, TreeOpts{})
	cmd := tr.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestTree_Focused(t *testing.T) {
	tr := NewTree(nil, TreeOpts{})
	if tr.Focused() {
		t.Error("should not be focused initially")
	}
}

// --- Viewport (Focused, truncateLine) ---

func TestViewport_Focused(t *testing.T) {
	vp := NewViewport()
	if vp.Focused() {
		t.Error("should not be focused initially")
	}
}

func TestTruncateLine(t *testing.T) {
	result := truncateLine("hello world", 5)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestTruncateLineNoOp(t *testing.T) {
	result := truncateLine("hi", 10)
	if result != "hi" {
		t.Errorf("expected 'hi' unchanged, got %q", result)
	}
}

// --- Picker (Items, Focused) ---

func TestPicker_Items(t *testing.T) {
	items := []PickerItem{
		{Title: "one"},
		{Title: "two"},
	}
	p := NewPicker(items, PickerOpts{})
	got := p.Items()
	if len(got) != 2 {
		t.Errorf("expected 2 items, got %d", len(got))
	}
}

func TestPicker_Focused(t *testing.T) {
	p := NewPicker(nil, PickerOpts{})
	if p.Focused() {
		t.Error("should not be focused initially")
	}
}

// --- Split (Focused, delegateKey) ---

func TestSplit_Focused(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Horizontal, 0.5, a, b)
	if s.Focused() {
		t.Error("should not be focused initially")
	}
}

func TestSplit_DelegateKey(t *testing.T) {
	a := &stubComponent{name: "a"}
	b := &stubComponent{name: "b"}
	s := NewSplit(Horizontal, 0.5, a, b)
	s.SetTheme(DefaultTheme())
	s.SetSize(80, 24)

	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 80, Height: 24}}
	cmd := s.delegateKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, ctx)
	_ = cmd
}

// --- LogViewer (LogAppendCmd) ---

func TestLogAppendCmd(t *testing.T) {
	cmd := LogAppendCmd(LogLine{Message: "test", Timestamp: time.Now()})
	if cmd == nil {
		t.Fatal("LogAppendCmd should return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(LogAppendMsg); !ok {
		t.Errorf("expected LogAppendMsg, got %T", msg)
	}
}

// --- Glyphs ---

func TestAsciiGlyphs(t *testing.T) {
	g := AsciiGlyphs()
	if g.TreeBranch == "" {
		t.Error("AsciiGlyphs should have non-empty TreeBranch")
	}
	if g.TreeLast == "" {
		t.Error("AsciiGlyphs should have non-empty TreeLast")
	}
}

// --- theme_importers clampF ---

func TestClampF(t *testing.T) {
	tests := []struct {
		v, want float64
	}{
		{0.5, 0.5},
		{-1, 0},
		{2, 1},
		{0, 0},
		{1, 1},
	}
	for _, tt := range tests {
		got := clampF(tt.v)
		if got != tt.want {
			t.Errorf("clampF(%g) = %g, want %g", tt.v, got, tt.want)
		}
	}
}

// --- updater_ratelimit Error ---

func TestRateLimitError_Error(t *testing.T) {
	e := &RateLimitError{StatusCode: 429}
	s := e.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}
}
