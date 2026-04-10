package tuikit

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moneycaringcoder/tuikit-go/internal/fuzzy"
)

// FilePickerOpts configures FilePicker behaviour.
type FilePickerOpts struct {
	// Root is the starting directory. Defaults to the current directory.
	Root string

	// PreviewPane renders a file preview pane on the right when true.
	PreviewPane bool

	// ShowHidden shows hidden files (dot-prefixed) when true.
	ShowHidden bool

	// OnSelect is called when the user presses Enter on a file node. Optional.
	OnSelect func(path string)

	// OnCancel is called when the user cancels. Optional.
	OnCancel func()
}

// filePickerLazyLoad is a message emitted after lazy-loading a directory.
type filePickerLazyLoad struct {
	node     *Node
	children []*Node
}

// FilePicker wraps a Tree over the local filesystem with lazy-loaded
// children and a Picker-style fuzzy search box.
type FilePicker struct {
	opts FilePickerOpts
	tree *Tree

	// search state
	input         textinput.Model
	searchActive  bool
	allNodes      []*Node  // flattened paths for search
	searchResults []string // matched absolute paths
	searchCursor  int

	// preview
	previewContent string
	previewNode    *Node

	theme   Theme
	focused bool
	width   int
	height  int
}

// NewFilePicker creates a FilePicker rooted at opts.Root (defaults to ".").
func NewFilePicker(opts FilePickerOpts) *FilePicker {
	if opts.Root == "" {
		opts.Root = "."
	}
	abs, err := filepath.Abs(opts.Root)
	if err != nil {
		abs = opts.Root
	}
	opts.Root = abs

	ti := textinput.New()
	ti.Placeholder = "Search files..."

	fp := &FilePicker{
		opts:  opts,
		input: ti,
	}

	root := fp.makeNode(opts.Root, true)
	root.Expanded = true
	fp.loadChildren(root)

	fp.tree = NewTree([]*Node{root}, TreeOpts{
		OnSelect: fp.onTreeSelect,
		OnToggle: fp.onTreeToggle,
	})

	fp.collectAll(fp.tree.Roots())
	return fp
}

// Init implements Component.
func (fp *FilePicker) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements Component.
func (fp *FilePicker) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case filePickerLazyLoad:
		msg.node.Children = msg.children
		fp.tree.rebuild()
		fp.collectAll(fp.tree.Roots())
		return fp, nil

	case tea.KeyMsg:
		if fp.searchActive {
			cmd := fp.handleSearchKey(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			cmd := fp.handleKey(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if fp.searchActive {
		var inputCmd tea.Cmd
		fp.input, inputCmd = fp.input.Update(msg)
		if inputCmd != nil {
			cmds = append(cmds, inputCmd)
		}
		fp.rebuildSearch()
	} else {
		var c Component
		var cmd tea.Cmd
		c, cmd = fp.tree.Update(msg, ctx)
		fp.tree = c.(*Tree)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return fp, tea.Batch(cmds...)
}

func (fp *FilePicker) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "/":
		fp.searchActive = true
		fp.input.Reset()
		fp.input.Focus()
		fp.rebuildSearch()
		return Consumed()
	}
	return nil
}

func (fp *FilePicker) handleSearchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		fp.searchActive = false
		fp.input.Blur()
		return Consumed()

	case "up", "ctrl+p":
		if fp.searchCursor > 0 {
			fp.searchCursor--
		}
		fp.invalidatePreview()
		return Consumed()

	case "down", "ctrl+n":
		if fp.searchCursor < len(fp.searchResults)-1 {
			fp.searchCursor++
		}
		fp.invalidatePreview()
		return Consumed()

	case "enter":
		if fp.searchCursor < len(fp.searchResults) {
			path := fp.searchResults[fp.searchCursor]
			if fp.opts.OnSelect != nil {
				fp.opts.OnSelect(path)
			}
		}
		return Consumed()
	}
	return nil
}

