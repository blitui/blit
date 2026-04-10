package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogButton represents a button in a Dialog.
type DialogButton struct {
	// Label is the text shown on the button.
	Label string
	// Action is called when the button is activated. Optional.
	Action func()
}

// DialogOpts configures a Dialog.
type DialogOpts struct {
	// Title is optional text rendered in the top border.
	Title string

	// Body is the main content of the dialog.
	Body string

	// Buttons are the action buttons shown at the bottom.
	// If empty, a single "OK" button that closes the dialog is added.
	Buttons []DialogButton

	// OnClose is called when the dialog is dismissed (Esc or button action).
	OnClose func()
}

// Dialog is a modal overlay that displays a message with action buttons.
// It implements the Overlay interface.
type Dialog struct {
	opts    DialogOpts
	active  bool
	cursor  int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewDialog creates and opens a Dialog with the given options.
func NewDialog(opts DialogOpts) *Dialog {
	if len(opts.Buttons) == 0 {
		opts.Buttons = []DialogButton{{Label: "OK"}}
	}
	return &Dialog{
		opts:   opts,
		active: true,
	}
}

// Init implements Component.
func (d *Dialog) Init() tea.Cmd { return nil }

// Update implements Component.
func (d *Dialog) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !d.active || !d.focused {
		return d, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}
	switch km.String() {
	case "left", "h", "shift+tab":
		if d.cursor > 0 {
			d.cursor--
		}
		return d, Consumed()
	case "right", "l", "tab":
		if d.cursor < len(d.opts.Buttons)-1 {
			d.cursor++
		}
		return d, Consumed()
	case "enter", " ":
		btn := d.opts.Buttons[d.cursor]
		if btn.Action != nil {
			btn.Action()
		}
		d.Close()
		return d, Consumed()
	case "esc":
		d.Close()
		return d, Consumed()
	}
	return d, nil
}

// View implements Component.
func (d *Dialog) View() string {
	if !d.active {
		return ""
	}

	borders := DefaultBorders()
	if d.theme.Borders != nil {
		borders = *d.theme.Borders
	}
	border := borders.Rounded

	textColor := lipgloss.Color(d.theme.Text)
	accentColor := lipgloss.Color(d.theme.Accent)
	mutedColor := lipgloss.Color(d.theme.Muted)
	cursorBg := lipgloss.Color(d.theme.Cursor)
	inverseFg := lipgloss.Color(d.theme.TextInverse)

	// Determine dialog content width.
	contentWidth := d.width / 2
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentWidth > 60 {
		contentWidth = 60
	}
	innerWidth := contentWidth - 2 // account for border

	// Title.
	var title string
	if d.opts.Title != "" {
		title = d.opts.Title
	}

	// Body wrapped to inner width.
	body := lipgloss.NewStyle().
		Width(innerWidth).
		Foreground(textColor).
		Render(d.opts.Body)

	// Buttons row.
	var btnParts []string
	for i, btn := range d.opts.Buttons {
		label := " " + btn.Label + " "
		if i == d.cursor {
			style := lipgloss.NewStyle().
				Background(cursorBg).
				Foreground(inverseFg).
				Bold(true)
			btnParts = append(btnParts, style.Render(label))
		} else {
			style := lipgloss.NewStyle().
				Foreground(accentColor)
			btnParts = append(btnParts, style.Render(label))
		}
	}
	buttons := strings.Join(btnParts, "  ")

	// Center buttons.
	buttonsLine := lipgloss.NewStyle().
		Width(innerWidth).
		Align(lipgloss.Center).
		Render(buttons)

	content := body + "\n\n" + buttonsLine

	boxStyle := lipgloss.NewStyle().
		Border(border).
		BorderForeground(mutedColor).
		Padding(1, 2).
		Width(contentWidth)

	if title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
		boxStyle = boxStyle.BorderTop(true)
		content = titleStyle.Render(title) + "\n\n" + content
	}

	box := boxStyle.Render(content)

	// Center the dialog in the available space.
	return lipgloss.Place(d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		box)
}

// KeyBindings implements Component.
func (d *Dialog) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "left/right", Label: "Navigate buttons", Group: "DIALOG"},
		{Key: "enter/space", Label: "Activate button", Group: "DIALOG"},
		{Key: "esc", Label: "Close", Group: "DIALOG"},
	}
}

// SetSize implements Component.
func (d *Dialog) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// Focused implements Component.
func (d *Dialog) Focused() bool { return d.focused }

// SetFocused implements Component.
func (d *Dialog) SetFocused(f bool) { d.focused = f }

// SetTheme implements Themed.
func (d *Dialog) SetTheme(theme Theme) { d.theme = theme }

// IsActive implements Overlay.
func (d *Dialog) IsActive() bool { return d.active }

// Close implements Overlay.
func (d *Dialog) Close() {
	d.active = false
	if d.opts.OnClose != nil {
		d.opts.OnClose()
	}
}

// CursorIndex returns the currently highlighted button index.
func (d *Dialog) CursorIndex() int { return d.cursor }
