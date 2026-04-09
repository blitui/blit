package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Orientation controls whether Split divides the space horizontally or vertically.
type Orientation int

const (
	// Horizontal splits left/right (A on left, B on right).
	Horizontal Orientation = iota
	// Vertical splits top/bottom (A on top, B on bottom).
	Vertical
)

// Split is a Component that divides its space between two child components
// with a visible divider. When Resizable is true the divider can be moved
// with alt+arrow keys (or mouse drag when mouse support is enabled).
type Split struct {
	// Orientation controls whether the split is left/right or top/bottom.
	Orientation Orientation

	// Ratio is the fraction of space given to pane A (0.0–1.0, default 0.5).
	Ratio float64

	// Resizable enables alt+arrow key (and mouse drag) divider movement.
	Resizable bool

	// A is the first child component (left or top).
	A Component

	// B is the second child component (right or bottom).
	B Component

	theme   Theme
	focused bool
	width   int
	height  int

	// focusA tracks which pane has keyboard focus (tab toggles).
	focusA bool
}

// NewSplit creates a Split with the given orientation, ratio, and children.
func NewSplit(o Orientation, ratio float64, a, b Component) *Split {
	if ratio <= 0 || ratio >= 1 {
		ratio = 0.5
	}
	return &Split{
		Orientation: o,
		Ratio:       ratio,
		A:           a,
		B:           b,
		focusA:      true,
	}
}

// --- size helpers ---

func (s *Split) sizeA() (w, h int) {
	switch s.Orientation {
	case Horizontal:
		w = int(float64(s.width-1) * s.Ratio)
		h = s.height
	case Vertical:
		w = s.width
		h = int(float64(s.height-1) * s.Ratio)
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return
}

func (s *Split) sizeB() (w, h int) {
	aw, ah := s.sizeA()
	switch s.Orientation {
	case Horizontal:
		w = s.width - aw - 1 // 1 for divider
		h = s.height
	case Vertical:
		w = s.width
		h = s.height - ah - 1 // 1 for divider
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return
}

func (s *Split) distributeSize() {
	if s.A != nil {
		aw, ah := s.sizeA()
		s.A.SetSize(aw, ah)
	}
	if s.B != nil {
		bw, bh := s.sizeB()
		s.B.SetSize(bw, bh)
	}
}

// --- Component interface ---

func (s *Split) Init() tea.Cmd {
	var cmds []tea.Cmd
	if s.A != nil {
		cmds = append(cmds, s.A.Init())
	}
	if s.B != nil {
		cmds = append(cmds, s.B.Init())
	}
	return tea.Batch(cmds...)
}

func (s *Split) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := s.handleKey(msg)
		if isConsumed(cmd) {
			return s, cmd
		}
		// Delegate to focused child.
		return s, s.delegateKey(msg, ctx)

	case tea.MouseMsg:
		// Delegate mouse to both children; the one that consumes wins.
		if s.A != nil {
			updated, cmd := s.A.Update(msg, ctx)
			s.A = updated
			if isConsumed(cmd) {
				return s, cmd
			}
		}
		if s.B != nil {
			updated, cmd := s.B.Update(msg, ctx)
			s.B = updated
			if isConsumed(cmd) {
				return s, cmd
			}
		}
	}
	return s, nil
}

func (s *Split) handleKey(msg tea.KeyMsg) tea.Cmd {
	// Tab switches focus between panes.
	if msg.String() == "tab" {
		s.focusA = !s.focusA
		s.syncChildFocus()
		return Consumed()
	}

	if !s.Resizable {
		return nil
	}

	step := 0.05
	switch s.Orientation {
	case Horizontal:
		switch msg.String() {
		case "alt+left":
			s.Ratio -= step
			s.clampRatio()
			s.distributeSize()
			return Consumed()
		case "alt+right":
			s.Ratio += step
			s.clampRatio()
			s.distributeSize()
			return Consumed()
		}
	case Vertical:
		switch msg.String() {
		case "alt+up":
			s.Ratio -= step
			s.clampRatio()
			s.distributeSize()
			return Consumed()
		case "alt+down":
			s.Ratio += step
			s.clampRatio()
			s.distributeSize()
			return Consumed()
		}
	}
	return nil
}

