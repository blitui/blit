package tuikit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Help is an overlay that auto-generates a keybinding reference from the registry.
// It requires zero configuration — just call NewHelp() and register it with the App.
type Help struct {
	reg     *registry
	theme   Theme
	active  bool
	focused bool
	width   int
	height  int
}

// NewHelp creates a new Help overlay.
func NewHelp() *Help {
	return &Help{}
}

func (h *Help) setRegistry(r *registry) { h.reg = r }
// SetTheme implements the Themed interface.
func (h *Help) SetTheme(t Theme)        { h.theme = t }

func (h *Help) Init() tea.Cmd { return nil }

func (h *Help) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "q":
			h.Close()
			return h, Consumed()
		}
	}
	return h, nil
}

func (h *Help) View() string {
	if !h.active || h.reg == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Accent)).
		Bold(true)

	groupStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Muted)).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Text)).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Muted))

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	sb.WriteString("\n\n")

	groups := h.reg.grouped()
	for i, group := range groups {
		sb.WriteString(groupStyle.Render(group.Name))
		sb.WriteString("\n")
		for _, kb := range group.Bindings {
			line := fmt.Sprintf("  %s  %s",
				keyStyle.Render(fmt.Sprintf("%-12s", kb.Key)),
				labelStyle.Render(kb.Label),
			)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		if i < len(groups)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(labelStyle.Render("Press ? or Esc to close"))

	content := sb.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(h.theme.Border)).
		Padding(1, 2).
		Width(h.width - 4).
		Height(h.height - 2)

	return lipgloss.Place(h.width, h.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content))
}

func (h *Help) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "esc", Label: "Close help", Group: "OTHER"},
	}
}

func (h *Help) SetSize(w, ht int) { h.width = w; h.height = ht }
func (h *Help) Focused() bool      { return h.focused }
func (h *Help) SetFocused(f bool)   { h.focused = f }
func (h *Help) IsActive() bool      { return h.active }
func (h *Help) Close()              { h.active = false }
