package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TimelineEvent is a single entry in a Timeline.
type TimelineEvent struct {
	// Time is the display label for the timestamp (e.g. "2024-01-15", "3:00 PM").
	Time string
	// Title is the event headline.
	Title string
	// Description is optional detail text.
	Description string
	// Status controls the icon/color: "done", "active", "pending", or custom.
	Status string
}

// TimelineOpts configures a Timeline.
type TimelineOpts struct {
	// OnSelect is called when enter is pressed on an event.
	OnSelect func(event TimelineEvent, index int)

	// Horizontal renders events left-to-right instead of top-to-bottom.
	Horizontal bool
}

// Timeline is a vertical or horizontal sequence of events with connectors.
// Implements Component and Themed.
type Timeline struct {
	opts    TimelineOpts
	events  []TimelineEvent
	cursor  int
	scroll  int
	theme   Theme
	focused bool
	width   int
	height  int
}

// NewTimeline creates a Timeline with the given events and options.
func NewTimeline(events []TimelineEvent, opts TimelineOpts) *Timeline {
	return &Timeline{
		opts:   opts,
		events: events,
	}
}

// Events returns the current events.
func (t *Timeline) Events() []TimelineEvent { return t.events }

// SetEvents replaces the events and resets the cursor.
func (t *Timeline) SetEvents(events []TimelineEvent) {
	t.events = events
	if t.cursor >= len(events) {
		t.cursor = max(0, len(events)-1)
	}
	t.scroll = 0
}

// CursorIndex returns the currently highlighted event index.
func (t *Timeline) CursorIndex() int { return t.cursor }

// Init implements Component.
func (t *Timeline) Init() tea.Cmd { return nil }

// Update implements Component.
func (t *Timeline) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !t.focused {
		return t, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return t, nil
	}
	switch km.String() {
	case "up", "k":
		if !t.opts.Horizontal && t.cursor > 0 {
			t.cursor--
			t.clampScroll()
		}
		return t, Consumed()
	case "down", "j":
		if !t.opts.Horizontal && t.cursor < len(t.events)-1 {
			t.cursor++
			t.clampScroll()
		}
		return t, Consumed()
	case "left", "h":
		if t.opts.Horizontal && t.cursor > 0 {
			t.cursor--
		}
		return t, Consumed()
	case "right", "l":
		if t.opts.Horizontal && t.cursor < len(t.events)-1 {
			t.cursor++
		}
		return t, Consumed()
	case "enter":
		if t.cursor < len(t.events) && t.opts.OnSelect != nil {
			t.opts.OnSelect(t.events[t.cursor], t.cursor)
		}
		return t, Consumed()
	}
	return t, nil
}

// View implements Component.
func (t *Timeline) View() string {
	if t.width == 0 || t.height == 0 || len(t.events) == 0 {
		return ""
	}
	if t.opts.Horizontal {
		return t.viewHorizontal()
	}
	return t.viewVertical()
}

