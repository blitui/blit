package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KanbanCard is a single card in a Kanban column.
type KanbanCard struct {
	// ID is a unique identifier for the card.
	ID string
	// Title is the display text.
	Title string
	// Description is optional secondary text.
	Description string
	// Tag is an optional label shown as a badge (e.g. "bug", "feat").
	Tag string
}

// KanbanColumn is a named column containing cards.
type KanbanColumn struct {
	// Title is the column header.
	Title string
	// Cards are the items in this column.
	Cards []KanbanCard
}

// KanbanOpts configures a Kanban board.
type KanbanOpts struct {
	// OnMove is called when a card is moved between columns.
	// It receives the card, source column index, and destination column index.
	OnMove func(card KanbanCard, from, to int)

	// OnSelect is called when enter is pressed on a card.
	OnSelect func(card KanbanCard, col int)
}

// Kanban is a multi-column board component for organizing cards.
// Implements Component and Themed.
type Kanban struct {
	opts    KanbanOpts
	columns []KanbanColumn
	colIdx  int // active column
	cardIdx int // cursor within active column
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewKanban creates a Kanban board with the given columns and options.
func NewKanban(columns []KanbanColumn, opts KanbanOpts) *Kanban {
	return &Kanban{
		opts:    opts,
		columns: columns,
	}
}

// Columns returns the current columns.
func (k *Kanban) Columns() []KanbanColumn { return k.columns }

// ActiveColumn returns the index of the focused column.
func (k *Kanban) ActiveColumn() int { return k.colIdx }

// ActiveCard returns the index of the focused card within the active column.
func (k *Kanban) ActiveCard() int { return k.cardIdx }

// Init implements Component.
func (k *Kanban) Init() tea.Cmd { return nil }

// Update implements Component.
func (k *Kanban) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !k.focused {
		return k, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return k, nil
	}
	switch km.String() {
	case "left", "h":
		if k.colIdx > 0 {
			k.colIdx--
			k.clampCard()
		}
		return k, Consumed()
	case "right", "l":
		if k.colIdx < len(k.columns)-1 {
			k.colIdx++
			k.clampCard()
		}
		return k, Consumed()
	case "up", "k":
		if k.cardIdx > 0 {
			k.cardIdx--
		}
		return k, Consumed()
	case "down", "j":
		col := k.activeCol()
		if col != nil && k.cardIdx < len(col.Cards)-1 {
			k.cardIdx++
		}
		return k, Consumed()
	case "enter":
		col := k.activeCol()
		if col != nil && k.cardIdx < len(col.Cards) && k.opts.OnSelect != nil {
			k.opts.OnSelect(col.Cards[k.cardIdx], k.colIdx)
		}
		return k, Consumed()
	case "L", "shift+right":
		k.moveCard(1)
		return k, Consumed()
	case "H", "shift+left":
		k.moveCard(-1)
		return k, Consumed()
	}
	return k, nil
}

// View implements Component.
func (k *Kanban) View() string {
	if k.width == 0 || k.height == 0 || len(k.columns) == 0 {
		return ""
	}

	colWidth := k.width / len(k.columns)
	if colWidth < 10 {
		colWidth = 10
	}

	mutedColor := lipgloss.Color(k.theme.Muted)
	textColor := lipgloss.Color(k.theme.Text)
	accentColor := lipgloss.Color(k.theme.Accent)
	cursorBg := lipgloss.Color(k.theme.Cursor)
	inverseFg := lipgloss.Color(k.theme.TextInverse)
	warnColor := lipgloss.Color(k.theme.Warn)

	var colViews []string
	for ci, col := range k.columns {
		isActiveCol := ci == k.colIdx

		// Column header.
		headerStyle := lipgloss.NewStyle().
			Width(colWidth - 2).
			Align(lipgloss.Center).
			Bold(true)
		if isActiveCol {
			headerStyle = headerStyle.Foreground(accentColor)
		} else {
			headerStyle = headerStyle.Foreground(mutedColor)
		}
		header := headerStyle.Render(col.Title)

		sep := lipgloss.NewStyle().
			Foreground(mutedColor).
			Render(strings.Repeat("─", colWidth-2))

		// Cards.
		var cardLines []string
		for ci2, card := range col.Cards {
			isCursor := isActiveCol && ci2 == k.cardIdx
			cardView := k.renderCard(card, colWidth-2, isCursor, textColor, cursorBg, inverseFg, accentColor, mutedColor, warnColor)
			cardLines = append(cardLines, cardView)
		}

		if len(cardLines) == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(mutedColor).
				Width(colWidth - 2).
				Align(lipgloss.Center)
			cardLines = append(cardLines, emptyStyle.Render("(empty)"))
		}

		content := header + "\n" + sep + "\n" + strings.Join(cardLines, "\n")

		borderColor := mutedColor
		if isActiveCol {
			borderColor = accentColor
		}

		borders := DefaultBorders()
		if k.theme.Borders != nil {
			borders = *k.theme.Borders
		}

		colStyle := lipgloss.NewStyle().
			Border(borders.Rounded).
			BorderForeground(borderColor).
			Width(colWidth).
			Height(k.height - 2) // account for border

		colViews = append(colViews, colStyle.Render(content))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, colViews...)
}

