package blit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- recordFrame (0%) ---

func TestDevConsole_RecordFrame(t *testing.T) {
	dc := newDevConsole()
	now := time.Now()

	dc.recordFrame(now)
	if dc.frameCount != 1 {
		t.Errorf("expected frameCount 1, got %d", dc.frameCount)
	}

	dc.recordFrame(now.Add(16 * time.Millisecond))
	if dc.frameCount != 2 {
		t.Errorf("expected frameCount 2, got %d", dc.frameCount)
	}
}

func TestDevConsole_RecordFrameWraparound(t *testing.T) {
	dc := newDevConsole()
	now := time.Now()

	// Fill the entire ring buffer and then some
	for i := 0; i < 70; i++ {
		dc.recordFrame(now.Add(time.Duration(i) * 16 * time.Millisecond))
	}
	if dc.frameCount < 60 {
		t.Errorf("frameCount should cap at 60, got %d", dc.frameCount)
	}
}

// --- fps with multiple frames ---

func TestDevConsole_FpsWithFrames(t *testing.T) {
	dc := newDevConsole()
	now := time.Now()

	for i := 0; i < 10; i++ {
		dc.recordFrame(now.Add(time.Duration(i) * 16 * time.Millisecond))
	}
	fps := dc.fps()
	if fps <= 0 {
		t.Errorf("expected positive FPS, got %f", fps)
	}
}

// --- frameTimeMs (40%) ---

func TestDevConsole_FrameTimeMs(t *testing.T) {
	dc := newDevConsole()
	now := time.Now()

	dc.recordFrame(now)
	dc.recordFrame(now.Add(16 * time.Millisecond))

	ft := dc.frameTimeMs()
	if ft <= 0 {
		t.Errorf("expected positive frame time, got %f", ft)
	}
}

func TestDevConsole_FrameTimeMsSingleFrame(t *testing.T) {
	dc := newDevConsole()
	dc.recordFrame(time.Now())

	ft := dc.frameTimeMs()
	if ft != 0 {
		t.Errorf("expected 0 with single frame, got %f", ft)
	}
}

// --- Focused / SetFocused / Close / SetActive (all 0%) ---

func TestDevConsole_Focused(t *testing.T) {
	dc := newDevConsole()
	if dc.Focused() {
		t.Error("should not be focused initially")
	}
	dc.SetFocused(true)
	if !dc.Focused() {
		t.Error("should be focused")
	}
}

func TestDevConsole_Close(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.Close()
	if dc.active {
		t.Error("Close should deactivate")
	}
}

func TestDevConsole_SetActive(t *testing.T) {
	dc := newDevConsole()
	dc.SetActive(true)
	if !dc.active {
		t.Error("SetActive(true) should activate")
	}
	dc.SetActive(false)
	if dc.active {
		t.Error("SetActive(false) should deactivate")
	}
}

// --- Init (0%) ---

func TestDevConsole_Init(t *testing.T) {
	dc := newDevConsole()
	cmd := dc.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

// --- Update coverage (46.5%) ---

func TestDevConsole_UpdateEsc(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	dc.Update(tea.KeyMsg{Type: tea.KeyEsc}, ctx)
	if dc.active {
		t.Error("Esc should deactivate dev console")
	}
}

func TestDevConsole_UpdateCtrlBackslash(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	dc.Update(tea.KeyMsg{Type: tea.KeyCtrlBackslash}, ctx)
	if dc.active {
		t.Error("ctrl+\\ should deactivate dev console")
	}
}

func TestDevConsole_UpdateNumberKeyOutOfRange(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	dc.activeTab = 2
	// Key "9" is out of range (only 6 providers), should not change tab
	dc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")}, ctx)
	if dc.activeTab != 2 {
		t.Errorf("out-of-range number key should not change tab, got %d", dc.activeTab)
	}
}

func TestDevConsole_UpdateWhenInactive(t *testing.T) {
	dc := newDevConsole()
	dc.active = false
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	// When inactive, key messages should pass through
	dc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}, ctx)
	// activeTab should not change when inactive
	if dc.activeTab != 0 {
		t.Errorf("inactive console should not handle keys, got tab %d", dc.activeTab)
	}
}

// --- SetSize edge cases ---

func TestDevConsole_SetSizeSmall(t *testing.T) {
	dc := newDevConsole()
	dc.SetSize(30, 10)
	// w should clamp to min 40
	if dc.w < 40 {
		t.Errorf("expected min w=40, got %d", dc.w)
	}
	// h should clamp to min 12
	if dc.h < 12 {
		t.Errorf("expected min h=12, got %d", dc.h)
	}
}

func TestDevConsole_SetSizeAlreadySet(t *testing.T) {
	dc := newDevConsole()
	dc.SetSize(120, 40)
	origW, origH := dc.w, dc.h
	// Second call should not reset if w > 0
	dc.SetSize(200, 60)
	if dc.w != origW || dc.h != origH {
		t.Error("SetSize should not reset existing panel dimensions")
	}
}

// --- View rendering paths ---

func TestDevConsole_ViewInactive(t *testing.T) {
	dc := newDevConsole()
	dc.active = false
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)

	view := dc.View()
	if view != "" {
		t.Error("inactive devConsole should render empty string")
	}
}

// --- Update: resize and move keys ---

func TestDevConsole_UpdateResize(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	origW, origH := dc.w, dc.h

	dc.Update(tea.KeyMsg{Type: tea.KeyUp, Alt: true}, ctx)
	if dc.h != origH-1 {
		t.Errorf("alt+up should shrink h, got %d", dc.h)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyDown, Alt: true}, ctx)
	if dc.h != origH {
		t.Errorf("alt+down should grow h, got %d", dc.h)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyLeft, Alt: true}, ctx)
	if dc.w != origW-1 {
		t.Errorf("alt+left should shrink w, got %d", dc.w)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: true}, ctx)
	if dc.w != origW {
		t.Errorf("alt+right should grow w, got %d", dc.w)
	}
}

func TestDevConsole_UpdateMove(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	origX, origY := dc.x, dc.y

	dc.Update(tea.KeyMsg{Type: tea.KeyShiftDown}, ctx)
	if dc.y != origY+1 {
		t.Errorf("shift+down should move y, got %d", dc.y)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyShiftUp}, ctx)
	if dc.y != origY {
		t.Errorf("shift+up should move y back, got %d", dc.y)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyShiftRight}, ctx)
	if dc.x != origX+1 {
		t.Errorf("shift+right should move x, got %d", dc.x)
	}

	dc.Update(tea.KeyMsg{Type: tea.KeyShiftLeft}, ctx)
	if dc.x != origX {
		t.Errorf("shift+left should move x back, got %d", dc.x)
	}
}

func TestDevConsole_UpdateMouseDrag(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	dc.Update(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionMotion,
		X:      10, Y: 5,
	}, ctx)
	if dc.x != 10 || dc.y != 5 {
		t.Errorf("expected x=10 y=5, got x=%d y=%d", dc.x, dc.y)
	}
}
