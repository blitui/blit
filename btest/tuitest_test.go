package btest_test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/blitui/blit/btest"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── test reporter ──────────────────────────────────────────────────────────

var (
	passIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("✓")
	failIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")
	dimText  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boldText = lipgloss.NewStyle().Bold(true)
	greenBg  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	redBg    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

type result struct {
	name    string
	group   string
	passed  bool
	elapsed time.Duration
}

var (
	tracker   []result
	trackerMu sync.Mutex
)

func track(t *testing.T) {
	start := time.Now()
	t.Cleanup(func() {
		trackerMu.Lock()
		defer trackerMu.Unlock()
		name := t.Name()
		group, short := splitTestName(name)
		tracker = append(tracker, result{
			name:    short,
			group:   group,
			passed:  !t.Failed(),
			elapsed: time.Since(start),
		})
	})
}

// splitTestName turns "TestScreenPlainText" into ("Screen", "PlainText").
func splitTestName(name string) (string, string) {
	name = strings.TrimPrefix(name, "Test")
	for i := 1; i < len(name); i++ {
		if unicode.IsUpper(rune(name[i])) {
			prefix := name[:i]
			// Known groups — keep multi-word prefixes together.
			switch prefix {
			case "Screen", "Region", "Assert", "Golden", "Test":
				rest := name[i:]
				if prefix == "Test" {
					prefix = "TestModel"
					rest = strings.TrimPrefix(rest, "Model")
				}
				return prefix, rest
			}
		}
	}
	return "", name
}

func TestMain(m *testing.M) {
	flag.Parse()

	// When -v is set, suppress Go's default === RUN / --- PASS output
	// by redirecting stdout to /dev/null during test execution.
	// Our report prints to stderr, bypassing the redirect.
	verbose := testing.Verbose()
	if verbose {
		origStdout := os.Stdout
		devNull, err := os.Open(os.DevNull)
		if err == nil {
			os.Stdout = devNull
			defer func() {
				os.Stdout = origStdout
				devNull.Close()
			}()
		}
	}

	start := time.Now()
	code := m.Run()
	elapsed := time.Since(start)

	w := os.Stderr

	fmt.Fprintln(w)
	fmt.Fprintln(w, "  "+accent.Render("btest")+" "+dimText.Render("· terminal test toolkit"))
	fmt.Fprintln(w)

	trackerMu.Lock()
	results := tracker
	trackerMu.Unlock()

	passed, failed := 0, 0
	currentGroup := ""
	for _, r := range results {
		if r.group != currentGroup {
			currentGroup = r.group
			fmt.Fprintln(w, "  "+boldText.Render(currentGroup))
		}
		icon := passIcon
		if !r.passed {
			icon = failIcon
			failed++
		} else {
			passed++
		}
		ms := fmt.Sprintf("%.3fms", float64(r.elapsed.Microseconds())/1000.0)
		fmt.Fprintf(w, "    %s %s %s\n", icon, r.name, dimText.Render(ms))
	}

	fmt.Fprintln(w)
	total := passed + failed
	if failed == 0 {
		fmt.Fprintf(w, "  %s %s %s\n",
			greenBg.Render("PASS"),
			boldText.Render(fmt.Sprintf("%d tests", total)),
			dimText.Render(fmt.Sprintf("(%dms)", elapsed.Milliseconds())),
		)
	} else {
		fmt.Fprintf(w, "  %s %s, %s %s\n",
			redBg.Render("FAIL"),
			boldText.Render(fmt.Sprintf("%d failed", failed)),
			dimText.Render(fmt.Sprintf("%d passed", passed)),
			dimText.Render(fmt.Sprintf("(%dms)", elapsed.Milliseconds())),
		)
	}
	fmt.Fprintln(w)

	os.Exit(code) //nolint:gocritic // TestMain requires os.Exit after m.Run
}

// ── Screen tests ───────────────────────────────────────────────────────────

func TestScreenPlainText(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	if got := s.Row(0); got != "Hello, World!" {
		t.Errorf("Row(0) = %q, want %q", got, "Hello, World!")
	}
	if got := s.Row(1); got != "" {
		t.Errorf("Row(1) = %q, want empty", got)
	}
}

func TestScreenMultipleLines(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Line 1\nLine 2\nLine 3")

	tests := []struct {
		row  int
		want string
	}{
		{0, "Line 1"},
		{1, "Line 2"},
		{2, "Line 3"},
		{3, ""},
	}
	for _, tt := range tests {
		if got := s.Row(tt.row); got != tt.want {
			t.Errorf("Row(%d) = %q, want %q", tt.row, got, tt.want)
		}
	}
}

func TestScreenTextAt(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	if got := s.TextAt(0, 0, 5); got != "Hello" {
		t.Errorf("TextAt(0, 0, 5) = %q, want %q", got, "Hello")
	}
	if got := s.TextAt(0, 7, 12); got != "World" {
		t.Errorf("TextAt(0, 7, 12) = %q, want %q", got, "World")
	}
}

func TestScreenTextAtBounds(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("test")

	if got := s.TextAt(-1, 0, 5); got != "" {
		t.Errorf("TextAt(-1, 0, 5) = %q, want empty", got)
	}
	if got := s.TextAt(10, 0, 5); got != "" {
		t.Errorf("TextAt(10, 0, 5) = %q, want empty", got)
	}
	if got := s.TextAt(0, 5, 3); got != "" {
		t.Errorf("TextAt(0, 5, 3) = %q, want empty", got)
	}
}

func TestScreenContains(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!\nSecond line here")

	if !s.Contains("Hello") {
		t.Error("Contains(Hello) should be true")
	}
	if !s.Contains("Second") {
		t.Error("Contains(Second) should be true")
	}
	if s.Contains("Missing") {
		t.Error("Contains(Missing) should be false")
	}
}

func TestScreenContainsAt(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	if !s.ContainsAt(0, 0, "Hello") {
		t.Error("ContainsAt(0, 0, Hello) should be true")
	}
	if !s.ContainsAt(0, 7, "World") {
		t.Error("ContainsAt(0, 7, World) should be true")
	}
	if s.ContainsAt(0, 1, "Hello") {
		t.Error("ContainsAt(0, 1, Hello) should be false")
	}
}

func TestScreenCursorPos(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hi")

	row, col := s.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

func TestScreenString(t *testing.T) {
	track(t)
	s := btest.NewScreen(20, 3)
	s.Render("Row 0\nRow 1\nRow 2")

	got := s.String()
	want := "Row 0\nRow 1\nRow 2"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestScreenStyleBold(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m Normal")

	style := s.StyleAt(0, 0)
	if !style.Bold {
		t.Error("expected bold at (0, 0)")
	}
	style = s.StyleAt(0, 5)
	if style.Bold {
		t.Error("expected non-bold at (0, 5)")
	}
}

func TestScreenStyleFgColor(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[31mRed\x1b[0m")

	style := s.StyleAt(0, 0)
	if style.Fg != "red" {
		t.Errorf("Fg = %q, want %q", style.Fg, "red")
	}
	style = s.StyleAt(0, 3)
	if style.Fg != "" {
		t.Errorf("Fg after reset = %q, want empty", style.Fg)
	}
}

func TestScreenStyleBgColor(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[44mBlue\x1b[0m")

	style := s.StyleAt(0, 0)
	if style.Bg != "blue" {
		t.Errorf("Bg = %q, want %q", style.Bg, "blue")
	}
}

func TestScreenStyleItalicUnderline(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[3;4mStyled\x1b[0m")

	style := s.StyleAt(0, 0)
	if !style.Italic {
		t.Error("expected italic at (0, 0)")
	}
	if !style.Underline {
		t.Error("expected underline at (0, 0)")
	}
}

func TestScreenStyleAtOutOfBounds(t *testing.T) {
	track(t)
	s := btest.NewScreen(10, 5)
	s.Render("test")

	style := s.StyleAt(-1, 0)
	if style != (btest.CellStyle{}) {
		t.Error("StyleAt(-1, 0) should return zero CellStyle")
	}
	style = s.StyleAt(0, 100)
	if style != (btest.CellStyle{}) {
		t.Error("StyleAt(0, 100) should return zero CellStyle")
	}
}

func TestScreenRenderOverwrites(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("First render")
	s.Render("Second")

	if s.Contains("First") {
		t.Error("first render content should be cleared")
	}
	if !s.Contains("Second") {
		t.Error("should contain second render content")
	}
}

// ── Region tests ───────────────────────────────────────────────────────────

func TestRegionContains(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA BBBB\nCCCC DDDD\nEEEE FFFF")

	r := s.Region(0, 5, 4, 3)
	if !r.Contains("BBBB") {
		t.Error("region should contain BBBB")
	}
	if !r.Contains("DDDD") {
		t.Error("region should contain DDDD")
	}
	if r.Contains("AAAA") {
		t.Error("region should not contain AAAA")
	}
}

func TestRegionRow(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA BBBB\nCCCC DDDD")

	r := s.Region(0, 5, 4, 2)
	if got := r.Row(0); got != "BBBB" {
		t.Errorf("Region.Row(0) = %q, want %q", got, "BBBB")
	}
	if got := r.Row(1); got != "DDDD" {
		t.Errorf("Region.Row(1) = %q, want %q", got, "DDDD")
	}
}

func TestRegionRowOutOfBounds(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("test")

	r := s.Region(0, 0, 4, 1)
	if got := r.Row(-1); got != "" {
		t.Errorf("Region.Row(-1) = %q, want empty", got)
	}
	if got := r.Row(5); got != "" {
		t.Errorf("Region.Row(5) = %q, want empty", got)
	}
}

func TestRegionString(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 3)
	s.Render("AB CD\nEF GH")

	r := s.Region(0, 3, 2, 2)
	got := r.String()
	want := "CD\nGH"
	if got != want {
		t.Errorf("Region.String() = %q, want %q", got, want)
	}
}

// ── Assert tests ───────────────────────────────────────────────────────────

type fakeTB struct {
	testing.TB
	failed bool
}

func (f *fakeTB) Helper()                        {}
func (f *fakeTB) Error(a ...any)                 { f.failed = true }
func (f *fakeTB) Errorf(format string, a ...any) { f.failed = true }
func (f *fakeTB) Fatal(a ...any)                 { f.failed = true }
func (f *fakeTB) Fatalf(format string, a ...any) { f.failed = true }
func (f *fakeTB) Log(a ...any)                   {}
func (f *fakeTB) Logf(format string, a ...any)   {}

func TestAssertContainsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	btest.AssertContains(ft, s, "Hello")
	if ft.failed {
		t.Error("AssertContains should not fail for existing text")
	}
}

func TestAssertContainsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	btest.AssertContains(ft, s, "Missing")
	if !ft.failed {
		t.Error("AssertContains should fail for missing text")
	}
}

func TestAssertContainsAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	btest.AssertContainsAt(ft, s, 0, 0, "Hello")
	if ft.failed {
		t.Error("AssertContainsAt should not fail for correct position")
	}
}

func TestAssertContainsAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	btest.AssertContainsAt(ft, s, 0, 1, "Hello")
	if !ft.failed {
		t.Error("AssertContainsAt should fail for wrong position")
	}
}

func TestAssertNotContainsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	btest.AssertNotContains(ft, s, "Missing")
	if ft.failed {
		t.Error("AssertNotContains should not fail for missing text")
	}
}

func TestAssertNotContainsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	btest.AssertNotContains(ft, s, "Hello")
	if !ft.failed {
		t.Error("AssertNotContains should fail for present text")
	}
}

func TestAssertCursorAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hi")

	ft := &fakeTB{}
	btest.AssertCursorAt(ft, s, 0, 2)
	if ft.failed {
		t.Error("AssertCursorAt should not fail for correct position")
	}
}

func TestAssertCursorAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hi")

	ft := &fakeTB{}
	btest.AssertCursorAt(ft, s, 1, 0)
	if !ft.failed {
		t.Error("AssertCursorAt should fail for wrong position")
	}
}

func TestAssertStyleAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m")

	ft := &fakeTB{}
	btest.AssertStyleAt(ft, s, 0, 0, btest.CellStyle{Bold: true})
	if ft.failed {
		t.Error("AssertStyleAt should not fail for matching style")
	}
}

func TestAssertStyleAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Normal")

	ft := &fakeTB{}
	btest.AssertStyleAt(ft, s, 0, 0, btest.CellStyle{Bold: true})
	if !ft.failed {
		t.Error("AssertStyleAt should fail for non-matching style")
	}
}

func TestAssertBoldAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m")

	ft := &fakeTB{}
	btest.AssertBoldAt(ft, s, 0, 0)
	if ft.failed {
		t.Error("AssertBoldAt should not fail for bold cell")
	}
}

func TestAssertBoldAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Normal")

	ft := &fakeTB{}
	btest.AssertBoldAt(ft, s, 0, 0)
	if !ft.failed {
		t.Error("AssertBoldAt should fail for non-bold cell")
	}
}

func TestAssertRowContainsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!\nSecond line")

	ft := &fakeTB{}
	btest.AssertRowContains(ft, s, 1, "Second")
	if ft.failed {
		t.Error("AssertRowContains should not fail for present text")
	}
}

func TestAssertRowContainsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	btest.AssertRowContains(ft, s, 0, "Missing")
	if !ft.failed {
		t.Error("AssertRowContains should fail for missing text")
	}
}

