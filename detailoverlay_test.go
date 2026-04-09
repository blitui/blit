package tuikit

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testItem struct {
	name  string
	value int
}

func newTestDetailOverlay() *DetailOverlay[testItem] {
	return NewDetailOverlay(DetailOverlayOpts[testItem]{
		Render: func(item testItem, w, h int, theme Theme) string {
			return item.name + " detail view"
		},
		Title: "Item Detail",
	})
}

func TestDetailOverlayInitInactive(t *testing.T) {
	d := newTestDetailOverlay()
	if d.IsActive() {
		t.Error("detail overlay should start inactive")
	}
}

func TestDetailOverlayShowActivates(t *testing.T) {
	d := newTestDetailOverlay()
	d.Show(testItem{name: "BTC", value: 100})
	if !d.IsActive() {
		t.Error("detail overlay should be active after Show()")
	}
}

func TestDetailOverlayItem(t *testing.T) {
	d := newTestDetailOverlay()
	d.Show(testItem{name: "BTC", value: 100})
	item := d.Item()
	if item.name != "BTC" {
		t.Errorf("expected 'BTC', got '%s'", item.name)
	}
	if item.value != 100 {
		t.Errorf("expected 100, got %d", item.value)
	}
}

func TestDetailOverlayClose(t *testing.T) {
	d := newTestDetailOverlay()
	d.Show(testItem{name: "BTC", value: 100})
	d.Close()
	if d.IsActive() {
		t.Error("detail overlay should be inactive after Close()")
	}
}

func TestDetailOverlayViewContainsRender(t *testing.T) {
	d := newTestDetailOverlay()
	d.SetTheme(DefaultTheme())
	d.SetSize(80, 24)
	d.Show(testItem{name: "BTC", value: 100})

	view := d.View()
	if !strings.Contains(view, "BTC detail view") {
		t.Error("view should contain rendered item content")
	}
}

func TestDetailOverlayViewContainsTitle(t *testing.T) {
	d := newTestDetailOverlay()
	d.SetTheme(DefaultTheme())
	d.SetSize(80, 24)
	d.Show(testItem{name: "BTC", value: 100})

	view := d.View()
	if !strings.Contains(view, "Item Detail") {
		t.Error("view should contain title")
	}
}

func TestDetailOverlayViewInactiveEmpty(t *testing.T) {
	d := newTestDetailOverlay()
	d.SetTheme(DefaultTheme())
	d.SetSize(80, 24)

	view := d.View()
	if view != "" {
		t.Error("inactive overlay should return empty view")
	}
}

func TestDetailOverlayOnKey(t *testing.T) {
	keyCalled := ""
	d := NewDetailOverlay(DetailOverlayOpts[testItem]{
		Render: func(item testItem, w, h int, theme Theme) string {
			return item.name
		},
		OnKey: func(item testItem, key tea.KeyMsg) tea.Cmd {
			keyCalled = key.String()
			return Consumed()
		},
	})
	d.SetTheme(DefaultTheme())
	d.SetSize(80, 24)
	d.Show(testItem{name: "ETH", value: 50})

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, Context{})

	if keyCalled != "s" {
		t.Errorf("OnKey should have been called with 's', got '%s'", keyCalled)
	}
}

func TestDetailOverlaySetActive(t *testing.T) {
	d := newTestDetailOverlay()
	d.SetActive(true)
	if !d.IsActive() {
		t.Error("SetActive(true) should activate")
	}
	d.SetActive(false)
	if d.IsActive() {
		t.Error("SetActive(false) should deactivate")
	}
}
