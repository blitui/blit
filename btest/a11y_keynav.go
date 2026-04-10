package btest

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyNavResult holds the outcome of a keyboard navigation check.
type KeyNavResult struct {
	// Screens holds the screen state after each tab press.
	Screens []string

	// UniqueScreens is the number of distinct screen states visited.
	UniqueScreens int

	// CycleLength is the number of tab presses before the screen state
	// repeats (i.e., the focus cycle length). Zero if no cycle was found
	// within maxTabs presses.
	CycleLength int

	// FocusChanges counts how many tab presses produced a visible change
	// in the rendered screen. A good component should change on every tab.
	FocusChanges int
}

// KeyNavConfig configures the keyboard navigation verifier.
type KeyNavConfig struct {
	// MaxTabs is the maximum number of tab presses to attempt before
	// giving up on finding a cycle. Defaults to 50.
	MaxTabs int

	// Cols and Lines set the terminal dimensions. Defaults to 80x24.
	Cols  int
	Lines int

	// ForwardKey is the key used to move focus forward. Defaults to "tab".
	ForwardKey string

	// BackwardKey is the key used to move focus backward. Defaults to "shift+tab".
	BackwardKey string
}

func (c *KeyNavConfig) defaults() {
	if c.MaxTabs <= 0 {
		c.MaxTabs = 50
	}
	if c.Cols <= 0 {
		c.Cols = 80
	}
	if c.Lines <= 0 {
		c.Lines = 24
	}
	if c.ForwardKey == "" {
		c.ForwardKey = "tab"
	}
	if c.BackwardKey == "" {
		c.BackwardKey = "shift+tab"
	}
}

// CheckKeyNav exercises keyboard navigation on a model by pressing tab
// repeatedly and tracking screen changes. It detects the focus cycle
// length and reports how many tab presses produce visible focus changes.
func CheckKeyNav(t testing.TB, model tea.Model, cfg KeyNavConfig) *KeyNavResult {
	t.Helper()
	cfg.defaults()

	tm := NewTestModel(t, model, cfg.Cols, cfg.Lines)
	initial := tm.Screen().String()

	seen := map[string]int{initial: 0}
	screens := []string{initial}
	focusChanges := 0
	cycleLen := 0

	prev := initial
	for i := 1; i <= cfg.MaxTabs; i++ {
		tm.SendKey(cfg.ForwardKey)
		current := tm.Screen().String()
		screens = append(screens, current)

		if current != prev {
			focusChanges++
		}
		prev = current

		if firstIdx, ok := seen[current]; ok {
			cycleLen = i - firstIdx
			break
		}
		seen[current] = i
	}

	unique := len(seen)

	return &KeyNavResult{
		Screens:       screens,
		UniqueScreens: unique,
		CycleLength:   cycleLen,
		FocusChanges:  focusChanges,
	}
}

// CheckKeyNavRoundTrip verifies that pressing tab N times and then
// shift+tab N times returns to the original screen state.
func CheckKeyNavRoundTrip(t testing.TB, model tea.Model, cfg KeyNavConfig) bool {
	t.Helper()
	cfg.defaults()

	tm := NewTestModel(t, model, cfg.Cols, cfg.Lines)
	initial := tm.Screen().String()

	// First find the cycle length.
	result := CheckKeyNav(t, model, cfg)
	n := result.CycleLength
	if n == 0 {
		// No cycle found; use FocusChanges as a fallback.
		n = result.FocusChanges
	}
	if n == 0 {
		return true // No navigation targets — trivially passes.
	}

	// Tab forward n times.
	for i := 0; i < n; i++ {
		tm.SendKey(cfg.ForwardKey)
	}

	// Tab backward n times.
	for i := 0; i < n; i++ {
		tm.SendKey(cfg.BackwardKey)
	}

	final := tm.Screen().String()
	return normalizeKeyNavScreen(initial) == normalizeKeyNavScreen(final)
}

// AssertKeyNavCycle fails the test if the model does not have a keyboard
// navigation cycle (at least minStops distinct focus states reachable via tab).
func AssertKeyNavCycle(t testing.TB, model tea.Model, minStops int, cfg KeyNavConfig) {
	t.Helper()
	result := CheckKeyNav(t, model, cfg)
	if result.CycleLength == 0 {
		t.Errorf("keyboard navigation: no focus cycle found within %d tab presses", cfg.MaxTabs)
		return
	}
	if result.UniqueScreens < minStops {
		t.Errorf("keyboard navigation: %d unique focus states, want ≥ %d",
			result.UniqueScreens, minStops)
	}
}

// AssertKeyNavRoundTrip fails the test if tabbing forward and then
// backward does not return to the initial state.
func AssertKeyNavRoundTrip(t testing.TB, model tea.Model, cfg KeyNavConfig) {
	t.Helper()
	if !CheckKeyNavRoundTrip(t, model, cfg) {
		t.Error("keyboard navigation: tab forward + shift+tab backward did not return to initial state")
	}
}

// KeyNavReport generates a human-readable summary of keyboard navigation.
func KeyNavReport(result *KeyNavResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Keyboard navigation:\n")
	fmt.Fprintf(&sb, "  Unique focus states: %d\n", result.UniqueScreens)
	fmt.Fprintf(&sb, "  Focus cycle length:  %d\n", result.CycleLength)
	fmt.Fprintf(&sb, "  Visible changes:     %d / %d tabs\n",
		result.FocusChanges, len(result.Screens)-1)
	if result.CycleLength > 0 {
		fmt.Fprintf(&sb, "  Status: PASS (cycle detected)\n")
	} else {
		fmt.Fprintf(&sb, "  Status: WARN (no cycle found)\n")
	}
	return sb.String()
}

// normalizeKeyNavScreen trims trailing whitespace from each line for
// comparison, since focus indicators may affect trailing spaces.
func normalizeKeyNavScreen(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}