// View implements Component.
func (fp *FilePicker) View() string {
	if fp.width == 0 || fp.height == 0 {
		return ""
	}

	treeWidth := fp.width
	if fp.opts.PreviewPane {
		treeWidth = fp.width * 6 / 10
	}

	var left string
	if fp.searchActive {
		left = fp.renderSearch(treeWidth)
	} else {
		fp.tree.SetSize(treeWidth, fp.height)
		left = fp.tree.View()
	}

	if !fp.opts.PreviewPane {
		return left
	}

	previewWidth := fp.width - treeWidth - 1
	preview := fp.renderPreview(previewWidth)
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.Border)).
		Render(strings.Repeat("|\n", fp.height-1) + "|")

	return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, preview)
}

func (fp *FilePicker) renderSearch(width int) string {
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(fp.theme.Accent)).
		Width(width-2).
		Padding(0, 1)
	inputView := inputStyle.Render(fp.input.View())
	inputHeight := strings.Count(inputView, "\n") + 1

	availHeight := fp.height - inputHeight - 1
	if availHeight < 1 {
		availHeight = 1
	}

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.TextInverse)).
		Background(lipgloss.Color(fp.theme.Cursor)).
		Width(width)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.Text)).
		Width(width)
	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.Muted))

	start := 0
	if fp.searchCursor >= availHeight {
		start = fp.searchCursor - availHeight + 1
	}
	end := start + availHeight
	if end > len(fp.searchResults) {
		end = len(fp.searchResults)
	}

	var itemLines []string
	for i := start; i < end; i++ {
		path := fp.searchResults[i]
		label := filepath.Base(path)
		dir := filepath.Dir(path)
		if i == fp.searchCursor {
			itemLines = append(itemLines, cursorStyle.Render(" "+label+" "+mutedStyle.Render(dir)))
		} else {
			itemLines = append(itemLines, normalStyle.Render(" "+label+" "+mutedStyle.Render(dir)))
		}
	}

	if len(itemLines) == 0 {
		itemLines = []string{mutedStyle.Render("  No results")}
	}

	count := mutedStyle.Render(" " + strconv.Itoa(len(fp.searchResults)) + " results")
	return lipgloss.JoinVertical(lipgloss.Left, inputView, strings.Join(itemLines, "\n"), count)
}

func (fp *FilePicker) renderPreview(width int) string {
	if width < 2 {
		return ""
	}

	node := fp.tree.CursorNode()
	if fp.searchActive && fp.searchCursor < len(fp.searchResults) {
		// For search results, build a stub node for preview title.
		path := fp.searchResults[fp.searchCursor]
		node = &Node{Title: path, Data: path}
	}

	if node == nil {
		return ""
	}

	// Invalidate if node changed.
	if node != fp.previewNode {
		fp.previewNode = node
		fp.previewContent = ""
	}

	if fp.previewContent == "" {
		if path, ok := node.Data.(string); ok {
			fp.previewContent = loadFilePreview(path)
		}
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.Accent)).
		Bold(true).
		Width(width)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fp.theme.Text)).
		Width(width).
		Height(fp.height - 2)

	var title string
	if node != nil {
		title = node.Title
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		contentStyle.Render(fp.previewContent),
	)
}

// KeyBindings implements Component.
func (fp *FilePicker) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "/", Label: "Search", Group: "FILEPICKER"},
		{Key: "esc", Label: "Close search", Group: "FILEPICKER"},
		{Key: "up/k", Label: "Move up", Group: "FILEPICKER"},
		{Key: "down/j", Label: "Move down", Group: "FILEPICKER"},
		{Key: "right/l", Label: "Expand directory", Group: "FILEPICKER"},
		{Key: "left/h", Label: "Collapse directory", Group: "FILEPICKER"},
		{Key: "enter", Label: "Select", Group: "FILEPICKER"},
		{Key: "space", Label: "Toggle expand", Group: "FILEPICKER"},
	}
}

// SetSize implements Component.
func (fp *FilePicker) SetSize(w, h int) {
	fp.width = w
	fp.height = h
	fp.tree.SetSize(w, h)
	fp.input.Width = w - 6
}

// Focused implements Component.
func (fp *FilePicker) Focused() bool { return fp.focused }

