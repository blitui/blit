package blit

import (
	"fmt"
	"testing"
)

func TestStatsCollectorRecordSuccess(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordSuccess("repo-a", 5)
	sc.RecordSuccess("repo-b", 3)

	snap := sc.Snapshot()
	if snap.TotalCalls != 2 {
		t.Errorf("TotalCalls = %d, want 2", snap.TotalCalls)
	}
	if snap.SuccessCalls != 2 {
		t.Errorf("SuccessCalls = %d, want 2", snap.SuccessCalls)
	}
	if snap.TotalItems != 8 {
		t.Errorf("TotalItems = %d, want 8", snap.TotalItems)
	}
	if snap.Sources["repo-a"].LastSuccess != true {
		t.Error("repo-a should be last success")
	}
	if snap.Sources["repo-a"].ItemCount != 5 {
		t.Errorf("repo-a ItemCount = %d, want 5", snap.Sources["repo-a"].ItemCount)
	}
}

func TestStatsCollectorRecordFailure(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordFailure("repo-a", fmt.Errorf("timeout"))
	sc.RecordFailure("repo-a", fmt.Errorf("timeout"))

	snap := sc.Snapshot()
	if snap.FailedCalls != 2 {
		t.Errorf("FailedCalls = %d, want 2", snap.FailedCalls)
	}
	if snap.Sources["repo-a"].FailStreak != 2 {
		t.Errorf("FailStreak = %d, want 2", snap.Sources["repo-a"].FailStreak)
	}
	if snap.Sources["repo-a"].LastSuccess {
		t.Error("repo-a should not be last success after failures")
	}
}

func TestStatsCollectorFailureThenSuccess(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordFailure("repo-a", fmt.Errorf("err"))
	sc.RecordSuccess("repo-a", 10)

	snap := sc.Snapshot()
	if snap.Sources["repo-a"].FailStreak != 0 {
		t.Errorf("FailStreak should reset to 0 after success, got %d", snap.Sources["repo-a"].FailStreak)
	}
	if !snap.Sources["repo-a"].LastSuccess {
		t.Error("should be last success after successful fetch")
	}
}

func TestStatsCollectorRecordCached(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordCached("repo-a", 5)

	snap := sc.Snapshot()
	if !snap.Sources["repo-a"].UsingCache {
		t.Error("should be using cache")
	}
}

func TestStatsCollectorRateLimit(t *testing.T) {
	sc := NewStatsCollector()
	sc.SetRateLimit(4500, 5000)

	snap := sc.Snapshot()
	if snap.RateLimit.Remaining != 4500 {
		t.Errorf("Remaining = %d, want 4500", snap.RateLimit.Remaining)
	}
	if snap.RateLimit.Limit != 5000 {
		t.Errorf("Limit = %d, want 5000", snap.RateLimit.Limit)
	}
}

func TestStatsCollectorSnapshotIsolation(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordSuccess("repo-a", 5)

	snap := sc.Snapshot()
	// Modify snapshot — should not affect collector
	snap.TotalCalls = 999

	snap2 := sc.Snapshot()
	if snap2.TotalCalls == 999 {
		t.Error("modifying snapshot should not affect collector")
	}
}

func TestStatsCollectorViewRenders(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordSuccess("repo-a", 5)
	sc.RecordFailure("repo-b", fmt.Errorf("err"))
	sc.SetRateLimit(100, 5000)

	th := DefaultTheme()
	view := sc.View(80, 40, th)
	if view == "" {
		t.Error("View should not be empty")
	}
	// Should contain stats numbers
	if !containsPlain(view, "2") {
		t.Error("View should contain total calls")
	}
}

func TestStatsCollectorData(t *testing.T) {
	sc := NewStatsCollector()
	sc.RecordSuccess("repo-a", 5)
	sc.SetRateLimit(100, 5000)

	data := sc.Data()
	if data["total_calls"] != 1 {
		t.Errorf("total_calls = %v, want 1", data["total_calls"])
	}
	if data["success_calls"] != 1 {
		t.Errorf("success_calls = %v, want 1", data["success_calls"])
	}
}

// containsPlain checks if s contains substr after stripping ANSI.
func containsPlain(s, substr string) bool {
	return containsStr(stripANSI(s), substr)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && containsAny(s, substr)
}

func containsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
