package tuikit

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moneycaringcoder/tuikit-go/internal/fuzzy"
)

// PickerItem is a single entry in a Picker.
type PickerItem struct {
	// Title is the primary display label.
	Title string

	// Subtitle is secondary text shown below the title. Optional.
	Subtitle string

	// Glyph is an optional icon/glyph prefix.
	Glyph string

	// Preview returns content shown in the preview pane when this item is
	// highlighted. Called lazily on cursor change. Optional.
	Preview func() string

	// Value is an arbitrary payload the caller can attach.
	Value any
}

// PickerOpts configures Picker behaviour.
type PickerOpts struct {
	// Placeholder text shown in the filter input when empty.
	Placeholder string

	// Preview enables the right-side preview pane (40% width).
	Preview bool

	// OnConfirm is called when the user presses Enter on an item. Optional.
	OnConfirm func(item PickerItem)

	// OnCancel is called when the user closes the picker without confirming.
	OnCancel func()
}

// Picker is a fuzzy-searchable command palette that can be used as either a
// full-screen Overlay (Cmd-K style) or as an embedded Component.
type Picker struct {
	opts PickerOpts

	items    []PickerItem
	filtered []rankedItem
	cursor   int

	input   textinput.Model
	theme   Theme
	focused bool
	active  bool

	width  int
	height int

	previewContent string
	previewCursor  int
}

type rankedItem struct {
	item  PickerItem
	score float64
	orig  int
}

// NewPicker creates a Picker with the given items and options.
func NewPicker(items []PickerItem, opts PickerOpts) *Picker {
	ti := textinput.New()
	ti.Placeholder = opts.Placeholder
	if ti.Placeholder == "" {
		ti.Placeholder = "Type to filter..."
	}
	ti.Focus()

	p := &Picker{
		opts:  opts,
		items: items,
		input: ti,
	}
	p.rebuildFiltered()
	return p
}

// SetItems replaces the item list and re-runs the filter.
func (p *Picker) SetItems(items []PickerItem) {
	p.items = items
	p.rebuildFiltered()
}

// Items returns the full (unfiltered) item list.
func (p *Picker) Items() []PickerItem { return p.items }

// CursorItem returns the currently highlighted item, or nil if there are none.
func (p *Picker) CursorItem() *PickerItem {
	if p.cursor >= 0 && p.cursor < len(p.filtered) {
		item := p.filtered[p.cursor].item
		return &item
	}
	return nil
}

// Init implements Component.
func (p *Picker) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements Component.
func (p *Picker) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := p.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	var inputCmd tea.Cmd
	p.input, inputCmd = p.input.Update(msg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	p.rebuildFiltered()

	return p, tea.Batch(cmds...)
}

func (p *Picker) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "ctrl+p":
		if p.cursor > 0 {
			p.cursor--
			p.invalidatePreview()
		}
		return Consumed()

	case "down", "ctrl+n":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
			p.invalidatePreview()
		}
		return Consumed()

	case "enter":
		if p.opts.OnConfirm != nil {
			if item := p.CursorItem(); item != nil {
				p.active = false
				p.opts.OnConfirm(*item)
			}
		}
		return Consumed()

	case "esc":
		p.active = false
		if p.opts.OnCancel != nil {
			p.opts.OnCancel()
		}
		return Consumed()

	case "ctrl+k":
		p.input.Reset()
		p.rebuildFiltered()
		return Consumed()
	}

	return nil
}

// View implements Component.
func (p *Picker) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	listWidth := p.width
	if p.opts.Preview {
		listWidth = p.width * 6 / 10
	}

	list := p.renderList(listWidth)

	if !p.opts.Preview {
		return list
	}

	preview := p.renderPreview(p.width - listWidth - 1)
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Border)).
		Render(strings.Repeat("|\n", p.height-1) + "|")

	return lipgloss.JoinHorizontal(lipgloss.Top, list, divider, preview)
}

