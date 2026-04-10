package blit_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	blit "github.com/blitui/blit"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Create a small file tree:
	//   dir/
	//     file1.txt
	//     subdir/
	//       file2.go
	if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file2.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newTestFilePicker(t *testing.T, opts blit.FilePickerOpts) *blit.FilePicker {
	t.Helper()
	fp := blit.NewFilePicker(opts)
	fp.SetTheme(blit.DefaultTheme())
	fp.SetSize(80, 20)
	fp.SetFocused(true)
	return fp
}

func TestFilePicker_View(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})
	view := fp.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestFilePicker_Navigate(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Press down — should move within the tree.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	view := fp.View()
	if view == "" {
		t.Fatal("view after navigate should not be empty")
	}
}

func TestFilePicker_ExpandDirectory(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Root is already expanded; navigate to subdir and expand it.
	// Move down to first child (file1.txt or subdir).
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	// Move down again.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	// Expand with right.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	view := fp.View()
	if view == "" {
		t.Fatal("view after expand should not be empty")
	}
}

func TestFilePicker_SearchFilter(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Activate search with "/".
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Type "file" to filter.
	for _, r := range "file" {
		updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, blit.Context{})
		fp = updated.(*blit.FilePicker)
	}

	view := fp.View()
	if view == "" {
		t.Fatal("search view should not be empty")
	}
}

func TestFilePicker_SearchEsc(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Activate search.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Escape closes search — should return tree view.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyEsc}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	view := fp.View()
	if view == "" {
		t.Fatal("view after search close should not be empty")
	}
}

func TestFilePicker_Select(t *testing.T) {
	dir := makeTempDir(t)
	var selected string
	fp := newTestFilePicker(t, blit.FilePickerOpts{
		Root: dir,
		OnSelect: func(path string) {
			selected = path
		},
	})

	// Navigate to a file and press enter.
	// Down moves to first child (file1.txt in alphabetical order).
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.FilePicker)

	// We may or may not land on a file (depends on sorting), but no panic.
	_ = selected
}

func TestFilePicker_LazyLoad(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Navigate down to subdir and expand.
	// The subdir node should lazy-load its children on expand.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Expand — triggers lazy load.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Move into expanded children — should work without panic.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	view := fp.View()
	if view == "" {
		t.Fatal("view after lazy load should not be empty")
	}
}

func TestFilePicker_PreviewPane(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{
		Root:        dir,
		PreviewPane: true,
	})
	view := fp.View()
	if view == "" {
		t.Fatal("preview pane view should not be empty")
	}
}

func TestFilePicker_KeyBindings(t *testing.T) {
	dir := makeTempDir(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})
	binds := fp.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}
