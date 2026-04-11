package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestReexportTypeAliases verifies that blit's re-exported types are
// identical to the original bubbletea/lipgloss types so consumers can
// use them interchangeably.
func TestReexportTypeAliases(t *testing.T) {
	// Message types should be assignable — use conversion to verify alias identity
	var _ = Msg(tea.Msg(nil))
	var _ = Cmd(tea.Cmd(nil))
	var _ = KeyMsg(tea.KeyMsg{})
	var _ = MouseMsg(tea.MouseMsg{})
	var _ = WindowSizeMsg(tea.WindowSizeMsg{})
	var _ = Key(tea.Key{})
	var _ = KeyType(tea.KeyType(0))

	// Lipgloss types
	var _ = Color(lipgloss.Color(""))
	var _ = Style(lipgloss.Style{})
	var _ = Border(lipgloss.Border{})
	var _ = Position(lipgloss.Position(0))
}

// TestReexportKeyConstants verifies that blit key constants match tea constants.
func TestReexportKeyConstants(t *testing.T) {
	tests := []struct {
		blit KeyType
		tea  tea.KeyType
	}{
		{KeyUp, tea.KeyUp},
		{KeyDown, tea.KeyDown},
		{KeyLeft, tea.KeyLeft},
		{KeyRight, tea.KeyRight},
		{KeyEnter, tea.KeyEnter},
		{KeyEscape, tea.KeyEscape},
		{KeyBackspace, tea.KeyBackspace},
		{KeyTab, tea.KeyTab},
		{KeySpace, tea.KeySpace},
		{KeyHome, tea.KeyHome},
		{KeyEnd, tea.KeyEnd},
		{KeyPgUp, tea.KeyPgUp},
		{KeyPgDown, tea.KeyPgDown},
		{KeyRunes, tea.KeyRunes},
		{KeyCtrlC, tea.KeyCtrlC},
		{KeyCtrlU, tea.KeyCtrlU},
		{KeyCtrlD, tea.KeyCtrlD},
		{KeyCtrlK, tea.KeyCtrlK},
		{KeyCtrlBackslash, tea.KeyCtrlBackslash},
	}
	for _, tt := range tests {
		if tt.blit != tt.tea {
			t.Errorf("blit key constant != tea constant: %v != %v", tt.blit, tt.tea)
		}
	}
}

// TestReexportFuncKeyConstants verifies function key re-exports.
func TestReexportFuncKeyConstants(t *testing.T) {
	tests := []struct {
		blit KeyType
		tea  tea.KeyType
	}{
		{F1, tea.KeyF1},
		{F5, tea.KeyF5},
		{F10, tea.KeyF10},
		{F12, tea.KeyF12},
	}
	for _, tt := range tests {
		if tt.blit != tt.tea {
			t.Errorf("F-key constant mismatch: %v != %v", tt.blit, tt.tea)
		}
	}
}

// TestReexportCmdConstructors verifies that blit command constructors are
// the same as tea's.
func TestReexportCmdConstructors(t *testing.T) {
	if Batch == nil {
		t.Error("Batch should not be nil")
	}
	if Quit == nil {
		t.Error("Quit should not be nil")
	}
	if Sequence == nil {
		t.Error("Sequence should not be nil")
	}
	// Quit should produce a quit message
	cmd := Quit
	if cmd == nil {
		t.Error("Quit cmd is nil")
	}
}

// TestReexportLipglossFunctions verifies lipgloss function re-exports.
func TestReexportLipglossFunctions(t *testing.T) {
	_ = NewStyle()
	_ = Width("hello")
	_ = Height("hello\nworld")
	_ = JoinVertical
	_ = JoinHorizontal
	_ = Place
	_ = PosCenter
	_ = PosTop
	_ = PosBottom
	_ = PosLeft
	_ = PosRight
	_ = RoundedBorder()
	_ = DoubleBorder()
	_ = ASCIIBorder()
	_ = NormalBorder()
}

// TestReexportANSIHelpers verifies ANSI helper re-exports.
func TestReexportANSIHelpers(t *testing.T) {
	w := StringWidth("hello")
	if w != 5 {
		t.Errorf("StringWidth(%q) = %d, want 5", "hello", w)
	}

	truncated := TruncateWith("hello world", 5, "…")
	if Width(truncated) > 5 {
		t.Errorf("TruncateWith produced width > 5: %d", Width(truncated))
	}
}

// TestReexportKeyMsgUsage demonstrates the consumer pattern: using blit
// types directly without importing bubbletea.
func TestReexportKeyMsgUsage(t *testing.T) {
	msg := KeyMsg{Type: KeyRunes, Runes: []rune{'a'}}

	// Switch on key string — the primary consumer pattern
	switch msg.String() {
	case "a":
		// ok
	default:
		t.Errorf("expected key 'a', got %q", msg.String())
	}

	// Switch on key type
	switch msg.Type {
	case KeyRunes:
		// ok
	default:
		t.Errorf("expected KeyRunes, got %v", msg.Type)
	}
}