// ── Golden tests ───────────────────────────────────────────────────────────

func TestGoldenCreateAndCompare(t *testing.T) {
	track(t)
	s := btest.NewScreen(20, 3)
	s.Render("Golden test\nLine 2")

	first := s.String()
	second := s.String()
	if first != second {
		t.Error("String() should be deterministic across calls")
	}
}

// ── TestModel tests ────────────────────────────────────────────────────────

type stubModel struct {
	text   string
	width  int
	height int
	keys   []string
}

func (m stubModel) Init() tea.Cmd { return nil }

func (m stubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		m.keys = append(m.keys, msg.String())
		if msg.String() == "a" {
			m.text += "a"
		}
	}
	return m, nil
}

func (m stubModel) View() string {
	if m.text == "" {
		return "empty"
	}
	return m.text
}

func TestTestModelScreen(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	scr := tm.Screen()

	if !scr.Contains("empty") {
		t.Error("initial screen should contain 'empty'")
	}
}

func TestTestModelSendKey(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendKey("a")

	scr := tm.Screen()
	if !scr.Contains("a") {
		t.Error("screen should contain 'a' after key press")
	}
}

func TestTestModelType(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	tm.Type("aaa")

	scr := tm.Screen()
	if !scr.Contains("aaa") {
		t.Error("screen should contain 'aaa' after typing")
	}
}

