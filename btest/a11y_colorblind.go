package btest

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// ColorBlindType represents a type of color vision deficiency.
type ColorBlindType int

const (
	// Protanopia is red-blind (missing L-cones).
	Protanopia ColorBlindType = iota
	// Deuteranopia is green-blind (missing M-cones).
	Deuteranopia
	// Tritanopia is blue-blind (missing S-cones).
	Tritanopia
)

// String returns the name of the color blindness type.
func (t ColorBlindType) String() string {
	switch t {
	case Protanopia:
		return "protanopia"
	case Deuteranopia:
		return "deuteranopia"
	case Tritanopia:
		return "tritanopia"
	default:
		return "unknown"
	}
}

// SimulateColorBlind transforms an RGB color to simulate how it appears
// to a person with the specified color vision deficiency. Uses the
// Brettel/Viénot/Mollon simulation matrices.
func SimulateColorBlind(c RGB, typ ColorBlindType) RGB {
	// Convert sRGB to linear RGB.
	r := linearize(float64(c.R) / 255.0)
	g := linearize(float64(c.G) / 255.0)
	b := linearize(float64(c.B) / 255.0)

	// Apply the simulation matrix.
	m := simMatrix(typ)
	sr := m[0]*r + m[1]*g + m[2]*b
	sg := m[3]*r + m[4]*g + m[5]*b
	sb := m[6]*r + m[7]*g + m[8]*b

	// Clamp and convert back to sRGB.
	return RGB{
		R: delinearize(clamp01(sr)),
		G: delinearize(clamp01(sg)),
		B: delinearize(clamp01(sb)),
	}
}

// ColorDistance computes the Euclidean distance between two colors in
// linear RGB space. Values range from 0 (identical) to ~1.73 (max).
func ColorDistance(a, b RGB) float64 {
	ar := linearize(float64(a.R) / 255.0)
	ag := linearize(float64(a.G) / 255.0)
	ab := linearize(float64(a.B) / 255.0)
	br := linearize(float64(b.R) / 255.0)
	bg := linearize(float64(b.G) / 255.0)
	bb := linearize(float64(b.B) / 255.0)
	dr := ar - br
	dg := ag - bg
	db := ab - bb
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// ColorPairDistinguishable checks whether two colors remain visually
// distinct under the given color blindness simulation. The threshold
// is the minimum color distance required (0.1 is a reasonable default).
func ColorPairDistinguishable(a, b RGB, typ ColorBlindType, threshold float64) bool {
	sa := SimulateColorBlind(a, typ)
	sb := SimulateColorBlind(b, typ)
	return ColorDistance(sa, sb) >= threshold
}

// AssertDistinguishable fails the test if the two colors are not
// distinguishable under the specified color blindness type.
func AssertDistinguishable(t testing.TB, a, b RGB, typ ColorBlindType) {
	t.Helper()
	const defaultThreshold = 0.1
	if !ColorPairDistinguishable(a, b, typ, defaultThreshold) {
		sa := SimulateColorBlind(a, typ)
		sb := SimulateColorBlind(b, typ)
		dist := ColorDistance(sa, sb)
		t.Errorf("%s: colors %s and %s are not distinguishable (simulated: %s vs %s, distance: %.3f, threshold: %.3f)",
			typ, a, b, sa, sb, dist, defaultThreshold)
	}
}

// AssertAllDistinguishable fails the test if any pair of colors in the
// palette is not distinguishable under the specified color blindness type.
func AssertAllDistinguishable(t testing.TB, colors []RGB, typ ColorBlindType) {
	t.Helper()
	const defaultThreshold = 0.1
	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			if !ColorPairDistinguishable(colors[i], colors[j], typ, defaultThreshold) {
				sa := SimulateColorBlind(colors[i], typ)
				sb := SimulateColorBlind(colors[j], typ)
				dist := ColorDistance(sa, sb)
				t.Errorf("%s: colors[%d]=%s and colors[%d]=%s not distinguishable (simulated: %s vs %s, distance: %.3f)",
					typ, i, colors[i], j, colors[j], sa, sb, dist)
			}
		}
	}
}

// ColorBlindReport holds the results of checking a palette against
// all three types of color vision deficiency.
type ColorBlindReport struct {
	Pairs []ColorBlindPairResult
}

// ColorBlindPairResult holds the result for a single color pair check.
type ColorBlindPairResult struct {
	A, B      RGB
	Type      ColorBlindType
	SimA      RGB
	SimB      RGB
	Distance  float64
	Threshold float64
	Pass      bool
}

// CheckPalette tests all pairs of colors against all three color blindness
// types and returns a detailed report.
func CheckPalette(colors []RGB, threshold float64) *ColorBlindReport {
	types := []ColorBlindType{Protanopia, Deuteranopia, Tritanopia}
	var pairs []ColorBlindPairResult
	for _, typ := range types {
		for i := 0; i < len(colors); i++ {
			for j := i + 1; j < len(colors); j++ {
				sa := SimulateColorBlind(colors[i], typ)
				sb := SimulateColorBlind(colors[j], typ)
				dist := ColorDistance(sa, sb)
				pairs = append(pairs, ColorBlindPairResult{
					A: colors[i], B: colors[j],
					Type: typ, SimA: sa, SimB: sb,
					Distance: dist, Threshold: threshold,
					Pass: dist >= threshold,
				})
			}
		}
	}
	return &ColorBlindReport{Pairs: pairs}
}

// Violations returns only the failing pair results.
func (r *ColorBlindReport) Violations() []ColorBlindPairResult {
	var out []ColorBlindPairResult
	for _, p := range r.Pairs {
		if !p.Pass {
			out = append(out, p)
		}
	}
	return out
}

// Summary returns a human-readable summary.
func (r *ColorBlindReport) Summary() string {
	violations := r.Violations()
	var sb strings.Builder
	fmt.Fprintf(&sb, "Color blindness check: %d pairs tested, %d violations\n",
		len(r.Pairs), len(violations))
	for _, v := range violations {
		fmt.Fprintf(&sb, "  %s: %s vs %s → distance %.3f (need ≥ %.3f)\n",
			v.Type, v.A, v.B, v.Distance, v.Threshold)
	}
	return sb.String()
}

// simMatrix returns the 3×3 simulation matrix (row-major) for the given
// color vision deficiency. These are the Viénot/Brettel/Mollon matrices
// widely used in accessibility tools.
func simMatrix(typ ColorBlindType) [9]float64 {
	switch typ {
	case Protanopia:
		return [9]float64{
			0.56667, 0.43333, 0.00000,
			0.55833, 0.44167, 0.00000,
			0.00000, 0.24167, 0.75833,
		}
	case Deuteranopia:
		return [9]float64{
			0.62500, 0.37500, 0.00000,
			0.70000, 0.30000, 0.00000,
			0.00000, 0.30000, 0.70000,
		}
	case Tritanopia:
		return [9]float64{
			0.95000, 0.05000, 0.00000,
			0.00000, 0.43333, 0.56667,
			0.00000, 0.47500, 0.52500,
		}
	default:
		// Identity matrix (no transformation).
		return [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}
	}
}

// delinearize converts a linear RGB channel value (0–1) back to sRGB (0–255).
func delinearize(v float64) uint8 {
	var s float64
	if v <= 0.0031308 {
		s = v * 12.92
	} else {
		s = 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	return uint8(math.Round(clamp01(s) * 255))
}

// clamp01 clamps v to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
