package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

func makeTestTimeline() *blit.Timeline {
	events := []blit.TimelineEvent{
		{Time: "9:00", Title: "Standup", Status: "done"},
		{Time: "10:00", Title: "Code review", Status: "active", Description: "PR #42"},
		{Time: "14:00", Title: "Deploy", Status: "pending"},
	}
	tl := blit.NewTimeline(events, blit.TimelineOpts{})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 24)
	tl.SetFocused(true)
	return tl
}

func TestTimeline_NewDefaults(t *testing.T) {
	tl := makeTestTimeline()
	if tl.CursorIndex() != 0 {
		t.Fatalf("CursorIndex() = %d, want 0", tl.CursorIndex())
	}
	if len(tl.Events()) != 3 {
		t.Fatalf("Events() = %d, want 3", len(tl.Events()))
	}
}

func TestTimeline_ViewVertical(t *testing.T) {
	tl := makeTestTimeline()
	view := tl.View()
	if view == "" {
		t.Fatal("View() should not be empty")
	}
	if !strings.Contains(view, "Standup") {
		t.Fatalf("view should contain 'Standup':\n%s", view)
	}
	if !strings.Contains(view, "Code review") {
		t.Fatalf("view should contain 'Code review':\n%s", view)
	}
	if !strings.Contains(view, "PR #42") {
		t.Fatalf("view should contain description 'PR #42':\n%s", view)
	}
}

func TestTimeline_NavigateVertical(t *testing.T) {
	tl := makeTestTimeline()

	// Down.
	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", tl.CursorIndex())
	}

	// Down again.
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2", tl.CursorIndex())
	}

	// Down at end stays.
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 2 {
		t.Fatalf("CursorIndex() = %d, want 2 (clamped)", tl.CursorIndex())
	}

	// Up.
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyUp}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 1 {
		t.Fatalf("CursorIndex() = %d, want 1", tl.CursorIndex())
	}
}

func TestTimeline_ViNavigation(t *testing.T) {
	tl := makeTestTimeline()

	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 1 {
		t.Fatalf("j: CursorIndex() = %d, want 1", tl.CursorIndex())
	}

	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 0 {
		t.Fatalf("k: CursorIndex() = %d, want 0", tl.CursorIndex())
	}
}

func TestTimeline_Select(t *testing.T) {
	var selected blit.TimelineEvent
	var selectedIdx int
	events := []blit.TimelineEvent{
		{Time: "9:00", Title: "Event1"},
		{Time: "10:00", Title: "Event2"},
	}
	tl := blit.NewTimeline(events, blit.TimelineOpts{
		OnSelect: func(ev blit.TimelineEvent, idx int) {
			selected = ev
			selectedIdx = idx
		},
	})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 24)
	tl.SetFocused(true)

	// Move to Event2 and select.
	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyEnter}, blit.Context{})
	_ = updated.(*blit.Timeline)

	if selected.Title != "Event2" {
		t.Fatalf("selected = %q, want Event2", selected.Title)
	}
	if selectedIdx != 1 {
		t.Fatalf("selectedIdx = %d, want 1", selectedIdx)
	}
}

func TestTimeline_Horizontal(t *testing.T) {
	events := []blit.TimelineEvent{
		{Time: "Q1", Title: "Plan", Status: "done"},
		{Time: "Q2", Title: "Build", Status: "active"},
		{Time: "Q3", Title: "Ship", Status: "pending"},
	}
	tl := blit.NewTimeline(events, blit.TimelineOpts{Horizontal: true})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 10)
	tl.SetFocused(true)

	view := tl.View()
	if view == "" {
		t.Fatal("horizontal View() should not be empty")
	}
	if !strings.Contains(view, "Q1") {
		t.Fatalf("view should contain 'Q1':\n%s", view)
	}
}

