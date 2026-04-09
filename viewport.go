package tuikit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewportGlyphs holds the scrollbar track and thumb symbols for Viewport.
// Falls back to defaults when nil.
type ViewportGlyphs struct {
	Track string
	Thumb string
}

// defaultViewportGlyphs returns the default scrollbar glyph set.
func defaultViewportGlyphs() ViewportGlyphs {
	return ViewportGlyphs{
		Track: "│",
		Thumb: "█",
	}
}

// Viewport is a themed scroll container for arbitrary string content.
// It renders a right-side scrollbar (track + thumb) pulled from the theme.
// Keybinds: j/k, pgup/pgdn, home/end, ctrl+u/ctrl+d, mouse wheel.
type Viewport struct {
	// Content is the full text to display (may contain newlines).
	Content string

	theme   Theme
	focused bool
	width   int
	height  int
	yOffset int   // current scroll position (line index)
	lines   []string
	ready   bool
}

// NewViewport creates a new Viewport with no content.
func NewViewport() *Viewport {
	return &Viewport{}
}

// SetContent sets the text content and resets the scroll position.
func (v *Viewport) SetContent(content string) {
	v.Content = content
	v.lines = viewportSplitLines(content)
	v.clampOffset()
}

func viewportSplitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// ScrollBy moves the viewport by delta lines (positive = down, negative = up).
func (v *Viewport) ScrollBy(delta int) {
	v.yOffset += delta
	v.clampOffset()
}

// GotoTop scrolls to the top.
func (v *Viewport) GotoTop() {
	v.yOffset = 0
}

// GotoBottom scrolls to the bottom.
func (v *Viewport) GotoBottom() {
	v.yOffset = v.maxOffset()
}

// YOffset returns the current scroll offset.
func (v *Viewport) YOffset() int { return v.yOffset }

// AtTop returns true when scrolled to the top.
func (v *Viewport) AtTop() bool { return v.yOffset == 0 }

// AtBottom returns true when scrolled to the bottom.
func (v *Viewport) AtBottom() bool { return v.yOffset >= v.maxOffset() }

func (v *Viewport) maxOffset() int {
	contentH := len(v.lines)
	viewH := v.viewHeight()
	if contentH <= viewH {
		return 0
	}
	return contentH - viewH
}

func (v *Viewport) viewHeight() int {
	if v.height < 1 {
		return 1
	}
	return v.height
}

// --- Component interface ---

func (v *Viewport) Init() tea.Cmd { return nil }

func (v *Viewport) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := v.HandleKey(msg)
		return v, cmd
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			v.ScrollBy(-3)
			return v, Consumed()
		case tea.MouseButtonWheelDown:
			v.ScrollBy(3)
			return v, Consumed()
		}
	}
	return v, nil
}

// HandleKey processes navigation keys. Returns Consumed() if handled.
func (v *Viewport) HandleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		v.ScrollBy(-1)
		return Consumed()
	case "down", "j":
		v.ScrollBy(1)
		return Consumed()
	case "pgup":
		v.ScrollBy(-v.viewHeight())
		return Consumed()
	case "pgdown":
		v.ScrollBy(v.viewHeight())
		return Consumed()
	case "home":
		v.GotoTop()
		return Consumed()
	case "end":
		v.GotoBottom()
		return Consumed()
	case "ctrl+u":
		v.ScrollBy(-v.viewHeight() / 2)
		return Consumed()
	case "ctrl+d":
		v.ScrollBy(v.viewHeight() / 2)
		return Consumed()
	}
	return nil
}

