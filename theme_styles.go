package blit

import "github.com/charmbracelet/lipgloss"

// ThemeStyles returns a curated set of commonly-needed styles derived from
// the theme's semantic color tokens. Call this once on component init and
// again whenever the theme changes (in SetTheme).
//
// The returned Styles struct covers the most common rendering patterns:
// text variants, semantic colors, structural styles, interactive states,
// and badge presets. Consumers can extend it with domain-specific styles
// using the Extra map on the Theme.
func ThemeStyles(t Theme) Styles {
	return Styles{
		// Text variants
		Text:       NewStyle().Foreground(t.Text),
		TextBold:   NewStyle().Foreground(t.Text).Bold(true),
		Muted:      NewStyle().Foreground(t.Muted),
		Accent:     NewStyle().Foreground(t.Accent),
		AccentBold: NewStyle().Foreground(t.Accent).Bold(true),

		// Semantic
		Positive: NewStyle().Foreground(t.Positive),
		Negative: NewStyle().Foreground(t.Negative),
		Warn:     NewStyle().Foreground(t.Warn),
		Info:     NewStyle().Foreground(t.SemanticColor("info", t.Accent)),

		// Structural
		Title:    NewStyle().Foreground(t.Text).Bold(true).PaddingLeft(1),
		Subtitle: NewStyle().Foreground(t.Muted).PaddingLeft(1),
		Label:    NewStyle().Foreground(t.Muted),
		Detail:   NewStyle().Foreground(t.Muted),
		Hint:     NewStyle().Foreground(t.Muted),

		// Interactive
		Cursor:   NewStyle().Background(lipgloss.Color(t.Cursor)).Foreground(lipgloss.Color(t.TextInverse)),
		Selected: NewStyle().Background(lipgloss.Color(t.Accent)).Foreground(lipgloss.Color(t.TextInverse)),
		Flash:    NewStyle().Background(lipgloss.Color(t.Flash)),
		Border:   NewStyle().Foreground(lipgloss.Color(t.Border)),
		Header:   NewStyle().Foreground(t.Accent).Bold(true),

		// Badge presets (foreground-only, matching Badge() utility)
		BadgePositive: NewStyle().Foreground(t.Positive),
		BadgeNegative: NewStyle().Foreground(t.Negative),
		BadgeWarn:     NewStyle().Foreground(t.Warn),
		BadgeAccent:   NewStyle().Foreground(t.Accent),
		BadgeMuted:    NewStyle().Foreground(t.Muted),
	}
}

// Styles holds a curated set of pre-built styles derived from a Theme.
// Consumers should call ThemeStyles(theme) to construct one and replace it
// on SetTheme.
type Styles struct {
	// Text variants
	Text       Style // primary text
	TextBold   Style // bold primary text
	Muted      Style // secondary/dimmed text
	Accent     Style // highlighted/active text
	AccentBold Style // bold accent text

	// Semantic
	Positive Style // green: success, online, gains
	Negative Style // red: errors, offline, losses
	Warn     Style // yellow: warnings, caution
	Info     Style // blue/cyan: informational

	// Structural
	Title    Style // bold text, left-padded
	Subtitle Style // muted text, left-padded
	Label    Style // muted label prefix (e.g. "Type:    ")
	Detail   Style // detail content
	Hint     Style // keyboard hints, dimmed

	// Interactive
	Cursor   Style // cursor highlight row (bg)
	Selected Style // selected row (bg)
	Flash    Style // flash background row (bg)
	Border   Style // border foreground
	Header   Style // section header, accent+bold

	// Badge presets (foreground-only)
	BadgePositive Style // positive badge
	BadgeNegative Style // negative badge
	BadgeWarn     Style // warn badge
	BadgeAccent   Style // accent badge
	BadgeMuted    Style // muted badge
}

// SemanticColor resolves a well-known semantic color name via the Theme's
// Extra map, falling back to a sensible default from the base tokens when
// the name is not present. This lets consumers define domain-specific
// colors (like "info", "create", "local") in Theme.Extra and have them
// resolve automatically, while still working with themes that don't
// define them.
//
// Well-known names and their defaults:
//
//	info    → Accent
//	create  → Positive
//	delete  → Negative
//	review  → Cursor
//	comment → Muted
//	issue   → Warn
//	release → Flash
//	local   → Accent
func (t Theme) SemanticColor(name string, fallback Color) Color {
	if t.Extra != nil {
		if c, ok := t.Extra[name]; ok {
			return c
		}
	}
	// Well-known defaults — only return these when the caller doesn't
	// supply a meaningful fallback. If the caller provides a non-zero
	// fallback, respect it.
	defaults := map[string]Color{
		"info":    t.Accent,
		"create":  t.Positive,
		"delete":  t.Negative,
		"review":  t.Cursor,
		"comment": t.Muted,
		"issue":   t.Warn,
		"release": t.Flash,
		"local":   t.Accent,
	}
	if c, ok := defaults[name]; ok {
		return c
	}
	return fallback
}

// HealthDot renders a coloured dot indicator for a status line.
// ok=true shows a positive (green) dot, ok=false shows a negative (red) dot,
// and unknown shows a muted dot. The label is rendered in Muted after the dot.
func HealthDot(label string, ok bool, theme Theme) string {
	glyphs := theme.glyphsOrDefault()
	switch {
	case ok:
		return Badge(glyphs.SelectedBullet, theme.Positive, false) + " " + NewStyle().Foreground(theme.Muted).Render(label)
	default:
		return Badge(glyphs.SelectedBullet, theme.Negative, false) + " " + NewStyle().Foreground(theme.Muted).Render(label)
	}
}

// HealthDotUnknown renders a muted empty dot for an unknown/unchecked status.
func HealthDotUnknown(label string, theme Theme) string {
	glyphs := theme.glyphsOrDefault()
	return Badge(glyphs.UnselectedBullet, theme.Muted, false) + " " + NewStyle().Foreground(theme.Muted).Render(label)
}
