package blit

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

	// Detail is optional secondary text rendered after the title.
	Detail string

	// Children are the child nodes. Non-nil means this node is expandable.
	Children []*Node

	// Data is an arbitrary payload the caller can attach.
	Data any

	// Expanded controls whether children are visible.
	Expanded bool

	// Selected marks this node as selected (used in Single/Multi selection modes).
	Selected bool
}

// SelectionMode controls how node selection behaves in a Tree.
type SelectionMode int

const (
	// SelectionNone disables selection state tracking. Enter triggers OnSelect only.
	SelectionNone SelectionMode = iota
	// SelectionSingle allows at most one selected node at a time.
	SelectionSingle
	// SelectionMulti allows any number of selected nodes (checkbox style).
	SelectionMulti
)

// NodeRenderFunc customises how a single node is rendered. It receives the
// node and whether it is currently highlighted by the cursor. Return the
// string that should appear in the tree line (connectors and selection
// indicators are still drawn by the Tree itself).
type NodeRenderFunc func(node *Node, isCursor bool) string

// TreeOpts configures Tree behaviour.
type TreeOpts struct {
	// OnSelect is called when the user presses Enter on a node. Optional.
	OnSelect func(node *Node)

	// OnToggle is called when the user expands or collapses a node. Optional.
	OnToggle func(node *Node)

	// OnContext is called when the user triggers the context action on a node
	// (default key: "c"). Optional.
	OnContext func(node *Node)

	// LoadChildren, if non-nil, is called the first time a node with nil
	// Children is expanded. It should return the children to attach.
	// This enables lazy loading of deep or dynamic trees.
	LoadChildren func(node *Node) []*Node

	// Selection sets the selection mode. Default is SelectionNone.
	Selection SelectionMode

	// RenderNode, if non-nil, overrides the default node label rendering.
	// The returned string replaces the glyph+title portion of each line.
	RenderNode NodeRenderFunc
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

	// filter is the active search/filter query. Empty means no filter.
	filter string
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

// SelectedNodes returns all nodes that have Selected == true, in tree order.
func (t *Tree) SelectedNodes() []*Node {
	var out []*Node
	t.collectSelected(t.roots, &out)
	return out
}

func (t *Tree) collectSelected(nodes []*Node, out *[]*Node) {
	for _, n := range nodes {
		if n.Selected {
			*out = append(*out, n)
		}
		t.collectSelected(n.Children, out)
	}
}

// SelectAll marks every node as Selected. Only meaningful in SelectionMulti mode.
func (t *Tree) SelectAll() {
	t.walkAll(t.roots, func(n *Node) { n.Selected = true })
}

// DeselectAll clears the Selected flag on every node.
func (t *Tree) DeselectAll() {
	t.walkAll(t.roots, func(n *Node) { n.Selected = false })
}

func (t *Tree) walkAll(nodes []*Node, fn func(*Node)) {
	for _, n := range nodes {
		fn(n)
		t.walkAll(n.Children, fn)
	}
}

// Filter returns the current filter query, or "" if no filter is active.
func (t *Tree) Filter() string { return t.filter }

// SetFilter applies a search filter. Only nodes whose Title contains the
// query (case-insensitive) are shown, along with their ancestors. An empty
// query clears the filter.
func (t *Tree) SetFilter(query string) {
	t.filter = query
	t.rebuild()
	t.cursor = 0
	t.scroll = 0
}

// ClearFilter removes the active filter and rebuilds the flat view.
func (t *Tree) ClearFilter() {
	t.SetFilter("")
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
		return t, t.handleKey(msg)
	case tea.MouseMsg:
		return t, t.handleMouse(msg)
	}
	return t, nil
}

