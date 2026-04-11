// Package main demonstrates a full-featured blit service monitor dashboard.
//
// This example showcases: Table with live-updating data, DualPane layout,
// CollapsibleSection sidebar, DetailOverlay, CommandBar, ConfigEditor,
// StatusBar, Help, Notifications, custom cell rendering, and keybindings.
//
// Theme: "Fleet Monitor" — track 30 microservices across regions with
// health, latency, memory, and error rates. Data updates every tick.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Service data ──────────────────────────────────────────────────────────

type service struct {
	Name    string
	Region  string
	Status  string
	Latency int // ms
	Memory  int // MB
	Errors  int
	Uptime  string
}

var (
	serviceNames = []string{
		"api-gateway", "auth-svc", "user-svc", "billing-svc",
		"notification-svc", "search-svc", "analytics-svc", "cdn-edge",
		"cache-layer", "db-proxy", "queue-worker", "scheduler",
		"ml-inference", "logging-svc", "metrics-svc", "config-svc",
		"payment-svc", "email-svc", "storage-svc", "webhook-svc",
		"rate-limiter", "load-balancer", "dns-resolver", "cert-manager",
		"secret-vault", "audit-trail", "feature-flags", "ab-testing",
		"health-check", "canary-deploy",
	}
	regions  = []string{"us-east-1", "us-west-2", "eu-west-1", "ap-south-1"}
	statuses = []string{"healthy", "healthy", "healthy", "healthy", "degraded", "critical"}
	uptimes  = []string{"45d", "12d", "89d", "3d", "156d", "7d", "23d", "67d"}
)

func generateServices() []service {
	svcs := make([]service, len(serviceNames))
	for i, name := range serviceNames {
		status := statuses[rand.Intn(len(statuses))]
		latency := 5 + rand.Intn(200)
		memory := 64 + rand.Intn(1024)
		errors := 0
		if status == "degraded" {
			latency += 300
			errors = rand.Intn(50)
		}
		if status == "critical" {
			latency += 800
			errors = 50 + rand.Intn(200)
			memory += 512
		}
		svcs[i] = service{
			Name:    name,
			Region:  regions[rand.Intn(len(regions))],
			Status:  status,
			Latency: latency,
			Memory:  memory,
			Errors:  errors,
			Uptime:  uptimes[rand.Intn(len(uptimes))],
		}
	}
	return svcs
}

func servicesToRows(svcs []service) []blit.Row {
	rows := make([]blit.Row, len(svcs))
	for i, s := range svcs {
		rows[i] = blit.Row{
			s.Name, s.Region, s.Status,
			fmt.Sprintf("%d", s.Latency),
			fmt.Sprintf("%d", s.Memory),
			fmt.Sprintf("%d", s.Errors),
			s.Uptime,
		}
	}
	return rows
}

// ── Sidebar panel ─────────────────────────────────────────────────────────

// FleetPanel shows aggregate fleet stats in collapsible sections.
type FleetPanel struct {
	theme    blit.Theme
	focused  bool
	width    int
	height   int
	services []service

	sectionHealth bool
	sectionRegion bool
	sectionAlerts bool
}

func newFleetPanel(svcs []service) *FleetPanel {
	return &FleetPanel{
		services:      svcs,
		sectionHealth: true,
		sectionRegion: true,
		sectionAlerts: true,
	}
}

func (p *FleetPanel) Init() tea.Cmd               { return nil }
func (p *FleetPanel) KeyBindings() []blit.KeyBind { return nil }
func (p *FleetPanel) SetSize(w, h int)            { p.width = w; p.height = h }
func (p *FleetPanel) Focused() bool               { return p.focused }
func (p *FleetPanel) SetFocused(f bool)           { p.focused = f }
func (p *FleetPanel) SetTheme(t blit.Theme)       { p.theme = t }
func (p *FleetPanel) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) {
	return p, nil
}
func (p *FleetPanel) UpdateServices(svcs []service) { p.services = svcs }

