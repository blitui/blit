package btest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Diff represents a line-by-line comparison between two screen states.
type Diff struct {
	Lines []DiffLine
}

// DiffLine represents a single line difference.
type DiffLine struct {
	Row      int
	Expected string
	Actual   string
	Changed  bool
}

// ScreenDiff compares two screens and returns a Diff showing changed lines.
func ScreenDiff(a, b *Screen) Diff {
	_, aLines := a.Size()
	_, bLines := b.Size()
	maxLines := aLines
	if bLines > maxLines {
		maxLines = bLines
	}

	var lines []DiffLine
	for r := 0; r < maxLines; r++ {
		var aRow, bRow string
		if r < aLines {
			aRow = a.Row(r)
		}
		if r < bLines {
			bRow = b.Row(r)
		}
		lines = append(lines, DiffLine{
			Row:      r,
			Expected: aRow,
			Actual:   bRow,
			Changed:  aRow != bRow,
		})
	}
	return Diff{Lines: lines}
}

// HasChanges reports whether any lines differ.
func (d Diff) HasChanges() bool {
	for _, l := range d.Lines {
		if l.Changed {
			return true
		}
	}
	return false
}

// ChangedLines returns only the lines that differ.
func (d Diff) ChangedLines() []DiffLine {
	var changed []DiffLine
	for _, l := range d.Lines {
		if l.Changed {
			changed = append(changed, l)
		}
	}
	return changed
}

// String returns a human-readable unified-style diff.
func (d Diff) String() string {
	var b strings.Builder
	for _, l := range d.Lines {
		if !l.Changed {
			continue
		}
		fmt.Fprintf(&b, "  row %d:\n", l.Row)
		fmt.Fprintf(&b, "    - %q\n", l.Expected)
		fmt.Fprintf(&b, "    + %q\n", l.Actual)
	}
	return b.String()
}

// diffStrings produces a simple line-by-line diff between two multi-line strings.
// Used internally by assertions for readable error output.
func diffStrings(expected, actual string) string {
	eLines := strings.Split(expected, "\n")
	aLines := strings.Split(actual, "\n")
	maxLen := len(eLines)
	if len(aLines) > maxLen {
		maxLen = len(aLines)
	}

	var b strings.Builder
	for i := 0; i < maxLen; i++ {
		var e, a string
		if i < len(eLines) {
			e = eLines[i]
		}
		if i < len(aLines) {
			a = aLines[i]
		}
		if e != a {
			fmt.Fprintf(&b, "  row %d:\n", i)
			fmt.Fprintf(&b, "    - %q\n", e)
			fmt.Fprintf(&b, "    + %q\n", a)
		}
	}
	return b.String()
}

// AssertScreensEqual fails if two screens don't have identical text content.
// On failure it also persists a FailureCapture so `blit diff` can replay it.
func AssertScreensEqual(t testing.TB, a, b *Screen) {
	t.Helper()
	diff := ScreenDiff(a, b)
	if diff.HasChanges() {
		SaveFailureCapture(t, FailureCapture{
			Kind:           FailureScreenEqual,
			ExpectedScreen: a.String(),
			ActualScreen:   b.String(),
		})
		t.Errorf("screens differ:\n%s", diff.String())
	}
}

// AssertScreensNotEqual fails if two screens have identical text content.
func AssertScreensNotEqual(t testing.TB, a, b *Screen) {
	t.Helper()
	diff := ScreenDiff(a, b)
	if !diff.HasChanges() {
		t.Error("expected screens to differ, but they are identical")
	}
}

// CellKind classifies the type of difference at a single cell position.
type CellKind int

const (
	// CellMatch means expected and actual are identical (text + style).
	CellMatch CellKind = iota
	// CellTextDiffer means the character content differs.
	CellTextDiffer
	// CellStyleDiffer means the character content matches but styles differ.
	CellStyleDiffer
)

// CellDiff holds the comparison result for a single terminal cell.
type CellDiff struct {
	Row           int
	Col           int
	ExpectedText  string
	ActualText    string
	ExpectedStyle CellStyle
	ActualStyle   CellStyle
	Kind          CellKind
}

