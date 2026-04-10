package main

import (
	"strings"
	"testing"
)

func TestBuildTape_V1Session(t *testing.T) {
	sess := session{
		Version: 1,
		Cols:    80,
		Lines:   24,
		Steps: []step{
			{Kind: "key", Key: "enter"},
			{Kind: "type", Text: "hello"},
		},
	}
	tape := buildTape(sess, 80, 24, 15, 80, "test.tuisess")
	if !strings.Contains(tape, "Set Width 640") {
		t.Errorf("tape missing width: %s", tape)
	}
	if !strings.Contains(tape, "Set Height 384") {
		t.Errorf("tape missing height: %s", tape)
	}
	if !strings.Contains(tape, "Key Enter") {
		t.Error("tape missing Enter key")
	}
	if !strings.Contains(tape, `Type "hello"`) {
		t.Error("tape missing type command")
	}
	if !strings.Contains(tape, "Sleep 2s") {
		t.Error("tape missing final hold")
	}
}

func TestBuildTape_V2Session(t *testing.T) {
	sess := session{
		Version: 2,
		Cols:    120,
		Lines:   40,
		Steps: []step{
			{Kind: "key", Key: "tab"},
		},
	}
	tape := buildTape(sess, 120, 40, 30, 100, "demo.tuisess")
	if !strings.Contains(tape, "Set Framerate 30") {
		t.Error("tape missing custom framerate")
	}
	if !strings.Contains(tape, "Key Tab") {
		t.Error("tape missing Tab key")
	}
	if !strings.Contains(tape, "demo.tuisess") {
		t.Error("tape missing source filename in comment")
	}
}

func TestBuildTape_ResizeStep(t *testing.T) {
	sess := session{
		Version: 2,
		Cols:    80,
		Lines:   24,
		Steps: []step{
			{Kind: "resize", Cols: 100, Lines: 30},
		},
	}
	tape := buildTape(sess, 80, 24, 15, 80, "test.tuisess")
	if !strings.Contains(tape, "# Resize 100x30") {
		t.Error("tape missing resize comment")
	}
}

func TestBuildTape_ScreenStep(t *testing.T) {
	sess := session{
		Version: 2,
		Cols:    80,
		Lines:   24,
		Steps: []step{
			{Kind: "screen", Screen: "snapshot data"},
		},
	}
	tape := buildTape(sess, 80, 24, 15, 80, "test.tuisess")
	if !strings.Contains(tape, "Sleep 160ms") {
		t.Error("tape missing screen step sleep (delay*2)")
	}
}

func TestBuildTape_StdinSource(t *testing.T) {
	sess := session{Version: 2, Cols: 80, Lines: 24}
	tape := buildTape(sess, 80, 24, 15, 80, "-")
	if !strings.Contains(tape, "stdin") {
		t.Error("tape should reference stdin when input is -")
	}
}

func TestBuildTape_EscapedQuotes(t *testing.T) {
	sess := session{
		Version: 2,
		Cols:    80,
		Lines:   24,
		Steps: []step{
			{Kind: "type", Text: `say "hi"`},
		},
	}
	tape := buildTape(sess, 80, 24, 15, 80, "test.tuisess")
	if !strings.Contains(tape, `say \"hi\"`) {
		t.Errorf("tape should escape quotes: %s", tape)
	}
}

func TestBuildTape_WidthHeightOverride(t *testing.T) {
	sess := session{Version: 2, Cols: 80, Lines: 24}
	tape := buildTape(sess, 100, 50, 15, 80, "test.tuisess")
	if !strings.Contains(tape, "Set Width 800") {
		t.Error("tape should use overridden width")
	}
	if !strings.Contains(tape, "Set Height 800") {
		t.Error("tape should use overridden height")
	}
}

func TestVhsKey_KnownKeys(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"enter", "Key Enter\n"},
		{"tab", "Key Tab\n"},
		{"esc", "Key Escape\n"},
		{"up", "Key Up\n"},
		{"down", "Key Down\n"},
		{"left", "Key Left\n"},
		{"right", "Key Right\n"},
		{"backspace", "Key Backspace\n"},
		{"space", "Key Space\n"},
		{"ctrl+c", "Key Ctrl+C\n"},
		{"pgup", "Key PageUp\n"},
		{"pgdown", "Key PageDown\n"},
		{"f1", "Key F1\n"},
	}
	for _, tc := range cases {
		got := vhsKey(tc.key, 80)
		if !strings.Contains(got, tc.want) {
			t.Errorf("vhsKey(%q) = %q, want to contain %q", tc.key, got, tc.want)
		}
		if !strings.Contains(got, "Sleep 80ms") {
			t.Errorf("vhsKey(%q) missing delay", tc.key)
		}
	}
}

func TestVhsKey_UnknownKey(t *testing.T) {
	got := vhsKey("x", 50)
	if !strings.Contains(got, "Key X") {
		t.Errorf("unknown key should capitalize: %q", got)
	}
}

func TestVhsKey_EmptyKey(t *testing.T) {
	got := vhsKey("", 50)
	if !strings.Contains(got, "# unknown key") {
		t.Errorf("empty key should be comment: %q", got)
	}
}
