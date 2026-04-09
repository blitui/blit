package tuikit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeClock is a tiny Clock implementation used only by poller_test.
// We avoid importing tuitest here to keep the dependency one-way.
type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time        { return f.now }
func (f *fakeClock) Advance(d time.Duration) { f.now = f.now.Add(d) }

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func dummyCmd() tea.Cmd {
	return func() tea.Msg { return "polled" }
}

func TestPoller_NewPollerUsesRealClock(t *testing.T) {
	p := NewPoller(time.Second, dummyCmd)
	if p.clock == nil {
		t.Fatal("NewPoller did not set a clock")
	}
	if _, ok := p.clock.(realClock); !ok {
		t.Errorf("NewPoller clock type = %T, want realClock", p.clock)
	}
}

func TestPoller_NilClockDefaultsReal(t *testing.T) {
	p := NewPollerWithClock(time.Second, dummyCmd, nil)
	if _, ok := p.clock.(realClock); !ok {
		t.Errorf("nil clock not replaced with realClock, got %T", p.clock)
	}
}

func TestPoller_CheckFiresAfterInterval(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(100*time.Millisecond, dummyCmd, c)
	// First call: lastPoll is zero, elapsed is huge, should fire.
	if cmd := p.Check(TickMsg{}); cmd == nil {
		t.Fatal("first Check should return a cmd")
	}
	// Second call immediately: interval not elapsed, no fire.
	if cmd := p.Check(TickMsg{}); cmd != nil {
		t.Fatal("Check should not fire before interval elapses")
	}
	// Advance 99ms: still not enough.
	c.Advance(99 * time.Millisecond)
	if cmd := p.Check(TickMsg{}); cmd != nil {
		t.Fatal("Check fired at 99ms, interval is 100ms")
	}
	// Advance 1ms more: now at 100ms, should fire.
	c.Advance(1 * time.Millisecond)
	if cmd := p.Check(TickMsg{}); cmd == nil {
		t.Fatal("Check should fire at 100ms elapsed")
	}
}

func TestPoller_Pause(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(100*time.Millisecond, dummyCmd, c)
	p.Check(TickMsg{}) // establish lastPoll
	p.Pause()
	if !p.IsPaused() {
		t.Fatal("IsPaused should be true after Pause")
	}
	c.Advance(5 * time.Second)
	if cmd := p.Check(TickMsg{}); cmd != nil {
		t.Fatal("paused poller should not fire")
	}
}

func TestPoller_ResumeFires(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(50*time.Millisecond, dummyCmd, c)
	p.Check(TickMsg{})
	p.Pause()
	c.Advance(1 * time.Second)
	p.Resume()
	if p.IsPaused() {
		t.Fatal("Resume did not clear paused flag")
	}
	if cmd := p.Check(TickMsg{}); cmd == nil {
		t.Fatal("Check after Resume should fire because interval elapsed")
	}
}

func TestPoller_TogglePause(t *testing.T) {
	p := NewPoller(time.Second, dummyCmd)
	if p.IsPaused() {
		t.Fatal("new poller should not be paused")
	}
	p.TogglePause()
	if !p.IsPaused() {
		t.Fatal("TogglePause 1 should pause")
	}
	p.TogglePause()
	if p.IsPaused() {
		t.Fatal("TogglePause 2 should unpause")
	}
}

func TestPoller_ForceRefreshBypassesPause(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(time.Hour, dummyCmd, c)
	p.Pause()
	p.ForceRefresh()
	cmd := p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("ForceRefresh should fire even when paused")
	}
	// ForceRefresh is one-shot.
	if cmd := p.Check(TickMsg{}); cmd != nil {
		t.Fatal("ForceRefresh should be one-shot")
	}
}

func TestPoller_ForceRefreshUpdatesLastPoll(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(100*time.Millisecond, dummyCmd, c)
	p.ForceRefresh()
	before := c.Now()
	p.Check(TickMsg{})
	if !p.LastPoll().Equal(before) {
		t.Errorf("LastPoll = %v, want %v", p.LastPoll(), before)
	}
}

func TestPoller_SetInterval(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithClock(time.Hour, dummyCmd, c)
	p.Check(TickMsg{}) // fire once
	p.SetInterval(10 * time.Millisecond)
	c.Advance(15 * time.Millisecond)
	if cmd := p.Check(TickMsg{}); cmd == nil {
		t.Fatal("Check should fire after SetInterval shortened the interval")
	}
}

func TestPoller_LastPollZeroOnCreate(t *testing.T) {
	p := NewPoller(time.Second, dummyCmd)
	if !p.LastPoll().IsZero() {
		t.Errorf("new poller LastPoll = %v, want zero", p.LastPoll())
	}
}