func TestTestModelSendResize(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendResize(80, 24)

	m := tm.Model().(stubModel)
	if m.width != 80 || m.height != 24 {
		t.Errorf("model size = (%d, %d), want (80, 24)", m.width, m.height)
	}
}

func TestTestModelReturnsTea(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	m := tm.Model()
	if _, ok := m.(stubModel); !ok {
		t.Error("Model() should return the underlying stubModel")
	}
}

func TestTestModelSendKeys(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendKeys("a", "a", "a")

	scr := tm.Screen()
	if !scr.Contains("aaa") {
		t.Error("screen should contain 'aaa' after SendKeys")
	}
}

func TestTestModelSendMsg(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendMsg(tea.WindowSizeMsg{Width: 100, Height: 50})

	m := tm.Model().(stubModel)
	if m.width != 100 || m.height != 50 {
		t.Errorf("model size = (%d, %d), want (100, 50)", m.width, m.height)
	}
}

func TestTestModelColsLines(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 80, 24)
	if tm.Cols() != 80 {
		t.Errorf("Cols() = %d, want 80", tm.Cols())
	}
	if tm.Lines() != 24 {
		t.Errorf("Lines() = %d, want 24", tm.Lines())
	}
}

func TestTestModelRequireScreen(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)

	called := false
	tm.RequireScreen(func(tb testing.TB, s *btest.Screen) {
		called = true
		if !s.Contains("empty") {
			tb.Error("expected 'empty'")
		}
	})
	if !called {
		t.Error("RequireScreen callback was not invoked")
	}
}

