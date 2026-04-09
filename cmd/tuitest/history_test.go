package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndLoadHistory(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint

	rec := RunRecord{
		RunAt:    time.Now(),
		Duration: 1.23,
		Passed:   5,
		Failed:   1,
		Total:    6,
	}

	if err := writeHistory(rec, 50); err != nil {
		t.Fatalf("writeHistory: %v", err)
	}

	// Verify the file exists under the expected path.
	day := rec.RunAt.Format("2006-01-02")
	ts := rec.RunAt.Format("150405")
	path := filepath.Join(historyDir, day, ts+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}

	var got RunRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Passed != rec.Passed || got.Failed != rec.Failed {
		t.Errorf("got %+v, want %+v", got, rec)
	}
}

func TestLoadHistoryEmpty(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint

	records, err := loadHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty, got %d records", len(records))
	}
}

func TestPruneHistory(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint

	base := time.Now()
	for i := 0; i < 10; i++ {
		rec := RunRecord{
			RunAt:    base.Add(time.Duration(i) * time.Second),
			Duration: float64(i),
			Passed:   i,
			Failed:   0,
			Total:    i,
		}
		if err := writeHistory(rec, 100); err != nil {
			t.Fatalf("writeHistory %d: %v", i, err)
		}
	}

	// Now prune to 5.
	if err := pruneHistory(5); err != nil {
		t.Fatalf("pruneHistory: %v", err)
	}

	records, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(records) > 5 {
		t.Errorf("expected <=5 records after prune, got %d", len(records))
	}
}

func TestSparkline(t *testing.T) {
	records := []RunRecord{
		{Passed: 10, Failed: 0, Total: 10},
		{Passed: 5, Failed: 5, Total: 10},
		{Passed: 0, Failed: 10, Total: 10},
	}
	s := sparkline(records, 3)
	if len([]rune(s)) != 3 {
		t.Errorf("expected 3 chars, got %d: %q", len([]rune(s)), s)
	}
}

func TestSparklineFewer(t *testing.T) {
	records := []RunRecord{
		{Passed: 10, Failed: 0, Total: 10},
	}
	s := sparkline(records, 5)
	// Should only produce as many chars as records available.
	if len([]rune(s)) != 1 {
		t.Errorf("expected 1 char, got %d: %q", len([]rune(s)), s)
	}
}

func TestHistoryNewestFirst(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint

	base := time.Now()
	for i := 0; i < 3; i++ {
		rec := RunRecord{
			RunAt:    base.Add(time.Duration(i) * time.Second),
			Duration: float64(i),
			Passed:   i + 1,
			Total:    i + 1,
		}
		if err := writeHistory(rec, 50); err != nil {
			t.Fatalf("writeHistory %d: %v", i, err)
		}
	}

	records, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3, got %d", len(records))
	}
	if !records[0].RunAt.After(records[1].RunAt) {
		t.Error("expected newest-first ordering")
	}
}

func TestWriteHistoryPrunesAutomatically(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint

	base := time.Now()
	keep := 3
	for i := 0; i < 5; i++ {
		rec := RunRecord{
			RunAt:    base.Add(time.Duration(i) * time.Second),
			Duration: float64(i),
			Passed:   i + 1,
			Total:    i + 1,
		}
		if err := writeHistory(rec, keep); err != nil {
			t.Fatalf("writeHistory %d: %v", i, err)
		}
	}

	records, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(records) > keep {
		t.Errorf("expected <=%d after auto-prune, got %d", keep, len(records))
	}
}
