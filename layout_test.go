package tuikit

import "testing"

func TestSinglePaneSize(t *testing.T) {
	sp := SinglePane{}
	main, side, sideVisible := sp.compute(120, 40)
	if main.width != 120 || main.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", main.width, main.height)
	}
	if sideVisible {
		t.Error("SinglePane should not have a side")
	}
	if side.width != 0 || side.height != 0 {
		t.Errorf("side should be 0x0, got %dx%d", side.width, side.height)
	}
}

func TestDualPaneSizeNormal(t *testing.T) {
	dp := DualPane{SideWidth: 30, MinMainWidth: 60, SideRight: true}
	main, side, sideVisible := dp.compute(120, 40)
	if !sideVisible {
		t.Error("side should be visible at width 120")
	}
	if main.width != 89 {
		t.Errorf("expected main width 89, got %d", main.width)
	}
	if side.width != 30 {
		t.Errorf("expected side width 30, got %d", side.width)
	}
	if main.height != 40 || side.height != 40 {
		t.Errorf("heights should be 40, got main=%d side=%d", main.height, side.height)
	}
}

func TestDualPaneAutoHide(t *testing.T) {
	dp := DualPane{SideWidth: 30, MinMainWidth: 60, SideRight: true}
	main, _, sideVisible := dp.compute(80, 40)
	if sideVisible {
		t.Error("side should auto-hide at width 80")
	}
	if main.width != 80 {
		t.Errorf("expected main width 80 when side hidden, got %d", main.width)
	}
}

func TestDualPaneToggle(t *testing.T) {
	dp := DualPane{SideWidth: 30, MinMainWidth: 60}
	dp.sideHidden = true
	main, _, sideVisible := dp.compute(120, 40)
	if sideVisible {
		t.Error("side should be hidden when toggled off")
	}
	if main.width != 120 {
		t.Errorf("expected main width 120 when side toggled off, got %d", main.width)
	}
}
