package btest

import (
	"sync"
	"time"
)

// Clock is an abstraction over time.Now used by Poller-like components so
// tests can drive time deterministically. The real clock uses time.Now and
// time.Sleep; FakeClock lets tests advance time manually.
type Clock interface {
	// Now returns the current time as seen by this clock.
	Now() time.Time
	// Sleep blocks until the clock has advanced by d.
	// FakeClock implementations return immediately but still honor Advance.
	Sleep(d time.Duration)
}

// RealClock is a Clock backed by the real time package. Safe for concurrent use.
type RealClock struct{}

// Now returns time.Now.
func (RealClock) Now() time.Time { return time.Now() }

// Sleep calls time.Sleep.
func (RealClock) Sleep(d time.Duration) { time.Sleep(d) }

// FakeClock is a deterministic Clock for tests. Create one with NewFakeClock
// and advance it with Advance. Now and Sleep are safe for concurrent use.
type FakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*FakeTimer
}

// NewFakeClock returns a FakeClock anchored at the given time. If t is the
// zero value, it is anchored at a fixed epoch (2026-01-01 UTC) so tests are
// reproducible across machines.
func NewFakeClock(t time.Time) *FakeClock {
	if t.IsZero() {
		t = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return &FakeClock{now: t}
}

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Advance moves the fake clock forward by d and fires any timers whose
// deadline has been reached. Negative durations are ignored.
func (f *FakeClock) Advance(d time.Duration) {
	if d <= 0 {
		return
	}
	f.mu.Lock()
	f.now = f.now.Add(d)
	f.mu.Unlock()
	f.fireTimers()
}

// Set moves the fake clock to an absolute time.
func (f *FakeClock) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t
}

// Sleep advances the fake clock by d and returns immediately. It does not
// block, so tests remain fast. Use Advance for clarity when the intent is
// "time passes" rather than "goroutine sleeps".
func (f *FakeClock) Sleep(d time.Duration) {
	f.Advance(d)
}

// AfterFunc registers a function to be called when the clock advances past
// the deadline (Now() + d). Returns a Timer that can be stopped. The function
// fires during the next Advance call that crosses the deadline.
func (f *FakeClock) AfterFunc(d time.Duration, fn func()) *FakeTimer {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := &FakeTimer{
		deadline: f.now.Add(d),
		fn:       fn,
	}
	f.timers = append(f.timers, t)
	return t
}

// Pending returns the number of timers waiting to fire.
func (f *FakeClock) Pending() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, t := range f.timers {
		if !t.stopped {
			count++
		}
	}
	return count
}

// fireTimers fires any timers whose deadline has been reached. Must be called
// with the lock NOT held, as timer callbacks may call back into the clock.
func (f *FakeClock) fireTimers() {
	f.mu.Lock()
	var ready []*FakeTimer
	var remaining []*FakeTimer
	for _, t := range f.timers {
		if t.stopped {
			continue
		}
		if !f.now.Before(t.deadline) {
			ready = append(ready, t)
		} else {
			remaining = append(remaining, t)
		}
	}
	f.timers = remaining
	f.mu.Unlock()

	for _, t := range ready {
		t.fn()
	}
}

// FakeTimer is a timer registered with FakeClock.AfterFunc.
type FakeTimer struct {
	deadline time.Time
	fn       func()
	stopped  bool
}

// Stop prevents the timer from firing. Returns true if the timer was stopped
// before it fired.
func (t *FakeTimer) Stop() bool {
	if t.stopped {
		return false
	}
	t.stopped = true
	return true
}
