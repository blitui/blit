package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

type model struct {
	fp       *tuikit.FilePicker
	selected string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return m.fp.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.fp.SetSize(msg.Width, msg.Height-2)
	}

	updated, cmd := m.fp.Update(msg)
	m.fp = updated.(*tuikit.FilePicker)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		if m.selected != "" {
			return fmt.Sprintf("Selected: %s\n", m.selected)
		}
		return "Cancelled.\n"
	}
	header := "File Tree  [/ search · enter select · q quit]\n"
	return header + m.fp.View()
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	var selected string
	fp := tuikit.NewFilePicker(tuikit.FilePickerOpts{
		Root:        root,
		PreviewPane: true,
		ShowHidden:  false,
		OnSelect: func(path string) {
			selected = path
		},
	})
	fp.SetTheme(tuikit.DefaultTheme())
	fp.SetFocused(true)

	m := model{fp: fp}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if selected != "" {
		fmt.Printf("Selected: %s\n", selected)
	}
}
