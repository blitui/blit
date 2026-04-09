// Package main demonstrates a full-featured tuikit dashboard.
//
// This example showcases: Table with sorting/filtering, DualPane layout,
// ConfigEditor overlay, StatusBar, Help, custom cell rendering with semantic
// colors, custom sort functions, and global keybindings — all composed via
// the v0.10 named-slot API.
//
// Theme: "Galactic Pizza Delivery" — 42 pizza deliveries across the galaxy
// with drivers, statuses, ETAs, and tips. Fun data makes for better demos.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// MissionControl is a sidebar panel showing delivery stats.
type MissionControl struct {
	theme         tuikit.Theme
	focused       bool
	width, height int
	stats         missionStats
}

type missionStats struct {
	total, delivered, lost int
	topDriver              string
}

func newMissionControl(s missionStats) *MissionControl { return &MissionControl{stats: s} }

func (m *MissionControl) Init() tea.Cmd { return nil }
func (m *MissionControl) Update(msg tea.Msg, ctx tuikit.Context) (tuikit.Component, tea.Cmd) {
	return m, nil
}
func (m *MissionControl) KeyBindings() []tuikit.KeyBind { return nil }
func (m *MissionControl) SetSize(w, h int)              { m.width, m.height = w, h }
func (m *MissionControl) Focused() bool                 { return m.focused }
func (m *MissionControl) SetFocused(f bool)             { m.focused = f }
func (m *MissionControl) SetTheme(t tuikit.Theme)       { m.theme = t }

func (m *MissionControl) View() string {
	title := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Accent)).Bold(true)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Muted))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Text)).Bold(true)
	pos := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Positive))
	neg := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Negative))
	rate := 0.0
	if m.stats.total > 0 {
		rate = float64(m.stats.delivered) / float64(m.stats.total) * 100
	}
	var sb strings.Builder
	sb.WriteString(title.Render("  MISSION CONTROL") + "\n\n")
	sb.WriteString(label.Render("  Total Deliveries: ") + value.Render(fmt.Sprintf("%d", m.stats.total)) + "\n")
	sb.WriteString(label.Render("  Delivered:        ") + pos.Render(fmt.Sprintf("%d", m.stats.delivered)) + "\n")
	sb.WriteString(label.Render("  Lost in Wormhole: ") + neg.Render(fmt.Sprintf("%d", m.stats.lost)) + "\n")
	sb.WriteString(label.Render("  Success Rate:     ") + value.Render(fmt.Sprintf("%.1f%%", rate)) + "\n\n")
	sb.WriteString(label.Render("  Top Driver:") + "\n" + value.Render("  "+m.stats.topDriver) + "\n\n")
	sb.WriteString(title.Render("  WORST PLANET") + "\n" + neg.Render("  Arrakis") + "\n")
	sb.WriteString(label.Render("  (sand everywhere)"))
	return sb.String()
}

