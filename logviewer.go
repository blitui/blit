package blit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents the severity of a log line.
type LogLevel int

const (
	// LogDebug is the lowest severity level.
	LogDebug LogLevel = iota
	// LogInfo is the informational severity level.
	LogInfo
	// LogWarn is the warning severity level.
	LogWarn
	// LogError is the highest severity level.
	LogError
)

// LogLine is a single entry in the LogViewer.
type LogLine struct {
	Level     LogLevel
	Timestamp time.Time
	Message   string
	Source    string
}

// LogViewer is an append-only, auto-scrolling log display with level filtering,
// substring search, pause/resume, and goroutine-safe Append.
//
// It implements Component and Themed.
type LogViewer struct {
	mu sync.RWMutex

	allLines      []LogLine
	filteredLines []LogLine

	viewport viewport.Model
	ready    bool
	width    int
	height   int
	focused  bool
	theme    Theme

	// auto-scroll state
	paused       bool
	userScrolled bool

	// filter state
	filterText  string
	filterLevel LogLevel // minimum level shown; cycles debug→info→warn→error→debug
	filterInput textinput.Model
	filtering   bool // slash-filter input mode active
}

// NewLogViewer creates a new LogViewer with default settings.
func NewLogViewer() *LogViewer {
	ti := textinput.New()
	ti.Placeholder = "filter…"
	ti.CharLimit = 120

	return &LogViewer{
		filterLevel: LogDebug,
		filterInput: ti,
	}
}

// Append adds a LogLine to the viewer. Safe to call from any goroutine.
func (lv *LogViewer) Append(line LogLine) {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.allLines = append(lv.allLines, line)
	lv.rebuildFilteredLocked()
}

// Clear removes all log lines.
func (lv *LogViewer) Clear() {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.allLines = lv.allLines[:0]
	lv.rebuildFilteredLocked()
}

// Lines returns a snapshot of all lines currently stored.
func (lv *LogViewer) Lines() []LogLine {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	out := make([]LogLine, len(lv.allLines))
	copy(out, lv.allLines)
	return out
}

// --- Component interface ---

// Init implements Component.
func (lv *LogViewer) Init() tea.Cmd { return nil }

// Update implements Component.
func (lv *LogViewer) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := lv.handleKey(msg)
		return lv, cmd

	case tea.MouseMsg:
		cmd := lv.handleMouse(msg)
		return lv, cmd

	case LogAppendMsg:
		lv.Append(msg.Line)
		return lv, nil
	}
	return lv, nil
}

func (lv *LogViewer) handleMouse(msg tea.MouseMsg) tea.Cmd {
	// Ignore events outside component bounds.
	if msg.X < 0 || msg.X >= lv.width || msg.Y < 0 || msg.Y >= lv.height {
		return nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		lv.mu.Lock()
		lv.viewport.ScrollUp(3)
		lv.userScrolled = true
		lv.paused = true
		lv.mu.Unlock()
		return Consumed()
	case tea.MouseButtonWheelDown:
		lv.mu.Lock()
		lv.viewport.ScrollDown(3)
		if lv.viewport.AtBottom() {
			lv.userScrolled = false
		}
		lv.mu.Unlock()
		return Consumed()
	}
	return nil
}

// LogAppendMsg is a Bubble Tea message that appends a line to the LogViewer.
// Use LogAppendCmd to produce this from a tea.Cmd.
type LogAppendMsg struct{ Line LogLine }

// LogAppendCmd returns a tea.Cmd that appends a LogLine to any LogViewer
// that receives it via Update.
func LogAppendCmd(line LogLine) tea.Cmd {
	return func() tea.Msg { return LogAppendMsg{Line: line} }
}

