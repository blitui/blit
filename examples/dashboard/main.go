// Package main demonstrates a full-featured tuikit dashboard.
//
// This example showcases: Table with sorting/filtering, DualPane layout,
// ConfigEditor overlay, StatusBar, Help, custom cell rendering with semantic
// colors, custom sort functions, and global keybindings.
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
	theme   tuikit.Theme
	focused bool
	width   int
	height  int
	stats   missionStats
}

type missionStats struct {
	total     int
	delivered int
	lost      int
	topDriver string
}

func newMissionControl(s missionStats) *MissionControl {
	return &MissionControl{stats: s}
}

func (m *MissionControl) Init() tea.Cmd                                { return nil }
func (m *MissionControl) Update(msg tea.Msg) (tuikit.Component, tea.Cmd) { return m, nil }
func (m *MissionControl) KeyBindings() []tuikit.KeyBind                { return nil }
func (m *MissionControl) SetSize(w, h int)                             { m.width = w; m.height = h }
func (m *MissionControl) Focused() bool                                { return m.focused }
func (m *MissionControl) SetFocused(f bool)                            { m.focused = f }
func (m *MissionControl) SetTheme(t tuikit.Theme)                      { m.theme = t }

func (m *MissionControl) View() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Accent)).
		Bold(true)

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Muted))

	value := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Bold(true)

	positive := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Positive))

	negative := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Negative))

	rate := 0.0
	if m.stats.total > 0 {
		rate = float64(m.stats.delivered) / float64(m.stats.total) * 100
	}

	var sb strings.Builder
	sb.WriteString(title.Render("  MISSION CONTROL"))
	sb.WriteString("\n\n")
	sb.WriteString(label.Render("  Total Deliveries: "))
	sb.WriteString(value.Render(fmt.Sprintf("%d", m.stats.total)))
	sb.WriteString("\n")
	sb.WriteString(label.Render("  Delivered:        "))
	sb.WriteString(positive.Render(fmt.Sprintf("%d", m.stats.delivered)))
	sb.WriteString("\n")
	sb.WriteString(label.Render("  Lost in Wormhole: "))
	sb.WriteString(negative.Render(fmt.Sprintf("%d", m.stats.lost)))
	sb.WriteString("\n")
	sb.WriteString(label.Render("  Success Rate:     "))
	sb.WriteString(value.Render(fmt.Sprintf("%.1f%%", rate)))
	sb.WriteString("\n\n")
	sb.WriteString(label.Render("  Top Driver:"))
	sb.WriteString("\n")
	sb.WriteString(value.Render("  " + m.stats.topDriver))
	sb.WriteString("\n\n")
	sb.WriteString(title.Render("  WORST PLANET"))
	sb.WriteString("\n")
	sb.WriteString(negative.Render("  Arrakis"))
	sb.WriteString("\n")
	sb.WriteString(label.Render("  (sand everywhere)"))

	return sb.String()
}