func TestTestModelWaitFor(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	// Screen starts with "empty", so UntilContains("empty") should succeed immediately.
	ok := tm.WaitFor(btest.UntilContains("empty"), 5)
	if !ok {
		t.Error("WaitFor(UntilContains) should succeed for existing text")
	}
}

func TestTestModelWaitForTimeout(t *testing.T) {
	track(t)
	tm := btest.NewTestModel(t, stubModel{}, 40, 5)
	ok := tm.WaitFor(btest.UntilContains("never-here"), 3)
	if ok {
		t.Error("WaitFor should return false when text never appears")
	}
}

// ── Screen extended tests ─────────────────────────────────────────────────

func TestScreenFindText(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello\nWorld\nFoo Bar")

	row, col := s.FindText("World")
	if row != 1 || col != 0 {
		t.Errorf("FindText(World) = (%d, %d), want (1, 0)", row, col)
	}

	row, col = s.FindText("Bar")
	if row != 2 || col != 4 {
		t.Errorf("FindText(Bar) = (%d, %d), want (2, 4)", row, col)
	}

	row, col = s.FindText("Missing")
	if row != -1 || col != -1 {
		t.Errorf("FindText(Missing) = (%d, %d), want (-1, -1)", row, col)
	}
}

func TestScreenFindAllText(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("ab ab\nab")

	results := s.FindAllText("ab")
	if len(results) != 3 {
		t.Errorf("FindAllText(ab) found %d, want 3", len(results))
	}
}

func TestScreenRowCount(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Line 1\nLine 2\nLine 3")

	if got := s.RowCount(); got != 3 {
		t.Errorf("RowCount() = %d, want 3", got)
	}
}

func TestScreenAllRows(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 3)
	s.Render("A\nB\nC")

	rows := s.AllRows()
	if len(rows) != 3 {
		t.Fatalf("AllRows() len = %d, want 3", len(rows))
	}
	if rows[0] != "A" || rows[1] != "B" || rows[2] != "C" {
		t.Errorf("AllRows() = %v, want [A B C]", rows)
	}
}

func TestScreenNonEmptyRows(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("A\n\nC")

	rows := s.NonEmptyRows()
	if len(rows) != 2 {
		t.Fatalf("NonEmptyRows() len = %d, want 2", len(rows))
	}
	if rows[0].Index != 0 || rows[0].Text != "A" {
		t.Errorf("NonEmptyRows[0] = %+v, want {0, A}", rows[0])
	}
	if rows[1].Index != 2 || rows[1].Text != "C" {
		t.Errorf("NonEmptyRows[1] = %+v, want {2, C}", rows[1])
	}
}

func TestScreenColumn(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 3)
	s.Render("ABC\nDEF\nGHI")

	col := s.Column(1, 0, 3)
	if col != "B\nE\nH" {
		t.Errorf("Column(1, 0, 3) = %q, want %q", col, "B\nE\nH")
	}
}

func TestScreenIsEmpty(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("")
	if !s.IsEmpty() {
		t.Error("IsEmpty() should be true for blank screen")
	}

	s.Render("text")
	if s.IsEmpty() {
		t.Error("IsEmpty() should be false for non-blank screen")
	}
}

func TestScreenCountOccurrences(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("aa aa\naa")

	if got := s.CountOccurrences("aa"); got != 3 {
		t.Errorf("CountOccurrences(aa) = %d, want 3", got)
	}
	if got := s.CountOccurrences("zz"); got != 0 {
		t.Errorf("CountOccurrences(zz) = %d, want 0", got)
	}
}

func TestScreenMatchesRegexp(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Error: 404 not found")

	if !s.MatchesRegexp(`\d{3}`) {
		t.Error("MatchesRegexp should find 3-digit number")
	}
	if s.MatchesRegexp(`\d{5}`) {
		t.Error("MatchesRegexp should not find 5-digit number")
	}
}

func TestScreenFindRegexp(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Name: Alice\nAge: 30")

	row, col := s.FindRegexp(`\d+`)
	if row != 1 || col != 5 {
		t.Errorf("FindRegexp(\\d+) = (%d, %d), want (1, 5)", row, col)
	}
}

