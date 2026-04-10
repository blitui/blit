package blit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Help is an overlay that auto-generates a keybinding reference from the registry.
// It requires zero configuration — just call NewHelp() and register it with the App.
type Help struct {
	reg     *Registry
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

func (h *Help) setRegistry(r *Registry) { h.reg = r }

// SetTheme implements the Themed interface.
func (h *Help) SetTheme(t Theme) { h.theme = t }

// Init initializes the Help component.
func (h *Help) Init() tea.Cmd { return nil }

// Update handles incoming messages and updates Help state.
func (h *Help) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc", "?", "q":
			h.Close()
			return h, Consumed()
		}
	}
	return h, nil
}

// View renders the Help as a string.
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
		groupName := group.Name
		if strings.HasPrefix(groupName, "md:") {
			// Render the description portion (after the "md:" prefix) through
			// Markdown so that inline formatting (bold, code, links) is styled.
			rendered := strings.TrimSpace(Markdown(strings.TrimPrefix(groupName, "md:"), h.theme))
			sb.WriteString(rendered)
		} else {
			sb.WriteString(groupStyle.Render(groupName))
		}
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

// KeyBindings returns the key bindings for the Help.
func (h *Help) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "esc", Label: "Close help", Group: "OTHER"},
	}
}

// SetSize sets the width and height of the Help.
func (h *Help) SetSize(w, ht int) { h.width = w; h.height = ht }
// Focused reports whether the Help is focused.
func (h *Help) Focused() bool     { return h.focused }
// SetFocused sets the focus state of the Help.
func (h *Help) SetFocused(f bool) { h.focused = f }
// IsActive reports whether the Help overlay is currently visible.
func (h *Help) IsActive() bool    { return h.active }
// Close deactivates the Help and resets its state.
func (h *Help) Close()            { h.active = false }
