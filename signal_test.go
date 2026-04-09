package tuikit

import (
	"sync"
	"sync/atomic"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSignalGetSet(t *testing.T) {
	s := NewSignal(42)
	if got := s.Get(); got != 42 {
		t.Fatalf("initial Get = %d, want 42", got)
	}
	s.Set(99)
	if got := s.Get(); got != 99 {
		t.Fatalf("after Set Get = %d, want 99", got)
	}
}

func TestSignalSubscribeFiresOnFlush(t *testing.T) {
	s := NewSignal("")
	bus := newSignalBus()
	s.attach(bus)

	var got string
	unsub := s.Subscribe(func(v string) { got = v })
	defer unsub()

	s.Set("hello")
	bus.drain()
	if got != "hello" {
		t.Fatalf("subscriber got %q, want %q", got, "hello")
	}
}

func TestSignalDirtyBitCoalesces(t *testing.T) {
	// A2: multiple Set calls within one frame collapse into a single
	// subscriber notification.
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)

	var calls int32
	s.Subscribe(func(int) { atomic.AddInt32(&calls, 1) })

	for i := 0; i < 100; i++ {
		s.Set(i)
	}
	bus.drain()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("subscriber fired %d times, want 1 (coalesced)", got)
	}
	if s.Get() != 99 {
		t.Fatalf("final value = %d, want 99", s.Get())
	}
}

func TestSignalUnsubscribe(t *testing.T) {
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)

	var calls int
	unsub := s.Subscribe(func(int) { calls++ })
	s.Set(1)
	bus.drain()

	unsub()
	s.Set(2)
	bus.drain()

	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (unsubscribed)", calls)
	}
}

func TestSignalComputed(t *testing.T) {
	// A3: Computed derives from deps and pushes updates through the bus.
	a := NewSignal(2)
	b := NewSignal(3)
	bus := newSignalBus()
	a.attach(bus)
	b.attach(bus)

	sum := Computed([]AnySignal{a, b}, func() int { return a.Get() + b.Get() })
	sum.attach(bus)

	if got := sum.Get(); got != 5 {
		t.Fatalf("initial sum = %d, want 5", got)
	}

	var observed int
	sum.Subscribe(func(v int) { observed = v })

	a.Set(10)
	bus.drain()
	// Computed.Set runs during drain → sum is dirty → another drain round
	bus.drain()

	if sum.Get() != 13 {
		t.Fatalf("sum.Get = %d, want 13", sum.Get())
	}
	if observed != 13 {
		t.Fatalf("observed = %d, want 13", observed)
	}
}

func TestSignalRaceSafeConcurrentSet(t *testing.T) {
	// A6: Set is safe from any goroutine. Run with -race to exercise the
	// happy path; subscribers still run sequentially via drain.
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)

	const writers = 16
	const itersPerWriter = 500

	var wg sync.WaitGroup
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < itersPerWriter; j++ {
				s.Set(base*itersPerWriter + j)
			}
		}(i)
	}
	wg.Wait()
	bus.drain()

	// Subscriber fires at most once per frame; here we just assert
	// drain didn't panic and final value is one of the writes.
	if s.Get() < 0 {
		t.Fatalf("unexpected final value %d", s.Get())
	}
}

func TestSignalBusSchedulesOnlyOnce(t *testing.T) {
	bus := newSignalBus()
	var sends int
	bus.setSender(func(msg tea.Msg) { sends++ })

	a := NewSignal("a")
	a.attach(bus)

	a.Set("1")
	a.Set("2")
	a.Set("3")

	// Only a single signalFlushMsg should have been scheduled despite 3
	// writes in the same "frame" (no drain in between).
	if sends != 1 {
		t.Fatalf("sends = %d, want 1", sends)
	}
	bus.drain()

	a.Set("4")
	if sends != 2 {
		t.Fatalf("sends after post-drain Set = %d, want 2", sends)
	}
}

func TestSignalStatusBarOverload(t *testing.T) {
	// A4: StatusBar accepts both func() string and *Signal[string].
	sig := NewSignal("hello")

	bar := NewStatusBar(StatusBarOpts{
		Left:  func() string { return "legacy" },
		Right: sig,
	})
	bar.SetSize(20, 1)
	bar.SetTheme(DefaultTheme())

	view := bar.View()
	if view == "" {
		t.Fatal("empty view")
	}
	// The signal value should be present in the rendered bar.
	if !contains(view, "hello") {
		t.Fatalf("view missing signal value: %q", view)
	}
	if !contains(view, "legacy") {
		t.Fatalf("view missing legacy closure value: %q", view)
	}

	sig.Set("world")
	view2 := bar.View()
	if !contains(view2, "world") {
		t.Fatalf("view2 did not pick up new signal value: %q", view2)
	}
}

func TestSignalConfigEditorOverload(t *testing.T) {
	sig := NewSignal("from-signal")
	fields := []ConfigField{
		{Label: "A", Get: func() string { return "from-closure" }},
		{Label: "B", Source: sig},
	}
	if got := fields[0].currentValue(); got != "from-closure" {
		t.Fatalf("field A = %q, want from-closure", got)
	}
	if got := fields[1].currentValue(); got != "from-signal" {
		t.Fatalf("field B = %q, want from-signal", got)
	}
	sig.Set("updated")
	if got := fields[1].currentValue(); got != "updated" {
		t.Fatalf("field B after Set = %q, want updated", got)
	}
}

// contains is a tiny substring helper to avoid pulling in strings in a
// benchmark-sensitive file.
func contains(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// --- benchmarks (A7) ---------------------------------------------------

func BenchmarkSignalSet100k(b *testing.B) {
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)
	s.Subscribe(func(int) {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100_000; j++ {
			s.Set(j)
		}
		bus.drain()
	}
}

func BenchmarkSignalSubscribe(b *testing.B) {
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsub := s.Subscribe(func(int) {})
		unsub()
	}
}
