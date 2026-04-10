// Package main demonstrates the simplest possible blit application.
// A working TUI in ~30 lines: a movie watchlist with navigation, status bar, and help.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

func main() {
	movies := []string{
		"The Matrix", "Blade Runner 2049", "Interstellar",
		"Arrival", "Ex Machina", "Dune", "WALL-E",
		"2001: A Space Odyssey", "Alien", "Moon",
	}

	list := blit.NewListView(blit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render(fmt.Sprintf("%d. %s", idx+1, item))
		},
		HeaderFunc: func(theme blit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  🎬 Sci-Fi Watchlist")
		},
	})
	list.SetItems(movies)

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("watchlist", list),
		blit.WithStatusBar(
			func() string { return " ↑/↓ navigate  ? help  q quit" },
			func() string { return fmt.Sprintf(" %d movies ", len(movies)) },
		),
		blit.WithHelp(),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
