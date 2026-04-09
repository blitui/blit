package tuikit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// flexStub is a minimal Component for flex layout tests.
// Named distinctly to avoid conflicts with stubComponent in app_test.go.
type flexStub struct {
	label   string
	width   int
	height  int
	focused bool
}

func newFlexStub(label string) *flexStub { return &flexStub{label: label} }

func (s *flexStub) Init() tea.Cmd                           { return nil }
func (s *flexStub) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) { return s, nil }
func (s *flexStub) View() string {
	line := s.label
	if s.width > 0 && len(line) > s.width {
		line = line[:s.width]
	}
	return line
}
func (s *flexStub) KeyBindings() []KeyBind  { return nil }
func (s *flexStub) SetSize(w, h int)        { s.width = w; s.height = h }
func (s *flexStub) Focused() bool           { return s.focused }
func (s *flexStub) SetFocused(focused bool) { s.focused = focused }

// --- HBox tests ---

func TestHBoxGap(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &HBox{
		Gap:   2,
		Items: []Component{Sized{W: 10, C: a}, Sized{W: 10, C: b}},
	}
	box.SetSize(24, 1)
	view := box.View()
	if !strings.Contains(view, "A") || !strings.Contains(view, "B") {
		t.Errorf("HBox gap view missing children: %q", view)
	}
}

func TestHBoxFlexGrow(t *testing.T) {
	a := newFlexStub("flex1")
	b := newFlexStub("flex2")
	box := &HBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(40, 5)
	if a.width != 20 {
		t.Errorf("expected flex child width 20, got %d", a.width)
	}
	if b.width != 20 {
		t.Errorf("expected flex child width 20, got %d", b.width)
	}
}

func TestHBoxFlexGrowUnequal(t *testing.T) {
	a := newFlexStub("grow1")
	b := newFlexStub("grow2")
	box := &HBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: a}, Flex{Grow: 3, C: b}},
	}
	box.SetSize(40, 5)
	if a.width != 10 {
		t.Errorf("expected grow=1 child width 10, got %d", a.width)
	}
	if b.width != 30 {
		t.Errorf("expected grow=3 child width 30, got %d", b.width)
	}
}

