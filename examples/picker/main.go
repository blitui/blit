// Package main demonstrates the blit Picker as a file browser.
//
// Navigate with up/down, type to fuzzy-filter file names, and press enter to
// select. The right pane shows a preview of the selected file content.
// Press ctrl+p to open the picker overlay.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	items, err := buildFileItems(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading directory: %v\n", err)
		os.Exit(1)
	}

	var selectedMsg string

	picker := blit.NewPicker(items, blit.PickerOpts{
		Placeholder: "Type to filter files...",
		Preview:     true,
		OnConfirm: func(item blit.PickerItem) {
			selectedMsg = "Selected: " + item.Title
		},
	})

	status := &helpStatus{msg: "Press ctrl+p to open the file picker"}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("main", status),
		blit.WithOverlay("File Picker", "ctrl+p", picker),
		blit.WithStatusBar(
			func() string {
				if selectedMsg != "" {
					return " " + selectedMsg
				}
				return " ctrl+p open picker  q quit"
			},
			func() string { return " File Browser " },
		),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func buildFileItems(dir string) ([]blit.PickerItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []blit.PickerItem
	for _, e := range entries {
		name := e.Name()
		fullPath := filepath.Join(dir, name)

		glyph := "  "
		subtitle := "file"
		if e.IsDir() {
			glyph = "  "
			subtitle = "directory"
		} else {
			switch strings.ToLower(filepath.Ext(name)) {
			case ".go":
				glyph = "  "
			case ".md":
				glyph = "  "
			case ".json", ".yaml", ".yml", ".toml":
				glyph = "  "
			}
		}

		path := fullPath
		items = append(items, blit.PickerItem{
			Title:    name,
			Subtitle: subtitle,
			Glyph:    glyph,
			Preview: func() string {
				return readPreview(path)
			},
			Value: fullPath,
		})
	}
	return items, nil
}

func readPreview(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "(error reading file)"
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "(error reading directory)"
		}
		var lines []string
		for _, e := range entries {
			prefix := "  "
			if e.IsDir() {
				prefix = "  "
			}
			lines = append(lines, prefix+e.Name())
		}
		return strings.Join(lines, "\n")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "(error reading file)"
	}
	if isBinary(data) {
		return fmt.Sprintf("(binary file, %d bytes)", len(data))
	}
	lines := strings.SplitN(string(data), "\n", 51)
	if len(lines) > 50 {
		lines = lines[:50]
		lines = append(lines, "... (truncated)")
	}
	return strings.Join(lines, "\n")
}

func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

type helpStatus struct {
	msg     string
	theme   blit.Theme
	focused bool
	width   int
	height  int
}

func (s *helpStatus) Init() tea.Cmd { return nil }
func (s *helpStatus) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	return s, nil
}
func (s *helpStatus) KeyBindings() []blit.KeyBind { return nil }
func (s *helpStatus) SetSize(w, h int)              { s.width = w; s.height = h }
func (s *helpStatus) Focused() bool                 { return s.focused }
func (s *helpStatus) SetFocused(f bool)             { s.focused = f }
func (s *helpStatus) SetTheme(t blit.Theme)       { s.theme = t }

func (s *helpStatus) View() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Muted)).
		Width(s.width).
		Height(s.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(s.msg)
}
