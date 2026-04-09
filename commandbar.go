package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command defines a single command for the CommandBar.
type Command struct {
	Name    string             // Primary name: "quit", "sort", "go"
	Aliases []string           // Alternatives: ["q"] for "quit"
	Args    bool               // Whether it takes an argument
	Hint    string             // Help text
	Run     func(string) tea.Cmd // Handler, receives args string
}

// CommandBar is a vim-style : command input overlay.
type CommandBar struct {
	commands []Command
	theme    Theme
	active   bool
	focused  bool
	width    int
	height   int
	input    string
	errMsg   string
}

// NewCommandBar creates a command bar with the given commands.
func NewCommandBar(commands []Command) *CommandBar {
	return &CommandBar{commands: commands}
}

func (c *CommandBar) Init() tea.Cmd { return nil }

func (c *CommandBar) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKey(msg)
	}
	return c, nil
}

func (c *CommandBar) handleKey(msg tea.KeyMsg) (Component, tea.Cmd) {
	c.errMsg = ""

	switch msg.Type {
	case tea.KeyEscape:
		c.Close()
		return c, Consumed()
	case tea.KeyEnter:
		cmd := c.execute()
		if cmd != nil {
			c.Close()
			return c, cmd
		}
		return c, Consumed()
	case tea.KeyBackspace:
		if len(c.input) == 0 {
			c.Close()
			return c, Consumed()
		}
		c.input = c.input[:len(c.input)-1]
		return c, Consumed()
	case tea.KeyTab:
		c.tabComplete()
		return c, Consumed()
	case tea.KeyRunes:
		c.input += string(msg.Runes)
		return c, Consumed()
	}
	return c, Consumed()
}

func (c *CommandBar) execute() tea.Cmd {
	raw := strings.TrimSpace(c.input)
	if raw == "" {
		c.Close()
		return nil
	}

	parts := strings.SplitN(raw, " ", 2)
	name := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	for _, cmd := range c.commands {
		if strings.ToLower(cmd.Name) == name {
			if cmd.Run != nil {
				return cmd.Run(args)
			}
			return nil
		}
		for _, alias := range cmd.Aliases {
			if strings.ToLower(alias) == name {
				if cmd.Run != nil {
					return cmd.Run(args)
				}
				return nil
			}
		}
	}

	c.errMsg = "unknown command: " + parts[0]
	return nil
}

func (c *CommandBar) tabComplete() {
	prefix := strings.ToLower(c.input)
	if prefix == "" {
		return
	}
	for _, cmd := range c.commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), prefix) {
			c.input = cmd.Name
			return
		}
	}
}

func (c *CommandBar) View() string {
	if !c.active {
		return ""
	}

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Accent)).
		Bold(true)
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Text))
	errStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Negative))

	var line string
	if c.errMsg != "" {
		line = errStyle.Render(c.errMsg)
	} else {
		line = promptStyle.Render(":") + inputStyle.Render(c.input+"█")
	}

	return line
}

func (c *CommandBar) KeyBindings() []KeyBind {
	var bindings []KeyBind
	for _, cmd := range c.commands {
		bindings = append(bindings, KeyBind{
			Key:   ":" + cmd.Name,
			Label: cmd.Hint,
			Group: "COMMANDS",
		})
	}
	return bindings
}

func (c *CommandBar) SetSize(w, h int)  { c.width = w; c.height = h }
func (c *CommandBar) Focused() bool     { return c.focused }
func (c *CommandBar) SetFocused(f bool) { c.focused = f }
func (c *CommandBar) IsActive() bool    { return c.active }
func (c *CommandBar) SetActive(v bool)  { c.active = v }
func (c *CommandBar) Close() {
	c.active = false
	c.input = ""
	c.errMsg = ""
}

// Inline returns true — the CommandBar renders as a line at the bottom,
// not a fullscreen overlay.
func (c *CommandBar) Inline() bool { return true }

// SetTheme implements the Themed interface.
func (c *CommandBar) SetTheme(t Theme) { c.theme = t }
