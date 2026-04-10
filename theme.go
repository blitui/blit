package blit

import (
	"math"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines semantic color tokens for consistent styling across components.
// Components reference these tokens instead of raw colors.
type Theme struct {
	Positive    lipgloss.Color            // Green: gains, success, online
	Negative    lipgloss.Color            // Red: losses, errors, offline
	Accent      lipgloss.Color            // Highlights, active elements
	Muted       lipgloss.Color            // Dimmed text, secondary info
	Text        lipgloss.Color            // Primary text
	TextInverse lipgloss.Color            // Text on colored backgrounds
	Cursor      lipgloss.Color            // Cursor/selection highlight
	Border      lipgloss.Color            // Borders, separators
	Flash       lipgloss.Color            // Temporary notification background
	Extra       map[string]lipgloss.Color // App-specific color tokens
	Glyphs      *Glyphs                   // Optional glyph override; nil uses DefaultGlyphs
	Borders     *BorderSet                // Optional border override; nil uses DefaultBorders
	registry    *styleRegistry            // Named style overrides; nil means built-ins only
}

// glyphsOrDefault returns the theme glyphs or DefaultGlyphs if nil.
func (t Theme) glyphsOrDefault() Glyphs {
	if t.Glyphs != nil {
		return *t.Glyphs
	}
	return DefaultGlyphs()
}

// Color returns an app-specific color from Extra by key, falling back to
// the provided default if the key does not exist.
func (t Theme) Color(key string, fallback lipgloss.Color) lipgloss.Color {
	if t.Extra != nil {
		if c, ok := t.Extra[key]; ok {
			return c
		}
	}
	return fallback
}

// BorderSet groups named lipgloss border styles for consistent component use.
type BorderSet struct {
	Rounded lipgloss.Border
	Double  lipgloss.Border
	Thick   lipgloss.Border
	Ascii   lipgloss.Border
	Minimal lipgloss.Border
}

// DefaultBorders returns the standard lipgloss border set.
func DefaultBorders() BorderSet {
	return BorderSet{
		Rounded: lipgloss.RoundedBorder(),
		Double:  lipgloss.DoubleBorder(),
		Thick:   lipgloss.ThickBorder(),
		Ascii:   lipgloss.ASCIIBorder(),
		Minimal: lipgloss.NormalBorder(),
	}
}

// DefaultTheme returns a dark theme suitable for most terminal backgrounds.
func DefaultTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#22c55e"),
		Negative:    lipgloss.Color("#ef4444"),
		Accent:      lipgloss.Color("#3b82f6"),
		Muted:       lipgloss.Color("#6b7280"),
		Text:        lipgloss.Color("#e5e7eb"),
		TextInverse: lipgloss.Color("#111827"),
		Cursor:      lipgloss.Color("#38bdf8"),
		Border:      lipgloss.Color("#374151"),
		Flash:       lipgloss.Color("#facc15"),
	}
}

// LightTheme returns a light theme for light terminal backgrounds.
func LightTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#16a34a"),
		Negative:    lipgloss.Color("#dc2626"),
		Accent:      lipgloss.Color("#2563eb"),
		Muted:       lipgloss.Color("#9ca3af"),
		Text:        lipgloss.Color("#111827"),
		TextInverse: lipgloss.Color("#f9fafb"),
		Cursor:      lipgloss.Color("#0284c7"),
		Border:      lipgloss.Color("#d1d5db"),
		Flash:       lipgloss.Color("#eab308"),
	}
}

// ThemeFromMap creates a Theme from a map of color names to hex values.
// Missing keys fall back to DefaultTheme values. Keys not matching a built-in
// token are placed in the Extra map.
func ThemeFromMap(m map[string]string) Theme {
	t := DefaultTheme()
	builtins := map[string]*lipgloss.Color{
		"positive":     &t.Positive,
		"negative":     &t.Negative,
		"accent":       &t.Accent,
		"muted":        &t.Muted,
		"text":         &t.Text,
		"text_inverse": &t.TextInverse,
		"cursor":       &t.Cursor,
		"border":       &t.Border,
		"flash":        &t.Flash,
	}
	for k, v := range m {
		if ptr, ok := builtins[k]; ok {
			*ptr = lipgloss.Color(v)
		} else {
			if t.Extra == nil {
				t.Extra = make(map[string]lipgloss.Color)
			}
			t.Extra[k] = lipgloss.Color(v)
		}
	}
	return t
}

// --- Theme registry ---

var (
	themeMu       sync.RWMutex
	themeRegistry = map[string]Theme{}
)

// Register adds a named theme to the global preset registry.
func Register(name string, t Theme) {
	themeMu.Lock()
	defer themeMu.Unlock()
	themeRegistry[name] = t
}

// Presets returns a copy of the global theme registry.
func Presets() map[string]Theme {
	themeMu.RLock()
	defer themeMu.RUnlock()
	out := make(map[string]Theme, len(themeRegistry))
	for k, v := range themeRegistry {
		out[k] = v
	}
	return out
}

// --- WCAG contrast ---

// Contrast computes the WCAG 2.1 contrast ratio between bg and fg colors.
// Returns a value in [1, 21]; 4.5 satisfies WCAG AA for normal text.
func Contrast(bg, fg lipgloss.Color) float64 {
	l1 := relativeLuminance(string(bg))
	l2 := relativeLuminance(string(fg))
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(hex string) float64 {
	r, g, b := parseHex(hex)
	return 0.2126*linearise(float64(r)/255) +
		0.7152*linearise(float64(g)/255) +
		0.0722*linearise(float64(b)/255)
}

func linearise(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// --- Runtime theme switching ---

// SetThemeMsg is a Bubble Tea message that asks the App to swap its theme.
type SetThemeMsg struct {
	Theme Theme
}

// SetThemeCmd returns a tea.Cmd that emits a SetThemeMsg.
func SetThemeCmd(t Theme) func() tea.Msg {
	return func() tea.Msg { return SetThemeMsg{Theme: t} }
}
