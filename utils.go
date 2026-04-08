package tuikit

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RelativeTime returns a short human-readable duration string like "3m ago" or "2h ago".
// If the duration is negative (future time), it returns "now".
func RelativeTime(t time.Time, now time.Time) string {
	d := now.Sub(t)
	if d < 0 {
		return "now"
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// OSC8Link wraps text in an OSC8 terminal hyperlink escape sequence.
// If url is empty, the plain text is returned unchanged.
func OSC8Link(url, text string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// OpenURL opens a URL in the user's default browser.
// It runs the command asynchronously and does not block.
func OpenURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return
	}
	go cmd.Wait()
}

// SparklineOpts configures sparkline rendering.
type SparklineOpts struct {
	// UpStyle styles bars that are higher than the previous bar.
	// If zero value, defaults to green (#00ff88).
	UpStyle lipgloss.Style
	// DownStyle styles bars that are lower than the previous bar.
	// If zero value, defaults to red (#ff4444).
	DownStyle lipgloss.Style
	// NeutralStyle styles bars with no change from the previous bar.
	// If zero value, defaults to dim gray (#666666).
	NeutralStyle lipgloss.Style
	// Mono renders all bars in a single style (NeutralStyle) instead of
	// coloring by direction. Useful for detail views.
	Mono bool
}

// Sparkline renders a Unicode block sparkline from a slice of float64 values.
// The output is trimmed or sampled to fit within maxWidth characters.
// Each bar uses one of ▁▂▃▄▅▆▇█ based on the value's position in the min-max range.
// Returns the styled string and the number of characters rendered.
func Sparkline(data []float64, maxWidth int, opts *SparklineOpts) (string, int) {
	if len(data) < 2 || maxWidth < 1 {
		return "", 0
	}

	// Sample if data exceeds width
	if len(data) > maxWidth {
		step := float64(len(data)) / float64(maxWidth)
		sampled := make([]float64, maxWidth)
		for i := range sampled {
			sampled[i] = data[int(float64(i)*step)]
		}
		data = sampled
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	mn, mx := data[0], data[0]
	for _, v := range data {
		mn = math.Min(mn, v)
		mx = math.Max(mx, v)
	}
	spread := mx - mn
	if spread == 0 {
		spread = 1
	}

	upStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88"))
	downStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444"))
	neutralStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	mono := false

	if opts != nil {
		if opts.UpStyle.Value() != "" {
			upStyle = opts.UpStyle
		}
		if opts.DownStyle.Value() != "" {
			downStyle = opts.DownStyle
		}
		if opts.NeutralStyle.Value() != "" {
			neutralStyle = opts.NeutralStyle
		}
		mono = opts.Mono
	}

	var sb strings.Builder
	for i, v := range data {
		idx := int((v - mn) / spread * float64(len(blocks)-1))
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		ch := string(blocks[idx])

		if mono {
			sb.WriteString(neutralStyle.Render(ch))
		} else if i == 0 {
			sb.WriteString(neutralStyle.Render(ch))
		} else if v > data[i-1] {
			sb.WriteString(upStyle.Render(ch))
		} else if v < data[i-1] {
			sb.WriteString(downStyle.Render(ch))
		} else {
			sb.WriteString(neutralStyle.Render(ch))
		}
	}
	return sb.String(), len(data)
}
