package main

import (
	"testing"

	blit "github.com/blitui/blit"
)

func TestSortedThemeNames(t *testing.T) {
	presets := blit.Presets()
	presets["default"] = blit.DefaultTheme()
	presets["light"] = blit.LightTheme()

	names := sortedThemeNames(presets)
	if len(names) < 2 {
		t.Fatalf("expected at least 2 themes, got %d", len(names))
	}
	// Verify sorted order.
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

func TestRunTheme_List(t *testing.T) {
	// Should not panic and return 0.
	code := runTheme(nil)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestRunTheme_Preview(t *testing.T) {
	code := runTheme([]string{"dracula"})
	if code != 0 {
		t.Errorf("expected exit 0 for dracula, got %d", code)
	}
}

func TestRunTheme_Unknown(t *testing.T) {
	code := runTheme([]string{"nonexistent-theme"})
	if code != 1 {
		t.Errorf("expected exit 1 for unknown theme, got %d", code)
	}
}
