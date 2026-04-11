package blit_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	blit "github.com/blitui/blit"
)

func TestPresets_AllRegistered(t *testing.T) {
	presets := blit.Presets()
	expected := []string{
		"dracula",
		"catppuccin-mocha",
		"tokyo-night",
		"nord",
		"gruvbox-dark",
		"rose-pine",
		"kanagawa",
		"one-dark",
		"solarized-dark",
		"everforest",
		"nightfox",
	}
	for _, name := range expected {
		if _, ok := presets[name]; !ok {
			t.Errorf("preset %q not registered", name)
		}
	}
}

func TestPresets_RequiredFields(t *testing.T) {
	presets := blit.Presets()
	for name, theme := range presets {
		if theme.Positive == lipgloss.Color("") {
			t.Errorf("%s: Positive is empty", name)
		}
		if theme.Negative == lipgloss.Color("") {
			t.Errorf("%s: Negative is empty", name)
		}
		if theme.Accent == lipgloss.Color("") {
			t.Errorf("%s: Accent is empty", name)
		}
		if theme.Muted == lipgloss.Color("") {
			t.Errorf("%s: Muted is empty", name)
		}
		if theme.Text == lipgloss.Color("") {
			t.Errorf("%s: Text is empty", name)
		}
		if theme.TextInverse == lipgloss.Color("") {
			t.Errorf("%s: TextInverse is empty", name)
		}
		if theme.Cursor == lipgloss.Color("") {
			t.Errorf("%s: Cursor is empty", name)
		}
		if theme.Border == lipgloss.Color("") {
			t.Errorf("%s: Border is empty", name)
		}
	}
}

func TestPresets_HaveExtraColors(t *testing.T) {
	presets := blit.Presets()
	requiredExtras := []string{"info", "create", "delete", "review", "comment", "issue", "release", "local"}
	for name, theme := range presets {
		if theme.Extra == nil {
			t.Errorf("%s: Extra map is nil", name)
			continue
		}
		for _, key := range requiredExtras {
			if _, ok := theme.Extra[key]; !ok {
				t.Errorf("%s: Extra[%q] missing", name, key)
			}
		}
	}
}

func TestPresets_ReturnsACopy(t *testing.T) {
	p1 := blit.Presets()
	p2 := blit.Presets()
	delete(p1, "dracula")
	if _, ok := p2["dracula"]; !ok {
		t.Fatal("Presets() should return independent copies")
	}
}

func TestDraculaTheme(t *testing.T) {
	th := blit.DraculaTheme()
	if th.Accent != lipgloss.Color("#bd93f9") {
		t.Fatalf("Dracula Accent = %v, want #bd93f9", th.Accent)
	}
}

func TestCatppuccinMochaTheme(t *testing.T) {
	th := blit.CatppuccinMochaTheme()
	if th.Accent != lipgloss.Color("#cba6f7") {
		t.Fatalf("CatppuccinMocha Accent = %v, want #cba6f7", th.Accent)
	}
}

func TestTokyoNightTheme(t *testing.T) {
	th := blit.TokyoNightTheme()
	if th.Accent != lipgloss.Color("#7aa2f7") {
		t.Fatalf("TokyoNight Accent = %v, want #7aa2f7", th.Accent)
	}
}

func TestNordTheme(t *testing.T) {
	th := blit.NordTheme()
	if th.Accent != lipgloss.Color("#81a1c1") {
		t.Fatalf("Nord Accent = %v, want #81a1c1", th.Accent)
	}
}

func TestGruvboxDarkTheme(t *testing.T) {
	th := blit.GruvboxDarkTheme()
	if th.Accent != lipgloss.Color("#fabd2f") {
		t.Fatalf("GruvboxDark Accent = %v, want #fabd2f", th.Accent)
	}
}

func TestRosePineTheme(t *testing.T) {
	th := blit.RosePineTheme()
	if th.Accent != lipgloss.Color("#c4a7e7") {
		t.Fatalf("RosePine Accent = %v, want #c4a7e7", th.Accent)
	}
}

func TestKanagawaTheme(t *testing.T) {
	th := blit.KanagawaTheme()
	if th.Accent != lipgloss.Color("#7e9cd8") {
		t.Fatalf("Kanagawa Accent = %v, want #7e9cd8", th.Accent)
	}
}

func TestOneDarkTheme(t *testing.T) {
	th := blit.OneDarkTheme()
	if th.Accent != lipgloss.Color("#61afef") {
		t.Fatalf("OneDark Accent = %v, want #61afef", th.Accent)
	}
}

func TestSolarizedDarkTheme(t *testing.T) {
	th := blit.SolarizedDarkTheme()
	if th.Accent != lipgloss.Color("#268bd2") {
		t.Fatalf("SolarizedDark Accent = %v, want #268bd2", th.Accent)
	}
}

func TestEverforestTheme(t *testing.T) {
	th := blit.EverforestTheme()
	if th.Accent != lipgloss.Color("#7fbbb3") {
		t.Fatalf("Everforest Accent = %v, want #7fbbb3", th.Accent)
	}
}

func TestNightfoxTheme(t *testing.T) {
	th := blit.NightfoxTheme()
	if th.Accent != lipgloss.Color("#719cd6") {
		t.Fatalf("Nightfox Accent = %v, want #719cd6", th.Accent)
	}
}
