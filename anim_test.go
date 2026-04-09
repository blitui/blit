package tuikit

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// fixed clock helpers — no wall-clock dependency
var t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func tAt(ms int) time.Time {
	return t0.Add(time.Duration(ms) * time.Millisecond)
}

// --- Ease functions ---

func TestLinear(t *testing.T) {
	for _, tc := range []struct{ in, want float64 }{
		{0, 0}, {0.5, 0.5}, {1, 1},
	} {
		if got := Linear(tc.in); got != tc.want {
			t.Errorf("Linear(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestEaseInOut(t *testing.T) {
	if got := EaseInOut(0); got != 0 {
		t.Errorf("EaseInOut(0) = %v", got)
	}
	if got := EaseInOut(1); got != 1 {
		t.Errorf("EaseInOut(1) = %v", got)
	}
	if got := EaseInOut(0.5); got != 0.5 {
		t.Errorf("EaseInOut(0.5) = %v, want 0.5", got)
	}
}

func TestEaseOutCubic(t *testing.T) {
	if got := EaseOutCubic(0); got != 0 {
		t.Errorf("EaseOutCubic(0) = %v", got)
	}
	if got := EaseOutCubic(1); got != 1 {
		t.Errorf("EaseOutCubic(1) = %v", got)
	}
}

func TestEaseInCubic(t *testing.T) {
	if got := EaseInCubic(0); got != 0 {
		t.Errorf("EaseInCubic(0) = %v", got)
	}
	if got := EaseInCubic(1); got != 1 {
		t.Errorf("EaseInCubic(1) = %v", got)
	}
}

func TestEaseOutExpo(t *testing.T) {
	if got := EaseOutExpo(0); got != 0 {
		t.Errorf("EaseOutExpo(0) = %v", got)
	}
	if got := EaseOutExpo(1); got != 1 {
		t.Errorf("EaseOutExpo(1) = %v", got)
	}
}

// --- Interpolate ---

func TestInterpolateFloat64(t *testing.T) {
	cases := []struct {
		from, to, t, want float64
	}{
		{0, 100, 0, 0},
		{0, 100, 1, 100},
		{0, 100, 0.5, 50},
		{20, 40, 0.25, 25},
	}
	for _, tc := range cases {
		got := Interpolate[float64](tc.from, tc.to, tc.t, Linear)
		if got != tc.want {
			t.Errorf("Interpolate float64(%v->%v @%v) = %v, want %v", tc.from, tc.to, tc.t, got, tc.want)
		}
	}
}

func TestInterpolateInt(t *testing.T) {
	cases := []struct {
		from, to int
		t        float64
		want     int
	}{
		{0, 10, 0, 0},
		{0, 10, 1, 10},
		{0, 10, 0.5, 5},
		{0, 10, 0.3, 3},
	}
	for _, tc := range cases {
		got := Interpolate[int](tc.from, tc.to, tc.t, Linear)
		if got != tc.want {
			t.Errorf("Interpolate int(%v->%v @%v) = %v, want %v", tc.from, tc.to, tc.t, got, tc.want)
		}
	}
}

func TestInterpolateColor(t *testing.T) {
	from := lipgloss.Color("#000000")
	to := lipgloss.Color("#ffffff")
	// 0 + 255*0.5 = 127.5, math.Round gives 128 = 0x80
	mid := Interpolate[lipgloss.Color](from, to, 0.5, Linear)
	if string(mid) != "#808080" {
		t.Errorf("Interpolate color midpoint = %q, want #808080", string(mid))
	}
	got0 := Interpolate[lipgloss.Color](from, to, 0, Linear)
	if string(got0) != "#000000" {
		t.Errorf("Interpolate color t=0 = %q, want #000000", string(got0))
	}
	got1 := Interpolate[lipgloss.Color](from, to, 1, Linear)
	if string(got1) != "#ffffff" {
		t.Errorf("Interpolate color t=1 = %q, want #ffffff", string(got1))
	}
}

// --- Tween (fixed clock, no wall-clock dependency) ---

func TestTweenProgress(t *testing.T) {
	if animDisabled {
		t.Skip("TUIKIT_NO_ANIM=1 set, skipping tween progress test")
	}
	tw := &Tween{Duration: 100 * time.Millisecond}
	tw.Start(t0)

	if p := tw.Progress(tAt(0)); p != 0 {
		t.Errorf("progress at 0ms = %v, want 0", p)
	}
	if p := tw.Progress(tAt(50)); p != 0.5 {
		t.Errorf("progress at 50ms = %v, want 0.5", p)
	}
	if p := tw.Progress(tAt(100)); p != 1 {
		t.Errorf("progress at 100ms = %v, want 1", p)
	}
	if !tw.Done() {
		t.Error("tween should be done at 100ms")
	}
	if p := tw.Progress(tAt(200)); p != 1 {
		t.Errorf("progress past end = %v, want 1", p)
	}
}

func TestTweenNotStarted(t *testing.T) {
	tw := &Tween{Duration: 100 * time.Millisecond}
	if p := tw.Progress(t0); p != 1 {
		t.Errorf("unstarted tween progress = %v, want 1", p)
	}
	if !tw.Done() {
		t.Error("unstarted tween should be done")
	}
}

func TestTweenZeroDuration(t *testing.T) {
	if animDisabled {
		t.Skip("TUIKIT_NO_ANIM=1 set")
	}
	tw := &Tween{Duration: 0}
	tw.Start(t0)
	if p := tw.Progress(t0); p != 1 {
		t.Errorf("zero-duration tween = %v, want 1", p)
	}
}

func TestTweenRunning(t *testing.T) {
	if animDisabled {
		t.Skip("TUIKIT_NO_ANIM=1 set, skipping running test")
	}
	tw := &Tween{Duration: 200 * time.Millisecond}
	if tw.Running() {
		t.Error("tween should not be running before Start")
	}
	tw.Start(t0)
	if !tw.Running() {
		t.Error("tween should be running after Start")
	}
	tw.Progress(tAt(200)) // advance to completion
	if tw.Running() {
		t.Error("tween should stop running after completion")
	}
}

func TestTweenDisabledSnapToEnd(t *testing.T) {
	if !animDisabled {
		t.Skip("TUIKIT_NO_ANIM not set, skipping disabled-snap test")
	}
	tw := &Tween{Duration: 500 * time.Millisecond}
	tw.Start(t0)
	if tw.Running() {
		t.Error("disabled: tween should not be running after Start")
	}
	if p := tw.Progress(t0); p != 1 {
		t.Errorf("disabled: progress = %v, want 1", p)
	}
}

// --- hex helpers ---

func TestParseHexColor(t *testing.T) {
	r, g, b := parseHexColor("#ff8000")
	if r != 255 || g != 128 || b != 0 {
		t.Errorf("parseHexColor = %d %d %d, want 255 128 0", r, g, b)
	}
	r2, _, _ := parseHexColor("invalid")
	if r2 != -1 {
		t.Error("parseHexColor invalid should return -1")
	}
}

func TestRgbToHex(t *testing.T) {
	if got := rgbToHex(255, 128, 0); got != "#ff8000" {
		t.Errorf("rgbToHex = %q, want #ff8000", got)
	}
	if got := rgbToHex(-1, 300, 0); got != "#00ff00" {
		t.Errorf("rgbToHex clamped = %q, want #00ff00", got)
	}
}