func TestScreenSize(t *testing.T) {
	track(t)
	s := btest.NewScreen(80, 24)
	cols, lines := s.Size()
	if cols != 80 || lines != 24 {
		t.Errorf("Size() = (%d, %d), want (80, 24)", cols, lines)
	}
}

// ── Region extended tests ─────────────────────────────────────────────────

func TestRegionRowCount(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA\nBBBB\n\nDDDD")

	r := s.Region(0, 0, 4, 4)
	if got := r.RowCount(); got != 3 {
		t.Errorf("Region.RowCount() = %d, want 3", got)
	}
}

func TestRegionIsEmpty(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA")

	r := s.Region(2, 0, 4, 2) // empty area
	if !r.IsEmpty() {
		t.Error("Region.IsEmpty() should be true for blank region")
	}
}

func TestRegionFindText(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA BBBB\nCCCC DDDD")

	r := s.Region(0, 5, 4, 2)
	row, col := r.FindText("DDDD")
	if row != 1 || col != 0 {
		t.Errorf("Region.FindText(DDDD) = (%d, %d), want (1, 0)", row, col)
	}
}

func TestRegionStyleAt(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m Normal")

	r := s.Region(0, 0, 10, 1)
	style := r.StyleAt(0, 0)
	if !style.Bold {
		t.Error("Region.StyleAt should return bold at (0, 0)")
	}
	style = r.StyleAt(0, 5)
	if style.Bold {
		t.Error("Region.StyleAt should return non-bold at (0, 5)")
	}
}

func TestRegionAllRows(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 3)
	s.Render("AB CD\nEF GH\nIJ KL")

	r := s.Region(0, 3, 2, 3)
	rows := r.AllRows()
	if len(rows) != 3 {
		t.Fatalf("Region.AllRows() len = %d, want 3", len(rows))
	}
	if rows[0] != "CD" || rows[1] != "GH" || rows[2] != "KL" {
		t.Errorf("Region.AllRows() = %v, want [CD GH KL]", rows)
	}
}

func TestRegionCountOccurrences(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 3)
	s.Render("ab ab\nab ab\nab")

	r := s.Region(0, 0, 5, 2)
	if got := r.CountOccurrences("ab"); got != 4 {
		t.Errorf("Region.CountOccurrences(ab) = %d, want 4", got)
	}
}

// ── Assert extended tests ─────────────────────────────────────────────────

func TestAssertRowEqualsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Exact Match")

	ft := &fakeTB{}
	btest.AssertRowEquals(ft, s, 0, "Exact Match")
	if ft.failed {
		t.Error("AssertRowEquals should not fail for exact match")
	}
}

func TestAssertRowEqualsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Exact Match")

	ft := &fakeTB{}
	btest.AssertRowEquals(ft, s, 0, "Wrong")
	if !ft.failed {
		t.Error("AssertRowEquals should fail for non-match")
	}
}

func TestAssertRowCountPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("A\nB\nC")

	ft := &fakeTB{}
	btest.AssertRowCount(ft, s, 3)
	if ft.failed {
		t.Error("AssertRowCount should not fail for correct count")
	}
}

func TestAssertRowCountFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("A\nB\nC")

	ft := &fakeTB{}
	btest.AssertRowCount(ft, s, 5)
	if !ft.failed {
		t.Error("AssertRowCount should fail for wrong count")
	}
}

func TestAssertEmptyPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("")

	ft := &fakeTB{}
	btest.AssertEmpty(ft, s)
	if ft.failed {
		t.Error("AssertEmpty should not fail for empty screen")
	}
}

func TestAssertEmptyFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("text")

	ft := &fakeTB{}
	btest.AssertEmpty(ft, s)
	if !ft.failed {
		t.Error("AssertEmpty should fail for non-empty screen")
	}
}

func TestAssertNotEmptyPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("text")

	ft := &fakeTB{}
	btest.AssertNotEmpty(ft, s)
	if ft.failed {
		t.Error("AssertNotEmpty should not fail for non-empty screen")
	}
}

func TestAssertNotEmptyFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("")

	ft := &fakeTB{}
	btest.AssertNotEmpty(ft, s)
	if !ft.failed {
		t.Error("AssertNotEmpty should fail for empty screen")
	}
}

func TestAssertMatchesPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Error: 404 not found")

	ft := &fakeTB{}
	btest.AssertMatches(ft, s, `\d{3}`)
	if ft.failed {
		t.Error("AssertMatches should not fail for matching pattern")
	}
}

func TestAssertMatchesFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello World")

	ft := &fakeTB{}
	btest.AssertMatches(ft, s, `\d+`)
	if !ft.failed {
		t.Error("AssertMatches should fail for non-matching pattern")
	}
}

func TestAssertRowMatchesPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Name: Alice\nAge: 30")

	ft := &fakeTB{}
	btest.AssertRowMatches(ft, s, 1, `Age: \d+`)
	if ft.failed {
		t.Error("AssertRowMatches should not fail for matching row")
	}
}

func TestAssertRowMatchesFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Name: Alice")

	ft := &fakeTB{}
	btest.AssertRowMatches(ft, s, 0, `\d+`)
	if !ft.failed {
		t.Error("AssertRowMatches should fail for non-matching row")
	}
}

func TestAssertContainsCountPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("ab ab\nab")

	ft := &fakeTB{}
	btest.AssertContainsCount(ft, s, "ab", 3)
	if ft.failed {
		t.Error("AssertContainsCount should not fail for correct count")
	}
}

func TestAssertContainsCountFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("ab ab\nab")

	ft := &fakeTB{}
	btest.AssertContainsCount(ft, s, "ab", 5)
	if !ft.failed {
		t.Error("AssertContainsCount should fail for wrong count")
	}
}

func TestAssertFgAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[31mRed\x1b[0m")

	ft := &fakeTB{}
	btest.AssertFgAt(ft, s, 0, 0, "red")
	if ft.failed {
		t.Error("AssertFgAt should not fail for correct color")
	}
}

func TestAssertFgAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[31mRed\x1b[0m")

	ft := &fakeTB{}
	btest.AssertFgAt(ft, s, 0, 0, "blue")
	if !ft.failed {
		t.Error("AssertFgAt should fail for wrong color")
	}
}

func TestAssertBgAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[44mBlue\x1b[0m")

	ft := &fakeTB{}
	btest.AssertBgAt(ft, s, 0, 0, "blue")
	if ft.failed {
		t.Error("AssertBgAt should not fail for correct color")
	}
}

func TestAssertItalicAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[3mItalic\x1b[0m")

	ft := &fakeTB{}
	btest.AssertItalicAt(ft, s, 0, 0)
	if ft.failed {
		t.Error("AssertItalicAt should not fail for italic cell")
	}
}

func TestAssertUnderlineAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("\x1b[4mUnderline\x1b[0m")

	ft := &fakeTB{}
	btest.AssertUnderlineAt(ft, s, 0, 0)
	if ft.failed {
		t.Error("AssertUnderlineAt should not fail for underlined cell")
	}
}

func TestAssertTextAtPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	btest.AssertTextAt(ft, s, 0, 7, "World")
	if ft.failed {
		t.Error("AssertTextAt should not fail for correct text")
	}
}

func TestAssertTextAtFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	btest.AssertTextAt(ft, s, 0, 7, "Wrong")
	if !ft.failed {
		t.Error("AssertTextAt should fail for wrong text")
	}
}

func TestAssertRegionContainsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA BBBB\nCCCC DDDD")

	ft := &fakeTB{}
	btest.AssertRegionContains(ft, s, 0, 5, 4, 2, "BBBB")
	if ft.failed {
		t.Error("AssertRegionContains should not fail for present text")
	}
}

func TestAssertRegionContainsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("AAAA BBBB\nCCCC DDDD")

	ft := &fakeTB{}
	btest.AssertRegionContains(ft, s, 0, 5, 4, 2, "AAAA")
	if !ft.failed {
		t.Error("AssertRegionContains should fail for absent text")
	}
}

func TestAssertScreenEqualsPass(t *testing.T) {
	track(t)
	s := btest.NewScreen(20, 3)
	s.Render("Row 0\nRow 1\nRow 2")

	ft := &fakeTB{}
	btest.AssertScreenEquals(ft, s, "Row 0\nRow 1\nRow 2")
	if ft.failed {
		t.Error("AssertScreenEquals should not fail for identical content")
	}
}

