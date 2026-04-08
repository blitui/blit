package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputModel struct {
	prompt   string
	validate func(string) error
	ti       textinput.Model
	errMsg   string
	value    string
	done     bool
	quitting bool
}

func newInputModel(prompt string, validate func(string) error) inputModel {
	ti := textinput.New()
	ti.Focus()
	return inputModel{
		prompt:   prompt,
		validate: validate,
		ti:       ti,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			m.done = true
			return m, tea.Quit
		case "enter":
			val := strings.TrimSpace(m.ti.Value())
			if m.validate != nil {
				if err := m.validate(val); err != nil {
					m.errMsg = err.Error()
					m.ti.SetValue("")
					return m, nil
				}
			}
			m.value = val
			m.done = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.done {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(m.prompt)
	sb.WriteString(" ")
	sb.WriteString(m.ti.View())
	if m.errMsg != "" {
		sb.WriteString("\n  ")
		sb.WriteString(m.errMsg)
	}
	return sb.String()
}

// Input prompts for text input with optional validation.
// If validate is non-nil, the input is re-prompted on validation failure.
// Returns an error only if the user cancels with Ctrl+C.
func Input(prompt string, validate func(string) error) (string, error) {
	m := newInputModel(prompt, validate)
	p := tea.NewProgram(m, tea.WithoutRenderer())
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("input: %w", err)
	}
	final, ok := result.(inputModel)
	if !ok {
		return "", fmt.Errorf("input: unexpected model type")
	}
	if final.quitting {
		return "", fmt.Errorf("input: cancelled")
	}
	return final.value, nil
}
