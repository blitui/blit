// Package main_test demonstrates tuitest session tests for myapp.
//
// These tests use the tuitest.Harness fluent API to drive a tea.Model
// programmatically and assert on rendered screen output — no real terminal
// or subprocess required.
package main_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

// listModel is a minimal tea.Model that renders a navigable list.
// Replace this with your actual app model in real tests.
type listModel struct {
	items  []string
	cursor int
	quit   bool
}

func newListModel() *listModel {
	return &listModel{
		items: []string{
			"Item One",
			"Item Two",
			"Item Three",
			"Item Four",
			"Item Five",
		},
	}
}

func (m *listModel) Init() tea.Cmd { return nil }

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *listModel) View() string {
	if m.quit {
		return ""
	}
	var sb strings.Builder
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, item))
	}
	return sb.String()
}

// TestNavigation verifies that pressing down moves the cursor.
func TestNavigation(t *testing.T) {
	tuitest.NewHarness(t, newListModel(), 80, 24).
		Expect("Item One").
		Keys("down").
		Expect("Item Two").
		Keys("down").
		Expect("Item Three").
		Done()
}

// TestNavigateUp verifies that pressing up after down returns to the first item.
func TestNavigateUp(t *testing.T) {
	tuitest.NewHarness(t, newListModel(), 80, 24).
		Keys("down").
		Keys("up").
		Expect("> Item One").
		Done()
}

// TestQuit verifies that pressing q triggers a clean exit.
func TestQuit(t *testing.T) {
	tuitest.NewHarness(t, newListModel(), 80, 24).
		Keys("q").
		Done()
}
