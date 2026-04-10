package tape_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blitui/blit/internal/tape"
	"github.com/blitui/blit/btest"
)

// fixtureSession returns a small hand-crafted session that matches
// the golden file at testdata/basic.tape.golden.
func fixtureSession() *btest.Session {
	return &btest.Session{
		Version: 2,
		Cols:    80,
		Lines:   24,
		Steps: []btest.SessionStep{
			{Kind: "type", Text: "hello"},
			{Kind: "screen", Screen: "hello"},
			{Kind: "key", Key: "enter"},
			{Kind: "screen", Screen: "hello\n"},
			{Kind: "key", Key: "down"},
			{Kind: "screen", Screen: "hello\n"},
			{Kind: "type", Text: "world"},
			{Kind: "screen", Screen: "hello\nworld"},
		},
	}
}

func TestGenerate_GoldenFixture(t *testing.T) {
	sess := fixtureSession()
	got := tape.Generate(sess)

	goldenPath := filepath.Join("testdata", "basic.tape.golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated golden file %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}

	if got != string(want) {
		t.Errorf("tape output does not match golden\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestGenerate_TerminalSize(t *testing.T) {
	sess := &btest.Session{Version: 2, Cols: 100, Lines: 30, Steps: nil}
	out := tape.Generate(sess)

	if want := "Set Width 800\n"; !contains(out, want) {
		t.Errorf("expected %q in output:\n%s", want, out)
	}
	if want := "Set Height 480\n"; !contains(out, want) {
		t.Errorf("expected %q in output:\n%s", want, out)
	}
}

func TestGenerate_KeyDirectives(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"enter", "Enter\n"},
		{"up", "Up\n"},
		{"down", "Down\n"},
		{"left", "Left\n"},
		{"right", "Right\n"},
		{"tab", "Tab\n"},
		{"backspace", "Backspace\n"},
		{"esc", "Escape\n"},
		{"ctrl+c", "Ctrl+C\n"},
		{"ctrl+d", "Ctrl+D\n"},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			sess := &btest.Session{
				Version: 2, Cols: 80, Lines: 24,
				Steps: []btest.SessionStep{
					{Kind: "key", Key: tc.key},
				},
			}
			out := tape.Generate(sess)
			if !contains(out, tc.want) {
				t.Errorf("key %q: expected %q in:\n%s", tc.key, tc.want, out)
			}
		})
	}
}

func TestGenerate_TypeEscapesQuotes(t *testing.T) {
	sess := &btest.Session{
		Version: 2, Cols: 80, Lines: 24,
		Steps: []btest.SessionStep{
			{Kind: "type", Text: `say "hello"`},
		},
	}
	out := tape.Generate(sess)
	if !contains(out, `Type "say \"hello\""`) {
		t.Errorf("expected escaped quotes in output:\n%s", out)
	}
}

func TestGenerate_ScreenStepsOmitted(t *testing.T) {
	sess := &btest.Session{
		Version: 2, Cols: 80, Lines: 24,
		Steps: []btest.SessionStep{
			{Kind: "screen", Screen: "some screen content"},
		},
	}
	out := tape.Generate(sess)
	if contains(out, "some screen content") {
		t.Errorf("screen step content should not appear in tape output:\n%s", out)
	}
}

func TestGenerate_ResizeEmitsComment(t *testing.T) {
	sess := &btest.Session{
		Version: 2, Cols: 80, Lines: 24,
		Steps: []btest.SessionStep{
			{Kind: "resize", Cols: 120, Lines: 40},
		},
	}
	out := tape.Generate(sess)
	if !contains(out, "# Resize 120x40\n") {
		t.Errorf("expected resize comment in output:\n%s", out)
	}
}

func TestGenerate_TickEmitsSleep(t *testing.T) {
	sess := &btest.Session{
		Version: 2, Cols: 80, Lines: 24,
		Steps: []btest.SessionStep{
			{Kind: "tick"},
		},
	}
	out := tape.Generate(sess)
	if !contains(out, "Sleep 100ms\n") {
		t.Errorf("expected Sleep directive for tick in output:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
