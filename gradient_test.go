package blit

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestParseHex(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input      string
		r, g, b    uint8
		wantZeroed bool
	}{
		{"#ff8000", 255, 128, 0, false},
		{"ff8000", 255, 128, 0, false},
		{"#000000", 0, 0, 0, false},
		{"#ffffff", 255, 255, 255, false},
		{"", 0, 0, 0, true},
		{"#fff", 0, 0, 0, true},
		{"#zzzzzz", 0, 0, 0, true},
		{"short", 0, 0, 0, true},
	}
	for _, tc := range cases {
		r, g, b := parseHex(tc.input)
		if r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("parseHex(%q) = (%d,%d,%d), want (%d,%d,%d)", tc.input, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestLerpU8(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b uint8
		t    float64
		want uint8
	}{
		{0, 255, 0, 0},
		{0, 255, 1, 255},
		{0, 255, 0.5, 128},
		{100, 200, 0.25, 125},
		{0, 0, 0.5, 0},
		{255, 255, 0.5, 255},
	}
	for _, tc := range cases {
		got := lerpU8(tc.a, tc.b, tc.t)
		if got != tc.want {
			t.Errorf("lerpU8(%d,%d,%.2f) = %d, want %d", tc.a, tc.b, tc.t, got, tc.want)
		}
	}
}

func TestGradientRenderAt(t *testing.T) {
	t.Parallel()
	g := Gradient{Start: lipgloss.Color("#000000"), End: lipgloss.Color("#ffffff")}

	cases := []struct {
		t    float64
		want string
	}{
		{0, "#000000"},
		{1, "#ffffff"},
		{0.5, "#808080"},
		{-1, "#000000"}, // clamped to 0
		{2, "#ffffff"},  // clamped to 1
	}
	for _, tc := range cases {
		got := string(g.RenderAt(tc.t))
		if got != tc.want {
			t.Errorf("RenderAt(%.1f) = %q, want %q", tc.t, got, tc.want)
		}
	}
}

func TestGradientRenderText(t *testing.T) {
	t.Parallel()
	g := Gradient{Start: lipgloss.Color("#ff0000"), End: lipgloss.Color("#0000ff")}

	// Empty string returns empty
	if got := g.RenderText(""); got != "" {
		t.Errorf("RenderText(\"\") = %q, want \"\"", got)
	}

	// Single char uses t=0 (start color)
	got := g.RenderText("A")
	if got == "" {
		t.Error("RenderText(\"A\") returned empty string")
	}

	// Multi-char produces non-empty output containing original text
	got = g.RenderText("Hello")
	if got == "" {
		t.Error("RenderText(\"Hello\") returned empty string")
	}
}

func TestRenderGradient(t *testing.T) {
	t.Parallel()
	g := Gradient{Start: lipgloss.Color("#ff0000"), End: lipgloss.Color("#0000ff")}
	got := RenderGradient("test", g)
	if got == "" {
		t.Error("RenderGradient returned empty string")
	}
}
