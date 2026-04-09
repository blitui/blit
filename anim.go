package tuikit

import (
	"math"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// animDisabled reports whether animations are suppressed via TUIKIT_NO_ANIM=1.
var animDisabled = os.Getenv("TUIKIT_NO_ANIM") == "1"

// Ease is an easing function mapping t in [0,1] to value in [0,1].
type Ease func(t float64) float64

// Linear easing - constant rate.
func Linear(t float64) float64 { return t }

// EaseInOut - slow start and end, fast middle.
func EaseInOut(t float64) float64 { return t * t * (3 - 2*t) }

// EaseOutCubic - fast start, slow end.
func EaseOutCubic(t float64) float64 { t = t - 1; return 1 + t*t*t }

// EaseInCubic - slow start, fast end.
func EaseInCubic(t float64) float64 { return t * t * t }

// EaseOutExpo - exponential deceleration.
func EaseOutExpo(t float64) float64 {
	if t >= 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// Interpolate interpolates between from and to using easing e at progress t in [0,1].
// Supported types: float64, int, lipgloss.Color.
func Interpolate[T any](from, to T, t float64, e Ease) T {
	if e != nil {
		t = e(t)
	}
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	switch f := any(from).(type) {
	case float64:
		return any(f + (any(to).(float64)-f)*t).(T)
	case int:
		return any(int(math.Round(float64(f) + (float64(any(to).(int))-float64(f))*t))).(T)
	case lipgloss.Color:
		return any(lipgloss.Color(interpolateColor(string(f), string(any(to).(lipgloss.Color)), t))).(T)
	}
	if t >= 0.5 {
		return to
	}
	return from
}

func interpolateColor(from, to string, t float64) string {
	fr, fg, fb := parseHexColor(from)
	tr, tg, tb := parseHexColor(to)
	if fr < 0 || tr < 0 {
		if t >= 0.5 {
			return to
		}
		return from
	}
	r := int(math.Round(float64(fr) + (float64(tr)-float64(fr))*t))
	g := int(math.Round(float64(fg) + (float64(tg)-float64(fg))*t))
	b := int(math.Round(float64(fb) + (float64(tb)-float64(fb))*t))
	return rgbToHex(r, g, b)
}

func parseHexColor(s string) (int, int, int) {
	if len(s) == 7 && s[0] == '#' {
		var r, g, b int
		_, err := scanHex(s[1:3], s[3:5], s[5:7], &r, &g, &b)
		if err {
			return -1, -1, -1
		}
		return r, g, b
	}
	return -1, -1, -1
}

func scanHex(rs, gs, bs string, r, g, b *int) (int, bool) {
	*r = hexPair(rs)
	*g = hexPair(gs)
	*b = hexPair(bs)
	if *r < 0 || *g < 0 || *b < 0 {
		return 0, true
	}
	return 0, false
}

func hexPair(s string) int {
	if len(s) != 2 {
		return -1
	}
	hi := hexDigit(s[0])
	lo := hexDigit(s[1])
	if hi < 0 || lo < 0 {
		return -1
	}
	return hi*16 + lo
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func rgbToHex(r, g, b int) string {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	return "#" + byteHex(clamp(r)) + byteHex(clamp(g)) + byteHex(clamp(b))
}

func byteHex(v int) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[v>>4], digits[v&0xf]})
}

// Tween tracks animation progress over a fixed duration.
// When TUIKIT_NO_ANIM=1, Start is a no-op and Progress always returns 1.
type Tween struct {
	Duration  time.Duration
	startedAt time.Time
	running   bool
}

// Start begins the tween at the given clock time.
func (tw *Tween) Start(now time.Time) {
	if animDisabled {
		tw.running = false
		return
	}
	tw.startedAt = now
	tw.running = true
}

// Progress returns t in [0,1] at the given clock time. Returns 1 when done.
func (tw *Tween) Progress(now time.Time) float64 {
	if !tw.running || animDisabled {
		return 1
	}
	if tw.Duration <= 0 {
		return 1
	}
	elapsed := now.Sub(tw.startedAt)
	t := float64(elapsed) / float64(tw.Duration)
	if t >= 1 {
		t = 1
		tw.running = false
	}
	return t
}

// Done reports whether the tween has completed.
func (tw *Tween) Done() bool { return !tw.running }

// Running reports whether the tween is currently active.
func (tw *Tween) Running() bool { return tw.running }

// animTickMsg is sent by the animation tick bus (~60fps while tweens are active).
type animTickMsg struct {
	time time.Time
}