func (t *Tree) handleMouse(msg tea.MouseMsg) tea.Cmd {
	// Ignore events outside component bounds.
	if msg.X < 0 || msg.X >= t.width || msg.Y < 0 || msg.Y >= t.height {
		return nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if t.cursor > 0 {
			t.cursor--
			t.clampScroll()
		}
		return Consumed()
	case tea.MouseButtonWheelDown:
		if t.cursor < len(t.flat)-1 {
			t.cursor++
			t.clampScroll()
		}
		return Consumed()
	}
	return nil
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
		if node != nil && !node.Expanded {
			// Lazy load children on first expand if callback is set.
			if node.Children == nil && t.opts.LoadChildren != nil {
				node.Children = t.opts.LoadChildren(node)
			}
			if len(node.Children) > 0 {
				node.Expanded = true
				t.rebuild()
				if t.opts.OnToggle != nil {
					t.opts.OnToggle(node)
				}
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
		if node != nil {
			t.applySelection(node)
			if t.opts.OnSelect != nil {
				t.opts.OnSelect(node)
			}
		}
		return Consumed()

	case " ":
		node := t.CursorNode()
		if node == nil {
			return Consumed()
		}
		// In multi-select, space toggles selection instead of expand/collapse.
		if t.opts.Selection == SelectionMulti {
			t.applySelection(node)
			return Consumed()
		}
		// Lazy load on first toggle.
		if node.Children == nil && t.opts.LoadChildren != nil {
			node.Children = t.opts.LoadChildren(node)
		}
		if len(node.Children) > 0 {
			node.Expanded = !node.Expanded
			t.rebuild()
			if t.opts.OnToggle != nil {
				t.opts.OnToggle(node)
			}
		}
		return Consumed()

	case "c":
		node := t.CursorNode()
		if node != nil && t.opts.OnContext != nil {
			t.opts.OnContext(node)
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
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.theme.Positive))

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
		switch {
		case fn.depth == 0:
			connector = ""
		case fn.isLast:
			connector = fn.prefix + g.TreeLast + " "
		default:
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

		// Selection indicator (checkbox-style) for Single/Multi modes.
		var selMark string
		if t.opts.Selection != SelectionNone {
			if node.Selected {
				selMark = g.SelectedBullet + " "
			} else {
				selMark = g.UnselectedBullet + " "
			}
		}

		// Node label: custom renderer or default glyph+title+detail.
		var label string
		if t.opts.RenderNode != nil {
			label = t.opts.RenderNode(node, isCursor)
		} else {
			glyph := node.Glyph
			if glyph != "" {
				glyph += " "
			}
			label = glyph + node.Title
			if node.Detail != "" {
				label += " " + node.Detail
			}
		}

		var line string
		if isCursor {
			prefixStyled := mutedStyle.Render(connector)
			line = prefixStyled + accentStyle.Render(arrow) + cursorStyle.Render(" "+selMark+label)
		} else {
			line = mutedStyle.Render(connector+arrow+" ") + selectedStyle.Render(selMark) + normalStyle.Render(label)
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
	binds := []KeyBind{
		{Key: "up/k", Label: "Move up", Group: "TREE"},
		{Key: "down/j", Label: "Move down", Group: "TREE"},
		{Key: "right/l", Label: "Expand", Group: "TREE"},
		{Key: "left/h", Label: "Collapse", Group: "TREE"},
		{Key: "enter", Label: "Select", Group: "TREE"},
		{Key: "space", Label: "Toggle expand", Group: "TREE"},
	}
	if t.opts.OnContext != nil {
		binds = append(binds, KeyBind{Key: "c", Label: "Context menu", Group: "TREE"})
	}
	return binds
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

// applySelection updates node selection state based on the current mode.
func (t *Tree) applySelection(node *Node) {
	switch t.opts.Selection {
	case SelectionSingle:
		wasSelected := node.Selected
		t.DeselectAll()
		node.Selected = !wasSelected
	case SelectionMulti:
		node.Selected = !node.Selected
	}
}

// rebuild linearises all currently visible nodes into t.flat.
func (t *Tree) rebuild() {
	t.flat = t.flat[:0]
	if t.filter == "" {
		for i, root := range t.roots {
			isLast := i == len(t.roots)-1
			t.buildFlat(root, 0, isLast, "")
		}
	} else {
		lowerFilter := strings.ToLower(t.filter)
		// Collect matching roots (nodes that match or have matching descendants).
		var matching []*Node
		for _, root := range t.roots {
			if t.nodeMatches(root, lowerFilter) {
				matching = append(matching, root)
			}
		}
		for i, root := range matching {
			isLast := i == len(matching)-1
			t.buildFlatFiltered(root, 0, isLast, "", lowerFilter)
		}
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

// nodeMatches returns true if the node or any descendant matches the filter.
func (t *Tree) nodeMatches(node *Node, lowerFilter string) bool {
	if strings.Contains(strings.ToLower(node.Title), lowerFilter) {
		return true
	}
	for _, child := range node.Children {
		if t.nodeMatches(child, lowerFilter) {
			return true
		}
	}
	return false
}

// buildFlatFiltered adds only matching nodes and their matching ancestors.
func (t *Tree) buildFlatFiltered(node *Node, depth int, isLast bool, prefix string, lowerFilter string) {
	t.flat = append(t.flat, flatNode{
		node:   node,
		depth:  depth,
		isLast: isLast,
		prefix: prefix,
	})
	// Always show children of matching ancestors when filtered.
	g := t.theme.glyphsOrDefault()
	childPrefix := prefix
	if depth > 0 {
		if isLast {
			childPrefix = prefix + g.TreeEmpty + " "
		} else {
			childPrefix = prefix + g.TreePipe + " "
		}
	}
	var matching []*Node
	for _, child := range node.Children {
		if t.nodeMatches(child, lowerFilter) {
			matching = append(matching, child)
		}
	}
	for i, child := range matching {
		childIsLast := i == len(matching)-1
		t.buildFlatFiltered(child, depth+1, childIsLast, childPrefix, lowerFilter)
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
