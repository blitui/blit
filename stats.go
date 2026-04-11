package blit

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StatsCollector is a thread-safe structured stats accumulator for API-backed
// TUIs. It tracks per-source success/failure counts, rate limits, and
// cache status, and can auto-wire into the DevConsole as a DebugDataProvider.
type StatsCollector struct {
	mu    sync.RWMutex
	stats StatsSnapshot
}

// StatsSnapshot holds a point-in-time view of all collected stats.
type StatsSnapshot struct {
	TotalCalls   int
	SuccessCalls int
	FailedCalls  int
	TotalItems   int
	LastActivity time.Time
	Sources      map[string]*SourceHealth
	RateLimit    RateLimitInfo
}

// SourceHealth tracks the health of a single data source.
type SourceHealth struct {
	LastSuccess bool
	FailStreak  int
	UsingCache  bool
	LastFetchAt time.Time
	ItemCount   int
}

// RateLimitInfo holds rate limit status for an API.
type RateLimitInfo struct {
	Remaining int
	Limit     int
}

// NewStatsCollector creates a new StatsCollector ready for use.
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: StatsSnapshot{
			Sources: make(map[string]*SourceHealth),
		},
	}
}

// RecordSuccess records a successful fetch from source with the given item count.
func (s *StatsCollector) RecordSuccess(source string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.TotalCalls++
	s.stats.SuccessCalls++
	s.stats.TotalItems += count
	s.stats.LastActivity = time.Now()
	h := s.getOrCreateSource(source)
	h.LastSuccess = true
	h.FailStreak = 0
	h.UsingCache = false
	h.LastFetchAt = time.Now()
	h.ItemCount = count
}

// RecordFailure records a failed fetch from source.
func (s *StatsCollector) RecordFailure(source string, _ error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.TotalCalls++
	s.stats.FailedCalls++
	s.stats.LastActivity = time.Now()
	h := s.getOrCreateSource(source)
	h.LastSuccess = false
	h.FailStreak++
	h.LastFetchAt = time.Now()
}

// RecordCached records that cached data was returned for source.
func (s *StatsCollector) RecordCached(source string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := s.getOrCreateSource(source)
	h.UsingCache = true
	h.ItemCount = count
}

// SetRateLimit updates the current rate limit info.
func (s *StatsCollector) SetRateLimit(remaining, limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.RateLimit = RateLimitInfo{Remaining: remaining, Limit: limit}
}

// Snapshot returns a copy of the current stats.
func (s *StatsCollector) Snapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := s.stats
	cp.Sources = make(map[string]*SourceHealth, len(s.stats.Sources))
	for k, v := range s.stats.Sources {
		vc := *v
		cp.Sources[k] = &vc
	}
	return cp
}

func (s *StatsCollector) getOrCreateSource(name string) *SourceHealth {
	h, ok := s.stats.Sources[name]
	if !ok {
		h = &SourceHealth{}
		s.stats.Sources[name] = h
	}
	return h
}

// --- DebugDataProvider implementation ------------------------------------------

// Name implements DebugProvider.
func (s *StatsCollector) Name() string { return "Stats" }

// View implements DebugProvider.
func (s *StatsCollector) View(width, height int, theme Theme) string {
	snap := s.Snapshot()
	var b strings.Builder
	dim := NewStyle().Foreground(theme.Muted)
	hdr := NewStyle().Foreground(theme.Accent).Bold(true)
	errStyle := NewStyle().Foreground(theme.Negative)

	b.WriteString(hdr.Render("API Stats") + "\n")
	b.WriteString(dim.Render(fmt.Sprintf("  Total calls:  %d", snap.TotalCalls)) + "\n")
	b.WriteString(dim.Render(fmt.Sprintf("  Successful:   %d", snap.SuccessCalls)) + "\n")
	if snap.FailedCalls > 0 {
		b.WriteString(errStyle.Render(fmt.Sprintf("  Failed:       %d", snap.FailedCalls)) + "\n")
	} else {
		b.WriteString(dim.Render(fmt.Sprintf("  Failed:       %d", snap.FailedCalls)) + "\n")
	}
	b.WriteString(dim.Render(fmt.Sprintf("  Total items:  %d", snap.TotalItems)) + "\n")
	if !snap.LastActivity.IsZero() {
		ago := time.Since(snap.LastActivity).Truncate(time.Second)
		b.WriteString(dim.Render(fmt.Sprintf("  Last activity: %s ago", ago)) + "\n")
	}

	// Per-source health
	keys := make([]string, 0, len(snap.Sources))
	for k := range snap.Sources {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(keys) > 0 {
		b.WriteString("\n" + hdr.Render("Source Health") + "\n")
		for _, k := range keys {
			h := snap.Sources[k]
			var badge string
			if h.LastSuccess {
				badge = Badge("OK", theme.Positive, true)
			} else {
				badge = Badge("FAIL", theme.Negative, true)
			}
			b.WriteString(fmt.Sprintf("  %s %s", badge, dim.Render(k)) + "\n")
		}
	}

	// Rate limit
	if snap.RateLimit.Limit > 0 {
		b.WriteString("\n" + hdr.Render("Rate Limit") + "\n")
		pct := float64(snap.RateLimit.Remaining) / float64(snap.RateLimit.Limit) * 100
		color := theme.Positive
		if pct < 20 {
			color = theme.Negative
		} else if pct < 50 {
			color = theme.Warn
		}
		b.WriteString(NewStyle().Foreground(color).Render(
			fmt.Sprintf("  %d/%d (%.0f%%)", snap.RateLimit.Remaining, snap.RateLimit.Limit, pct)) + "\n")
	}

	return b.String()
}

// Data implements DebugDataProvider.
func (s *StatsCollector) Data() map[string]any {
	snap := s.Snapshot()
	data := map[string]any{
		"total_calls":   snap.TotalCalls,
		"success_calls": snap.SuccessCalls,
		"failed_calls":  snap.FailedCalls,
		"total_items":   snap.TotalItems,
		"rate_limit":    snap.RateLimit,
	}
	if !snap.LastActivity.IsZero() {
		data["last_activity"] = snap.LastActivity.Format(time.RFC3339)
	}
	sources := make(map[string]any, len(snap.Sources))
	for k, v := range snap.Sources {
		sources[k] = map[string]any{
			"last_success": v.LastSuccess,
			"fail_streak":  v.FailStreak,
			"using_cache":  v.UsingCache,
			"item_count":   v.ItemCount,
		}
	}
	data["sources"] = sources
	return data
}

// DebugProvider returns this StatsCollector as a DebugDataProvider
// suitable for registration with the DevConsole.
func (s *StatsCollector) DebugProvider() DebugDataProvider {
	return s
}
