package blit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestCommandBar() *CommandBar {
	return NewCommandBar([]Command{
		{Name: "quit", Aliases: []string{"q"}, Hint: "Quit the app", Run: func(args string) tea.Cmd { return tea.Quit }},
		{Name: "sort", Args: true, Hint: "Sort by column", Run: func(args string) tea.Cmd { return nil }},
		{Name: "go", Aliases: []string{"find"}, Args: true, Hint: "Jump to symbol", Run: func(args string) tea.Cmd { return nil }},
	})
}

func TestCommandBarInitInactive(t *testing.T) {
	cb := newTestCommandBar()
	if cb.IsActive() {
		t.Error("command bar should start inactive")
	}
}

func TestCommandBarActivateDeactivate(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	if !cb.IsActive() {
		t.Error("command bar should be active after SetActive(true)")
	}
	cb.Close()
	if cb.IsActive() {
		t.Error("command bar should be inactive after Close()")
	}
}

func TestCommandBarMatchByName(t *testing.T) {
	var called string
	cb := NewCommandBar([]Command{
		{Name: "sort", Args: true, Run: func(args string) tea.Cmd {
			called = args
			return nil
		}},
	})
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	// Type "sort price"
	for _, r := range "sort price" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}
	cb.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})

	if called != "price" {
		t.Errorf("expected args 'price', got '%s'", called)
	}
}

func TestCommandBarMatchByAlias(t *testing.T) {
	quitCalled := false
	cb := NewCommandBar([]Command{
		{Name: "quit", Aliases: []string{"q"}, Run: func(args string) tea.Cmd {
			quitCalled = true
			return nil
		}},
	})
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, Context{})
	cb.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})

	if !quitCalled {
		t.Error("quit command should have been called via alias")
	}
}

func TestCommandBarUnknownCommand(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	for _, r := range "bogus" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}
	cb.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})

	if cb.errMsg == "" {
		t.Error("should show error for unknown command")
	}
	if !cb.IsActive() {
		t.Error("command bar should stay open on error")
	}
}

func TestCommandBarEscCloses(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	for _, r := range "sort" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}
	cb.Update(tea.KeyMsg{Type: tea.KeyEscape}, Context{})

	if cb.IsActive() {
		t.Error("esc should close the command bar")
	}
	if cb.input != "" {
		t.Error("input should be cleared on close")
	}
}

func TestCommandBarBackspaceOnEmptyCloses(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	cb.Update(tea.KeyMsg{Type: tea.KeyBackspace}, Context{})

	if cb.IsActive() {
		t.Error("backspace on empty should close the command bar")
	}
}

func TestCommandBarBackspaceDeletesChar(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, Context{})
	cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}, Context{})
	cb.Update(tea.KeyMsg{Type: tea.KeyBackspace}, Context{})

	if cb.input != "a" {
		t.Errorf("expected 'a' after backspace, got '%s'", cb.input)
	}
}

func TestCommandBarView(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	for _, r := range "sort" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}

	view := cb.View()
	if !strings.Contains(view, ":") {
		t.Error("view should contain ':' prompt")
	}
	if !strings.Contains(view, "sort") {
		t.Error("view should contain typed input")
	}
}

func TestCommandBarKeyBindingsForHelp(t *testing.T) {
	cb := newTestCommandBar()
	bindings := cb.KeyBindings()

	found := false
	for _, kb := range bindings {
		if kb.Group == "COMMANDS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("command bar should register commands in COMMANDS group")
	}
}

func TestCommandBarCaseInsensitive(t *testing.T) {
	called := false
	cb := NewCommandBar([]Command{
		{Name: "quit", Run: func(args string) tea.Cmd {
			called = true
			return nil
		}},
	})
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	for _, r := range "QUIT" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}
	cb.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})

	if !called {
		t.Error("command matching should be case-insensitive")
	}
}

func TestCommandBarTabCompletion(t *testing.T) {
	cb := newTestCommandBar()
	cb.SetActive(true)
	cb.SetTheme(DefaultTheme())
	cb.SetSize(80, 24)

	// Type "so" then tab — should complete to "sort"
	for _, r := range "so" {
		cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, Context{})
	}
	cb.Update(tea.KeyMsg{Type: tea.KeyTab}, Context{})

	if cb.input != "sort" {
		t.Errorf("expected tab completion to 'sort', got '%s'", cb.input)
	}
}
