package blit

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListViewOpts configures a ListView component.
type ListViewOpts[T any] struct {
	// RenderItem renders a single item to a string.
	// Parameters: item, index in the list, whether this item has cursor, theme.
	RenderItem func(item T, idx int, isCursor bool, theme Theme) string

	// HeaderFunc renders a header above the list content. Optional.
	// The returned string can be multi-line. ListView accounts for its height.
	HeaderFunc func(theme Theme) string

	// DetailFunc renders a detail bar below the list for the selected item. Optional.
	// Only shown when the list is focused and there is a cursor item.
	// The returned string can be multi-line. ListView accounts for its height.
	DetailFunc func(item T, theme Theme) string

	// FlashFunc returns true if an item should show the flash marker. Optional.
	// Called during render with the current time for duration-based flash effects.
	FlashFunc func(item T, now time.Time) bool

	// OnSelect is called when Enter is pressed on an item. Optional.
	OnSelect func(item T, idx int)

	// DetailHeight is the number of lines to reserve for the detail bar.
	// When DetailFunc is set, this space is always reserved (blank when unfocused)
	// to prevent viewport jitter on focus change. Default: 3.
	DetailHeight int
}

// ListView is a generic scrollable list with cursor navigation, header, detail
// bar, and flash highlighting. It implements Component and Themed.
//
// Use it standalone as a registered component, or embed it inside a custom
// component and delegate via HandleKey and View.
type ListView[T any] struct {
	items       []T
	opts        ListViewOpts[T]
	theme       Theme
	focused     bool
	width       int
	height      int
	cursor      int
	viewport    viewport.Model
	ready       bool
	cursorTween Tween // 120ms cursor highlight fade-in
}

// NewListView creates a new ListView with the given options.
func NewListView[T any](opts ListViewOpts[T]) *ListView[T] {
	if opts.DetailHeight == 0 && opts.DetailFunc != nil {
		opts.DetailHeight = 3
	}
	return &ListView[T]{opts: opts}
}

// SetItems replaces the list data and rebuilds the view.
func (l *ListView[T]) SetItems(items []T) {
	l.items = items
	l.clampCursor()
	if l.ready {
		l.rebuildContent()
	}
}

// Items returns the current items.
func (l *ListView[T]) Items() []T {
	return l.items
}

// CursorItem returns a pointer to the item at the cursor, or nil if empty.
func (l *ListView[T]) CursorItem() *T {
	if l.cursor >= 0 && l.cursor < len(l.items) {
		return &l.items[l.cursor]
	}
	return nil
}

// CursorIndex returns the current cursor position.
func (l *ListView[T]) CursorIndex() int {
	return l.cursor
}

// SetCursor moves the cursor to the given index.
func (l *ListView[T]) SetCursor(idx int) {
	l.cursor = idx
	l.clampCursor()
	if l.ready {
		l.rebuildContent()
		l.ensureCursorVisible()
	}
}

// ScrollToTop moves the cursor and viewport to the top.
func (l *ListView[T]) ScrollToTop() {
	l.cursor = 0
	l.clampCursor()
	if l.ready {
		l.rebuildContent()
		l.viewport.GotoTop()
	}
}

// ScrollToBottom moves the cursor and viewport to the bottom.
func (l *ListView[T]) ScrollToBottom() {
	l.cursor = max(0, len(l.items)-1)
	l.clampCursor()
	if l.ready {
		l.rebuildContent()
		l.viewport.GotoBottom()
	}
}

// IsAtTop returns true if the viewport is scrolled to the top.
func (l *ListView[T]) IsAtTop() bool {
	return l.viewport.YOffset == 0
}

// IsAtBottom returns true if the viewport is scrolled to the bottom.
func (l *ListView[T]) IsAtBottom() bool {
	return l.viewport.AtBottom()
}

// Refresh re-renders the list content without changing items.
// Use this to update flash effects or dynamic rendering.
func (l *ListView[T]) Refresh() {
	if l.ready {
		l.rebuildContent()
	}
}

// ItemCount returns the number of items in the list.
func (l *ListView[T]) ItemCount() int {
	return len(l.items)
}

func (l *ListView[T]) Init() tea.Cmd { return nil }

func (l *ListView[T]) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := l.HandleKey(msg)
		return l, cmd
	case animTickMsg:
		if l.cursorTween.Running() {
			l.rebuildContent()
			return l, nil
		}
	}
	return l, nil
}

// HandleKey processes navigation keys. Use this when embedding ListView
// in another component. Returns Consumed() if handled, nil otherwise.
func (l *ListView[T]) HandleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if l.cursor > 0 {
			l.cursor--
			l.startCursorTween()
		}
		l.rebuildContent()
		l.ensureCursorVisible()
		return Consumed()
	case "down", "j":
		if l.cursor < len(l.items)-1 {
			l.cursor++
			l.startCursorTween()
		}
		l.rebuildContent()
		l.ensureCursorVisible()
		return Consumed()
	case "home", "g":
		l.startCursorTween()
		l.cursor = 0
		l.rebuildContent()
		l.ensureCursorVisible()
		return Consumed()
	case "end", "G":
		l.startCursorTween()
		l.cursor = max(0, len(l.items)-1)
		l.rebuildContent()
		l.ensureCursorVisible()
		return Consumed()
	case "enter":
		if l.opts.OnSelect != nil && l.cursor >= 0 && l.cursor < len(l.items) {
			l.opts.OnSelect(l.items[l.cursor], l.cursor)
			return Consumed()
		}
	}
	return nil
}

