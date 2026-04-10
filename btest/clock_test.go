package btest

import (
	"sync"
	"testing"
	"time"
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
