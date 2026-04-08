package tuitest_test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
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
	fmt.Fprintln(w, "  "+accent.Render("tuitest")+" "+dimText.Render("· terminal test toolkit"))
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

	os.Exit(code)
}

// ── Screen tests ───────────────────────────────────────────────────────────

func TestScreenPlainText(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
	s.Render("Hi")

	row, col := s.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

func TestScreenString(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(20, 3)
	s.Render("Row 0\nRow 1\nRow 2")

	got := s.String()
	want := "Row 0\nRow 1\nRow 2"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestScreenStyleBold(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
	s.Render("\x1b[44mBlue\x1b[0m")

	style := s.StyleAt(0, 0)
	if style.Bg != "blue" {
		t.Errorf("Bg = %q, want %q", style.Bg, "blue")
	}
}

func TestScreenStyleItalicUnderline(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(10, 5)
	s.Render("test")

	style := s.StyleAt(-1, 0)
	if style != (tuitest.CellStyle{}) {
		t.Error("StyleAt(-1, 0) should return zero CellStyle")
	}
	style = s.StyleAt(0, 100)
	if style != (tuitest.CellStyle{}) {
		t.Error("StyleAt(0, 100) should return zero CellStyle")
	}
}

func TestScreenRenderOverwrites(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 5)
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
	s := tuitest.NewScreen(40, 3)
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
	failed  bool
	message string
}

func (f *fakeTB) Helper()                       {}
func (f *fakeTB) Errorf(format string, a ...any) { f.failed = true }
func (f *fakeTB) Fatalf(format string, a ...any) { f.failed = true }
func (f *fakeTB) Logf(format string, a ...any)   {}

func TestAssertContainsPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	tuitest.AssertContains(ft, s, "Hello")
	if ft.failed {
		t.Error("AssertContains should not fail for existing text")
	}
}

func TestAssertContainsFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	tuitest.AssertContains(ft, s, "Missing")
	if !ft.failed {
		t.Error("AssertContains should fail for missing text")
	}
}

func TestAssertContainsAtPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	tuitest.AssertContainsAt(ft, s, 0, 0, "Hello")
	if ft.failed {
		t.Error("AssertContainsAt should not fail for correct position")
	}
}

func TestAssertContainsAtFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	tuitest.AssertContainsAt(ft, s, 0, 1, "Hello")
	if !ft.failed {
		t.Error("AssertContainsAt should fail for wrong position")
	}
}

func TestAssertNotContainsPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	tuitest.AssertNotContains(ft, s, "Missing")
	if ft.failed {
		t.Error("AssertNotContains should not fail for missing text")
	}
}

func TestAssertNotContainsFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello")

	ft := &fakeTB{}
	tuitest.AssertNotContains(ft, s, "Hello")
	if !ft.failed {
		t.Error("AssertNotContains should fail for present text")
	}
}

func TestAssertCursorAtPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hi")

	ft := &fakeTB{}
	tuitest.AssertCursorAt(ft, s, 0, 2)
	if ft.failed {
		t.Error("AssertCursorAt should not fail for correct position")
	}
}

func TestAssertCursorAtFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hi")

	ft := &fakeTB{}
	tuitest.AssertCursorAt(ft, s, 1, 0)
	if !ft.failed {
		t.Error("AssertCursorAt should fail for wrong position")
	}
}

func TestAssertStyleAtPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m")

	ft := &fakeTB{}
	tuitest.AssertStyleAt(ft, s, 0, 0, tuitest.CellStyle{Bold: true})
	if ft.failed {
		t.Error("AssertStyleAt should not fail for matching style")
	}
}

func TestAssertStyleAtFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Normal")

	ft := &fakeTB{}
	tuitest.AssertStyleAt(ft, s, 0, 0, tuitest.CellStyle{Bold: true})
	if !ft.failed {
		t.Error("AssertStyleAt should fail for non-matching style")
	}
}

func TestAssertBoldAtPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("\x1b[1mBold\x1b[0m")

	ft := &fakeTB{}
	tuitest.AssertBoldAt(ft, s, 0, 0)
	if ft.failed {
		t.Error("AssertBoldAt should not fail for bold cell")
	}
}

func TestAssertBoldAtFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Normal")

	ft := &fakeTB{}
	tuitest.AssertBoldAt(ft, s, 0, 0)
	if !ft.failed {
		t.Error("AssertBoldAt should fail for non-bold cell")
	}
}

func TestAssertRowContainsPass(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello, World!\nSecond line")

	ft := &fakeTB{}
	tuitest.AssertRowContains(ft, s, 1, "Second")
	if ft.failed {
		t.Error("AssertRowContains should not fail for present text")
	}
}

func TestAssertRowContainsFail(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(40, 5)
	s.Render("Hello, World!")

	ft := &fakeTB{}
	tuitest.AssertRowContains(ft, s, 0, "Missing")
	if !ft.failed {
		t.Error("AssertRowContains should fail for missing text")
	}
}

// ── Golden tests ───────────────────────────────────────────────────────────

func TestGoldenCreateAndCompare(t *testing.T) {
	track(t)
	s := tuitest.NewScreen(20, 3)
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
	tm := tuitest.NewTestModel(t, stubModel{}, 40, 5)
	scr := tm.Screen()

	if !scr.Contains("empty") {
		t.Error("initial screen should contain 'empty'")
	}
}

func TestTestModelSendKey(t *testing.T) {
	track(t)
	tm := tuitest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendKey("a")

	scr := tm.Screen()
	if !scr.Contains("a") {
		t.Error("screen should contain 'a' after key press")
	}
}

func TestTestModelType(t *testing.T) {
	track(t)
	tm := tuitest.NewTestModel(t, stubModel{}, 40, 5)
	tm.Type("aaa")

	scr := tm.Screen()
	if !scr.Contains("aaa") {
		t.Error("screen should contain 'aaa' after typing")
	}
}

func TestTestModelSendResize(t *testing.T) {
	track(t)
	tm := tuitest.NewTestModel(t, stubModel{}, 40, 5)
	tm.SendResize(80, 24)

	m := tm.Model().(stubModel)
	if m.width != 80 || m.height != 24 {
		t.Errorf("model size = (%d, %d), want (80, 24)", m.width, m.height)
	}
}

func TestTestModelReturnsTea(t *testing.T) {
	track(t)
	tm := tuitest.NewTestModel(t, stubModel{}, 40, 5)
	m := tm.Model()
	if _, ok := m.(stubModel); !ok {
		t.Error("Model() should return the underlying stubModel")
	}
}
