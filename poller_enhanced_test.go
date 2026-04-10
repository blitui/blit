package blit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- BackoffStrategy tests ---

func TestExponentialBackoff(t *testing.T) {
	b := ExponentialBackoff(100*time.Millisecond, 2*time.Second)

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1600 * time.Millisecond},
		{5, 2 * time.Second}, // capped
		{6, 2 * time.Second}, // still capped
	}

	for _, tt := range tests {
		got := b.NextBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("attempt %d: got %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestFixedBackoff(t *testing.T) {
	b := FixedBackoff(500*time.Millisecond, 3)

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 500 * time.Millisecond},
		{1, 500 * time.Millisecond},
		{2, 500 * time.Millisecond},
		{3, -1}, // exceeded maxAttempts
		{4, -1},
	}

	for _, tt := range tests {
		got := b.NextBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("attempt %d: got %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestNoBackoff(t *testing.T) {
	b := NoBackoff()
	for i := 0; i < 5; i++ {
		got := b.NextBackoff(i)
		if got >= 0 {
			t.Errorf("attempt %d: got %v, want negative", i, got)
		}
	}
}

// --- Enhanced Poller tests ---

func TestPollerWithOpts_BasicPoll(t *testing.T) {
	c := newFakeClock()
	fetchCalls := 0

	p := NewPollerWithOpts(PollerOpts{
		Name:     "test",
		Interval: 200 * time.Millisecond,
		Fetch: func() (tea.Msg, error) {
			fetchCalls++
			return testMsg{"fetched"}, nil
		},
		Backoff:    NoBackoff(),
		MaxRetries: 1,
		Clock:      c,
	})

	// First check should fire (lastPoll is zero).
	cmd := p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("first Check should return a cmd")
	}
	msg := cmd()
	sm, ok := msg.(PollerSuccessMsg)
	if !ok {
		t.Fatalf("got %T, want PollerSuccessMsg", msg)
	}
	if sm.Name != "test" {
		t.Errorf("name = %q, want %q", sm.Name, "test")
	}
	inner, ok := sm.Msg.(testMsg)
	if !ok || inner.val != "fetched" {
		t.Errorf("inner msg = %v, want testMsg{fetched}", sm.Msg)
	}
	if fetchCalls != 1 {
		t.Errorf("fetchCalls = %d, want 1", fetchCalls)
	}

	// Immediate second check: should not fire.
	cmd = p.Check(TickMsg{})
	if cmd != nil {
		t.Fatal("second Check should not fire before interval")
	}

	// Advance past interval.
	c.Advance(250 * time.Millisecond)
	cmd = p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("Check should fire after interval elapsed")
	}
}

func TestPollerWithOpts_RateLimit(t *testing.T) {
	c := newFakeClock()
	var rateLimitTime time.Time

	p := NewPollerWithOpts(PollerOpts{
		Name:     "rl-test",
		Interval: 100 * time.Millisecond,
		Fetch: func() (tea.Msg, error) {
			return testMsg{"ok"}, nil
		},
		OnRateLimit: func(resetAt time.Time) {
			rateLimitTime = resetAt
		},
		Clock: c,
	})

	// Fire once to establish lastPoll.
	p.Check(TickMsg{})

	// Set rate limited.
	resetAt := c.Now().Add(5 * time.Second)
	p.SetRateLimited(resetAt)

	if !rateLimitTime.Equal(resetAt) {
		t.Errorf("OnRateLimit called with %v, want %v", rateLimitTime, resetAt)
	}

	stats := p.Stats()
	if !stats.IsRateLimited {
		t.Fatal("Stats should show rate limited")
	}

	// Advance past interval but still rate limited.
	c.Advance(1 * time.Second)
	cmd := p.Check(TickMsg{})
	if cmd != nil {
		t.Fatal("rate-limited poller should not fire")
	}

	// Advance past rate limit reset.
	c.Advance(5 * time.Second)
	cmd = p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("Check should fire after rate limit expires")
	}

	stats = p.Stats()
	if stats.IsRateLimited {
		t.Fatal("rate limit should be cleared after expiry")
	}
}

