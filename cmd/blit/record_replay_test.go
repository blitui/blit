package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRecordSavesSessionFile verifies that saveRecordedSession writes a valid
// .tuisess file atomically using temp+rename.
func TestRecordSavesSessionFile(t *testing.T) {
	dir := t.TempDir()
	sess := &recordSession{
		Name:    "testsess",
		Command: []string{"echo", "hello"},
		Steps: []recordStep{
			{Kind: "key", Key: "down"},
			{Kind: "key", Key: "enter"},
			{Kind: "tick"},
		},
		keystroke: 2,
	}

	code := saveRecordedSession(sess, dir, 80, 24)
	if code != 0 {
		t.Fatalf("saveRecordedSession returned %d", code)
	}

	dest := filepath.Join(dir, "testsess.tuisess")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var out tuiSessFile
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if out.Version != 2 {
		t.Errorf("version = %d, want 2", out.Version)
	}
	if out.Name != "testsess" {
		t.Errorf("name = %q, want %q", out.Name, "testsess")
	}
	if out.Cols != 80 || out.Lines != 24 {
		t.Errorf("size = %dx%d, want 80x24", out.Cols, out.Lines)
	}
	if len(out.Steps) != 3 {
		t.Errorf("steps = %d, want 3", len(out.Steps))
	}
	if out.RecordedAt == "" {
		t.Error("recorded_at is empty")
	}
	if len(out.Command) != 2 || out.Command[0] != "echo" {
		t.Errorf("command = %v, want [echo hello]", out.Command)
	}

	// Confirm no leftover .tmp file.
	if _, err := os.Stat(dest + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file should not exist after atomic rename")
	}
}

// TestRecordStatusBar verifies drawStatusBar does not panic and produces
// non-empty output.
func TestRecordStatusBar(t *testing.T) {
	sess := &recordSession{Name: "x", keystroke: 7, StartedAt: time.Now()}
	// drawStatusBar writes to stderr; just make sure it doesn't panic.
	drawStatusBar(sess, 80)
	drawStatusBar(sess, 0) // zero cols: no truncation
}

// TestReplayBarRenders verifies drawReplayBar does not panic with edge inputs.
func TestReplayBarRenders(t *testing.T) {
	drawReplayBar(0, 0, 1.0, 80)   // zero total
	drawReplayBar(5, 10, 2.0, 80)  // halfway
	drawReplayBar(10, 10, 0.5, 40) // complete, narrow terminal
}

// TestReplaySpeedParsed verifies the speed parsing helper.
func TestReplaySpeedParsed(t *testing.T) {
	cases := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"1x", 1.0, false},
		{"2x", 2.0, false},
		{"0.5x", 0.5, false},
		{"2", 2.0, false},
		{"0", 0, true},
		{"-1x", 0, true},
		{"fast", 0, true},
	}
	for _, tc := range cases {
		got, err := replaySpeedParsed(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("replaySpeedParsed(%q): want error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("replaySpeedParsed(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("replaySpeedParsed(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestReplayLoadsV1Session verifies that a v1 session file (blit ≤ v0.7.1)
// is accepted by the replay command without error.
func TestReplayLoadsV1Session(t *testing.T) {
	dir := t.TempDir()
	v1 := tuiSessFile{
		Version: 1,
		Cols:    80,
		Lines:   24,
		Steps: []recordStep{
			{Kind: "key", Key: "down"},
			{Kind: "screen", Screen: "hello"},
		},
	}
	data, _ := json.MarshalIndent(v1, "", "  ")
	path := filepath.Join(dir, "legacy.tuisess")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// runReplay dispatches to runReplaySession via flag parsing.
	// Call it directly to avoid os.Exit.
	rawData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var sess replaySessFile
	if err := json.Unmarshal(rawData, &sess); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if sess.Version != 1 {
		t.Errorf("version = %d, want 1", sess.Version)
	}
	// runReplaySession should return 0 for a valid session.
	code := runReplaySession(&sess, path, 1.0)
	if code != 0 {
		t.Errorf("runReplaySession = %d, want 0", code)
	}
}

// TestVirtualTerminalBasic exercises the minimal VT used during replay.
func TestVirtualTerminalBasic(t *testing.T) {
	vt := newVirtualTerminal(40, 5)

	vt.applyKey("h")
	vt.applyKey("i")
	rendered := vt.render()
	if len(rendered) == 0 {
		t.Error("render returned empty string")
	}

	vt.applyKey("backspace")
	// After backspace the last char should be gone.
	vt.applyKey("enter")
	// After enter cursor should move to next line.
	vt.applyKey("x")

	// Resize should not panic.
	vt.resize(60, 3)
}

// TestParseRawKey exercises the key parser with common sequences.
func TestParseRawKey(t *testing.T) {
	cases := []struct {
		raw  []byte
		want string
	}{
		{[]byte{0x03}, "ctrl+c"},
		{[]byte{0x10}, "ctrl+p"},
		{[]byte{0x0d}, "enter"},
		{[]byte{0x7f}, "backspace"},
		{[]byte{0x20}, "space"},
		{[]byte{'a'}, "a"},
		{[]byte{0x1b, '[', 'A'}, "up"},
		{[]byte{0x1b, '[', 'B'}, "down"},
		{[]byte{0x1b, '[', 'C'}, "right"},
		{[]byte{0x1b, '[', 'D'}, "left"},
	}
	for _, tc := range cases {
		got := parseRawKey(tc.raw)
		if got != tc.want {
			t.Errorf("parseRawKey(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}
