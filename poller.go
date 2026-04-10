package blit

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Clock abstracts time.Now so Poller can be tested deterministically.
// Production code uses realClock (via NewPoller). Tests can pass a
// FakeClock from the blit package (or any type that implements this
// interface).
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Poller manages periodic command execution on top of blit's TickMsg.
// Create one with NewPoller, then call Check from your component's Update
// when it receives a TickMsg. The Poller handles interval timing, pause/resume,
// and force-refresh.
type Poller struct {
	interval     time.Duration
	cmd          func() tea.Cmd
	lastPoll     time.Time
	paused       bool
	needsRefresh bool
	clock        Clock
}

// NewPoller creates a Poller that runs cmd at the given interval using the
// real system clock. For deterministic tests, use NewPollerWithClock.
func NewPoller(interval time.Duration, cmd func() tea.Cmd) *Poller {
	return NewPollerWithClock(interval, cmd, realClock{})
}

// NewPollerWithClock creates a Poller that uses the supplied Clock for all
// timing decisions. Pass a FakeClock in tests to advance time deterministically.
// A nil clock is treated as the real clock.
func NewPollerWithClock(interval time.Duration, cmd func() tea.Cmd, clock Clock) *Poller {
	if clock == nil {
		clock = realClock{}
	}
	return &Poller{
		interval: interval,
		cmd:      cmd,
		clock:    clock,
	}
}

// Check should be called from your component's Update when receiving a TickMsg.
// Returns a tea.Cmd if it's time to poll, nil otherwise.
// ForceRefresh takes priority and works even when paused.
func (p *Poller) Check(msg TickMsg) tea.Cmd {
	now := p.clock.Now()
	if p.needsRefresh {
		p.needsRefresh = false
		p.lastPoll = now
		return p.cmd()
	}

	if p.paused {
		return nil
	}

	if now.Sub(p.lastPoll) >= p.interval {
		p.lastPoll = now
		return p.cmd()
	}
	return nil
}

// SetInterval changes the polling interval.
func (p *Poller) SetInterval(d time.Duration) {
	p.interval = d
}

// Pause stops periodic polling. ForceRefresh still works when paused.
func (p *Poller) Pause() { p.paused = true }

// Resume resumes periodic polling.
func (p *Poller) Resume() { p.paused = false }

// TogglePause toggles between paused and active.
func (p *Poller) TogglePause() { p.paused = !p.paused }

// ForceRefresh triggers a poll on the next Check call, even if paused.
func (p *Poller) ForceRefresh() { p.needsRefresh = true }

// IsPaused returns whether polling is paused.
func (p *Poller) IsPaused() bool { return p.paused }

// LastPoll returns the time of the last successful poll.
func (p *Poller) LastPoll() time.Time { return p.lastPoll }
