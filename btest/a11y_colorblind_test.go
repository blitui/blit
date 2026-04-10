package btest

import (
	"math"
	"testing"
)

func TestSimulateColorBlind_BlackUnchanged(t *testing.T) {
	// Black should remain black under all simulations.
	black := RGB{0, 0, 0}
	for _, typ := range []ColorBlindType{Protanopia, Deuteranopia, Tritanopia} {
		got := SimulateColorBlind(black, typ)
		if got != black {
			t.Errorf("%s: black → %v, want {0,0,0}", typ, got)
		}
	}
}

func TestSimulateColorBlind_WhiteNearWhite(t *testing.T) {
	// White should remain close to white (matrices are not exactly identity).
	white := RGB{255, 255, 255}
	for _, typ := range []ColorBlindType{Protanopia, Deuteranopia, Tritanopia} {
		got := SimulateColorBlind(white, typ)
		dist := ColorDistance(white, got)
		if dist > 0.1 {
			t.Errorf("%s: white → %v (distance %.3f), expected near-white", typ, got, dist)
		}
	}
}

func TestSimulateColorBlind_RedGreenMerge(t *testing.T) {
	// Under protanopia, red and green should become less distinguishable.
	red := RGB{255, 0, 0}
	green := RGB{0, 255, 0}

	normalDist := ColorDistance(red, green)
	simRed := SimulateColorBlind(red, Protanopia)
	simGreen := SimulateColorBlind(green, Protanopia)
	protDist := ColorDistance(simRed, simGreen)

	if protDist >= normalDist {
		t.Errorf("protanopia should reduce red-green distance: normal=%.3f, simulated=%.3f",
			normalDist, protDist)
	}
}

func TestSimulateColorBlind_Deuteranopia(t *testing.T) {
	red := RGB{255, 0, 0}
	green := RGB{0, 255, 0}

	normalDist := ColorDistance(red, green)
	simRed := SimulateColorBlind(red, Deuteranopia)
	simGreen := SimulateColorBlind(green, Deuteranopia)
	deutDist := ColorDistance(simRed, simGreen)

	if deutDist >= normalDist {
		t.Errorf("deuteranopia should reduce red-green distance: normal=%.3f, simulated=%.3f",
			normalDist, deutDist)
	}
}

func TestColorDistance_Identical(t *testing.T) {
	c := RGB{128, 64, 200}
	dist := ColorDistance(c, c)
	if dist != 0 {
		t.Errorf("distance to self = %.3f, want 0", dist)
	}
}

func TestColorDistance_BlackWhite(t *testing.T) {
	dist := ColorDistance(RGB{0, 0, 0}, RGB{255, 255, 255})
	// Should be ~1.73 (sqrt(3) in linear space).
	if math.Abs(dist-math.Sqrt(3)) > 0.01 {
		t.Errorf("black-white distance = %.3f, want ~%.3f", dist, math.Sqrt(3))
	}
}

func TestColorPairDistinguishable(t *testing.T) {
	// Black and white are always distinguishable.
	ok := ColorPairDistinguishable(RGB{0, 0, 0}, RGB{255, 255, 255}, Protanopia, 0.1)
	if !ok {
		t.Error("black and white should be distinguishable under protanopia")
	}
}

func TestColorPairDistinguishable_RedGreen(t *testing.T) {
	// Red and certain shades of green may become indistinguishable under protanopia
	// with a high threshold.
	red := RGB{200, 50, 50}
	green := RGB{50, 200, 50}
	// With a very high threshold, they should fail.
	ok := ColorPairDistinguishable(red, green, Protanopia, 2.0)
	if ok {
		t.Error("red and green should not be distinguishable at threshold 2.0 under protanopia")
	}
}

func TestCheckPalette(t *testing.T) {
	colors := []RGB{
		{255, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
	}
	report := CheckPalette(colors, 0.05)
	if len(report.Pairs) == 0 {
		t.Error("report should have pairs")
	}
	// 3 colors = 3 pairs × 3 types = 9 total checks.
	if len(report.Pairs) != 9 {
		t.Errorf("pairs = %d, want 9", len(report.Pairs))
	}
}

func TestCheckPalette_Violations(t *testing.T) {
	// Two similar colors that merge under simulation.
	colors := []RGB{
		{200, 50, 50},
		{50, 200, 50},
	}
	report := CheckPalette(colors, 0.5)
	violations := report.Violations()
	// At least protanopia and deuteranopia should have violations.
	if len(violations) == 0 {
		t.Error("expected at least one violation for red/green palette at threshold 0.5")
	}
}

func TestCheckPalette_Summary(t *testing.T) {
	colors := []RGB{{255, 0, 0}, {0, 255, 0}}
	report := CheckPalette(colors, 0.1)
	summary := report.Summary()
	if summary == "" {
		t.Error("summary should not be empty")
	}
}

func TestColorBlindType_String(t *testing.T) {
	cases := []struct {
		typ  ColorBlindType
		want string
	}{
		{Protanopia, "protanopia"},
		{Deuteranopia, "deuteranopia"},
		{Tritanopia, "tritanopia"},
	}
	for _, tc := range cases {
		if tc.typ.String() != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.typ, tc.typ.String(), tc.want)
		}
	}
}

func TestDelinearize_RoundTrip(t *testing.T) {
	// linearize → delinearize should approximately round-trip.
	for v := 0; v <= 255; v++ {
		input := float64(v) / 255.0
		lin := linearize(input)
		got := delinearize(lin)
		if abs(int(got)-v) > 1 {
			t.Errorf("round-trip(%d) = %d", v, got)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
