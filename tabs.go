package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabItem is a single tab entry with a title, optional glyph, and content component.
type TabItem struct {
	// Title is the label shown in the tab bar.
	Title string
	// Glyph is an optional icon prefix shown before the title.
	Glyph string
	// Content is the component rendered when this tab is active.
	Content Component
}

// Tabs is a Component that shows a row or column of named tabs with switchable content panes.
// Keybinds: tab/shift+tab to cycle, 1-9 to jump, left-click to select.
type Tabs struct {
	items       []TabItem
	active      int
	orientation Orientation
	onChange    func(int)
	theme       Theme
	focused     bool
	width       int
	height      int
}

// TabsOpts configures a Tabs component.
type TabsOpts struct {
	// Orientation is Horizontal (default) or Vertical.
	Orientation Orientation
	// OnChange is called whenever the active tab changes.
	OnChange func(int)
}

// NewTabs creates a new Tabs component with the given items and options.
func NewTabs(items []TabItem, opts TabsOpts) *Tabs {
	t := &Tabs{
		items:       items,
		orientation: opts.Orientation,
		onChange:    opts.OnChange,
		theme:       DefaultTheme(),
	}
	// Propagate initial size/theme to content components.
	return t
}

// SetActive sets the active tab by index (clamped to valid range).
func (t *Tabs) SetActive(i int) {
	if len(t.items) == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(t.items) {
		i = len(t.items) - 1
	}
	if i == t.active {
		return
	}
	prev := t.active
	t.active = i
	t.syncContentSize()
	if t.onChange != nil && prev != t.active {
		t.onChange(t.active)
	}
}

// OnChange registers a callback invoked when the active tab changes.
func (t *Tabs) OnChange(fn func(int)) {
	t.onChange = fn
}

// ActiveIndex returns the currently active tab index.
func (t *Tabs) ActiveIndex() int {
	return t.active
}

// Init implements Component.
func (t *Tabs) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, item := range t.items {
		if item.Content != nil {
			if cmd := item.Content.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

// Update implements Component.
func (t *Tabs) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := t.handleKey(msg)
		if isConsumed(cmd) {
			return t, cmd
		}
		// Forward unconsumed keys to active content.
		if c := t.activeContent(); c != nil && c.Focused() {
			updated, cmd2 := c.Update(msg, ctx)
			t.items[t.active].Content = updated
			return t, cmd2
		}
		return t, cmd
	case tea.MouseMsg:
		cmd := t.handleMouse(msg)
		if isConsumed(cmd) {
			return t, cmd
		}
		// Forward mouse to active content.
		if c := t.activeContent(); c != nil {
			updated, cmd2 := c.Update(msg, ctx)
			t.items[t.active].Content = updated
			return t, cmd2
		}
		return t, nil
	default:
		// Broadcast to active content.
		if c := t.activeContent(); c != nil {
			updated, cmd := c.Update(msg, ctx)
			t.items[t.active].Content = updated
			return t, cmd
		}
	}
	return t, nil
}

func (t *Tabs) handleKey(msg tea.KeyMsg) tea.Cmd {
	if len(t.items) == 0 {
		return nil
	}
	switch msg.String() {
	case "tab":
		next := (t.active + 1) % len(t.items)
		t.SetActive(next)
		return Consumed()
	case "shift+tab":
		prev := (t.active - 1 + len(t.items)) % len(t.items)
		t.SetActive(prev)
		return Consumed()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.Runes[0] - '1')
		if idx < len(t.items) {
			t.SetActive(idx)
			return Consumed()
		}
	}
	return nil
}

func (t *Tabs) handleMouse(msg tea.MouseMsg) tea.Cmd {
	if msg.Button != tea.MouseButtonLeft {
		return nil
	}
	if msg.Action == tea.MouseActionRelease {
		return nil
	}
	if t.orientation == Horizontal {
		return t.handleMouseHorizontal(msg)
	}
	return t.handleMouseVertical(msg)
}

func (t *Tabs) handleMouseHorizontal(msg tea.MouseMsg) tea.Cmd {
	// Tab bar is row 0. Calculate which tab was clicked by scanning X positions.
	if msg.Y != 0 {
		return nil
	}
	x := 0
	for i, item := range t.items {
		label := t.tabLabel(item)
		w := lipgloss.Width(label) + 2 // +2 for padding
		if msg.X >= x && msg.X < x+w {
			t.SetActive(i)
			return Consumed()
		}
		x += w + 1 // +1 for separator
	}
	return nil
}

func (t *Tabs) handleMouseVertical(msg tea.MouseMsg) tea.Cmd {
	// Tab bar is column 0..tabBarWidth-1.
	tabBarWidth := t.verticalBarWidth()
	if msg.X >= tabBarWidth {
		return nil
	}
	if msg.Y >= 0 && msg.Y < len(t.items) {
		t.SetActive(msg.Y)
		return Consumed()
	}
	return nil
}

// View implements Component.
func (t *Tabs) View() string {
	if len(t.items) == 0 {
		return ""
	}
	if t.orientation == Vertical {
		return t.viewVertical()
	}
	return t.viewHorizontal()
}

