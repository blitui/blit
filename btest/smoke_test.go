package btest_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/blitui/blit/btest"
)

// smokeStubModel is a minimal tea.Model for smoke testing.
type smokeStubModel struct {
	width, height int
}

func (m *smokeStubModel) Init() tea.Cmd { return nil }
func (m *smokeStubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}
func (m *smokeStubModel) View() string { return "hello smoke test" }

func TestSmoke_AllPass(t *testing.T) {
	report := btest.Smoke(t, &smokeStubModel{}, nil)
	if report.Failed > 0 {
		for _, r := range report.Results {
			if !r.Passed {
				t.Errorf("unexpected failure: %s — %s", r.Name, r.Detail)
			}
		}
	}
	if report.Passed == 0 {
		t.Error("expected at least one passing check")
	}
	// Default: init+render, 3 resize, key-dispatch, mouse = 6 checks
	if len(report.Results) != 6 {
		t.Errorf("expected 6 results, got %d", len(report.Results))
	}
}

func TestSmoke_CustomOpts(t *testing.T) {
	opts := &btest.SmokeOpts{
		Sizes: []btest.SmokeSize{{60, 20}},
		Keys:  []string{"a", "b"},
	}
	report := btest.Smoke(t, &smokeStubModel{}, opts)
	// init+render, 1 resize, key-dispatch, mouse = 4 checks
	if len(report.Results) != 4 {
		t.Errorf("expected 4 results, got %d", len(report.Results))
	}
	if report.Failed > 0 {
		t.Errorf("expected all pass, got %d failures", report.Failed)
	}
}

func TestSmokeTest_Passes(t *testing.T) {
	// SmokeTest should not fail for a well-behaved model.
	btest.SmokeTest(t, &smokeStubModel{}, nil)
}

// panicModel panics in View to test smoke failure detection.
type panicModel struct {
	panicked bool
}

func (m *panicModel) Init() tea.Cmd { return nil }
func (m *panicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
func (m *panicModel) View() string {
	if m.panicked {
		panic("intentional panic")
	}
	return "ok"
}

// emptyModel returns empty view to test failure detection.
type emptyModel struct{}

func (m *emptyModel) Init() tea.Cmd                            { return nil }
func (m *emptyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *emptyModel) View() string                             { return "" }

func TestSmoke_EmptyViewFails(t *testing.T) {
	report := btest.Smoke(t, &emptyModel{}, nil)
	// init+render should fail
	found := false
	for _, r := range report.Results {
		if r.Name == "init+render" && !r.Passed {
			found = true
		}
	}
	if !found {
		t.Error("expected init+render to fail for empty view")
	}
}
