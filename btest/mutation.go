// Package btest mutation testing for TUI models.
//
// Mutation testing verifies the quality of a test suite by injecting small
// behavioral changes ("mutations") into a tea.Model and checking whether the
// tests catch them. A mutation that goes undetected indicates a gap in the
// test suite.
//
// Usage:
//
//	result := btest.MutationTest(t, func() tea.Model { return NewMyModel() }, btest.MutationConfig{
//	    Test: func(t testing.TB, model tea.Model) {
//	        tm := btest.NewTestModel(t, model, 80, 24)
//	        tm.SendKey("enter")
//	        btest.AssertContains(t, tm.Screen(), "expected")
//	    },
//	})
//	// result.Killed / result.Survived / result.Total
package btest

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// MutationType identifies the kind of mutation applied.
type MutationType int

const (
	// MutationDropKey silently drops a key message.
	MutationDropKey MutationType = iota
	// MutationSwapKeys swaps a common key pair (e.g., up↔down).
	MutationSwapKeys
	// MutationEmptyView forces View() to return an empty string.
	MutationEmptyView
	// MutationStaticView forces View() to return a fixed string.
	MutationStaticView
	// MutationDropCmd discards commands returned from Update.
	MutationDropCmd
	// MutationNilInit makes Init() return nil.
	MutationNilInit
)

// String returns a human-readable label for the mutation type.
func (mt MutationType) String() string {
	switch mt {
	case MutationDropKey:
		return "drop-key"
	case MutationSwapKeys:
		return "swap-keys"
	case MutationEmptyView:
		return "empty-view"
	case MutationStaticView:
		return "static-view"
	case MutationDropCmd:
		return "drop-cmd"
	case MutationNilInit:
		return "nil-init"
	default:
		return fmt.Sprintf("unknown(%d)", int(mt))
	}
}

// Mutation describes a single behavioral change applied to a model.
type Mutation struct {
	// Type identifies the mutation kind.
	Type MutationType
	// Description is a human-readable explanation.
	Description string
}

// MutationResult is the outcome of applying a single mutation.
type MutationResult struct {
	Mutation Mutation
	// Killed is true if the test suite detected the mutation (test failed).
	Killed bool
}

// MutationReport summarizes the results of a full mutation test run.
type MutationReport struct {
	Results  []MutationResult
	Killed   int
	Survived int
	Total    int
}

// Score returns the mutation score as a percentage (killed / total * 100).
func (r *MutationReport) Score() float64 {
	if r.Total == 0 {
		return 100
	}
	return float64(r.Killed) / float64(r.Total) * 100
}

// Summary returns a human-readable summary of the mutation test.
func (r *MutationReport) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Mutation testing: %d/%d killed (%.1f%%)\n", r.Killed, r.Total, r.Score())
	for _, res := range r.Results {
		status := "KILLED"
		if !res.Killed {
			status = "SURVIVED"
		}
		fmt.Fprintf(&sb, "  [%s] %s: %s\n", status, res.Mutation.Type, res.Mutation.Description)
	}
	return sb.String()
}

// Survivors returns all mutations that were not caught by the test suite.
func (r *MutationReport) Survivors() []MutationResult {
	var out []MutationResult
	for _, res := range r.Results {
		if !res.Killed {
			out = append(out, res)
		}
	}
	return out
}

// MutationConfig configures a mutation test run.
type MutationConfig struct {
	// Test is the test function to run against each mutated model.
	// It should exercise the model and make assertions using t.
	Test func(t testing.TB, model tea.Model)

	// Mutations limits which mutation types to apply. If nil, all built-in
	// mutations are used.
	Mutations []MutationType

	// Cols and Lines set the terminal size for built-in mutations that
	// need to construct a TestModel internally. Defaults to 80x24.
	Cols  int
	Lines int
}

func (c *MutationConfig) defaults() {
	if c.Cols == 0 {
		c.Cols = 80
	}
	if c.Lines == 0 {
		c.Lines = 24
	}
}

// MutationTest runs the test function against each mutation of the model
// and returns a report showing which mutations were killed (caught).
//
// The factory function is called fresh for each mutation so mutations
// don't leak state between runs.
func MutationTest(t testing.TB, factory func() tea.Model, cfg MutationConfig) *MutationReport {
	t.Helper()
	cfg.defaults()

	mutations := cfg.Mutations
	if len(mutations) == 0 {
		mutations = []MutationType{
			MutationDropKey,
			MutationSwapKeys,
			MutationEmptyView,
			MutationStaticView,
			MutationDropCmd,
			MutationNilInit,
		}
	}

	var report MutationReport
	for _, mt := range mutations {
		muts := buildMutations(mt)
		for _, mut := range muts {
			model := factory()
			mutated := applyMutation(model, mut)
			killed := runMutant(cfg.Test, mutated)
			report.Results = append(report.Results, MutationResult{
				Mutation: mut,
				Killed:   killed,
			})
			if killed {
				report.Killed++
			} else {
				report.Survived++
			}
			report.Total++
		}
	}

	return &report
}

// AssertMutationScore fails the test if the mutation score is below min.
func AssertMutationScore(t testing.TB, report *MutationReport, min float64) {
	t.Helper()
	if report.Score() < min {
		t.Errorf("mutation score %.1f%% below minimum %.1f%%\n%s",
			report.Score(), min, report.Summary())
	}
}