// ScreenCellDiff performs a per-cell comparison between two screens,
// returning one CellDiff per cell that differs (text or style).
func ScreenCellDiff(expected, actual *Screen) []CellDiff {
	eCols, eLines := expected.Size()
	aCols, aLines := actual.Size()
	cols := eCols
	if aCols > cols {
		cols = aCols
	}
	lines := eLines
	if aLines > lines {
		lines = aLines
	}

	var diffs []CellDiff
	for r := 0; r < lines; r++ {
		for c := 0; c < cols; c++ {
			et := expected.TextAt(r, c, c+1)
			at := actual.TextAt(r, c, c+1)
			es := expected.StyleAt(r, c)
			as := actual.StyleAt(r, c)

			if et == at && es == as {
				continue
			}
			kind := CellTextDiffer
			if et == at {
				kind = CellStyleDiffer
			}
			diffs = append(diffs, CellDiff{
				Row: r, Col: c,
				ExpectedText: et, ActualText: at,
				ExpectedStyle: es, ActualStyle: as,
				Kind: kind,
			})
		}
	}
	return diffs
}

// FailureKind identifies what kind of assertion failed.
type FailureKind string

const (
	// FailureScreenEqual is produced by screen-equality assertions.
	FailureScreenEqual FailureKind = "screen_equal"
	// FailureGolden is produced by AssertGolden when the file differs.
	FailureGolden FailureKind = "golden"
)

// FailureCapture holds the screens (or golden bytes) captured when an
// assertion fails. It is persisted to .blit/failures/<testname>.json
// so that `blit diff <testname>` can replay the diff offline.
type FailureCapture struct {
	TestName string      `json:"test_name"`
	Kind     FailureKind `json:"kind"`
	// For FailureScreenEqual: both fields populated.
	ExpectedScreen string `json:"expected_screen,omitempty"`
	ActualScreen   string `json:"actual_screen,omitempty"`
	// For FailureGolden: golden file path + expected/actual bytes.
	GoldenPath     string `json:"golden_path,omitempty"`
	GoldenExpected string `json:"golden_expected,omitempty"`
	GoldenActual   string `json:"golden_actual,omitempty"`
}

// failureCaptureDir is the directory where FailureCapture JSON files land.
// Exported as a var so tests can redirect it to a temp directory.
var failureCaptureDir = ".blit/failures"

// namer is the subset of testing.TB that exposes Name(). The standard
// *testing.T and *testing.B implement it; minimal fake TBs used in meta-tests
// may not, so we guard with an interface check.
type namer interface {
	Name() string
}

// safeTestName calls t.Name() via the namer interface, recovering from any
// panic that occurs when t is a minimal fake whose embedded testing.TB is nil.
func safeTestName(t testing.TB) (name string) {
	defer func() { recover() }() //nolint:errcheck
	if n, ok := t.(namer); ok {
		name = n.Name()
	}
	return
}

// SaveFailureCapture persists fc to .blit/failures/<testname>.json.
// If fc.TestName is empty and t implements Name(), the test name is filled in.
// Errors are logged via t but never fatal — failure capture is best-effort.
func SaveFailureCapture(t testing.TB, fc FailureCapture) {
	t.Helper()
	if fc.TestName == "" {
		fc.TestName = safeTestName(t)
	}
	if fc.TestName == "" {
		return // nothing useful to save
	}
	if err := os.MkdirAll(failureCaptureDir, 0o755); err != nil {
		t.Logf("btest: could not create failure dir: %v", err)
		return
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", " ", "_", ":", "_").Replace(fc.TestName)
	path := filepath.Join(failureCaptureDir, safe+".json")
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		t.Logf("btest: could not marshal failure capture: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Logf("btest: could not write failure capture: %v", err)
	}
}

// LoadFailureCapture reads the persisted capture for a test name.
func LoadFailureCapture(testName string) (*FailureCapture, error) {
	safe := strings.NewReplacer("/", "_", "\\", "_", " ", "_", ":", "_").Replace(testName)
	path := filepath.Join(failureCaptureDir, safe+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load failure capture: %w", err)
	}
	var fc FailureCapture
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse failure capture: %w", err)
	}
	return &fc, nil
}

// ListFailureCaptures returns test names for all persisted failure captures.
func ListFailureCaptures() ([]string, error) {
	entries, err := os.ReadDir(failureCaptureDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list failure captures: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	return names, nil
}
