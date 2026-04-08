package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type selectModel struct {
	prompt    string
	items     []string
	filtered  []int // indices into items
	cursor    int
	filter    string
	useFilter bool
	selected  int
	quitting  bool
	done      bool
}

func newSelectModel(prompt string, items []string) selectModel {
	indices := make([]int, len(items))
	for i := range items {
		indices[i] = i
	}
	return selectModel{
		prompt:    prompt,
		items:     items,
		filtered:  indices,
		selected:  -1,
		useFilter: len(items) > 10,
	}
}

func (m *selectModel) applyFilter() {
	m.filtered = m.filtered[:0]
	lower := strings.ToLower(m.filter)
	for i, item := range m.items {
		if strings.Contains(strings.ToLower(item), lower) {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if len(m.filtered) > 0 {
				m.selected = m.filtered[m.cursor]
				m.done = true
				return m, tea.Quit
			}
			return m, nil
		case "backspace":
			if m.useFilter && len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}
			return m, nil
		default:
			if m.useFilter && len(msg.Runes) == 1 {
				m.filter += string(msg.Runes)
				m.applyFilter()
			}
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.done {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(m.prompt)
	sb.WriteString("\n")
	if m.useFilter {
		sb.WriteString(filterStyle.Render("  Filter: "))
		sb.WriteString(m.filter)
		sb.WriteString("\n")
	}
	for i, idx := range m.filtered {
		item := m.items[idx]
		if i == m.cursor {
			sb.WriteString(cursorStyle.Render("  > "))
			sb.WriteString(selectedStyle.Render(item))
		} else {
			sb.WriteString(dimStyle.Render("    "))
			sb.WriteString(dimStyle.Render(item))
		}
		sb.WriteString("\n")
	}
	if len(m.filtered) == 0 {
		sb.WriteString(dimStyle.Render("    (no matches)\n"))
	}
	return sb.String()
}

// SelectOne presents a list of options and returns the selected item and its index.
// Uses arrow keys or j/k for navigation, Enter to confirm.
// Supports type-to-filter when there are more than 10 items.
// Returns an error if the user cancels with Esc or Ctrl+C.
func SelectOne(prompt string, items []string) (string, int, error) {
	if len(items) == 0 {
		return "", -1, fmt.Errorf("select: no items provided")
	}
	m := newSelectModel(prompt, items)
	p := tea.NewProgram(m, tea.WithoutRenderer())
	result, err := p.Run()
	if err != nil {
		return "", -1, fmt.Errorf("select: %w", err)
	}
	final, ok := result.(selectModel)
	if !ok {
		return "", -1, fmt.Errorf("select: unexpected model type")
	}
	if final.quitting {
		return "", -1, fmt.Errorf("select: cancelled")
	}
	return items[final.selected], final.selected, nil
}
