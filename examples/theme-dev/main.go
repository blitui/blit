// Package main demonstrates blit's theme hot-reload feature.
// Edit examples/theme-dev/theme.yaml while this program is running and watch
// the colors update live without restarting.
package main

import (
	"fmt"
	"os"

	blit "github.com/blitui/blit"
	"github.com/charmbracelet/lipgloss"
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

	list := blit.NewListView(blit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render("  " + item)
		},
		HeaderFunc: func(theme blit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  Theme Token Preview — edit theme.yaml to see live changes")
		},
	})
	list.SetItems(items)

	app := blit.NewApp(
		blit.WithComponent("tokens", list),
		blit.WithStatusBar(
			func() string { return fmt.Sprintf(" watching: %s", themeFile) },
			func() string { return " q quit" },
		),
		blit.WithThemeHotReload(themeFile),
		blit.WithHelp(),
	)

	if _, err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
