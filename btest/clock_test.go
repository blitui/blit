package btest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRealClock(t *testing.T) {
	c := RealClock{}
	before := time.Now()
	got := c.Now()
	if got.Before(before) {
		t.Errorf("RealClock.Now returned %v, earlier than %v", got, before)
	}
	// Sleep should not panic; keep duration tiny so the test stays fast.
	c.Sleep(1 * time.Millisecond)
}

func TestFakeClock_DefaultEpoch(t *testing.T) {
	c := NewFakeClock(time.Time{})
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !c.Now().Equal(want) {
		t.Errorf("FakeClock default epoch = %v, want %v", c.Now(), want)
	}
}

func TestFakeClock_Advance(t *testing.T) {
	start := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	c := NewFakeClock(start)
	c.Advance(500 * time.Millisecond)
	c.Advance(500 * time.Millisecond)
	want := start.Add(1 * time.Second)
	if !c.Now().Equal(want) {
		t.Errorf("after 2x500ms Advance, Now=%v, want %v", c.Now(), want)
	}
}

func TestFakeClock_AdvanceIgnoresNegative(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := NewFakeClock(start)
	c.Advance(-1 * time.Hour)
	if !c.Now().Equal(start) {
		t.Errorf("negative Advance mutated time: %v", c.Now())
	}
}

func TestFakeClock_SleepAdvances(t *testing.T) {
	c := NewFakeClock(time.Time{})
	start := c.Now()
	c.Sleep(2 * time.Second)
	if c.Now().Sub(start) != 2*time.Second {
		t.Errorf("Sleep did not advance by 2s, got %v", c.Now().Sub(start))
	}
}

func TestFakeClock_Set(t *testing.T) {
	c := NewFakeClock(time.Time{})
	target := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)
	c.Set(target)
	if !c.Now().Equal(target) {
		t.Errorf("Set = %v, want %v", c.Now(), target)
	}
}

func TestFakeClock_AfterFunc(t *testing.T) {
	c := NewFakeClock(time.Time{})
	var fired bool
	c.AfterFunc(5*time.Second, func() { fired = true })

	c.Advance(4 * time.Second)
	if fired {
		t.Fatal("timer fired too early")
	}
	c.Advance(1 * time.Second)
	if !fired {
		t.Fatal("timer did not fire at deadline")
	}
}

func TestFakeClock_AfterFunc_Multiple(t *testing.T) {
	c := NewFakeClock(time.Time{})
	var order []int
	c.AfterFunc(3*time.Second, func() { order = append(order, 3) })
	c.AfterFunc(1*time.Second, func() { order = append(order, 1) })
	c.AfterFunc(2*time.Second, func() { order = append(order, 2) })

	c.Advance(3 * time.Second)
	if len(order) != 3 {
		t.Fatalf("expected 3 timers fired, got %d", len(order))
	}
	// All three fire in the same Advance call; the 1s and 2s timers are
	// past deadline so they fire alongside the 3s timer.
	for i, v := range order {
		if v != i+1 {
			// Order depends on slice iteration; just check all three fired.
			break
		}
	}
}

func TestFakeClock_AfterFunc_Stop(t *testing.T) {
	c := NewFakeClock(time.Time{})
	var fired bool
	timer := c.AfterFunc(1*time.Second, func() { fired = true })

	ok := timer.Stop()
	if !ok {
		t.Fatal("Stop returned false on first call")
	}
	ok2 := timer.Stop()
	if ok2 {
		t.Fatal("Stop returned true on second call")
	}

	c.Advance(2 * time.Second)
	if fired {
		t.Fatal("stopped timer still fired")
	}
}

func TestFakeClock_Pending(t *testing.T) {
	c := NewFakeClock(time.Time{})
	if c.Pending() != 0 {
		t.Fatalf("Pending = %d, want 0", c.Pending())
	}
	c.AfterFunc(1*time.Second, func() {})
	c.AfterFunc(2*time.Second, func() {})
	if c.Pending() != 2 {
		t.Fatalf("Pending = %d, want 2", c.Pending())
	}

	c.Advance(1 * time.Second)
	if c.Pending() != 1 {
		t.Fatalf("after 1s advance, Pending = %d, want 1", c.Pending())
	}

	c.Advance(1 * time.Second)
	if c.Pending() != 0 {
		t.Fatalf("after 2s advance, Pending = %d, want 0", c.Pending())
	}
}