func main() {
	planets := []string{
		"Mars", "Jupiter", "Saturn", "Neptune", "Pluto",
		"Kepler-442b", "Proxima b", "Tatooine", "Arrakis",
		"Gallifrey", "Vulcan", "Krypton", "Ego",
	}
	drivers := []string{
		"Zorp McBlast", "Captain Pepperoni", "Turbo Jenkins",
		"Sally Starfighter", "Buzz Crust", "The Dough Knight",
	}
	statuses := []string{
		"In Transit", "Delivered", "Lost in Wormhole",
		"Dodging Asteroids", "Refueling", "Abducted by Aliens",
	}

	var rows []tuikit.Row
	delivered, lost := 0, 0
	for i := 0; i < 42; i++ {
		planet := planets[rand.Intn(len(planets))]
		driver := drivers[rand.Intn(len(drivers))]
		status := statuses[rand.Intn(len(statuses))]
		eta := fmt.Sprintf("%d", rand.Intn(500)+1)
		tip := fmt.Sprintf("%d.%02d", rand.Intn(1000), rand.Intn(100))

		if status == "Delivered" {
			delivered++
			eta = "0"
		}
		if status == "Lost in Wormhole" {
			lost++
			eta = "999"
		}

		rows = append(rows, tuikit.Row{planet, driver, status, eta, tip})
	}

	columns := []tuikit.Column{
		{Title: "Planet", Width: 20, Sortable: true},
		{Title: "Driver", Width: 25, Sortable: true},
		{Title: "Status", Width: 25},
		{Title: "ETA (light-min)", Width: 15, MinWidth: 100, Align: tuikit.Right, Sortable: true},
		{Title: "Tip ($)", Width: 10, MinWidth: 120, Align: tuikit.Right, Sortable: true},
	}

	// Custom cell renderer: color cells based on context
	cellRenderer := func(row tuikit.Row, colIdx int, isCursor bool, theme tuikit.Theme) string {
		if colIdx >= len(row) {
			return ""
		}
		val := row[colIdx]

		// Status column gets semantic colors
		if colIdx == 2 {
			var style lipgloss.Style
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
			if isCursor {
				style = style.Background(lipgloss.Color(theme.Cursor))
			}
			return style.Render(val)
		}

		// Tip column in green
		if colIdx == 4 {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive))
			if isCursor {
				style = style.Background(lipgloss.Color(theme.Cursor))
			}
			return style.Render("$" + val)
		}

		// Default rendering
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
		if isCursor {
			style = style.Background(lipgloss.Color(theme.Cursor)).
				Foreground(lipgloss.Color(theme.TextInverse))
		}
		return style.Render(val)
	}

	// Custom numeric sort for ETA and Tip columns
	numericSort := func(a, b tuikit.Row, sortCol int, sortAsc bool) bool {
		va, _ := strconv.ParseFloat(a[sortCol], 64)
		vb, _ := strconv.ParseFloat(b[sortCol], 64)
		if sortAsc {
			return va < vb
		}
		return va > vb
	}

	table := tuikit.NewTable(columns, rows, tuikit.TableOpts{
		Sortable:     true,
		Filterable:   true,
		CellRenderer: cellRenderer,
		SortFunc:     numericSort,
	})

	panel := newMissionControl(missionStats{
		total:     42,
		delivered: delivered,
		lost:      lost,
		topDriver: "Captain Pepperoni",
	})

	pineapple := "absolutely not"
	wormholeLevel := "11"
	defaultTip := "15"

	configEditor := tuikit.NewConfigEditor([]tuikit.ConfigField{
		{
			Label: "Pineapple Allowed",
			Group: "Pizza Policy",
			Hint:  "this is a serious matter",
			Get:   func() string { return pineapple },
			Set:   func(v string) error { pineapple = v; return nil },
		},
		{
			Label: "Wormhole Avoidance",
			Group: "Navigation",
			Hint:  "scale of 1-11 (11 = maximum avoidance)",
			Get:   func() string { return wormholeLevel },
			Set:   func(v string) error { wormholeLevel = v; return nil },
		},
		{
			Label: "Default Tip %",
			Group: "Finance",
			Hint:  "be generous, they're crossing galaxies",
			Get:   func() string { return defaultTip },
			Set:   func(v string) error { defaultTip = v; return nil },
		},
	})

	// Filter mode cycling: all -> delivered -> lost -> all
	filterModes := []string{"all", "delivered", "lost"}
	filterIdx := 0

	// Set up predicate filter — closure reads filterIdx
	table.SetFilter(func(row tuikit.Row) bool {
		if len(row) < 3 {
			return true
		}
		switch filterModes[filterIdx] {
		case "delivered":
			return row[2] == "Delivered"
		case "lost":
			return row[2] == "Lost in Wormhole"
		default:
			return true
		}
	})

	var app *tuikit.App
	app = tuikit.NewApp(
		tuikit.WithTheme(tuikit.DefaultTheme()),
		tuikit.WithLayout(&tuikit.DualPane{
			Main:         table,
			Side:         panel,
			SideWidth:    32,
			MinMainWidth: 60,
			SideRight:    true,
			ToggleKey:    "p",
		}),
		tuikit.WithStatusBar(
			func() string {
				return fmt.Sprintf(" ? help  / search  s sort  c config  f filter[%s]  p panel  q quit", filterModes[filterIdx])
			},
			func() string { return "42 active deliveries  Galactic Pizza Corp " },
		),
		tuikit.WithHelp(),
		tuikit.WithOverlay("Settings", "c", configEditor),
		tuikit.WithKeyBind(tuikit.KeyBind{
			Key:   "f",
			Label: "Cycle filter",
			Group: "DATA",
			Handler: func() {
				filterIdx = (filterIdx + 1) % len(filterModes)
				// Re-apply the filter (the closure already reads filterIdx)
				table.SetRows(rows)
			},
		}),
		tuikit.WithKeyBind(tuikit.KeyBind{
			Key:   "i",
			Label: "Info toast",
			Group: "TOAST",
			Handler: func() {
				app.Send(tuikit.ToastMsg{Severity: tuikit.SeverityInfo, Title: "Info", Body: "Mission is a go!", Duration: 4 * time.Second})
			},
		}),
		tuikit.WithKeyBind(tuikit.KeyBind{
			Key:   "s",
			Label: "Success toast",
			Group: "TOAST",
			Handler: func() {
				app.Send(tuikit.ToastMsg{Severity: tuikit.SeveritySuccess, Title: "Delivered!", Body: "Pizza arrived at destination.", Duration: 4 * time.Second})
			},
		}),
		tuikit.WithKeyBind(tuikit.KeyBind{
			Key:   "w",
			Label: "Warning toast",
			Group: "TOAST",
			Handler: func() {
				app.Send(tuikit.ToastMsg{Severity: tuikit.SeverityWarn, Title: "Asteroid Field", Body: "Wormhole instability detected.", Duration: 4 * time.Second})
			},
		}),
		tuikit.WithKeyBind(tuikit.KeyBind{
			Key:   "e",
			Label: "Error toast",
			Group: "TOAST",
			Handler: func() {
				app.Send(tuikit.ToastMsg{Severity: tuikit.SeverityError, Title: "Delivery Failed", Body: "Lost in wormhole. No pizza.", Duration: 4 * time.Second})
			},
		}),
		tuikit.WithMouseSupport(),
		tuikit.WithTickInterval(100*time.Millisecond),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
