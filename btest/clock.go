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
	mu  sync.Mutex
	now time.Time
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

// Advance moves the fake clock forward by d. Negative durations are ignored.
func (f *FakeClock) Advance(d time.Duration) {
	if d <= 0 {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
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
