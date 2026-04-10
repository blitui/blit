package blit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestPicker(items []PickerItem, opts PickerOpts) *Picker {
	p := NewPicker(items, opts)
	p.SetTheme(DefaultTheme())
	p.SetSize(80, 24)
	p.SetFocused(true)
	return p
}

func samplePickerItems() []PickerItem {
	return []PickerItem{
		{Title: "go build", Subtitle: "compile packages", Glyph: "  "},
		{Title: "go test", Subtitle: "run tests", Glyph: "  "},
		{Title: "go fmt", Subtitle: "format source", Glyph: "  "},
		{Title: "git commit", Subtitle: "record changes", Glyph: "  "},
		{Title: "git push", Subtitle: "upload changes", Glyph: "  "},
		{Title: "make build", Subtitle: "run make", Glyph: "  "},
	}
}

func sendPickerKey(p *Picker, key string) {
	var msg tea.KeyMsg
	switch key {
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+k":
		msg = tea.KeyMsg{Type: tea.KeyCtrlK}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	p.handleKey(msg)
}

func TestPicker_InitialState(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	if len(p.filtered) != len(samplePickerItems()) {
		t.Errorf("filtered len = %d, want %d", len(p.filtered), len(samplePickerItems()))
	}
	if p.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", p.cursor)
	}
}

func TestPicker_FilterReducesResults(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	p.input.SetValue("go")
	p.rebuildFiltered()
	if len(p.filtered) == 0 {
		t.Fatal("filter go should match at least one item")
	}
}

func TestPicker_Navigate_Down(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	sendPickerKey(p, "down")
	if p.cursor != 1 {
		t.Errorf("after down, cursor = %d, want 1", p.cursor)
	}
}

func TestPicker_Navigate_Up(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	sendPickerKey(p, "down")
	sendPickerKey(p, "down")
	sendPickerKey(p, "up")
	if p.cursor != 1 {
		t.Errorf("after down down up, cursor = %d, want 1", p.cursor)
	}
}

func TestPicker_Navigate_ClampTop(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	sendPickerKey(p, "up")
	if p.cursor != 0 {
		t.Errorf("up at top, cursor = %d, want 0", p.cursor)
	}
}

func TestPicker_Navigate_ClampBottom(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	for i := 0; i < 20; i++ {
		sendPickerKey(p, "down")
	}
	if p.cursor != len(p.filtered)-1 {
		t.Errorf("past bottom, cursor = %d, want %d", p.cursor, len(p.filtered)-1)
	}
}

func TestPicker_Confirm(t *testing.T) {
	var confirmed PickerItem
	p := newTestPicker(samplePickerItems(), PickerOpts{
		OnConfirm: func(item PickerItem) { confirmed = item },
	})
	sendPickerKey(p, "enter")
	if confirmed.Title == "" {
		t.Error("OnConfirm should have been called with an item")
	}
	if p.IsActive() {
		t.Error("picker should be inactive after confirm")
	}
}

func TestPicker_Cancel(t *testing.T) {
	cancelled := false
	p := newTestPicker(samplePickerItems(), PickerOpts{
		OnCancel: func() { cancelled = true },
	})
	sendPickerKey(p, "esc")
	if !cancelled {
		t.Error("OnCancel should have been called")
	}
	if p.IsActive() {
		t.Error("picker should be inactive after esc")
	}
}

func TestPicker_ClearFilter(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	p.input.SetValue("go")
	p.rebuildFiltered()
	filtered := len(p.filtered)

	sendPickerKey(p, "ctrl+k")
	p.rebuildFiltered()
	if len(p.filtered) <= filtered {
		t.Errorf("after ctrl+k clear, filtered=%d should be > %d", len(p.filtered), filtered)
	}
}

func TestPicker_PreviewUpdate(t *testing.T) {
	items := []PickerItem{
		{Title: "item-a", Preview: func() string { return "preview-a" }},
		{Title: "item-b", Preview: func() string { return "preview-b" }},
	}
	p := newTestPicker(items, PickerOpts{Preview: true})

	view := p.View()
	if !strings.Contains(view, "preview-a") {
		t.Errorf("first item preview not shown: %q", view)
	}

	sendPickerKey(p, "down")
	p.ensurePreview()
	if p.previewContent != "preview-b" {
		t.Errorf("preview after down = %q, want preview-b", p.previewContent)
	}
}

func TestPicker_ViewRendersItems(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	view := p.View()
	if !strings.Contains(view, "go build") {
		t.Errorf("view missing go build: %s", view)
	}
}

func TestPicker_ViewNoResults(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	p.input.SetValue("zzzxxx")
	p.rebuildFiltered()
	view := p.View()
	if !strings.Contains(view, "No results") {
		t.Errorf("view with no matches should show No results: %s", view)
	}
}

func TestPicker_Open(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	p.Close()
	if p.IsActive() {
		t.Error("should be inactive after Close")
	}
	p.Open()
	if !p.IsActive() {
		t.Error("should be active after Open")
	}
}

func TestPicker_SetItems(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	p.SetItems([]PickerItem{{Title: "only"}})
	if len(p.items) != 1 {
		t.Errorf("SetItems: items len = %d, want 1", len(p.items))
	}
	if len(p.filtered) != 1 {
		t.Errorf("SetItems: filtered len = %d, want 1", len(p.filtered))
	}
}

func TestPicker_KeyBindings(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	binds := p.KeyBindings()
	if len(binds) < 5 {
		t.Errorf("expected >=5 keybinds, got %d", len(binds))
	}
}

func TestPicker_Init(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	cmd := p.Init()
	if cmd == nil {
		t.Error("Init should return textinput.Blink cmd")
	}
}

func TestPicker_Overlay_Interface(t *testing.T) {
	p := newTestPicker(samplePickerItems(), PickerOpts{})
	var _ Overlay = p
}
