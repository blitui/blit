package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BreadcrumbItem is a single segment in a Breadcrumb path.
type BreadcrumbItem struct {
	// Label is the display text.
	Label string
	// Data is an arbitrary payload the caller can attach.
	Data any
}

// BreadcrumbOpts configures a Breadcrumb.
type BreadcrumbOpts struct {
	// Separator is the string between items. Defaults to " > ".
	Separator string

	// OnSelect is called when enter is pressed on an item.
	OnSelect func(item BreadcrumbItem, index int)
}

// Breadcrumb is an interactive, cursor-driven breadcrumb navigation component
// with selection support. It displays a navigable path of items and allows the
// user to move between them with keyboard input and invoke an OnSelect callback.
// It implements Component and Themed.
//
// For a simple display-only path trail, see Breadcrumbs.
type Breadcrumb struct {
	opts    BreadcrumbOpts
	items   []BreadcrumbItem
	cursor  int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewBreadcrumb creates a Breadcrumb with the given items and options.
func NewBreadcrumb(items []BreadcrumbItem, opts BreadcrumbOpts) *Breadcrumb {
	if opts.Separator == "" {
		opts.Separator = " > "
	}
	return &Breadcrumb{
		opts:   opts,
		items:  items,
		cursor: max(0, len(items)-1), // default to last item
	}
}

// Items returns the current breadcrumb items.
func (b *Breadcrumb) Items() []BreadcrumbItem { return b.items }

// SetItems replaces the items and resets the cursor to the last item.
func (b *Breadcrumb) SetItems(items []BreadcrumbItem) {
	b.items = items
	b.cursor = max(0, len(items)-1)
}

// Push appends an item and moves the cursor to it.
func (b *Breadcrumb) Push(item BreadcrumbItem) {
	b.items = append(b.items, item)
	b.cursor = len(b.items) - 1
}

// Pop removes and returns the last item, moving the cursor back. Returns
// the zero value if there are no items.
func (b *Breadcrumb) Pop() BreadcrumbItem {
	if len(b.items) == 0 {
		return BreadcrumbItem{}
	}
	last := b.items[len(b.items)-1]
	b.items = b.items[:len(b.items)-1]
	if b.cursor >= len(b.items) {
		b.cursor = max(0, len(b.items)-1)
	}
	return last
}

// CursorIndex returns the currently highlighted item index.
func (b *Breadcrumb) CursorIndex() int { return b.cursor }

// Init implements Component.
func (b *Breadcrumb) Init() tea.Cmd { return nil }

// Update implements Component.
func (b *Breadcrumb) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !b.focused {
		return b, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return b, nil
	}
	switch km.String() {
	case "left", "h":
		if b.cursor > 0 {
			b.cursor--
		}
		return b, Consumed()
	case "right", "l":
		if b.cursor < len(b.items)-1 {
			b.cursor++
		}
		return b, Consumed()
	case "enter":
		if b.cursor < len(b.items) && b.opts.OnSelect != nil {
			b.opts.OnSelect(b.items[b.cursor], b.cursor)
		}
		return b, Consumed()
	}
	return b, nil
}

// View implements Component.
func (b *Breadcrumb) View() string {
	if len(b.items) == 0 {
		return ""
	}

	textColor := lipgloss.Color(b.theme.Text)
	accentColor := lipgloss.Color(b.theme.Accent)
	mutedColor := lipgloss.Color(b.theme.Muted)
	cursorBg := lipgloss.Color(b.theme.Cursor)
	inverseFg := lipgloss.Color(b.theme.TextInverse)

	sepStyle := lipgloss.NewStyle().Foreground(mutedColor)

	var parts []string
	for i, item := range b.items {
		isCursor := i == b.cursor
		isLast := i == len(b.items)-1

		var styled string
		switch {
		case isCursor:
			style := lipgloss.NewStyle().
				Background(cursorBg).
				Foreground(inverseFg)
			styled = style.Render(" " + item.Label + " ")
		case isLast:
			styled = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(item.Label)
		default:
			styled = lipgloss.NewStyle().Foreground(textColor).Render(item.Label)
		}

		parts = append(parts, styled)
		if !isLast {
			parts = append(parts, sepStyle.Render(b.opts.Separator))
		}
	}

	return strings.Join(parts, "")
}

// KeyBindings implements Component.
func (b *Breadcrumb) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "left/h", Label: "Previous", Group: "BREADCRUMB"},
		{Key: "right/l", Label: "Next", Group: "BREADCRUMB"},
		{Key: "enter", Label: "Select", Group: "BREADCRUMB"},
	}
}

// SetSize implements Component.
func (b *Breadcrumb) SetSize(w, h int) {
	b.width = w
	b.height = h
}

// Focused implements Component.
func (b *Breadcrumb) Focused() bool { return b.focused }

// SetFocused implements Component.
func (b *Breadcrumb) SetFocused(f bool) { b.focused = f }

// SetTheme implements Themed.
func (b *Breadcrumb) SetTheme(theme Theme) { b.theme = theme }
