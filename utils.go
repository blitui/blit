package blit

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

// Hyperlink is an alias for OSC8Link — wraps text in a clickable terminal hyperlink.
// Supported in modern terminals (iTerm2, Kitty, WezTerm, Windows Terminal, etc.).
func Hyperlink(url, text string) string {
	return OSC8Link(url, text)
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
	go func() { _ = cmd.Wait() }()
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

		switch {
		case mono, i == 0, v == data[i-1]:
			sb.WriteString(neutralStyle.Render(ch))
		case v > data[i-1]:
			sb.WriteString(upStyle.Render(ch))
		case v < data[i-1]:
			sb.WriteString(downStyle.Render(ch))
		default:
			sb.WriteString(neutralStyle.Render(ch))
		}
	}
	return sb.String(), len(data)
}

// Divider renders a horizontal line of ─ characters at the given width,
// styled with the theme's Border color.
func Divider(width int, theme Theme) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Border)).
		Render(strings.Repeat("─", width))
}

// Truncate shortens s to maxWidth characters, appending "…" if truncated.
// Uses visual width (lipgloss.Width) so ANSI sequences are handled correctly.
// If s fits within maxWidth, it is returned unchanged.
func Truncate(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Strip to plain text for safe truncation, then re-measure
	plain := strings.TrimRight(s, " ")
	runes := []rune(plain)
	for i := len(runes) - 1; i >= 0; i-- {
		candidate := string(runes[:i]) + "…"
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return "…"
}

// Badge renders a short text label with foreground color and optional bold.
// Useful for category tags in detail bars and status displays.
func Badge(text string, fg lipgloss.Color, bold bool) string {
	s := lipgloss.NewStyle().Foreground(fg)
	if bold {
		s = s.Bold(true)
	}
	return s.Render(text)
}