func (p *FleetPanel) View() string {
	title := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Accent)).Bold(true)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Muted))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Text)).Bold(true)
	positive := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Positive))
	negative := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Negative))
	flash := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Flash))

	healthy, degraded, critical := 0, 0, 0
	totalLatency, totalErrors, totalMem := 0, 0, 0
	regionCounts := map[string]int{}
	for _, s := range p.services {
		switch s.Status {
		case "healthy":
			healthy++
		case "degraded":
			degraded++
		case "critical":
			critical++
		}
		totalLatency += s.Latency
		totalErrors += s.Errors
		totalMem += s.Memory
		regionCounts[s.Region]++
	}
	avgLatency := 0
	if len(p.services) > 0 {
		avgLatency = totalLatency / len(p.services)
	}

	var sb strings.Builder
	sb.WriteString(title.Render("  FLEET OVERVIEW"))
	sb.WriteString("\n\n")

	// Health section
	healthSection := blit.CollapsibleSection{
		Title:     "HEALTH",
		Collapsed: !p.sectionHealth,
	}
	sb.WriteString(healthSection.Render(p.theme, func() string {
		var content strings.Builder
		content.WriteString(label.Render("  Healthy:  "))
		content.WriteString(positive.Render(fmt.Sprintf("%-4d", healthy)))
		content.WriteString("\n")
		content.WriteString(label.Render("  Degraded: "))
		content.WriteString(flash.Render(fmt.Sprintf("%-4d", degraded)))
		content.WriteString("\n")
		content.WriteString(label.Render("  Critical: "))
		content.WriteString(negative.Render(fmt.Sprintf("%-4d", critical)))
		content.WriteString("\n")
		content.WriteString(label.Render("  Avg Lat:  "))
		latStyle := positive
		if avgLatency > 200 {
			latStyle = flash
		}
		if avgLatency > 500 {
			latStyle = negative
		}
		content.WriteString(latStyle.Render(fmt.Sprintf("%dms", avgLatency)))
		return content.String()
	}))
	sb.WriteString("\n\n")

	// Region section
	regionSection := blit.CollapsibleSection{
		Title:     "REGIONS",
		Collapsed: !p.sectionRegion,
	}
	sb.WriteString(regionSection.Render(p.theme, func() string {
		var content strings.Builder
		sortedRegions := make([]string, 0, len(regionCounts))
		for r := range regionCounts {
			sortedRegions = append(sortedRegions, r)
		}
		sort.Strings(sortedRegions)
		for _, r := range sortedRegions {
			content.WriteString(label.Render(fmt.Sprintf("  %-12s", r)))
			content.WriteString(value.Render(fmt.Sprintf("%d svcs", regionCounts[r])))
			content.WriteString("\n")
		}
		return content.String()
	}))
	sb.WriteString("\n\n")

	// Alerts section
	alertSection := blit.CollapsibleSection{
		Title:     "ALERTS",
		Collapsed: !p.sectionAlerts,
	}
	sb.WriteString(alertSection.Render(p.theme, func() string {
		var content strings.Builder
		if totalErrors == 0 {
			content.WriteString(positive.Render("  No active alerts"))
		} else {
			content.WriteString(negative.Render(fmt.Sprintf("  %d total errors", totalErrors)))
			content.WriteString("\n")
			for _, s := range p.services {
				if s.Errors > 20 {
					content.WriteString(negative.Render(fmt.Sprintf("  ⚠ %s: %d err", s.Name, s.Errors)))
					content.WriteString("\n")
				}
			}
		}
		return content.String()
	}))
	sb.WriteString("\n\n")

	// Resource summary
	sb.WriteString(label.Render("  ─────────────────────\n"))
	sb.WriteString(label.Render("  Total Memory: "))
	sb.WriteString(value.Render(fmt.Sprintf("%.1f GB", float64(totalMem)/1024.0)))
	sb.WriteString("\n")
	sb.WriteString(label.Render("  Services:     "))
	sb.WriteString(value.Render(fmt.Sprintf("%d", len(p.services))))
	sb.WriteString("\n")

	return sb.String()
}

// ── Main ──────────────────────────────────────────────────────────────────

