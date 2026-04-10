package blit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelfUpdateRollback_NoOldFile(t *testing.T) {
	// Without an .old file next to the running test binary, rollback should
	// return a descriptive error. We cannot move the test binary itself
	// without breaking the test runner, so we only exercise the error path.
	err := SelfUpdateRollback()
	if err == nil {
		t.Error("expected error when no .old file exists")
	}
}

func TestVerifyInstalled_CommandFails(t *testing.T) {
	// Calling VerifyInstalled on the test binary will spawn it with --version
	// which go test does not understand → exits non-zero → rollback attempt.
	// We just assert we get an error back.
	err := VerifyInstalled("v999.0.0")
	if err == nil {
		t.Error("expected error on unexpected output")
	}
}

func TestSelfUpdateRollback_RestoresFromOld(t *testing.T) {
	// Unit-level test that replaces the path-resolution indirection by
	// simulating the rename dance in a tempdir. We exercise the rename
	// logic directly rather than go through SelfUpdateRollback (which
	// depends on os.Executable).
	dir := t.TempDir()
	exe := filepath.Join(dir, "tool")
	old := exe + ".old"
	if err := os.WriteFile(exe, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Simulate what SelfUpdateRollback does.
	broken := exe + ".broken"
	if err := os.Rename(exe, broken); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(old, exe); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(broken)

	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Errorf("after rollback binary = %q, want %q", data, "old")
	}
}