func (l *ListView[T]) View() string {
	if !l.ready {
		return ""
	}

	var sections []string

	// Header
	if l.opts.HeaderFunc != nil {
		header := l.opts.HeaderFunc(l.theme)
		if header != "" {
			sections = append(sections, header)
		}
	}

	// Viewport content
	vpView := strings.TrimRight(l.viewport.View(), "\n")
	sections = append(sections, vpView)

	// Detail bar
	if l.opts.DetailFunc != nil {
		detail := l.renderDetail()
		sections = append(sections, detail)
	}

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().MaxWidth(l.width).Render(view)
}

func (l *ListView[T]) renderDetail() string {
	if l.focused {
		if item := l.CursorItem(); item != nil {
			detail := l.opts.DetailFunc(*item, l.theme)
			if detail != "" {
				return detail
			}
		}
	}
	// Reserve blank lines to prevent viewport jitter
	blank := strings.Repeat("\n", l.opts.DetailHeight-1)
	return strings.Repeat(" ", l.width) + blank
}

func (l *ListView[T]) KeyBindings() []KeyBind {
	bindings := []KeyBind{
		{Key: "up/k", Label: "Scroll up", Group: "NAVIGATION"},
		{Key: "down/j", Label: "Scroll down", Group: "NAVIGATION"},
		{Key: "home/g", Label: "Go to top", Group: "NAVIGATION"},
		{Key: "end/G", Label: "Go to bottom", Group: "NAVIGATION"},
	}
	if l.opts.OnSelect != nil {
		bindings = append(bindings, KeyBind{Key: "enter", Label: "Select", Group: "NAVIGATION"})
	}
	return bindings
}

func (l *ListView[T]) SetSize(w, h int) {
	l.width = w
	l.height = h

	vpHeight := l.viewportHeight()
	if !l.ready {
		l.viewport = viewport.New(w, vpHeight)
		l.ready = true
	} else {
		l.viewport.Width = w
		l.viewport.Height = vpHeight
	}
	l.rebuildContent()
}

func (l *ListView[T]) viewportHeight() int {
	h := l.height

	// Subtract header height
	if l.opts.HeaderFunc != nil {
		header := l.opts.HeaderFunc(l.theme)
		if header != "" {
			h -= strings.Count(header, "\n") + 1
		}
	}

	// Subtract detail height (always reserved when DetailFunc is set)
	if l.opts.DetailFunc != nil {
		h -= l.opts.DetailHeight
	}

	if h < 1 {
		h = 1
	}
	return h
}

func (l *ListView[T]) Focused() bool     { return l.focused }
func (l *ListView[T]) SetFocused(f bool) { l.focused = f; l.rebuildContent() }
func (l *ListView[T]) SetTheme(th Theme) { l.theme = th }

func (l *ListView[T]) startCursorTween() {
	l.cursorTween = Tween{Duration: 120 * time.Millisecond}
	l.cursorTween.Start(time.Now())
}

func (l *ListView[T]) rebuildContent() {
	if !l.ready {
		return
	}

	now := time.Now()

	// Compute animated cursor colors
	cursorBgColor := lipgloss.Color(l.theme.Cursor)
	cursorFgColor := lipgloss.Color(l.theme.Accent)
	if l.cursorTween.Running() {
		tval := l.cursorTween.Progress(now)
		cursorBgColor = Interpolate[lipgloss.Color](
			lipgloss.Color(l.theme.Muted),
			lipgloss.Color(l.theme.Cursor),
			tval, EaseOutCubic,
		)
		cursorFgColor = Interpolate[lipgloss.Color](
			lipgloss.Color(l.theme.Muted),
			lipgloss.Color(l.theme.Accent),
			tval, EaseOutCubic,
		)
	}

	cursorMarker := lipgloss.NewStyle().
		Foreground(cursorFgColor).
		Bold(true)
	cursorBg := lipgloss.NewStyle().
		Background(cursorBgColor)
	flashStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.theme.Flash)).
		Bold(true)

	glyphs := l.theme.glyphsOrDefault()
	var lines []string
	for i, item := range l.items {
		isCursor := l.focused && i == l.cursor
		line := l.opts.RenderItem(item, i, isCursor, l.theme)

		switch {
		case isCursor:
			line = cursorMarker.Render(glyphs.CursorMarker) + " " + cursorBg.Render(line)
			// Pad to full width with cursor background
			vis := lipgloss.Width(line)
			if vis < l.width {
				line += cursorBg.Render(strings.Repeat(" ", l.width-vis))
			}
		case l.opts.FlashFunc != nil && l.opts.FlashFunc(item, now):
			line = flashStyle.Render(glyphs.FlashMarker) + " " + line
		default:
			line = "  " + line
		}
		lines = append(lines, line)
	}

	l.viewport.SetContent(strings.Join(lines, "\n"))

	// Sync viewport height in case header changed
	vpHeight := l.viewportHeight()
	if l.viewport.Height != vpHeight {
		l.viewport.Height = vpHeight
	}
}

func (l *ListView[T]) ensureCursorVisible() {
	vpHeight := l.viewport.Height
	yOffset := l.viewport.YOffset
	if l.cursor < yOffset {
		l.viewport.SetYOffset(l.cursor)
	} else if l.cursor >= yOffset+vpHeight {
		l.viewport.SetYOffset(l.cursor - vpHeight + 1)
	}
}

func (l *ListView[T]) clampCursor() {
	if l.cursor < 0 {
		l.cursor = 0
	}
	maxCursor := len(l.items) - 1
	if maxCursor < 0 {
		maxCursor = 0
	}
	if l.cursor > maxCursor {
		l.cursor = maxCursor
	}
}
