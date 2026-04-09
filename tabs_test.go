package tuikit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// stubContent is a minimal Component used as tab content in tests.
type stubContent struct {
	name    string
	focused bool
	width   int
	height  int
	theme   Theme
}

func newStub(name string) *stubContent { return &stubContent{name: name} }

func (s *stubContent) Init() tea.Cmd                           { return nil }
func (s *stubContent) Update(msg tea.Msg) (Component, tea.Cmd) { return s, nil }
func (s *stubContent) View() string                            { return "[" + s.name + "]" }
func (s *stubContent) KeyBindings() []KeyBind                  { return nil }
func (s *stubContent) SetSize(w, h int)                        { s.width = w; s.height = h }
func (s *stubContent) Focused() bool                           { return s.focused }
func (s *stubContent) SetFocused(f bool)                       { s.focused = f }
func (s *stubContent) SetTheme(th Theme)                       { s.theme = th }

func newTestTabs(n int) *Tabs {
	items := make([]TabItem, n)
	for i := 0; i < n; i++ {
		items[i] = TabItem{
			Title:   tabTitle(i),
			Content: newStub(tabTitle(i)),
		}
	}
	t := NewTabs(items, TabsOpts{})
	t.SetTheme(DefaultTheme())
	t.SetSize(80, 24)
	return t
}

func tabTitle(i int) string {
	names := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	if i < len(names) {
		return names[i]
	}
	return "Tab"
}

// ── D4: SetActive / OnChange ──────────────────────────────────────────────────

func TestTabs_SetActive(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetActive(2)
	if tabs.ActiveIndex() != 2 {
		t.Errorf("SetActive(2) → ActiveIndex = %d, want 2", tabs.ActiveIndex())
	}
}

func TestTabs_SetActiveClamps(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetActive(100)
	if tabs.ActiveIndex() != 2 {
		t.Errorf("SetActive(100) → ActiveIndex = %d, want 2", tabs.ActiveIndex())
	}
	tabs.SetActive(-5)
	if tabs.ActiveIndex() != 0 {
		t.Errorf("SetActive(-5) → ActiveIndex = %d, want 0", tabs.ActiveIndex())
	}
}

func TestTabs_OnChange(t *testing.T) {
	var got []int
	tabs := newTestTabs(3)
	tabs.OnChange(func(i int) { got = append(got, i) })

	tabs.SetActive(1)
	tabs.SetActive(2)
	tabs.SetActive(2) // no-op: same index

	if len(got) != 2 {
		t.Fatalf("OnChange called %d times, want 2", len(got))
	}
	if got[0] != 1 || got[1] != 2 {
		t.Errorf("OnChange args = %v, want [1 2]", got)
	}
}

// ── D2: Keybinds ─────────────────────────────────────────────────────────────

func TestTabs_TabKeyCycles(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetFocused(true)

	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	if tabs.ActiveIndex() != 1 {
		t.Errorf("tab → index = %d, want 1", tabs.ActiveIndex())
	}
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	if tabs.ActiveIndex() != 2 {
		t.Errorf("tab → index = %d, want 2", tabs.ActiveIndex())
	}
	// wraps around
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	if tabs.ActiveIndex() != 0 {
		t.Errorf("tab wrap → index = %d, want 0", tabs.ActiveIndex())
	}
}

func TestTabs_ShiftTabCycles(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetFocused(true)

	tabs.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if tabs.ActiveIndex() != 2 {
		t.Errorf("shift+tab from 0 → index = %d, want 2", tabs.ActiveIndex())
	}
}

func TestTabs_NumberJump(t *testing.T) {
	tabs := newTestTabs(5)
	tabs.SetFocused(true)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if tabs.ActiveIndex() != 2 {
		t.Errorf("key '3' → index = %d, want 2", tabs.ActiveIndex())
	}
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if tabs.ActiveIndex() != 0 {
		t.Errorf("key '1' → index = %d, want 0", tabs.ActiveIndex())
	}
}

func TestTabs_NumberJumpOutOfRange(t *testing.T) {
	tabs := newTestTabs(2)
	tabs.SetFocused(true)
	before := tabs.ActiveIndex()

	// Key '5' is out of range for a 2-tab set — index unchanged.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	if tabs.ActiveIndex() != before {
		t.Errorf("out-of-range key changed index to %d", tabs.ActiveIndex())
	}
}

// ── D2: Mouse click ───────────────────────────────────────────────────────────

func TestTabs_MouseClickHorizontal(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetFocused(true)

	// Row 0, X=0 → first tab (always starts at x=0).
	tabs.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 0, Y: 0})
	if tabs.ActiveIndex() != 0 {
		t.Errorf("click x=0 y=0 → index = %d, want 0", tabs.ActiveIndex())
	}
}

func TestTabs_MouseClickIgnoresRelease(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetActive(2)

	tabs.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease, X: 0, Y: 0})
	if tabs.ActiveIndex() != 2 {
		t.Errorf("release event changed active tab, should be no-op")
	}
}

func TestTabs_MouseClickVertical(t *testing.T) {
	items := []TabItem{
		{Title: "One", Content: newStub("One")},
		{Title: "Two", Content: newStub("Two")},
		{Title: "Three", Content: newStub("Three")},
	}
	tabs := NewTabs(items, TabsOpts{Orientation: Vertical})
	tabs.SetTheme(DefaultTheme())
	tabs.SetSize(80, 24)
	tabs.SetFocused(true)

	// Click on row 1 (second tab) within bar width.
	tabs.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 0, Y: 1})
	if tabs.ActiveIndex() != 1 {
		t.Errorf("vertical click row=1 → index = %d, want 1", tabs.ActiveIndex())
	}
}

