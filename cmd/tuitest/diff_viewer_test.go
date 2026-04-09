package main

import (
	"strings"
	"testing"

	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

// failingFixture returns a FailureCapture representing a known screen mismatch.
func failingFixture() *tuitest.FailureCapture {
	return &tuitest.FailureCapture{
		TestName:       "TestDiffViewerFixture",
		Kind:           tuitest.FailureScreenEqual,
		ExpectedScreen: "hello world\nfoo bar baz\nline three",
		ActualScreen:   "hello world\nfoo BAR baz\nline FOUR",
	}
}

// goldenFixture returns a FailureCapture representing a golden file mismatch.
func goldenFixture() *tuitest.FailureCapture {
	return &tuitest.FailureCapture{
		TestName:        "TestDiffViewerGolden",
		Kind:            tuitest.FailureGolden,
		GoldenPath:      "testdata/fixture.golden",
		GoldenExpected:  "line one\nline two",
		GoldenActual:    "line one\nline TWO",
	}
}

func TestDiffViewerSideBySide(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(80, 20)
	dv.SetMode(tuitest.DiffModeSideBySide)

	view := dv.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}

	// Header must mention test name and mode.
	if !strings.Contains(view, "TestDiffViewerFixture") {
		t.Errorf("side-by-side view missing test name; got:\n%s", view)
	}
	if !strings.Contains(view, "side-by-side") {
		t.Errorf("side-by-side view missing mode label; got:\n%s", view)
	}

	// Both pane labels must be present.
	if !strings.Contains(view, "EXPECTED") {
		t.Errorf("side-by-side view missing EXPECTED label; got:\n%s", view)
	}
	if !strings.Contains(view, "ACTUAL") {
		t.Errorf("side-by-side view missing ACTUAL label; got:\n%s", view)
	}

	// Content from both sides must appear.
	if !strings.Contains(view, "hello world") {
		t.Errorf("side-by-side view missing shared content; got:\n%s", view)
	}
}

func TestDiffViewerUnified(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(80, 20)
	dv.SetMode(tuitest.DiffModeUnified)

	view := dv.View()
	if view == "" {
		t.Fatal("View() returned empty string for unified mode")
	}

	// Unified diff must show +/- lines for the differing rows.
	if !strings.Contains(view, "- ") {
		t.Errorf("unified view missing deletion line; got:\n%s", view)
	}
	if !strings.Contains(view, "+ ") {
		t.Errorf("unified view missing addition line; got:\n%s", view)
	}

	// The differing content must appear.
	if !strings.Contains(view, "foo bar baz") {
		t.Errorf("unified view missing expected content 'foo bar baz'; got:\n%s", view)
	}
	if !strings.Contains(view, "foo BAR baz") {
		t.Errorf("unified view missing actual content 'foo BAR baz'; got:\n%s", view)
	}
}

func TestDiffViewerCellsOnly(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(80, 20)
	dv.SetMode(tuitest.DiffModeCellsOnly)

	view := dv.View()
	if view == "" {
		t.Fatal("View() returned empty string for cells-only mode")
	}

	// Should list differing cell coordinates.
	if !strings.Contains(view, "text") {
		t.Errorf("cells-only view missing 'text' kind label; got:\n%s", view)
	}
}

func TestDiffViewerGoldenFixture(t *testing.T) {
	dv := tuitest.NewDiffViewer(goldenFixture())
	dv.SetSize(80, 20)
	dv.SetMode(tuitest.DiffModeSideBySide)

	view := dv.View()
	if !strings.Contains(view, "line one") {
		t.Errorf("golden fixture view missing shared content; got:\n%s", view)
	}
	if !strings.Contains(view, "TestDiffViewerGolden") {
		t.Errorf("golden fixture view missing test name; got:\n%s", view)
	}
}

func TestDiffViewerKeyToggle(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(80, 20)

	// Default is side-by-side.
	if dv.Mode() != tuitest.DiffModeSideBySide {
		t.Errorf("default mode = %v, want DiffModeSideBySide", dv.Mode())
	}

	// Use the screen harness to send keys through the component's Update.
	sendKey := func(key string) {
		msg := tuitest.KeyMsgForTesting(key)
		dv2, _ := dv.Update(msg, nil)
		*dv = *dv2
	}

	sendKey("u")
	if dv.Mode() != tuitest.DiffModeUnified {
		t.Errorf("after 'u' mode = %v, want DiffModeUnified", dv.Mode())
	}

	sendKey("d")
	if dv.Mode() != tuitest.DiffModeCellsOnly {
		t.Errorf("after 'd' mode = %v, want DiffModeCellsOnly", dv.Mode())
	}

	sendKey("s")
	if dv.Mode() != tuitest.DiffModeSideBySide {
		t.Errorf("after 's' mode = %v, want DiffModeSideBySide", dv.Mode())
	}
}

func TestDiffViewerBackKey(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(80, 20)

	msg := tuitest.KeyMsgForTesting("q")
	_, cmd := dv.Update(msg, nil)
	if cmd == nil {
		t.Fatal("expected non-nil cmd after 'q'")
	}
	result := cmd()
	if _, ok := result.(tuitest.DiffViewerBackMsg); !ok {
		t.Errorf("'q' produced %T, want DiffViewerBackMsg", result)
	}
}

func TestDiffViewerEmptySize(t *testing.T) {
	dv := tuitest.NewDiffViewer(failingFixture())
	dv.SetSize(0, 0)
	if got := dv.View(); got != "" {
		t.Errorf("View() with zero size = %q, want empty string", got)
	}
}

func TestDiffViewerNilCapture(t *testing.T) {
	dv := tuitest.NewDiffViewer(nil)
	dv.SetSize(80, 20)
	if got := dv.View(); got != "" {
		t.Errorf("View() with nil capture = %q, want empty string", got)
	}
}
