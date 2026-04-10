// Package btest session record/replay.
//
// A .tuisess file captures a deterministic sequence of input events against
// a tea.Model and the resulting screen after each step. Consumer apps can
// commit these files under `testdata/sessions/` as end-to-end regression
// fixtures; running Replay against the same model should yield identical
// screens at every step.
//
// Format (JSON):
//
//	{
//	    "version": 1,
//	    "cols": 80,
//	    "lines": 24,
//	    "steps": [
//	        {"kind": "key",    "key":   "down"},
//	        {"kind": "screen", "screen": "..."},
//	        {"kind": "type",   "text":  "abc"},
//	        {"kind": "screen", "screen": "..."}
//	    ]
//	}
//
// The "screen" steps are the expected screen strings captured right after
// the previous input step. Replay compares the live screen to the expected
// screen and fails the test on any divergence.
package btest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// SessionFormatVersion is the on-disk version marker. Bump whenever the
// step schema changes in a backwards-incompatible way.
//
// History:
//
//	v1 – initial format (blit ≤ v0.7.1)
//	v2 – adds Name, RecordedAt, Command metadata fields (blit v0.11+)
const SessionFormatVersion = 2

// Session is the on-disk representation of a .tuisess file.
type Session struct {
	Version    int           `json:"version"`
	Cols       int           `json:"cols"`
	Lines      int           `json:"lines"`
	Name       string        `json:"name,omitempty"`        // v2+: logical session name
	RecordedAt string        `json:"recorded_at,omitempty"` // v2+: RFC3339 timestamp
	Command    []string      `json:"command,omitempty"`     // v2+: recorded command args
	Steps      []SessionStep `json:"steps"`
}

// SessionStep is a single input or screen assertion.
type SessionStep struct {
	Kind   string `json:"kind"`             // "key", "type", "resize", "screen"
	Key    string `json:"key,omitempty"`    // for kind=key
	Text   string `json:"text,omitempty"`   // for kind=type
	Cols   int    `json:"cols,omitempty"`   // for kind=resize
	Lines  int    `json:"lines,omitempty"`  // for kind=resize
	Screen string `json:"screen,omitempty"` // for kind=screen (expected post-state)
}

// SessionRecorder captures input + screen steps against a live TestModel.
type SessionRecorder struct {
	tm    *TestModel
	steps []SessionStep
}

// NewSessionRecorder returns a recorder bound to an existing TestModel.
func NewSessionRecorder(tm *TestModel) *SessionRecorder {
	return &SessionRecorder{tm: tm}
}

// Key sends a named key and records the key + resulting screen.
func (r *SessionRecorder) Key(key string) *SessionRecorder {
	r.tm.SendKey(key)
	r.steps = append(r.steps,
		SessionStep{Kind: "key", Key: key},
		SessionStep{Kind: "screen", Screen: r.tm.Screen().String()},
	)
	return r
}

// Type sends text one char at a time and records the aggregated step.
func (r *SessionRecorder) Type(text string) *SessionRecorder {
	r.tm.Type(text)
	r.steps = append(r.steps,
		SessionStep{Kind: "type", Text: text},
		SessionStep{Kind: "screen", Screen: r.tm.Screen().String()},
	)
	return r
}

// Resize updates the simulated terminal size and records the step.
func (r *SessionRecorder) Resize(cols, lines int) *SessionRecorder {
	r.tm.SendResize(cols, lines)
	r.steps = append(r.steps,
		SessionStep{Kind: "resize", Cols: cols, Lines: lines},
		SessionStep{Kind: "screen", Screen: r.tm.Screen().String()},
	)
	return r
}

// Save writes the recorded session to the given path as a .tuisess file.
// Missing parent directories are created. If path has no extension it is
// given ".tuisess".
func (r *SessionRecorder) Save(path string) error {
	if filepath.Ext(path) == "" {
		path += ".tuisess"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	sess := Session{
		Version: SessionFormatVersion,
		Cols:    r.tm.Cols(),
		Lines:   r.tm.Lines(),
		Steps:   r.steps,
	}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSession reads a .tuisess file. Versions 1 and 2 are both accepted;
// v1 files are loaded without the metadata fields introduced in v2.
func LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if s.Version != 1 && s.Version != SessionFormatVersion {
		return nil, fmt.Errorf("unsupported session version %d (want 1 or %d)", s.Version, SessionFormatVersion)
	}
	return &s, nil
}

// Replay drives the given model through a recorded session, asserting
// that each screen step matches the live screen. It fails the test on any
// divergence. Cols/lines from the session override the model's initial size.
func Replay(t testing.TB, model tea.Model, path string) {
	t.Helper()
	sess, err := LoadSession(path)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	tm := NewTestModel(t, model, sess.Cols, sess.Lines)
	stepIdx := 0
	for _, step := range sess.Steps {
		stepIdx++
		switch step.Kind {
		case "key":
			tm.SendKey(step.Key)
		case "type":
			tm.Type(step.Text)
		case "resize":
			tm.SendResize(step.Cols, step.Lines)
		case "screen":
			got := tm.Screen().String()
			if !sessionScreensEqual(got, step.Screen) {
				t.Errorf("Replay %s step %d: screen mismatch\n--- expected ---\n%s\n--- got ---\n%s",
					filepath.Base(path), stepIdx, step.Screen, got)
				return
			}
		default:
			t.Errorf("Replay: unknown step kind %q at step %d", step.Kind, stepIdx)
			return
		}
	}
}

// sessionScreensEqual compares two screen dumps with CRLF/trailing-space
// normalization so cross-platform replays stay stable.
func sessionScreensEqual(a, b string) bool {
	return normalizeSessionScreen(a) == normalizeSessionScreen(b)
}

func normalizeSessionScreen(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	// Drop trailing blank lines.
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