func TestTimeline_HorizontalNavigation(t *testing.T) {
	events := []blit.TimelineEvent{
		{Time: "A", Title: "A"},
		{Time: "B", Title: "B"},
	}
	tl := blit.NewTimeline(events, blit.TimelineOpts{Horizontal: true})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 10)
	tl.SetFocused(true)

	// Right moves cursor.
	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyRight}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 1 {
		t.Fatalf("right: CursorIndex() = %d, want 1", tl.CursorIndex())
	}

	// Left moves back.
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyLeft}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 0 {
		t.Fatalf("left: CursorIndex() = %d, want 0", tl.CursorIndex())
	}

	// Up/Down should NOT move in horizontal mode.
	updated, _ = tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 0 {
		t.Fatal("down should not move cursor in horizontal mode")
	}
}

func TestTimeline_SetEvents(t *testing.T) {
	tl := makeTestTimeline()

	// Move cursor to end.
	tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)

	// Replace with fewer events.
	tl.SetEvents([]blit.TimelineEvent{{Time: "1", Title: "Only"}})
	if tl.CursorIndex() != 0 {
		t.Fatalf("CursorIndex() = %d, want 0 after SetEvents", tl.CursorIndex())
	}
}

func TestTimeline_Empty(t *testing.T) {
	tl := blit.NewTimeline([]blit.TimelineEvent{}, blit.TimelineOpts{})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 24)
	if tl.View() != "" {
		t.Fatal("empty timeline should return empty view")
	}
}

func TestTimeline_ZeroSize(t *testing.T) {
	tl := makeTestTimeline()
	tl.SetSize(0, 0)
	if tl.View() != "" {
		t.Fatal("zero-sized timeline should return empty view")
	}
}

func TestTimeline_UnfocusedIgnoresInput(t *testing.T) {
	tl := makeTestTimeline()
	tl.SetFocused(false)

	updated, _ := tl.Update(tea.KeyMsg{Type: tea.KeyDown}, blit.Context{})
	tl = updated.(*blit.Timeline)
	if tl.CursorIndex() != 0 {
		t.Fatal("unfocused timeline should not process input")
	}
}

func TestTimeline_KeyBindingsVertical(t *testing.T) {
	tl := blit.NewTimeline([]blit.TimelineEvent{{Title: "A"}}, blit.TimelineOpts{})
	binds := tl.KeyBindings()
	found := false
	for _, b := range binds {
		if strings.Contains(b.Key, "up") {
			found = true
		}
	}
	if !found {
		t.Fatal("vertical timeline should have up/down bindings")
	}
}

func TestTimeline_KeyBindingsHorizontal(t *testing.T) {
	tl := blit.NewTimeline([]blit.TimelineEvent{{Title: "A"}}, blit.TimelineOpts{Horizontal: true})
	binds := tl.KeyBindings()
	found := false
	for _, b := range binds {
		if strings.Contains(b.Key, "left") {
			found = true
		}
	}
	if !found {
		t.Fatal("horizontal timeline should have left/right bindings")
	}
}

func TestTimeline_ComponentInterface(t *testing.T) {
	tl := makeTestTimeline()
	if cmd := tl.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}
	tl.SetFocused(false)
	if tl.Focused() {
		t.Fatal("should not be focused")
	}
	tl.SetFocused(true)
	if !tl.Focused() {
		t.Fatal("should be focused")
	}
}

func TestTimeline_StatusIcons(t *testing.T) {
	events := []blit.TimelineEvent{
		{Time: "1", Title: "Done", Status: "done"},
		{Time: "2", Title: "Active", Status: "active"},
		{Time: "3", Title: "Pending", Status: "pending"},
		{Time: "4", Title: "Custom", Status: "other"},
	}
	tl := blit.NewTimeline(events, blit.TimelineOpts{})
	tl.SetTheme(blit.DefaultTheme())
	tl.SetSize(80, 30)

	g := blit.DefaultGlyphs()
	view := tl.View()
	if !strings.Contains(view, g.Check) {
		t.Fatal("done event should show check icon")
	}
	if !strings.Contains(view, g.Star) {
		t.Fatal("active event should show star icon")
	}
}