func TestHBoxSizedChildren(t *testing.T) {
	a := newFlexStub("fixed")
	b := newFlexStub("flex")
	box := &HBox{
		Gap:   0,
		Items: []Component{Sized{W: 15, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(40, 5)
	if a.width != 15 {
		t.Errorf("expected Sized child width 15, got %d", a.width)
	}
	if b.width != 25 {
		t.Errorf("expected Flex child width 25, got %d", b.width)
	}
}

func TestHBoxJustifyStart(t *testing.T) {
	a := newFlexStub("A")
	box := &HBox{
		Gap:     0,
		Justify: FlexJustifyStart,
		Items:   []Component{Sized{W: 5, C: a}},
	}
	box.SetSize(20, 1)
	view := box.View()
	if !strings.HasPrefix(strings.Split(view, "\n")[0], "A") {
		t.Errorf("JustifyStart: expected A at start, got %q", view)
	}
}

func TestHBoxJustifyEnd(t *testing.T) {
	a := newFlexStub("A")
	box := &HBox{
		Gap:     0,
		Justify: FlexJustifyEnd,
		Items:   []Component{Sized{W: 1, C: a}},
	}
	box.SetSize(10, 1)
	view := box.View()
	firstLine := strings.Split(view, "\n")[0]
	if !strings.HasSuffix(strings.TrimRight(firstLine, " "), "A") {
		t.Errorf("JustifyEnd: expected A near end, got %q", firstLine)
	}
}

func TestHBoxJustifyCenter(t *testing.T) {
	a := newFlexStub("A")
	box := &HBox{
		Gap:     0,
		Justify: FlexJustifyCenter,
		Items:   []Component{Sized{W: 1, C: a}},
	}
	box.SetSize(11, 1)
	view := box.View()
	firstLine := strings.Split(view, "\n")[0]
	trimmed := strings.TrimLeft(firstLine, " ")
	leading := len(firstLine) - len(trimmed)
	if leading < 4 {
		t.Errorf("JustifyCenter: expected leading spaces ≥4, got %d in %q", leading, firstLine)
	}
}

func TestHBoxJustifySpaceBetween(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &HBox{
		Gap:     0,
		Justify: FlexJustifySpaceBetween,
		Items:   []Component{Sized{W: 1, C: a}, Sized{W: 1, C: b}},
	}
	box.SetSize(10, 1)
	view := box.View()
	firstLine := strings.Split(view, "\n")[0]
	idxA := strings.Index(firstLine, "A")
	idxB := strings.Index(firstLine, "B")
	if idxA < 0 || idxB < 0 {
		t.Fatalf("SpaceBetween: missing children in %q", firstLine)
	}
	if idxB-idxA < 2 {
		t.Errorf("SpaceBetween: not enough space between A and B in %q", firstLine)
	}
}

func TestHBoxJustifySpaceAround(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &HBox{
		Gap:     0,
		Justify: FlexJustifySpaceAround,
		Items:   []Component{Sized{W: 1, C: a}, Sized{W: 1, C: b}},
	}
	box.SetSize(10, 1)
	view := box.View()
	firstLine := strings.Split(view, "\n")[0]
	if !strings.Contains(firstLine, "A") || !strings.Contains(firstLine, "B") {
		t.Errorf("SpaceAround: missing children in %q", firstLine)
	}
}

func TestHBoxAlignStart(t *testing.T) {
	a := newFlexStub("A")
	box := &HBox{
		Align: FlexAlignStart,
		Items: []Component{Sized{W: 5, C: a}},
	}
	box.SetSize(20, 5)
	view := box.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 1 || !strings.Contains(lines[0], "A") {
		t.Errorf("AlignStart: A should be in first line, got lines=%v", lines)
	}
}

func TestHBoxAlignStretch(t *testing.T) {
	a := newFlexStub("A")
	box := &HBox{
		Align: FlexAlignStretch,
		Items: []Component{Sized{W: 5, C: a}},
	}
	box.SetSize(20, 5)
	box.View()
	if a.height != 5 {
		t.Errorf("AlignStretch: expected height 5, got %d", a.height)
	}
}

// --- VBox tests ---

func TestVBoxGap(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &VBox{
		Gap:   2,
		Items: []Component{Sized{W: 1, C: a}, Sized{W: 1, C: b}},
	}
	box.SetSize(10, 10)
	view := box.View()
	lines := strings.Split(view, "\n")
	hasBlank := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			hasBlank = true
			break
		}
	}
	if !hasBlank {
		t.Errorf("VBox gap: expected blank lines, got %v", lines)
	}
}

func TestVBoxFlexGrow(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &VBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(10, 20)
	if a.height != 10 {
		t.Errorf("expected flex child height 10, got %d", a.height)
	}
	if b.height != 10 {
		t.Errorf("expected flex child height 10, got %d", b.height)
	}
}

func TestVBoxSizedChildren(t *testing.T) {
	a := newFlexStub("fixed")
	b := newFlexStub("flex")
	box := &VBox{
		Gap:   0,
		Items: []Component{Sized{W: 5, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(10, 20)
	if a.height != 5 {
		t.Errorf("expected Sized child height 5, got %d", a.height)
	}
	if b.height != 15 {
		t.Errorf("expected Flex child height 15, got %d", b.height)
	}
}

func TestVBoxJustifyStart(t *testing.T) {
	a := newFlexStub("A")
	box := &VBox{
		Justify: FlexJustifyStart,
		Items:   []Component{Sized{W: 1, C: a}},
	}
	box.SetSize(10, 5)
	view := box.View()
	lines := strings.Split(view, "\n")
	if len(lines) == 0 || !strings.Contains(lines[0], "A") {
		t.Errorf("VBox JustifyStart: A should be first line, got %v", lines)
	}
}

func TestVBoxJustifyCenter(t *testing.T) {
	a := newFlexStub("A")
	box := &VBox{
		Justify: FlexJustifyCenter,
		Items:   []Component{Sized{W: 1, C: a}},
	}
	box.SetSize(10, 9)
	view := box.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "A") {
		t.Errorf("VBox JustifyCenter: A should not be on first line, got %v", lines)
	}
}

func TestVBoxJustifyEnd(t *testing.T) {
	a := newFlexStub("A")
	box := &VBox{
		Justify: FlexJustifyEnd,
		Items:   []Component{Sized{W: 1, C: a}},
	}
	box.SetSize(10, 5)
	view := box.View()
	lines := strings.Split(view, "\n")
	last := lines[len(lines)-1]
	if !strings.Contains(last, "A") {
		t.Errorf("VBox JustifyEnd: A should be last line, got %v", lines)
	}
}

func TestVBoxJustifySpaceBetween(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &VBox{
		Justify: FlexJustifySpaceBetween,
		Items:   []Component{Sized{W: 1, C: a}, Sized{W: 1, C: b}},
	}
	box.SetSize(10, 10)
	view := box.View()
	lines := strings.Split(view, "\n")
	idxA, idxB := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "A") {
			idxA = i
		}
		if strings.Contains(l, "B") {
			idxB = i
		}
	}
	if idxA < 0 || idxB < 0 {
		t.Fatalf("VBox SpaceBetween: missing children, lines=%v", lines)
	}
	if idxB-idxA < 2 {
		t.Errorf("VBox SpaceBetween: expected gap between A and B, idxA=%d idxB=%d", idxA, idxB)
	}
}

// --- Nesting ---

func TestHBoxNestedVBox(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	c := newFlexStub("C")
	inner := &VBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: b}, Flex{Grow: 1, C: c}},
	}
	outer := &HBox{
		Gap:   1,
		Items: []Component{Sized{W: 10, C: a}, Flex{Grow: 1, C: inner}},
	}
	outer.SetSize(30, 10)
	if a.width != 10 {
		t.Errorf("nested: outer fixed child width should be 10, got %d", a.width)
	}
	// inner gets 30 - 10 - 1 gap = 19
	if inner.width != 19 {
		t.Errorf("nested: inner VBox width should be 19, got %d", inner.width)
	}
	// inner children each get height 5
	if b.height != 5 || c.height != 5 {
		t.Errorf("nested: inner children height should be 5, got b=%d c=%d", b.height, c.height)
	}
}