func (k *Kanban) renderCard(card KanbanCard, width int, isCursor bool, textColor, cursorBg, inverseFg, accentColor, mutedColor, warnColor lipgloss.Color) string {
	var parts []string

	title := card.Title
	if card.Tag != "" {
		tagStyle := lipgloss.NewStyle().Foreground(warnColor)
		title = tagStyle.Render("["+card.Tag+"]") + " " + title
	}

	if isCursor {
		style := lipgloss.NewStyle().
			Background(cursorBg).
			Foreground(inverseFg).
			Width(width)
		parts = append(parts, style.Render(title))
	} else {
		style := lipgloss.NewStyle().
			Foreground(textColor).
			Width(width)
		parts = append(parts, style.Render(title))
	}

	if card.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(width)
		parts = append(parts, descStyle.Render("  "+card.Description))
	}

	return strings.Join(parts, "\n")
}

// KeyBindings implements Component.
func (k *Kanban) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "left/h", Label: "Previous column", Group: "KANBAN"},
		{Key: "right/l", Label: "Next column", Group: "KANBAN"},
		{Key: "up/k", Label: "Previous card", Group: "KANBAN"},
		{Key: "down/j", Label: "Next card", Group: "KANBAN"},
		{Key: "enter", Label: "Select card", Group: "KANBAN"},
		{Key: "H/L", Label: "Move card left/right", Group: "KANBAN"},
	}
}

// SetSize implements Component.
func (k *Kanban) SetSize(w, h int) {
	k.width = w
	k.height = h
}

// Focused implements Component.
func (k *Kanban) Focused() bool { return k.focused }

// SetFocused implements Component.
func (k *Kanban) SetFocused(f bool) { k.focused = f }

// SetTheme implements Themed.
func (k *Kanban) SetTheme(theme Theme) { k.theme = theme }

func (k *Kanban) activeCol() *KanbanColumn {
	if k.colIdx >= 0 && k.colIdx < len(k.columns) {
		return &k.columns[k.colIdx]
	}
	return nil
}

func (k *Kanban) clampCard() {
	col := k.activeCol()
	if col == nil || len(col.Cards) == 0 {
		k.cardIdx = 0
		return
	}
	if k.cardIdx >= len(col.Cards) {
		k.cardIdx = len(col.Cards) - 1
	}
}

func (k *Kanban) moveCard(dir int) {
	col := k.activeCol()
	if col == nil || len(col.Cards) == 0 {
		return
	}
	destIdx := k.colIdx + dir
	if destIdx < 0 || destIdx >= len(k.columns) {
		return
	}
	if k.cardIdx >= len(col.Cards) {
		return
	}

	card := col.Cards[k.cardIdx]
	fromIdx := k.colIdx

	// Remove from source.
	col.Cards = append(col.Cards[:k.cardIdx], col.Cards[k.cardIdx+1:]...)
	k.columns[k.colIdx] = *col

	// Add to destination.
	dest := &k.columns[destIdx]
	dest.Cards = append(dest.Cards, card)

	// Move cursor to destination column.
	k.colIdx = destIdx
	k.cardIdx = len(dest.Cards) - 1

	if k.opts.OnMove != nil {
		k.opts.OnMove(card, fromIdx, destIdx)
	}
}
