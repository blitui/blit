package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Node is a single node in a Tree.
type Node struct {
	// Title is the display label for this node.
	Title string

	// Glyph is an optional icon prefix shown before the title.
	Glyph string

	// Children are the child nodes. Non-nil means this node is expandable.
	Children []*Node

	// Data is an arbitrary payload the caller can attach.
	Data any

	// Expanded controls whether children are visible.
	Expanded bool
}

// TreeOpts configures Tree behaviour.
type TreeOpts struct {
	// OnSelect is called when the user presses Enter on a node. Optional.
	OnSelect func(node *Node)

	// OnToggle is called when the user expands or collapses a node. Optional.
	OnToggle func(node *Node)
}

// Tree is a recursive Component that renders a tree of Nodes with indent
// connector glyphs pulled from the theme glyph pack.
type Tree struct {
	opts  TreeOpts
	roots []*Node

	// flat is the ordered list of visible nodes for keyboard navigation.
	flat    []flatNode
	cursor  int
	theme   Theme
	focused bool
	width   int
	height  int
	scroll  int
}

// flatNode is an entry in the linearised visible-node list.
type flatNode struct {
	node   *Node
	depth  int
	isLast bool
	prefix string
}

// NewTree creates a Tree with the given root nodes and options.
func NewTree(roots []*Node, opts TreeOpts) *Tree {
	t := &Tree{
		opts:  opts,
		roots: roots,
	}
	t.rebuild()
	return t
}

// SetRoots replaces the root nodes and rebuilds the flat view.
func (t *Tree) SetRoots(roots []*Node) {
	t.roots = roots
	t.rebuild()
	if t.cursor >= len(t.flat) {
		t.cursor = max(0, len(t.flat)-1)
	}
}

// Roots returns the root nodes.
func (t *Tree) Roots() []*Node { return t.roots }

// CursorNode returns the currently highlighted node, or nil if the tree is empty.
func (t *Tree) CursorNode() *Node {
	if t.cursor >= 0 && t.cursor < len(t.flat) {
		return t.flat[t.cursor].node
	}
	return nil
}

// Init implements Component.
func (t *Tree) Init() tea.Cmd { return nil }

// Update implements Component.
func (t *Tree) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !t.focused {
		return t, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := t.handleKey(msg)
		return t, cmd
	}
	return t, nil
}

func (t *Tree) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if t.cursor > 0 {
			t.cursor--
			t.clampScroll()
		}
		return Consumed()

	case "down", "j":
		if t.cursor < len(t.flat)-1 {
			t.cursor++
			t.clampScroll()
		}
		return Consumed()

	case "right", "l":
		node := t.CursorNode()
		if node != nil && len(node.Children) > 0 && !node.Expanded {
			node.Expanded = true
			t.rebuild()
			if t.opts.OnToggle != nil {
				t.opts.OnToggle(node)
			}
		}
		return Consumed()

	case "left", "h":
		node := t.CursorNode()
		if node != nil && node.Expanded {
			node.Expanded = false
			t.rebuild()
			if t.opts.OnToggle != nil {
				t.opts.OnToggle(node)
			}
		}
		return Consumed()

	case "enter":
		node := t.CursorNode()
		if node != nil && t.opts.OnSelect != nil {
			t.opts.OnSelect(node)
		}
		return Consumed()

	case " ":
		node := t.CursorNode()
		if node != nil && len(node.Children) > 0 {
			node.Expanded = !node.Expanded
			t.rebuild()
			if t.opts.OnToggle != nil {
				t.opts.OnToggle(node)
			}
		}
		return Consumed()
	}
	return nil
}

