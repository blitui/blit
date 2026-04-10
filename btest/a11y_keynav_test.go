package btest_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/blitui/blit/btest"
)

// tabbableModel has three focus states that cycle on tab.
type tabbableModel struct {
	focus int
	items int
}

func newTabbableModel(items int) *tabbableModel {
	return &tabbableModel{items: items}
}

func (m *tabbableModel) Init() tea.Cmd { return nil }
func (m *tabbableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "tab":
			m.focus = (m.focus + 1) % m.items
		case "shift+tab":
			m.focus = (m.focus - 1 + m.items) % m.items
		}
	}
	return m, nil
}
func (m *tabbableModel) View() string {
	return fmt.Sprintf("focus=%d/%d", m.focus, m.items)
}

// staticModel never changes on tab — represents a non-interactive component.
type staticModel struct{}

func (m *staticModel) Init() tea.Cmd                           { return nil }
func (m *staticModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *staticModel) View() string                            { return "static" }

func TestCheckKeyNav_CycleDetected(t *testing.T) {
	result := btest.CheckKeyNav(t, newTabbableModel(3), btest.KeyNavConfig{})
	if result.CycleLength != 3 {
		t.Errorf("CycleLength = %d, want 3", result.CycleLength)
	}
	if result.UniqueScreens != 3 {
		t.Errorf("UniqueScreens = %d, want 3", result.UniqueScreens)
	}
	if result.FocusChanges != 3 {
		t.Errorf("FocusChanges = %d, want 3", result.FocusChanges)
	}
}

func TestCheckKeyNav_Static(t *testing.T) {
	// A static model's screen never changes, so the cycle is trivially 1
	// (same screen on every tab press) with zero visible focus changes.
	result := btest.CheckKeyNav(t, &staticModel{}, btest.KeyNavConfig{MaxTabs: 10})
	if result.CycleLength != 1 {
		t.Errorf("CycleLength = %d, want 1 for static model (trivial cycle)", result.CycleLength)
	}
	if result.FocusChanges != 0 {
		t.Errorf("FocusChanges = %d, want 0 for static model", result.FocusChanges)
	}
	if result.UniqueScreens != 1 {
		t.Errorf("UniqueScreens = %d, want 1", result.UniqueScreens)
	}
}

func TestCheckKeyNav_SingleItem(t *testing.T) {
	// A model with 1 item cycles back to itself immediately.
	result := btest.CheckKeyNav(t, newTabbableModel(1), btest.KeyNavConfig{})
	// focus stays at 0/1 after tab (0+1 mod 1 = 0), so cycle is 1.
	if result.CycleLength != 1 {
		t.Errorf("CycleLength = %d, want 1", result.CycleLength)
	}
}

func TestCheckKeyNavRoundTrip_Pass(t *testing.T) {
	ok := btest.CheckKeyNavRoundTrip(t, newTabbableModel(4), btest.KeyNavConfig{})
	if !ok {
		t.Error("round trip should pass for tabbable model")
	}
}

func TestCheckKeyNavRoundTrip_Static(t *testing.T) {
	ok := btest.CheckKeyNavRoundTrip(t, &staticModel{}, btest.KeyNavConfig{MaxTabs: 5})
	if !ok {
		t.Error("round trip should trivially pass for static model")
	}
}

func TestAssertKeyNavCycle(t *testing.T) {
	// Should not fail the test — 3-item model has 3 stops.
	btest.AssertKeyNavCycle(t, newTabbableModel(3), 3, btest.KeyNavConfig{})
}

func TestAssertKeyNavRoundTrip(t *testing.T) {
	btest.AssertKeyNavRoundTrip(t, newTabbableModel(3), btest.KeyNavConfig{})
}

func TestKeyNavReport(t *testing.T) {
	result := btest.CheckKeyNav(t, newTabbableModel(3), btest.KeyNavConfig{})
	report := btest.KeyNavReport(result)
	if report == "" {
		t.Error("report should not be empty")
	}
	if !stringContains(report, "cycle detected") {
		t.Errorf("report should mention cycle detected: %s", report)
	}
}

func TestKeyNavReport_Static(t *testing.T) {
	result := btest.CheckKeyNav(t, &staticModel{}, btest.KeyNavConfig{MaxTabs: 5})
	report := btest.KeyNavReport(result)
	// Static model has trivial cycle (length 1), so report says PASS.
	if !stringContains(report, "Unique focus states: 1") {
		t.Errorf("report should show 1 unique state: %s", report)
	}
	if !stringContains(report, "Visible changes:     0") {
		t.Errorf("report should show 0 visible changes: %s", report)
	}
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