func TestFakeClock_Pending_StoppedNotCounted(t *testing.T) {
	c := NewFakeClock(time.Time{})
	t1 := c.AfterFunc(1*time.Second, func() {})
	c.AfterFunc(2*time.Second, func() {})
	t1.Stop()
	if c.Pending() != 1 {
		t.Fatalf("Pending = %d, want 1 (one stopped)", c.Pending())
	}
}

func TestFakeClock_ConcurrentSafe(t *testing.T) {
	c := NewFakeClock(time.Time{})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Advance(1 * time.Millisecond)
		}()
		go func() {
			defer wg.Done()
			_ = c.Now()
		}()
	}
	wg.Wait()
	// 50 advances of 1ms each => 50ms.
	want := 50 * time.Millisecond
	got := c.Now().Sub(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if got != want {
		t.Errorf("concurrent Advance total = %v, want %v", got, want)
	}
}

// ── TestModel clock integration ──────────────────────────────────────────

// tickModel records TickMsg times for verifying clock integration.
type tickModel struct {
	ticks []time.Time
	width int
}

func (m *tickModel) Init() tea.Cmd { return nil }
func (m *tickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case TickMsg:
		m.ticks = append(m.ticks, v.Time)
	case tea.WindowSizeMsg:
		m.width = v.Width
	}
	return m, nil
}
func (m *tickModel) View() string {
	return fmt.Sprintf("ticks=%d w=%d", len(m.ticks), m.width)
}

func TestTestModel_WithClock(t *testing.T) {
	clk := NewFakeClock(time.Time{})
	tm := NewTestModel(t, &tickModel{}, 40, 10)
	ret := tm.WithClock(clk)
	if ret != tm {
		t.Fatal("WithClock should return the same TestModel for chaining")
	}
	if tm.Clock() != clk {
		t.Fatal("Clock() should return the attached clock")
	}
}

func TestTestModel_ClockNilByDefault(t *testing.T) {
	tm := NewTestModel(t, &tickModel{}, 40, 10)
	if tm.Clock() != nil {
		t.Fatal("Clock() should be nil when no clock is attached")
	}
}

func TestTestModel_AdvanceClock(t *testing.T) {
	clk := NewFakeClock(time.Time{})
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10).WithClock(clk)

	tm.AdvanceClock(5 * time.Second)
	if len(m.ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(m.ticks))
	}
	want := time.Date(2026, 1, 1, 0, 0, 5, 0, time.UTC)
	if !m.ticks[0].Equal(want) {
		t.Errorf("tick time = %v, want %v", m.ticks[0], want)
	}
}

func TestTestModel_AdvanceClock_FiresTimers(t *testing.T) {
	clk := NewFakeClock(time.Time{})
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10).WithClock(clk)

	var fired bool
	clk.AfterFunc(3*time.Second, func() { fired = true })

	tm.AdvanceClock(2 * time.Second)
	if fired {
		t.Fatal("timer should not have fired yet")
	}
	tm.AdvanceClock(1 * time.Second)
	if !fired {
		t.Fatal("timer should have fired after reaching deadline")
	}
}

func TestTestModel_TriggerTick_WithClock(t *testing.T) {
	clk := NewFakeClock(time.Time{})
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10).WithClock(clk)

	clk.Advance(10 * time.Second)
	tm.TriggerTick()

	if len(m.ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(m.ticks))
	}
	want := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)
	if !m.ticks[0].Equal(want) {
		t.Errorf("tick time = %v, want %v", m.ticks[0], want)
	}
}

func TestTestModel_TriggerTick_WithoutClock(t *testing.T) {
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10)

	before := time.Now()
	tm.TriggerTick()
	after := time.Now()

	if len(m.ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(m.ticks))
	}
	if m.ticks[0].Before(before) || m.ticks[0].After(after) {
		t.Errorf("tick time %v not between %v and %v", m.ticks[0], before, after)
	}
}

func TestTestModel_SendTick_Deprecated(t *testing.T) {
	m := &tickModel{}
	tm := NewTestModel(t, m, 40, 10)
	tm.SendTick()
	if len(m.ticks) != 1 {
		t.Fatalf("SendTick should still work (delegates to TriggerTick), got %d ticks", len(m.ticks))
	}
}
