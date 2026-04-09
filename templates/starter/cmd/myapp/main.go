// Package main is the entry point for myapp, a tuikit-go starter application.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/OWNER/myapp/internal/updatewire"
)

func main() {
	items := []string{
		"Item One",
		"Item Two",
		"Item Three",
		"Item Four",
		"Item Five",
	}

	list := tuikit.NewListView(tuikit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme tuikit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render(fmt.Sprintf("  %d. %s", idx+1, item))
		},
		HeaderFunc: func(theme tuikit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  myapp")
		},
	})
	list.SetItems(items)

	app := tuikit.NewApp(
		tuikit.WithTheme(tuikit.DefaultTheme()),
		tuikit.WithComponent("list", list),
		tuikit.WithStatusBar(
			func() string { return " ↑/↓ navigate  ? help  q quit" },
			func() string { return fmt.Sprintf(" %d items ", len(items)) },
		),
		tuikit.WithHelp(),
		tuikit.WithAutoUpdate(updatewire.Config()),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
