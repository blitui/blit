package blit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReleaseNotesOverlay_EmptyShowsPlaceholder(t *testing.T) {
	o := NewReleaseNotesOverlay("v1.2.3", "")
	if !strings.Contains(o.View(), "no release notes") {
		t.Errorf("expected placeholder, got:\n%s", o.View())
	}
}

func TestReleaseNotesOverlay_RendersVersionAndBody(t *testing.T) {
	o := NewReleaseNotesOverlay("v1.2.3", "First line\nSecond line\nThird line")
	v := o.View()
	for _, want := range []string{"v1.2.3", "First line", "Second line", "Third line"} {
		if !strings.Contains(v, want) {
			t.Errorf("missing %q in view:\n%s", want, v)
		}
	}
}

func TestReleaseNotesOverlay_ScrollDownUp(t *testing.T) {
	notes := strings.Repeat("x\n", 100)
	o := NewReleaseNotesOverlay("v1", notes)
	o.Height = 10 // visible = 4
	before := o.Offset
	o.Update(tea.KeyMsg{Type: tea.KeyDown})
	if o.Offset != before+1 {
		t.Errorf("down: offset = %d, want %d", o.Offset, before+1)
	}
	o.Update(tea.KeyMsg{Type: tea.KeyUp})
	if o.Offset != before {
		t.Errorf("up: offset = %d, want %d", o.Offset, before)
	}
}

func TestReleaseNotesOverlay_PageDownPageUp(t *testing.T) {
	lines := strings.Repeat("x\n", 100)
	o := NewReleaseNotesOverlay("v1", lines)
	o.Height = 10 // visible = 4
	o.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if o.Offset != 4 {
		t.Errorf("pgdown: offset = %d, want 4", o.Offset)
	}
	o.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if o.Offset != 0 {
		t.Errorf("pgup: offset = %d, want 0", o.Offset)
	}
}

func TestReleaseNotesOverlay_HomeEnd(t *testing.T) {
	lines := strings.Repeat("x\n", 100)
	o := NewReleaseNotesOverlay("v1", lines)
	o.Height = 10
	o.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if o.Offset != o.MaxOffset() {
		t.Errorf("end: offset = %d, want %d", o.Offset, o.MaxOffset())
	}
	o.Update(tea.KeyMsg{Type: tea.KeyHome})
	if o.Offset != 0 {
		t.Errorf("home: offset = %d, want 0", o.Offset)
	}
}

func TestReleaseNotesOverlay_ScrollDownClamped(t *testing.T) {
	o := NewReleaseNotesOverlay("v1", "one\ntwo\nthree")
	o.Height = 24
	for i := 0; i < 50; i++ {
		o.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if o.Offset > o.MaxOffset() {
		t.Errorf("offset %d exceeds max %d", o.Offset, o.MaxOffset())
	}
}

func TestReleaseNotesOverlay_QuitClosesOverlay(t *testing.T) {
	o := NewReleaseNotesOverlay("v1", "line")
	o.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !o.Closed {
		t.Error("esc should close overlay")
	}
	o2 := NewReleaseNotesOverlay("v1", "line")
	o2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if !o2.Closed {
		t.Error("q should close overlay")
	}
}

func TestReleaseNotesOverlay_WindowResize(t *testing.T) {
	o := NewReleaseNotesOverlay("v1", "line")
	o.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if o.Width != 120 || o.Height != 40 {
		t.Errorf("resize: got %dx%d", o.Width, o.Height)
	}
}

func TestReleaseNotesOverlay_PositionIndicator(t *testing.T) {
	o := NewReleaseNotesOverlay("v1", "one\ntwo")
	o.Height = 24
	if !strings.Contains(o.View(), "all") {
		t.Errorf("short notes should show 'all':\n%s", o.View())
	}

	long := strings.Repeat("x\n", 100)
	o2 := NewReleaseNotesOverlay("v1", long)
	o2.Height = 10
	if !strings.Contains(o2.View(), "top") {
		t.Errorf("expected 'top' at start:\n%s", o2.View())
	}
	o2.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if !strings.Contains(o2.View(), "end") {
		t.Errorf("expected 'end' after End key:\n%s", o2.View())
	}
}
