package blit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func makeTempDirWithHidden(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFilePicker_HiddenFiles(t *testing.T) {
	dir := makeTempDirWithHidden(t)

	// Without ShowHidden, .hidden should not appear.
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})
	view := fp.View()
	if strings.Contains(view, ".hidden") {
		t.Fatal("hidden file should not appear when ShowHidden is false")
	}

	// With ShowHidden, .hidden should appear.
	fp2 := newTestFilePicker(t, blit.FilePickerOpts{Root: dir, ShowHidden: true})
	view2 := fp2.View()
	if !strings.Contains(view2, ".hidden") {
		t.Fatalf("hidden file should appear when ShowHidden is true:\n%s", view2)
	}
}

func TestFilePicker_SearchNavigate(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Activate search.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Navigate down in search results.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Navigate up.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	view := fp.View()
	if view == "" {
		t.Fatal("search view after navigation should not be empty")
	}
}

func TestFilePicker_SearchCtrlPN(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Activate search.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// ctrl+n moves down.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyCtrlN}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// ctrl+p moves up.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyCtrlP}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	view := fp.View()
	if view == "" {
		t.Fatal("view should not be empty")
	}
}

func TestFilePicker_SearchEnter(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	var selected string
	fp := newTestFilePicker(t, blit.FilePickerOpts{
		Root:     dir,
		OnSelect: func(path string) { selected = path },
	})

	// Activate search.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Press enter on first result.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.FilePicker)

	// OnSelect may or may not fire depending on what's at cursor, but no panic.
	_ = selected
}

func TestFilePicker_PreviewPaneContent(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{
		Root:        dir,
		PreviewPane: true,
	})

	// Navigate to a file.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	view := fp.View()
	if view == "" {
		t.Fatal("preview pane view should not be empty")
	}
}

func TestFilePicker_ZeroSize(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := blit.NewFilePicker(blit.FilePickerOpts{Root: dir})
	fp.SetTheme(blit.DefaultTheme())
	fp.SetSize(0, 0)
	if fp.View() != "" {
		t.Fatal("zero-sized filepicker should return empty view")
	}
}

func TestFilePicker_Unfocused(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})
	fp.SetFocused(false)

	if fp.Focused() {
		t.Fatal("should not be focused")
	}
}

func TestFilePicker_Init(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})
	cmd := fp.Init()
	if cmd == nil {
		t.Fatal("Init() should return a blink command")
	}
}

func TestFilePicker_DefaultRoot(t *testing.T) {
	// Not specifying a root should default to current directory.
	fp := blit.NewFilePicker(blit.FilePickerOpts{})
	fp.SetTheme(blit.DefaultTheme())
	fp.SetSize(80, 20)
	fp.SetFocused(true)
	view := fp.View()
	if view == "" {
		t.Fatal("default root filepicker should have a non-empty view")
	}
}

func TestFilePicker_CollapseDirectory(t *testing.T) {
	dir := makeTempDirWithHidden(t)
	fp := newTestFilePicker(t, blit.FilePickerOpts{Root: dir})

	// Navigate to subdir.
	updated, _ := fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Expand.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	// Collapse with left.
	updated, _ = fp.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	fp = updated.(*blit.FilePicker)

	view := fp.View()
	if view == "" {
		t.Fatal("view after collapse should not be empty")
	}
}
