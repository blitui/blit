package blit

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestLogViewer() *LogViewer {
	lv := NewLogViewer()
	lv.SetTheme(DefaultTheme())
	lv.SetSize(80, 20)
	return lv
}

func makeLogLine(level LogLevel, msg, src string) LogLine {
	return LogLine{
		Level:     level,
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Message:   msg,
		Source:    src,
	}
}

func TestLogViewer_AppendAndLines(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "hello world", "app"))
	lv.Append(makeLogLine(LogError, "something failed", "db"))

	lines := lv.Lines()
	if len(lines) != 2 {
		t.Fatalf("Lines() = %d, want 2", len(lines))
	}
	if lines[0].Message != "hello world" {
		t.Errorf("lines[0].Message = %q, want %q", lines[0].Message, "hello world")
	}
	if lines[1].Level != LogError {
		t.Errorf("lines[1].Level = %d, want LogError", lines[1].Level)
	}
}

func TestLogViewer_Clear(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "msg", ""))
	lv.Append(makeLogLine(LogWarn, "msg2", ""))
	lv.Clear()

	if n := len(lv.Lines()); n != 0 {
		t.Errorf("after Clear(), Lines() = %d, want 0", n)
	}
}

func TestLogViewer_LevelFilter(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogDebug, "debug msg", ""))
	lv.Append(makeLogLine(LogInfo, "info msg", ""))
	lv.Append(makeLogLine(LogWarn, "warn msg", ""))
	lv.Append(makeLogLine(LogError, "error msg", ""))

	// Default: debug+ shows all
	if len(lv.filteredLines) != 4 {
		t.Errorf("debug+ filter: got %d lines, want 4", len(lv.filteredLines))
	}

	// Cycle to info+
	lv.cycleLevel()
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 3 {
		t.Errorf("info+ filter: got %d lines, want 3", len(lv.filteredLines))
	}

	// Cycle to warn+
	lv.cycleLevel()
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 2 {
		t.Errorf("warn+ filter: got %d lines, want 2", len(lv.filteredLines))
	}

	// Cycle to error
	lv.cycleLevel()
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 1 {
		t.Errorf("error filter: got %d lines, want 1", len(lv.filteredLines))
	}

	// Cycle back to debug
	lv.cycleLevel()
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 4 {
		t.Errorf("after full cycle, debug+ filter: got %d lines, want 4", len(lv.filteredLines))
	}
}

func TestLogViewer_SubstringFilter(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "database connected", "db"))
	lv.Append(makeLogLine(LogInfo, "user logged in", "auth"))
	lv.Append(makeLogLine(LogError, "database timeout", "db"))

	lv.filterText = "database"
	lv.rebuildFiltered()

	if len(lv.filteredLines) != 2 {
		t.Errorf("filter 'database': got %d lines, want 2", len(lv.filteredLines))
	}

	// Filter by source
	lv.filterText = "auth"
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 1 {
		t.Errorf("filter 'auth' (source): got %d lines, want 1", len(lv.filteredLines))
	}

	// Clear filter
	lv.filterText = ""
	lv.rebuildFiltered()
	if len(lv.filteredLines) != 3 {
		t.Errorf("cleared filter: got %d lines, want 3", len(lv.filteredLines))
	}
}

func TestLogViewer_PauseResumeKey(t *testing.T) {
	lv := newTestLogViewer()
	lv.SetFocused(true)

	if lv.paused {
		t.Error("should not start paused")
	}

	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}, Context{})
	if cmd == nil {
		t.Error("p key should return consumed cmd")
	}
	if !lv.paused {
		t.Error("should be paused after p")
	}

	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}, Context{})
	if lv.paused {
		t.Error("should be resumed after second p")
	}
}

func TestLogViewer_ClearKey(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "msg", ""))
	lv.Append(makeLogLine(LogInfo, "msg2", ""))

	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}, Context{})
	if n := len(lv.Lines()); n != 0 {
		t.Errorf("after c key, Lines() = %d, want 0", n)
	}
}

func TestLogViewer_EndKeyJumpsToBottom(t *testing.T) {
	lv := newTestLogViewer()
	for i := 0; i < 50; i++ {
		lv.Append(makeLogLine(LogInfo, "line", ""))
	}

	// Scroll up first
	lv.Update(tea.KeyMsg{Type: tea.KeyUp}, Context{})
	lv.paused = true

	// End jumps down and resumes
	lv.Update(tea.KeyMsg{Type: tea.KeyEnd}, Context{})
	if lv.paused {
		t.Error("end key should resume auto-scroll")
	}
	if !lv.viewport.AtBottom() {
		t.Error("end key should scroll to bottom")
	}
}

