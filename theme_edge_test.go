package blit

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestThemeFromMapAllBuiltinKeys verifies that every built-in key is set.
func TestThemeFromMapAllBuiltinKeys(t *testing.T) {
	m := map[string]string{
		"positive":     "#111111",
		"negative":     "#222222",
		"accent":       "#333333",
		"muted":        "#444444",
		"text":         "#555555",
		"text_inverse": "#666666",
		"cursor":       "#777777",
		"border":       "#888888",
		"flash":        "#999999",
		"warn":         "#aaaaaa",
	}
	theme := ThemeFromMap(m)

	checks := []struct {
		name string
		got  lipgloss.Color
		want lipgloss.Color
	}{
		{"Positive", theme.Positive, lipgloss.Color("#111111")},
		{"Negative", theme.Negative, lipgloss.Color("#222222")},
		{"Accent", theme.Accent, lipgloss.Color("#333333")},
		{"Muted", theme.Muted, lipgloss.Color("#444444")},
		{"Text", theme.Text, lipgloss.Color("#555555")},
		{"TextInverse", theme.TextInverse, lipgloss.Color("#666666")},
		{"Cursor", theme.Cursor, lipgloss.Color("#777777")},
		{"Border", theme.Border, lipgloss.Color("#888888")},
		{"Flash", theme.Flash, lipgloss.Color("#999999")},
		{"Warn", theme.Warn, lipgloss.Color("#aaaaaa")},
	}
	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
			}
		})
	}

	if len(theme.Extra) > 0 {
		t.Errorf("Extra should be empty when only built-in keys given, got %v", theme.Extra)
	}
}

// TestThemeFromMapMixedBuiltinAndExtra verifies built-in + extra keys
// go to the right places.
func TestThemeFromMapMixedBuiltinAndExtra(t *testing.T) {
	m := map[string]string{
		"accent": "#abcdef",
		"custom": "#fedcba",
	}
	theme := ThemeFromMap(m)

	if theme.Accent != lipgloss.Color("#abcdef") {
		t.Errorf("Accent = %v, want #abcdef", theme.Accent)
	}
	if theme.Extra == nil || theme.Extra["custom"] != lipgloss.Color("#fedcba") {
		t.Errorf("Extra[custom] = %v, want #fedcba", theme.Extra["custom"])
	}
	// Unset built-in should default
	defaults := DefaultTheme()
	if theme.Positive != defaults.Positive {
		t.Errorf("Positive should default, got %v", theme.Positive)
	}
}

// TestThemeRegisterPresetsRoundtrip verifies Register -> Presets returns
// the registered theme.
func TestThemeRegisterPresetsRoundtrip(t *testing.T) {
	custom := Theme{
		Positive: lipgloss.Color("#010101"),
		Negative: lipgloss.Color("#020202"),
		Text:     lipgloss.Color("#030303"),
	}
	Register("test-roundtrip", custom)
	defer func() {
		// Cleanup
		themeMu.Lock()
		delete(themeRegistry, "test-roundtrip")
		themeMu.Unlock()
	}()

	presets := Presets()
	got, ok := presets["test-roundtrip"]
	if !ok {
		t.Fatal("registered theme not found in Presets()")
	}
	if got.Positive != custom.Positive {
		t.Errorf("Positive = %v, want %v", got.Positive, custom.Positive)
	}
	if got.Text != custom.Text {
		t.Errorf("Text = %v, want %v", got.Text, custom.Text)
	}
}

// TestThemePresetsReturnsCopy verifies that Presets returns a copy
// (mutating the map doesn't affect the registry).
func TestThemePresetsReturnsCopy(t *testing.T) {
	presets := Presets()
	originalLen := len(presets)
	presets["injected"] = Theme{}

	presets2 := Presets()
	if len(presets2) != originalLen {
		t.Errorf("Presets should return a copy; mutation leaked (len %d vs %d)", len(presets2), originalLen)
	}
}

// TestContrastBlackOnWhite verifies a known high-contrast pair.
func TestContrastBlackOnWhite(t *testing.T) {
	ratio := Contrast(lipgloss.Color("#000000"), lipgloss.Color("#ffffff"))
	if ratio < 20.9 {
		t.Errorf("black/white contrast = %.2f, want >= 21.0", ratio)
	}
}

// TestContrastSameColor verifies that identical colors have ratio 1.0.
func TestContrastSameColor(t *testing.T) {
	ratio := Contrast(lipgloss.Color("#abcdef"), lipgloss.Color("#abcdef"))
	if ratio < 0.99 || ratio > 1.01 {
		t.Errorf("same color contrast = %.4f, want ~1.0", ratio)
	}
}

// TestSetThemeCmdProducesMsg verifies SetThemeCmd round-trips.
func TestSetThemeCmdProducesMsg(t *testing.T) {
	theme := LightTheme()
	cmd := SetThemeCmd(theme)
	msg := cmd()
	stm, ok := msg.(SetThemeMsg)
	if !ok {
		t.Fatalf("expected SetThemeMsg, got %T", msg)
	}
	if stm.Theme.Text != theme.Text {
		t.Errorf("Text = %v, want %v", stm.Theme.Text, theme.Text)
	}
}

// TestThemeGlyphsOrDefault verifies glyphsOrDefault returns correct glyphs.
func TestThemeGlyphsOrDefault(t *testing.T) {
	// nil Glyphs -> defaults
	theme := DefaultTheme()
	g := theme.glyphsOrDefault()
	def := DefaultGlyphs()
	if g.TreeBranch != def.TreeBranch {
		t.Errorf("glyphsOrDefault with nil Glyphs.TreeBranch = %q, want %q", g.TreeBranch, def.TreeBranch)
	}

	// Custom glyphs
	custom := DefaultGlyphs()
	custom.TreeBranch = "+-"
	theme.Glyphs = &custom
	g2 := theme.glyphsOrDefault()
	if g2.TreeBranch != "+-" {
		t.Errorf("glyphsOrDefault with custom Glyphs.TreeBranch = %q, want %q", g2.TreeBranch, "+-")
	}
}

// TestDefaultBordersNotZero verifies DefaultBorders returns populated set.
func TestDefaultBordersNotZero(t *testing.T) {
	b := DefaultBorders()
	if b.Rounded.Top == "" {
		t.Error("DefaultBorders().Rounded should not have empty Top")
	}
}