func (p *Picker) renderList(width int) string {
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.theme.Accent)).
		Width(width-2).
		Padding(0, 1)

	inputView := inputStyle.Render(p.input.View())
	inputHeight := strings.Count(inputView, "\n") + 1

	availHeight := p.height - inputHeight - 1
	if availHeight < 1 {
		availHeight = 1
	}

	var itemLines []string
	start := 0
	if p.cursor >= availHeight {
		start = p.cursor - availHeight + 1
	}
	end := start + availHeight
	if end > len(p.filtered) {
		end = len(p.filtered)
	}

	// Use row.cursor from the style registry when available.
	var cursorStyle lipgloss.Style
	if ss, ok := p.theme.Style("row.cursor"); ok {
		cursorStyle = ss.Focus.Width(width)
	} else {
		cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.theme.TextInverse)).
			Background(lipgloss.Color(p.theme.Cursor)).
			Width(width)
	}
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Text)).
		Width(width)
	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Muted))
	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Accent))

	for i := start; i < end; i++ {
		ri := p.filtered[i]
		isCursor := i == p.cursor

		glyph := ri.item.Glyph
		if glyph == "" {
			glyph = "  "
		}

		title := ri.item.Title
		sub := ri.item.Subtitle

		var line string
		if isCursor {
			line = accentStyle.Render(glyph) + cursorStyle.Render(" "+title)
			if sub != "" {
				line += "\n" + cursorStyle.Render("    "+sub)
			}
		} else {
			line = mutedStyle.Render(glyph) + normalStyle.Render(" "+title)
			if sub != "" {
				line += "\n" + mutedStyle.Render("    "+sub)
			}
		}
		itemLines = append(itemLines, line)
	}

	if len(itemLines) == 0 {
		empty := mutedStyle.Render("  No results")
		itemLines = []string{empty}
	}

	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Muted))
	info := countStyle.Render(" " + pickerItoa(len(p.filtered)) + "/" + pickerItoa(len(p.items)) + " items")

	listContent := strings.Join(itemLines, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, inputView, listContent, info)
}

func (p *Picker) renderPreview(width int) string {
	if width < 2 {
		return ""
	}

	p.ensurePreview()

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Accent)).
		Bold(true).
		Width(width)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.theme.Text)).
		Width(width).
		Height(p.height - 2)

	var title string
	if item := p.CursorItem(); item != nil {
		title = item.Title
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		contentStyle.Render(p.previewContent),
	)
}

func (p *Picker) ensurePreview() {
	if p.previewCursor == p.cursor && p.previewContent != "" {
		return
	}
	p.previewCursor = p.cursor
	p.previewContent = ""
	if item := p.CursorItem(); item != nil && item.Preview != nil {
		p.previewContent = item.Preview()
	}
}

func (p *Picker) invalidatePreview() {
	p.previewContent = ""
}

// KeyBindings implements Component.
func (p *Picker) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/ctrl+p", Label: "Move up", Group: "PICKER"},
		{Key: "down/ctrl+n", Label: "Move down", Group: "PICKER"},
		{Key: "enter", Label: "Confirm", Group: "PICKER"},
		{Key: "esc", Label: "Cancel", Group: "PICKER"},
		{Key: "ctrl+k", Label: "Clear filter", Group: "PICKER"},
	}
}

// SetSize implements Component.
func (p *Picker) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.input.Width = w - 6
}

// Focused implements Component.
func (p *Picker) Focused() bool { return p.focused }

// SetFocused implements Component.
func (p *Picker) SetFocused(f bool) {
	p.focused = f
	if f {
		p.input.Focus()
		p.active = true
	} else {
		p.input.Blur()
	}
}

// SetTheme implements Themed.
func (p *Picker) SetTheme(t Theme) {
	p.theme = t
	p.input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Text))
	p.input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted))
	p.input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
}

// IsActive implements Overlay.
func (p *Picker) IsActive() bool { return p.active }

// Close implements Overlay.
func (p *Picker) Close() {
	p.active = false
	p.input.Blur()
}

// Open activates the picker and focuses the input.
func (p *Picker) Open() {
	p.active = true
	p.input.Reset()
	p.cursor = 0
	p.rebuildFiltered()
	p.input.Focus()
}

func (p *Picker) rebuildFiltered() {
	query := p.input.Value()
	p.filtered = p.filtered[:0]

	for i, item := range p.items {
		if query == "" {
			p.filtered = append(p.filtered, rankedItem{item: item, score: 1, orig: i})
			continue
		}
		target := item.Title
		if item.Subtitle != "" {
			target += " " + item.Subtitle
		}
		m := fuzzy.Score(query, target)
		if m.Score > 0 {
			p.filtered = append(p.filtered, rankedItem{item: item, score: m.Score, orig: i})
		}
	}

	sortRanked(p.filtered)

	if p.cursor >= len(p.filtered) {
		p.cursor = max(0, len(p.filtered)-1)
	}

	p.invalidatePreview()
}

func sortRanked(items []rankedItem) {
	for i := 1; i < len(items); i++ {
		j := i
		for j > 0 && items[j].score > items[j-1].score {
			items[j], items[j-1] = items[j-1], items[j]
			j--
		}
	}
}

func pickerItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