func main() {
	planets := []string{"Mars", "Jupiter", "Saturn", "Neptune", "Pluto", "Kepler-442b", "Proxima b", "Tatooine", "Arrakis", "Gallifrey", "Vulcan", "Krypton", "Ego"}
	drivers := []string{"Zorp McBlast", "Captain Pepperoni", "Turbo Jenkins", "Sally Starfighter", "Buzz Crust", "The Dough Knight"}
	statuses := []string{"In Transit", "Delivered", "Lost in Wormhole", "Dodging Asteroids", "Refueling", "Abducted by Aliens"}

	var rows []tuikit.Row
	delivered, lost := 0, 0
	for i := 0; i < 42; i++ {
		status := statuses[rand.Intn(len(statuses))]
		eta := fmt.Sprintf("%d", rand.Intn(500)+1)
		if status == "Delivered" {
			delivered++
			eta = "0"
		}
		if status == "Lost in Wormhole" {
			lost++
			eta = "999"
		}
		rows = append(rows, tuikit.Row{
			planets[rand.Intn(len(planets))],
			drivers[rand.Intn(len(drivers))],
			status, eta,
			fmt.Sprintf("%d.%02d", rand.Intn(1000), rand.Intn(100)),
		})
	}

	columns := []tuikit.Column{
		{Title: "Planet", Width: 20, Sortable: true},
		{Title: "Driver", Width: 25, Sortable: true},
		{Title: "Status", Width: 25},
		{Title: "ETA (light-min)", Width: 15, MinWidth: 100, Align: tuikit.Right, Sortable: true},
		{Title: "Tip ($)", Width: 10, MinWidth: 120, Align: tuikit.Right, Sortable: true},
	}

	// Custom cell renderer: color cells based on context.
	cellRenderer := func(row tuikit.Row, colIdx int, isCursor bool, theme tuikit.Theme) string {
		if colIdx >= len(row) {
			return ""
		}
		val := row[colIdx]
		var style lipgloss.Style
		switch {
		case colIdx == 2: // Status — semantic colors
			switch val {
			case "Delivered":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive))
			case "Lost in Wormhole":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative))
			case "Dodging Asteroids", "Abducted by Aliens":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Flash))
			default:
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent))
			}
		case colIdx == 4: // Tip — always green, prefixed $
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive))
			val = "$" + val
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
		}
		if isCursor {
			style = style.Background(lipgloss.Color(theme.Cursor))
			if colIdx != 2 && colIdx != 4 {
				style = style.Foreground(lipgloss.Color(theme.TextInverse))
			}
		}
		return style.Render(val)
	}

	// Numeric sort for ETA and Tip columns.
	numericSort := func(a, b tuikit.Row, col int, asc bool) bool {
		va, _ := strconv.ParseFloat(a[col], 64)
		vb, _ := strconv.ParseFloat(b[col], 64)
		if asc {
			return va < vb
		}
		return va > vb
	}

	table := tuikit.NewTable(columns, rows, tuikit.TableOpts{
		Sortable: true, Filterable: true,
		CellRenderer: cellRenderer, SortFunc: numericSort,
	})
	panel := newMissionControl(missionStats{total: 42, delivered: delivered, lost: lost, topDriver: "Captain Pepperoni"})

	pineapple, wormholeLevel, defaultTip := "absolutely not", "11", "15"
	configEditor := tuikit.NewConfigEditor([]tuikit.ConfigField{
		{Label: "Pineapple Allowed", Group: "Pizza Policy", Hint: "this is a serious matter",
			Get: func() string { return pineapple }, Set: func(v string) error { pineapple = v; return nil }},
		{Label: "Wormhole Avoidance", Group: "Navigation", Hint: "scale of 1-11 (11 = maximum avoidance)",
			Get: func() string { return wormholeLevel }, Set: func(v string) error { wormholeLevel = v; return nil }},
		{Label: "Default Tip %", Group: "Finance", Hint: "be generous, they're crossing galaxies",
			Get: func() string { return defaultTip }, Set: func(v string) error { defaultTip = v; return nil }},
	})

	filterModes := []string{"all", "delivered", "lost"}
	filterIdx := 0
	themeList := []tuikit.Theme{tuikit.DefaultTheme()}
	for _, th := range tuikit.Presets() {
		themeList = append(themeList, th)
	}
	themeIdx := 0

	table.SetFilter(func(row tuikit.Row) bool {
		if len(row) < 3 {
			return true
		}
		switch filterModes[filterIdx] {
		case "delivered":
			return row[2] == "Delivered"
		case "lost":
			return row[2] == "Lost in Wormhole"
		}
		return true
	})

	// toastKey builds a keybind that sends a toast — collapses 4 near-
	// identical bindings into a compact table.
	var app *tuikit.App
	toastKey := func(key, label string, sev tuikit.ToastSeverity, title, body string) tuikit.KeyBind {
		return tuikit.KeyBind{Key: key, Label: label, Group: "TOAST", Handler: func() {
			app.Send(tuikit.ToastMsg{Severity: sev, Title: title, Body: body, Duration: 4 * time.Second})
		}}
	}

	app = tuikit.NewApp(
		tuikit.WithTheme(tuikit.DefaultTheme()),
		tuikit.WithSlot(tuikit.SlotMain, table),
		tuikit.WithSlot(tuikit.SlotSidebar, panel),
		tuikit.WithStatusBar(
			func() string {
				return fmt.Sprintf(" ? help  / search  s sort  c config  f filter[%s]  p panel  q quit", filterModes[filterIdx])
			},
			func() string { return "42 active deliveries  Galactic Pizza Corp " },
		),
		tuikit.WithHelp(),
		tuikit.WithSlotOverlay("Settings", "c", configEditor),
		tuikit.WithKeyBind(tuikit.KeyBind{Key: "f", Label: "Cycle filter", Group: "DATA", Handler: func() {
			filterIdx = (filterIdx + 1) % len(filterModes)
			table.SetRows(rows)
		}}),
		tuikit.WithKeyBind(tuikit.KeyBind{Key: "ctrl+t", Label: "Cycle theme", Group: "VIEW", HandlerCmd: func() tea.Cmd {
			themeIdx = (themeIdx + 1) % len(themeList)
			return tuikit.SetThemeCmd(themeList[themeIdx])
		}}),
		tuikit.WithKeyBind(toastKey("i", "Info toast", tuikit.SeverityInfo, "Info", "Mission is a go!")),
		tuikit.WithKeyBind(toastKey("s", "Success toast", tuikit.SeveritySuccess, "Delivered!", "Pizza arrived at destination.")),
		tuikit.WithKeyBind(toastKey("w", "Warning toast", tuikit.SeverityWarn, "Asteroid Field", "Wormhole instability detected.")),
		tuikit.WithKeyBind(toastKey("e", "Error toast", tuikit.SeverityError, "Delivery Failed", "Lost in wormhole. No pizza.")),
		tuikit.WithMouseSupport(),
		tuikit.WithTickInterval(100*time.Millisecond),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