func TestLogViewer_ScrollUpPausesAutoScroll(t *testing.T) {
	lv := newTestLogViewer()
	for i := 0; i < 50; i++ {
		lv.Append(makeLogLine(LogInfo, "line", ""))
	}

	lv.Update(tea.KeyMsg{Type: tea.KeyUp}, Context{})
	if !lv.paused {
		t.Error("scrolling up should pause auto-scroll")
	}
}

func TestLogViewer_LevelCycleKey(t *testing.T) {
	lv := newTestLogViewer()
	if lv.filterLevel != LogDebug {
		t.Errorf("initial filterLevel = %d, want LogDebug", lv.filterLevel)
	}

	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, Context{})
	if lv.filterLevel != LogInfo {
		t.Errorf("after l, filterLevel = %d, want LogInfo", lv.filterLevel)
	}
}

func TestLogViewer_ViewRendersContent(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "hello world", "myapp"))
	lv.Append(makeLogLine(LogError, "boom", "db"))

	view := lv.View()
	if !strings.Contains(view, "hello world") {
		t.Errorf("view missing 'hello world': %s", view)
	}
	if !strings.Contains(view, "boom") {
		t.Errorf("view missing 'boom': %s", view)
	}
}

func TestLogViewer_ViewContainsLevelChips(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogDebug, "d", ""))
	lv.Append(makeLogLine(LogInfo, "i", ""))
	lv.Append(makeLogLine(LogWarn, "w", ""))
	lv.Append(makeLogLine(LogError, "e", ""))

	view := lv.View()
	for _, chip := range []string{"DBG", "INF", "WRN", "ERR"} {
		if !strings.Contains(view, chip) {
			t.Errorf("view missing level chip %q", chip)
		}
	}
}

func TestLogViewer_AppendGoroutineSafe(t *testing.T) {
	lv := newTestLogViewer()
	done := make(chan struct{})
	const n = 200
	go func() {
		for i := 0; i < n; i++ {
			lv.Append(makeLogLine(LogInfo, "concurrent", "goroutine"))
		}
		close(done)
	}()
	for i := 0; i < n; i++ {
		lv.Append(makeLogLine(LogDebug, "main", "main"))
	}
	<-done

	total := len(lv.Lines())
	if total != 2*n {
		t.Errorf("after concurrent appends, Lines() = %d, want %d", total, 2*n)
	}
}

func TestLogViewer_KeyBindings(t *testing.T) {
	lv := newTestLogViewer()
	binds := lv.KeyBindings()
	if len(binds) < 7 {
		t.Errorf("expected >= 7 keybindings, got %d", len(binds))
	}
	keys := make(map[string]bool)
	for _, b := range binds {
		keys[b.Key] = true
	}
	for _, want := range []string{"p", "c", "/", "l", "end"} {
		if !keys[want] {
			t.Errorf("missing keybinding %q", want)
		}
	}
}

func TestLogViewer_InitReturnsNil(t *testing.T) {
	lv := newTestLogViewer()
	if cmd := lv.Init(); cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestLogViewer_SetFocused(t *testing.T) {
	lv := newTestLogViewer()
	if lv.Focused() {
		t.Error("new LogViewer should not be focused")
	}
	lv.SetFocused(true)
	if !lv.Focused() {
		t.Error("SetFocused(true) should set focus")
	}
}

func TestLogViewer_LogAppendMsg(t *testing.T) {
	lv := newTestLogViewer()
	line := makeLogLine(LogWarn, "via message", "src")
	lv.Update(LogAppendMsg{Line: line}, Context{})

	lines := lv.Lines()
	if len(lines) != 1 {
		t.Fatalf("Lines() = %d after LogAppendMsg, want 1", len(lines))
	}
	if lines[0].Message != "via message" {
		t.Errorf("Lines()[0].Message = %q, want 'via message'", lines[0].Message)
	}
}

func TestLogViewer_FilterInputMode(t *testing.T) {
	lv := newTestLogViewer()
	lv.Append(makeLogLine(LogInfo, "match me", ""))
	lv.Append(makeLogLine(LogInfo, "other", ""))

	// Enter filter mode
	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, Context{})
	if !lv.filtering {
		t.Fatal("/ should activate filter mode")
	}

	// Type filter text
	for _, ch := range "match" {
		lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}, Context{})
	}

	// Confirm with Enter
	lv.Update(tea.KeyMsg{Type: tea.KeyEnter}, Context{})
	if lv.filtering {
		t.Error("Enter should exit filter mode")
	}
	if lv.filterText != "match" {
		t.Errorf("filterText = %q, want 'match'", lv.filterText)
	}
	if len(lv.filteredLines) != 1 {
		t.Errorf("filtered lines = %d, want 1", len(lv.filteredLines))
	}
}
