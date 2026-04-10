package btest

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Harness is a fluent wrapper around TestModel for concise, chainable test
// scripts. Each method returns *Harness so calls can be chained:
//
//	btest.NewHarness(t, model, 80, 24).
//	    Keys("down", "down", "enter").
//	    Expect("Loaded").
//	    ExpectRow(2, "selected").
//	    Done()
type Harness struct {
	t        testing.TB
	tm       *TestModel
	setup    []func()
	teardown []func()
	done     bool
}

// NewHarness creates a new test harness around the given model.
func NewHarness(t testing.TB, model tea.Model, cols, lines int) *Harness {
	t.Helper()
	return &Harness{
		t:  t,
		tm: NewTestModel(t, model, cols, lines),
	}
}

// TestModel returns the underlying TestModel for calls not covered by the
// fluent API. Escape hatch — prefer to add a method on Harness if possible.
func (h *Harness) TestModel() *TestModel { return h.tm }

// OnSetup registers a function to run once when the first action fires.
// Use this for things like seeding state that depends on Harness being
// wired up but shouldn't run in NewHarness.
func (h *Harness) OnSetup(fn func()) *Harness {
	h.setup = append(h.setup, fn)
	return h
}

// OnTeardown registers a function to run when Done() is called. Registered
// functions run in reverse order (last-registered first) so they match a
// deferred-cleanup mental model.
func (h *Harness) OnTeardown(fn func()) *Harness {
	h.teardown = append(h.teardown, fn)
	return h
}

func (h *Harness) runSetup() {
	for _, fn := range h.setup {
		fn()
	}
	h.setup = nil
}

// Keys sends a sequence of named keys (down, enter, ctrl+c, a…).
func (h *Harness) Keys(keys ...string) *Harness {
	h.t.Helper()
	h.runSetup()
	h.tm.SendKeys(keys...)
	return h
}

// Type sends each character in text as an individual key event.
func (h *Harness) Type(text string) *Harness {
	h.t.Helper()
	h.runSetup()
	h.tm.Type(text)
	return h
}

// Click dispatches a left-click at (x, y).
func (h *Harness) Click(x, y int) *Harness {
	h.t.Helper()
	h.runSetup()
	h.tm.SendMouse(x, y, tea.MouseButtonLeft)
	return h
}

// Resize updates the simulated terminal size.
func (h *Harness) Resize(cols, lines int) *Harness {
	h.t.Helper()
	h.runSetup()
	h.tm.SendResize(cols, lines)
	return h
}

// Send dispatches an arbitrary tea.Msg.
func (h *Harness) Send(msg tea.Msg) *Harness {
	h.t.Helper()
	h.runSetup()
	h.tm.SendMsg(msg)
	return h
}

// Expect asserts that the current screen contains text anywhere. Fails the
// test with a helpful diff if not.
func (h *Harness) Expect(text string) *Harness {
	h.t.Helper()
	scr := h.tm.Screen()
	if !scr.Contains(text) {
		h.t.Errorf("Expect(%q): not found on screen:\n%s", text, scr.String())
	}
	return h
}

// ExpectNot asserts that text is NOT present on the current screen.
func (h *Harness) ExpectNot(text string) *Harness {
	h.t.Helper()
	scr := h.tm.Screen()
	if scr.Contains(text) {
		h.t.Errorf("ExpectNot(%q): unexpectedly found on screen:\n%s", text, scr.String())
	}
	return h
}

// ExpectRow asserts that the given row contains text.
func (h *Harness) ExpectRow(row int, text string) *Harness {
	h.t.Helper()
	got := h.tm.Screen().Row(row)
	if !strings.Contains(got, text) {
		h.t.Errorf("ExpectRow(%d, %q): got %q", row, text, got)
	}
	return h
}

// Screen returns the current screen for ad-hoc assertions outside the
// fluent API.
func (h *Harness) Screen() *Screen { return h.tm.Screen() }

// Snapshot stores or compares a screen snapshot under the given name.
func (h *Harness) Snapshot(name string) *Harness {
	h.t.Helper()
	AssertSnapshot(h.t, h.tm.Screen(), name)
	return h
}

// Advance is a placeholder for time-based drivers. It accepts a duration
// so call sites read naturally; the underlying TestModel does not yet
// integrate a FakeClock directly, but tests that own the FakeClock can
// chain .Advance(d) purely for documentation.
func (h *Harness) Advance(d time.Duration) *Harness {
	h.t.Helper()
	// Send a zero-valued tick so any time-driven Update paths fire.
	h.tm.SendMsg(TickMsgPlaceholder{At: time.Now().Add(d)})
	return h
}

// TickMsgPlaceholder is a btest-local tick shape used by Harness.Advance
// so we don't depend on an external TickMsg definition here. Consumers
// that care about tick semantics should use SendMsg directly.
type TickMsgPlaceholder struct {
	At time.Time
}

// NewAppHarness creates a test harness from a full application, using the real
// component tree (Tabs, DualPane, StatusBar, overlays, etc.). This ensures
// golden snapshots cover the actual production layout, not a simplified
// reconstruction. app must implement Model() tea.Model (e.g., *blit.App).
//
//	h := btest.NewAppHarness(t, myApp, 80, 24)
//	defer h.Done()
//	h.Send(someMsg).Expect("loaded").Snapshot("initial")
func NewAppHarness(t testing.TB, app interface{ Model() tea.Model }, cols, lines int) *Harness {
	t.Helper()
	return NewHarness(t, app.Model(), cols, lines)
}

// SnapshotApp is a one-liner that creates a harness from a full application,
// takes a named golden snapshot, and cleans up. Use this when you only need to
// verify the initial render of the real app layout:
//
//	btest.SnapshotApp(t, myApp, 80, 24, "initial")
//
// For interactive tests (key presses, resizes, multiple snapshots), use
// NewAppHarness instead.
func SnapshotApp(t testing.TB, app interface{ Model() tea.Model }, cols, lines int, name string) {
	t.Helper()
	h := NewAppHarness(t, app, cols, lines)
	defer h.Done()
	h.Snapshot(name)
}

// Done runs registered teardown callbacks in LIFO order. Safe to call
// multiple times; subsequent calls are no-ops.
func (h *Harness) Done() {
	if h.done {
		return
	}
	h.done = true
	for i := len(h.teardown) - 1; i >= 0; i-- {
		h.teardown[i]()
	}
}