func TestPollerWithOpts_Backoff(t *testing.T) {
	c := newFakeClock()
	fetchCalls := 0
	var lastOnError error

	p := NewPollerWithOpts(PollerOpts{
		Name:     "backoff-test",
		Interval: 100 * time.Millisecond,
		Fetch: func() (tea.Msg, error) {
			fetchCalls++
			return nil, fmt.Errorf("fail-%d", fetchCalls)
		},
		OnError: func(err error) {
			lastOnError = err
		},
		// Use NoBackoff so the test doesn't sleep.
		Backoff:    NoBackoff(),
		MaxRetries: 3,
		Clock:      c,
	})

	// First check fires.
	cmd := p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("first Check should return a cmd")
	}

	msg := cmd()
	errMsg, ok := msg.(PollerErrorMsg)
	if !ok {
		t.Fatalf("got %T, want PollerErrorMsg", msg)
	}
	if errMsg.Name != "backoff-test" {
		t.Errorf("name = %q, want %q", errMsg.Name, "backoff-test")
	}
	// With NoBackoff and MaxRetries=3, first attempt fails, backoff returns -1, so we break.
	// That means 1 attempt + 0 retries = 1 fetch call total (backoff says no retry).
	// Wait — actually the loop goes attempt 0..maxRetries. Let me re-check.
	// attempt=0: fetch fails, attempt < maxRetries (3), backoff.NextBackoff(0) = -1, break.
	// So Attempts = maxRetries+1 = 4? No, it reports the total.
	// Actually the code does: Attempts: p.maxRetries + 1
	// So it always reports maxRetries+1 as the "attempted" count, but NoBackoff breaks early.
	// The fetchCalls tells us the real story.
	if fetchCalls != 1 {
		t.Errorf("fetchCalls = %d, want 1 (NoBackoff stops after first failure)", fetchCalls)
	}
	if lastOnError == nil {
		t.Fatal("OnError should have been called")
	}

	// Verify error tracking.
	stats := p.Stats()
	if stats.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", stats.ErrorCount)
	}
	if stats.ConsecutiveErrors != 1 {
		t.Errorf("ConsecutiveErrors = %d, want 1", stats.ConsecutiveErrors)
	}
	if stats.LastError == nil {
		t.Fatal("LastError should be set")
	}
}

func TestPollerWithOpts_BackoffWithRetries(t *testing.T) {
	c := newFakeClock()
	fetchCalls := 0

	p := NewPollerWithOpts(PollerOpts{
		Name:     "retry-test",
		Interval: 100 * time.Millisecond,
		Fetch: func() (tea.Msg, error) {
			fetchCalls++
			if fetchCalls < 3 {
				return nil, fmt.Errorf("transient")
			}
			return testMsg{"recovered"}, nil
		},
		// FixedBackoff with 0 delay so tests don't sleep.
		Backoff:    FixedBackoff(0, 5),
		MaxRetries: 3,
		Clock:      c,
	})

	cmd := p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("Check should return a cmd")
	}

	msg := cmd()
	sm, ok := msg.(PollerSuccessMsg)
	if !ok {
		t.Fatalf("got %T, want PollerSuccessMsg (retry should succeed)", msg)
	}
	if inner, ok := sm.Msg.(testMsg); !ok || inner.val != "recovered" {
		t.Errorf("inner = %v, want testMsg{recovered}", sm.Msg)
	}
	if fetchCalls != 3 {
		t.Errorf("fetchCalls = %d, want 3", fetchCalls)
	}
}

func TestPollerWithOpts_Stats(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithOpts(PollerOpts{
		Name:     "stats-test",
		Interval: 5 * time.Second,
		Fetch: func() (tea.Msg, error) {
			return testMsg{"ok"}, nil
		},
		Clock: c,
	})

	// Before any poll.
	stats := p.Stats()
	if stats.Name != "stats-test" {
		t.Errorf("Name = %q, want %q", stats.Name, "stats-test")
	}
	if stats.Interval != 5*time.Second {
		t.Errorf("Interval = %v, want 5s", stats.Interval)
	}
	if stats.IsPaused {
		t.Error("should not be paused")
	}

	// Poll once.
	cmd := p.Check(TickMsg{})
	if cmd != nil {
		cmd() // execute the fetch
	}

	stats = p.Stats()
	if stats.LastPoll.IsZero() {
		t.Error("LastPoll should be set after Check")
	}
	if stats.NextPoll.IsZero() {
		t.Error("NextPoll should be set after Check")
	}
	expectedNext := stats.LastPoll.Add(5 * time.Second)
	if !stats.NextPoll.Equal(expectedNext) {
		t.Errorf("NextPoll = %v, want %v", stats.NextPoll, expectedNext)
	}
}

func TestPollerWithOpts_PauseResume(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithOpts(PollerOpts{
		Name:     "pause-test",
		Interval: 100 * time.Millisecond,
		Fetch: func() (tea.Msg, error) {
			return testMsg{"ok"}, nil
		},
		Clock: c,
	})

	// Fire once.
	p.Check(TickMsg{})

	p.Pause()
	stats := p.Stats()
	if !stats.IsPaused {
		t.Fatal("Stats should show paused")
	}

	c.Advance(1 * time.Second)
	cmd := p.Check(TickMsg{})
	if cmd != nil {
		t.Fatal("paused poller should not fire")
	}

	p.Resume()
	stats = p.Stats()
	if stats.IsPaused {
		t.Fatal("Stats should show resumed")
	}

	cmd = p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("resumed poller should fire (interval elapsed)")
	}
}

