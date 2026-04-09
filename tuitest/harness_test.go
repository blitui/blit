package tuitest_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

// harnessModel is a minimal tea.Model that records keys and renders them.
type harnessModel struct {
	keys    []string
	width   int
	height  int
	lastMsg tea.Msg
}

func (m *harnessModel) Init() tea.Cmd { return nil }
func (m *harnessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.lastMsg = msg
	switch v := msg.(type) {
	case tea.KeyMsg:
		m.keys = append(m.keys, v.String())
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
	}
	return m, nil
}
func (m *harnessModel) View() string {
	return fmt.Sprintf("keys=%v size=%dx%d", m.keys, m.width, m.height)
}

func TestHarness_KeysAndExpect(t *testing.T) {
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	h.Keys("down", "enter").Expect("down enter")
}

func TestHarness_TypeExpect(t *testing.T) {
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	h.Type("abc").Expect("a b c")
}

func TestHarness_ExpectNot(t *testing.T) {
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	h.ExpectNot("zzzzz")
}

func TestHarness_Resize(t *testing.T) {
	m := &harnessModel{}
	h := tuitest.NewHarness(t, m, 10, 5)
	h.Resize(30, 10)
	if m.width != 30 || m.height != 10 {
		t.Errorf("resize not applied: %dx%d", m.width, m.height)
	}
}

func TestHarness_SetupTeardownOrder(t *testing.T) {
	var order []string
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	h.OnSetup(func() { order = append(order, "setup-1") })
	h.OnSetup(func() { order = append(order, "setup-2") })
	h.OnTeardown(func() { order = append(order, "td-1") })
	h.OnTeardown(func() { order = append(order, "td-2") })
	h.Keys("a") // triggers setup
	h.Done()
	// Setup runs in registration order, teardown in LIFO.
	want := []string{"setup-1", "setup-2", "td-2", "td-1"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want[i])
		}
	}
}

func TestHarness_DoneIdempotent(t *testing.T) {
	calls := 0
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	h.OnTeardown(func() { calls++ })
	h.Done()
	h.Done()
	if calls != 1 {
		t.Errorf("teardown should run exactly once, got %d", calls)
	}
}

func TestHarness_Screen(t *testing.T) {
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	scr := h.Keys("x").Screen()
	if !scr.Contains("x") {
		t.Errorf("screen should contain x")
	}
}

func TestHarness_TestModelEscape(t *testing.T) {
	h := tuitest.NewHarness(t, &harnessModel{}, 40, 5)
	if h.TestModel() == nil {
		t.Error("TestModel should expose the underlying wrapper")
	}
}
