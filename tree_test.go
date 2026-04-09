package tuikit_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func makeTestTree() (*tuikit.Tree, []*tuikit.Node) {
	child1 := &tuikit.Node{Title: "child1"}
	child2 := &tuikit.Node{Title: "child2"}
	parent := &tuikit.Node{Title: "parent", Children: []*tuikit.Node{child1, child2}}
	leaf := &tuikit.Node{Title: "leaf"}

	roots := []*tuikit.Node{parent, leaf}
	t := tuikit.NewTree(roots, tuikit.TreeOpts{})
	t.SetTheme(tuikit.DefaultTheme())
	t.SetSize(80, 20)
	t.SetFocused(true)
	return t, roots
}

func sendKey(c tuikit.Component, key string) tuikit.Component {
	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated
}

func sendSpecialKey(c tuikit.Component, keyType tea.KeyType) tuikit.Component {
	updated, _ := c.Update(tea.KeyMsg{Type: keyType})
	return updated
}

func TestTree_Navigate(t *testing.T) {
	tree, roots := makeTestTree()

	// Initially cursor should be at index 0 (parent node).
	if tree.CursorNode() != roots[0] {
		t.Fatalf("expected cursor at parent, got %v", tree.CursorNode())
	}

	// Move down.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree = updated.(*tuikit.Tree)
	// parent is collapsed, so cursor moves to leaf (index 1).
	if tree.CursorNode() != roots[1] {
		t.Fatalf("expected cursor at leaf after down, got %v", tree.CursorNode())
	}

	// Move back up.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyUp})
	tree = updated.(*tuikit.Tree)
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
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight})
	tree = updated.(*tuikit.Tree)
	if !parent.Expanded {
		t.Fatal("parent should be expanded after right arrow")
	}

	// Move down — cursor should be at child1 now.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyDown})
	tree = updated.(*tuikit.Tree)
	if tree.CursorNode().Title != "child1" {
		t.Fatalf("expected child1, got %s", tree.CursorNode().Title)
	}

	// Move back to parent and collapse with left arrow.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyUp})
	tree = updated.(*tuikit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyLeft})
	tree = updated.(*tuikit.Tree)
	if parent.Expanded {
		t.Fatal("parent should be collapsed after left arrow")
	}
}

func TestTree_SpaceToggle(t *testing.T) {
	tree, roots := makeTestTree()
	parent := roots[0]

	// Space expands.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeySpace})
	tree = updated.(*tuikit.Tree)
	if !parent.Expanded {
		t.Fatal("space should expand parent")
	}

	// Space collapses.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeySpace})
	tree = updated.(*tuikit.Tree)
	if parent.Expanded {
		t.Fatal("space should collapse parent")
	}
}

func TestTree_Select(t *testing.T) {
	var selected *tuikit.Node
	tree, roots := makeTestTree()
	tree2 := tuikit.NewTree(roots, tuikit.TreeOpts{
		OnSelect: func(n *tuikit.Node) { selected = n },
	})
	tree2.SetTheme(tuikit.DefaultTheme())
	tree2.SetSize(80, 20)
	tree2.SetFocused(true)

	// Enter on parent (which has children, no file) should still call OnSelect.
	updated, _ := tree2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	tree2 = updated.(*tuikit.Tree)
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
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tree = updated.(*tuikit.Tree)
	if tree.CursorNode() != roots[1] {
		t.Fatalf("j should move cursor down to leaf, got %v", tree.CursorNode())
	}

	// k moves up.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	tree = updated.(*tuikit.Tree)
	if tree.CursorNode() != roots[0] {
		t.Fatalf("k should move cursor up to parent, got %v", tree.CursorNode())
	}

	// l expands.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	tree = updated.(*tuikit.Tree)
	if !roots[0].Expanded {
		t.Fatal("l should expand node")
	}

	// h collapses.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	tree = updated.(*tuikit.Tree)
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
	tree := tuikit.NewTree([]*tuikit.Node{}, tuikit.TreeOpts{})
	tree.SetTheme(tuikit.DefaultTheme())
	tree.SetSize(80, 20)
	view := tree.View()
	if view == "" {
		t.Fatal("empty tree should still render something")
	}
}

func TestTree_OnToggle(t *testing.T) {
	var toggled *tuikit.Node
	child := &tuikit.Node{Title: "child"}
	parent := &tuikit.Node{Title: "parent", Children: []*tuikit.Node{child}}
	tree := tuikit.NewTree([]*tuikit.Node{parent}, tuikit.TreeOpts{
		OnToggle: func(n *tuikit.Node) { toggled = n },
	})
	tree.SetTheme(tuikit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight})
	tree = updated.(*tuikit.Tree)
	if toggled != parent {
		t.Fatalf("expected OnToggle with parent, got %v", toggled)
	}
}
