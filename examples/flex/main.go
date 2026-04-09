// Package main demonstrates HBox/VBox flex layout with mixed fixed and flex children.
//
// Layout:
//
//	┌──────────────────────────────────────────────────────────┐
//	│  HEADER (fixed 3 rows, full width)                       │
//	├─────────────────────────────┬────────────────────────────┤
//	│  LEFT SIDEBAR (fixed 24w)   │  CONTENT AREA (flex, grow) │
//	│  · Nav item 1               │  · Large text area         │
//	│  · Nav item 2               │    rendered here           │
//	│  · Nav item 3               │                            │
//	├─────────────────────────────┴────────────────────────────┤
//	│  FOOTER BAR  (fixed 1 row, SpaceBetween two labels)      │
//	└──────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// --- Simple stub components ---

type headerPane struct {
	width, height int
	theme         tuikit.Theme
}

func (h *headerPane) Init() tea.Cmd                            { return nil }
func (h *headerPane) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return h, nil }
func (h *headerPane) KeyBindings() []tuikit.KeyBind            { return nil }
func (h *headerPane) SetSize(w, ht int)                        { h.width = w; h.height = ht }
func (h *headerPane) Focused() bool                            { return false }
func (h *headerPane) SetFocused(bool)                          {}
func (h *headerPane) SetTheme(t tuikit.Theme)                  { h.theme = t }
func (h *headerPane) View() string {
	style := lipgloss.NewStyle().
		Width(h.width).
		Background(lipgloss.Color(h.theme.Accent)).
		Foreground(lipgloss.Color(h.theme.TextInverse)).
		Bold(true).
		Padding(0, 2)
	title := style.Render("  tuikit HBox/VBox Flex Layout Demo")
	sub := lipgloss.NewStyle().
		Width(h.width).
		Foreground(lipgloss.Color(h.theme.Muted)).
		Padding(0, 2).
		Render("Mixed fixed + flex children, nested boxes, gap, align, justify")
	sep := lipgloss.NewStyle().
		Width(h.width).
		Foreground(lipgloss.Color(h.theme.Border)).
		Render(strings.Repeat("─", h.width))
	return strings.Join([]string{title, sub, sep}, "\n")
}

type sidebarPane struct {
	width, height int
	theme         tuikit.Theme
	focused       bool
}

func (s *sidebarPane) Init() tea.Cmd                            { return nil }
func (s *sidebarPane) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return s, nil }
func (s *sidebarPane) KeyBindings() []tuikit.KeyBind            { return nil }
func (s *sidebarPane) SetSize(w, h int)                         { s.width = w; s.height = h }
func (s *sidebarPane) Focused() bool                            { return s.focused }
func (s *sidebarPane) SetFocused(f bool)                        { s.focused = f }
func (s *sidebarPane) SetTheme(t tuikit.Theme)                  { s.theme = t }
func (s *sidebarPane) View() string {
	borderColor := s.theme.Border
	if s.focused {
		borderColor = s.theme.Accent
	}
	box := lipgloss.NewStyle().
		Width(s.width - 2).
		Height(s.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1)

	items := []string{"Navigation", "", "  > Dashboard", "    Reports", "    Analytics", "    Settings", "", "  > Components", "    HBox", "    VBox", "    Sized", "    Flex"}
	content := strings.Join(items, "\n")
	return box.Render(content)
}

type contentPane struct {
	width, height int
	theme         tuikit.Theme
	focused       bool
}

func (c *contentPane) Init() tea.Cmd                            { return nil }
func (c *contentPane) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return c, nil }
func (c *contentPane) KeyBindings() []tuikit.KeyBind            { return nil }
func (c *contentPane) SetSize(w, h int)                         { c.width = w; c.height = h }
func (c *contentPane) Focused() bool                            { return c.focused }
func (c *contentPane) SetFocused(f bool)                        { c.focused = f }
func (c *contentPane) SetTheme(t tuikit.Theme)                  { c.theme = t }
func (c *contentPane) View() string {
	borderColor := c.theme.Border
	if c.focused {
		borderColor = c.theme.Accent
	}
	box := lipgloss.NewStyle().
		Width(c.width - 2).
		Height(c.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Accent)).
		Bold(true).
		Render("Flex Layout Content Area")

	desc := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.theme.Text)).
		Render("This pane has Grow=1 and expands to fill remaining width.\n" +
			"The sidebar is Sized{W:26} — fixed regardless of terminal width.\n\n" +
			"HBox children:\n" +
			"  · Sized{W:26}  — sidebar (fixed)\n" +
			"  · Flex{Grow:1} — this pane (fills rest)\n\n" +
			"VBox outer:\n" +
			"  · Sized{W:3}   — header (3 rows)\n" +
			"  · Flex{Grow:1} — body HBox\n" +
			"  · Sized{W:1}   — footer (1 row)\n\n" +
			"Gap=1 separates each row. AlignStretch fills cross axis.")

	return box.Render(title + "\n\n" + desc)
}