func TestAssertScreenEqualsFail(t *testing.T) {
	track(t)
	s := btest.NewScreen(20, 3)
	s.Render("Row 0\nRow 1\nRow 2")

	ft := &fakeTB{}
	btest.AssertScreenEquals(ft, s, "Row 0\nWrong\nRow 2")
	if !ft.failed {
		t.Error("AssertScreenEquals should fail for different content")
	}
}

// ── Diff tests ────────────────────────────────────────────────────────────

func TestDiffIdentical(t *testing.T) {
	track(t)
	a := btest.NewScreen(20, 3)
	a.Render("Same\nContent")
	b := btest.NewScreen(20, 3)
	b.Render("Same\nContent")

	diff := btest.ScreenDiff(a, b)
	if diff.HasChanges() {
		t.Error("ScreenDiff should have no changes for identical screens")
	}
}

func TestDiffChanged(t *testing.T) {
	track(t)
	a := btest.NewScreen(20, 3)
	a.Render("Line A\nLine B")
	b := btest.NewScreen(20, 3)
	b.Render("Line A\nLine C")

	diff := btest.ScreenDiff(a, b)
	if !diff.HasChanges() {
		t.Error("ScreenDiff should detect changes")
	}
	changed := diff.ChangedLines()
	if len(changed) != 1 {
		t.Fatalf("ChangedLines() = %d, want 1", len(changed))
	}
	if changed[0].Row != 1 {
		t.Errorf("changed row = %d, want 1", changed[0].Row)
	}
}

func TestDiffString(t *testing.T) {
	track(t)
	a := btest.NewScreen(20, 2)
	a.Render("Same\nOld")
	b := btest.NewScreen(20, 2)
	b.Render("Same\nNew")

	diff := btest.ScreenDiff(a, b)
	output := diff.String()
	if !strings.Contains(output, "row 1") {
		t.Error("diff output should mention the changed row")
	}
}

func TestAssertScreensEqualPass(t *testing.T) {
	track(t)
	a := btest.NewScreen(20, 3)
	a.Render("Same")
	b := btest.NewScreen(20, 3)
	b.Render("Same")

	ft := &fakeTB{}
	btest.AssertScreensEqual(ft, a, b)
	if ft.failed {
		t.Error("AssertScreensEqual should not fail for identical screens")
	}
}

func TestAssertScreensNotEqualPass(t *testing.T) {
	track(t)
	a := btest.NewScreen(20, 3)
	a.Render("A")
	b := btest.NewScreen(20, 3)
	b.Render("B")

	ft := &fakeTB{}
	btest.AssertScreensNotEqual(ft, a, b)
	if ft.failed {
		t.Error("AssertScreensNotEqual should not fail for different screens")
	}
}

// ── Predicate tests ───────────────────────────────────────────────────────

func TestUntilContains(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello World")

	pred := btest.UntilContains("World")
	if !pred(s) {
		t.Error("UntilContains should return true when text exists")
	}
}

func TestUntilNotContains(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("Hello World")

	pred := btest.UntilNotContains("Missing")
	if !pred(s) {
		t.Error("UntilNotContains should return true when text is absent")
	}
}

func TestUntilRowContains(t *testing.T) {
	track(t)
	s := btest.NewScreen(40, 5)
	s.Render("First\nSecond")

	pred := btest.UntilRowContains(1, "Second")
	if !pred(s) {
		t.Error("UntilRowContains should return true when row contains text")
	}
}

// ── KeyNames test ─────────────────────────────────────────────────────────

func TestKeyNames(t *testing.T) {
	track(t)
	names := btest.KeyNames()
	if len(names) < 30 {
		t.Errorf("KeyNames() returned %d keys, want at least 30", len(names))
	}
}

// ── Stopwatch test ────────────────────────────────────────────────────────

func TestStopwatch(t *testing.T) {
	track(t)
	sw := btest.StartStopwatch()
	elapsed := sw.Elapsed()
	if elapsed < 0 {
		t.Error("Stopwatch.Elapsed() should be non-negative")
	}
}
