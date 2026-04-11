package blit

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestInterpolate_NilEasing(t *testing.T) {
	got := Interpolate[float64](0, 100, 0.5, nil)
	if got != 50 {
		t.Errorf("Interpolate with nil easing = %v, want 50", got)
	}
}

func TestInterpolate_ClampNegativeT(t *testing.T) {
	got := Interpolate[float64](0, 100, -0.5, Linear)
	if got != 0 {
		t.Errorf("Interpolate t=-0.5 = %v, want 0", got)
	}
}

func TestInterpolate_ClampOverOneT(t *testing.T) {
	got := Interpolate[float64](0, 100, 1.5, Linear)
	if got != 100 {
		t.Errorf("Interpolate t=1.5 = %v, want 100", got)
	}
}

func TestInterpolate_UnsupportedType_High(t *testing.T) {
	// Unsupported type with t >= 0.5 should return 'to'.
	got := Interpolate[string]("hello", "world", 0.7, Linear)
	if got != "world" {
		t.Errorf("Interpolate string t=0.7 = %q, want %q", got, "world")
	}
}

func TestInterpolate_UnsupportedType_Low(t *testing.T) {
	// Unsupported type with t < 0.5 should return 'from'.
	got := Interpolate[string]("hello", "world", 0.3, Linear)
	if got != "hello" {
		t.Errorf("Interpolate string t=0.3 = %q, want %q", got, "hello")
	}
}

func TestInterpolateColor_InvalidFrom(t *testing.T) {
	// Invalid 'from' color: t >= 0.5 should return 'to'.
	got := interpolateColor("bad", "#ff0000", 0.6)
	if got != "#ff0000" {
		t.Errorf("interpolateColor bad from t=0.6 = %q, want #ff0000", got)
	}
}

func TestInterpolateColor_InvalidTo(t *testing.T) {
	// Invalid 'to' color: t < 0.5 should return 'from'.
	got := interpolateColor("#ff0000", "bad", 0.3)
	if got != "#ff0000" {
		t.Errorf("interpolateColor bad to t=0.3 = %q, want #ff0000", got)
	}
}

func TestInterpolateColor_BothInvalid(t *testing.T) {
	got := interpolateColor("bad", "also-bad", 0.5)
	if got != "also-bad" {
		t.Errorf("interpolateColor both invalid t=0.5 = %q, want also-bad", got)
	}
}

func TestInterpolate_IntNilEasing(t *testing.T) {
	got := Interpolate[int](0, 10, 0.5, nil)
	if got != 5 {
		t.Errorf("Interpolate int nil easing = %v, want 5", got)
	}
}

func TestInterpolate_ColorNilEasing(t *testing.T) {
	from := lipgloss.Color("#000000")
	to := lipgloss.Color("#ffffff")
	got := Interpolate[lipgloss.Color](from, to, 0.5, nil)
	if string(got) != "#808080" {
		t.Errorf("Interpolate color nil easing = %q, want #808080", string(got))
	}
}

func TestParseHexColor_TooShort(t *testing.T) {
	r, _, _ := parseHexColor("#fff")
	if r != -1 {
		t.Error("parseHexColor short should return -1")
	}
}

func TestParseHexColor_NoPound(t *testing.T) {
	r, _, _ := parseHexColor("ff8000x")
	if r != -1 {
		t.Error("parseHexColor no # should return -1")
	}
}