func main() {
	services := generateServices()
	rows := servicesToRows(services)

	columns := []blit.Column{
		{Title: "Service", Width: 20, Sortable: true},
		{Title: "Region", Width: 14, Sortable: true},
		{Title: "Status", Width: 12, Sortable: true},
		{Title: "Latency", Width: 10, MinWidth: 90, Align: blit.Right, Sortable: true},
		{Title: "Memory", Width: 10, MinWidth: 100, Align: blit.Right, Sortable: true},
		{Title: "Errors", Width: 8, MinWidth: 110, Align: blit.Right, Sortable: true},
		{Title: "Uptime", Width: 8, MinWidth: 120, Align: blit.Right, Sortable: true},
	}

	cellRenderer := func(row blit.Row, colIdx int, isCursor bool, theme blit.Theme) string {
		if colIdx >= len(row) {
			return ""
		}
		val := row[colIdx]
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
		cursorBg := lipgloss.NewStyle().Background(lipgloss.Color(theme.Cursor)).Foreground(lipgloss.Color(theme.TextInverse))

		style := base
		if isCursor {
			style = cursorBg
		}

		switch colIdx {
		case 0: // Service name — bold
			style = style.Bold(true)
		case 2: // Status — semantic color
			switch val {
			case "healthy":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive))
				val = "● " + val
			case "degraded":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Flash))
				val = "◐ " + val
			case "critical":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative)).Bold(true)
				val = "✗ " + val
			}
			if isCursor {
				style = style.Background(lipgloss.Color(theme.Cursor))
			}
		case 3: // Latency — threshold coloring
			ms, _ := strconv.Atoi(val)
			if ms > 500 {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative))
			} else if ms > 200 {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Flash))
			}
			if isCursor {
				style = style.Background(lipgloss.Color(theme.Cursor))
			}
			val += "ms"
		case 4: // Memory
			val += "MB"
		case 5: // Errors — red if > 0
			n, _ := strconv.Atoi(val)
			if n > 0 {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative))
			} else {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
			}
			if isCursor {
				style = style.Background(lipgloss.Color(theme.Cursor))
			}
		}

		return style.Render(val)
	}

	numericSort := func(a, b blit.Row, col int, asc bool) bool {
		va, _ := strconv.ParseFloat(a[col], 64)
		vb, _ := strconv.ParseFloat(b[col], 64)
		if asc {
			return va < vb
		}
		return va > vb
	}

	table := blit.NewTable(columns, rows, blit.TableOpts{
		Sortable:     true,
		Filterable:   true,
		CellRenderer: cellRenderer,
		SortFunc:     numericSort,
	})

	panel := newFleetPanel(services)

	// Detail overlay for inspecting a service
	detail := blit.NewDetailOverlay(blit.DetailOverlayOpts[blit.Row]{
		Render: func(row blit.Row, width, height int, theme blit.Theme) string {
			if len(row) < 7 {
				return ""
			}
			title := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted)).Width(14)
			val := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text)).Bold(true)

			var sb strings.Builder
			sb.WriteString(title.Render(fmt.Sprintf("  SERVICE: %s", row[0])))
			sb.WriteString("\n\n")
			pairs := [][2]string{
				{"Region", row[1]},
				{"Status", row[2]},
				{"Latency", row[3] + "ms"},
				{"Memory", row[4] + "MB"},
				{"Errors", row[5]},
				{"Uptime", row[6]},
			}
			for _, p := range pairs {
				sb.WriteString("  ")
				sb.WriteString(label.Render(p[0] + ":"))
				sb.WriteString(val.Render(p[1]))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted)).Render("  Press esc to close"))
			return sb.String()
		},
	})

	// Config editor
	refreshRate := "1s"
	alertThreshold := "50"
	configEditor := blit.NewConfigEditor([]blit.ConfigField{
		{
			Label: "Refresh Rate",
			Group: "General",
			Hint:  "how often to update service data",
			Get:   func() string { return refreshRate },
			Set:   func(v string) error { refreshRate = v; return nil },
		},
		{
			Label: "Alert Threshold",
			Group: "Alerts",
			Hint:  "error count to trigger alert",
			Get:   func() string { return alertThreshold },
			Set:   func(v string) error { alertThreshold = v; return nil },
		},
	})

	// Command bar
	commandBar := blit.NewCommandBar([]blit.Command{
		{Name: "sort", Args: true, Hint: "sort <column>", Run: func(args string) tea.Cmd { return nil }},
		{Name: "filter", Args: true, Hint: "filter <status>", Run: func(args string) tea.Cmd { return nil }},
		{Name: "region", Args: true, Hint: "region <name>", Run: func(args string) tea.Cmd { return nil }},
	})

	// Filter state
	filterModes := []string{"all", "healthy", "degraded", "critical"}
	filterIdx := 0
	table.SetFilter(func(row blit.Row) bool {
		if len(row) < 3 {
			return true
		}
		mode := filterModes[filterIdx]
		if mode == "all" {
			return true
		}
		return row[2] == mode
	})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithLayout(&blit.DualPane{
			Main:         table,
			Side:         panel,
			SideWidth:    30,
			MinMainWidth: 60,
			SideRight:    true,
			ToggleKey:    "p",
		}),
		blit.WithStatusBar(
			func() string {
				return fmt.Sprintf(" ? help  / search  s sort  c config  : cmd  f filter[%s]  d detail  p panel  q quit",
					filterModes[filterIdx])
			},
			func() string {
				h, d, c := 0, 0, 0
				for _, s := range services {
					switch s.Status {
					case "healthy":
						h++
					case "degraded":
						d++
					case "critical":
						c++
					}
				}
				return fmt.Sprintf(" %d healthy  %d degraded  %d critical ", h, d, c)
			},
		),
		blit.WithHelp(),
		blit.WithOverlay("Settings", "c", configEditor),
		blit.WithOverlay("Command", ":", commandBar),
		blit.WithOverlay("Detail", "d", detail),
		blit.WithKeyBind(blit.KeyBind{
			Key:   "f",
			Label: "Cycle filter",
			Group: "DATA",
			Handler: func() {
				filterIdx = (filterIdx + 1) % len(filterModes)
				table.SetRows(rows)
			},
		}),
		blit.WithKeyBind(blit.KeyBind{
			Key:   "r",
			Label: "Refresh data",
			Group: "DATA",
			Handler: func() {
				services = generateServices()
				rows = servicesToRows(services)
				table.SetRows(rows)
				panel.UpdateServices(services)
			},
		}),
		blit.WithMouseSupport(),
		blit.WithTickInterval(100*time.Millisecond),
	)

	if _, err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
