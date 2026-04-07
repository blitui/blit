package tuikit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigField defines a single editable field in the config editor.
type ConfigField struct {
	Label string             // Display label
	Group string             // Group heading (e.g., "General", "Display")
	Hint  string             // Help text shown below the field
	Get   func() string      // Returns the current value
	Set   func(string) error // Sets a new value, returns error if invalid
}

// ConfigEditor is an overlay for editing application settings.
// Fields are declared with getter/setter closures for config-format-agnostic editing.
type ConfigEditor struct {
	fields  []ConfigField
	theme   Theme
	active  bool
	focused bool
	width   int
	height  int
	cursor  int
	editing bool
	editBuf string
	errMsg  string
	dirty   bool
}

// NewConfigEditor creates a config editor overlay with the given fields.
func NewConfigEditor(fields []ConfigField) *ConfigEditor {
	return &ConfigEditor{fields: fields}
}

// SetTheme implements the Themed interface.
func (c *ConfigEditor) SetTheme(t Theme) { c.theme = t }

func (c *ConfigEditor) Init() tea.Cmd { return nil }

func (c *ConfigEditor) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKey(msg)
	}
	return c, nil
}

func (c *ConfigEditor) handleKey(msg tea.KeyMsg) (Component, tea.Cmd) {
	if c.editing {
		return c.handleEditKey(msg)
	}

	switch msg.String() {
	case "up", "k":
		if c.cursor > 0 {
			c.cursor--
			c.errMsg = ""
		}
		return c, Consumed()
	case "down", "j":
		if c.cursor < len(c.fields)-1 {
			c.cursor++
			c.errMsg = ""
		}
		return c, Consumed()
	case "enter":
		f := c.fields[c.cursor]
		c.editing = true
		c.editBuf = f.Get()
		c.errMsg = ""
		return c, Consumed()
	case "esc", "q":
		c.Close()
		return c, Consumed()
	}
	return c, nil
}

func (c *ConfigEditor) handleEditKey(msg tea.KeyMsg) (Component, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		f := c.fields[c.cursor]
		if f.Set != nil {
			if err := f.Set(c.editBuf); err != nil {
				c.errMsg = err.Error()
				return c, Consumed()
			}
		}
		c.editing = false
		c.editBuf = ""
		c.errMsg = ""
		c.dirty = true
		return c, Consumed()
	case tea.KeyEscape:
		c.editing = false
		c.editBuf = ""
		c.errMsg = ""
		return c, Consumed()
	case tea.KeyBackspace:
		if len(c.editBuf) > 0 {
			c.editBuf = c.editBuf[:len(c.editBuf)-1]
		}
		return c, Consumed()
	case tea.KeyRunes:
		c.editBuf += string(msg.Runes)
		return c, Consumed()
	}
	return c, Consumed()
}

func (c *ConfigEditor) View() string {
	if !c.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Accent)).
		Bold(true)

	groupStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Muted)).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Text))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Accent))

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(c.theme.Cursor)).
		Foreground(lipgloss.Color(c.theme.TextInverse))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Muted)).
		Italic(true)

	errStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Negative))

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Settings"))
	sb.WriteString("\n\n")

	currentGroup := ""
	for i, f := range c.fields {
		if f.Group != currentGroup {
			if currentGroup != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(groupStyle.Render(f.Group))
			sb.WriteString("\n")
			currentGroup = f.Group
		}

		isCursor := i == c.cursor
		label := fmt.Sprintf("  %-20s", f.Label)
		val := f.Get()

		if isCursor && c.editing {
			val = c.editBuf + "█"
		}

		if isCursor {
			sb.WriteString(cursorStyle.Render(label) + " " + valueStyle.Render(val))
		} else {
			sb.WriteString(labelStyle.Render(label) + " " + valueStyle.Render(val))
		}
		sb.WriteString("\n")

		if isCursor && f.Hint != "" {
			sb.WriteString(hintStyle.Render("    " + f.Hint))
			sb.WriteString("\n")
		}
		if isCursor && c.errMsg != "" {
			sb.WriteString(errStyle.Render("    error: " + c.errMsg))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("  Enter to edit • Esc to close"))

	content := sb.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(c.theme.Border)).
		Padding(1, 2).
		Width(c.width - 4).
		Height(c.height - 2)

	return lipgloss.Place(c.width, c.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content))
}

func (c *ConfigEditor) KeyBindings() []KeyBind {
	if c.editing {
		return []KeyBind{
			{Key: "enter", Label: "Confirm", Group: "CONFIG"},
			{Key: "esc", Label: "Cancel edit", Group: "CONFIG"},
		}
	}
	return []KeyBind{
		{Key: "up/k", Label: "Previous field", Group: "CONFIG"},
		{Key: "down/j", Label: "Next field", Group: "CONFIG"},
		{Key: "enter", Label: "Edit field", Group: "CONFIG"},
		{Key: "esc", Label: "Close settings", Group: "CONFIG"},
	}
}

func (c *ConfigEditor) SetSize(w, h int) { c.width = w; c.height = h }
func (c *ConfigEditor) Focused() bool     { return c.focused }
func (c *ConfigEditor) SetFocused(f bool)  { c.focused = f }
func (c *ConfigEditor) IsActive() bool     { return c.active }
func (c *ConfigEditor) Close()             { c.active = false; c.editing = false; c.errMsg = "" }
