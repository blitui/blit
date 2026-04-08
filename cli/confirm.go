// Package cli provides interactive CLI primitives for tools that need
// more than plain stdout but less than a full Bubble Tea TUI.
package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	confirmPromptStyle = lipgloss.NewStyle().Bold(true)
	confirmHintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	confirmAnswerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
)

type confirmModel struct {
	prompt     string
	defaultYes bool
	result     bool
	done       bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n":
			m.result = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.result = m.defaultYes
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			m.result = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		answer := "No"
		if m.result {
			answer = "Yes"
		}
		return confirmPromptStyle.Render(m.prompt) + " " + confirmAnswerStyle.Render(answer) + "\n"
	}
	hint := "[y/N]"
	if m.defaultYes {
		hint = "[Y/n]"
	}
	return confirmPromptStyle.Render(m.prompt) + " " + confirmHintStyle.Render(hint) + " "
}

// Confirm prints a yes/no prompt and returns the user's choice.
// Default value is used when the user presses Enter without typing.
// Shows [Y/n] if defaultYes is true, [y/N] otherwise.
func Confirm(prompt string, defaultYes bool) bool {
	m := confirmModel{prompt: prompt, defaultYes: defaultYes}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return defaultYes
	}
	if final, ok := result.(confirmModel); ok {
		return final.result
	}
	return defaultYes
}