// View implements Component.
func (lv *LogViewer) View() string {
	if !lv.ready {
		return ""
	}

	// Snapshot viewport under read-lock to avoid racing with Append().
	lv.mu.RLock()
	vpView := strings.TrimRight(lv.viewport.View(), "\n")
	lv.mu.RUnlock()

	var sections []string

	// Status bar: pause indicator + level filter + search text
	sections = append(sections, lv.renderStatusBar())

	// Viewport
	sections = append(sections, vpView)

	// Filter input (shown when in filtering mode)
	if lv.filtering {
		sections = append(sections, lv.renderFilterInput())
	}

	return lipgloss.NewStyle().MaxWidth(lv.width).
		Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// KeyBindings implements Component.
func (lv *LogViewer) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/k", Label: "Scroll up", Group: "LOG"},
		{Key: "down/j", Label: "Scroll down", Group: "LOG"},
		{Key: "end", Label: "Jump to latest", Group: "LOG"},
		{Key: "p", Label: "Pause/resume auto-scroll", Group: "LOG"},
		{Key: "c", Label: "Clear log", Group: "LOG"},
		{Key: "/", Label: "Filter by substring", Group: "LOG"},
		{Key: "l", Label: "Cycle level filter", Group: "LOG"},
	}
}

// SetSize implements Component.
func (lv *LogViewer) SetSize(w, h int) {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.width = w
	lv.height = h

	vpHeight := lv.viewportHeight()
	if !lv.ready {
		lv.viewport = viewport.New(w, vpHeight)
		lv.ready = true
	} else {
		lv.viewport.Width = w
		lv.viewport.Height = vpHeight
	}
	lv.rebuildContent()
}

// Focused implements Component.
func (lv *LogViewer) Focused() bool { return lv.focused }

// SetFocused implements Component.
func (lv *LogViewer) SetFocused(f bool) { lv.focused = f }

// SetTheme implements Themed.
func (lv *LogViewer) SetTheme(th Theme) { lv.theme = th }

// --- internal ---

func (lv *LogViewer) handleKey(msg tea.KeyMsg) tea.Cmd {
	// If filter input is active, feed keys to it
	if lv.filtering {
		switch msg.String() {
		case "enter", "esc":
			lv.filtering = false
			lv.filterInput.Blur()
			lv.filterText = lv.filterInput.Value()
			lv.rebuildFiltered()
			return Consumed()
		default:
			var cmd tea.Cmd
			lv.filterInput, cmd = lv.filterInput.Update(msg)
			lv.filterText = lv.filterInput.Value()
			lv.rebuildFiltered()
			return cmd
		}
	}

	switch msg.String() {
	case "up", "k":
		lv.viewport.ScrollUp(1)
		lv.userScrolled = true
		lv.paused = true
		return Consumed()
	case "down", "j":
		lv.viewport.ScrollDown(1)
		if lv.viewport.AtBottom() {
			lv.userScrolled = false
		}
		return Consumed()
	case "pgup":
		lv.viewport.HalfPageUp()
		lv.userScrolled = true
		lv.paused = true
		return Consumed()
	case "pgdown":
		lv.viewport.HalfPageDown()
		if lv.viewport.AtBottom() {
			lv.userScrolled = false
		}
		return Consumed()
	case "end":
		lv.viewport.GotoBottom()
		lv.userScrolled = false
		lv.paused = false
		return Consumed()
	case "p":
		lv.paused = !lv.paused
		if !lv.paused {
			lv.viewport.GotoBottom()
			lv.userScrolled = false
		}
		return Consumed()
	case "c":
		lv.Clear()
		return Consumed()
	case "/":
		lv.filtering = true
		lv.filterInput.SetValue(lv.filterText)
		lv.filterInput.Focus()
		lv.filterInput.CursorEnd()
		return Consumed()
	case "l":
		lv.cycleLevel()
		lv.rebuildFiltered()
		return Consumed()
	}
	return nil
}

func (lv *LogViewer) cycleLevel() {
	switch lv.filterLevel {
	case LogDebug:
		lv.filterLevel = LogInfo
	case LogInfo:
		lv.filterLevel = LogWarn
	case LogWarn:
		lv.filterLevel = LogError
	case LogError:
		lv.filterLevel = LogDebug
	}
}

func (lv *LogViewer) rebuildFiltered() {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.rebuildFilteredLocked()
}