func (s *Split) delegateKey(msg tea.KeyMsg, ctx Context) tea.Cmd {
	if s.focusA && s.A != nil {
		updated, cmd := s.A.Update(msg, ctx)
		s.A = updated
		return cmd
	}
	if !s.focusA && s.B != nil {
		updated, cmd := s.B.Update(msg, ctx)
		s.B = updated
		return cmd
	}
	return nil
}

func (s *Split) syncChildFocus() {
	if s.A != nil {
		s.A.SetFocused(s.focusA && s.focused)
	}
	if s.B != nil {
		s.B.SetFocused(!s.focusA && s.focused)
	}
}

func (s *Split) View() string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	aView := ""
	bView := ""
	if s.A != nil {
		aView = s.A.View()
	}
	if s.B != nil {
		bView = s.B.View()
	}

	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Border))

	switch s.Orientation {
	case Horizontal:
		return s.viewHorizontal(aView, bView, dividerStyle)
	case Vertical:
		return s.viewVertical(aView, bView, dividerStyle)
	}
	return ""
}

func (s *Split) viewHorizontal(aView, bView string, divStyle lipgloss.Style) string {
	aw, ah := s.sizeA()
	bw, bh := s.sizeB()
	maxH := max(ah, bh)

	aLines := padLines(strings.Split(aView, "\n"), aw, maxH)
	bLines := padLines(strings.Split(bView, "\n"), bw, maxH)

	// Build divider column (one character wide).
	divLines := make([]string, maxH)
	for i := range divLines {
		divLines[i] = divStyle.Render("│")
	}

	var rows []string
	for i := 0; i < maxH; i++ {
		rows = append(rows, aLines[i]+divLines[i]+bLines[i])
	}
	return strings.Join(rows, "\n")
}

func (s *Split) viewVertical(aView, bView string, divStyle lipgloss.Style) string {
	// Build divider row (full width, one line).
	divRow := divStyle.Render(strings.Repeat("─", s.width))

	var parts []string
	if aView != "" {
		parts = append(parts, aView)
	}
	parts = append(parts, divRow)
	if bView != "" {
		parts = append(parts, bView)
	}
	return strings.Join(parts, "\n")
}

// padLines pads/trims a slice of lines to exactly h lines, each w columns wide.
func padLines(lines []string, w, h int) []string {
	out := make([]string, h)
	for i := 0; i < h; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		vis := lipgloss.Width(line)
		if vis < w {
			line += strings.Repeat(" ", w-vis)
		} else if vis > w {
			line = truncateLine(line, w)
		}
		out[i] = line
	}
	return out
}

func (s *Split) KeyBindings() []KeyBind {
	binds := []KeyBind{
		{Key: "tab", Label: "Switch pane focus", Group: "SPLIT"},
	}
	if s.Resizable {
		switch s.Orientation {
		case Horizontal:
			binds = append(binds,
				KeyBind{Key: "alt+left", Label: "Shrink left pane", Group: "SPLIT"},
				KeyBind{Key: "alt+right", Label: "Grow left pane", Group: "SPLIT"},
			)
		case Vertical:
			binds = append(binds,
				KeyBind{Key: "alt+up", Label: "Shrink top pane", Group: "SPLIT"},
				KeyBind{Key: "alt+down", Label: "Grow top pane", Group: "SPLIT"},
			)
		}
	}
	return binds
}

func (s *Split) SetSize(w, h int) {
	s.width = w
	s.height = h
	s.distributeSize()
}

func (s *Split) Focused() bool { return s.focused }

func (s *Split) SetFocused(f bool) {
	s.focused = f
	s.syncChildFocus()
}

func (s *Split) SetTheme(th Theme) {
	s.theme = th
	if t, ok := s.A.(Themed); ok {
		t.SetTheme(th)
	}
	if t, ok := s.B.(Themed); ok {
		t.SetTheme(th)
	}
}

func (s *Split) clampRatio() {
	if s.Ratio < 0.1 {
		s.Ratio = 0.1
	}
	if s.Ratio > 0.9 {
		s.Ratio = 0.9
	}
}
