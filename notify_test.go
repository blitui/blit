package tuikit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNotifyCmd_ZeroDuration(t *testing.T) {
	cmd := NotifyCmd("", 0)
	if cmd == nil {
		t.Fatal("nil")
	}
	msg, ok := cmd().(NotifyMsg)
	if !ok {
		t.Fatalf("%T", cmd())
	}
	if msg.Text != "" || msg.Duration != 0 {
		t.Error(msg)
	}
}

func TestNotifyCmd_Nil(t *testing.T) {
	cmd := NotifyCmd("hi", 0)
	a, b := cmd().(NotifyMsg), cmd().(NotifyMsg)
	if a != b {
		t.Error(a, b)
	}
}

func TestToastCmd_Info(t *testing.T) {
	cmd := ToastCmd(SeverityInfo, "title", "body", 3*time.Second)
	msg := cmd().(ToastMsg)
	if msg.Severity != SeverityInfo {
		t.Error(msg.Severity)
	}
	if msg.Title != "title" {
		t.Error(msg.Title)
	}
}

func TestToastCmd_WithActions(t *testing.T) {
	called := false
	cmd := ToastCmd(SeverityError, "e", "b", time.Second, ToastAction{"R", func() { called = true }})
	cmd().(ToastMsg).Actions[0].Handler()
	if !called {
		t.Error("not called")
	}
}

func TestToastManager_AddAndStack(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{MaxVisible: 3})
	tm.add(ToastMsg{Title: "one", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "two", Duration: 5 * time.Second})
	if len(tm.toasts) != 2 {
		t.Fatalf("%d", len(tm.toasts))
	}
}

func TestToastManager_MaxVisible(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{MaxVisible: 2})
	tm.add(ToastMsg{Title: "a", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "b", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "c", Duration: 5 * time.Second})
	if len(tm.toasts) != 2 {
		t.Fatalf("%d", len(tm.toasts))
	}
	if tm.toasts[0].Title != "b" {
		t.Error(tm.toasts[0].Title)
	}
}

func TestToastManager_DismissTop(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.add(ToastMsg{Title: "first", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "second", Duration: 5 * time.Second})
	tm.dismissTop()
	if len(tm.toasts) != 1 {
		t.Fatalf("%d", len(tm.toasts))
	}
	if tm.toasts[0].Title != "first" {
		t.Error(tm.toasts[0].Title)
	}
}

func TestToastManager_DismissAt(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.add(ToastMsg{Title: "a", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "b", Duration: 5 * time.Second})
	tm.add(ToastMsg{Title: "c", Duration: 5 * time.Second})
	tm.dismissAt(1)
	if len(tm.toasts) != 2 {
		t.Fatalf("%d", len(tm.toasts))
	}
	if tm.toasts[0].Title != "a" || tm.toasts[1].Title != "c" {
		t.Error()
	}
}

func TestToastManager_AutoExpire(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.noAnim = true
	tm.add(ToastMsg{Title: "x", Duration: 10 * time.Millisecond})
	time.Sleep(20 * time.Millisecond)
	if !tm.tick(time.Now()) {
		t.Error("changed should be true")
	}
	if tm.hasActive() {
		t.Error("should be gone")
	}
}

func TestToastManager_DefaultDuration(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.noAnim = true
	tm.add(ToastMsg{Title: "x", Duration: 0})
	if tm.toasts[0].Duration != 4*time.Second {
		t.Error(tm.toasts[0].Duration)
	}
}

func TestToastManager_SeverityColors(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.theme = DefaultTheme()
	if tm.severityColor(SeveritySuccess) != tm.theme.Positive {
		t.Error()
	}
	if tm.severityColor(SeverityWarn) != tm.theme.Flash {
		t.Error()
	}
	if tm.severityColor(SeverityError) != tm.theme.Negative {
		t.Error()
	}
	if tm.severityColor(SeverityInfo) != tm.theme.Accent {
		t.Error()
	}
}

func TestToastManager_View(t *testing.T) {
	tm := newToastManager(ToastManagerOpts{})
	tm.noAnim = true
	tm.theme = DefaultTheme()
	if tm.view(80) != "" {
		t.Error()
	}
	tm.add(ToastMsg{Title: "OK", Body: "done", Duration: 5 * time.Second})
	if tm.view(80) == "" {
		t.Error()
	}
}

func TestSeverityIcon(t *testing.T) {
	if severityIcon(SeverityInfo) != "i" {
		t.Error()
	}
	if severityIcon(SeveritySuccess) != "✓" {
		t.Error()
	}
	if severityIcon(SeverityWarn) != "⚠" {
		t.Error()
	}
	if severityIcon(SeverityError) != "✗" {
		t.Error()
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error()
	}
	got := truncate("hello world", 5)
	if []rune(got)[len([]rune(got))-1] != '…' {
		t.Errorf("last rune should be ellipsis, got %q", got)
	}
	if len([]rune(got)) != 5 {
		t.Errorf("truncated length should be 5 runes, got %d: %q", len([]rune(got)), got)
	}
}

func TestWrapText(t *testing.T) {
	lines := wrapText("one two three four", 9)
	if len(lines) < 2 {
		t.Errorf("%v", lines)
	}
}

func TestAppToastMsg(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.Update(ToastMsg{Severity: SeverityWarn, Title: "w", Duration: 5 * time.Second})
	if !a.toasts.hasActive() {
		t.Error()
	}
	if a.toasts.toasts[0].Severity != SeverityWarn {
		t.Error()
	}
}

func TestAppToastEscDismiss(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.Update(ToastMsg{Severity: SeverityInfo, Title: "hi", Duration: 5 * time.Second})
	if !a.toasts.hasActive() {
		t.Fatal()
	}
	a.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if a.toasts.hasActive() {
		t.Error("should be dismissed")
	}
}

func TestAppNotifyRoutesToToast(t *testing.T) {
	a := newAppModel(WithTheme(DefaultTheme()))
	a.Update(NotifyMsg{Text: "compat", Duration: 3 * time.Second})
	if !a.toasts.hasActive() {
		t.Error()
	}
	if a.toasts.toasts[0].Severity != SeverityInfo {
		t.Error()
	}
}