// rebuildFilteredLocked rebuilds the filtered line list and content. Caller
// must hold lv.mu.
func (lv *LogViewer) rebuildFilteredLocked() {
	text := strings.ToLower(lv.filterText)
	filtered := lv.allLines[:0:0]
	for _, line := range lv.allLines {
		if line.Level < lv.filterLevel {
			continue
		}
		if text != "" {
			if !strings.Contains(strings.ToLower(line.Message), text) &&
				!strings.Contains(strings.ToLower(line.Source), text) {
				continue
			}
		}
		filtered = append(filtered, line)
	}
	lv.filteredLines = filtered
	lv.rebuildContent()
}

func (lv *LogViewer) rebuildContent() {
	if !lv.ready {
		return
	}

	var sb strings.Builder
	for i, line := range lv.filteredLines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(lv.renderLine(line))
	}
	lv.viewport.SetContent(sb.String())

	// Sync height
	vpHeight := lv.viewportHeight()
	if lv.viewport.Height != vpHeight {
		lv.viewport.Height = vpHeight
	}

	// Auto-scroll to bottom unless paused or user scrolled up
	if !lv.paused && !lv.userScrolled {
		lv.viewport.GotoBottom()
	}
}

func (lv *LogViewer) renderLine(line LogLine) string {
	chip := lv.levelChip(line.Level)
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color(lv.theme.Muted)).
		Render(line.Timestamp.Format("15:04:05.000"))
	msg := lipgloss.NewStyle().
		Foreground(lipgloss.Color(lv.theme.Text)).
		Render(line.Message)

	parts := []string{ts, chip}
	if line.Source != "" {
		src := lipgloss.NewStyle().
			Foreground(lipgloss.Color(lv.theme.Accent)).
			Render(line.Source)
		parts = append(parts, src)
	}
	parts = append(parts, msg)
	return strings.Join(parts, " ")
}

func (lv *LogViewer) levelChip(level LogLevel) string {
	var text string
	var color lipgloss.Color

	switch level {
	case LogDebug:
		text = "DBG"
		color = lipgloss.Color(lv.theme.Muted)
	case LogInfo:
		text = "INF"
		color = lipgloss.Color(lv.theme.Accent)
	case LogWarn:
		text = "WRN"
		color = lipgloss.Color(lv.theme.Flash)
	case LogError:
		text = "ERR"
		color = lipgloss.Color(lv.theme.Negative)
	default:
		text = "???"
		color = lipgloss.Color(lv.theme.Muted)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(lv.theme.TextInverse)).
		Background(color).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

func (lv *LogViewer) renderStatusBar() string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(lv.theme.Muted))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(lv.theme.Accent))

	var parts []string

	if lv.paused {
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.Color(lv.theme.Flash)).
			Bold(true).
			Render("⏸ PAUSED"))
	} else {
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.Color(lv.theme.Positive)).
			Render("▶ LIVE"))
	}

	levelLabel := lv.levelName(lv.filterLevel)
	parts = append(parts, muted.Render("level:")+accent.Render(levelLabel))

	if lv.filterText != "" {
		parts = append(parts, muted.Render("filter:")+accent.Render(lv.filterText))
	}

	lv.mu.RLock()
	total := len(lv.filteredLines)
	all := len(lv.allLines)
	lv.mu.RUnlock()
	parts = append(parts, muted.Render(fmt.Sprintf("%d/%d lines", total, all)))

	return strings.Join(parts, muted.Render("  ·  "))
}

func (lv *LogViewer) renderFilterInput() string {
	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color(lv.theme.Accent)).
		Bold(true).
		Render("/")
	return prompt + lv.filterInput.View()
}

func (lv *LogViewer) levelName(level LogLevel) string {
	switch level {
	case LogDebug:
		return "debug+"
	case LogInfo:
		return "info+"
	case LogWarn:
		return "warn+"
	case LogError:
		return "error"
	default:
		return "all"
	}
}

func (lv *LogViewer) viewportHeight() int {
	h := lv.height
	h-- // status bar
	if lv.filtering {
		h-- // filter input line
	}
	if h < 1 {
		h = 1
	}
	return h
}
