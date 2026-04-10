package blit_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	blit "github.com/blitui/blit"
)

func makeTestTree() (*blit.Tree, []*blit.Node) {
	child1 := &blit.Node{Title: "child1"}
	child2 := &blit.Node{Title: "child2"}
	parent := &blit.Node{Title: "parent", Children: []*blit.Node{child1, child2}}
	leaf := &blit.Node{Title: "leaf"}

	roots := []*blit.Node{parent, leaf}
	t := blit.NewTree(roots, blit.TreeOpts{})
	t.SetTheme(blit.DefaultTheme())
	t.SetSize(80, 20)
	t.SetFocused(true)
	return t, roots
}

func TestTree_Navigate(t *testing.T) {
	tree, roots := makeTestTree()

	// Initially cursor should be at index 0 (parent node).
	if tree.CursorNode() != roots[0] {
		t.Fatalf("expected cursor at parent, got %v", tree.CursorNode())
	}

	// Move down.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tree = updated.(*blit.Tree)
	// parent is collapsed, so cursor moves to leaf (index 1).
	if tree.CursorNode() != roots[1] {
		t.Fatalf("expected cursor at leaf after down, got %v", tree.CursorNode())
	}

	// Move back up.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	tree = updated.(*blit.Tree)
	if tree.CursorNode() != roots[0] {
		t.Fatalf("expected cursor back at parent after up, got %v", tree.CursorNode())
	}
}

func TestTree_ExpandCollapse(t *testing.T) {
	tree, roots := makeTestTree()
	parent := roots[0]

	if parent.Expanded {
		t.Fatal("parent should start collapsed")
	}

	// Expand with right arrow.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !parent.Expanded {
		t.Fatal("parent should be expanded after right arrow")
	}

	// Move down — cursor should be at child1 now.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tree = updated.(*blit.Tree)
	if tree.CursorNode().Title != "child1" {
		t.Fatalf("expected child1, got %s", tree.CursorNode().Title)
	}

	// Move back to parent and collapse with left arrow.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	tree = updated.(*blit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	_ = updated.(*blit.Tree)
	if parent.Expanded {
		t.Fatal("parent should be collapsed after left arrow")
	}
}

func TestTree_SpaceToggle(t *testing.T) {
	tree, roots := makeTestTree()
	parent := roots[0]

	// Space expands.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !parent.Expanded {
		t.Fatal("space should expand parent")
	}

	// Space collapses.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	_ = updated.(*blit.Tree)
	if parent.Expanded {
		t.Fatal("space should collapse parent")
	}
}

func TestTree_Select(t *testing.T) {
	var selected *blit.Node
	tree, roots := makeTestTree()
	tree2 := blit.NewTree(roots, blit.TreeOpts{
		OnSelect: func(n *blit.Node) { selected = n },
	})
	tree2.SetTheme(blit.DefaultTheme())
	tree2.SetSize(80, 20)
	tree2.SetFocused(true)

	// Enter on parent (which has children, no file) should still call OnSelect.
	updated, _ := tree2.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Tree)
	if selected != roots[0] {
		t.Fatalf("expected OnSelect called with parent, got %v", selected)
	}

	_ = tree
}

func TestTree_View(t *testing.T) {
	tree, _ := makeTestTree()
	view := tree.View()
	if view == "" {
		t.Fatal("View() should not be empty")
	}
	if len(view) == 0 {
		t.Fatal("expected non-empty view")
	}
}

func TestTree_ViAlias(t *testing.T) {
	tree, roots := makeTestTree()

	// j moves down.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, blit.Context{})
	tree = updated.(*blit.Tree)
	if tree.CursorNode() != roots[1] {
		t.Fatalf("j should move cursor down to leaf, got %v", tree.CursorNode())
	}

	// k moves up.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, blit.Context{})
	tree = updated.(*blit.Tree)
	if tree.CursorNode() != roots[0] {
		t.Fatalf("k should move cursor up to parent, got %v", tree.CursorNode())
	}

	// l expands.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !roots[0].Expanded {
		t.Fatal("l should expand node")
	}

	// h collapses.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, blit.Context{})
	_ = updated.(*blit.Tree)
	if roots[0].Expanded {
		t.Fatal("h should collapse node")
	}
}

func TestTree_KeyBindings(t *testing.T) {
	tree, _ := makeTestTree()
	binds := tree.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}

func TestTree_EmptyRoots(t *testing.T) {
	tree := blit.NewTree([]*blit.Node{}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	view := tree.View()
	if view == "" {
		t.Fatal("empty tree should still render something")
	}
}

func TestTree_OnToggle(t *testing.T) {
	var toggled *blit.Node
	child := &blit.Node{Title: "child"}
	parent := &blit.Node{Title: "parent", Children: []*blit.Node{child}}
	tree := blit.NewTree([]*blit.Node{parent}, blit.TreeOpts{
		OnToggle: func(n *blit.Node) { toggled = n },
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	_ = updated.(*blit.Tree)
	if toggled != parent {
		t.Fatalf("expected OnToggle with parent, got %v", toggled)
	}
}
