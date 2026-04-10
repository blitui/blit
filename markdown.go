package blit

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

// Markdown renders md as a terminal-formatted string styled using the semantic
// color tokens from theme. Headings use theme.Accent, code block backgrounds
// use theme.Muted, body text uses theme.Text.
//
// The rendered string includes trailing newlines added by glamour; callers may
// trim with strings.TrimSpace if needed.
func Markdown(md string, theme Theme) string {
	cfg := themeToStyleConfig(theme)
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(cfg),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return out
}

// ptr returns a pointer to v, used when building glamour StylePrimitive fields.
func ptr[T any](v T) *T { return &v }

// themeToStyleConfig converts a blit Theme into a glamour ansi.StyleConfig
// by mapping semantic color tokens to glamour's ANSIColor scheme.
func themeToStyleConfig(t Theme) ansi.StyleConfig {
	accent := string(t.Accent)
	muted := string(t.Muted)
	text := string(t.Text)
	positive := string(t.Positive)
	negative := string(t.Negative)

	headingBlock := ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: &accent,
			Bold:  ptr(true),
		},
	}

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &text,
			},
		},
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &text,
			},
		},
		Heading: headingBlock,
		H1:      headingBlock,
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &accent,
				Bold:  ptr(true),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &accent,
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &accent,
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &accent,
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &muted,
			},
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           &muted,
				BackgroundColor: &muted,
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color:           &text,
					BackgroundColor: &muted,
				},
			},
		},
		Text: ansi.StylePrimitive{
			Color: &text,
		},
		Strong: ansi.StylePrimitive{
			Color: &accent,
			Bold:  ptr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  &muted,
			Italic: ptr(true),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  &muted,
				Italic: ptr(true),
			},
		},
		Item: ansi.StylePrimitive{
			Color: &text,
		},
		Enumeration: ansi.StylePrimitive{
			Color: &accent,
		},
		Link: ansi.StylePrimitive{
			Color:     &positive,
			Underline: ptr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: &positive,
		},
		HorizontalRule: ansi.StylePrimitive{
			Color: &muted,
		},
		Strikethrough: ansi.StylePrimitive{
			Color:      &negative,
			CrossedOut: ptr(true),
		},
	}
}

// ansiEscape matches ANSI escape sequences for stripping.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences from s.
func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// highlightSections returns a copy of rendered with lines whose plain-text
// content contains "BREAKING" or "SECURITY" re-colored using the theme's
// Negative and Flash tokens respectively.
func highlightSections(rendered string, theme Theme) string {
	negStyle := lipgloss.NewStyle().Foreground(theme.Negative).Bold(true)
	warnStyle := lipgloss.NewStyle().Foreground(theme.Flash).Bold(true)

	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		plain := strings.TrimSpace(stripANSI(line))
		switch {
		case strings.Contains(plain, "BREAKING"):
			lines[i] = negStyle.Render(plain)
		case strings.Contains(plain, "SECURITY"):
			lines[i] = warnStyle.Render(plain)
		}
	}
	return strings.Join(lines, "\n")
}
