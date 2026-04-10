package main

const dashboardMainTmpl = `package main

import (
	"fmt"
	"os"

	app "{{.ModulePath}}/internal/{{.BinaryName}}"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
`

const dashboardAppTmpl = `package {{.BinaryName}}

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

// New creates a dashboard application with a table and split pane layout.
func New() *blit.App {
	columns := []blit.Column{
		{Title: "Name", Width: 20, Sortable: true},
		{Title: "Status", Width: 15},
		{Title: "Value", Width: 10, Align: blit.Right, Sortable: true},
	}

	rows := []blit.Row{
		{"Server Alpha", "Running", "98.5"},
		{"Server Beta", "Running", "87.2"},
		{"Server Gamma", "Warning", "45.1"},
		{"Server Delta", "Stopped", "0.0"},
		{"Server Epsilon", "Running", "91.7"},
	}

	table := blit.NewTable(columns, rows, blit.TableOpts{
		Sortable:   true,
		Filterable: true,
	})

	sidebar := newSidebar(len(rows))

	return blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithSlot(blit.SlotMain, table),
		blit.WithSlot(blit.SlotSidebar, sidebar),
		blit.WithStatusBar(
			func() string { return " / search  s sort  ? help  q quit" },
			func() string { return fmt.Sprintf(" %d items ", len(rows)) },
		),
		blit.WithHelp(),
	)
}

// sidebar is a simple stats panel.
type sidebar struct {
	theme         blit.Theme
	focused       bool
	width, height int
	total         int
}

func newSidebar(total int) *sidebar { return &sidebar{total: total} }

func (s *sidebar) Init() blit.Cmd                                         { return nil }
func (s *sidebar) Update(_ blit.Msg, _ blit.Context) (blit.Component, blit.Cmd) { return s, nil }
func (s *sidebar) KeyBindings() []blit.KeyBind                            { return nil }
func (s *sidebar) SetSize(w, h int)                                       { s.width, s.height = w, h }
func (s *sidebar) Focused() bool                                          { return s.focused }
func (s *sidebar) SetFocused(f bool)                                      { s.focused = f }
func (s *sidebar) SetTheme(t blit.Theme)                                  { s.theme = t }

func (s *sidebar) View() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Accent)).
		Bold(true)
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Muted))
	value := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.Text)).
		Bold(true)

	return fmt.Sprintf("%s\n\n%s %s\n%s %s\n%s %s",
		title.Render("  Dashboard"),
		label.Render("  Total:"),
		value.Render(fmt.Sprintf("%d", s.total)),
		label.Render("  Active:"),
		value.Render(fmt.Sprintf("%d", s.total-1)),
		label.Render("  Alerts:"),
		value.Render("1"),
	)
}
`

const dashboardAppTestTmpl = `package {{.BinaryName}}_test

import (
	"testing"

	"github.com/blitui/blit/btest"

	app "{{.ModulePath}}/internal/{{.BinaryName}}"
)

func TestApp_Renders(t *testing.T) {
	a := app.New()
	tm := btest.NewTestModel(t, a.Model(), 120, 40)

	tm.RequireScreen(func(t testing.TB, s *btest.Screen) {
		btest.AssertNotEmpty(t, s)
		btest.AssertContains(t, s, "Dashboard")
		btest.AssertContains(t, s, "Name")
		btest.AssertContains(t, s, "Status")
	})
}
`
