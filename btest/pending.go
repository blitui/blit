package btest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PendingGolden describes a single pending snapshot review item: the
// canonical .golden path plus the candidate .golden.new content.
type PendingGolden struct {
	// GoldenPath is the path to the accepted golden file (may not exist yet
	// if the golden was never written).
	GoldenPath string
	// NewPath is the .golden.new file written by AssertGolden on mismatch.
	NewPath string
	// Expected is the current accepted content (empty if golden doesn't exist).
	Expected string
	// Actual is the candidate content from the .golden.new file.
	Actual string
}

// TestName returns a short human-readable label derived from the golden path.
// It strips the leading testdata/ prefix and the .golden suffix.
func (p PendingGolden) TestName() string {
	name := p.GoldenPath
	name = strings.TrimPrefix(name, "testdata/")
	name = strings.TrimPrefix(name, "testdata\\")
	name = strings.TrimSuffix(name, ".golden")
	return name
}

// Accept writes Actual atomically to GoldenPath and removes NewPath.
// It uses a temp-file + rename so the update is atomic.
func (p PendingGolden) Accept() error {
	dir := filepath.Dir(p.GoldenPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("accept golden: mkdir: %w", err)
	}
	tmp := p.GoldenPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(p.Actual), 0o644); err != nil {
		return fmt.Errorf("accept golden: write tmp: %w", err)
	}
	if err := os.Rename(tmp, p.GoldenPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("accept golden: rename: %w", err)
	}
	_ = os.Remove(p.NewPath)
	return nil
}

// Reject removes the .golden.new file without touching the accepted golden.
func (p PendingGolden) Reject() error {
	if err := os.Remove(p.NewPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reject golden: %w", err)
	}
	return nil
}

// FindPendingGoldens walks root recursively and returns all .golden.new files
// as PendingGolden items. root is typically "." (the package under test).
func FindPendingGoldens(root string) ([]PendingGolden, error) {
	var out []PendingGolden
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".golden.new") {
			return nil
		}
		goldenPath := strings.TrimSuffix(path, ".new")
		newContent, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable
		}
		expectedContent, _ := os.ReadFile(goldenPath) // empty if not found
		out = append(out, PendingGolden{
			GoldenPath: goldenPath,
			NewPath:    path,
			Expected:   string(expectedContent),
			Actual:     string(newContent),
		})
		return nil
	})
	return out, err
}
