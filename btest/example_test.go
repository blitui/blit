package btest_test

import (
	"fmt"
	"testing"

	"github.com/blitui/blit/btest"
	tea "github.com/charmbracelet/bubbletea"
)

// listModel is a minimal model used in examples.
type listModel struct {
	items  []string
	cursor int
	chosen string
}

func newListModel() *listModel {
	return &listModel{items: []string{"Alpha", "Beta", "Gamma"}}
}

func (m *listModel) Init() tea.Cmd { return nil }

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyEnter:
			m.chosen = m.items[m.cursor]
		}
	}
	return m, nil
}

func (m *listModel) View() string {
	s := ""
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		s += cursor + item + "\n"
	}
	if m.chosen != "" {
		s += "\nSelected: " + m.chosen
	}
	return s
}

func ExampleNewTestModel() {
	// Create a TestModel to drive a tea.Model synchronously.
	t := &testing.T{}
	model := newListModel()
	tm := btest.NewTestModel(t, model, 40, 10)

	// Send keys and inspect the screen.
	tm.SendKey("down")
	tm.SendKey("enter")
	fmt.Println(model.chosen)
	// Output: Beta
}

func ExampleAssertContains() {
	t := &testing.T{}
	model := newListModel()
	tm := btest.NewTestModel(t, model, 40, 10)

	// AssertContains checks that the screen contains the given text.
	btest.AssertContains(t, tm.Screen(), "Alpha")
	btest.AssertContains(t, tm.Screen(), "Beta")
	btest.AssertContains(t, tm.Screen(), "Gamma")
	fmt.Println("all items visible")
	// Output: all items visible
}

func ExampleAssertRowContains() {
	t := &testing.T{}
	model := newListModel()
	tm := btest.NewTestModel(t, model, 40, 10)

	// First row should have the cursor on Alpha.
	btest.AssertRowContains(t, tm.Screen(), 0, "> Alpha")
	fmt.Println("cursor on Alpha")
	// Output: cursor on Alpha
}

func ExampleNewHarness() {
	t := &testing.T{}
	model := newListModel()

	// Harness provides a fluent API for test scripts.
	btest.NewHarness(t, model, 40, 10).
		Keys("down", "down").
		Expect("Gamma").
		Done()
	fmt.Println("harness done")
	// Output: harness done
}

func ExampleScreen_FindText() {
	t := &testing.T{}
	model := newListModel()
	tm := btest.NewTestModel(t, model, 40, 10)

	row, col := tm.Screen().FindText("Beta")
	fmt.Printf("Beta at row=%d col=%d\n", row, col)
	// Output: Beta at row=1 col=2
}

func ExampleContrastRatio() {
	// Check WCAG contrast ratio between two colors.
	black := btest.RGB{R: 0, G: 0, B: 0}
	white := btest.RGB{R: 255, G: 255, B: 255}
	ratio := btest.ContrastRatio(black, white)
	fmt.Printf("contrast ratio: %.1f:1\n", ratio)
	// Output: contrast ratio: 21.0:1
}

func ExampleCheckContrast() {
	fg := btest.RGB{R: 0, G: 0, B: 0}
	bg := btest.RGB{R: 255, G: 255, B: 255}
	result := btest.CheckContrast(fg, bg)
	fmt.Printf("AA=%v AAA=%v ratio=%.1f\n", result.PassAA, result.PassAAA, result.Ratio)
	// Output: AA=true AAA=true ratio=21.0
}

func ExampleFuzz() {
	t := &testing.T{}
	model := newListModel()

	// Fuzz sends random inputs to find panics.
	result := btest.Fuzz(t, model, btest.FuzzConfig{
		Seed:       42,
		Iterations: 100,
	})
	fmt.Printf("panicked=%v iterations=%d\n", result.Panicked, result.Iterations)
	// Output: panicked=false iterations=100
}

func ExampleMutationTest() {
	t := &testing.T{}
	factory := func() tea.Model { return newListModel() }

	report := btest.MutationTest(t, factory, btest.MutationConfig{
		Mutations: []btest.MutationType{btest.MutationEmptyView, btest.MutationDropKey},
		Test: func(t testing.TB, model tea.Model) {
			tm := btest.NewTestModel(t, model, 40, 10)
			btest.AssertContains(t, tm.Screen(), "Alpha")
			tm.SendKey("down")
			tm.SendKey("enter")
			btest.AssertContains(t, tm.Screen(), "Selected")
		},
	})
	fmt.Printf("killed=%d total=%d score=%.0f%%\n", report.Killed, report.Total, report.Score())
	// Output: killed=2 total=2 score=100%
}

func ExampleSimulateColorBlind() {
	red := btest.RGB{R: 255, G: 0, B: 0}
	simulated := btest.SimulateColorBlind(red, btest.Protanopia)
	fmt.Printf("protanopia red → R=%d G=%d B=%d\n", simulated.R, simulated.G, simulated.B)
	// Output: protanopia red → R=198 G=197 B=0
}

func ExampleSmokeTest() {
	t := &testing.T{}
	model := newListModel()

	// SmokeTest runs automated checks against the model.
	btest.SmokeTest(t, model, nil)
	fmt.Println("smoke test passed")
	// Output: smoke test passed
}

func ExampleNewSessionRecorder() {
	t := &testing.T{}
	model := newListModel()
	tm := btest.NewTestModel(t, model, 40, 10)

	// Record a session.
	rec := btest.NewSessionRecorder(tm)
	rec.Key("down")
	rec.Key("enter")
	fmt.Println("recorded 2 actions")
	// Output: recorded 2 actions
}

func ExampleAssertSnapshot() {
	// AssertSnapshot compares screen content against a stored .snap file.
	// On first run, the snapshot is created. On subsequent runs, mismatches
	// cause the test to fail.
	//
	// Usage:
	//   tm := btest.NewTestModel(t, model, 80, 24)
	//   btest.AssertSnapshot(t, tm.Screen(), "my-component")
	//
	// Regenerate snapshots:
	//   go test ./... -args -btest.update
	fmt.Println("see usage in doc comment")
	// Output: see usage in doc comment
}