func (v *Viewport) View() string {
	if !v.ready || v.width < 2 || v.height < 1 {
		return ""
	}

	// Reserve 1 column for the scrollbar.
	contentW := v.width - 1
	if contentW < 1 {
		contentW = 1
	}
	viewH := v.viewHeight()

	// Slice visible lines.
	visibleLines := make([]string, viewH)
	for i := 0; i < viewH; i++ {
		lineIdx := v.yOffset + i
		if lineIdx < len(v.lines) {
			line := v.lines[lineIdx]
			// Truncate to content width.
			if lipgloss.Width(line) > contentW {
				line = truncateLine(line, contentW)
			}
			visibleLines[i] = line
		}
	}

	// Build scrollbar.
	scrollbar := v.renderScrollbar(viewH)

	// Combine each content line with its scrollbar character.
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(v.theme.Text)).
		Width(contentW)

	var rows []string
	for i, line := range visibleLines {
		contentCell := contentStyle.Render(line)
		sbChar := " "
		if i < len(scrollbar) {
			sbChar = scrollbar[i]
		}
		rows = append(rows, contentCell+sbChar)
	}

	return strings.Join(rows, "\n")
}

// renderScrollbar returns a slice of glyph strings, one per visible row.
func (v *Viewport) renderScrollbar(viewH int) []string {
	glyphs := v.theme.glyphsOrDefault()
	vpGlyphs := defaultViewportGlyphs()
	_ = glyphs // may be used for custom overrides later

	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(v.theme.Border))
	thumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(v.theme.Accent))

	totalLines := len(v.lines)

	bar := make([]string, viewH)
	// If all content fits, no scrollbar needed.
	if totalLines <= viewH {
		for i := range bar {
			bar[i] = trackStyle.Render(vpGlyphs.Track)
		}
		return bar
	}

	// Compute thumb position and size.
	thumbH := max(1, viewH*viewH/totalLines)
	thumbTop := (v.yOffset * (viewH - thumbH)) / max(1, totalLines-viewH)

	for i := range bar {
		if i >= thumbTop && i < thumbTop+thumbH {
			bar[i] = thumbStyle.Render(vpGlyphs.Thumb)
		} else {
			bar[i] = trackStyle.Render(vpGlyphs.Track)
		}
	}
	return bar
}

func (v *Viewport) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/k", Label: "Scroll up", Group: "VIEWPORT"},
		{Key: "down/j", Label: "Scroll down", Group: "VIEWPORT"},
		{Key: "pgup", Label: "Page up", Group: "VIEWPORT"},
		{Key: "pgdn", Label: "Page down", Group: "VIEWPORT"},
		{Key: "home", Label: "Go to top", Group: "VIEWPORT"},
		{Key: "end", Label: "Go to bottom", Group: "VIEWPORT"},
		{Key: "ctrl+u", Label: "Half page up", Group: "VIEWPORT"},
		{Key: "ctrl+d", Label: "Half page down", Group: "VIEWPORT"},
	}
}

func (v *Viewport) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.ready = true
	v.clampOffset()
}

func (v *Viewport) Focused() bool       { return v.focused }
func (v *Viewport) SetFocused(f bool)   { v.focused = f }
func (v *Viewport) SetTheme(th Theme)   { v.theme = th }

func (v *Viewport) clampOffset() {
	if v.yOffset < 0 {
		v.yOffset = 0
	}
	if m := v.maxOffset(); v.yOffset > m {
		v.yOffset = m
	}
}

// truncateLine truncates a string to at most maxW visible columns.
func truncateLine(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	var out strings.Builder
	w := 0
	for _, r := range s {
		rw := runeWidth(r)
		if w+rw > maxW {
			break
		}
		out.WriteRune(r)
		w += rw
	}
	return out.String()
}

// runeWidth returns the display width of a rune (1 for ASCII, 2 for wide CJK, etc.).
func runeWidth(r rune) int {
	if r < 0x1100 {
		return 1
	}
	// Wide character ranges (simplified).
	if r >= 0x1100 && r <= 0x115F ||
		r >= 0x2E80 && r <= 0x9FFF ||
		r >= 0xA000 && r <= 0xABFF ||
		r >= 0xF900 && r <= 0xFAFF ||
		r >= 0xFE10 && r <= 0xFE1F ||
		r >= 0xFE30 && r <= 0xFE4F ||
		r >= 0xFF00 && r <= 0xFF60 ||
		r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	}
	return 1
}
