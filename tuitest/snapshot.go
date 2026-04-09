package tuitest

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateSnapshots is set by the -tuitest.update flag. When true,
// AssertSnapshot writes the current screen content to disk instead of
// comparing. Regenerate all snapshots with:
//
//	go test ./... -args -tuitest.update
var updateSnapshots = flag.Bool("tuitest.update", false,
	"regenerate tuitest screen snapshots instead of comparing")

// SnapshotDir is the subdirectory (relative to the test file's package)
// where snapshot files are written. Kept exported so users can override
// it for a specific package if they dislike the default.
var SnapshotDir = filepath.Join("testdata", "__snapshots__")

// AssertSnapshot compares scr's current screen contents against a
// previously stored snapshot named <name>.snap under SnapshotDir. On first
// run (or when -tuitest.update is passed), the snapshot is (re)generated.
// Line endings are normalized to \n so snapshots round-trip across OSes.
//
// Example:
//
//	scr := tm.Screen()
//	tuitest.AssertSnapshot(t, scr, "login-form")
func AssertSnapshot(t testing.TB, scr *Screen, name string) {
	t.Helper()
	if scr == nil {
		t.Fatal("AssertSnapshot: nil screen")
	}
	if name == "" {
		t.Fatal("AssertSnapshot: empty snapshot name")
	}
	got := normalizeSnapshot(scr.String())
	path := SnapshotPath(name)

	if *updateSnapshots {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("AssertSnapshot: mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("AssertSnapshot: write: %v", err)
		}
		t.Logf("tuitest: wrote snapshot %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First run with no stored snapshot: create it and pass.
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("AssertSnapshot: mkdir: %v", err)
			}
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatalf("AssertSnapshot: initial write: %v", err)
			}
			t.Logf("tuitest: created snapshot %s (first run)", path)
			return
		}
		t.Fatalf("AssertSnapshot: read: %v", err)
	}

	wantStr := normalizeSnapshot(string(want))
	if got != wantStr {
		t.Errorf("snapshot %q mismatch\n--- want ---\n%s\n--- got ---\n%s\n"+
			"(rerun with -args -tuitest.update to regenerate)",
			name, wantStr, got)
	}
}

// SnapshotPath returns the full on-disk path for the named snapshot.
// Exposed so tests can assert on where a snapshot landed.
func SnapshotPath(name string) string {
	if !strings.HasSuffix(name, ".snap") {
		name += ".snap"
	}
	return filepath.Join(SnapshotDir, name)
}

// normalizeSnapshot collapses \r\n → \n and trims trailing whitespace on
// each line to keep snapshots stable across OSes and terminal widths.
func normalizeSnapshot(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t")
	}
	// Drop trailing blank lines so incidental padding doesn't churn diffs.
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
