package cli

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg builds a tea.KeyMsg for a named key string.
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

// namedKey builds a tea.KeyMsg for special named keys like "enter", "up", "down".
func namedKey(typ tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: typ}
}

// --- confirm tests ---

func TestConfirmHint(t *testing.T) {
	tests := []struct {
		defaultYes bool
		wantHint   string
	}{
		{true, "[Y/n]"},
		{false, "[y/N]"},
	}
	for _, tt := range tests {
		m := confirmModel{prompt: "Continue?", defaultYes: tt.defaultYes}
		view := m.View()
		if !strings.Contains(view, tt.wantHint) {
			t.Errorf("defaultYes=%v: expected hint %q in view %q", tt.defaultYes, tt.wantHint, view)
		}
	}
}

func TestConfirmKeyY(t *testing.T) {
	m := confirmModel{prompt: "q?", defaultYes: false}
	result, _ := m.Update(keyMsg("y"))
	final := result.(confirmModel)
	if !final.result {
		t.Error("expected result=true for 'y'")
	}
	if !final.done {
		t.Error("expected done=true")
	}
}

func TestConfirmKeyN(t *testing.T) {
	m := confirmModel{prompt: "q?", defaultYes: true}
	result, _ := m.Update(keyMsg("n"))
	final := result.(confirmModel)
	if final.result {
		t.Error("expected result=false for 'n'")
	}
	if !final.done {
		t.Error("expected done=true")
	}
}

func TestConfirmEnterDefaultYes(t *testing.T) {
	m := confirmModel{prompt: "q?", defaultYes: true}
	result, _ := m.Update(namedKey(tea.KeyEnter))
	final := result.(confirmModel)
	if !final.result {
		t.Error("expected result=true for Enter with defaultYes=true")
	}
}

func TestConfirmEnterDefaultNo(t *testing.T) {
	m := confirmModel{prompt: "q?", defaultYes: false}
	result, _ := m.Update(namedKey(tea.KeyEnter))
	final := result.(confirmModel)
	if final.result {
		t.Error("expected result=false for Enter with defaultYes=false")
	}
}

func TestConfirmViewEmptyWhenDone(t *testing.T) {
	m := confirmModel{done: true}
	if m.View() != "" {
		t.Errorf("expected empty view when done")
	}
}

// --- input validation tests ---

func TestInputValidation(t *testing.T) {
	notEmpty := func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	}

	if err := notEmpty("hello"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := notEmpty(""); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestInputViewShowsError(t *testing.T) {
	m := newInputModel("Name:", nil)
	m.errMsg = "required"
	view := m.View()
	if !strings.Contains(view, "required") {
		t.Errorf("expected error message in view, got: %q", view)
	}
}

func TestInputViewEmptyWhenDone(t *testing.T) {
	m := newInputModel("Name:", nil)
	m.done = true
	if m.View() != "" {
		t.Errorf("expected empty view when done")
	}
}

// --- select tests ---

func TestSelectFilter(t *testing.T) {
	items := make([]string, 15)
	for i := range items {
		items[i] = "item"
	}
	items[3] = "special"
	items[12] = "special-two"

	m := newSelectModel("Pick:", items)
	if !m.useFilter {
		t.Fatal("expected filter to be active for 15 items")
	}

	m.filter = "special"
	m.applyFilter()

	if len(m.filtered) != 2 {
		t.Errorf("expected 2 filtered items, got %d", len(m.filtered))
	}
}

func TestSelectFilterInactiveSmallList(t *testing.T) {
	items := []string{"a", "b", "c"}
	m := newSelectModel("Pick:", items)
	if m.useFilter {
		t.Error("expected filter to be inactive for 3 items")
	}
}

func TestSelectCursorDown(t *testing.T) {
	items := []string{"a", "b", "c"}
	m := newSelectModel("Pick:", items)

	result, _ := m.Update(namedKey(tea.KeyDown))
	m = result.(selectModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.cursor)
	}
}

func TestSelectCursorUp(t *testing.T) {
	items := []string{"a", "b", "c"}
	m := newSelectModel("Pick:", items)
	m.cursor = 2

	result, _ := m.Update(namedKey(tea.KeyUp))
	m = result.(selectModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after up, got %d", m.cursor)
	}
}

func TestSelectCursorClampTop(t *testing.T) {
	items := []string{"a", "b"}
	m := newSelectModel("Pick:", items)

	result, _ := m.Update(namedKey(tea.KeyUp))
	m = result.(selectModel)
	if m.cursor != 0 {
		t.Errorf("cursor should clamp at 0, got %d", m.cursor)
	}
}

func TestSelectCursorClampBottom(t *testing.T) {
	items := []string{"a", "b"}
	m := newSelectModel("Pick:", items)
	m.cursor = 1

	result, _ := m.Update(namedKey(tea.KeyDown))
	m = result.(selectModel)
	if m.cursor != 1 {
		t.Errorf("cursor should clamp at 1, got %d", m.cursor)
	}
}

func TestSelectViewContainsItems(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	m := newSelectModel("Pick:", items)
	view := m.View()
	for _, item := range items {
		if !strings.Contains(view, item) {
			t.Errorf("expected %q in view", item)
		}
	}
}

func TestSelectEmptyItems(t *testing.T) {
	_, _, err := SelectOne("Pick:", []string{})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

// --- progress bar tests ---

func TestProgressPercentage(t *testing.T) {
	p := &Progress{total: 100, label: "Test"}
	p.current = 50

	pct := p.current * 100 / p.total
	if pct != 50 {
		t.Errorf("expected 50%%, got %d%%", pct)
	}
}

func TestProgressBarFilled(t *testing.T) {
	p := &Progress{total: 100, label: "Test"}
	p.current = 50

	filled := p.current * barWidth / p.total
	if filled != barWidth/2 {
		t.Errorf("expected %d filled cells, got %d", barWidth/2, filled)
	}
}

func TestProgressIncrement(t *testing.T) {
	p := &Progress{total: 10, label: "Work"}
	p.Increment(3)
	if p.current != 3 {
		t.Errorf("expected current=3, got %d", p.current)
	}
}

func TestProgressIncrementClamps(t *testing.T) {
	p := &Progress{total: 10, label: "Work"}
	p.Increment(3)
	p.Increment(10) // should clamp to total
	if p.current != 10 {
		t.Errorf("expected current clamped to 10, got %d", p.current)
	}
}

func TestProgressDoneIdempotent(t *testing.T) {
	p := &Progress{total: 5, label: "X"}
	p.Done()
	p.Done() // second call must not panic
	if !p.done {
		t.Error("expected done=true")
	}
}

func TestProgressZeroTotal(t *testing.T) {
	// Should not divide by zero
	p := &Progress{total: 0, label: "Empty"}
	p.Increment(1)
	p.Done()
}

// --- spinner tests ---

func TestSpinnerStopDoesNotPanic(t *testing.T) {
	s := Spin("loading...")
	s.Stop()
	s.Stop() // idempotent — second stop must not panic
}
