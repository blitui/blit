package tuitest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempSnapshotDir temporarily points SnapshotDir at t.TempDir() so the
// test doesn't litter the repo.
func withTempSnapshotDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "__snapshots__")
	orig := SnapshotDir
	SnapshotDir = dir
	t.Cleanup(func() { SnapshotDir = orig })
	return dir
}

func makeScreen(t *testing.T, content string) *Screen {
	t.Helper()
	s := NewScreen(20, 3)
	s.Render(content)
	return s
}

func TestAssertSnapshot_FirstRunCreates(t *testing.T) {
	dir := withTempSnapshotDir(t)
	scr := makeScreen(t, "hello\nworld\n")
	AssertSnapshot(t, scr, "first-run")
	path := filepath.Join(dir, "first-run.snap")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot not created: %v", err)
	}
}

func TestAssertSnapshot_Matches(t *testing.T) {
	_ = withTempSnapshotDir(t)
	scr := makeScreen(t, "same content\n")
	AssertSnapshot(t, scr, "match")
	// Second call should match.
	AssertSnapshot(t, scr, "match")
}

func TestAssertSnapshot_Mismatch(t *testing.T) {
	_ = withTempSnapshotDir(t)
	scr1 := makeScreen(t, "original text\n")
	AssertSnapshot(t, scr1, "mismatch")

	// Use a subtest to detect the failure.
	fakeT := &capturingT{TB: t}
	scr2 := makeScreen(t, "different text\n")
	AssertSnapshot(fakeT, scr2, "mismatch")
	if !fakeT.failed {
		t.Error("mismatched snapshot should have failed the test")
	}
	if !strings.Contains(fakeT.errMsg, "mismatch") {
		t.Errorf("error did not contain snapshot name: %q", fakeT.errMsg)
	}
}

func TestAssertSnapshot_UpdateFlagRegenerates(t *testing.T) {
	dir := withTempSnapshotDir(t)
	scr1 := makeScreen(t, "before\n")
	AssertSnapshot(t, scr1, "regen")

	// Flip the update flag and feed new content.
	orig := *updateSnapshots
	*updateSnapshots = true
	t.Cleanup(func() { *updateSnapshots = orig })

	scr2 := makeScreen(t, "after\n")
	AssertSnapshot(t, scr2, "regen")

	data, err := os.ReadFile(filepath.Join(dir, "regen.snap"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "after") {
		t.Errorf("snapshot not regenerated, content=%q", string(data))
	}
}

func TestNormalizeSnapshot_LineEndings(t *testing.T) {
	in := "a\r\nb\r\nc"
	out := normalizeSnapshot(in)
	if out != "a\nb\nc" {
		t.Errorf("normalizeSnapshot = %q, want a\\nb\\nc", out)
	}
}

func TestNormalizeSnapshot_StripsTrailingWhitespace(t *testing.T) {
	in := "hello   \nworld\t\n"
	out := normalizeSnapshot(in)
	if out != "hello\nworld" {
		t.Errorf("normalizeSnapshot = %q", out)
	}
}

func TestSnapshotPath_AppendsSuffix(t *testing.T) {
	SnapshotDir = "test-snaps"
	defer func() { SnapshotDir = filepath.Join("testdata", "__snapshots__") }()
	if got := SnapshotPath("foo"); got != filepath.Join("test-snaps", "foo.snap") {
		t.Errorf("SnapshotPath = %q", got)
	}
	if got := SnapshotPath("bar.snap"); got != filepath.Join("test-snaps", "bar.snap") {
		t.Errorf("SnapshotPath with suffix = %q", got)
	}
}

// capturingT captures t.Error/Errorf/Fatalf so tests can inspect failure
// messages without actually failing the outer test.
type capturingT struct {
	testing.TB
	failed bool
	errMsg string
}

func (c *capturingT) Errorf(format string, args ...interface{}) {
	c.failed = true
	c.errMsg = sprintf(format, args...)
}

func (c *capturingT) Error(args ...interface{}) {
	c.failed = true
	c.errMsg = sprintln(args...)
}

func (c *capturingT) Fatalf(format string, args ...interface{}) {
	c.failed = true
	c.errMsg = sprintf(format, args...)
}

func (c *capturingT) Fatal(args ...interface{}) {
	c.failed = true
	c.errMsg = sprintln(args...)
}

func (c *capturingT) Helper() {}

// Tiny wrappers to avoid importing fmt in a test helper that might confuse
// the vet printf check on capturingT.
func sprintf(format string, args ...interface{}) string {
	return fmtSprintf(format, args...)
}

func sprintln(args ...interface{}) string {
	return fmtSprintln(args...)
}
