package blit

import (
	"encoding/base64"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// CopyToClipboardCmd returns a tea.Cmd that copies text to the system
// clipboard using the OSC 52 escape sequence. This works over SSH and in
// most modern terminal emulators without requiring external tools.
func CopyToClipboardCmd(text string) tea.Cmd {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return tea.Printf("\033]52;c;%s\a", encoded)
}

// CopyToClipboardMsg is sent after a successful clipboard copy to notify
// the application. Components can listen for this to show confirmation.
type CopyToClipboardMsg struct {
	Text string
}

// CopyToClipboardWithNotify returns a tea.Batch that copies text to the
// clipboard and sends a CopyToClipboardMsg so the app can show feedback.
func CopyToClipboardWithNotify(text string) tea.Cmd {
	return tea.Batch(
		CopyToClipboardCmd(text),
		func() tea.Msg { return CopyToClipboardMsg{Text: text} },
	)
}

// osc52Sequence returns the raw OSC 52 escape sequence for testing.
func osc52Sequence(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\033]52;c;%s\a", encoded)
}
