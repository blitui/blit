package blit

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme.Positive == "" {
		t.Error("DefaultTheme().Positive should not be empty")
	}
	if theme.Negative == "" {
		t.Error("DefaultTheme().Negative should not be empty")
	}
	if theme.Text == "" {
		t.Error("DefaultTheme().Text should not be empty")
	}
}

func TestLightTheme(t *testing.T) {
	theme := LightTheme()
	if theme.Positive == "" {
		t.Error("LightTheme().Positive should not be empty")
	}
	dark := DefaultTheme()
	if theme.Text == dark.Text {
		t.Error("LightTheme and DefaultTheme should have different Text colors")
	}
}

func TestThemeFromMap(t *testing.T) {
	m := map[string]string{
		"positive": "#00ff00",
		"negative": "#ff0000",
		"accent":   "#0000ff",
	}
	theme := ThemeFromMap(m)
	if theme.Positive != lipgloss.Color("#00ff00") {
		t.Errorf("expected #00ff00, got %v", theme.Positive)
	}
	if theme.Negative != lipgloss.Color("#ff0000") {
		t.Errorf("expected #ff0000, got %v", theme.Negative)
	}
	if theme.Accent != lipgloss.Color("#0000ff") {
		t.Errorf("expected #0000ff, got %v", theme.Accent)
	}
}

func TestThemeFromMapDefaults(t *testing.T) {
	theme := ThemeFromMap(map[string]string{})
	defaults := DefaultTheme()
	if theme.Positive != defaults.Positive {
		t.Errorf("expected default Positive %v, got %v", defaults.Positive, theme.Positive)
	}
}

func TestThemeFromMapExtra(t *testing.T) {
	m := map[string]string{
		"positive": "#00ff00",
		"push":     "#22c55e",
		"pr":       "#3b82f6",
		"review":   "#a855f7",
	}
	theme := ThemeFromMap(m)
	if theme.Positive != lipgloss.Color("#00ff00") {
		t.Errorf("expected #00ff00, got %v", theme.Positive)
	}
	if theme.Extra == nil {
		t.Fatal("Extra should not be nil when unknown keys are provided")
	}
	if theme.Extra["push"] != lipgloss.Color("#22c55e") {
		t.Errorf("Extra[push] = %v, want #22c55e", theme.Extra["push"])
	}
	if theme.Extra["pr"] != lipgloss.Color("#3b82f6") {
		t.Errorf("Extra[pr] = %v, want #3b82f6", theme.Extra["pr"])
	}
	if theme.Extra["review"] != lipgloss.Color("#a855f7") {
		t.Errorf("Extra[review] = %v, want #a855f7", theme.Extra["review"])
	}
}

func TestThemeColor(t *testing.T) {
	theme := DefaultTheme()
	theme.Extra = map[string]lipgloss.Color{
		"star": lipgloss.Color("#ffaa00"),
	}

	// Existing key
	got := theme.Color("star", lipgloss.Color("#000000"))
	if got != lipgloss.Color("#ffaa00") {
		t.Errorf("Color(star) = %v, want #ffaa00", got)
	}

	// Missing key falls back
	got = theme.Color("missing", lipgloss.Color("#111111"))
	if got != lipgloss.Color("#111111") {
		t.Errorf("Color(missing) = %v, want #111111 (fallback)", got)
	}
}

func TestThemeColorNilExtra(t *testing.T) {
	theme := DefaultTheme()
	// Extra is nil by default
	got := theme.Color("anything", lipgloss.Color("#aabbcc"))
	if got != lipgloss.Color("#aabbcc") {
		t.Errorf("Color with nil Extra should return fallback, got %v", got)
	}
}

func TestThemeFromMapNoExtraForBuiltins(t *testing.T) {
	m := map[string]string{
		"positive": "#00ff00",
		"text":     "#ffffff",
	}
	theme := ThemeFromMap(m)
	// Built-in keys should NOT appear in Extra
	if len(theme.Extra) > 0 {
		t.Errorf("Extra should be empty for built-in keys only, got %v", theme.Extra)
	}
}