// buildMutations returns all concrete mutation instances for a type.
func buildMutations(mt MutationType) []Mutation {
	switch mt {
	case MutationDropKey:
		return []Mutation{
			{Type: mt, Description: "drop all key messages in Update"},
		}
	case MutationSwapKeys:
		return []Mutation{
			{Type: mt, Description: "swap up/down arrow keys"},
			{Type: mt, Description: "swap tab/shift+tab keys"},
		}
	case MutationEmptyView:
		return []Mutation{
			{Type: mt, Description: "View() always returns empty string"},
		}
	case MutationStaticView:
		return []Mutation{
			{Type: mt, Description: "View() always returns 'MUTANT'"},
		}
	case MutationDropCmd:
		return []Mutation{
			{Type: mt, Description: "discard all commands from Update"},
		}
	case MutationNilInit:
		return []Mutation{
			{Type: mt, Description: "Init() returns nil instead of original command"},
		}
	default:
		return nil
	}
}

// applyMutation wraps the model with the given mutation.
func applyMutation(model tea.Model, mut Mutation) tea.Model {
	switch mut.Type {
	case MutationDropKey:
		return &dropKeyMutant{inner: model}
	case MutationSwapKeys:
		if strings.Contains(mut.Description, "up/down") {
			return &swapKeysMutant{inner: model, a: tea.KeyUp, b: tea.KeyDown}
		}
		return &swapKeysMutant{inner: model, a: tea.KeyTab, b: tea.KeyShiftTab}
	case MutationEmptyView:
		return &emptyViewMutant{inner: model}
	case MutationStaticView:
		return &staticViewMutant{inner: model, text: "MUTANT"}
	case MutationDropCmd:
		return &dropCmdMutant{inner: model}
	case MutationNilInit:
		return &nilInitMutant{inner: model}
	default:
		return model
	}
}

// runMutant runs the test function against a mutated model and returns
// true if the test detected the mutation (i.e., the test failed).
func runMutant(testFn func(testing.TB, tea.Model), model tea.Model) bool {
	mt := &mutantTB{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				mt.failed = true
			}
		}()
		testFn(mt, model)
	}()
	return mt.failed
}

// mutantTB is a minimal testing.TB that captures failures without aborting.
type mutantTB struct {
	testing.TB
	failed bool
}

func (m *mutantTB) Helper()                        {}
func (m *mutantTB) Error(a ...any)                 { m.failed = true }
func (m *mutantTB) Errorf(format string, a ...any) { m.failed = true }
func (m *mutantTB) Fatal(a ...any)                 { m.failed = true }
func (m *mutantTB) Fatalf(format string, a ...any) { m.failed = true }
func (m *mutantTB) Log(a ...any)                   {}
func (m *mutantTB) Logf(format string, a ...any)   {}
func (m *mutantTB) Name() string                   { return "mutant" }

// ── Mutant wrappers ─────────────────────────────────────────────────────

// dropKeyMutant silently swallows all tea.KeyMsg messages.
type dropKeyMutant struct{ inner tea.Model }

func (m *dropKeyMutant) Init() tea.Cmd { return m.inner.Init() }
func (m *dropKeyMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, nil
	}
	inner, cmd := m.inner.Update(msg)
	m.inner = inner
	return m, cmd
}
func (m *dropKeyMutant) View() string { return m.inner.View() }

// swapKeysMutant swaps two key types in Update.
type swapKeysMutant struct {
	inner tea.Model
	a, b  tea.KeyType
}

func (m *swapKeysMutant) Init() tea.Cmd { return m.inner.Init() }
func (m *swapKeysMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case m.a:
			km.Type = m.b
		case m.b:
			km.Type = m.a
		}
		msg = km
	}
	inner, cmd := m.inner.Update(msg)
	m.inner = inner
	return m, cmd
}
func (m *swapKeysMutant) View() string { return m.inner.View() }

// emptyViewMutant always returns "" from View.
type emptyViewMutant struct{ inner tea.Model }

func (m *emptyViewMutant) Init() tea.Cmd { return m.inner.Init() }
func (m *emptyViewMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.inner.Update(msg)
	m.inner = inner
	return m, cmd
}
func (m *emptyViewMutant) View() string { return "" }

// staticViewMutant always returns a fixed string from View.
type staticViewMutant struct {
	inner tea.Model
	text  string
}

func (m *staticViewMutant) Init() tea.Cmd { return m.inner.Init() }
func (m *staticViewMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.inner.Update(msg)
	m.inner = inner
	return m, cmd
}
func (m *staticViewMutant) View() string { return m.text }

// dropCmdMutant discards all commands returned from Update.
type dropCmdMutant struct{ inner tea.Model }

func (m *dropCmdMutant) Init() tea.Cmd { return m.inner.Init() }
func (m *dropCmdMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, _ := m.inner.Update(msg)
	m.inner = inner
	return m, nil
}
func (m *dropCmdMutant) View() string { return m.inner.View() }

// nilInitMutant returns nil from Init instead of the original command.
type nilInitMutant struct{ inner tea.Model }

func (m *nilInitMutant) Init() tea.Cmd { return nil }
func (m *nilInitMutant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.inner.Update(msg)
	m.inner = inner
	return m, cmd
}
func (m *nilInitMutant) View() string { return m.inner.View() }
