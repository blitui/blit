package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	blit "github.com/blitui/blit"
	tea "github.com/charmbracelet/bubbletea"
)

// newTestModel returns a watchModel sized for tests.
func newTestModel(t *testing.T) *watchModel {
	t.Helper()
	m := newWatchModel([]string{"./testpkg/..."})
	m.width = 80
	m.height = 24
	m.statusBar.SetSize(80, 1)
	return m
}

// sendKey drives a single rune key through the model and returns the updated model.
func sendKey(t *testing.T, m *watchModel, key string) *watchModel {
	t.Helper()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	wm, ok := next.(*watchModel)
	if !ok {
		t.Fatalf("Update returned unexpected type %T", next)
	}
	return wm
}

// sendKeySpecial drives a named special key through the model.
func sendKeySpecial(t *testing.T, m *watchModel, keyType tea.KeyType) *watchModel {
	t.Helper()
	next, _ := m.Update(tea.KeyMsg{Type: keyType})
	wm, ok := next.(*watchModel)
	if !ok {
		t.Fatalf("Update returned unexpected type %T", next)
	}
	return wm
}

// ── A5: key routing ──────────────────────────────────────────────────────────

func TestKeyRouting_Quit(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected a cmd from q, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg from q")
	}
}

func TestKeyRouting_ToggleUpdateSnap(t *testing.T) {
	m := newTestModel(t)
	if m.filters.UpdateSnap {
		t.Fatal("UpdateSnap should start false")
	}
	m = sendKey(t, m, "u")
	if !m.filters.UpdateSnap {
		t.Fatal("UpdateSnap should be true after u")
	}
	m = sendKey(t, m, "u")
	if m.filters.UpdateSnap {
		t.Fatal("UpdateSnap should toggle back to false")
	}
}

func TestKeyRouting_CycleLogLevel(t *testing.T) {
	m := newTestModel(t)
	if m.filters.LogLevel != WatchLogQuiet {
		t.Fatalf("expected initial level 0 (quiet), got %v", m.filters.LogLevel)
	}
	m = sendKey(t, m, "l")
	if m.filters.LogLevel != WatchLogNormal {
		t.Fatalf("expected normal after l, got %v", m.filters.LogLevel)
	}
	m = sendKey(t, m, "l")
	if m.filters.LogLevel != WatchLogVerbose {
		t.Fatalf("expected verbose after ll, got %v", m.filters.LogLevel)
	}
	m = sendKey(t, m, "l")
	if m.filters.LogLevel != WatchLogQuiet {
		t.Fatalf("expected quiet after lll, got %v", m.filters.LogLevel)
	}
}

func TestKeyRouting_ClearFilters(t *testing.T) {
	m := newTestModel(t)
	m.filters.Path = "some/path"
	m.filters.NameRegex = "TestFoo"
	m.filters.FailedOnly = true

	m = sendKey(t, m, "c")

	if m.filters.Path != "" {
		t.Errorf("Path should be cleared, got %q", m.filters.Path)
	}
	if m.filters.NameRegex != "" {
		t.Errorf("NameRegex should be cleared, got %q", m.filters.NameRegex)
	}
	if m.filters.FailedOnly {
		t.Error("FailedOnly should be cleared")
	}
}

func TestKeyRouting_EnterPromptPath(t *testing.T) {
	m := newTestModel(t)
	if m.mode != promptNone {
		t.Fatal("expected initial mode promptNone")
	}
	m = sendKey(t, m, "p")
	if m.mode != promptPath {
		t.Fatalf("expected promptPath after p, got %v", m.mode)
	}
}

func TestKeyRouting_EnterPromptName(t *testing.T) {
	m := newTestModel(t)
	m = sendKey(t, m, "t")
	if m.mode != promptName {
		t.Fatalf("expected promptName after t, got %v", m.mode)
	}
}

func TestKeyRouting_EscCancelsPrompt(t *testing.T) {
	m := newTestModel(t)
	m = sendKey(t, m, "p")
	if m.mode != promptPath {
		t.Fatal("should be in promptPath mode")
	}
	m = sendKeySpecial(t, m, tea.KeyEsc)
	if m.mode != promptNone {
		t.Fatalf("expected promptNone after Esc, got %v", m.mode)
	}
}

// ── A5: filter persistence ───────────────────────────────────────────────────

func TestFilterPersistence_Path(t *testing.T) {
	m := newTestModel(t)
	m = sendKey(t, m, "p")

	for _, r := range "internal/foo" {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(*watchModel)
	}
	m = sendKeySpecial(t, m, tea.KeyEnter)

	if m.filters.Path != "internal/foo" {
		t.Errorf("expected Path=%q, got %q", "internal/foo", m.filters.Path)
	}
	if m.mode != promptNone {
		t.Errorf("expected promptNone after enter, got %v", m.mode)
	}
}

