package blit

import (
	"sync/atomic"
	"testing"
)

// TestSignalOrphanSetNoPanic verifies that Set on a signal with no bus
// (orphan) does not panic.
func TestSignalOrphanSetNoPanic(t *testing.T) {
	s := NewSignal(42)
	// No attach — bus is nil
	s.Set(100) // must not panic
	if got := s.Get(); got != 100 {
		t.Errorf("orphan Set: Get = %d, want 100", got)
	}
}

// TestSignalGenerationCounter verifies that each Set increments gen.
func TestSignalGenerationCounter(t *testing.T) {
	s := NewSignal(0)

	if s.gen != 0 {
		t.Fatalf("initial gen = %d, want 0", s.gen)
	}

	s.Set(1)
	if s.gen != 1 {
		t.Fatalf("after first Set gen = %d, want 1", s.gen)
	}

	s.Set(2)
	s.Set(3)
	if s.gen != 3 {
		t.Fatalf("after 3 Sets gen = %d, want 3", s.gen)
	}
}

// TestSignalMultipleSubscribers verifies that all subscribers fire on flush.
func TestSignalMultipleSubscribers(t *testing.T) {
	s := NewSignal("")
	bus := newSignalBus()
	s.attach(bus)

	var got1, got2 string
	s.Subscribe(func(v string) { got1 = v })
	s.Subscribe(func(v string) { got2 = v })

	s.Set("hello")
	bus.drain()

	if got1 != "hello" {
		t.Errorf("subscriber 1 got %q, want %q", got1, "hello")
	}
	if got2 != "hello" {
		t.Errorf("subscriber 2 got %q, want %q", got2, "hello")
	}
}

// TestSignalUnsubscribeMiddle verifies that unsubscribing one of several
// subscribers only removes that one.
func TestSignalUnsubscribeMiddle(t *testing.T) {
	s := NewSignal(0)
	bus := newSignalBus()
	s.attach(bus)

	var a, b, c int32
	s.Subscribe(func(int) { atomic.AddInt32(&a, 1) })
	unsub := s.Subscribe(func(int) { atomic.AddInt32(&b, 1) })
	s.Subscribe(func(int) { atomic.AddInt32(&c, 1) })

	s.Set(1)
	bus.drain()

	if atomic.LoadInt32(&a) != 1 || atomic.LoadInt32(&b) != 1 || atomic.LoadInt32(&c) != 1 {
		t.Fatal("all three should fire once before unsub")
	}

	unsub()
	s.Set(2)
	bus.drain()

	if atomic.LoadInt32(&a) != 2 {
		t.Error("subscriber a should fire again")
	}
	if atomic.LoadInt32(&b) != 1 {
		t.Error("subscriber b should NOT fire after unsub")
	}
	if atomic.LoadInt32(&c) != 2 {
		t.Error("subscriber c should fire again")
	}
}

// TestSignalCoalescingMultipleSets verifies that N Set calls between
// drains produce exactly one subscriber invocation.
func TestSignalCoalescingMultipleSets(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"2 sets", 2},
		{"10 sets", 10},
		{"1000 sets", 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSignal(0)
			bus := newSignalBus()
			s.attach(bus)

			var calls int32
			s.Subscribe(func(int) { atomic.AddInt32(&calls, 1) })

			for i := 0; i < tt.count; i++ {
				s.Set(i)
			}
			bus.drain()

			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Errorf("subscriber fired %d times, want 1", got)
			}
		})
	}
}

// TestSignalComputedReDerives verifies Computed re-runs calc when deps change.
func TestSignalComputedReDerives(t *testing.T) {
	x := NewSignal(10)
	y := NewSignal(20)
	bus := newSignalBus()
	x.attach(bus)
	y.attach(bus)

	product := Computed([]AnySignal{x, y}, func() int { return x.Get() * y.Get() })
	product.attach(bus)

	if got := product.Get(); got != 200 {
		t.Fatalf("initial product = %d, want 200", got)
	}

	x.Set(5)
	bus.drain() // triggers Computed recalc
	bus.drain() // drain the product signal's own dirty

	if got := product.Get(); got != 100 {
		t.Fatalf("after x=5, product = %d, want 100", got)
	}

	y.Set(3)
	bus.drain()
	bus.drain()

	if got := product.Get(); got != 15 {
		t.Fatalf("after y=3, product = %d, want 15", got)
	}
}

// TestSignalComputedNoDeps verifies Computed with empty deps still computes
// the initial value.
func TestSignalComputedNoDeps(t *testing.T) {
	c := Computed([]AnySignal{}, func() string { return "static" })
	if got := c.Get(); got != "static" {
		t.Errorf("Computed with no deps = %q, want %q", got, "static")
	}
}

