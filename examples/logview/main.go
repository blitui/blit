// Package main demonstrates the LogViewer component streaming fake logs at 50 Hz.
//
// Fake log lines are generated from a background goroutine across four levels
// (debug, info, warn, error). Use the keybindings below to interact:
//
//	p       — pause / resume auto-scroll
//	c       — clear all log lines
//	/       — enter substring filter (Enter/Esc to confirm/cancel)
//	l       — cycle minimum level filter (debug+ → info+ → warn+ → error)
//	end     — jump to latest (also resumes auto-scroll)
//	up/k    — scroll up (pauses auto-scroll)
//	down/j  — scroll down
//	q       — quit
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// tickMsg drives the fake log generator.
type tickMsg struct{}

// model is the root Bubble Tea model.
type model struct {
	lv    *tuikit.LogViewer
	width int
	height int
}

func newModel() model {
	lv := tuikit.NewLogViewer()
	lv.SetTheme(tuikit.DefaultTheme())
	return model{lv: lv}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		m.lv.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.lv.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		comp, cmd := m.lv.Update(msg)
		m.lv = comp.(*tuikit.LogViewer)
		return m, cmd

	case tickMsg:
		appendFakeLog(m.lv)
		return m, tick()
	}

	comp, cmd := m.lv.Update(msg)
	m.lv = comp.(*tuikit.LogViewer)
	return m, cmd
}

func (m model) View() string {
	return m.lv.View()
}

// tick fires at ~50 Hz (20 ms interval).
func tick() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

var sources = []string{"auth", "db", "api", "cache", "queue", "scheduler", "gateway"}

type msgTemplate struct {
	text    string
	hasVerb bool
}

var msgTemplates = []msgTemplate{
	{"connection established", false},
	{"request received", false},
	{"cache miss, fetching from db", false},
	{"query executed in %dms", true},
	{"token validated", false},
	{"rate limit checked", false},
	{"background job started", false},
	{"retry attempt %d", true},
	{"record not found", false},
	{"response sent", false},
	{"connection pool full", false},
	{"slow query detected (%dms)", true},
	{"panic recovered", false},
	{"health check passed", false},
}

func appendFakeLog(lv *tuikit.LogViewer) {
	levels := []tuikit.LogLevel{
		tuikit.LogDebug,
		tuikit.LogInfo,
		tuikit.LogInfo,
		tuikit.LogInfo,
		tuikit.LogWarn,
		tuikit.LogError,
	}
	level := levels[rand.Intn(len(levels))]
	src := sources[rand.Intn(len(sources))]
	tmpl := msgTemplates[rand.Intn(len(msgTemplates))]
	var msg string
	if tmpl.hasVerb {
		msg = fmt.Sprintf(tmpl.text, rand.Intn(500)+1)
	} else {
		msg = tmpl.text
	}

	lv.Append(tuikit.LogLine{
		Level:     level,
		Timestamp: time.Now(),
		Message:   msg,
		Source:    src,
	})
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