func TestFilterPersistence_NameRegex(t *testing.T) {
	m := newTestModel(t)
	m = sendKey(t, m, "t")

	for _, r := range "TestBar" {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(*watchModel)
	}
	m = sendKeySpecial(t, m, tea.KeyEnter)

	if m.filters.NameRegex != "TestBar" {
		t.Errorf("expected NameRegex=%q, got %q", "TestBar", m.filters.NameRegex)
	}
}

func TestFilterPersistence_InvalidRegex(t *testing.T) {
	m := newTestModel(t)
	m.filters.NameRegex = "TestExisting"

	m = sendKey(t, m, "t")
	for _, r := range "[invalid" {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(*watchModel)
	}
	m = sendKeySpecial(t, m, tea.KeyEnter)

	if m.filters.NameRegex != "TestExisting" {
		t.Errorf("invalid regex should not update filter; got %q", m.filters.NameRegex)
	}
	if m.mode != promptNone {
		t.Errorf("expected promptNone after rejected regex, got %v", m.mode)
	}
}

func TestFilterPersistence_SurvivesRerun(t *testing.T) {
	m := newTestModel(t)
	m.filters.Path = "cmd/"
	m.filters.NameRegex = "TestFoo"
	m.filters.LogLevel = WatchLogVerbose

	id := m.runID + 1
	m.runID = id
	m.running = true

	next, _ := m.Update(runEndMsg{id: id, exitCode: 0})
	m = next.(*watchModel)

	if m.filters.Path != "cmd/" {
		t.Errorf("Path should persist across run; got %q", m.filters.Path)
	}
	if m.filters.NameRegex != "TestFoo" {
		t.Errorf("NameRegex should persist; got %q", m.filters.NameRegex)
	}
	if m.filters.LogLevel != WatchLogVerbose {
		t.Errorf("LogLevel should persist; got %v", m.filters.LogLevel)
	}
}

// ── A5: re-run on change ─────────────────────────────────────────────────────

func TestRerunOnChange(t *testing.T) {
	m := newTestModel(t)
	prevID := m.runID

	next, cmd := m.Update(fileChangeMsg{})
	m = next.(*watchModel)

	if m.runID <= prevID {
		t.Errorf("runID should increment on fileChangeMsg; before=%d after=%d", prevID, m.runID)
	}
	if !m.running {
		t.Error("running should be true after fileChangeMsg")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from fileChangeMsg")
	}
}

func TestPollDebounce(t *testing.T) {
	m := newTestModel(t)
	// Set debounceEnd to the past so it fires on next poll tick.
	m.debounceEnd = time.Now().Add(-1 * time.Millisecond)

	next, cmd := m.Update(pollTickMsg{})
	m = next.(*watchModel)

	if cmd == nil {
		t.Error("expected a cmd after debounce fires")
	}
	if !m.debounceEnd.IsZero() {
		t.Error("debounceEnd should be zeroed after firing")
	}
}

func TestFailedOnlyFallback(t *testing.T) {
	m := newTestModel(t)
	// No failedPkgs — should fall back to running all.
	prevID := m.runID

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m = next.(*watchModel)

	if m.runID <= prevID {
		t.Error("runID should increment even on fallback run")
	}
	if cmd == nil {
		t.Error("expected cmd from f key")
	}
}

// ── filter helpers ───────────────────────────────────────────────────────────

func TestBuildPackages_PathFilter(t *testing.T) {
	m := newWatchModel([]string{"./internal/foo/...", "./internal/bar/...", "./cmd/..."})
	m.filters.Path = "internal"

	pkgs := m.buildPackages()
	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages matching 'internal', got %v", pkgs)
	}
}

func TestBuildPackages_NoMatch(t *testing.T) {
	m := newWatchModel([]string{"./internal/foo/...", "./cmd/..."})
	m.filters.Path = "zzznomatch"

	pkgs := m.buildPackages()
	if len(pkgs) != 2 {
		t.Errorf("expected all 2 packages on no-match, got %v", pkgs)
	}
}

// ── classifyLine ─────────────────────────────────────────────────────────────

func TestClassifyLine(t *testing.T) {
	cases := []struct {
		line   string
		failed bool
		want   blit.LogLevel
	}{
		{"FAIL\tgithub.com/x/y [build failed]", true, blit.LogError},
		{"ok  \tgithub.com/x/y", false, blit.LogInfo},
		{"--- FAIL: TestFoo", true, blit.LogWarn},
		{"--- PASS: TestBar", false, blit.LogInfo},
		{"some random line", false, blit.LogDebug},
	}
	for _, tc := range cases {
		got := classifyLine(tc.line, tc.failed)
		if got != tc.want {
			t.Errorf("classifyLine(%q, %v) = %v, want %v", tc.line, tc.failed, got, tc.want)
		}
	}
}

// ── snapshotTree ─────────────────────────────────────────────────────────────

func TestSnapshotTree(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "x.go")
	if err := os.WriteFile(goFile, []byte("package x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	h := snapshotTree(dir)
	if h == "" {
		t.Error("snapshotTree returned empty string for dir with .go file")
	}
}
