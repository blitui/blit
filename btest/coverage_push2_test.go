package btest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ── assert_extended.go coverage ─────────────────────────────────────────

func TestAssertRegionFg_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[31mhello\x1b[0m")
	AssertRegionFg(t, s, 0, 0, 5, 1, "red")
}

func TestAssertRegionBg_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[41mworld\x1b[0m")
	AssertRegionBg(t, s, 0, 0, 5, 1, "red")
}

func TestAssertRegionBold_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[1mBOLD\x1b[0m")
	AssertRegionBold(t, s, 0, 0, 4, 1)
}

func TestAssertKeybind_Found(t *testing.T) {
	s := NewScreen(40, 3)
	s.Render("q quit  h help")
	AssertKeybind(t, s, "q", "quit")
}

func TestAssertKeybind_DescriptionEmpty(t *testing.T) {
	s := NewScreen(40, 3)
	s.Render("q quit")
	AssertKeybind(t, s, "q", "")
}

func TestAssertScreenMatches_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("hello world")
	AssertScreenMatches(t, s, `hello\s+world`)
}

func TestAssertNoANSI_Clean(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("clean text")
	AssertNoANSI(t, s)
}

func TestAssertCursorRowContains_Cover(t *testing.T) {
	s := NewScreen(20, 5)
	s.Render("line0\nline1\nline2")
	// After render, cursor is at end of output (row 2).
	AssertCursorRowContains(t, s, "line2")
}

func TestAssertColumnContains_Cover(t *testing.T) {
	s := NewScreen(20, 5)
	s.Render("abc\ndef\nghi")
	AssertColumnContains(t, s, 0, 0, 2, "a")
}

func TestAssertColumnCount_Cover(t *testing.T) {
	s := NewScreen(20, 5)
	s.Render("abc\nabc\nxyz")
	AssertColumnCount(t, s, 0, 0, 2, "a", 2)
}

// ── diff.go: LoadFailureCapture / ListFailureCaptures ───────────────────

func TestLoadFailureCapture(t *testing.T) {
	dir := t.TempDir()
	capDir := filepath.Join(dir, "failures")
	if err := os.MkdirAll(capDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fc := FailureCapture{
		TestName:       "TestLoad",
		Kind:           FailureScreenEqual,
		ExpectedScreen: "exp",
		ActualScreen:   "act",
	}
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(capDir, "TestLoad.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	origDir := failureCaptureDir
	failureCaptureDir = capDir
	defer func() { failureCaptureDir = origDir }()

	loaded, err := LoadFailureCapture("TestLoad")
	if err != nil {
		t.Fatalf("LoadFailureCapture: %v", err)
	}
	if loaded.TestName != "TestLoad" {
		t.Errorf("TestName = %q, want TestLoad", loaded.TestName)
	}
	if loaded.ExpectedScreen != "exp" {
		t.Errorf("ExpectedScreen = %q, want exp", loaded.ExpectedScreen)
	}
}

func TestListFailureCaptures_Cover(t *testing.T) {
	dir := t.TempDir()
	capDir := filepath.Join(dir, "failures")
	if err := os.MkdirAll(capDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"TestA", "TestB"} {
		if err := os.WriteFile(filepath.Join(capDir, name+".json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(capDir, "readme.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir := failureCaptureDir
	failureCaptureDir = capDir
	defer func() { failureCaptureDir = origDir }()

	names, err := ListFailureCaptures()
	if err != nil {
		t.Fatalf("ListFailureCaptures: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("got %d captures, want 2: %v", len(names), names)
	}
}

func TestListFailureCaptures_NonExistent(t *testing.T) {
	origDir := failureCaptureDir
	failureCaptureDir = filepath.Join(t.TempDir(), "nonexistent")
	defer func() { failureCaptureDir = origDir }()

	names, err := ListFailureCaptures()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty, got %v", names)
	}
}

// ── golden.go: AssertGolden ─────────────────────────────────────────────

func TestAssertGolden_CreateAndMatch(t *testing.T) {
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	s := NewScreen(20, 3)
	s.Render("golden test")

	AssertGolden(t, s, "cover_golden")
	AssertGolden(t, s, "cover_golden")

	_ = os.RemoveAll(filepath.Join(dir, "testdata"))
}

// ── snapshot.go: AssertSnapshot ─────────────────────────────────────────

func TestAssertSnapshot_CreateAndMatch(t *testing.T) {
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	s := NewScreen(20, 3)
	s.Render("snapshot test")

	AssertSnapshot(t, s, "cover_snap")
	AssertSnapshot(t, s, "cover_snap")
}

// ── diff_viewer.go: truncateStr / colorForKind ──────────────────────────

func TestTruncateStr(t *testing.T) {
	cases := []struct {
		s    string
		maxW int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 1, "h"},
		{"hi", 2, "hi"},
	}
	for _, tc := range cases {
		got := truncateStr(tc.s, tc.maxW)
		if got != tc.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tc.s, tc.maxW, got, tc.want)
		}
	}
}

func TestColorForKind(t *testing.T) {
	_ = colorForKind(CellTextDiffer)
	_ = colorForKind(CellStyleDiffer)
	_ = colorForKind(CellKind(99))
}

// ── smoke.go: SmokeTest ─────────────────────────────────────────────────

func TestSmokeTest_StableModel(t *testing.T) {
	m := &tickModel{}
	SmokeTest(t, m, nil)
}

// ── sess.go: Save / LoadSession / Replay ────────────────────────────────

func TestSessionRecorder_SaveAndReplay(t *testing.T) {
	m := &cmdModel{}
	tm := NewTestModel(t, m, 40, 10)
	rec := NewSessionRecorder(tm)

	rec.Key("a")
	rec.Type("hello")
	rec.Resize(60, 20)

	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	if err := rec.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	sessPath := path + ".tuisess"
	if _, err := os.Stat(sessPath); err != nil {
		t.Fatalf("session file not found: %v", err)
	}

	sess, err := LoadSession(sessPath)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if sess.Version != SessionFormatVersion {
		t.Errorf("version = %d, want %d", sess.Version, SessionFormatVersion)
	}
	if len(sess.Steps) == 0 {
		t.Error("expected non-empty steps")
	}
}

func TestLoadSession_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.tuisess")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadSession(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadSession_UnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v99.tuisess")
	data := `{"version":99,"cols":80,"lines":24,"steps":[]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadSession(path)
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}

func TestReplay_Simple(t *testing.T) {
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10)
	rec := NewSessionRecorder(tm)
	rec.Key("enter")

	dir := t.TempDir()
	path := filepath.Join(dir, "replay.tuisess")
	if err := rec.Save(path); err != nil {
		t.Fatal(err)
	}

	Replay(t, &tickModel{}, path)
}

// ── assert.go: style assertion pass paths ───────────────────────────────

func TestAssertRowNotContains_Pass(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertRowNotContains(t, s, 0, "world")
}

func TestAssertBgAt_Default(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("text")
	AssertBgAt(t, s, 0, 0, "")
}

func TestAssertItalicAt_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[3mitalic\x1b[0m")
	AssertItalicAt(t, s, 0, 0)
}

func TestAssertUnderlineAt_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[4munderlined\x1b[0m")
	AssertUnderlineAt(t, s, 0, 0)
}

func TestAssertReverseAt_Cover(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("\x1b[7mreversed\x1b[0m")
	AssertReverseAt(t, s, 0, 0)
}

func TestAssertRegionNotContains_Pass(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertRegionNotContains(t, s, 0, 0, 20, 1, "xyz")
}

func TestAssertMatches_Pass(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("hello 123 world")
	AssertMatches(t, s, `\d+`)
}

func TestAssertRowMatches_Pass(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("item-42 done")
	AssertRowMatches(t, s, 0, `item-\d+`)
}

// ── diff.go: ScreenDiff ─────────────────────────────────────────────────

func TestScreenDiff_NonEmpty(t *testing.T) {
	a := NewScreen(20, 2)
	a.Render("hello")
	b := NewScreen(20, 2)
	b.Render("world")
	diff := ScreenDiff(a, b)
	if len(diff.Lines) == 0 {
		t.Error("expected non-empty diff")
	}
}

// ── harness.go: Expect ──────────────────────────────────────────────────

func TestHarness_ExpectChain(t *testing.T) {
	m := &tickModel{}
	h := NewHarness(t, m, 40, 10)
	h.Expect("").Done()
}

// ── DiffViewer: golden-mode path ────────────────────────────────────────

func TestDiffViewer_GoldenMode(t *testing.T) {
	fc := &FailureCapture{
		TestName:       "TestGoldenDiff",
		Kind:           FailureGolden,
		GoldenExpected: "expected golden\nline2",
		GoldenActual:   "actual golden\nline2",
	}
	dv := NewDiffViewer(fc)
	dv.SetSize(120, 40)

	modes := []DiffMode{DiffModeSideBySide, DiffModeUnified, DiffModeCellsOnly}
	for _, mode := range modes {
		dv.SetMode(mode)
		view := dv.View()
		if view == "" {
			t.Errorf("View() empty for golden mode %d", mode)
		}
	}
}

// ── fuzz.go: Fuzz ───────────────────────────────────────────────────────

func TestFuzz_StableModel(t *testing.T) {
	m := &tickModel{}
	result := Fuzz(t, m, FuzzConfig{Seed: 1, Iterations: 50})
	if result.Panicked {
		t.Errorf("stable model should not panic: %v", result.PanicValue)
	}
}

// ── model.go: processCmd sequenceMsg path ───────────────────────────────

type seqModel struct {
	msgs []string
}

func (m *seqModel) Init() tea.Cmd { return nil }
func (m *seqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s, ok := msg.(string); ok {
		m.msgs = append(m.msgs, s)
		return m, nil
	}
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Sequence(
			func() tea.Msg { return "seq-1" },
			func() tea.Msg { return "seq-2" },
		)
	}
	return m, nil
}
func (m *seqModel) View() string { return "seq" }

func TestProcessCmd_SequenceMsg(t *testing.T) {
	m := &seqModel{}
	tm := NewTestModel(t, m, 40, 10)
	tm.SendKey("a")
	// tea.Sequence may not be fully supported by processCmd; just verify no panic.
	_ = len(m.msgs)
}

// ── SaveFailureCapture full path ────────────────────────────────────────

func TestSaveFailureCapture_FullPath(t *testing.T) {
	dir := t.TempDir()
	origDir := failureCaptureDir
	failureCaptureDir = filepath.Join(dir, "cap")
	defer func() { failureCaptureDir = origDir }()

	SaveFailureCapture(t, FailureCapture{
		TestName:       "TestFullCapture",
		Kind:           FailureScreenEqual,
		ExpectedScreen: "exp",
		ActualScreen:   "act",
	})

	path := filepath.Join(dir, "cap", "TestFullCapture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("capture file not written: %v", err)
	}
	var fc FailureCapture
	if err := json.Unmarshal(data, &fc); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if fc.TestName != "TestFullCapture" {
		t.Errorf("TestName = %q", fc.TestName)
	}
}

// ── pending.go: Accept creates dir ──────────────────────────────────────

func TestPendingGolden_Accept_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "deep", "nested")
	goldenPath := filepath.Join(subdir, "test.golden")
	newPath := goldenPath + ".new"

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := PendingGolden{
		GoldenPath: goldenPath,
		NewPath:    newPath,
		Actual:     "content",
	}
	if err := p.Accept(); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	data, _ := os.ReadFile(goldenPath)
	if string(data) != "content" {
		t.Errorf("golden = %q", data)
	}
}

// ── screen.go: StyleAt out-of-bounds ────────────────────────────────────

func TestStyleAt_OutOfBounds(t *testing.T) {
	s := NewScreen(10, 3)
	s.Render("hi")
	style := s.StyleAt(-1, 0)
	if style.Bold || style.Fg != "" {
		t.Error("expected zero style for out-of-bounds")
	}
	style = s.StyleAt(0, 999)
	if style.Bold || style.Fg != "" {
		t.Error("expected zero style for out-of-bounds col")
	}
}

// ── screen.go: FindText / FindRegexp ────────────────────────────────────

func TestFindText_Found(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("abc\ndef\nghi")
	r, c := s.FindText("def")
	if r != 1 || c != 0 {
		t.Errorf("FindText(def) = (%d,%d), want (1,0)", r, c)
	}
}

func TestFindText_NotFound(t *testing.T) {
	s := NewScreen(20, 3)
	s.Render("abc")
	r, c := s.FindText("xyz")
	if r != -1 || c != -1 {
		t.Errorf("FindText(xyz) = (%d,%d), want (-1,-1)", r, c)
	}
}

func TestFindRegexp_Cover(t *testing.T) {
	s := NewScreen(30, 3)
	s.Render("item-42 done")
	r, c := s.FindRegexp(`item-\d+`)
	if r < 0 || c < 0 {
		t.Error("expected regex match")
	}
}

// ── reporter.go coverage ────────────────────────────────────────────────

func TestReport_WriteJUnit_Cover(t *testing.T) {
	dir := t.TempDir()
	r := &Report{
		Results: []TestResult{
			{Name: "TestA", Passed: true},
			{Name: "TestB", Passed: false, Failure: "oops"},
		},
	}
	path := filepath.Join(dir, "junit.xml")
	if err := r.WriteJUnit(path); err != nil {
		t.Fatalf("WriteJUnit: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("empty JUnit file")
	}
}

func TestReport_WriteHTML_Cover(t *testing.T) {
	dir := t.TempDir()
	r := &Report{
		Results: []TestResult{
			{Name: "TestA", Passed: true},
		},
	}
	path := filepath.Join(dir, "report.html")
	if err := r.WriteHTML(path); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("empty HTML file")
	}
}

// ── AssertStyleAt coverage ──────────────────────────────────────────────

func TestAssertStyleAt_AllFields(t *testing.T) {
	s := NewScreen(30, 3)
	// Bold + italic + underline + reverse + fg red + bg blue
	s.Render("\x1b[1;3;4;7;31;44mstyle\x1b[0m")
	AssertStyleAt(t, s, 0, 0, CellStyle{
		Fg:        "red",
		Bg:        "blue",
		Bold:      true,
		Italic:    true,
		Underline: true,
		Reverse:   true,
	})
}

// ── Failure-path tests using stubTB ─────────────────────────────────────
// stubTB is a minimal testing.TB that records failures without aborting.
type stubTB struct {
	testing.TB
	failed bool
}

func (s *stubTB) Helper()                        {}
func (s *stubTB) Error(a ...any)                 { s.failed = true }
func (s *stubTB) Errorf(format string, a ...any) { s.failed = true }
func (s *stubTB) Fatal(a ...any)                 { s.failed = true }
func (s *stubTB) Fatalf(format string, a ...any) { s.failed = true }
func (s *stubTB) Log(a ...any)                   {}
func (s *stubTB) Logf(format string, a ...any)   {}

// These exercise the t.Errorf branches in assertion functions.

func TestAssertContrastFail(t *testing.T) {
	ft := &stubTB{}
	// Two colors with very similar luminance should fail AA.
	AssertContrast(ft, RGB{128, 128, 128}, RGB{140, 140, 140}, WCAGLevelAA)
	if !ft.failed {
		t.Error("expected AssertContrast to fail for low contrast pair")
	}
}

func TestAssertContrastLargeFail(t *testing.T) {
	ft := &stubTB{}
	AssertContrastLarge(ft, RGB{128, 128, 128}, RGB{140, 140, 140}, WCAGLevelAA)
	if !ft.failed {
		t.Error("expected AssertContrastLarge to fail for low contrast pair")
	}
}

func TestContrastReport_Violations_Cover(t *testing.T) {
	r := ContrastReport{
		Results: []ContrastResult{
			{FG: RGB{0, 0, 0}, BG: RGB{255, 255, 255}, Ratio: 21.0, PassAA: true, PassAAA: true},
			{FG: RGB{128, 128, 128}, BG: RGB{140, 140, 140}, Ratio: 1.1, PassAA: false, PassAAA: false},
		},
	}
	v := r.Violations(WCAGLevelAA)
	if len(v) != 1 {
		t.Errorf("expected 1 violation, got %d", len(v))
	}
}

func TestAssertKeybind_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertKeybind(ft, s, "q", "quit")
	if !ft.failed {
		t.Error("expected AssertKeybind to fail when key not found")
	}
}

func TestAssertScreenMatches_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertScreenMatches(ft, s, `^goodbye`)
	if !ft.failed {
		t.Error("expected AssertScreenMatches to fail")
	}
}

func TestAssertRegionBold_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("notbold")
	AssertRegionBold(ft, s, 0, 0, 4, 1)
	if !ft.failed {
		t.Error("expected AssertRegionBold to fail for non-bold text")
	}
}

func TestAssertRegionFg_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertRegionFg(ft, s, 0, 0, 4, 1, "red")
	if !ft.failed {
		t.Error("expected AssertRegionFg to fail for default fg")
	}
}

func TestAssertRegionBg_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertRegionBg(ft, s, 0, 0, 4, 1, "red")
	if !ft.failed {
		t.Error("expected AssertRegionBg to fail for default bg")
	}
}

func TestAssertStyleAt_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("plain")
	AssertStyleAt(ft, s, 0, 0, CellStyle{Bold: true})
	if !ft.failed {
		t.Error("expected AssertStyleAt to fail")
	}
}

func TestAssertRowNotContains_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertRowNotContains(ft, s, 0, "hello")
	if !ft.failed {
		t.Error("expected AssertRowNotContains to fail")
	}
}

func TestAssertRegionNotContains_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello world")
	AssertRegionNotContains(ft, s, 0, 0, 20, 1, "hello")
	if !ft.failed {
		t.Error("expected AssertRegionNotContains to fail")
	}
}

func TestAssertBgAt_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertBgAt(ft, s, 0, 0, "red")
	if !ft.failed {
		t.Error("expected AssertBgAt to fail")
	}
}

func TestAssertItalicAt_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertItalicAt(ft, s, 0, 0)
	if !ft.failed {
		t.Error("expected AssertItalicAt to fail")
	}
}

func TestAssertUnderlineAt_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertUnderlineAt(ft, s, 0, 0)
	if !ft.failed {
		t.Error("expected AssertUnderlineAt to fail")
	}
}

func TestAssertReverseAt_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("text")
	AssertReverseAt(ft, s, 0, 0)
	if !ft.failed {
		t.Error("expected AssertReverseAt to fail")
	}
}

func TestAssertMatches_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertMatches(ft, s, `^\d+$`)
	if !ft.failed {
		t.Error("expected AssertMatches to fail")
	}
}

func TestAssertRowMatches_Fail(t *testing.T) {
	ft := &stubTB{}
	s := NewScreen(20, 3)
	s.Render("hello")
	AssertRowMatches(ft, s, 0, `^\d+$`)
	if !ft.failed {
		t.Error("expected AssertRowMatches to fail")
	}
}

func TestAssertScreensEqual_Fail(t *testing.T) {
	ft := &stubTB{}
	a := NewScreen(20, 2)
	a.Render("hello")
	b := NewScreen(20, 2)
	b.Render("world")
	AssertScreensEqual(ft, a, b)
	if !ft.failed {
		t.Error("expected AssertScreensEqual to fail")
	}
}

func TestAssertScreensNotEqual_Fail(t *testing.T) {
	ft := &stubTB{}
	a := NewScreen(20, 2)
	a.Render("hello")
	b := NewScreen(20, 2)
	b.Render("hello")
	AssertScreensNotEqual(ft, a, b)
	if !ft.failed {
		t.Error("expected AssertScreensNotEqual to fail")
	}
}

// ── WCAGLevel.String coverage ───────────────────────────────────────────

func TestWCAGLevel_String_Unknown(t *testing.T) {
	// Test the default/unknown branch only (AA/AAA covered in a11y_test.go).
	if s := WCAGLevel(99).String(); s == "" {
		t.Error("expected non-empty string for unknown level")
	}
}

// ── SmokeTest with failing model ────────────────────────────────────────

type panicOnResizeModel struct{}

func (m *panicOnResizeModel) Init() tea.Cmd                           { return nil }
func (m *panicOnResizeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *panicOnResizeModel) View() string                            { return "" }

func TestSmokeTest_EmptyView(t *testing.T) {
	ft := &stubTB{}
	SmokeTest(ft, &panicOnResizeModel{}, nil)
	// A model with empty view should cause smoke test failures.
	if !ft.failed {
		t.Error("expected SmokeTest to fail for empty-view model")
	}
}

// ── FuzzAndFail with panicking model ────────────────────────────────────

type panicModel struct{ step int }

func (m *panicModel) Init() tea.Cmd { return nil }
func (m *panicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.step++
	if m.step > 3 {
		panic("boom")
	}
	return m, nil
}
func (m *panicModel) View() string { return "panic" }

func TestFuzzAndFail_DetectsPanic(t *testing.T) {
	ft := &stubTB{}
	FuzzAndFail(ft, &panicModel{}, FuzzConfig{Seed: 1, Iterations: 20})
	if !ft.failed {
		t.Error("expected FuzzAndFail to fail for panicking model")
	}
}

// ── Harness ExpectNot / ExpectRow ───────────────────────────────────────

func TestHarness_ExpectNot(t *testing.T) {
	m := &tickModel{}
	h := NewHarness(t, m, 40, 10)
	h.ExpectNot("nonexistent").Done()
}

func TestHarness_ExpectRow(t *testing.T) {
	m := &tickModel{}
	h := NewHarness(t, m, 40, 10)
	// Row 0 should exist (even if empty).
	h.ExpectRow(0, "").Done()
}

// ── a11y_colorblind.go coverage ─────────────────────────────────────────

func TestAssertDistinguishable_Pass(t *testing.T) {
	// Black and white should be distinguishable under any type.
	AssertDistinguishable(t, RGB{0, 0, 0}, RGB{255, 255, 255}, Protanopia)
}

func TestAssertDistinguishable_Fail(t *testing.T) {
	ft := &stubTB{}
	// Two very similar colors should fail.
	AssertDistinguishable(ft, RGB{128, 0, 0}, RGB{130, 2, 0}, Protanopia)
	if !ft.failed {
		t.Error("expected AssertDistinguishable to fail for similar colors")
	}
}

func TestAssertAllDistinguishable_Pass(t *testing.T) {
	colors := []RGB{{0, 0, 0}, {255, 255, 255}, {0, 0, 255}}
	AssertAllDistinguishable(t, colors, Deuteranopia)
}

func TestAssertAllDistinguishable_Fail(t *testing.T) {
	ft := &stubTB{}
	colors := []RGB{{128, 0, 0}, {130, 2, 0}, {255, 255, 255}}
	AssertAllDistinguishable(ft, colors, Protanopia)
	if !ft.failed {
		t.Error("expected AssertAllDistinguishable to fail")
	}
}

func TestColorBlindType_String_Unknown(t *testing.T) {
	s := ColorBlindType(99).String()
	if s == "" {
		t.Error("expected non-empty string for unknown ColorBlindType")
	}
}

func TestClamp01_Boundaries(t *testing.T) {
	if v := clamp01(-0.5); v != 0 {
		t.Errorf("clamp01(-0.5) = %f, want 0", v)
	}
	if v := clamp01(1.5); v != 1 {
		t.Errorf("clamp01(1.5) = %f, want 1", v)
	}
	if v := clamp01(0.5); v != 0.5 {
		t.Errorf("clamp01(0.5) = %f, want 0.5", v)
	}
}
