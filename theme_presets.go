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
	Register("solarized-dark", SolarizedDarkTheme())
	Register("everforest", EverforestTheme())
	Register("nightfox", NightfoxTheme())
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
		Warn:        lipgloss.Color("#f1fa8c"),
		Extra: map[string]lipgloss.Color{
			"info":    "#8be9fd",
			"create":  "#50fa7b",
			"delete":  "#ff5555",
			"review":  "#bd93f9",
			"comment": "#6272a4",
			"issue":   "#ffb86c",
			"release": "#ff79c6",
			"local":   "#8be9fd",
		},
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
		Warn:        lipgloss.Color("#f9e2af"),
		Extra: map[string]lipgloss.Color{
			"info":    "#89dceb",
			"create":  "#a6e3a1",
			"delete":  "#f38ba8",
			"review":  "#cba6f7",
			"comment": "#585b70",
			"issue":   "#fab387",
			"release": "#f5c2e7",
			"local":   "#94e2d5",
		},
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
		Warn:        lipgloss.Color("#e0af68"),
		Extra: map[string]lipgloss.Color{
			"info":    "#7dcfff",
			"create":  "#9ece6a",
			"delete":  "#f7768e",
			"review":  "#bb9af7",
			"comment": "#565f89",
			"issue":   "#ff9e64",
			"release": "#c0caf5",
			"local":   "#73daca",
		},
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
		Warn:        lipgloss.Color("#ebcb8b"),
		Extra: map[string]lipgloss.Color{
			"info":    "#88c0d0",
			"create":  "#a3be8c",
			"delete":  "#bf616a",
			"review":  "#b48ead",
			"comment": "#4c566a",
			"issue":   "#d08770",
			"release": "#81a1c1",
			"local":   "#8fbcbb",
		},
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
		Warn:        lipgloss.Color("#fabd2f"),
		Extra: map[string]lipgloss.Color{
			"info":    "#83a598",
			"create":  "#b8bb26",
			"delete":  "#fb4934",
			"review":  "#d3869b",
			"comment": "#928374",
			"issue":   "#fe8019",
			"release": "#fabd2f",
			"local":   "#8ec07c",
		},
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
		Warn:        lipgloss.Color("#f6c177"),
		Extra: map[string]lipgloss.Color{
			"info":    "#9ccfd8",
			"create":  "#31748f",
			"delete":  "#eb6f92",
			"review":  "#c4a7e7",
			"comment": "#6e6a86",
			"issue":   "#f6c177",
			"release": "#ebbcba",
			"local":   "#9ccfd8",
		},
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
		Warn:        lipgloss.Color("#e98a00"),
		Extra: map[string]lipgloss.Color{
			"info":    "#7fb4ca",
			"create":  "#98bb6c",
			"delete":  "#e46876",
			"review":  "#957fb8",
			"comment": "#727169",
			"issue":   "#ffa066",
			"release": "#d27e99",
			"local":   "#7aa89f",
		},
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
		Warn:        lipgloss.Color("#e5c07b"),
		Extra: map[string]lipgloss.Color{
			"info":    "#56b6c2",
			"create":  "#98c379",
			"delete":  "#e06c75",
			"review":  "#c678dd",
			"comment": "#5c6370",
			"issue":   "#d19a66",
			"release": "#61afef",
			"local":   "#56b6c2",
		},
	}
}

// SolarizedDarkTheme returns the Solarized Dark colour theme.
func SolarizedDarkTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#859900"),
		Negative:    lipgloss.Color("#dc322f"),
		Accent:      lipgloss.Color("#268bd2"),
		Muted:       lipgloss.Color("#586e75"),
		Text:        lipgloss.Color("#839496"),
		TextInverse: lipgloss.Color("#002b36"),
		Cursor:      lipgloss.Color("#6c71c4"),
		Border:      lipgloss.Color("#073642"),
		Flash:       lipgloss.Color("#b58900"),
		Warn:        lipgloss.Color("#cb4b16"),
		Extra: map[string]lipgloss.Color{
			"info":    "#2aa198",
			"create":  "#859900",
			"delete":  "#dc322f",
			"review":  "#6c71c4",
			"comment": "#586e75",
			"issue":   "#cb4b16",
			"release": "#d33682",
			"local":   "#2aa198",
		},
	}
}

// EverforestTheme returns the Everforest Dark colour theme.
func EverforestTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#a7c080"),
		Negative:    lipgloss.Color("#e67e80"),
		Accent:      lipgloss.Color("#7fbbb3"),
		Muted:       lipgloss.Color("#859289"),
		Text:        lipgloss.Color("#d3c6aa"),
		TextInverse: lipgloss.Color("#2d353b"),
		Cursor:      lipgloss.Color("#d699b6"),
		Border:      lipgloss.Color("#475258"),
		Flash:       lipgloss.Color("#dbbc7f"),
		Warn:        lipgloss.Color("#e69875"),
		Extra: map[string]lipgloss.Color{
			"info":    "#7fbbb3",
			"create":  "#a7c080",
			"delete":  "#e67e80",
			"review":  "#d699b6",
			"comment": "#859289",
			"issue":   "#e69875",
			"release": "#7fbbb3",
			"local":   "#83c092",
		},
	}
}

// NightfoxTheme returns the Nightfox colour theme.
func NightfoxTheme() Theme {
	return Theme{
		Positive:    lipgloss.Color("#81b29a"),
		Negative:    lipgloss.Color("#c94f6d"),
		Accent:      lipgloss.Color("#719cd6"),
		Muted:       lipgloss.Color("#738091"),
		Text:        lipgloss.Color("#cdcecf"),
		TextInverse: lipgloss.Color("#192330"),
		Cursor:      lipgloss.Color("#9d79d6"),
		Border:      lipgloss.Color("#2b3b51"),
		Flash:       lipgloss.Color("#dbc074"),
		Warn:        lipgloss.Color("#f4a261"),
		Extra: map[string]lipgloss.Color{
			"info":    "#63cdcf",
			"create":  "#81b29a",
			"delete":  "#c94f6d",
			"review":  "#9d79d6",
			"comment": "#738091",
			"issue":   "#f4a261",
			"release": "#719cd6",
			"local":   "#63cdcf",
		},
	}
}
