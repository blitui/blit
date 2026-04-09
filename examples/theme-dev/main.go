// Package main demonstrates tuikit's theme hot-reload feature.
// Edit examples/theme-dev/theme.yaml while this program is running and watch
// the colors update live without restarting.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func main() {
	themeFile := "examples/theme-dev/theme.yaml"
	if len(os.Args) > 1 {
		themeFile = os.Args[1]
	}

	items := []string{
		"Positive  — gains, success, online",
		"Negative  — losses, errors, offline",
		"Accent    — highlights, active elements",
		"Muted     — dimmed text, secondary info",
		"Cursor    — selection highlight",
		"Flash     — temporary notifications",
		"Border    — borders and separators",
	}

	list := tuikit.NewListView(tuikit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme tuikit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render("  " + item)
		},
		HeaderFunc: func(theme tuikit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  Theme Token Preview — edit theme.yaml to see live changes")
		},
	})
	list.SetItems(items)

	app := tuikit.NewApp(
		tuikit.WithComponent("tokens", list),
		tuikit.WithStatusBar(
			func() string { return fmt.Sprintf(" watching: %s", themeFile) },
			func() string { return " q quit" },
		),
		tuikit.WithThemeHotReload(themeFile),
		tuikit.WithHelp(),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