// TestSignalAttachIdempotent verifies that attaching the same signal to
// multiple buses only binds to the first one.
func TestSignalAttachIdempotent(t *testing.T) {
	s := NewSignal(0)
	bus1 := newSignalBus()
	bus2 := newSignalBus()

	s.attach(bus1)
	s.attach(bus2) // should be a no-op

	var called bool
	s.Subscribe(func(int) { called = true })

	s.Set(1)
	bus1.drain()
	if !called {
		t.Error("subscriber should fire via bus1")
	}

	called = false
	s.Set(2)
	bus2.drain() // bus2 should NOT have the signal
	if called {
		t.Error("subscriber should NOT fire via bus2 (attach is idempotent)")
	}
	// Clean up: drain bus1 to actually trigger
	bus1.drain()
}

// TestSignalBusDrainClearsPending verifies drain empties the queue.
func TestSignalBusDrainClearsPending(t *testing.T) {
	bus := newSignalBus()
	s := NewSignal(0)
	s.attach(bus)

	var calls int32
	s.Subscribe(func(int) { atomic.AddInt32(&calls, 1) })

	s.Set(1)
	bus.drain()
	bus.drain() // second drain should be a no-op

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (second drain is no-op)", got)
	}
}

// TestSignalSubscribeAny verifies the untyped subscribeAny interface.
func TestSignalSubscribeAny(t *testing.T) {
	s := NewSignal(42)
	bus := newSignalBus()
	s.attach(bus)

	var fired bool
	var anySig AnySignal = s
	unsub := anySig.subscribeAny(func() { fired = true })

	s.Set(99)
	bus.drain()

	if !fired {
		t.Error("subscribeAny callback should fire on flush")
	}

	fired = false
	unsub()
	s.Set(100)
	bus.drain()

	if fired {
		t.Error("subscribeAny should not fire after unsubscribe")
	}
}

// --- StringSource adapter tests ---

func TestFuncStringNil(t *testing.T) {
	src := FuncString(nil)
	if got := src.Value(); got != "" {
		t.Errorf("FuncString(nil).Value() = %q, want empty", got)
	}
}

func TestFuncStringNormal(t *testing.T) {
	src := FuncString(func() string { return "hello" })
	if got := src.Value(); got != "hello" {
		t.Errorf("FuncString.Value() = %q, want %q", got, "hello")
	}
}

func TestSignalStringNil(t *testing.T) {
	src := SignalString(nil)
	if got := src.Value(); got != "" {
		t.Errorf("SignalString(nil).Value() = %q, want empty", got)
	}
}

func TestSignalStringNormal(t *testing.T) {
	sig := NewSignal("world")
	src := SignalString(sig)
	if got := src.Value(); got != "world" {
		t.Errorf("SignalString.Value() = %q, want %q", got, "world")
	}
	sig.Set("updated")
	if got := src.Value(); got != "updated" {
		t.Errorf("after Set, SignalString.Value() = %q, want %q", got, "updated")
	}
}

func TestToStringSourceNil(t *testing.T) {
	src := toStringSource(nil)
	if src != nil {
		t.Errorf("toStringSource(nil) should return nil, got %v", src)
	}
}

func TestToStringSourceSignal(t *testing.T) {
	sig := NewSignal("sig")
	src := toStringSource(sig)
	if src == nil {
		t.Fatal("toStringSource(*Signal[string]) should not return nil")
	}
	if got := src.Value(); got != "sig" {
		t.Errorf("got %q, want %q", got, "sig")
	}
}

func TestToStringSourceFunc(t *testing.T) {
	fn := func() string { return "fn" }
	src := toStringSource(fn)
	if src == nil {
		t.Fatal("toStringSource(func() string) should not return nil")
	}
	if got := src.Value(); got != "fn" {
		t.Errorf("got %q, want %q", got, "fn")
	}
}

func TestToStringSourceStringSource(t *testing.T) {
	inner := FuncString(func() string { return "inner" })
	src := toStringSource(inner)
	if got := src.Value(); got != "inner" {
		t.Errorf("got %q, want %q", got, "inner")
	}
}

func TestToStringSourceUnsupported(t *testing.T) {
	src := toStringSource(42) // unsupported type
	if src == nil {
		t.Fatal("unsupported type should return a fallback source, not nil")
	}
	got := src.Value()
	if got == "" {
		t.Error("unsupported type should produce a non-empty fallback string")
	}
}
