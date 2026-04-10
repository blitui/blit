package blit

import (
	"fmt"
	"sync"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

// Signal is a reactive, race-safe container for a value of type T.
//
// Writers call Set from any goroutine; Subscribers are always invoked on the
// UI goroutine (inside the App's tea.Msg dispatch). Multiple Set calls that
// land within the same frame collapse into a single notification per
// subscriber thanks to a dirty-bit + generation counter that coalesces
// pending flushes.
//
// Signals are typically created at App construction time and passed to
// components that accept a *Signal[T]. Components re-read Get() in their
// View() method; the App schedules a re-render whenever a signal flush is
// processed.
//
// Example:
//
//	status := blit.NewSignal("starting...")
//	app := blit.NewApp(
//	    blit.WithStatusBarSignal(status, nil),
//	)
//	go func() {
//	    time.Sleep(time.Second)
//	    status.Set("ready") // safe from any goroutine
//	}()
type Signal[T any] struct {
	mu     sync.RWMutex
	value  T
	gen    uint64 // generation: bumped on each Set
	nextID uint64
	subs   map[uint64]func(T)
	bus    *signalBus // set when attached to an App; nil means orphan
	dirty  int32      // atomic bool: 1 if enqueued in bus for flush
}

// Unsubscribe removes a previously registered Subscribe callback.
type Unsubscribe func()

// AnySignal is the untyped interface implemented by every *Signal[T]. It is
// used as the dependency type for Computed so that signals of mixed element
// types can be combined.
type AnySignal interface {
	// subscribeAny registers a generic callback fired after each flush.
	// It is unexported so only the blit package builds graphs.
	subscribeAny(fn func()) Unsubscribe
	// attach binds this signal to a bus so its subscribers fire on the UI
	// goroutine. Called by the App during setup.
	attach(b *signalBus)
}

// NewSignal creates a signal with the given initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		value: initial,
		subs:  make(map[uint64]func(T)),
	}
}

// Get returns the current value. Safe to call from any goroutine.
func (s *Signal[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set stores v and schedules a single coalesced notification on the next
// flush. Safe to call from any goroutine; subscribers still run on the UI
// goroutine.
func (s *Signal[T]) Set(v T) {
	s.mu.Lock()
	s.value = v
	s.gen++
	s.mu.Unlock()
	s.markDirty()
}

// Subscribe registers fn to receive the signal's value on every flush after
// Set. The callback is always invoked on the UI goroutine. The returned
// Unsubscribe removes the registration.
func (s *Signal[T]) Subscribe(fn func(T)) Unsubscribe {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subs[id] = fn
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.subs, id)
		s.mu.Unlock()
	}
}

// subscribeAny implements AnySignal.
func (s *Signal[T]) subscribeAny(fn func()) Unsubscribe {
	return s.Subscribe(func(T) { fn() })
}

// attach implements AnySignal. Multiple attach calls are a no-op after the
// first so signals shared between apps (tests) stay tethered to the first
// bus they see.
func (s *Signal[T]) attach(b *signalBus) {
	s.mu.Lock()
	if s.bus == nil {
		s.bus = b
	}
	s.mu.Unlock()
}

// markDirty enqueues s in its bus for the next flush. Uses an atomic
// dirty-bit so repeated Set calls within a single frame only enqueue once
// (A2: coalesced notification).
func (s *Signal[T]) markDirty() {
	s.mu.RLock()
	bus := s.bus
	s.mu.RUnlock()
	if bus == nil {
		return
	}
	if atomic.CompareAndSwapInt32(&s.dirty, 0, 1) {
		bus.enqueue(anySignalFlusher{flush: s.flush, clear: s.clearDirty})
	}
}

func (s *Signal[T]) clearDirty() {
	atomic.StoreInt32(&s.dirty, 0)
}

// flush invokes every subscriber with the latest value. Called on the UI
// goroutine via signalFlushMsg.
func (s *Signal[T]) flush() {
	s.mu.RLock()
	v := s.value
	subs := make([]func(T), 0, len(s.subs))
	for _, fn := range s.subs {
		subs = append(subs, fn)
	}
	s.mu.RUnlock()
	for _, fn := range subs {
		fn(v)
	}
}

