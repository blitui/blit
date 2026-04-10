package blit

import (
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotifyMsg triggers a timed notification in the App.
type NotifyMsg struct {
	Text     string
	Duration time.Duration
}

// NotifyCmd returns a tea.Cmd that sends a NotifyMsg.
func NotifyCmd(text string, duration time.Duration) tea.Cmd {
	return func() tea.Msg { return NotifyMsg{Text: text, Duration: duration} }
}

// ToastSeverity classifies a toast notification.
type ToastSeverity int

const (
	SeverityInfo ToastSeverity = iota
	SeveritySuccess
	SeverityWarn
	SeverityError
)

// ToastAction is a labelled action button on a toast.
type ToastAction struct {
	Label   string
	Handler func()
}

// ToastMsg triggers a toast notification via the App ToastManager.
type ToastMsg struct {
	Severity ToastSeverity
	Title    string
	Body     string
	Duration time.Duration
	Actions  []ToastAction
}

// ToastCmd returns a tea.Cmd that sends a ToastMsg.
func ToastCmd(severity ToastSeverity, title, body string, duration time.Duration, actions ...ToastAction) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{Severity: severity, Title: title, Body: body, Duration: duration, Actions: actions}
	}
}

type dismissTopToastMsg struct{}
type dismissToastMsg struct{ index int }

// ToastManagerOpts configures the ToastManager.
type ToastManagerOpts struct {
	MaxVisible   int
	AnimDuration time.Duration
}

type toastEntry struct {
	Severity    ToastSeverity
	Title       string
	Body        string
	Duration    time.Duration
	Actions     []ToastAction
	expiry      time.Time
	hovered     bool
	slideOffset int
}

type toastManager struct {
	toasts []*toastEntry
	opts   ToastManagerOpts
	theme  Theme
	noAnim bool
}

func newToastManager(opts ToastManagerOpts) *toastManager {
	if opts.MaxVisible <= 0 {
		opts.MaxVisible = 5
	}
	if opts.AnimDuration <= 0 {
		opts.AnimDuration = 300 * time.Millisecond
	}
	return &toastManager{opts: opts, theme: DefaultTheme(), noAnim: os.Getenv("BLIT_NO_ANIM") == "1"}
}

func (tm *toastManager) add(msg ToastMsg) {
	dur := msg.Duration
	if dur <= 0 {
		dur = 4 * time.Second
	}
	e := &toastEntry{
		Severity: msg.Severity, Title: msg.Title, Body: msg.Body,
		Duration: dur, Actions: msg.Actions, expiry: time.Now().Add(dur),
	}
	if !tm.noAnim {
		e.slideOffset = toastPanelWidth
	}
	tm.toasts = append(tm.toasts, e)
	if len(tm.toasts) > tm.opts.MaxVisible {
		tm.toasts = tm.toasts[len(tm.toasts)-tm.opts.MaxVisible:]
	}
}

func (tm *toastManager) dismissTop() {
	if len(tm.toasts) > 0 {
		tm.toasts = tm.toasts[:len(tm.toasts)-1]
	}
}

func (tm *toastManager) dismissAt(idx int) {
	if idx < 0 || idx >= len(tm.toasts) {
		return
	}
	tm.toasts = append(tm.toasts[:idx], tm.toasts[idx+1:]...)
}

func (tm *toastManager) tick(now time.Time) bool {
	changed := false
	i := 0
	for i < len(tm.toasts) {
		e := tm.toasts[i]
		if !e.hovered && now.After(e.expiry) {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			changed = true
			continue
		}
		if !tm.noAnim && e.slideOffset > 0 {
			step := e.slideOffset / 3
			if step < 1 {
				step = 1
			}
			e.slideOffset -= step
			if e.slideOffset < 0 {
				e.slideOffset = 0
			}
			changed = true
		}
		i++
	}
	return changed
}

func (tm *toastManager) hasActive() bool { return len(tm.toasts) > 0 }

func severityIcon(s ToastSeverity) string {
	switch s {
	case SeveritySuccess:
		return "✓"
	case SeverityWarn:
		return "⚠"
	case SeverityError:
		return "✗"
	default:
		return "i"
	}
}

func (tm *toastManager) severityColor(s ToastSeverity) lipgloss.Color {
	switch s {
	case SeveritySuccess:
		return tm.theme.Positive
	case SeverityWarn:
		return tm.theme.Flash
	case SeverityError:
		return tm.theme.Negative
	default:
		return tm.theme.Accent
	}
}

const toastPanelWidth = 40

func (tm *toastManager) view(screenWidth int) string {
	if len(tm.toasts) == 0 {
		return ""
	}
	var panels []string
	for _, e := range tm.toasts {
		panels = append(panels, tm.renderEntry(e, screenWidth))
	}
	return strings.Join(panels, "\n")
}

func (tm *toastManager) renderEntry(e *toastEntry, screenWidth int) string {
	color := tm.severityColor(e.Severity)
	accentBar := lipgloss.NewStyle().Foreground(color).Render("┃ ")
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(tm.theme.Text)).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(tm.theme.Muted))
	iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	innerWidth := toastPanelWidth - 6
	header := iconStyle.Render(severityIcon(e.Severity)) + " " + titleStyle.Render(truncate(e.Title, innerWidth))
	var lines []string
	lines = append(lines, accentBar+header)
	if e.Body != "" {
		for _, l := range wrapText(e.Body, innerWidth) {
			lines = append(lines, accentBar+bodyStyle.Render(l))
		}
	}
	for _, a := range e.Actions {
		lines = append(lines, accentBar+lipgloss.NewStyle().Foreground(color).Underline(true).Render("["+a.Label+"]"))
	}
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(tm.theme.Border)).
		Background(lipgloss.Color(tm.theme.TextInverse)).
		Padding(0, 1).Width(toastPanelWidth).
		Render(strings.Join(lines, "\n"))
	if !tm.noAnim && e.slideOffset > 0 {
		pad := strings.Repeat(" ", e.slideOffset)
		pl := strings.Split(panel, "\n")
		for i, l := range pl {
			pl[i] = pad + l
		}
		return strings.Join(pl, "\n")
	}
	if screenWidth > toastPanelWidth+2 {
		pad := strings.Repeat(" ", screenWidth-toastPanelWidth-2)
		pl := strings.Split(panel, "\n")
		for i, l := range pl {
			pl[i] = pad + l
		}
		return strings.Join(pl, "\n")
	}
	return panel
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	var lines []string
	cur := ""
	for _, w := range words {
		switch {
		case cur == "":
			cur = w
		case len(cur)+1+len(w) <= width:
			cur += " " + w
		default:
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
