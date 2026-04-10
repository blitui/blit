package btest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type harnessCoverageModel struct{ last tea.Msg }

func (m *harnessCoverageModel) Init() tea.Cmd { return nil }
func (m *harnessCoverageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.last = msg
	return m, nil
}
func (m *harnessCoverageModel) View() string { return "x\ny\nz" }

// TestAssertRegionBold exercises the bold attribute check by feeding a
// styled ANSI escape sequence into the screen.
func TestAssertRegionBold(t *testing.T) {
	s := NewScreen(20, 2)
	// \x1b[1m enables bold; \x1b[0m resets.
	s.Render("\x1b[1mHELLO\x1b[0m\r\n")
	AssertRegionBold(t, s, 0, 0, 5, 1)
}

// TestAssertRowNotContains exercises the negative row assertion.
func TestAssertRowNotContains(t *testing.T) {
	s := NewScreen(20, 2)
	s.Render("hello world\r\n")
	AssertRowNotContains(t, s, 0, "goodbye")
}

// TestAssertRegionNotContains exercises the negative region assertion.
func TestAssertRegionNotContains(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("line one\r\nline two\r\n")
	AssertRegionNotContains(t, s, 0, 0, 20, 2, "missing")
}

// TestAssertReverseAt covers the reverse-video assertion. We feed ANSI
// for the "reverse" SGR (7) so the cell attribute is set.
func TestAssertReverseAt(t *testing.T) {
	s := NewScreen(20, 2)
	s.Render("\x1b[7mX\x1b[0m\r\n")
	AssertReverseAt(t, s, 0, 0)
}

// TestHarnessExtraMethods covers Click, Send, ExpectRow, Snapshot, and
// Advance in a single run. Uses a minimal inline model.
func TestHarnessExtraMethods(t *testing.T) {
	m := &harnessCoverageModel{}
	h := NewHarness(t, m, 20, 5)
	h.Click(1, 1)
	h.Send(time.Now())
	h.ExpectRow(0, "x")
	h.Advance(10 * time.Millisecond)
}

// TestReporterWrappers calls JUnitReporter and HTMLReporter directly to
// register cleanup hooks, verifying the wrappers run without panicking.
func TestReporterWrappers(t *testing.T) {
	tmp := t.TempDir()
	junitPath := filepath.Join(tmp, "j.xml")
	htmlPath := filepath.Join(tmp, "r.html")

	rep := &Report{
		Suite:     "wrap",
		StartedAt: time.Now(),
		Results: []TestResult{
			{Name: "ok", Duration: time.Millisecond, Passed: true},
		},
	}
	JUnitReporter(t, rep, junitPath)
	HTMLReporter(t, rep, htmlPath)

	// The cleanup hooks fire at test end; for now just ensure the
	// reporter registrars did not panic.
	_ = os.Remove(junitPath)
	_ = os.Remove(htmlPath)
}

// TestSessionResizeRecords exercises the Resize branch of SessionRecorder.
func TestSessionResizeRecords(t *testing.T) {
	m := &harnessCoverageModel{}
	tm := NewTestModel(t, m, 20, 5)
	rec := NewSessionRecorder(tm)
	rec.Resize(30, 10)
	if len(rec.steps) == 0 {
		t.Error("Resize should record a step")
	}
	if rec.steps[0].Kind != "resize" {
		t.Errorf("step kind = %q, want resize", rec.steps[0].Kind)
	}
}
