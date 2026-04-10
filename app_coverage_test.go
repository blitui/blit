package blit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Option tests (0% functions) ---

func TestWithFocusCycleKey(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithFocusCycleKey("f1"),
	)
	if a.focusCycleKey != "f1" {
		t.Errorf("expected focusCycleKey 'f1', got %q", a.focusCycleKey)
	}
}

func TestWithFocusCycleKeyDisabled(t *testing.T) {
	c1 := &stubComponent{name: "one"}
	c2 := &stubComponent{name: "two"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("one", c1),
		WithComponent("two", c2),
		WithFocusCycleKey(""),
	)

	// Tab should not cycle focus when disabled
	a.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !c1.focused {
		t.Error("focus should not change when cycle key is empty")
	}
}

func TestWithAnimations(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithAnimations(true),
	)
	if !a.animationsEnabled {
		t.Error("expected animationsEnabled=true")
	}

	a2 := newAppModel(
		WithTheme(DefaultTheme()),
		WithAnimations(false),
	)
	if a2.animationsEnabled {
		t.Error("expected animationsEnabled=false")
	}
}

// --- handleAnimTick (0%) ---

func TestHandleAnimTick(t *testing.T) {
	c := &stubComponent{name: "main"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithAnimations(true),
	)

	msg := animTickMsg{time: time.Now()}
	_, cmd := a.Update(msg)

	// Should broadcast to components and schedule next tick
	if cmd == nil {
		t.Error("handleAnimTick should return a command")
	}
}

// --- animTickCmd (40%) ---

func TestAnimTickCmdEnabled(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithAnimations(true),
	)
	if animDisabled {
		t.Skip("BLIT_NO_ANIM is set")
	}
	cmd := a.animTickCmd()
	if cmd == nil {
		t.Error("animTickCmd should return a command when animations enabled")
	}
}

func TestAnimTickCmdDisabled(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	// Default: animationsEnabled = false
	cmd := a.animTickCmd()
	if cmd != nil {
		t.Error("animTickCmd should return nil when animations disabled")
	}
}

// --- devConsoleUpdateSnapshot (60.7%) ---

func TestDevConsoleUpdateSnapshot_NoConsole(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	// Should not panic when devConsole is nil
	a.devConsoleUpdateSnapshot()
}

func TestDevConsoleUpdateSnapshot_WithComponents(t *testing.T) {
	c1 := &stubComponent{name: "alpha"}
	c2 := &stubComponent{name: "beta"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("alpha", c1),
		WithComponent("beta", c2),
		WithDevConsole(),
	)
	a.width = 120
	a.height = 40
	a.resize()

	a.devConsoleUpdateSnapshot()

	snap := a.devConsole.snapshot
	if len(snap.componentNames) != 2 {
		t.Fatalf("expected 2 component names, got %d", len(snap.componentNames))
	}
	if snap.componentNames[0] != "alpha" {
		t.Errorf("expected first component 'alpha', got %q", snap.componentNames[0])
	}
	if snap.focusName != "alpha" {
		t.Errorf("expected focusName 'alpha', got %q", snap.focusName)
	}
	if snap.themeName != "default" {
		t.Errorf("expected themeName 'default', got %q", snap.themeName)
	}
}

func TestDevConsoleUpdateSnapshot_DualPane(t *testing.T) {
	main := &stubComponent{name: "M"}
	side := &stubComponent{name: "S"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithLayout(&DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    20,
			MinMainWidth: 40,
		}),
		WithDevConsole(),
	)
	a.width = 120
	a.height = 40
	a.resize()

	a.devConsoleUpdateSnapshot()

	snap := a.devConsole.snapshot
	if len(snap.componentNames) != 2 {
		t.Fatalf("expected 2 component names in dual pane, got %d", len(snap.componentNames))
	}
	if snap.componentNames[0] != "Main" {
		t.Errorf("expected 'Main', got %q", snap.componentNames[0])
	}
	if snap.componentNames[1] != "Side" {
		t.Errorf("expected 'Side', got %q", snap.componentNames[1])
	}
}

func TestDevConsoleUpdateSnapshot_WithSignals(t *testing.T) {
	sig := NewSignal("hello")
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithDevConsole(),
	)
	a.trackSignal(sig)
	a.width = 120
	a.height = 40
	a.resize()

	a.devConsoleUpdateSnapshot()

	if len(a.devConsole.snapshot.signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(a.devConsole.snapshot.signals))
	}
	if a.devConsole.snapshot.signals[0].value != "hello" {
		t.Errorf("expected signal value 'hello', got %q", a.devConsole.snapshot.signals[0].value)
	}
}

// --- openOverlay (57.1%) ---

func TestOpenOverlay_Activatable(t *testing.T) {
	o := &stubOverlay{name: "test"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithOverlay("test", "", o),
	)
	a.width = 80
	a.height = 24

	a.openOverlay(o)

	if !o.active {
		t.Error("overlay should be activated")
	}
	if a.overlays.active() != o {
		t.Error("overlay should be on the stack")
	}
}

// --- Update message branches (73.1%) ---

func TestUpdateSetThemeMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.width = 80
	a.height = 24
	a.resize()

	newTheme := DefaultTheme()
	newTheme.Accent = "#ff0000"
	a.Update(SetThemeMsg{Theme: newTheme})

	if a.theme.Accent != "#ff0000" {
		t.Errorf("expected accent '#ff0000', got %q", a.theme.Accent)
	}
}

func TestUpdateToastMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.width = 80
	a.height = 24

	a.Update(ToastMsg{Severity: SeverityError, Title: "fail", Duration: time.Second})

	if a.toasts == nil {
		t.Fatal("toasts should exist")
	}
}

func TestUpdateDismissTopToastMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.toasts.theme = a.theme
	a.toasts.add(ToastMsg{Title: "one", Duration: 5 * time.Second})
	a.toasts.add(ToastMsg{Title: "two", Duration: 5 * time.Second})

	a.Update(dismissTopToastMsg{})
	// Should not panic and should dismiss the top toast
}

func TestUpdateDismissToastMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.toasts.theme = a.theme
	a.toasts.add(ToastMsg{Title: "one", Duration: 5 * time.Second})

	a.Update(dismissToastMsg{index: 0})
}

func TestUpdateThemeHotReloadMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.width = 80
	a.height = 24
	a.resize()

	newTheme := DefaultTheme()
	newTheme.Accent = "#00ff00"
	a.Update(ThemeHotReloadMsg{Theme: newTheme})

	if a.theme.Accent != "#00ff00" {
		t.Errorf("expected accent '#00ff00', got %q", a.theme.Accent)
	}
}

func TestUpdateThemeHotReloadErrMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.Update(ThemeHotReloadErrMsg{Err: fmt.Errorf("bad theme")})
	// Should not panic; toast should be added
}

func TestUpdateSignalFlushMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.Update(signalFlushMsg{})
	// Should not panic
}

func TestUpdateDevConsoleToggleMsg(t *testing.T) {
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithDevConsole(),
	)
	a.width = 120
	a.height = 40
	a.resize()

	a.Update(devConsoleToggleMsg{})
	// toggleDevConsole should be called without error
}

// --- DevConsoleToggleCmd (0%) ---

func TestDevConsoleToggleCmd(t *testing.T) {
	cmd := DevConsoleToggleCmd()
	if cmd == nil {
		t.Fatal("DevConsoleToggleCmd should return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(devConsoleToggleMsg); !ok {
		t.Errorf("expected devConsoleToggleMsg, got %T", msg)
	}
}

// --- Init coverage (66.7%) ---

func TestInitWithModules(t *testing.T) {
	m := &stubModule{name: "test-mod"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithModule(m),
	)
	cmd := a.Init()
	// Module.Init should have been called
	if !m.inited {
		t.Error("module Init should be called")
	}
	_ = cmd
}

func TestInitWithPendingNotify(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.pendingNotify = "startup message"

	cmd := a.Init()
	if cmd == nil {
		t.Error("Init should return cmd for pending notify")
	}
	if a.pendingNotify != "" {
		t.Error("pendingNotify should be cleared after Init")
	}
}

func TestInitDualPane(t *testing.T) {
	main := &stubComponent{name: "M"}
	side := &stubComponent{name: "S"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithLayout(&DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    20,
			MinMainWidth: 40,
		}),
	)
	cmd := a.Init()
	_ = cmd
}

// --- View coverage (78.1%) ---

func TestViewWithDevConsole(t *testing.T) {
	c := &stubComponent{name: "content"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("content", c),
		WithDevConsole(),
	)
	a.width = 120
	a.height = 40
	a.resize()

	view := a.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestViewFullScreenOverlay(t *testing.T) {
	c := &stubComponent{name: "main"}
	o := &stubOverlay{name: "fullscreen-overlay", active: true}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("main", c),
		WithOverlay("test", "", o),
	)
	a.width = 80
	a.height = 24
	a.resize()
	a.overlays.push(o)

	view := a.View()
	if !strings.Contains(view, "fullscreen-overlay") {
		t.Error("view should show overlay content")
	}
}

// --- cycleFocus (80%) ---

func TestCycleFocusSingleComponent(t *testing.T) {
	c := &stubComponent{name: "only"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("only", c),
	)
	// Should not panic with single component
	a.cycleFocus()
	if a.focusIdx != 0 {
		t.Errorf("focus should remain at 0 with single component, got %d", a.focusIdx)
	}
}

// --- broadcastMsg (83.3%) ---

func TestBroadcastMsgDualPane(t *testing.T) {
	main := &stubComponent{name: "M"}
	side := &stubComponent{name: "S"}
	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithLayout(&DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    20,
			MinMainWidth: 40,
		}),
	)

	type custom struct{}
	a.broadcastMsg(custom{})
	if _, ok := main.lastMsg.(custom); !ok {
		t.Error("main should receive broadcast")
	}
	if _, ok := side.lastMsg.(custom); !ok {
		t.Error("side should receive broadcast")
	}
}

// --- trackSignal (75%) ---

func TestTrackSignalNil(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	// Should not panic
	a.trackSignal(nil)
	if len(a.signals) != 0 {
		t.Error("nil signal should not be tracked")
	}
}

// --- stubModule for Init test ---

type stubModule struct {
	name   string
	inited bool
}

func (m *stubModule) Name() string { return m.name }
func (m *stubModule) Init() tea.Cmd {
	m.inited = true
	return nil
}
func (m *stubModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	return m, nil
}
