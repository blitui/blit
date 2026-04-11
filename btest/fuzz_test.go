package btest_test

import (
	"fmt"
	"testing"

	"github.com/blitui/blit/btest"
	tea "github.com/charmbracelet/bubbletea"
)

// stableModel is a model that never panics — used to verify the fuzzer
// completes all iterations without false positives.
type stableModel struct {
	keys   int
	width  int
	height int
}

func (m *stableModel) Init() tea.Cmd { return nil }
func (m *stableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.keys++
	}
	if v, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = v.Width
		m.height = v.Height
	}
	return m, nil
}
func (m *stableModel) View() string {
	return fmt.Sprintf("keys=%d size=%dx%d", m.keys, m.width, m.height)
}

// panicOnQModel panics when it receives the "q" key.
type panicOnQModel struct{}

func (m *panicOnQModel) Init() tea.Cmd { return nil }
func (m *panicOnQModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.String() == "q" {
			panic("received q key")
		}
	}
	return m, nil
}
func (m *panicOnQModel) View() string { return "ok" }

func TestFuzz_StableModel(t *testing.T) {
	result := btest.Fuzz(t, &stableModel{}, btest.FuzzConfig{
		Seed:       42,
		Iterations: 500,
	})
	if result.Panicked {
		t.Fatalf("stable model should not panic, panicked at event %d: %v",
			result.PanicEvent, result.PanicValue)
	}
	if result.Iterations != 500 {
		t.Errorf("iterations = %d, want 500", result.Iterations)
	}
}

func TestFuzz_DetectsPanic(t *testing.T) {
	// The panicOnQModel panics on "q". With enough iterations and the
	// right seed, the fuzzer should eventually generate a "q" key.
	result := btest.Fuzz(t, &panicOnQModel{}, btest.FuzzConfig{
		Seed:         42,
		Iterations:   5000,
		KeyWeight:    10,
		MouseWeight:  1,
		ResizeWeight: 1,
		TickWeight:   1,
	})
	if !result.Panicked {
		t.Fatal("expected panic from panicOnQModel but none occurred")
	}
	if result.PanicValue == nil {
		t.Error("PanicValue should not be nil")
	}
}

func TestFuzz_FailingSequence(t *testing.T) {
	result := btest.Fuzz(t, &panicOnQModel{}, btest.FuzzConfig{
		Seed:         42,
		Iterations:   5000,
		KeyWeight:    10,
		MouseWeight:  1,
		ResizeWeight: 1,
		TickWeight:   1,
	})
	if !result.Panicked {
		t.Fatal("expected panic")
	}
	seq := result.FailingSequence()
	if len(seq) == 0 {
		t.Fatal("FailingSequence should not be empty on panic")
	}
	// Last event in the failing sequence should be the trigger.
	last := seq[len(seq)-1]
	if last.Kind != "key" || last.Key != "q" {
		t.Errorf("last event = %v, want key(q)", last)
	}
}

func TestFuzz_FailingSequenceNoPanic(t *testing.T) {
	result := &btest.FuzzResult{Panicked: false}
	if seq := result.FailingSequence(); seq != nil {
		t.Errorf("FailingSequence should be nil when no panic, got %v", seq)
	}
}

func TestFuzz_FormatSequence(t *testing.T) {
	events := []btest.FuzzEvent{
		{Kind: "key", Key: "a"},
		{Kind: "mouse", X: 10, Y: 5, Button: tea.MouseButtonLeft},
		{Kind: "resize", Cols: 120, Lines: 40},
		{Kind: "tick"},
	}
	got := btest.FormatSequence(events)
	if got == "" {
		t.Error("FormatSequence returned empty string")
	}
	// Should contain all event representations.
	for _, want := range []string{"key(a)", "mouse(10,5", "resize(120x40)", "tick"} {
		if !containsStr(got, want) {
			t.Errorf("FormatSequence missing %q in %q", want, got)
		}
	}
}

func TestFuzz_DefaultConfig(t *testing.T) {
	// Verify defaults are applied when zero values are passed.
	result := btest.Fuzz(t, &stableModel{}, btest.FuzzConfig{})
	if result.Iterations != 1000 {
		t.Errorf("default iterations = %d, want 1000", result.Iterations)
	}
}

func TestFuzz_CustomWeights(t *testing.T) {
	// Keys only — no mouse, resize, or tick events.
	result := btest.Fuzz(t, &stableModel{}, btest.FuzzConfig{
		Seed:         99,
		Iterations:   100,
		KeyWeight:    1,
		MouseWeight:  0,
		ResizeWeight: 0,
		TickWeight:   0,
	})
	if result.Panicked {
		t.Fatal("stable model should not panic")
	}
	// All events should be keys.
	for i, ev := range result.Events {
		if ev.Kind != "key" {
			t.Errorf("event %d: kind = %q, want key (weights: keys only)", i, ev.Kind)
			break
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
