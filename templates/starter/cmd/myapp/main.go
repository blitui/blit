// Package main is the entry point for myapp, a blit-go starter application.
package main

import (
	"fmt"
	"os"

	"github.com/OWNER/myapp/internal/updatewire"
	blit "github.com/blitui/blit"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	items := []string{
		"Item One",
		"Item Two",
		"Item Three",
		"Item Four",
		"Item Five",
	}

	list := blit.NewListView(blit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render(fmt.Sprintf("  %d. %s", idx+1, item))
		},
		HeaderFunc: func(theme blit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  myapp")
		},
	})
	list.SetItems(items)

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("list", list),
		blit.WithStatusBar(
			func() string { return " ↑/↓ navigate  ? help  q quit" },
			func() string { return fmt.Sprintf(" %d items ", len(items)) },
		),
		blit.WithHelp(),
		blit.WithAutoUpdate(updatewire.Config()),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
