// Package main demonstrates the tuikit Tabs component.
//
// This example nests a Table, a simple list, and a text pane behind three tabs.
// Keybinds: tab/shift+tab to cycle tabs, 1-3 to jump, mouse click on tab bar.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// ── Tab 1: Table ─────────────────────────────────────────────────────────────

func buildTableTab() tuikit.Component {
	columns := []tuikit.Column{
		{Title: "Language", Width: 18, Sortable: true},
		{Title: "Paradigm", Width: 20},
		{Title: "Year", Width: 8, Align: tuikit.Right, Sortable: true},
		{Title: "Creator", Width: 22},
	}
	rows := []tuikit.Row{
		{"Go", "Compiled/Concurrent", "2009", "Google"},
		{"Rust", "Systems/Safe", "2010", "Mozilla"},
		{"Python", "Dynamic/OOP", "1991", "van Rossum"},
		{"TypeScript", "Typed/OOP", "2012", "Microsoft"},
		{"Haskell", "Pure Functional", "1990", "Committee"},
		{"Zig", "Systems/Explicit", "2016", "Kelley"},
		{"Elixir", "Functional/Concurrent", "2011", "Valim"},
		{"Swift", "Multi-paradigm", "2014", "Apple"},
	}
	return tuikit.NewTable(columns, rows, tuikit.TableOpts{
		Sortable:   true,
		Filterable: true,
	})
}

// ── Tab 2: ListView ──────────────────────────────────────────────────────────

type noteItem struct {
	id   int
	text string
}

func buildListTab() tuikit.Component {
	items := []noteItem{
		{1, "Read the tuikit-go docs"},
		{2, "Implement Tabs component"},
		{3, "Write tuitest coverage"},
		{4, "Ship v0.8.0"},
		{5, "Celebrate with pizza"},
	}
	lv := tuikit.NewListView[noteItem](tuikit.ListViewOpts[noteItem]{
		RenderItem: func(item noteItem, idx int, isCursor bool, theme tuikit.Theme) string {
			check := "○"
			if item.id%2 == 0 {
				check = "●"
			}
			return fmt.Sprintf("%s  %s", check, item.text)
		},
		HeaderFunc: func(theme tuikit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  TODO LIST")
		},
	})
	lv.SetItems(items)
	return lv
}

// ── Tab 3: Info pane ─────────────────────────────────────────────────────────

type infoPane struct {
	theme   tuikit.Theme
	focused bool
	width   int
	height  int
}

func (p *infoPane) Init() tea.Cmd                                      { return nil }
func (p *infoPane) Update(msg tea.Msg) (tuikit.Component, tea.Cmd)     { return p, nil }
func (p *infoPane) KeyBindings() []tuikit.KeyBind                      { return nil }
func (p *infoPane) SetSize(w, h int)                                   { p.width = w; p.height = h }
func (p *infoPane) Focused() bool                                      { return p.focused }
func (p *infoPane) SetFocused(f bool)                                  { p.focused = f }
func (p *infoPane) SetTheme(t tuikit.Theme)                            { p.theme = t }

func (p *infoPane) View() string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Accent)).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Muted))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Text))

	lines := []string{
		accent.Render("  tuikit-go Tabs Demo"),
		"",
		text.Render("  This pane is tab #3 — a plain component."),
		text.Render("  Switch tabs with:"),
		"",
		muted.Render("    tab / shift+tab") + text.Render("  — cycle forward/back"),
		muted.Render("    1, 2, 3        ") + text.Render("  — jump directly"),
		muted.Render("    click on title ") + text.Render("  — mouse select"),
		"",
		text.Render("  The Table and List in the other tabs retain"),
		text.Render("  their scroll positions across tab switches."),
	}
	return strings.Join(lines, "\n")
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	tabs := tuikit.NewTabs([]tuikit.TabItem{
		{Title: "Languages", Glyph: "▦", Content: buildTableTab()},
		{Title: "TODO", Glyph: "✓", Content: buildListTab()},
		{Title: "About", Glyph: "i", Content: &infoPane{}},
	}, tuikit.TabsOpts{
		OnChange: func(i int) {
			_ = i // could update a status bar here
		},
	})

	app := tuikit.NewApp(
		tuikit.WithTheme(tuikit.DefaultTheme()),
		tuikit.WithComponent("tabs", tabs),
		tuikit.WithStatusBar(
			func() string { return " tab/shift+tab — cycle  1-3 — jump  q — quit" },
			func() string { return " tuikit-go tabs demo " },
		),
		tuikit.WithHelp(),
		tuikit.WithMouseSupport(),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