// View implements Component.
func (t *Tree) View() string {
	if t.width == 0 || t.height == 0 {
		return ""
	}

	g := t.theme.glyphsOrDefault()

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.theme.TextInverse)).
		Background(lipgloss.Color(t.theme.Cursor)).
		Width(t.width)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.theme.Text)).
		Width(t.width)
	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.theme.Muted))
	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.theme.Accent))

	end := t.scroll + t.height
	if end > len(t.flat) {
		end = len(t.flat)
	}

	var lines []string
	for i := t.scroll; i < end; i++ {
		fn := t.flat[i]
		node := fn.node
		isCursor := i == t.cursor

		// Build connector glyph.
		var connector string
		if fn.depth == 0 {
			connector = ""
		} else if fn.isLast {
			connector = fn.prefix + g.TreeLast + " "
		} else {
			connector = fn.prefix + g.TreeBranch + " "
		}

		// Build expand/collapse arrow or leaf indicator.
		var arrow string
		if len(node.Children) > 0 {
			if node.Expanded {
				arrow = g.ExpandedArrow
			} else {
				arrow = g.CollapsedArrow
			}
		} else {
			arrow = g.Dot
		}

		glyph := node.Glyph
		if glyph != "" {
			glyph = glyph + " "
		}

		label := connector + arrow + " " + glyph + node.Title

		var line string
		if isCursor {
			line = accentStyle.Render(arrow) + cursorStyle.Render(" "+glyph+node.Title)
			// re-render with connector prepended (not styled so connector remains muted)
			prefixStyled := mutedStyle.Render(connector)
			line = prefixStyled + accentStyle.Render(arrow) + cursorStyle.Render(" "+glyph+node.Title)
			_ = label
		} else {
			line = mutedStyle.Render(connector+arrow+" ") + normalStyle.Render(glyph+node.Title)
		}

		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Muted)).Render("  (empty)")
	}

	return strings.Join(lines, "\n")
}

// KeyBindings implements Component.
func (t *Tree) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/k", Label: "Move up", Group: "TREE"},
		{Key: "down/j", Label: "Move down", Group: "TREE"},
		{Key: "right/l", Label: "Expand", Group: "TREE"},
		{Key: "left/h", Label: "Collapse", Group: "TREE"},
		{Key: "enter", Label: "Select", Group: "TREE"},
		{Key: "space", Label: "Toggle expand", Group: "TREE"},
	}
}

// SetSize implements Component.
func (t *Tree) SetSize(w, h int) {
	t.width = w
	t.height = h
	t.clampScroll()
}

// Focused implements Component.
func (t *Tree) Focused() bool { return t.focused }

// SetFocused implements Component.
func (t *Tree) SetFocused(f bool) { t.focused = f }

// SetTheme implements Themed.
func (t *Tree) SetTheme(theme Theme) { t.theme = theme }

// rebuild linearises all currently visible nodes into t.flat.
func (t *Tree) rebuild() {
	t.flat = t.flat[:0]
	for i, root := range t.roots {
		isLast := i == len(t.roots)-1
		t.buildFlat(root, 0, isLast, "")
	}
}

func (t *Tree) buildFlat(node *Node, depth int, isLast bool, prefix string) {
	t.flat = append(t.flat, flatNode{
		node:   node,
		depth:  depth,
		isLast: isLast,
		prefix: prefix,
	})
	if node.Expanded {
		g := t.theme.glyphsOrDefault()
		childPrefix := prefix
		if depth > 0 {
			if isLast {
				childPrefix = prefix + g.TreeEmpty + " "
			} else {
				childPrefix = prefix + g.TreePipe + " "
			}
		}
		for i, child := range node.Children {
			childIsLast := i == len(node.Children)-1
			t.buildFlat(child, depth+1, childIsLast, childPrefix)
		}
	}
}

func (t *Tree) clampScroll() {
	if t.height <= 0 {
		return
	}
	if t.cursor < t.scroll {
		t.scroll = t.cursor
	}
	if t.cursor >= t.scroll+t.height {
		t.scroll = t.cursor - t.height + 1
	}
	if t.scroll < 0 {
		t.scroll = 0
	}
}
