package btest

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// counterModel increments a counter on each key press and displays it.
type counterModel struct {
	count int
}

func (m *counterModel) Init() tea.Cmd { return nil }
func (m *counterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.count++
	}
	return m, nil
}
func (m *counterModel) View() string {
	if m.count == 0 {
		return "count: 0"
	}
	return "count: " + string(rune('0'+m.count))
}

func TestMutationTest_Basic(t *testing.T) {
	factory := func() tea.Model { return &counterModel{} }
	cfg := MutationConfig{
		Test: func(t testing.TB, model tea.Model) {
			tm := NewTestModel(t, model, 40, 10)
			tm.SendKey("a")
			AssertContains(t, tm.Screen(), "count: 1")
		},
	}

	report := MutationTest(t, factory, cfg)
	if report.Total == 0 {
		t.Fatal("expected at least one mutation")
	}
	if report.Killed == 0 {
		t.Error("expected at least one mutation to be killed")
	}
	// The test checks view content after key press, so:
	// - drop-key: killed (count stays 0)
	// - empty-view: killed (no content)
	// - static-view: killed (shows MUTANT not count)
	t.Logf("mutation report:\n%s", report.Summary())
}

func TestMutationTest_SpecificMutations(t *testing.T) {
	factory := func() tea.Model { return &counterModel{} }
	cfg := MutationConfig{
		Mutations: []MutationType{MutationDropKey, MutationEmptyView},
		Test: func(t testing.TB, model tea.Model) {
			tm := NewTestModel(t, model, 40, 10)
			tm.SendKey("a")
			AssertContains(t, tm.Screen(), "count: 1")
		},
	}

	report := MutationTest(t, factory, cfg)
	if report.Total != 2 {
		t.Errorf("total = %d, want 2", report.Total)
	}
	if report.Killed != 2 {
		t.Errorf("killed = %d, want 2", report.Killed)
	}
}

func TestMutationReport_Score(t *testing.T) {
	r := &MutationReport{Killed: 3, Survived: 1, Total: 4}
	if score := r.Score(); score != 75 {
		t.Errorf("Score() = %.1f, want 75.0", score)
	}
}

func TestMutationReport_ScoreEmpty(t *testing.T) {
	r := &MutationReport{}
	if score := r.Score(); score != 100 {
		t.Errorf("Score() = %.1f, want 100.0 (no mutations)", score)
	}
}

func TestMutationReport_Summary(t *testing.T) {
	r := &MutationReport{
		Results: []MutationResult{
			{Mutation: Mutation{Type: MutationDropKey, Description: "drop keys"}, Killed: true},
			{Mutation: Mutation{Type: MutationEmptyView, Description: "empty view"}, Killed: false},
		},
		Killed: 1, Survived: 1, Total: 2,
	}
	s := r.Summary()
	if s == "" {
		t.Error("Summary() returned empty")
	}
	if !strings.Contains(s, "KILLED") || !strings.Contains(s, "SURVIVED") {
		t.Errorf("Summary missing status labels:\n%s", s)
	}
}

func TestMutationReport_Survivors(t *testing.T) {
	r := &MutationReport{
		Results: []MutationResult{
			{Mutation: Mutation{Type: MutationDropKey}, Killed: true},
			{Mutation: Mutation{Type: MutationEmptyView}, Killed: false},
		},
	}
	survivors := r.Survivors()
	if len(survivors) != 1 {
		t.Errorf("Survivors() = %d, want 1", len(survivors))
	}
	if survivors[0].Mutation.Type != MutationEmptyView {
		t.Errorf("survivor type = %v, want MutationEmptyView", survivors[0].Mutation.Type)
	}
}

func TestAssertMutationScore_Pass(t *testing.T) {
	r := &MutationReport{Killed: 9, Total: 10}
	AssertMutationScore(t, r, 80.0)
}

func TestAssertMutationScore_Fail(t *testing.T) {
	ft := &stubTB{}
	r := &MutationReport{Killed: 1, Survived: 9, Total: 10}
	AssertMutationScore(ft, r, 80.0)
	if !ft.failed {
		t.Error("expected AssertMutationScore to fail for low score")
	}
}

func TestMutationType_String(t *testing.T) {
	cases := []struct {
		mt   MutationType
		want string
	}{
		{MutationDropKey, "drop-key"},
		{MutationSwapKeys, "swap-keys"},
		{MutationEmptyView, "empty-view"},
		{MutationStaticView, "static-view"},
		{MutationDropCmd, "drop-cmd"},
		{MutationNilInit, "nil-init"},
		{MutationType(99), "unknown(99)"},
	}
	for _, tc := range cases {
		if got := tc.mt.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.mt, got, tc.want)
		}
	}
}

// initModel has a meaningful Init command.
type initModel struct {
	ready bool
}

func (m *initModel) Init() tea.Cmd {
	return func() tea.Msg { return "ready" }
}
func (m *initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s, ok := msg.(string); ok && s == "ready" {
		m.ready = true
	}
	return m, nil
}
func (m *initModel) View() string {
	if m.ready {
		return "READY"
	}
	return "NOT READY"
}

func TestMutantTB_Works(t *testing.T) {
	mt := &mutantTB{}
	mt.Errorf("test %s", "error")
	if !mt.failed {
		t.Error("mutantTB.Errorf should set failed=true")
	}
}

func TestRunMutant_DirectCall(t *testing.T) {
	testFn := func(t testing.TB, model tea.Model) {
		t.Errorf("forced failure")
	}
	killed := runMutant(testFn, &counterModel{})
	if !killed {
		t.Error("runMutant should detect t.Errorf call")
	}
}

func TestMutationTest_NilInit(t *testing.T) {
	factory := func() tea.Model { return &initModel{} }

	// Directly test the runMutant function.
	mutated := applyMutation(factory(), Mutation{Type: MutationNilInit})

	// Verify Init returns nil.
	if cmd := mutated.Init(); cmd != nil {
		t.Fatal("nilInitMutant Init should return nil")
	}

	// Verify view is NOT READY.
	if v := mutated.View(); v != "NOT READY" {
		t.Fatalf("expected NOT READY, got %q", v)
	}

	// Run with a test that checks the view exactly.
	testFn := func(tb testing.TB, m tea.Model) {
		v := m.View()
		if v != "READY" {
			tb.Errorf("expected exactly READY, got %q", v)
		}
	}
	killed := runMutant(testFn, mutated)
	if !killed {
		t.Errorf("direct runMutant should detect nil-init mutation, view=%q", mutated.View())
	}
}

func TestMutationTest_SwapKeys(t *testing.T) {
	// Verify swap mutations are generated correctly.
	muts := buildMutations(MutationSwapKeys)
	if len(muts) != 2 {
		t.Errorf("buildMutations(SwapKeys) = %d mutations, want 2", len(muts))
	}
}

func TestMutationTest_DropCmd(t *testing.T) {
	// Model that returns a command from Update.
	factory := func() tea.Model { return &initCmdModel{} }
	cfg := MutationConfig{
		Mutations: []MutationType{MutationDropCmd},
		Test: func(t testing.TB, model tea.Model) {
			tm := NewTestModel(t, model, 40, 10)
			tm.SendKey("a")
			// The initCmdModel should have been initialized by its own init cmd.
			// With drop-cmd, the cmd from SendKey's Update is lost.
			// Since initCmdModel tracks "initialized" via a string msg,
			// this tests whether the cmd pipeline matters.
		},
	}
	report := MutationTest(t, factory, cfg)
	if report.Total != 1 {
		t.Errorf("total = %d, want 1", report.Total)
	}
}

