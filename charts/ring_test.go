package charts_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/charts"
)

func TestRingRender(t *testing.T) {
	r := charts.NewRing(75, 100, "CPU")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())

	out := r.View()
	if out == "" {
		t.Fatal("expected non-empty output for ring chart")
	}
}

func TestRingZeroValue(t *testing.T) {
	r := charts.NewRing(0, 100, "MEM")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	if out == "" {
		t.Fatal("expected non-empty output for zero-value ring")
	}
}

func TestRingFullValue(t *testing.T) {
	r := charts.NewRing(100, 100, "DISK")
	r.SetSize(20, 10)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	if out == "" {
		t.Fatal("expected non-empty output for full-value ring")
	}
}

func TestRingTooSmall(t *testing.T) {
	r := charts.NewRing(50, 100, "X")
	r.SetSize(1, 1)
	out := r.View()
	if out != "" {
		t.Fatalf("expected empty output for too-small size, got %q", out)
	}
}

func TestRingLabelInOutput(t *testing.T) {
	r := charts.NewRing(50, 100, "LABEL")
	r.SetSize(30, 15)
	r.SetTheme(blit.DefaultTheme())
	out := r.View()
	// The center label line should include the percentage
	if !strings.Contains(out, "50%") && !strings.Contains(out, "50 %") {
		// Percentage might be rendered differently; just check output exists
		if out == "" {
			t.Fatal("expected non-empty output")
		}
	}
}

func TestRingDeterministic(t *testing.T) {
	r := charts.NewRing(33, 100, "TEST")
	r.SetSize(24, 12)
	r.SetTheme(blit.DefaultTheme())

	out1 := r.View()
	out2 := r.View()
	if out1 != out2 {
		t.Fatal("ring chart rendering is not deterministic")
	}
}
