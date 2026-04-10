// Package main demonstrates the Split, Viewport, and Breadcrumbs components.
//
// The layout shows a simple file-tree placeholder on the left pane,
// a Viewport with scrollable content on the right pane,
// and Breadcrumbs at the top reflecting the "selected" tree node.
//
// Key bindings:
//
//	tab        — switch focus between left and right panes
//	j/k        — navigate list (left) or scroll (right)
//	alt+left   — shrink left pane
//	alt+right  — grow left pane
//	q          — quit
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

// ---- tree placeholder -------------------------------------------------------

// treeNode is a simple entry in the placeholder tree.
type treeNode struct {
	label    string
	children []string
	expanded bool
}

// TreePlaceholder is a minimal list-style tree placeholder that satisfies
// blit.Component. It will be replaced by the real Tree from Track A once
// that branch is merged.
type TreePlaceholder struct {
	nodes   []treeNode
	cursor  int
	focused bool
	width   int
	height  int
	theme   blit.Theme
}

func newTreePlaceholder() *TreePlaceholder {
	return &TreePlaceholder{
		nodes: []treeNode{
			{label: "src", children: []string{"main.go", "app.go", "theme.go"}, expanded: true},
			{label: "examples", children: []string{"dashboard", "splitpane", "picker"}},
			{label: "blit", children: []string{"harness.go", "assert.go"}},
			{label: "go.mod"},
			{label: "README.md"},
		},
	}
}

func (t *TreePlaceholder) Init() tea.Cmd                 { return nil }
func (t *TreePlaceholder) KeyBindings() []blit.KeyBind { return nil }
func (t *TreePlaceholder) SetSize(w, h int)              { t.width = w; t.height = h }
func (t *TreePlaceholder) Focused() bool                 { return t.focused }
func (t *TreePlaceholder) SetFocused(f bool)             { t.focused = f }
func (t *TreePlaceholder) SetTheme(th blit.Theme)      { t.theme = th }

func (t *TreePlaceholder) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
			return t, blit.Consumed()
		case "down", "j":
			if t.cursor < len(t.nodes)-1 {
				t.cursor++
			}
			return t, blit.Consumed()
		case "right", "l", "enter":
			t.nodes[t.cursor].expanded = !t.nodes[t.cursor].expanded
			return t, blit.Consumed()
		}
	}
	return t, nil
}

func (t *TreePlaceholder) View() string {
	cursor := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Accent)).Bold(true)
	normal := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Text))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Muted))
	child := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Text))

	var sb strings.Builder
	for i, n := range t.nodes {
		isCursor := t.focused && i == t.cursor
		arrow := "  "
		if len(n.children) > 0 {
			if n.expanded {
				arrow = "▾ "
			} else {
				arrow = "▸ "
			}
		}
		line := arrow + n.label
		if isCursor {
			sb.WriteString(cursor.Render("> " + line))
		} else {
			sb.WriteString(normal.Render("  " + line))
		}
		sb.WriteString("\n")
		if n.expanded && len(n.children) > 0 {
			for _, c := range n.children {
				sb.WriteString(muted.Render("  ├─ ") + child.Render(c) + "\n")
			}
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// SelectedPath returns the current breadcrumb path for the cursor node.
func (t *TreePlaceholder) SelectedPath() []string {
	if t.cursor < 0 || t.cursor >= len(t.nodes) {
		return nil
	}
	n := t.nodes[t.cursor]
	return []string{"blit", n.label}
}

// ---- model ------------------------------------------------------------------

type model struct {
	split       *blit.Split
	tree        *TreePlaceholder
	viewport    *blit.Viewport
	breadcrumbs *blit.Breadcrumbs
	width       int
	height      int
	theme       blit.Theme
}

func newModel() model {
	theme := blit.DefaultTheme()

	tree := newTreePlaceholder()
	tree.SetTheme(theme)

	vp := blit.NewViewport()
	vp.SetTheme(theme)
	vp.SetContent(buildViewportContent())

	sp := blit.NewSplit(blit.Horizontal, 0.35, tree, vp)
	sp.Resizable = true
	sp.SetTheme(theme)
	sp.SetFocused(true)

	bc := blit.NewBreadcrumbs(tree.SelectedPath())
	bc.SetTheme(theme)

	return model{
		split:       sp,
		tree:        tree,
		viewport:    vp,
		breadcrumbs: bc,
		theme:       theme,
	}
}

func buildViewportContent() string {
	lines := []string{
		"blit — v0.9.0 Viewport Demo",
		strings.Repeat("─", 50),
		"",
		"This pane is a Viewport component. It supports:",
		"  • j/k        scroll one line",
		"  • pgup/pgdn  scroll one page",
		"  • home/end   jump to top / bottom",
		"  • ctrl+u/d   half-page scroll",
		"  • mouse wheel",
		"",
		"The right-hand scrollbar shows a themed track (│)",
		"and thumb (█) that update as you scroll.",
		"",
		strings.Repeat("─", 50),
		"",
	}
	// Pad with more lines so scrolling is meaningful.
	for i := 1; i <= 40; i++ {
		lines = append(lines, fmt.Sprintf("  Line %3d — Lorem ipsum dolor sit amet, consectetur adipiscing elit.", i))
	}
	lines = append(lines, "", strings.Repeat("─", 50), "  [end of content]")
	return strings.Join(lines, "\n")
}

func (m model) Init() tea.Cmd { return m.split.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve 1 line for breadcrumbs, 1 for status bar hint.
		m.split.SetSize(msg.Width, msg.Height-2)
		m.breadcrumbs.SetSize(msg.Width, 1)
		m.breadcrumbs.MaxWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		updated, cmd := m.split.Update(msg, blit.Context{})
		m.split = updated.(*blit.Split)
		// Refresh breadcrumbs to reflect cursor movement.
		m.breadcrumbs.Segments = m.tree.SelectedPath()
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Muted)).
		Render("  tab focus  alt+←/→ resize  q quit")

	return m.breadcrumbs.View() + "\n" + m.split.View() + "\n" + hint
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
