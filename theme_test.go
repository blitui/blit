package tuikit

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
