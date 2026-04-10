package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	historyDir     = ".blit/history"
	defaultKeep    = 50
	sparklineWidth = 20
)

// RunRecord is the JSON shape written for every blit run.
type RunRecord struct {
	RunAt    time.Time `json:"run_at"`
	Duration float64   `json:"duration_s"`
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Total    int       `json:"total"`
	Packages []string  `json:"packages,omitempty"`
}

// writeHistory writes a RunRecord to .blit/history/YYYY-MM-DD/HHMMSS.json
// and then prunes old records, keeping at most keep entries.
func writeHistory(rec RunRecord, keep int) error {
	if keep <= 0 {
		keep = defaultKeep
	}

	day := rec.RunAt.Format("2006-01-02")
	ts := rec.RunAt.Format("150405")
	dir := filepath.Join(historyDir, day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("history mkdir: %w", err)
	}

	path := filepath.Join(dir, ts+".json")
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("history marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("history write: %w", err)
	}

	return pruneHistory(keep)
}

// loadHistory reads all RunRecords from .blit/history, sorted newest-first.
func loadHistory() ([]RunRecord, error) {
	var records []RunRecord

	err := filepath.WalkDir(historyDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var rec RunRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil
		}
		records = append(records, rec)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load history: %w", err)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].RunAt.After(records[j].RunAt)
	})
	return records, nil
}

// pruneHistory deletes the oldest records, keeping at most keep entries.
func pruneHistory(keep int) error {
	type entry struct {
		path string
		t    time.Time
	}
	var entries []entry

	_ = filepath.WalkDir(historyDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		entries = append(entries, entry{path: path, t: info.ModTime()})
		return nil
	})

	if len(entries) <= keep {
		return nil
	}

	// Sort oldest first.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].t.Before(entries[j].t)
	})

	for _, e := range entries[:len(entries)-keep] {
		_ = os.Remove(e.path)
		// Remove parent day-dir if empty.
		_ = os.Remove(filepath.Dir(e.path))
	}
	return nil
}

// sparkline builds a Unicode block-bar sparkline from pass-rate values (0–1).
// It uses the last n values, newest-rightmost.
func sparkline(records []RunRecord, n int) string {
	if n > len(records) {
		n = len(records)
	}
	// records are newest-first; we want oldest-first for left→right display.
	slice := make([]RunRecord, n)
	for i := 0; i < n; i++ {
		slice[n-1-i] = records[i]
	}

	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder
	for _, r := range slice {
		rate := 1.0
		if r.Total > 0 {
			rate = float64(r.Passed) / float64(r.Total)
		}
		idx := int(rate * float64(len(bars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		sb.WriteRune(bars[idx])
	}
	return sb.String()
}

// cmdHistory is the entry point for `blit history`.
func cmdHistory(keep int) int {
	records, err := loadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit history] error: %v\n", err)
		return 1
	}
	if len(records) == 0 {
		fmt.Println("No history found. Run some tests first.")
		return 0
	}

	spark := sparkline(records, sparklineWidth)

	// Summary header.
	fmt.Printf("Pass rate (last %d runs): %s\n\n", min(sparklineWidth, len(records)), spark)
	fmt.Printf("%-20s  %6s  %6s  %6s  %8s\n", "Run At", "Passed", "Failed", "Total", "Duration")
	fmt.Println(strings.Repeat("-", 58))

	limit := keep
	if limit > len(records) {
		limit = len(records)
	}
	for _, r := range records[:limit] {
		fmt.Printf("%-20s  %6d  %6d  %6d  %7.2fs\n",
			r.RunAt.Format("2006-01-02 15:04:05"),
			r.Passed, r.Failed, r.Total,
			r.Duration,
		)
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
