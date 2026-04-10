package btest

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent by TriggerTick and AdvanceClock to simulate timer events.
// Components can type-assert on this message to handle ticks deterministically
// in tests.
type TickMsg struct {
	Time time.Time
}

// TestModel wraps a tea.Model for easy testing.
type TestModel struct {
	t     testing.TB
	model tea.Model
	cols  int
	lines int
	clock *FakeClock
}

// NewTestModel creates a test wrapper around a Bubble Tea model.
// Calls Init() and processes any returned commands that produce messages.
func NewTestModel(t testing.TB, model tea.Model, cols, lines int) *TestModel {
	t.Helper()
	cmd := model.Init()
	if cmd != nil {
		if msg := cmd(); msg != nil {
			model, _ = model.Update(msg)
		}
	}
	// Send an initial WindowSizeMsg so the model knows its dimensions.
	model, _ = model.Update(tea.WindowSizeMsg{Width: cols, Height: lines})
	return &TestModel{
		t:     t,
		model: model,
		cols:  cols,
		lines: lines,
	}
}

// SendKey sends a key event to the model. Supports special key names like
// "enter", "tab", "up", "down", "left", "right", "esc", "backspace",
// "space", and single characters.
func (tm *TestModel) SendKey(key string) {
	tm.t.Helper()
	msg := keyToMsg(key)
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// SendMouse sends a mouse event to the model.
func (tm *TestModel) SendMouse(x, y int, button tea.MouseButton) {
	tm.t.Helper()
	msg := tea.MouseMsg{X: x, Y: y, Button: button}
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// SendResize sends a window resize event.
func (tm *TestModel) SendResize(cols, lines int) {
	tm.t.Helper()
	tm.cols = cols
	tm.lines = lines
	msg := tea.WindowSizeMsg{Width: cols, Height: lines}
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// Type sends a sequence of key events (one per character).
func (tm *TestModel) Type(text string) {
	tm.t.Helper()
	for _, ch := range text {
		tm.SendKey(string(ch))
	}
}

// Screen renders the model's View() and returns the parsed screen.
func (tm *TestModel) Screen() *Screen {
	s := NewScreen(tm.cols, tm.lines)
	s.Render(tm.model.View())
	return s
}

// Model returns the current underlying tea.Model.
func (tm *TestModel) Model() tea.Model {
	return tm.model
}

// SendMsg sends an arbitrary tea.Msg to the model and processes any resulting command.
func (tm *TestModel) SendMsg(msg tea.Msg) {
	tm.t.Helper()
	var cmd tea.Cmd
	tm.model, cmd = tm.model.Update(msg)
	tm.processCmd(cmd)
}

// SendKeys sends a sequence of named keys. Each key is processed individually.
// Example: tm.SendKeys("down", "down", "enter")
func (tm *TestModel) SendKeys(keys ...string) {
	tm.t.Helper()
	for _, k := range keys {
		tm.SendKey(k)
	}
}

// WithClock attaches a FakeClock to this TestModel. AdvanceClock and
// TriggerTick will use it. Returns the TestModel for chaining.
func (tm *TestModel) WithClock(clock *FakeClock) *TestModel {
	tm.clock = clock
	return tm
}

// Clock returns the attached FakeClock, or nil if none was set.
func (tm *TestModel) Clock() *FakeClock {
	return tm.clock
}

// AdvanceClock advances the attached FakeClock by d and sends a TickMsg to the
// model. Any timers registered with AfterFunc that cross their deadline will
// fire. Panics if no clock is attached.
func (tm *TestModel) AdvanceClock(d time.Duration) {
	tm.t.Helper()
	if tm.clock == nil {
		tm.t.Fatal("btest: AdvanceClock called but no clock attached (use WithClock)")
	}
	tm.clock.Advance(d)
	tm.SendMsg(TickMsg{Time: tm.clock.Now()})
}

// TriggerTick sends a TickMsg with the current clock time (or time.Now if no
// clock is attached). Use this to simulate a timer/poller tick.
func (tm *TestModel) TriggerTick() {
	tm.t.Helper()
	var now time.Time
	if tm.clock != nil {
		now = tm.clock.Now()
	} else {
		now = time.Now()
	}
	tm.SendMsg(TickMsg{Time: now})
}

// SendTick sends a generic tick message. Useful for testing time-based
// components like pollers, spinners, and animations.
//
// Deprecated: Use TriggerTick or AdvanceClock instead.
func (tm *TestModel) SendTick() {
	tm.t.Helper()
	tm.TriggerTick()
}

// WaitFor repeatedly sends tick messages until the predicate returns true or
// maxTicks is reached. Returns true if the predicate was satisfied. This is
// useful for waiting on async state changes driven by tea.Cmd chains.
func (tm *TestModel) WaitFor(predicate func(*Screen) bool, maxTicks int) bool {
	tm.t.Helper()
	for i := 0; i < maxTicks; i++ {
		if predicate(tm.Screen()) {
			return true
		}
		// Process any pending commands by sending an empty message.
		tm.SendMsg(nil)
	}
	return predicate(tm.Screen())
}

// RequireScreen renders the model and runs assertions against the screen in one call.
// The callback receives the screen and can use assert helpers on it.
func (tm *TestModel) RequireScreen(fn func(t testing.TB, s *Screen)) {
	tm.t.Helper()
	fn(tm.t, tm.Screen())
}

// Cols returns the current terminal width.
func (tm *TestModel) Cols() int { return tm.cols }

// Lines returns the current terminal height.
func (tm *TestModel) Lines() int { return tm.lines }

// processCmd executes a command and feeds its message back into the model,
// looping until no more commands are produced (or maxProcessCmdIterations is
// hit, in which case the test fails). Handles tea.BatchMsg by flattening all
// sub-commands into the work queue so nested batches resolve correctly.
func (tm *TestModel) processCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	queue := []tea.Cmd{cmd}
	iter := 0
	for len(queue) > 0 {
		iter++
		if iter > maxProcessCmdIterations {
			tm.t.Helper()
			tm.t.Fatalf("btest: processCmd did not quiesce after %d iterations "+
				"(likely an infinite command loop)", maxProcessCmdIterations)
			return
		}
		cur := queue[0]
		queue = queue[1:]
		if cur == nil {
			continue
		}
		msg := cur()
		if msg == nil {
			continue
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, batch...)
			continue
		}
		var nextCmd tea.Cmd
		tm.model, nextCmd = tm.model.Update(msg)
		if nextCmd != nil {
			queue = append(queue, nextCmd)
		}
	}
}

// maxProcessCmdIterations caps processCmd's work loop so a buggy model with a
// command that keeps returning new commands fails the test quickly rather
// than hanging.
const maxProcessCmdIterations = 64

// keyToMsg converts a key name string to a tea.KeyMsg.
func keyToMsg(key string) tea.KeyMsg {
	if kt, ok := keyMap[key]; ok {
		return tea.KeyMsg{Type: kt}
	}
	runes := []rune(key)
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
}

// keyMap provides named key → tea.KeyType mappings for all common keys.
var keyMap = map[string]tea.KeyType{
	// Navigation
	"enter":     tea.KeyEnter,
	"tab":       tea.KeyTab,
	"shift+tab": tea.KeyShiftTab,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"delete":    tea.KeyDelete,
	"insert":    tea.KeyInsert,
	"backspace": tea.KeyBackspace,
	"esc":       tea.KeyEscape,
	"escape":    tea.KeyEscape,
	"space":     tea.KeySpace,

	// Ctrl combinations
	"ctrl+a": tea.KeyCtrlA,
	"ctrl+b": tea.KeyCtrlB,
	"ctrl+c": tea.KeyCtrlC,
	"ctrl+d": tea.KeyCtrlD,
	"ctrl+e": tea.KeyCtrlE,
	"ctrl+f": tea.KeyCtrlF,
	"ctrl+g": tea.KeyCtrlG,
	"ctrl+h": tea.KeyCtrlH,
	"ctrl+j": tea.KeyCtrlJ,
	"ctrl+k": tea.KeyCtrlK,
	"ctrl+l": tea.KeyCtrlL,
	"ctrl+n": tea.KeyCtrlN,
	"ctrl+o": tea.KeyCtrlO,
	"ctrl+p": tea.KeyCtrlP,
	"ctrl+q": tea.KeyCtrlQ,
	"ctrl+r": tea.KeyCtrlR,
	"ctrl+s": tea.KeyCtrlS,
	"ctrl+t": tea.KeyCtrlT,
	"ctrl+u": tea.KeyCtrlU,
	"ctrl+v": tea.KeyCtrlV,
	"ctrl+w": tea.KeyCtrlW,
	"ctrl+x": tea.KeyCtrlX,
	"ctrl+y": tea.KeyCtrlY,
	"ctrl+z": tea.KeyCtrlZ,

	// Function keys
	"f1":  tea.KeyF1,
	"f2":  tea.KeyF2,
	"f3":  tea.KeyF3,
	"f4":  tea.KeyF4,
	"f5":  tea.KeyF5,
	"f6":  tea.KeyF6,
	"f7":  tea.KeyF7,
	"f8":  tea.KeyF8,
	"f9":  tea.KeyF9,
	"f10": tea.KeyF10,
	"f11": tea.KeyF11,
	"f12": tea.KeyF12,
}

// KeyMsgForTesting converts a key name to a tea.KeyMsg for use in unit tests
// that call Component.Update directly without a full TestModel.
func KeyMsgForTesting(key string) tea.KeyMsg {
	return keyToMsg(key)
}

// KeyNames returns the list of all recognized key names for documentation.
func KeyNames() []string {
	names := make([]string, 0, len(keyMap))
	for k := range keyMap {
		names = append(names, k)
	}
	return names
}

// Convenience type alias for terser test code.
type screenPredicate = func(*Screen) bool

// UntilContains returns a predicate that is satisfied when the screen contains text.
func UntilContains(text string) screenPredicate {
	return func(s *Screen) bool { return s.Contains(text) }
}

// UntilNotContains returns a predicate satisfied when the screen no longer contains text.
func UntilNotContains(text string) screenPredicate {
	return func(s *Screen) bool { return !s.Contains(text) }
}

// UntilRowContains returns a predicate satisfied when the given row contains text.
func UntilRowContains(row int, text string) screenPredicate {
	return func(s *Screen) bool {
		return strings.Contains(s.Row(row), text)
	}
}

// Stopwatch measures elapsed time for performance assertions in tests.
type Stopwatch struct {
	start time.Time
}

// StartStopwatch begins a new timing measurement.
func StartStopwatch() Stopwatch {
	return Stopwatch{start: time.Now()}
}

// Elapsed returns the duration since the stopwatch was started.
func (sw Stopwatch) Elapsed() time.Duration {
	return time.Since(sw.start)
}

// AssertUnder fails the test if the elapsed time exceeds the given duration.
func (sw Stopwatch) AssertUnder(t testing.TB, d time.Duration) {
	t.Helper()
	elapsed := sw.Elapsed()
	if elapsed > d {
		t.Errorf("operation took %v, want under %v", elapsed, d)
	}
}