// ── D1: View / content rendering ─────────────────────────────────────────────

func TestTabs_ViewRendersActiveContent(t *testing.T) {
	tabs := newTestTabs(3)
	view := tabs.View()
	if !strings.Contains(view, "[Alpha]") {
		t.Errorf("view missing active content '[Alpha]': %q", view)
	}
	// Inactive content should not appear in view
	if strings.Contains(view, "[Beta]") {
		t.Errorf("view should not contain inactive content '[Beta]'")
	}
}

func TestTabs_ViewContainsTabTitles(t *testing.T) {
	tabs := newTestTabs(3)
	view := tabs.View()
	for _, title := range []string{"Alpha", "Beta", "Gamma"} {
		if !strings.Contains(view, title) {
			t.Errorf("view missing tab title %q", title)
		}
	}
}

func TestTabs_ViewChangesOnSetActive(t *testing.T) {
	tabs := newTestTabs(3)
	tabs.SetActive(1)
	view := tabs.View()
	if !strings.Contains(view, "[Beta]") {
		t.Errorf("after SetActive(1), view missing '[Beta]': %q", view)
	}
}

func TestTabs_VerticalViewRendersActiveContent(t *testing.T) {
	items := []TabItem{
		{Title: "One", Content: newStub("One")},
		{Title: "Two", Content: newStub("Two")},
	}
	tabs := NewTabs(items, TabsOpts{Orientation: Vertical})
	tabs.SetTheme(DefaultTheme())
	tabs.SetSize(80, 24)
	view := tabs.View()
	if !strings.Contains(view, "[One]") {
		t.Errorf("vertical view missing active content '[One]': %q", view)
	}
}

// ── D1: Glyph support ─────────────────────────────────────────────────────────

func TestTabs_GlyphInTabLabel(t *testing.T) {
	items := []TabItem{
		{Title: "Files", Glyph: "F", Content: newStub("Files")},
	}
	tabs := NewTabs(items, TabsOpts{})
	tabs.SetTheme(DefaultTheme())
	tabs.SetSize(80, 24)
	view := tabs.View()
	if !strings.Contains(view, "F Files") {
		t.Errorf("view missing glyph label 'F Files': %q", view)
	}
}

// ── D6: Inner component state persists across tab switches ────────────────────

func TestTabs_InnerStatePreservedAcrossSwitches(t *testing.T) {
	// Use a stub that records SetFocused calls so we can verify it keeps its state.
	s0 := newStub("First")
	s1 := newStub("Second")
	items := []TabItem{
		{Title: "First", Content: s0},
		{Title: "Second", Content: s1},
	}
	tabs := NewTabs(items, TabsOpts{})
	tabs.SetTheme(DefaultTheme())
	tabs.SetSize(80, 24)
	tabs.SetFocused(true)

	// Switch to second tab and back — s0 should still have its name unchanged.
	tabs.SetActive(1)
	tabs.SetActive(0)

	view := tabs.View()
	if !strings.Contains(view, "[First]") {
		t.Errorf("after round-trip, expected '[First]' in view: %q", view)
	}
}

// ── Component interface compliance ────────────────────────────────────────────

func TestTabs_InitReturnsNilForEmptyContent(t *testing.T) {
	items := []TabItem{
		{Title: "A", Content: nil},
	}
	tabs := NewTabs(items, TabsOpts{})
	cmd := tabs.Init()
	_ = cmd // should not panic
}

func TestTabs_KeyBindings(t *testing.T) {
	tabs := newTestTabs(3)
	binds := tabs.KeyBindings()
	if len(binds) == 0 {
		t.Error("KeyBindings should return non-empty slice")
	}
	found := map[string]bool{}
	for _, b := range binds {
		found[b.Key] = true
	}
	for _, want := range []string{"tab", "shift+tab", "1-9"} {
		if !found[want] {
			t.Errorf("KeyBindings missing %q", want)
		}
	}
}

func TestTabs_SetFocused(t *testing.T) {
	tabs := newTestTabs(2)
	tabs.SetFocused(true)
	if !tabs.Focused() {
		t.Error("Focused() should return true after SetFocused(true)")
	}
	// Active content should be focused.
	active := tabs.items[tabs.active].Content
	if active != nil && !active.Focused() {
		t.Error("active content should be focused when tabs are focused")
	}
	tabs.SetFocused(false)
	if tabs.Focused() {
		t.Error("Focused() should return false after SetFocused(false)")
	}
}

func TestTabs_SetTheme(t *testing.T) {
	tabs := newTestTabs(2)
	th := LightTheme()
	tabs.SetTheme(th)
	// Verify a comparable field was updated.
	if tabs.theme.Accent != th.Accent {
		t.Error("SetTheme did not update theme field")
	}
	// Content components should also receive the theme.
	for _, item := range tabs.items {
		if s, ok := item.Content.(*stubContent); ok {
			if s.theme.Accent != th.Accent {
				t.Errorf("content %q did not receive theme", s.name)
			}
		}
	}
}

func TestTabs_EmptyTabs(t *testing.T) {
	tabs := NewTabs(nil, TabsOpts{})
	tabs.SetTheme(DefaultTheme())
	tabs.SetSize(80, 24)
	_ = tabs.Init()
	_ = tabs.View()
	tabs.SetActive(0)
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
}
