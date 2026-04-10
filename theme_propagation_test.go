package blit_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/btest"
)

// themedStub is a minimal Component that tracks SetTheme calls.
type themedStub struct {
	name       string
	theme      blit.Theme
	themeCalls int
	focused    bool
	width      int
	height     int
}

func (s *themedStub) Init() tea.Cmd                                              { return nil }
func (s *themedStub) Update(msg tea.Msg, ctx blit.Context) (blit.Component, tea.Cmd) { return s, nil }
func (s *themedStub) View() string                                               { return s.name }
func (s *themedStub) KeyBindings() []blit.KeyBind                                { return nil }
func (s *themedStub) SetSize(w, h int)                                           { s.width = w; s.height = h }
func (s *themedStub) Focused() bool                                              { return s.focused }
func (s *themedStub) SetFocused(f bool)                                          { s.focused = f }
func (s *themedStub) SetTheme(t blit.Theme)                                      { s.theme = t; s.themeCalls++ }

// themedOverlayStub is a minimal Overlay that tracks SetTheme calls.
type themedOverlayStub struct {
	themedStub
	active bool
}

func (o *themedOverlayStub) IsActive() bool { return o.active }
func (o *themedOverlayStub) Close()         { o.active = false }

// TestThemePropagation_WithComponent verifies that every component registered
// via WithComponent receives SetTheme during app setup.
func TestThemePropagation_WithComponent(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"single component", 1},
		{"three components", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubs := make([]*themedStub, tt.count)
			var opts []blit.Option
			opts = append(opts, blit.WithTheme(blit.DefaultTheme()))
			for i := range stubs {
				stubs[i] = &themedStub{name: "c"}
				opts = append(opts, blit.WithComponent("c", stubs[i]))
			}
			app := blit.NewApp(opts...)
			_ = app.Model() // ensure setup ran

			for i, s := range stubs {
				if s.themeCalls == 0 {
					t.Errorf("component %d: SetTheme not called", i)
				}
				if string(s.theme.Accent) != string(blit.DefaultTheme().Accent) {
					t.Errorf("component %d: theme accent = %q, want %q", i, s.theme.Accent, blit.DefaultTheme().Accent)
				}
			}
		})
	}
}

// TestThemePropagation_HotReload verifies that SetThemeMsg propagates to all
// components without losing their registration.
func TestThemePropagation_HotReload(t *testing.T) {
	c1 := &themedStub{name: "one"}
	c2 := &themedStub{name: "two"}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("one", c1),
		blit.WithComponent("two", c2),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Record initial theme call count after setup
	c1Before := c1.themeCalls
	c2Before := c2.themeCalls

	// Send a theme switch
	light := blit.LightTheme()
	h.Send(blit.SetThemeMsg{Theme: light})

	if c1.themeCalls <= c1Before {
		t.Error("c1: SetTheme not called on hot-reload")
	}
	if c2.themeCalls <= c2Before {
		t.Error("c2: SetTheme not called on hot-reload")
	}
	if string(c1.theme.Accent) != string(light.Accent) {
		t.Errorf("c1: theme accent = %q, want %q", c1.theme.Accent, light.Accent)
	}
	if string(c2.theme.Accent) != string(light.Accent) {
		t.Errorf("c2: theme accent = %q, want %q", c2.theme.Accent, light.Accent)
	}
}

// TestThemePropagation_StatePersistence verifies that switching themes does
// not reset component state such as table cursor position.
func TestThemePropagation_StatePersistence(t *testing.T) {
	cols := []blit.Column{{Title: "Name", Width: 20}}
	rows := []blit.Row{{"Alice"}, {"Bob"}, {"Carol"}, {"Dave"}}
	tbl := blit.NewTable(cols, rows, blit.TableOpts{})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("table", tbl),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Move cursor down twice
	h.Keys("down", "down")
	cursorBefore := tbl.CursorIndex()
	if cursorBefore != 2 {
		t.Fatalf("cursor before theme switch = %d, want 2", cursorBefore)
	}

	// Switch theme
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})

	cursorAfter := tbl.CursorIndex()
	if cursorAfter != cursorBefore {
		t.Errorf("cursor after theme switch = %d, want %d", cursorAfter, cursorBefore)
	}
}

// TestThemePropagation_DualPane verifies that theme propagates to both
// Main and Side components in a DualPane layout.
func TestThemePropagation_DualPane(t *testing.T) {
	main := &themedStub{name: "main"}
	side := &themedStub{name: "side"}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithLayout(&blit.DualPane{
			Main:      main,
			Side:      side,
			SideWidth: 20,
		}),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if main.themeCalls == 0 {
		t.Error("DualPane.Main: SetTheme not called")
	}
	if side.themeCalls == 0 {
		t.Error("DualPane.Side: SetTheme not called")
	}

	// Hot-reload should propagate to both panes
	mainBefore := main.themeCalls
	sideBefore := side.themeCalls
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})

	if main.themeCalls <= mainBefore {
		t.Error("DualPane.Main: SetTheme not called on hot-reload")
	}
	if side.themeCalls <= sideBefore {
		t.Error("DualPane.Side: SetTheme not called on hot-reload")
	}
}

// TestThemePropagation_Split verifies that Split propagates theme to both
// child panes A and B.
func TestThemePropagation_Split(t *testing.T) {
	a := &themedStub{name: "paneA"}
	b := &themedStub{name: "paneB"}

	split := &blit.Split{
		A:     a,
		B:     b,
		Ratio: 0.5,
	}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("split", split),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if a.themeCalls == 0 {
		t.Error("Split.A: SetTheme not called")
	}
	if b.themeCalls == 0 {
		t.Error("Split.B: SetTheme not called")
	}

	// Hot-reload
	aBefore := a.themeCalls
	bBefore := b.themeCalls
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})

	if a.themeCalls <= aBefore {
		t.Error("Split.A: SetTheme not called on hot-reload")
	}
	if b.themeCalls <= bBefore {
		t.Error("Split.B: SetTheme not called on hot-reload")
	}
}