// SetFocused implements Component.
func (fp *FilePicker) SetFocused(f bool) {
	fp.focused = f
	fp.tree.SetFocused(f)
	if !f {
		fp.searchActive = false
		fp.input.Blur()
	}
}

// SetTheme implements Themed.
func (fp *FilePicker) SetTheme(t Theme) {
	fp.theme = t
	fp.tree.SetTheme(t)
	fp.input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Text))
	fp.input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted))
	fp.input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
}

// onTreeSelect handles tree node selection.
func (fp *FilePicker) onTreeSelect(node *Node) {
	if path, ok := node.Data.(string); ok {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			if fp.opts.OnSelect != nil {
				fp.opts.OnSelect(path)
			}
		}
	}
}

// onTreeToggle handles tree expand/collapse, triggering lazy load.
func (fp *FilePicker) onTreeToggle(node *Node) {
	if node.Expanded && len(node.Children) == 0 {
		if path, ok := node.Data.(string); ok {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				fp.loadChildren(node)
				fp.collectAll(fp.tree.Roots())
			}
		}
	}
}

// makeNode creates a Node for a given path.
func (fp *FilePicker) makeNode(path string, isDir bool) *Node {
	g := fp.theme.glyphsOrDefault()
	glyph := g.Dot
	if isDir {
		glyph = g.CollapsedArrow
	}
	return &Node{
		Title: filepath.Base(path),
		Glyph: glyph,
		Data:  path,
	}
}

// loadChildren synchronously loads the direct children of a directory node.
func (fp *FilePicker) loadChildren(node *Node) {
	path, ok := node.Data.(string)
	if !ok {
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	var children []*Node
	for _, entry := range entries {
		name := entry.Name()
		if !fp.opts.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}
		childPath := filepath.Join(path, name)
		child := fp.makeNode(childPath, entry.IsDir())
		if entry.IsDir() {
			// Placeholder child so the node shows as expandable.
			child.Children = []*Node{}
		}
		children = append(children, child)
	}
	node.Children = children

	// Update the glyph to reflect it's a known directory.
	g := fp.theme.glyphsOrDefault()
	if node.Expanded {
		node.Glyph = g.ExpandedArrow
	} else {
		node.Glyph = g.CollapsedArrow
	}
}

// collectAll builds the flat list of all nodes for search.
func (fp *FilePicker) collectAll(nodes []*Node) {
	fp.allNodes = fp.allNodes[:0]
	fp.walkNodes(nodes)
}

func (fp *FilePicker) walkNodes(nodes []*Node) {
	for _, n := range nodes {
		fp.allNodes = append(fp.allNodes, n)
		if len(n.Children) > 0 {
			fp.walkNodes(n.Children)
		}
	}
}

// rebuildSearch filters allNodes by the current query.
func (fp *FilePicker) rebuildSearch() {
	query := fp.input.Value()
	fp.searchResults = fp.searchResults[:0]

	for _, n := range fp.allNodes {
		path, ok := n.Data.(string)
		if !ok {
			continue
		}
		if query == "" {
			fp.searchResults = append(fp.searchResults, path)
			continue
		}
		name := filepath.Base(path)
		m := fuzzy.Score(query, name)
		if m.Score > 0 {
			fp.searchResults = append(fp.searchResults, path)
		}
	}

	if fp.searchCursor >= len(fp.searchResults) {
		fp.searchCursor = max(0, len(fp.searchResults)-1)
	}
	fp.invalidatePreview()
}

func (fp *FilePicker) invalidatePreview() {
	fp.previewContent = ""
	fp.previewNode = nil
}

// loadFilePreview reads up to 4KB of a file for the preview pane.
func loadFilePreview(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "(cannot read)"
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "(cannot read directory)"
		}
		var sb strings.Builder
		for i, e := range entries {
			if i >= 20 {
				sb.WriteString("…\n")
				break
			}
			if e.IsDir() {
				sb.WriteString("  " + e.Name() + "/\n")
			} else {
				sb.WriteString("  " + e.Name() + "\n")
			}
		}
		return sb.String()
	}

	f, err := os.Open(path)
	if err != nil {
		return "(cannot open)"
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	return string(buf[:n])
}
