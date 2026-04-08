package tuikit

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// NotifyMsg triggers a timed notification in the App.
// Components return this via tea.Cmd to show notifications without
// needing a direct reference to the App.
type NotifyMsg struct {
	Text     string
	Duration time.Duration
}

// NotifyCmd returns a tea.Cmd that sends a NotifyMsg.
func NotifyCmd(text string, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return NotifyMsg{Text: text, Duration: duration}
	}
}
