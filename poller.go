package blit

import (
	"fmt"
	"strings"
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

// BackoffStrategy defines retry behavior on transient errors.
type BackoffStrategy interface {
	// NextBackoff returns the delay before the next retry attempt.
	// A negative return value means no more retries.
	NextBackoff(attempt int) time.Duration
}

// exponentialBackoff doubles the delay each attempt, capped at max.
type exponentialBackoff struct {
	initial time.Duration
	max     time.Duration
}

// ExponentialBackoff returns a strategy starting at initial, doubling each attempt, capped at max.
func ExponentialBackoff(initial, max time.Duration) BackoffStrategy {
	return &exponentialBackoff{initial: initial, max: max}
}

// NextBackoff returns the delay for the given attempt, doubling from initial up to max.
func (e *exponentialBackoff) NextBackoff(attempt int) time.Duration {
	d := e.initial
	for i := 0; i < attempt; i++ {
		d *= 2
		if d > e.max {
			return e.max
		}
	}
	if d > e.max {
		return e.max
	}
	return d
}

// fixedBackoff uses a constant delay for up to maxAttempts.
type fixedBackoff struct {
	delay       time.Duration
	maxAttempts int
}

// FixedBackoff returns a strategy with constant delay for up to maxAttempts.
func FixedBackoff(delay time.Duration, maxAttempts int) BackoffStrategy {
	return &fixedBackoff{delay: delay, maxAttempts: maxAttempts}
}

// NextBackoff returns the fixed delay if under maxAttempts, or -1 to stop.
func (f *fixedBackoff) NextBackoff(attempt int) time.Duration {
	if attempt >= f.maxAttempts {
		return -1
	}
	return f.delay
}

// noBackoff never retries.
type noBackoff struct{}

// NoBackoff returns a strategy that never retries.
func NoBackoff() BackoffStrategy { return &noBackoff{} }

// NextBackoff always returns -1 (no retry).
func (noBackoff) NextBackoff(int) time.Duration { return -1 }

// PollerOpts configures an enhanced Poller with rate limiting and backoff.
type PollerOpts struct {
	// Name identifies this poller in DevConsole and log messages.
	Name string

	// Interval between poll attempts. Must be >= 100ms. Default: 30s.
	Interval time.Duration

	// Fetch is called on each poll. Returns a tea.Msg on success or error.
	Fetch func() (tea.Msg, error)

	// OnRateLimit is called when rate limiting is detected. Optional.
	OnRateLimit func(resetAt time.Time)

	// OnError is called when a fetch fails after all retries. Optional.
	OnError func(err error)

	// Backoff configures retry behavior. Default: ExponentialBackoff(1s, 5m).
	Backoff BackoffStrategy

	// MaxRetries is the maximum number of retry attempts. Default: 3.
	MaxRetries int

	// Clock for testing. Default: real clock.
	Clock Clock
}

// PollerStats contains the current state of a Poller for DevConsole display.
type PollerStats struct {
	Name              string
	LastPoll          time.Time
	NextPoll          time.Time
	Interval          time.Duration
	IsPaused          bool
	ErrorCount        int
	ConsecutiveErrors int
	LastError         error
	IsRateLimited     bool
	RateLimitReset    time.Time
}

// PollerStartMsg is sent when a fetch begins.
type PollerStartMsg struct{ Name string }

// PollerSuccessMsg wraps the successful fetch result.
type PollerSuccessMsg struct {
	Name string
	Msg  tea.Msg
}

// PollerErrorMsg is sent when all retries are exhausted.
type PollerErrorMsg struct {
	Name     string
	Err      error
	Attempts int
}

// PollerRateLimitedMsg is sent on rate limit detection.
type PollerRateLimitedMsg struct {
	Name    string
	ResetAt time.Time
}

// Poller manages periodic command execution on top of blit's TickMsg.
// Create one with NewPoller, then call Check from your component's Update
// when it receives a TickMsg. The Poller handles interval timing, pause/resume,
// and force-refresh.
//
// For enhanced features (rate limiting, backoff, DevConsole integration), use
// NewPollerWithOpts.
type Poller struct {
	interval     time.Duration
	cmd          func() tea.Cmd
	lastPoll     time.Time
	paused       bool
	needsRefresh bool
	clock        Clock

	// Enhanced fields (only used when created via NewPollerWithOpts).
	name              string
	fetch             func() (tea.Msg, error)
	onRateLimit       func(resetAt time.Time)
	onError           func(err error)
	backoff           BackoffStrategy
	maxRetries        int
	errorCount        int
	consecutiveErrors int
	lastError         error
	isRateLimited     bool
	rateLimitReset    time.Time
	enhanced          bool
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
// For enhanced pollers, rate limiting is respected (except on ForceRefresh).
func (p *Poller) Check(msg TickMsg) tea.Cmd {
	now := p.clock.Now()

	// Auto-clear expired rate limits for enhanced pollers.
	if p.enhanced && p.isRateLimited && !p.rateLimitReset.IsZero() && now.After(p.rateLimitReset) {
		p.ClearRateLimit()
	}

	if p.needsRefresh {
		p.needsRefresh = false
		p.lastPoll = now
		return p.cmd()
	}

	if p.paused {
		return nil
	}

	// Respect rate limiting for enhanced pollers.
	if p.enhanced && p.isRateLimited {
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

// NewPollerWithOpts creates an enhanced Poller configured via PollerOpts.
// The enhanced Poller supports rate limiting, backoff strategies, error tracking,
// and DevConsole integration. The returned Poller is fully compatible with
// the existing Check/Pause/Resume API.
func NewPollerWithOpts(opts PollerOpts) *Poller {
	if opts.Interval < 100*time.Millisecond {
		opts.Interval = 30 * time.Second
	}
	if opts.Backoff == nil {
		opts.Backoff = ExponentialBackoff(time.Second, 5*time.Minute)
	}
	if opts.MaxRetries < 1 {
		opts.MaxRetries = 3
	}
	if opts.Clock == nil {
		opts.Clock = realClock{}
	}

	p := &Poller{
		interval:    opts.Interval,
		clock:       opts.Clock,
		name:        opts.Name,
		fetch:       opts.Fetch,
		onRateLimit: opts.OnRateLimit,
		onError:     opts.OnError,
		backoff:     opts.Backoff,
		maxRetries:  opts.MaxRetries,
		enhanced:    true,
	}

	// Wire up the legacy cmd field to use the enhanced fetch pipeline.
	p.cmd = func() tea.Cmd {
		return p.enhancedFetch()
	}

	return p
}

// enhancedFetch builds a tea.Cmd that runs the Fetch function with retry/backoff
// and emits PollerStartMsg, PollerSuccessMsg, PollerErrorMsg, or PollerRateLimitedMsg.
func (p *Poller) enhancedFetch() tea.Cmd {
	return func() tea.Msg {
		for attempt := 0; attempt <= p.maxRetries; attempt++ {
			msg, err := p.fetch()
			if err == nil {
				p.consecutiveErrors = 0
				p.lastError = nil
				return PollerSuccessMsg{Name: p.name, Msg: msg}
			}

			p.errorCount++
			p.consecutiveErrors++
			p.lastError = err

			// Check if this is the last attempt.
			if attempt >= p.maxRetries {
				break
			}

			// Consult the backoff strategy.
			delay := p.backoff.NextBackoff(attempt)
			if delay < 0 {
				break
			}
			time.Sleep(delay)
		}

		if p.onError != nil {
			p.onError(p.lastError)
		}
		return PollerErrorMsg{Name: p.name, Err: p.lastError, Attempts: p.maxRetries + 1}
	}
}

// Stats returns the current state of this Poller for DevConsole display.
// For pollers created with NewPoller (non-enhanced), Stats returns a
// minimal PollerStats with interval and pause state.
func (p *Poller) Stats() PollerStats {
	now := p.clock.Now()
	var nextPoll time.Time
	if !p.paused && !p.lastPoll.IsZero() {
		nextPoll = p.lastPoll.Add(p.interval)
		if nextPoll.Before(now) {
			nextPoll = now
		}
	}

	return PollerStats{
		Name:              p.name,
		LastPoll:          p.lastPoll,
		NextPoll:          nextPoll,
		Interval:          p.interval,
		IsPaused:          p.paused,
		ErrorCount:        p.errorCount,
		ConsecutiveErrors: p.consecutiveErrors,
		LastError:         p.lastError,
		IsRateLimited:     p.isRateLimited,
		RateLimitReset:    p.rateLimitReset,
	}
}

// NextPoll returns the estimated time of the next poll. Returns the zero
// value if the poller is paused or has never polled.
func (p *Poller) NextPoll() time.Time {
	if p.paused || p.lastPoll.IsZero() {
		return time.Time{}
	}
	next := p.lastPoll.Add(p.interval)
	now := p.clock.Now()
	if next.Before(now) {
		return now
	}
	return next
}

// SetRateLimited marks this poller as rate limited until resetAt. While rate
// limited, Check will not fire. The OnRateLimit callback is invoked if set.
func (p *Poller) SetRateLimited(resetAt time.Time) {
	p.isRateLimited = true
	p.rateLimitReset = resetAt
	if p.onRateLimit != nil {
		p.onRateLimit(resetAt)
	}
}

// ClearRateLimit removes the rate limit flag.
func (p *Poller) ClearRateLimit() {
	p.isRateLimited = false
	p.rateLimitReset = time.Time{}
}

// DebugProvider returns a DebugDataProvider for this Poller suitable for
// registration with the DevConsole. Returns nil for non-enhanced pollers.
func (p *Poller) DebugProvider() DebugDataProvider {
	if !p.enhanced {
		return nil
	}
	return &pollerDebugProvider{poller: p}
}

// pollerDebugProvider implements DebugDataProvider to show poller stats in the DevConsole.
type pollerDebugProvider struct {
	poller *Poller
}

// Name returns the display name for this provider.
func (d *pollerDebugProvider) Name() string {
	if d.poller.name != "" {
		return d.poller.name + " Poller"
	}
	return "Poller"
}

// View renders the poller stats as a compact text block.
func (d *pollerDebugProvider) View(width, height int, theme Theme) string {
	s := d.poller.Stats()
	var b strings.Builder

	fmt.Fprintf(&b, "Poller: %s\n", s.Name)
	fmt.Fprintf(&b, "Interval: %s\n", s.Interval)
	switch {
	case s.IsPaused:
		b.WriteString("Status: PAUSED\n")
	case s.IsRateLimited:
		fmt.Fprintf(&b, "Status: RATE LIMITED (reset %s)\n", s.RateLimitReset.Format(time.RFC3339))
	default:
		b.WriteString("Status: Active\n")
	}
	if !s.LastPoll.IsZero() {
		fmt.Fprintf(&b, "Last Poll: %s\n", s.LastPoll.Format(time.RFC3339))
	}
	if !s.NextPoll.IsZero() {
		fmt.Fprintf(&b, "Next Poll: %s\n", s.NextPoll.Format(time.RFC3339))
	}
	if s.ErrorCount > 0 {
		fmt.Fprintf(&b, "Errors: %d total, %d consecutive\n", s.ErrorCount, s.ConsecutiveErrors)
	}
	if s.LastError != nil {
		fmt.Fprintf(&b, "Last Error: %s\n", s.LastError)
	}
	return b.String()
}

// Data returns structured key-value data for this provider.
func (d *pollerDebugProvider) Data() map[string]any {
	s := d.poller.Stats()
	data := map[string]any{
		"name":               s.Name,
		"interval":           s.Interval.String(),
		"paused":             s.IsPaused,
		"error_count":        s.ErrorCount,
		"consecutive_errors": s.ConsecutiveErrors,
		"rate_limited":       s.IsRateLimited,
	}
	if !s.LastPoll.IsZero() {
		data["last_poll"] = s.LastPoll.Format(time.RFC3339)
	}
	if !s.NextPoll.IsZero() {
		data["next_poll"] = s.NextPoll.Format(time.RFC3339)
	}
	if s.LastError != nil {
		data["last_error"] = s.LastError.Error()
	}
	if s.IsRateLimited {
		data["rate_limit_reset"] = s.RateLimitReset.Format(time.RFC3339)
	}
	return data
}

// RetryOpts configures retry behavior for RetryCmd.
type RetryOpts struct {
	// MaxAttempts is the total number of attempts (including the first).
	// Must be >= 1. Default: 3.
	MaxAttempts int

	// Backoff is the initial delay between retries. Each subsequent retry
	// doubles the delay (exponential backoff). Default: 500ms.
	Backoff time.Duration
}

// RetryErrorMsg is returned when all retry attempts are exhausted.
type RetryErrorMsg struct {
	Err      error
	Attempts int
}

// RetryCmd wraps a fallible function with exponential backoff retry.
// On the first successful call, the resulting tea.Msg is returned.
// If all attempts fail, a RetryErrorMsg is returned.
//
// Usage with Poller:
//
//	poller := NewPoller(30*time.Second, func() tea.Cmd {
//	    return RetryCmd(fetchData, RetryOpts{MaxAttempts: 3, Backoff: 500 * time.Millisecond})
//	})
func RetryCmd(fn func() (tea.Msg, error), opts RetryOpts) tea.Cmd {
	if opts.MaxAttempts < 1 {
		opts.MaxAttempts = 3
	}
	if opts.Backoff <= 0 {
		opts.Backoff = 500 * time.Millisecond
	}
	return func() tea.Msg {
		var lastErr error
		backoff := opts.Backoff
		for attempt := 0; attempt < opts.MaxAttempts; attempt++ {
			msg, err := fn()
			if err == nil {
				return msg
			}
			lastErr = err
			if attempt < opts.MaxAttempts-1 {
				time.Sleep(backoff)
				backoff *= 2
			}
		}
		return RetryErrorMsg{Err: lastErr, Attempts: opts.MaxAttempts}
	}
}
