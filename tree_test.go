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

func TestTree_SingleSelection(t *testing.T) {
	child := &blit.Node{Title: "child"}
	parent := &blit.Node{Title: "parent", Children: []*blit.Node{child}}
	leaf := &blit.Node{Title: "leaf"}
	roots := []*blit.Node{parent, leaf}

	tree := blit.NewTree(roots, blit.TreeOpts{
		Selection: blit.SelectionSingle,
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Enter selects parent.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !parent.Selected {
		t.Fatal("parent should be selected after enter")
	}

	// Move down and select leaf — parent should be deselected.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tree = updated.(*blit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	tree = updated.(*blit.Tree)
	if parent.Selected {
		t.Fatal("parent should be deselected in single mode")
	}
	if !leaf.Selected {
		t.Fatal("leaf should be selected")
	}

	selected := tree.SelectedNodes()
	if len(selected) != 1 || selected[0] != leaf {
		t.Fatalf("SelectedNodes() = %v, want [leaf]", selected)
	}

	// Enter again deselects.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	tree = updated.(*blit.Tree)
	if leaf.Selected {
		t.Fatal("leaf should be deselected after second enter")
	}
	if len(tree.SelectedNodes()) != 0 {
		t.Fatal("expected no selected nodes")
	}
}

func TestTree_MultiSelection(t *testing.T) {
	a := &blit.Node{Title: "a"}
	b := &blit.Node{Title: "b"}
	c := &blit.Node{Title: "c"}
	roots := []*blit.Node{a, b, c}

	tree := blit.NewTree(roots, blit.TreeOpts{
		Selection: blit.SelectionMulti,
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Space toggles selection in multi mode (doesn't expand/collapse).
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !a.Selected {
		t.Fatal("a should be selected after space")
	}

	// Move to b and select with space.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tree = updated.(*blit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	tree = updated.(*blit.Tree)
	if !b.Selected {
		t.Fatal("b should be selected")
	}
	if !a.Selected {
		t.Fatal("a should still be selected in multi mode")
	}

	selected := tree.SelectedNodes()
	if len(selected) != 2 {
		t.Fatalf("SelectedNodes() = %d, want 2", len(selected))
	}

	// Deselect a with enter.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	tree = updated.(*blit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Tree)
	if a.Selected {
		t.Fatal("a should be deselected after enter toggle")
	}
}

func TestTree_SelectAllDeselectAll(t *testing.T) {
	child := &blit.Node{Title: "child"}
	parent := &blit.Node{Title: "parent", Children: []*blit.Node{child}}
	leaf := &blit.Node{Title: "leaf"}
	roots := []*blit.Node{parent, leaf}

	tree := blit.NewTree(roots, blit.TreeOpts{
		Selection: blit.SelectionMulti,
	})

	tree.SelectAll()
	if !parent.Selected || !child.Selected || !leaf.Selected {
		t.Fatal("SelectAll should select all nodes including children")
	}
	if len(tree.SelectedNodes()) != 3 {
		t.Fatalf("SelectedNodes() = %d, want 3", len(tree.SelectedNodes()))
	}

	tree.DeselectAll()
	if parent.Selected || child.Selected || leaf.Selected {
		t.Fatal("DeselectAll should clear all selections")
	}
	if len(tree.SelectedNodes()) != 0 {
		t.Fatal("expected no selected nodes after DeselectAll")
	}
}

func TestTree_CustomRenderNode(t *testing.T) {
	node := &blit.Node{Title: "README.md", Glyph: "📄", Detail: "1.2KB"}
	roots := []*blit.Node{node}

	tree := blit.NewTree(roots, blit.TreeOpts{
		RenderNode: func(n *blit.Node, isCursor bool) string {
			return "[" + n.Glyph + "] " + n.Title + " (" + n.Detail + ")"
		},
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	view := tree.View()
	if view == "" {
		t.Fatal("View() should not be empty with custom renderer")
	}
	// The custom renderer output should appear in the view.
	if !containsStr(view, "[📄] README.md (1.2KB)") {
		t.Fatalf("custom render output not found in view:\n%s", view)
	}
}

func TestTree_Detail(t *testing.T) {
	node := &blit.Node{Title: "file.go", Detail: "42 lines"}
	roots := []*blit.Node{node}

	tree := blit.NewTree(roots, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)

	view := tree.View()
	if !containsStr(view, "42 lines") {
		t.Fatalf("Detail not rendered in view:\n%s", view)
	}
}

func TestTree_SelectionIndicator(t *testing.T) {
	node := &blit.Node{Title: "item", Selected: true}
	roots := []*blit.Node{node}

	tree := blit.NewTree(roots, blit.TreeOpts{
		Selection: blit.SelectionMulti,
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)

	view := tree.View()
	g := blit.DefaultGlyphs()
	if !containsStr(view, g.SelectedBullet) {
		t.Fatalf("selected bullet not found in view:\n%s", view)
	}

	// Deselected node should show unselected bullet.
	node.Selected = false
	view = tree.View()
	if !containsStr(view, g.UnselectedBullet) {
		t.Fatalf("unselected bullet not found in view:\n%s", view)
	}
}

func TestTree_SelectionNone_NoIndicator(t *testing.T) {
	node := &blit.Node{Title: "item"}
	roots := []*blit.Node{node}

	tree := blit.NewTree(roots, blit.TreeOpts{
		Selection: blit.SelectionNone,
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)

	view := tree.View()
	g := blit.DefaultGlyphs()
	if containsStr(view, g.SelectedBullet) || containsStr(view, g.UnselectedBullet) {
		t.Fatalf("selection indicators should not appear in SelectionNone mode:\n%s", view)
	}
}

func TestTree_SetRoots(t *testing.T) {
	tree, _ := makeTestTree()

	newRoots := []*blit.Node{{Title: "new1"}, {Title: "new2"}}
	tree.SetRoots(newRoots)

	if tree.CursorNode() != newRoots[0] {
		t.Fatal("cursor should be at first new root")
	}
}

func TestTree_MouseClick(t *testing.T) {
	tree, _ := makeTestTree()

	// Mouse outside bounds should be ignored.
	_, cmd := tree.Update(tea.MouseMsg{X: -1, Y: 0, Button: tea.MouseButtonWheelUp}, blit.Context{})
	if cmd != nil {
		t.Fatal("mouse outside bounds should return nil cmd")
	}
}

func TestTree_UnfocusedIgnoresInput(t *testing.T) {
	tree, roots := makeTestTree()
	tree.SetFocused(false)

	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tree = updated.(*blit.Tree)
	if tree.CursorNode() != roots[0] {
		t.Fatal("unfocused tree should not process key input")
	}
}

func TestTree_ZeroSize(t *testing.T) {
	tree := blit.NewTree([]*blit.Node{{Title: "x"}}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(0, 0)
	if tree.View() != "" {
		t.Fatal("zero-sized tree should return empty view")
	}
}

func TestTree_LazyLoading(t *testing.T) {
	loadCalled := 0
	parent := &blit.Node{Title: "parent"} // nil Children — lazy loadable
	roots := []*blit.Node{parent}

	tree := blit.NewTree(roots, blit.TreeOpts{
		LoadChildren: func(node *blit.Node) []*blit.Node {
			loadCalled++
			return []*blit.Node{
				{Title: "lazy-child-1"},
				{Title: "lazy-child-2"},
			}
		},
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Right arrow should trigger lazy load.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	tree = updated.(*blit.Tree)
	if loadCalled != 1 {
		t.Fatalf("LoadChildren called %d times, want 1", loadCalled)
	}
	if !parent.Expanded {
		t.Fatal("parent should be expanded after right arrow")
	}
	if len(parent.Children) != 2 {
		t.Fatalf("parent.Children = %d, want 2", len(parent.Children))
	}

	// Collapse and re-expand should NOT call LoadChildren again.
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	tree = updated.(*blit.Tree)
	updated, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	_ = updated.(*blit.Tree)
	if loadCalled != 1 {
		t.Fatalf("LoadChildren called %d times on re-expand, want 1", loadCalled)
	}
}

func TestTree_LazyLoadingSpaceToggle(t *testing.T) {
	loadCalled := 0
	parent := &blit.Node{Title: "parent"} // nil Children
	roots := []*blit.Node{parent}

	tree := blit.NewTree(roots, blit.TreeOpts{
		LoadChildren: func(node *blit.Node) []*blit.Node {
			loadCalled++
			return []*blit.Node{{Title: "child"}}
		},
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Space should trigger lazy load too.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeySpace}, blit.Context{})
	_ = updated.(*blit.Tree)
	if loadCalled != 1 {
		t.Fatalf("LoadChildren via space called %d times, want 1", loadCalled)
	}
	if !parent.Expanded {
		t.Fatal("parent should be expanded after space")
	}
}

func TestTree_LazyLoadingNoCallback(t *testing.T) {
	// Node with nil Children and no LoadChildren — should not panic on expand.
	parent := &blit.Node{Title: "parent"}
	tree := blit.NewTree([]*blit.Node{parent}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	_ = updated.(*blit.Tree)
	if parent.Expanded {
		t.Fatal("node with nil children and no loader should not expand")
	}
}

func TestTree_Filter(t *testing.T) {
	child1 := &blit.Node{Title: "apple"}
	child2 := &blit.Node{Title: "banana"}
	parent := &blit.Node{Title: "fruits", Children: []*blit.Node{child1, child2}, Expanded: true}
	leaf := &blit.Node{Title: "vegetable"}
	roots := []*blit.Node{parent, leaf}

	tree := blit.NewTree(roots, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Filter for "apple" — should show parent (ancestor) + apple.
	tree.SetFilter("apple")
	if tree.Filter() != "apple" {
		t.Fatalf("Filter() = %q, want %q", tree.Filter(), "apple")
	}

	view := tree.View()
	if !containsStr(view, "apple") {
		t.Fatalf("filtered view should contain 'apple':\n%s", view)
	}
	if containsStr(view, "banana") {
		t.Fatalf("filtered view should NOT contain 'banana':\n%s", view)
	}
	if containsStr(view, "vegetable") {
		t.Fatalf("filtered view should NOT contain 'vegetable':\n%s", view)
	}

	// Clear filter — all nodes visible again.
	tree.ClearFilter()
	if tree.Filter() != "" {
		t.Fatal("Filter() should be empty after ClearFilter")
	}
	view = tree.View()
	if !containsStr(view, "vegetable") {
		t.Fatalf("unfiltered view should contain 'vegetable':\n%s", view)
	}
}

func TestTree_FilterCaseInsensitive(t *testing.T) {
	node := &blit.Node{Title: "MyDocument"}
	tree := blit.NewTree([]*blit.Node{node}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)

	tree.SetFilter("mydoc")
	view := tree.View()
	if !containsStr(view, "MyDocument") {
		t.Fatalf("case-insensitive filter should match:\n%s", view)
	}
}

func TestTree_FilterNoMatch(t *testing.T) {
	node := &blit.Node{Title: "hello"}
	tree := blit.NewTree([]*blit.Node{node}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)

	tree.SetFilter("zzz")
	view := tree.View()
	// Should show empty tree indicator.
	if containsStr(view, "hello") {
		t.Fatalf("filtered view should not contain 'hello':\n%s", view)
	}
}

func TestTree_OnContext(t *testing.T) {
	var contextNode *blit.Node
	child := &blit.Node{Title: "child"}
	parent := &blit.Node{Title: "parent", Children: []*blit.Node{child}}
	tree := blit.NewTree([]*blit.Node{parent}, blit.TreeOpts{
		OnContext: func(n *blit.Node) { contextNode = n },
	})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// Press 'c' to trigger context menu on cursor node.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}, blit.Context{})
	_ = updated.(*blit.Tree)
	if contextNode != parent {
		t.Fatalf("OnContext should be called with parent, got %v", contextNode)
	}
}

func TestTree_OnContextNoCallback(t *testing.T) {
	node := &blit.Node{Title: "item"}
	tree := blit.NewTree([]*blit.Node{node}, blit.TreeOpts{})
	tree.SetTheme(blit.DefaultTheme())
	tree.SetSize(80, 20)
	tree.SetFocused(true)

	// 'c' with no OnContext should not panic.
	updated, _ := tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}, blit.Context{})
	_ = updated.(*blit.Tree)
}

func TestTree_ContextKeyBinding(t *testing.T) {
	// With OnContext set, KeyBindings should include "c".
	tree := blit.NewTree([]*blit.Node{{Title: "x"}}, blit.TreeOpts{
		OnContext: func(n *blit.Node) {},
	})
	binds := tree.KeyBindings()
	found := false
	for _, b := range binds {
		if b.Key == "c" {
			found = true
		}
	}
	if !found {
		t.Fatal("KeyBindings should include 'c' when OnContext is set")
	}

	// Without OnContext, no "c" binding.
	tree2 := blit.NewTree([]*blit.Node{{Title: "x"}}, blit.TreeOpts{})
	for _, b := range tree2.KeyBindings() {
		if b.Key == "c" {
			t.Fatal("KeyBindings should NOT include 'c' when OnContext is nil")
		}
	}
}

// containsStr is a simple helper to avoid importing strings in _test.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
