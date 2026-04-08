package tuitest

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModel wraps a tea.Model for easy testing.
type TestModel struct {
	t     testing.TB
	model tea.Model
	cols  int
	lines int
}

// NewTestModel creates a test wrapper around a Bubble Tea model.
// Calls Init() and processes any returned commands that produce messages.
func NewTestModel(t testing.TB, model tea.Model, cols, lines int) *TestModel {
	t.Helper()
	cmd := model.Init()
	if cmd != nil {
		if msg := cmd(); msg != nil {
			model, _ = model.Update(msg)
		}
	}
	// Send an initial WindowSizeMsg so the model knows its dimensions.
	model, _ = model.Update(tea.WindowSizeMsg{Width: cols, Height: lines})
	return &TestModel{
		t:     t,
		model: model,
		cols:  cols,
		lines: lines,
	}
}

// SendKey sends a key event to the model. Supports special key names like
// "enter", "tab", "up", "down", "left", "right", "esc", "backspace",
// "space", and single characters.
func (tm *TestModel) SendKey(key string) {
	tm.t.Helper()
	msg := keyToMsg(key)
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// SendMouse sends a mouse event to the model.
func (tm *TestModel) SendMouse(x, y int, button tea.MouseButton) {
	tm.t.Helper()
	msg := tea.MouseMsg{X: x, Y: y, Button: button}
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// SendResize sends a window resize event.
func (tm *TestModel) SendResize(cols, lines int) {
	tm.t.Helper()
	tm.cols = cols
	tm.lines = lines
	msg := tea.WindowSizeMsg{Width: cols, Height: lines}
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// Type sends a sequence of key events (one per character).
func (tm *TestModel) Type(text string) {
	tm.t.Helper()
	for _, ch := range text {
		tm.SendKey(string(ch))
	}
}

// Screen renders the model's View() and returns the parsed screen.
func (tm *TestModel) Screen() *Screen {
	s := NewScreen(tm.cols, tm.lines)
	s.Render(tm.model.View())
	return s
}

// Model returns the current underlying tea.Model.
func (tm *TestModel) Model() tea.Model {
	return tm.model
}

// processCmd executes a command and feeds its message back into the model.
// Only processes a single level to avoid infinite loops.
func (tm *TestModel) processCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	tm.model, _ = tm.model.Update(msg)
}

// keyToMsg converts a key name string to a tea.KeyMsg.
func keyToMsg(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "esc", "escape":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		runes := []rune(key)
		if len(runes) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
	}
}
