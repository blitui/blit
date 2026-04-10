package blit

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("dracula", DraculaTheme())
	Register("catppuccin-mocha", CatppuccinMochaTheme())
	Register("tokyo-night", TokyoNightTheme())
	Register("nord", NordTheme())
	Register("gruvbox-dark", GruvboxDarkTheme())
	Register("rose-pine", RosePineTheme())
	Register("kanagawa", KanagawaTheme())
	Register("one-dark", OneDarkTheme())
}

// DraculaTheme returns the Dracula colour theme.
func DraculaTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#50fa7b"),
		Negative:    lipgloss.Color("#ff5555"),
		Accent:      lipgloss.Color("#bd93f9"),
		Muted:       lipgloss.Color("#6272a4"),
		Text:        lipgloss.Color("#f8f8f2"),
		TextInverse: lipgloss.Color("#282a36"),
		Cursor:      lipgloss.Color("#ff79c6"),
		Border:      lipgloss.Color("#44475a"),
		Flash:       lipgloss.Color("#f1fa8c"),
	}
}

// CatppuccinMochaTheme returns the Catppuccin Mocha colour theme.
func CatppuccinMochaTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#a6e3a1"),
		Negative:    lipgloss.Color("#f38ba8"),
		Accent:      lipgloss.Color("#cba6f7"),
		Muted:       lipgloss.Color("#585b70"),
		Text:        lipgloss.Color("#cdd6f4"),
		TextInverse: lipgloss.Color("#1e1e2e"),
		Cursor:      lipgloss.Color("#89b4fa"),
		Border:      lipgloss.Color("#313244"),
		Flash:       lipgloss.Color("#f9e2af"),
	}
}

// TokyoNightTheme returns the Tokyo Night colour theme.
func TokyoNightTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#9ece6a"),
		Negative:    lipgloss.Color("#f7768e"),
		Accent:      lipgloss.Color("#7aa2f7"),
		Muted:       lipgloss.Color("#565f89"),
		Text:        lipgloss.Color("#c0caf5"),
		TextInverse: lipgloss.Color("#1a1b26"),
		Cursor:      lipgloss.Color("#bb9af7"),
		Border:      lipgloss.Color("#292e42"),
		Flash:       lipgloss.Color("#e0af68"),
	}
}

// NordTheme returns the Nord colour theme.
func NordTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#a3be8c"),
		Negative:    lipgloss.Color("#bf616a"),
		Accent:      lipgloss.Color("#81a1c1"),
		Muted:       lipgloss.Color("#4c566a"),
		Text:        lipgloss.Color("#eceff4"),
		TextInverse: lipgloss.Color("#2e3440"),
		Cursor:      lipgloss.Color("#88c0d0"),
		Border:      lipgloss.Color("#3b4252"),
		Flash:       lipgloss.Color("#ebcb8b"),
	}
}

// GruvboxDarkTheme returns the Gruvbox Dark colour theme.
func GruvboxDarkTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#b8bb26"),
		Negative:    lipgloss.Color("#fb4934"),
		Accent:      lipgloss.Color("#fabd2f"),
		Muted:       lipgloss.Color("#928374"),
		Text:        lipgloss.Color("#ebdbb2"),
		TextInverse: lipgloss.Color("#282828"),
		Cursor:      lipgloss.Color("#83a598"),
		Border:      lipgloss.Color("#504945"),
		Flash:       lipgloss.Color("#d3869b"),
	}
}

// RosePineTheme returns the Rose Pine colour theme.
func RosePineTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#31748f"),
		Negative:    lipgloss.Color("#eb6f92"),
		Accent:      lipgloss.Color("#c4a7e7"),
		Muted:       lipgloss.Color("#6e6a86"),
		Text:        lipgloss.Color("#e0def4"),
		TextInverse: lipgloss.Color("#191724"),
		Cursor:      lipgloss.Color("#9ccfd8"),
		Border:      lipgloss.Color("#403d52"),
		Flash:       lipgloss.Color("#f6c177"),
	}
}

// KanagawaTheme returns the Kanagawa colour theme.
func KanagawaTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#98bb6c"),
		Negative:    lipgloss.Color("#e46876"),
		Accent:      lipgloss.Color("#7e9cd8"),
		Muted:       lipgloss.Color("#727169"),
		Text:        lipgloss.Color("#dcd7ba"),
		TextInverse: lipgloss.Color("#1f1f28"),
		Cursor:      lipgloss.Color("#957fb8"),
		Border:      lipgloss.Color("#2a2a37"),
		Flash:       lipgloss.Color("#e98a00"),
	}
}

// OneDarkTheme returns the One Dark colour theme.
func OneDarkTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#98c379"),
		Negative:    lipgloss.Color("#e06c75"),
		Accent:      lipgloss.Color("#61afef"),
		Muted:       lipgloss.Color("#5c6370"),
		Text:        lipgloss.Color("#abb2bf"),
		TextInverse: lipgloss.Color("#282c34"),
		Cursor:      lipgloss.Color("#c678dd"),
		Border:      lipgloss.Color("#3e4451"),
		Flash:       lipgloss.Color("#e5c07b"),
	}
}