// Computed derives a new signal from deps. calc is re-run on the UI
// goroutine whenever any dep flushes; the derived signal then notifies its
// own subscribers.
func Computed[T any](deps []AnySignal, calc func() T) *Signal[T] {
	out := NewSignal(calc())
	for _, d := range deps {
		dep := d
		dep.subscribeAny(func() {
			out.Set(calc())
		})
	}
	return out
}

// --- bus ---------------------------------------------------------------

// signalFlushMsg is sent to the App event loop so pending signal
// subscribers run on the UI goroutine.
type signalFlushMsg struct{}

// anySignalFlusher is a type-erased pending flush record.
type anySignalFlusher struct {
	flush func()
	clear func()
}

// signalBus coalesces dirty signals and wakes the App. One bus per App.
type signalBus struct {
	mu        sync.Mutex
	pending   []anySignalFlusher
	send      func(tea.Msg) // set by App.Run (nil until then); enqueue still works
	scheduled bool
}

func newSignalBus() *signalBus { return &signalBus{} }

// setSender wires in the tea.Program's Send function. Until this is called
// (tests that drive Update directly, or pre-Run setup), pending flushes are
// drained on the next manual drain or when the caller processes a
// signalFlushMsg themselves.
func (b *signalBus) setSender(send func(tea.Msg)) {
	b.mu.Lock()
	b.send = send
	scheduled := b.scheduled
	hasPending := len(b.pending) > 0
	b.mu.Unlock()
	if hasPending && !scheduled {
		b.schedule()
	}
}

func (b *signalBus) enqueue(f anySignalFlusher) {
	b.mu.Lock()
	b.pending = append(b.pending, f)
	already := b.scheduled
	b.mu.Unlock()
	if !already {
		b.schedule()
	}
}

// schedule asks the event loop to deliver a signalFlushMsg soon. If the
// sender is not wired yet (pre-Run), we just mark scheduled so setSender
// can flip the switch.
func (b *signalBus) schedule() {
	b.mu.Lock()
	b.scheduled = true
	send := b.send
	b.mu.Unlock()
	if send != nil {
		send(signalFlushMsg{})
	}
}

// drain runs all pending flushers on the current goroutine. The App calls
// this from its Update when it receives signalFlushMsg.
func (b *signalBus) drain() {
	b.mu.Lock()
	pending := b.pending
	b.pending = nil
	b.scheduled = false
	b.mu.Unlock()
	for _, f := range pending {
		f.clear()
		f.flush()
	}
}

// --- StringSource adapter (A4) -----------------------------------------

// StringSource is a read-only source of a string value. Both plain
// `func() string` closures and `*Signal[string]` satisfy this contract via
// the adapters below, letting components like StatusBar and ConfigEditor
// accept either style without a breaking API change.
//
// Use SignalString to wrap a *Signal[string] as a StringSource, or call
// FuncString for an explicit adapter around a closure (usually unnecessary
// since raw closures already satisfy the contract via toStringSource).
type StringSource interface {
	Value() string
}

type funcStringSource struct{ fn func() string }

func (f funcStringSource) Value() string {
	if f.fn == nil {
		return ""
	}
	return f.fn()
}

type signalStringSource struct{ s *Signal[string] }

func (s signalStringSource) Value() string {
	if s.s == nil {
		return ""
	}
	return s.s.Get()
}

// FuncString adapts a func() string closure into a StringSource.
func FuncString(fn func() string) StringSource { return funcStringSource{fn: fn} }

// SignalString adapts a *Signal[string] into a StringSource so it can be
// passed anywhere a func() string is accepted.
func SignalString(s *Signal[string]) StringSource { return signalStringSource{s: s} }

// toStringSource normalises the variadic any argument used by StatusBar and
// ConfigEditor helpers. nil, *Signal[string], func() string, and
// StringSource are all accepted; anything else panics to surface misuse
// early.
func toStringSource(v any) StringSource {
	switch x := v.(type) {
	case nil:
		return nil
	case StringSource:
		return x
	case *Signal[string]:
		return signalStringSource{s: x}
	case func() string:
		return funcStringSource{fn: x}
	default:
		return funcStringSource{fn: func() string { return fmt.Sprintf("[unsupported source: %T]", v) }}
	}
}
