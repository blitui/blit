package btest

import (
	"math"
	"testing"
)

func TestContrastRatio_BlackOnWhite(t *testing.T) {
	ratio := ContrastRatio(RGB{0, 0, 0}, RGB{255, 255, 255})
	// Black on white should be 21:1.
	if math.Abs(ratio-21.0) > 0.1 {
		t.Errorf("black on white = %.2f, want ~21.0", ratio)
	}
}

func TestContrastRatio_WhiteOnBlack(t *testing.T) {
	ratio := ContrastRatio(RGB{255, 255, 255}, RGB{0, 0, 0})
	// Order shouldn't matter — lighter is always numerator.
	if math.Abs(ratio-21.0) > 0.1 {
		t.Errorf("white on black = %.2f, want ~21.0", ratio)
	}
}

func TestContrastRatio_SameColor(t *testing.T) {
	ratio := ContrastRatio(RGB{128, 128, 128}, RGB{128, 128, 128})
	if math.Abs(ratio-1.0) > 0.01 {
		t.Errorf("same color = %.2f, want 1.0", ratio)
	}
}

func TestContrastRatio_LowContrast(t *testing.T) {
	// Light gray on white — low contrast.
	ratio := ContrastRatio(RGB{200, 200, 200}, RGB{255, 255, 255})
	if ratio >= 4.5 {
		t.Errorf("light gray on white = %.2f, should be < 4.5", ratio)
	}
}

func TestCheckContrast_AA(t *testing.T) {
	// Black on white passes everything.
	r := CheckContrast(RGB{0, 0, 0}, RGB{255, 255, 255})
	if !r.PassAA {
		t.Error("black on white should pass AA")
	}
	if !r.PassAAA {
		t.Error("black on white should pass AAA")
	}
	if !r.PassAALarge {
		t.Error("black on white should pass AA large")
	}
	if !r.PassAAALarge {
		t.Error("black on white should pass AAA large")
	}
}

func TestCheckContrast_MediumContrast(t *testing.T) {
	// Dark gray on white — should pass AA but maybe not AAA.
	r := CheckContrast(RGB{90, 90, 90}, RGB{255, 255, 255})
	if !r.PassAA {
		t.Errorf("dark gray on white = %.2f, should pass AA (≥ 4.5)", r.Ratio)
	}
	if !r.PassAALarge {
		t.Errorf("dark gray on white = %.2f, should pass AA large (≥ 3.0)", r.Ratio)
	}
}

func TestCheckContrast_FailAA(t *testing.T) {
	// Light gray on white — fails AA.
	r := CheckContrast(RGB{200, 200, 200}, RGB{255, 255, 255})
	if r.PassAA {
		t.Errorf("light gray on white = %.2f, should fail AA", r.Ratio)
	}
}

func TestAssertContrast_Pass(t *testing.T) {
	// Should not fail the test.
	AssertContrast(t, RGB{0, 0, 0}, RGB{255, 255, 255}, WCAGLevelAA)
}

func TestAssertContrastLarge_Pass(t *testing.T) {
	AssertContrastLarge(t, RGB{0, 0, 0}, RGB{255, 255, 255}, WCAGLevelAA)
}

func TestANSI256ToRGB_StandardColors(t *testing.T) {
	cases := []struct {
		idx  uint8
		want RGB
	}{
		{0, RGB{0, 0, 0}},
		{1, RGB{128, 0, 0}},
		{7, RGB{192, 192, 192}},
		{15, RGB{255, 255, 255}},
	}
	for _, tc := range cases {
		got := ANSI256ToRGB(tc.idx)
		if got != tc.want {
			t.Errorf("ANSI256ToRGB(%d) = %v, want %v", tc.idx, got, tc.want)
		}
	}
}

func TestANSI256ToRGB_ColorCube(t *testing.T) {
	// Index 16 = cube(0,0,0) = black.
	got := ANSI256ToRGB(16)
	if got != (RGB{0, 0, 0}) {
		t.Errorf("ANSI256ToRGB(16) = %v, want {0,0,0}", got)
	}
	// Index 231 = cube(5,5,5) = brightest.
	got = ANSI256ToRGB(231)
	want := RGB{255, 255, 255}
	if got != want {
		t.Errorf("ANSI256ToRGB(231) = %v, want %v", got, want)
	}
}

func TestANSI256ToRGB_Grayscale(t *testing.T) {
	// Index 232 = first grayscale = 8.
	got := ANSI256ToRGB(232)
	if got != (RGB{8, 8, 8}) {
		t.Errorf("ANSI256ToRGB(232) = %v, want {8,8,8}", got)
	}
	// Index 255 = last grayscale = 238.
	got = ANSI256ToRGB(255)
	if got != (RGB{238, 238, 238}) {
		t.Errorf("ANSI256ToRGB(255) = %v, want {238,238,238}", got)
	}
}

func TestContrastReport_Violations(t *testing.T) {
	report := &ContrastReport{
		Results: []ContrastResult{
			CheckContrast(RGB{0, 0, 0}, RGB{255, 255, 255}),     // pass
			CheckContrast(RGB{200, 200, 200}, RGB{255, 255, 255}), // fail
		},
	}
	violations := report.Violations(WCAGLevelAA)
	if len(violations) != 1 {
		t.Errorf("violations = %d, want 1", len(violations))
	}
}

func TestContrastReport_Summary(t *testing.T) {
	report := &ContrastReport{
		Results: []ContrastResult{
			CheckContrast(RGB{0, 0, 0}, RGB{255, 255, 255}),
			CheckContrast(RGB{200, 200, 200}, RGB{255, 255, 255}),
		},
	}
	summary := report.Summary(WCAGLevelAA)
	if summary == "" {
		t.Error("summary should not be empty")
	}
	if !containsString(summary, "1 violations") {
		t.Errorf("summary should mention 1 violation: %s", summary)
	}
}

func TestWCAGLevel_String(t *testing.T) {
	if WCAGLevelAA.String() != "AA" {
		t.Errorf("AA.String() = %q", WCAGLevelAA.String())
	}
	if WCAGLevelAAA.String() != "AAA" {
		t.Errorf("AAA.String() = %q", WCAGLevelAAA.String())
	}
}

func TestRGB_String(t *testing.T) {
	c := RGB{255, 128, 0}
	want := "#FF8000"
	if c.String() != want {
		t.Errorf("RGB.String() = %q, want %q", c.String(), want)
	}
}

func containsString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
