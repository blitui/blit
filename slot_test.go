package blit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSlotMainRegisters(t *testing.T) {
	main := &stubComponent{name: "main"}
	a := newAppModel(WithSlot(SlotMain, main))

	if got := a.slots.get(SlotMain); got != main {
		t.Fatalf("SlotMain get = %v, want %v", got, main)
	}
	if len(a.components) != 1 || a.components[0].component != main {
		t.Fatalf("expected main component registered, got %#v", a.components)
	}
	if !main.focused {
		t.Errorf("expected main to be focused initially")
	}
}

func TestSlotMainAndSidebarBuildDualPane(t *testing.T) {
	main := &stubComponent{name: "main"}
	side := &stubComponent{name: "side"}
	a := newAppModel(
		WithSlot(SlotMain, main),
		WithSlot(SlotSidebar, side),
	)

	if a.dualPane == nil {
		t.Fatal("expected implicit DualPane from Main+Sidebar slots")
	}
	if a.dualPane.Main != main {
		t.Errorf("dualPane.Main = %v, want main stub", a.dualPane.Main)
	}
	if a.dualPane.Side != side {
		t.Errorf("dualPane.Side = %v, want side stub", a.dualPane.Side)
	}
}

func TestSlotFocusOrderCyclesThroughSlots(t *testing.T) {
	main := &stubComponent{name: "main"}
	side := &stubComponent{name: "side"}
	a := newAppModel(
		WithSlot(SlotMain, main),
		WithSlot(SlotSidebar, side),
	)
	// Need a width so the sidebar is visible.
	a.Update(tea.WindowSizeMsg{Width: 200, Height: 40})

	if !main.focused {
		t.Fatal("main should start focused")
	}
	a.Update(tea.KeyMsg{Type: tea.KeyTab})
	if main.focused || !side.focused {
		t.Fatalf("after tab: main.focused=%v side.focused=%v", main.focused, side.focused)
	}
	a.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !main.focused || side.focused {
		t.Fatalf("after second tab: main.focused=%v side.focused=%v", main.focused, side.focused)
	}
}

func TestSlotOverlayPushAndPopViaTriggerKey(t *testing.T) {
	main := &stubComponent{name: "main"}
	overlay := &stubOverlay{name: "ovl"}
	a := newAppModel(
		WithSlot(SlotMain, main),
		WithSlotOverlay("settings", "c", overlay),
	)
	a.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Trigger key should push the overlay onto the stack.
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if a.overlays.active() != overlay {
		t.Fatalf("expected overlay active after trigger key, got %v", a.overlays.active())
	}
	if !overlay.active {
		t.Errorf("expected overlay.active=true after push")
	}

	// Esc should pop it back off.
	a.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if a.overlays.active() != nil {
		t.Fatalf("expected overlay popped after esc, got %v", a.overlays.active())
	}
}

func TestSlotOverlayStacksMultipleEntries(t *testing.T) {
	main := &stubComponent{name: "main"}
	o1 := &stubOverlay{name: "o1"}
	o2 := &stubOverlay{name: "o2"}
	a := newAppModel(
		WithSlot(SlotMain, main),
		WithSlotOverlay("first", "c", o1),
		WithSlotOverlay("second", "x", o2),
	)

	entries := a.slots.all(SlotOverlay)
	if len(entries) != 2 {
		t.Fatalf("expected 2 overlay slot entries, got %d", len(entries))
	}
	if entries[0].component != o1 || entries[1].component != o2 {
		t.Errorf("overlay entry order wrong: %#v", entries)
	}

	// Each overlay is independently triggerable via its own key.
	a.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if a.overlays.active() != o1 {
		t.Errorf("expected o1 active after 'c', got %v", a.overlays.active())
	}
	a.Update(tea.KeyMsg{Type: tea.KeyEsc})
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if a.overlays.active() != o2 {
		t.Errorf("expected o2 active after 'x', got %v", a.overlays.active())
	}
}

func TestSlotToastStackRoutedThroughApp(t *testing.T) {
	main := &stubComponent{name: "main"}
	a := newAppModel(WithSlot(SlotMain, main))
	a.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Dispatch three toasts — they should all land in the manager.
	for i := 0; i < 3; i++ {
		a.Update(ToastMsg{
			Severity: SeverityInfo,
			Title:    "hello",
			Duration: 2 * time.Second,
		})
	}
	if a.toasts == nil || !a.toasts.hasActive() {
		t.Fatal("expected toasts to be active after dispatch")
	}
	if len(a.toasts.toasts) != 3 {
		t.Errorf("expected 3 toasts, got %d", len(a.toasts.toasts))
	}

	// Dismiss top — stack should shrink.
	a.Update(dismissTopToastMsg{})
	if len(a.toasts.toasts) != 2 {
		t.Errorf("expected 2 toasts after dismissTop, got %d", len(a.toasts.toasts))
	}
}

func TestLegacyWithComponentRoutesThroughSlotMain(t *testing.T) {
	c := &stubComponent{name: "legacy"}
	a := newAppModel(WithComponent("legacy", c))
	if got := a.slots.get(SlotMain); got != c {
		t.Fatalf("legacy WithComponent did not populate SlotMain: got %v", got)
	}
}

func TestLegacyWithLayoutRoutesMainAndSidebar(t *testing.T) {
	main := &stubComponent{name: "m"}
	side := &stubComponent{name: "s"}
	a := newAppModel(WithLayout(&DualPane{Main: main, Side: side, SideWidth: 20, MinMainWidth: 40}))

	if a.slots.get(SlotMain) != main {
		t.Errorf("WithLayout did not set SlotMain")
	}
	if a.slots.get(SlotSidebar) != side {
		t.Errorf("WithLayout did not set SlotSidebar")
	}
}

func TestLegacyWithStatusBarRoutesThroughFooterSlot(t *testing.T) {
	a := newAppModel(WithStatusBar(
		func() string { return "left" },
		func() string { return "right" },
	))
	if a.slots.get(SlotFooter) == nil {
		t.Errorf("WithStatusBar did not populate SlotFooter")
	}
	if a.statusBar == nil {
		t.Errorf("statusBar field should remain populated for legacy render path")
	}
}

func TestLegacyWithOverlayRoutesThroughOverlaySlot(t *testing.T) {
	o := &stubOverlay{name: "legacy-ovl"}
	a := newAppModel(WithOverlay("legacy", "o", o))
	entries := a.slots.all(SlotOverlay)
	if len(entries) != 1 || entries[0].component != o {
		t.Fatalf("legacy WithOverlay not in SlotOverlay: %#v", entries)
	}
	if entries[0].overlayKey != "o" || entries[0].overlayName != "legacy" {
		t.Errorf("overlay metadata lost: %#v", entries[0])
	}
}