// TestThemePropagation_HBox verifies that HBox propagates theme to children.
func TestThemePropagation_HBox(t *testing.T) {
	c1 := &themedStub{name: "left"}
	c2 := &themedStub{name: "right"}

	hbox := &blit.HBox{Items: []blit.Component{c1, c2}}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("hbox", hbox),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if c1.themeCalls == 0 {
		t.Error("HBox child 0: SetTheme not called")
	}
	if c2.themeCalls == 0 {
		t.Error("HBox child 1: SetTheme not called")
	}

	// Hot-reload
	c1Before := c1.themeCalls
	c2Before := c2.themeCalls
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})

	if c1.themeCalls <= c1Before {
		t.Error("HBox child 0: SetTheme not called on hot-reload")
	}
	if c2.themeCalls <= c2Before {
		t.Error("HBox child 1: SetTheme not called on hot-reload")
	}
}

// TestThemePropagation_VBox verifies that VBox propagates theme to children.
func TestThemePropagation_VBox(t *testing.T) {
	c1 := &themedStub{name: "top"}
	c2 := &themedStub{name: "bottom"}

	vbox := &blit.VBox{Items: []blit.Component{c1, c2}}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("vbox", vbox),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if c1.themeCalls == 0 {
		t.Error("VBox child 0: SetTheme not called")
	}
	if c2.themeCalls == 0 {
		t.Error("VBox child 1: SetTheme not called")
	}
}

// TestThemePropagation_FlexWrapped verifies that Flex-wrapped children
// inside HBox receive theme propagation.
func TestThemePropagation_FlexWrapped(t *testing.T) {
	c1 := &themedStub{name: "flex1"}
	c2 := &themedStub{name: "flex2"}

	hbox := &blit.HBox{Items: []blit.Component{
		blit.Flex{Grow: 1, C: c1},
		blit.Flex{Grow: 2, C: c2},
	}}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("hbox", hbox),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if c1.themeCalls == 0 {
		t.Error("Flex child 0: SetTheme not called")
	}
	if c2.themeCalls == 0 {
		t.Error("Flex child 1: SetTheme not called")
	}
}

// TestThemePropagation_Overlay verifies that named overlays receive theme
// on registration.
func TestThemePropagation_Overlay(t *testing.T) {
	overlay := &themedOverlayStub{themedStub: themedStub{name: "overlay"}}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithOverlay("test", "o", overlay),
	)
	_ = app.Model()

	if overlay.themeCalls == 0 {
		t.Error("overlay: SetTheme not called on registration")
	}
	if string(overlay.theme.Accent) != string(blit.DefaultTheme().Accent) {
		t.Errorf("overlay: theme accent = %q, want %q", overlay.theme.Accent, blit.DefaultTheme().Accent)
	}
}

// TestThemePropagation_DarkLightRoundTrip verifies that switching from dark
// to light and back preserves table cursor position and component state.
func TestThemePropagation_DarkLightRoundTrip(t *testing.T) {
	cols := []blit.Column{{Title: "ID", Width: 10}}
	rows := []blit.Row{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}}
	tbl := blit.NewTable(cols, rows, blit.TableOpts{})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("tbl", tbl),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Move cursor to row 3
	h.Keys("down", "down", "down")
	if tbl.CursorIndex() != 3 {
		t.Fatalf("cursor = %d, want 3", tbl.CursorIndex())
	}

	// Dark -> Light -> Dark
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})
	if tbl.CursorIndex() != 3 {
		t.Errorf("cursor after light switch = %d, want 3", tbl.CursorIndex())
	}

	h.Send(blit.SetThemeMsg{Theme: blit.DefaultTheme()})
	if tbl.CursorIndex() != 3 {
		t.Errorf("cursor after dark switch = %d, want 3", tbl.CursorIndex())
	}
}

// TestThemePropagation_StatusBar verifies the status bar receives theme.
func TestThemePropagation_StatusBar(t *testing.T) {
	// StatusBar is internal to the app, so we verify indirectly by ensuring
	// the app renders without panic after a theme switch with a status bar.
	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithStatusBar(
			func() string { return "left" },
			func() string { return "right" },
		),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Switch theme — should not panic
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})
	h.Expect("left")
}

// TestThemePropagation_NestedSplitInDualPane verifies theme propagation
// through a Split nested inside a DualPane.
func TestThemePropagation_NestedSplitInDualPane(t *testing.T) {
	inner1 := &themedStub{name: "inner1"}
	inner2 := &themedStub{name: "inner2"}
	side := &themedStub{name: "side"}

	split := &blit.Split{A: inner1, B: inner2, Ratio: 0.5}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithLayout(&blit.DualPane{
			Main:      split,
			Side:      side,
			SideWidth: 20,
		}),
	)
	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	if inner1.themeCalls == 0 {
		t.Error("nested Split.A: SetTheme not called")
	}
	if inner2.themeCalls == 0 {
		t.Error("nested Split.B: SetTheme not called")
	}
	if side.themeCalls == 0 {
		t.Error("DualPane.Side: SetTheme not called")
	}

	// Hot-reload should reach nested children
	inner1Before := inner1.themeCalls
	h.Send(blit.SetThemeMsg{Theme: blit.LightTheme()})

	if inner1.themeCalls <= inner1Before {
		t.Error("nested Split.A: SetTheme not called on hot-reload")
	}
}
