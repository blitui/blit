package main

import (
	"os"
	"path/filepath"
	"testing"
)

// sampleProfile is a minimal go cover profile for testing.
const sampleProfile = `mode: set
github.com/example/pkg/foo.go:10.20,15.5 3 1
github.com/example/pkg/foo.go:20.10,25.5 2 0
github.com/example/pkg/bar.go:5.10,10.5 5 5
github.com/example/pkg/baz.go:1.10,5.5 1 0
`

func TestParseCoverProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(path, []byte(sampleProfile), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	files, err := parseCoverProfile(path)
	if err != nil {
		t.Fatalf("parseCoverProfile: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("want 3 files, got %d", len(files))
	}

	// foo.go: 3 covered out of 5 total (first block covered, second not)
	foo := files[0]
	if foo.Total != 5 {
		t.Errorf("foo.Total: want 5, got %d", foo.Total)
	}
	if foo.Covered != 3 {
		t.Errorf("foo.Covered: want 3, got %d", foo.Covered)
	}

	// bar.go: 5/5
	bar := files[1]
	if bar.Total != 5 || bar.Covered != 5 {
		t.Errorf("bar: want 5/5, got %d/%d", bar.Covered, bar.Total)
	}

	// baz.go: 0/1
	baz := files[2]
	if baz.Total != 1 || baz.Covered != 0 {
		t.Errorf("baz: want 0/1, got %d/%d", baz.Covered, baz.Total)
	}
}

func TestFileCoveragePct(t *testing.T) {
	cases := []struct {
		covered, total int
		want           float64
	}{
		{0, 0, 100},
		{5, 5, 100},
		{0, 10, 0},
		{3, 5, 60},
	}
	for _, c := range cases {
		fc := fileCoverage{Covered: c.covered, Total: c.total}
		got := fc.Pct()
		if got != c.want {
			t.Errorf("Pct(%d/%d): want %.1f, got %.1f", c.covered, c.total, c.want, got)
		}
	}
}

func TestTotalCoverage(t *testing.T) {
	files := []fileCoverage{
		{Covered: 3, Total: 5},
		{Covered: 5, Total: 5},
		{Covered: 0, Total: 1},
	}
	got := totalCoverage(files)
	// 8 covered / 11 total ≈ 72.73
	want := float64(8) / float64(11) * 100
	if got != want {
		t.Errorf("totalCoverage: want %.4f, got %.4f", want, got)
	}
}

func TestTotalCoverageEmpty(t *testing.T) {
	if got := totalCoverage(nil); got != 100 {
		t.Errorf("totalCoverage(nil): want 100, got %.1f", got)
	}
}

func TestShortName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"github.com/foo/bar/baz.go", "foo/bar/baz.go"},
		{"noslash", "noslash"},
		{"a/b", "b"},
	}
	for _, c := range cases {
		got := shortName(c.in)
		if got != c.want {
			t.Errorf("shortName(%q): want %q, got %q", c.in, c.want, got)
		}
	}
}

// TestCoveragePrintPanel exercises the full render path (G2/G4) without panicking.
func TestCoveragePrintPanel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(path, []byte(sampleProfile), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	// Redirect stdout so output doesn't clutter test output.
	old := os.Stdout
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer func() { os.Stdout = old; null.Close() }()
	os.Stdout = null

	if err := printCoveragePanel(path); err != nil {
		t.Fatalf("printCoveragePanel: %v", err)
	}
}

func TestLatestCoverageFile(t *testing.T) {
	// When neither .tuitest/coverage.out nor any .out file exists, returns "".
	// Run from a temp dir to avoid picking up real coverage files.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if got := latestCoverageFile(); got != "" {
		t.Errorf("latestCoverageFile: want empty, got %q", got)
	}

	// Create a .tuitest/coverage.out and verify it is found.
	if err := os.MkdirAll(".tuitest", 0o755); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(".tuitest", "coverage.out")
	if err := os.WriteFile(outPath, []byte(sampleProfile), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := latestCoverageFile(); got != coverageOut {
		t.Errorf("latestCoverageFile: want %q, got %q", coverageOut, got)
	}
}