func (t *Tabs) viewHorizontal() string {
	// ── Tab bar (row of tabs) ──────────────────────────────────────────────
	tabBar := t.renderHorizontalBar()

	// ── Content area ──────────────────────────────────────────────────────
	content := ""
	if c := t.activeContent(); c != nil {
		content = c.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}

func (t *Tabs) viewVertical() string {
	barWidth := t.verticalBarWidth()

	// Build sidebar lines.
	var sideLines []string
	for i, item := range t.items {
		label := t.tabLabel(item)
		var style lipgloss.Style
		if i == t.active {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.theme.TextInverse)).
				Background(lipgloss.Color(t.theme.Accent)).
				Bold(true).
				Width(barWidth).
				PaddingLeft(1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.theme.Muted)).
				Width(barWidth).
				PaddingLeft(1)
		}
		sideLines = append(sideLines, style.Render(label))
	}

	sidebar := strings.Join(sideLines, "\n")
	contentStr := ""
	if c := t.activeContent(); c != nil {
		contentStr = c.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, contentStr)
}

func (t *Tabs) renderHorizontalBar() string {
	var parts []string
	for i, item := range t.items {
		label := t.tabLabel(item)
		if i == t.active {
			// Active tab: accent foreground, bold, underline indicator below
			tab := lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.theme.Accent)).
				Bold(true).
				PaddingLeft(1).PaddingRight(1).
				Render(label)
			// Underline using a separate render so it occupies a consistent width
			parts = append(parts, tab)
		} else {
			tab := lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.theme.Muted)).
				PaddingLeft(1).PaddingRight(1).
				Render(label)
			parts = append(parts, tab)
		}
	}

	tabRow := strings.Join(parts, lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Border)).Render("│"))

	// Build underline row: accent chars under active tab, dashes elsewhere.
	underlineParts := t.buildUnderline(parts)
	underlineRow := strings.Join(underlineParts, "┼")

	borderLine := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Border)).Render(
		strings.Repeat("─", t.width),
	)

	return lipgloss.JoinVertical(lipgloss.Left, tabRow, underlineRow, borderLine)
}

func (t *Tabs) buildUnderline(renderedParts []string) []string {
	parts := make([]string, len(renderedParts))
	for i, p := range renderedParts {
		w := lipgloss.Width(p)
		if i == t.active {
			// Gradient-like underline using accent color block characters.
			bar := strings.Repeat("▔", w)
			parts[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.theme.Accent)).
				Render(bar)
		} else {
			parts[i] = strings.Repeat(" ", w)
		}
	}
	return parts
}

func (t *Tabs) tabLabel(item TabItem) string {
	if item.Glyph != "" {
		return item.Glyph + " " + item.Title
	}
	return item.Title
}

func (t *Tabs) verticalBarWidth() int {
	max := 0
	for _, item := range t.items {
		w := lipgloss.Width(t.tabLabel(item)) + 2 // +2 for padding
		if w > max {
			max = w
		}
	}
	if max < 12 {
		max = 12
	}
	return max
}

// KeyBindings implements Component.
func (t *Tabs) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "tab", Label: "Next tab", Group: "TABS"},
		{Key: "shift+tab", Label: "Previous tab", Group: "TABS"},
		{Key: "1-9", Label: "Jump to tab", Group: "TABS"},
	}
}

// SetSize implements Component.
func (t *Tabs) SetSize(w, h int) {
	t.width = w
	t.height = h
	t.syncContentSize()
}

// syncContentSize propagates the available content area size to all content components.
func (t *Tabs) syncContentSize() {
	if t.width == 0 && t.height == 0 {
		return
	}
	cw, ch := t.contentSize()
	for _, item := range t.items {
		if item.Content != nil {
			item.Content.SetSize(cw, ch)
		}
	}
}

// contentSize returns the width/height available for content (excluding the tab bar).
func (t *Tabs) contentSize() (int, int) {
	if t.orientation == Vertical {
		barW := t.verticalBarWidth()
		return t.width - barW, t.height
	}
	// Horizontal: 3 rows for bar (tab row + underline + border line)
	barH := 3
	h := t.height - barH
	if h < 0 {
		h = 0
	}
	return t.width, h
}

// Focused implements Component.
func (t *Tabs) Focused() bool { return t.focused }

// SetFocused implements Component. When focused, the active content is also focused.
func (t *Tabs) SetFocused(focused bool) {
	t.focused = focused
	for i, item := range t.items {
		if item.Content != nil {
			item.Content.SetFocused(focused && i == t.active)
		}
	}
}

// SetTheme implements Themed.
func (t *Tabs) SetTheme(th Theme) {
	t.theme = th
	for _, item := range t.items {
		if item.Content != nil {
			if themed, ok := item.Content.(Themed); ok {
				themed.SetTheme(th)
			}
		}
	}
}

func (t *Tabs) activeContent() Component {
	if t.active >= 0 && t.active < len(t.items) {
		return t.items[t.active].Content
	}
	return nil
}
