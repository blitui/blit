package tuikit

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigEditorNavigation(t *testing.T) {
	value := "hello"
	fields := []ConfigField{
		{Label: "Field1", Group: "General", Get: func() string { return value }},
		{Label: "Field2", Group: "General", Get: func() string { return "world" }},
	}
	ce := NewConfigEditor(fields)
	ce.SetTheme(DefaultTheme())
	ce.SetSize(80, 24)
	ce.active = true

	if ce.cursor != 0 {
		t.Errorf("initial cursor should be 0, got %d", ce.cursor)
	}

	ce.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ce.cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", ce.cursor)
	}

	ce.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ce.cursor != 1 {
		t.Errorf("cursor should clamp to 1, got %d", ce.cursor)
	}

	ce.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ce.cursor != 0 {
		t.Errorf("cursor should be 0 after up, got %d", ce.cursor)
	}
}

func TestConfigEditorEdit(t *testing.T) {
	value := "old"
	fields := []ConfigField{
		{
			Label: "Name",
			Group: "General",
			Get:   func() string { return value },
			Set: func(v string) error {
				value = v
				return nil
			},
		},
	}
	ce := NewConfigEditor(fields)
	ce.SetTheme(DefaultTheme())
	ce.SetSize(80, 24)
	ce.active = true

	// Enter edit mode
	ce.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !ce.editing {
		t.Error("should be in edit mode after Enter")
	}

	// Clear the buffer (it starts with "old"), then type "new"
	// Backspace 3 times to clear "old"
	ce.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ce.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ce.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	for _, ch := range "new" {
		ce.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	ce.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if value != "new" {
		t.Errorf("expected value 'new', got '%s'", value)
	}
	if ce.editing {
		t.Error("should exit edit mode after Enter")
	}
}

func TestConfigEditorValidation(t *testing.T) {
	fields := []ConfigField{
		{
			Label: "Count",
			Group: "General",
			Get:   func() string { return "5" },
			Set: func(v string) error {
				return fmt.Errorf("invalid value")
			},
		},
	}
	ce := NewConfigEditor(fields)
	ce.SetTheme(DefaultTheme())
	ce.SetSize(80, 24)
	ce.active = true

	ce.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ce.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	ce.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ce.errMsg == "" {
		t.Error("should have error message after validation failure")
	}
}

func TestConfigEditorView(t *testing.T) {
	fields := []ConfigField{
		{Label: "Name", Group: "General", Hint: "your name", Get: func() string { return "Alice" }},
		{Label: "Color", Group: "Display", Get: func() string { return "blue" }},
	}
	ce := NewConfigEditor(fields)
	ce.SetTheme(DefaultTheme())
	ce.SetSize(80, 24)
	ce.active = true

	view := ce.View()
	if !strings.Contains(view, "General") {
		t.Error("view should show 'General' group")
	}
	if !strings.Contains(view, "Display") {
		t.Error("view should show 'Display' group")
	}
	if !strings.Contains(view, "Name") {
		t.Error("view should show 'Name' field")
	}
	if !strings.Contains(view, "Alice") {
		t.Error("view should show current value 'Alice'")
	}
}
