package blit

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func colorOf(s string) lipgloss.Color { return lipgloss.Color(s) }

// TestPresetWCAGAA asserts every registered preset achieves WCAG AA contrast
// (>= 4.5) between its Text and TextInverse colors.
func TestPresetWCAGAA(t *testing.T) {
	const minContrast = 4.5
	presets := Presets()
	if len(presets) == 0 {
		t.Fatal("Presets() returned empty map — init() may not have run")
	}
	for name, theme := range presets {
		t.Run(name, func(t *testing.T) {
			ratio := Contrast(theme.TextInverse, theme.Text)
			if ratio < minContrast {
				t.Errorf("%s: Text/TextInverse contrast %.2f < 4.5 (WCAG AA)", name, ratio)
			}
		})
	}
}

// TestContrastKnownValues checks the Contrast function against known ratios.
func TestContrastKnownValues(t *testing.T) {
	cases := []struct {
		bg, fg   string
		minRatio float64
		label    string
	}{
		{"#000000", "#ffffff", 21.0, "black/white"},
		{"#ffffff", "#000000", 21.0, "white/black"},
		{"#ffffff", "#ffffff", 1.0, "same color"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got := Contrast(colorOf(tc.bg), colorOf(tc.fg))
			if tc.label == "same color" {
				if fmt.Sprintf("%.2f", got) != "1.00" {
					t.Errorf("same color contrast = %.2f, want 1.00", got)
				}
			} else if got < tc.minRatio-0.1 {
				t.Errorf("%s contrast = %.2f, want >= %.2f", tc.label, got, tc.minRatio)
			}
		})
	}
}

// TestPresetPresetsCount ensures at least 8 built-in presets are registered.
func TestPresetPresetsCount(t *testing.T) {
	presets := Presets()
	if len(presets) < 8 {
		t.Errorf("expected >= 8 presets, got %d", len(presets))
	}
}

// TestPresetNames checks all expected preset names are registered.
func TestPresetNames(t *testing.T) {
	expected := []string{
		"dracula", "catppuccin-mocha", "tokyo-night", "nord",
		"gruvbox-dark", "rose-pine", "kanagawa", "one-dark",
	}
	presets := Presets()
	for _, name := range expected {
		if _, ok := presets[name]; !ok {
			t.Errorf("preset %q not found in registry", name)
		}
	}
}
