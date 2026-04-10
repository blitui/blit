package btest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// AssertGolden compares the screen content against a golden file.
// If the file doesn't exist or -update flag is set, it creates/updates the golden file.
// Golden files are stored in testdata/ relative to the test file.
func AssertGolden(t testing.TB, s *Screen, name string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")
	actual := s.String()

	if *update {
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create golden file directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(goldenPath)
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				t.Fatalf("failed to create golden file directory: %v", mkErr)
			}
			if wErr := os.WriteFile(goldenPath, []byte(actual), 0o644); wErr != nil {
				t.Fatalf("failed to write golden file: %v", wErr)
			}
			t.Logf("golden file %s created; re-run test to verify", goldenPath)
			return
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	if string(expected) != actual {
		SaveFailureCapture(t, FailureCapture{
			Kind:           FailureGolden,
			GoldenPath:     goldenPath,
			GoldenExpected: string(expected),
			GoldenActual:   actual,
		})
		// Write the new content to <original>.golden.new so `blit review`
		// can present a review queue without clobbering the accepted golden.
		newPath := goldenPath + ".new"
		if err := os.WriteFile(newPath, []byte(actual), 0o644); err != nil {
			t.Logf("btest: could not write pending golden %s: %v", newPath, err)
		}
		t.Errorf("screen content does not match golden file %s\nwant:\n%s\ngot:\n%s",
			goldenPath, string(expected), actual)
	}
}

// PendingGoldenPath returns the path of the pending-review file for a golden.
// The pending file is <goldenPath>.new and is written by AssertGolden on
// mismatch so that `blit review` can enumerate it.
func PendingGoldenPath(goldenPath string) string {
	return goldenPath + ".new"
}
