package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem is a single entry in a Menu.
type MenuItem struct {
	// Label is the display text.
	Label string

	// Shortcut is an optional key hint shown right-aligned (e.g. "ctrl+s").
	Shortcut string

	// Action is called when the item is activated. Optional.
	Action func()

	// Disabled greys out the item and prevents activation.
	Disabled bool

	// Separator, when true, renders this item as a horizontal divider.
	// Label, Shortcut, Action, and Disabled are ignored.
	Separator bool
}

// MenuOpts configures a Menu.
type MenuOpts struct {
	// Items are the menu entries.
	Items []MenuItem

	// OnClose is called when the menu is dismissed.
	OnClose func()

	// MinWidth sets a minimum width for the menu. 0 means auto.
	MinWidth int
}

// Menu is a popup overlay that displays a list of selectable items.
// It implements the Overlay interface.
type Menu struct {
	opts    MenuOpts
	active  bool
	cursor  int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewMenu creates and opens a Menu with the given options.
func NewMenu(opts MenuOpts) *Menu {
	m := &Menu{
		opts:   opts,
		active: true,
	}
	m.skipSeparators(1)
	return m
}

// Init implements Component.
func (m *Menu) Init() tea.Cmd { return nil }

// Update implements Component.
func (m *Menu) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !m.active || !m.focused {
		return m, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "up", "k":
		m.moveCursor(-1)
		return m, Consumed()
	case "down", "j":
		m.moveCursor(1)
		return m, Consumed()
	case "enter", " ":
		if m.cursor >= 0 && m.cursor < len(m.opts.Items) {
			item := m.opts.Items[m.cursor]
			if !item.Disabled && !item.Separator {
				if item.Action != nil {
					item.Action()
				}
				m.Close()
			}
		}
		return m, Consumed()
	case "esc":
		m.Close()
		return m, Consumed()
	}
	return m, nil
}

// View implements Component.
func (m *Menu) View() string {
	if !m.active || len(m.opts.Items) == 0 {
		return ""
	}

	borders := DefaultBorders()
	if m.theme.Borders != nil {
		borders = *m.theme.Borders
	}

	textColor := lipgloss.Color(m.theme.Text)
	accentColor := lipgloss.Color(m.theme.Accent)
	mutedColor := lipgloss.Color(m.theme.Muted)
	cursorBg := lipgloss.Color(m.theme.Cursor)
	inverseFg := lipgloss.Color(m.theme.TextInverse)

	// Calculate content width.
	contentWidth := m.opts.MinWidth
	for _, item := range m.opts.Items {
		if item.Separator {
			continue
		}
		w := lipgloss.Width(item.Label)
		if item.Shortcut != "" {
			w += 2 + lipgloss.Width(item.Shortcut) // "  " gap + shortcut
		}
		if w > contentWidth {
			contentWidth = w
		}
	}
	contentWidth += 2 // padding

	var lines []string
	for i, item := range m.opts.Items {
		if item.Separator {
			sep := strings.Repeat("─", contentWidth)
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render(sep))
			continue
		}

		label := item.Label
		if item.Shortcut != "" {
			gap := contentWidth - lipgloss.Width(label) - lipgloss.Width(item.Shortcut)
			if gap < 2 {
				gap = 2
			}
			label = label + strings.Repeat(" ", gap) + item.Shortcut
		} else {
			// Pad to full width.
			gap := contentWidth - lipgloss.Width(label)
			if gap > 0 {
				label += strings.Repeat(" ", gap)
			}
		}

		isCursor := i == m.cursor
		switch {
		case item.Disabled:
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render(label))
		case isCursor:
			style := lipgloss.NewStyle().
				Background(cursorBg).
				Foreground(inverseFg)
			lines = append(lines, style.Render(label))
		default:
			style := lipgloss.NewStyle().Foreground(textColor)
			if item.Shortcut != "" {
				// Render label and shortcut with different colors.
				labelPart := lipgloss.NewStyle().Foreground(textColor).Render(item.Label)
				gap := contentWidth - lipgloss.Width(item.Label) - lipgloss.Width(item.Shortcut)
				if gap < 2 {
					gap = 2
				}
				shortPart := lipgloss.NewStyle().Foreground(accentColor).Render(item.Shortcut)
				lines = append(lines, labelPart+strings.Repeat(" ", gap)+shortPart)
				continue
			}
			lines = append(lines, style.Render(label))
		}
	}

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(borders.Rounded).
		BorderForeground(mutedColor).
		Padding(0, 1)

	box := boxStyle.Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}

// KeyBindings implements Component.
func (m *Menu) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/k", Label: "Move up", Group: "MENU"},
		{Key: "down/j", Label: "Move down", Group: "MENU"},
		{Key: "enter/space", Label: "Select", Group: "MENU"},
		{Key: "esc", Label: "Close", Group: "MENU"},
	}
}

// SetSize implements Component.
func (m *Menu) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Focused implements Component.
func (m *Menu) Focused() bool { return m.focused }

// SetFocused implements Component.
func (m *Menu) SetFocused(f bool) { m.focused = f }

// SetTheme implements Themed.
func (m *Menu) SetTheme(theme Theme) { m.theme = theme }

// IsActive implements Overlay.
func (m *Menu) IsActive() bool { return m.active }

// Close implements Overlay.
func (m *Menu) Close() {
	m.active = false
	if m.opts.OnClose != nil {
		m.opts.OnClose()
	}
}

// CursorIndex returns the currently highlighted item index.
func (m *Menu) CursorIndex() int { return m.cursor }

// moveCursor moves the cursor by delta, skipping separators and disabled items.
func (m *Menu) moveCursor(delta int) {
	n := len(m.opts.Items)
	if n == 0 {
		return
	}
	next := m.cursor
	for range n {
		next += delta
		if next < 0 || next >= n {
			return // Don't wrap.
		}
		item := m.opts.Items[next]
		if !item.Separator {
			m.cursor = next
			return
		}
	}
}

// skipSeparators moves cursor to the first non-separator item.
func (m *Menu) skipSeparators(dir int) {
	for i := range m.opts.Items {
		idx := i
		if dir < 0 {
			idx = len(m.opts.Items) - 1 - i
		}
		if !m.opts.Items[idx].Separator {
			m.cursor = idx
			return
		}
	}
}
