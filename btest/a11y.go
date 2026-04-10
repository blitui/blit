// Package btest accessibility helpers.
//
// This file implements WCAG 2.1 contrast ratio checking for terminal
// color pairs. It converts ANSI 256-color and true-color values to
// relative luminance and computes contrast ratios per the W3C formula.
//
// References:
//   - https://www.w3.org/TR/WCAG21/#dfn-contrast-ratio
//   - https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
package btest

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// WCAGLevel represents a WCAG conformance level for contrast.
type WCAGLevel int

const (
	// WCAGLevelAA requires a contrast ratio of at least 4.5:1 for normal text
	// and 3:1 for large text.
	WCAGLevelAA WCAGLevel = iota
	// WCAGLevelAAA requires a contrast ratio of at least 7:1 for normal text
	// and 4.5:1 for large text.
	WCAGLevelAAA
)

// String returns the level name.
func (l WCAGLevel) String() string {
	switch l {
	case WCAGLevelAA:
		return "AA"
	case WCAGLevelAAA:
		return "AAA"
	default:
		return "unknown"
	}
}

// ContrastResult holds the outcome of a contrast ratio check.
type ContrastResult struct {
	// FG and BG are the foreground and background colors as RGB.
	FG, BG RGB

	// Ratio is the computed contrast ratio (1.0–21.0).
	Ratio float64

	// PassAA is true if the ratio meets WCAG AA for normal text (≥ 4.5).
	PassAA bool

	// PassAAA is true if the ratio meets WCAG AAA for normal text (≥ 7.0).
	PassAAA bool

	// PassAALarge is true if the ratio meets WCAG AA for large text (≥ 3.0).
	PassAALarge bool

	// PassAAALarge is true if the ratio meets WCAG AAA for large text (≥ 4.5).
	PassAAALarge bool
}

// RGB represents a color as 8-bit red, green, blue channels.
type RGB struct {
	R, G, B uint8
}

// String returns the color as a hex string.
func (c RGB) String() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// ContrastRatio computes the WCAG 2.1 contrast ratio between two colors.
// The result is in the range [1.0, 21.0].
func ContrastRatio(fg, bg RGB) float64 {
	l1 := relativeLuminance(fg)
	l2 := relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// CheckContrast evaluates the contrast between fg and bg against WCAG criteria.
func CheckContrast(fg, bg RGB) ContrastResult {
	ratio := ContrastRatio(fg, bg)
	return ContrastResult{
		FG:           fg,
		BG:           bg,
		Ratio:        ratio,
		PassAA:       ratio >= 4.5,
		PassAAA:      ratio >= 7.0,
		PassAALarge:  ratio >= 3.0,
		PassAAALarge: ratio >= 4.5,
	}
}

// AssertContrast fails the test if the contrast ratio between fg and bg
// does not meet the specified WCAG level for normal text.
func AssertContrast(t testing.TB, fg, bg RGB, level WCAGLevel) {
	t.Helper()
	result := CheckContrast(fg, bg)
	var pass bool
	var threshold float64
	switch level {
	case WCAGLevelAA:
		pass = result.PassAA
		threshold = 4.5
	case WCAGLevelAAA:
		pass = result.PassAAA
		threshold = 7.0
	}
	if !pass {
		t.Errorf("contrast ratio %s on %s = %.2f:1, want ≥ %.1f:1 (WCAG %s)",
			fg, bg, result.Ratio, threshold, level)
	}
}

// AssertContrastLarge fails the test if the contrast ratio between fg and bg
// does not meet the specified WCAG level for large text.
func AssertContrastLarge(t testing.TB, fg, bg RGB, level WCAGLevel) {
	t.Helper()
	result := CheckContrast(fg, bg)
	var pass bool
	var threshold float64
	switch level {
	case WCAGLevelAA:
		pass = result.PassAALarge
		threshold = 3.0
	case WCAGLevelAAA:
		pass = result.PassAAALarge
		threshold = 4.5
	}
	if !pass {
		t.Errorf("contrast ratio %s on %s = %.2f:1, want ≥ %.1f:1 (WCAG %s large text)",
			fg, bg, result.Ratio, threshold, level)
	}
}

// ContrastReport holds results of checking multiple color pairs.
type ContrastReport struct {
	Results []ContrastResult
}

// Violations returns only the results that fail the specified level
// for normal text.
func (r *ContrastReport) Violations(level WCAGLevel) []ContrastResult {
	var out []ContrastResult
	for _, res := range r.Results {
		switch level {
		case WCAGLevelAA:
			if !res.PassAA {
				out = append(out, res)
			}
		case WCAGLevelAAA:
			if !res.PassAAA {
				out = append(out, res)
			}
		}
	}
	return out
}

// Summary returns a human-readable summary of the report.
func (r *ContrastReport) Summary(level WCAGLevel) string {
	violations := r.Violations(level)
	var sb strings.Builder
	fmt.Fprintf(&sb, "Contrast check: %d pairs tested, %d violations (WCAG %s)\n",
		len(r.Results), len(violations), level)
	for _, v := range violations {
		fmt.Fprintf(&sb, "  %s on %s → %.2f:1 (FAIL)\n", v.FG, v.BG, v.Ratio)
	}
	return sb.String()
}

// relativeLuminance computes the WCAG 2.1 relative luminance of an sRGB color.
// Formula: L = 0.2126*R + 0.7152*G + 0.0722*B
// where each channel is linearized from sRGB.
func relativeLuminance(c RGB) float64 {
	r := linearize(float64(c.R) / 255.0)
	g := linearize(float64(c.G) / 255.0)
	b := linearize(float64(c.B) / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// linearize converts an sRGB channel value (0–1) to linear RGB.
func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// ANSI256ToRGB converts an ANSI 256-color index to an RGB value.
// Colors 0–15 are the standard terminal palette (implementation-defined;
// we use the xterm defaults). Colors 16–231 are a 6×6×6 color cube.
// Colors 232–255 are a grayscale ramp.
func ANSI256ToRGB(idx uint8) RGB {
	if idx < 16 {
		return ansi16[idx]
	}
	if idx < 232 {
		// 6×6×6 color cube: index = 16 + 36*r + 6*g + b
		i := int(idx) - 16
		b := i % 6
		g := (i / 6) % 6
		r := i / 36
		return RGB{
			R: cubeValue(r),
			G: cubeValue(g),
			B: cubeValue(b),
		}
	}
	// Grayscale ramp: 232–255 → 8, 18, 28, ..., 238
	v := uint8(8 + 10*(int(idx)-232))
	return RGB{R: v, G: v, B: v}
}

// cubeValue maps a 6-level cube index (0–5) to a channel value.
func cubeValue(level int) uint8 {
	if level == 0 {
		return 0
	}
	return uint8(55 + 40*level)
}

// ansi16 contains the default xterm colors for ANSI indices 0–15.
var ansi16 = [16]RGB{
	{0, 0, 0},       // 0: black
	{128, 0, 0},     // 1: red
	{0, 128, 0},     // 2: green
	{128, 128, 0},   // 3: yellow
	{0, 0, 128},     // 4: blue
	{128, 0, 128},   // 5: magenta
	{0, 128, 128},   // 6: cyan
	{192, 192, 192}, // 7: white
	{128, 128, 128}, // 8: bright black
	{255, 0, 0},     // 9: bright red
	{0, 255, 0},     // 10: bright green
	{255, 255, 0},   // 11: bright yellow
	{0, 0, 255},     // 12: bright blue
	{255, 0, 255},   // 13: bright magenta
	{0, 255, 255},   // 14: bright cyan
	{255, 255, 255}, // 15: bright white
}