type footerPane struct {
	width, height int
	theme         tuikit.Theme
}

func (f *footerPane) Init() tea.Cmd                            { return nil }
func (f *footerPane) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return f, nil }
func (f *footerPane) KeyBindings() []tuikit.KeyBind            { return nil }
func (f *footerPane) SetSize(w, h int)                         { f.width = w; f.height = h }
func (f *footerPane) Focused() bool                            { return false }
func (f *footerPane) SetFocused(bool)                          {}
func (f *footerPane) SetTheme(t tuikit.Theme)                  { f.theme = t }
func (f *footerPane) View() string {
	left := lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.theme.Muted)).
		Render("  tuikit-go v0.9 · Flex layout")
	right := lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.theme.Muted)).
		Render("q quit  ")
	// SpaceBetween via HBox
	inner := &tuikit.HBox{
		Gap:     0,
		Justify: tuikit.FlexJustifySpaceBetween,
		Items: []tuikit.Component{
			tuikit.Sized{W: len([]rune(lipgloss.NewStyle().Render(left))), C: &labelComp{text: left}},
			tuikit.Sized{W: len([]rune(lipgloss.NewStyle().Render(right))), C: &labelComp{text: right}},
		},
	}
	inner.SetSize(f.width, 1)
	return inner.View()
}

// labelComp is a minimal text component used inside the footer HBox.
type labelComp struct {
	text          string
	width, height int
}

func (l *labelComp) Init() tea.Cmd                            { return nil }
func (l *labelComp) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return l, nil }
func (l *labelComp) KeyBindings() []tuikit.KeyBind            { return nil }
func (l *labelComp) SetSize(w, h int)                         { l.width = w; l.height = h }
func (l *labelComp) Focused() bool                            { return false }
func (l *labelComp) SetFocused(bool)                          {}
func (l *labelComp) View() string                             { return l.text }

// --- Root layout component ---

// flexDemo is the root Component that owns the VBox/HBox tree.
type flexDemo struct {
	width, height int
	theme         tuikit.Theme

	header  *headerPane
	sidebar *sidebarPane
	content *contentPane
	footer  *footerPane

	body  *tuikit.HBox
	outer *tuikit.VBox
}

func newFlexDemo() *flexDemo {
	d := &flexDemo{
		header:  &headerPane{},
		sidebar: &sidebarPane{},
		content: &contentPane{},
		footer:  &footerPane{},
	}

	d.body = &tuikit.HBox{
		Gap:   1,
		Align: tuikit.FlexAlignStretch,
		Items: []tuikit.Component{
			tuikit.Sized{W: 26, C: d.sidebar},
			tuikit.Flex{Grow: 1, C: d.content},
		},
	}

	d.outer = &tuikit.VBox{
		Gap:   0,
		Align: tuikit.FlexAlignStretch,
		Items: []tuikit.Component{
			tuikit.Sized{W: 3, C: d.header},
			tuikit.Flex{Grow: 1, C: d.body},
			tuikit.Sized{W: 1, C: d.footer},
		},
	}
	return d
}

func (d *flexDemo) Init() tea.Cmd { return d.outer.Init() }

func (d *flexDemo) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) {
	updated, cmd := d.outer.Update(msg)
	d.outer = updated.(*tuikit.VBox)
	return d, cmd
}

func (d *flexDemo) View() string { return d.outer.View() }

func (d *flexDemo) KeyBindings() []tuikit.KeyBind { return nil }

func (d *flexDemo) SetSize(w, h int) {
	d.width = w
	d.height = h
	d.outer.SetSize(w, h)
}

func (d *flexDemo) Focused() bool      { return false }
func (d *flexDemo) SetFocused(f bool)  {}

func (d *flexDemo) SetTheme(t tuikit.Theme) {
	d.theme = t
	d.header.SetTheme(t)
	d.sidebar.SetTheme(t)
	d.content.SetTheme(t)
	d.footer.SetTheme(t)
}

func main() {
	demo := newFlexDemo()

	app := tuikit.NewApp(
		tuikit.WithTheme(tuikit.DefaultTheme()),
		tuikit.WithComponent("main", demo),
		tuikit.WithStatusBar(
			func() string { return "  HBox/VBox flex layout demo" },
			func() string { return "q quit " },
		),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
