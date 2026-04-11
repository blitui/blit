package btest_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/blitui/blit/btest"
	tea "github.com/charmbracelet/bubbletea"
)

// sessModel is a tiny deterministic model that just echoes the keys it saw.
type sessModel struct {
	keys []string
}

func (m *sessModel) Init() tea.Cmd { return nil }
func (m *sessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		m.keys = append(m.keys, k.String())
	}
	return m, nil
}
func (m *sessModel) View() string {
	return fmt.Sprintf("keys=%v", m.keys)
}

func TestSessionRoundTrip(t *testing.T) {
	tm := btest.NewTestModel(t, &sessModel{}, 40, 5)
	rec := btest.NewSessionRecorder(tm).
		Key("down").
		Key("enter").
		Type("ab")

	path := filepath.Join(t.TempDir(), "sample.tuisess")
	if err := rec.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	sess, err := btest.LoadSession(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if sess.Version != btest.SessionFormatVersion {
		t.Errorf("version = %d", sess.Version)
	}
	if len(sess.Steps) == 0 {
		t.Error("no steps")
	}
}

func TestSessionReplayMatches(t *testing.T) {
	tm := btest.NewTestModel(t, &sessModel{}, 40, 5)
	rec := btest.NewSessionRecorder(tm).
		Key("a").
		Key("b").
		Key("c")

	path := filepath.Join(t.TempDir(), "flow.tuisess")
	if err := rec.Save(path); err != nil {
		t.Fatal(err)
	}

	// Replay against a fresh model — should match step-for-step.
	btest.Replay(t, &sessModel{}, path)
}

func TestSessionReplayDetectsDivergence(t *testing.T) {
	tm := btest.NewTestModel(t, &sessModel{}, 40, 5)
	rec := btest.NewSessionRecorder(tm).Key("a")
	path := filepath.Join(t.TempDir(), "div.tuisess")
	if err := rec.Save(path); err != nil {
		t.Fatal(err)
	}

	// Replay against a model that ignores keys → screen mismatch.
	fake := &fakeT{TB: t}
	btest.Replay(fake, &ignoringModel{}, path)
	if !fake.failed {
		t.Error("expected divergence to fail the replay")
	}
}

// ignoringModel never records anything so its View stays constant.
type ignoringModel struct{}

func (m *ignoringModel) Init() tea.Cmd                           { return nil }
func (m *ignoringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *ignoringModel) View() string                            { return "nothing" }

// fakeT captures Errorf/Fatalf calls without actually failing the parent test.
type fakeT struct {
	testing.TB
	failed bool
}

func (f *fakeT) Errorf(format string, args ...interface{}) { f.failed = true }
func (f *fakeT) Fatalf(format string, args ...interface{}) { f.failed = true }
func (f *fakeT) Helper()                                   {}
