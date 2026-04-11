package blit

import (
	"fmt"
	"strings"
)

// ComponentHeader renders a structured header with title, subtitle, and
// status line parts. This is a common pattern in TUI components that
// display a title bar with health indicators, filter state, or sparklines.
//
// Usage:
//
//	h := ComponentHeader{
//	    Title:    "myapp",
//	    Subtitle: "Watching: owner/repo-a, owner/repo-b",
//	    StatusParts: []HeaderStatusPart{
//	        {Text: "Poll 5s ago"},
//	        {Text: "● repo-a", Color: theme.Positive},
//	        {Text: "API 4500/5000", Color: theme.Warn},
//	    },
//	}
//	header := h.Render(width, theme)
type ComponentHeader struct {
	Title       string
	Subtitle    string
	StatusParts []HeaderStatusPart
}

// HeaderStatusPart is a single segment in the status line.
type HeaderStatusPart struct {
	Text  string
	Color Color // zero value = use Muted text
}

// Render produces a multi-line header string: title, subtitle, status.
func (h ComponentHeader) Render(width int, theme Theme) string {
	s := ThemeStyles(theme)

	title := s.Title.Render(h.Title)
	subtitle := s.Subtitle.Render(h.Subtitle)

	var statusParts []string
	for _, p := range h.StatusParts {
		if p.Color != "" {
			statusParts = append(statusParts, NewStyle().Foreground(p.Color).Render(p.Text))
		} else {
			statusParts = append(statusParts, s.Muted.Render(p.Text))
		}
	}
	status := s.Subtitle.Render(strings.Join(statusParts, "  "))

	return JoinVertical(PosLeft, title, subtitle, status)
}

// StatusPaused renders a "[PAUSED]" indicator in the warn color.
func StatusPaused(theme Theme) HeaderStatusPart {
	return HeaderStatusPart{
		Text:  NewStyle().Foreground(theme.Warn).Render("[PAUSED]"),
		Color: theme.Warn,
	}
}

// StatusPollAgo renders a "Poll X ago" indicator in muted color.
func StatusPollAgo(ago string) HeaderStatusPart {
	return HeaderStatusPart{Text: "Poll " + ago + " ago"}
}

// StatusRateLimit renders a rate limit indicator with color based on
// remaining percentage.
func StatusRateLimit(remaining, limit int, theme Theme) HeaderStatusPart {
	pct := float64(remaining) / float64(limit) * 100
	color := theme.Positive
	if pct < 20 {
		color = theme.Negative
	} else if pct < 50 {
		color = theme.Warn
	}
	return HeaderStatusPart{
		Text:  fmt.Sprintf("API %d/%d", remaining, limit),
		Color: color,
	}
}
