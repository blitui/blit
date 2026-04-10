package cli

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn and returns what it wrote to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return string(out)
}

func TestSuccess(t *testing.T) {
	got := captureStdout(t, func() { Success("ok") })
	if !strings.Contains(got, "ok") {
		t.Errorf("Success output missing message: %q", got)
	}
}

func TestWarning(t *testing.T) {
	got := captureStdout(t, func() { Warning("caution") })
	if !strings.Contains(got, "caution") {
		t.Errorf("Warning output missing message: %q", got)
	}
}

func TestError(t *testing.T) {
	got := captureStdout(t, func() { Error("fail") })
	if !strings.Contains(got, "fail") {
		t.Errorf("Error output missing message: %q", got)
	}
}

func TestInfo(t *testing.T) {
	got := captureStdout(t, func() { Info("note") })
	if !strings.Contains(got, "note") {
		t.Errorf("Info output missing message: %q", got)
	}
}

func TestSuccessf(t *testing.T) {
	got := captureStdout(t, func() { Successf("count: %d", 42) })
	if !strings.Contains(got, "count: 42") {
		t.Errorf("Successf output missing formatted message: %q", got)
	}
}

func TestWarningf(t *testing.T) {
	got := captureStdout(t, func() { Warningf("warn: %s", "hot") })
	if !strings.Contains(got, "warn: hot") {
		t.Errorf("Warningf output missing formatted message: %q", got)
	}
}

func TestErrorf(t *testing.T) {
	got := captureStdout(t, func() { Errorf("err: %v", "bad") })
	if !strings.Contains(got, "err: bad") {
		t.Errorf("Errorf output missing formatted message: %q", got)
	}
}

func TestInfof(t *testing.T) {
	got := captureStdout(t, func() { Infof("info: %d", 7) })
	if !strings.Contains(got, "info: 7") {
		t.Errorf("Infof output missing formatted message: %q", got)
	}
}

func TestStep(t *testing.T) {
	got := captureStdout(t, func() { Step(1, 3, "installing") })
	if !strings.Contains(got, "1/3") {
		t.Errorf("Step output missing step numbers: %q", got)
	}
	if !strings.Contains(got, "installing") {
		t.Errorf("Step output missing message: %q", got)
	}
}

func TestTitle(t *testing.T) {
	got := captureStdout(t, func() { Title("Setup") })
	if !strings.Contains(got, "Setup") {
		t.Errorf("Title output missing message: %q", got)
	}
}

func TestSection(t *testing.T) {
	got := captureStdout(t, func() { Section("Config") })
	if !strings.Contains(got, "Config") {
		t.Errorf("Section output missing message: %q", got)
	}
}

func TestSeparator(t *testing.T) {
	got := captureStdout(t, func() { Separator() })
	if !strings.Contains(got, "─") {
		t.Errorf("Separator output missing divider chars: %q", got)
	}
}

func TestDim(t *testing.T) {
	got := captureStdout(t, func() { Dim("muted") })
	if !strings.Contains(got, "muted") {
		t.Errorf("Dim output missing message: %q", got)
	}
}

func TestKeyValue(t *testing.T) {
	got := captureStdout(t, func() { KeyValue("name", "alice") })
	if !strings.Contains(got, "name") || !strings.Contains(got, "alice") {
		t.Errorf("KeyValue output missing key or value: %q", got)
	}
}