func TestPollerWithOpts_NextPoll(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithOpts(PollerOpts{
		Name:     "next-test",
		Interval: 10 * time.Second,
		Fetch: func() (tea.Msg, error) {
			return testMsg{"ok"}, nil
		},
		Clock: c,
	})

	// Before any poll, NextPoll is zero.
	if !p.NextPoll().IsZero() {
		t.Error("NextPoll should be zero before first poll")
	}

	// Poll once.
	p.Check(TickMsg{})
	next := p.NextPoll()
	expected := c.Now().Add(10 * time.Second)
	if !next.Equal(expected) {
		t.Errorf("NextPoll = %v, want %v", next, expected)
	}

	// Paused: NextPoll is zero.
	p.Pause()
	if !p.NextPoll().IsZero() {
		t.Error("NextPoll should be zero when paused")
	}
}

func TestPollerDebugProvider(t *testing.T) {
	c := newFakeClock()
	p := NewPollerWithOpts(PollerOpts{
		Name:     "debug-test",
		Interval: 1 * time.Second,
		Fetch: func() (tea.Msg, error) {
			return testMsg{"ok"}, nil
		},
		Clock: c,
	})

	dp := p.DebugProvider()
	if dp == nil {
		t.Fatal("DebugProvider should not be nil for enhanced poller")
	}

	if dp.Name() != "debug-test Poller" {
		t.Errorf("Name = %q, want %q", dp.Name(), "debug-test Poller")
	}

	// Fire a poll to populate stats.
	cmd := p.Check(TickMsg{})
	if cmd != nil {
		cmd()
	}

	view := dp.View(80, 24, Theme{})
	if !strings.Contains(view, "debug-test") {
		t.Errorf("View should contain poller name, got:\n%s", view)
	}
	if !strings.Contains(view, "Active") {
		t.Errorf("View should show Active status, got:\n%s", view)
	}

	data := dp.Data()
	if data["name"] != "debug-test" {
		t.Errorf("Data name = %v, want debug-test", data["name"])
	}
	if data["paused"] != false {
		t.Errorf("Data paused = %v, want false", data["paused"])
	}
}

func TestPollerDebugProvider_NilForLegacy(t *testing.T) {
	p := NewPoller(time.Second, dummyCmd)
	if p.DebugProvider() != nil {
		t.Error("DebugProvider should be nil for legacy poller")
	}
}

func TestBackwardCompat(t *testing.T) {
	// Verify the old API still works exactly as before.
	c := newFakeClock()
	calls := 0
	p := NewPollerWithClock(100*time.Millisecond, func() tea.Cmd {
		return func() tea.Msg {
			calls++
			return "legacy"
		}
	}, c)

	// First check fires.
	cmd := p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("legacy first Check should fire")
	}
	msg := cmd()
	if msg != "legacy" {
		t.Errorf("got %v, want legacy", msg)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}

	// Pause, resume, toggle all work.
	p.Pause()
	if !p.IsPaused() {
		t.Fatal("Pause should work")
	}
	p.Resume()
	if p.IsPaused() {
		t.Fatal("Resume should work")
	}
	p.TogglePause()
	if !p.IsPaused() {
		t.Fatal("TogglePause should work")
	}
	p.TogglePause()

	// ForceRefresh works.
	p.ForceRefresh()
	cmd = p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("ForceRefresh should fire")
	}

	// SetInterval works.
	p.SetInterval(50 * time.Millisecond)
	c.Advance(60 * time.Millisecond)
	cmd = p.Check(TickMsg{})
	if cmd == nil {
		t.Fatal("Check after SetInterval should fire")
	}

	// LastPoll returns a value.
	if p.LastPoll().IsZero() {
		t.Error("LastPoll should not be zero after polling")
	}

	// Stats returns minimal info.
	stats := p.Stats()
	if stats.Interval != 50*time.Millisecond {
		t.Errorf("Stats Interval = %v, want 50ms", stats.Interval)
	}
}

func TestPollerWithOpts_DefaultInterval(t *testing.T) {
	p := NewPollerWithOpts(PollerOpts{
		Name: "defaults",
		Fetch: func() (tea.Msg, error) {
			return nil, nil
		},
		// Interval is 0, should default to 30s.
	})
	if p.interval != 30*time.Second {
		t.Errorf("default interval = %v, want 30s", p.interval)
	}
}

func TestPollerWithOpts_MinInterval(t *testing.T) {
	p := NewPollerWithOpts(PollerOpts{
		Name:     "min",
		Interval: 50 * time.Millisecond, // below 100ms minimum
		Fetch: func() (tea.Msg, error) {
			return nil, nil
		},
	})
	if p.interval != 30*time.Second {
		t.Errorf("interval below 100ms should default to 30s, got %v", p.interval)
	}
}