func TestVBoxNestedHBox(t *testing.T) {
	a := newFlexStub("header")
	b := newFlexStub("col1")
	c := newFlexStub("col2")
	inner := &HBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: b}, Flex{Grow: 1, C: c}},
	}
	outer := &VBox{
		Gap:   0,
		Items: []Component{Sized{W: 2, C: a}, Flex{Grow: 1, C: inner}},
	}
	outer.SetSize(20, 12)
	if a.height != 2 {
		t.Errorf("nested: fixed header height should be 2, got %d", a.height)
	}
	if inner.height != 10 {
		t.Errorf("nested: inner HBox height should be 10, got %d", inner.height)
	}
	if b.width != 10 || c.width != 10 {
		t.Errorf("nested: inner cols width should be 10, got b=%d c=%d", b.width, c.width)
	}
}

// --- DualPane composition (CRITICAL: must not break existing DualPane) ---

// TestDualPaneInsideHBox proves existing DualPane still works when composed
// inside an HBox — the flex layout must not break DualPane.
func TestDualPaneInsideHBox(t *testing.T) {
	main := newFlexStub("main")
	side := newFlexStub("side")
	dp := &DualPane{
		Main:         main,
		Side:         side,
		SideWidth:    20,
		MinMainWidth: 40,
		SideRight:    true,
	}

	dpMain, dpSide, visible := dp.compute(80, 20)
	if !visible {
		t.Fatal("DualPane side should be visible at width 80")
	}
	if dpMain.width != 57 { // 80 - 20 - 3
		t.Errorf("DualPane main width: expected 57, got %d", dpMain.width)
	}
	if dpSide.width != 20 {
		t.Errorf("DualPane side width: expected 20, got %d", dpSide.width)
	}

	// Wrap a "dashboard" stub inside HBox to prove flex + DualPane coexist
	dashboard := newFlexStub("dashboard")
	toolbar := newFlexStub("toolbar")
	hbox := &HBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: dashboard}, Sized{W: 20, C: toolbar}},
	}
	hbox.SetSize(100, 30)
	if dashboard.width != 80 {
		t.Errorf("HBox flex child (dashboard) width: expected 80, got %d", dashboard.width)
	}
	if toolbar.width != 20 {
		t.Errorf("HBox fixed child (toolbar) width: expected 20, got %d", toolbar.width)
	}
}

// --- Resize ---

func TestHBoxResize(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &HBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(40, 10)
	if a.width != 20 || b.width != 20 {
		t.Fatalf("initial: expected 20/20, got %d/%d", a.width, b.width)
	}
	box.SetSize(60, 10)
	if a.width != 30 || b.width != 30 {
		t.Errorf("after resize: expected 30/30, got %d/%d", a.width, b.width)
	}
}

func TestVBoxResize(t *testing.T) {
	a := newFlexStub("A")
	b := newFlexStub("B")
	box := &VBox{
		Gap:   0,
		Items: []Component{Flex{Grow: 1, C: a}, Flex{Grow: 1, C: b}},
	}
	box.SetSize(10, 20)
	if a.height != 10 || b.height != 10 {
		t.Fatalf("initial: expected 10/10, got %d/%d", a.height, b.height)
	}
	box.SetSize(10, 40)
	if a.height != 20 || b.height != 20 {
		t.Errorf("after resize: expected 20/20, got %d/%d", a.height, b.height)
	}
}