func (t *Timeline) viewVertical() string {
	g := t.theme.glyphsOrDefault()
	textColor := lipgloss.Color(t.theme.Text)
	accentColor := lipgloss.Color(t.theme.Accent)
	mutedColor := lipgloss.Color(t.theme.Muted)
	positiveColor := lipgloss.Color(t.theme.Positive)
	cursorBg := lipgloss.Color(t.theme.Cursor)
	inverseFg := lipgloss.Color(t.theme.TextInverse)

	end := t.scroll + t.height/2 // each event takes ~2 lines
	if end > len(t.events) {
		end = len(t.events)
	}
	if end <= t.scroll {
		end = len(t.events)
	}

	var lines []string
	for i := t.scroll; i < end; i++ {
		ev := t.events[i]
		isCursor := i == t.cursor
		isLast := i == len(t.events)-1

		// Status icon.
		icon := t.statusIcon(ev.Status, g)
		iconColor := t.statusColor(ev.Status, positiveColor, accentColor, mutedColor)

		iconStyled := lipgloss.NewStyle().Foreground(iconColor).Render(icon)

		// Connector line.
		connector := lipgloss.NewStyle().Foreground(mutedColor).Render("│")

		// Time label.
		timeStyle := lipgloss.NewStyle().Foreground(mutedColor)
		timePart := timeStyle.Render(ev.Time)

		// Title.
		var titlePart string
		if isCursor {
			style := lipgloss.NewStyle().
				Background(cursorBg).
				Foreground(inverseFg)
			titlePart = style.Render(" " + ev.Title + " ")
		} else {
			titlePart = lipgloss.NewStyle().Foreground(textColor).Render(ev.Title)
		}

		line := iconStyled + " " + timePart + "  " + titlePart

		if ev.Description != "" {
			descStyle := lipgloss.NewStyle().Foreground(mutedColor)
			line += "\n" + connector + "   " + descStyle.Render(ev.Description)
		}

		if !isLast {
			line += "\n" + connector
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (t *Timeline) viewHorizontal() string {
	textColor := lipgloss.Color(t.theme.Text)
	accentColor := lipgloss.Color(t.theme.Accent)
	mutedColor := lipgloss.Color(t.theme.Muted)
	positiveColor := lipgloss.Color(t.theme.Positive)
	cursorBg := lipgloss.Color(t.theme.Cursor)
	inverseFg := lipgloss.Color(t.theme.TextInverse)
	g := t.theme.glyphsOrDefault()

	var iconParts []string
	var lineParts []string
	var labelParts []string

	segWidth := t.width / max(len(t.events), 1)
	if segWidth < 6 {
		segWidth = 6
	}

	for i, ev := range t.events {
		isCursor := i == t.cursor
		icon := t.statusIcon(ev.Status, g)
		iconColor := t.statusColor(ev.Status, positiveColor, accentColor, mutedColor)

		iconStyle := lipgloss.NewStyle().
			Width(segWidth).
			Align(lipgloss.Center).
			Foreground(iconColor)
		iconParts = append(iconParts, iconStyle.Render(icon))

		// Connector.
		dash := strings.Repeat("─", segWidth)
		lineStyle := lipgloss.NewStyle().Foreground(mutedColor)
		lineParts = append(lineParts, lineStyle.Render(dash))

		// Label.
		label := ev.Time
		if label == "" {
			label = ev.Title
		}
		var labelStyle lipgloss.Style
		if isCursor {
			labelStyle = lipgloss.NewStyle().
				Width(segWidth).
				Align(lipgloss.Center).
				Background(cursorBg).
				Foreground(inverseFg)
		} else {
			labelStyle = lipgloss.NewStyle().
				Width(segWidth).
				Align(lipgloss.Center).
				Foreground(textColor)
		}
		labelParts = append(labelParts, labelStyle.Render(label))
	}

	row1 := strings.Join(iconParts, "")
	row2 := strings.Join(lineParts, "")
	row3 := strings.Join(labelParts, "")
	return row1 + "\n" + row2 + "\n" + row3
}

func (t *Timeline) statusIcon(status string, g Glyphs) string {
	switch status {
	case "done":
		return g.Check
	case "active":
		return g.Star
	case "pending":
		return g.Dot
	default:
		return g.Dot
	}
}

func (t *Timeline) statusColor(status string, positive, accent, muted lipgloss.Color) lipgloss.Color {
	switch status {
	case "done":
		return positive
	case "active":
		return accent
	default:
		return muted
	}
}

// KeyBindings implements Component.
func (t *Timeline) KeyBindings() []KeyBind {
	if t.opts.Horizontal {
		return []KeyBind{
			{Key: "left/h", Label: "Previous", Group: "TIMELINE"},
			{Key: "right/l", Label: "Next", Group: "TIMELINE"},
			{Key: "enter", Label: "Select", Group: "TIMELINE"},
		}
	}
	return []KeyBind{
		{Key: "up/k", Label: "Previous", Group: "TIMELINE"},
		{Key: "down/j", Label: "Next", Group: "TIMELINE"},
		{Key: "enter", Label: "Select", Group: "TIMELINE"},
	}
}

// SetSize implements Component.
func (t *Timeline) SetSize(w, h int) {
	t.width = w
	t.height = h
	t.clampScroll()
}

// Focused implements Component.
func (t *Timeline) Focused() bool { return t.focused }

// SetFocused implements Component.
func (t *Timeline) SetFocused(f bool) { t.focused = f }

// SetTheme implements Themed.
func (t *Timeline) SetTheme(theme Theme) { t.theme = theme }

func (t *Timeline) clampScroll() {
	if t.height <= 0 {
		return
	}
	visible := t.height / 2
	if visible < 1 {
		visible = 1
	}
	if t.cursor < t.scroll {
		t.scroll = t.cursor
	}
	if t.cursor >= t.scroll+visible {
		t.scroll = t.cursor - visible + 1
	}
	if t.scroll < 0 {
		t.scroll = 0
	}
}
