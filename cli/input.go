package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inputPromptStyle = lipgloss.NewStyle().Bold(true)
	inputValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	inputErrStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
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
	ti.Prompt = ""
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
	if m.done && !m.quitting {
		return inputPromptStyle.Render(m.prompt) + " " + inputValueStyle.Render(m.value) + "\n"
	}
	if m.done {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(inputPromptStyle.Render(m.prompt))
	sb.WriteString(" ")
	sb.WriteString(m.ti.View())
	if m.errMsg != "" {
		sb.WriteString("\n  ")
		sb.WriteString(inputErrStyle.Render(m.errMsg))
	}
	return sb.String()
}

// Input prompts for text input with optional validation.
// If validate is non-nil, the input is re-prompted on validation failure.
// Returns an error only if the user cancels with Ctrl+C.
func Input(prompt string, validate func(string) error) (string, error) {
	m := newInputModel(prompt, validate)
	p := tea.NewProgram(m)
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
