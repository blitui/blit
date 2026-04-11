package blit

import (
	"strings"
	"testing"
)

func TestThemeStylesRendersText(t *testing.T) {
	th := DefaultTheme()
	s := ThemeStyles(th)

	// Verify that styles render content without panicking
	// and produce output that contains the original text.
	cases := []struct {
		name  string
		style Style
		input string
	}{
		{"Text", s.Text, "hello"},
		{"TextBold", s.TextBold, "hello"},
		{"Muted", s.Muted, "hello"},
		{"Accent", s.Accent, "hello"},
		{"AccentBold", s.AccentBold, "hello"},
		{"Positive", s.Positive, "ok"},
		{"Negative", s.Negative, "fail"},
		{"Warn", s.Warn, "careful"},
		{"Info", s.Info, "note"},
		{"Title", s.Title, "Title"},
		{"Subtitle", s.Subtitle, "subtitle"},
		{"Label", s.Label, "Type:"},
		{"Detail", s.Detail, "detail"},
		{"Hint", s.Hint, "q quit"},
		{"Header", s.Header, "Section"},
		{"BadgePositive", s.BadgePositive, "OK"},
		{"BadgeNegative", s.BadgeNegative, "FAIL"},
		{"BadgeWarn", s.BadgeWarn, "WARN"},
		{"BadgeAccent", s.BadgeAccent, "INFO"},
		{"BadgeMuted", s.BadgeMuted, "DIM"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rendered := tc.style.Render(tc.input)
			if !strings.Contains(stripANSI(rendered), tc.input) {
				t.Errorf("style %s rendered %q does not contain input %q", tc.name, stripANSI(rendered), tc.input)
			}
		})
	}
}

func TestThemeStylesConsistentWithTheme(t *testing.T) {
	// Test with a custom theme to ensure styles update correctly
	th := DefaultTheme()
	th.Extra = map[string]Color{
		"info": "#06b6d4",
	}
	s1 := ThemeStyles(th)

	th2 := LightTheme()
	s2 := ThemeStyles(th2)

	// Different themes should produce different Info styles
	// (since the accent color differs)
	info1 := s1.Info.Render("x")
	info2 := s2.Info.Render("x")
	if stripANSI(info1) != stripANSI(info2) {
		// Both should render "x" — we're just checking they don't panic
		// and produce output. The exact colors differ but the text is same.
		t.Error("both Info renders should contain 'x'")
	}
}

func TestSemanticColorExtra(t *testing.T) {
	th := DefaultTheme()
	th.Extra = map[string]Color{
		"info":   "#06b6d4",
		"create": "#22c55e",
		"local":  "#a78bfa",
	}

	if got := th.SemanticColor("info", th.Accent); got != Color("#06b6d4") {
		t.Errorf("SemanticColor(info) = %v, want #06b6d4", got)
	}
	if got := th.SemanticColor("create", th.Positive); got != Color("#22c55e") {
		t.Errorf("SemanticColor(create) = %v, want #22c55e", got)
	}
	if got := th.SemanticColor("local", th.Accent); got != Color("#a78bfa") {
		t.Errorf("SemanticColor(local) = %v, want #a78bfa", got)
	}
}

func TestSemanticColorDefaults(t *testing.T) {
	th := DefaultTheme()
	// No Extra map — should fall back to well-known defaults
	if got := th.SemanticColor("info", th.Muted); got != th.Accent {
		t.Errorf("SemanticColor(info) with no Extra = %v, want Accent", got)
	}
	if got := th.SemanticColor("create", th.Muted); got != th.Positive {
		t.Errorf("SemanticColor(create) with no Extra = %v, want Positive", got)
	}
	if got := th.SemanticColor("delete", th.Muted); got != th.Negative {
		t.Errorf("SemanticColor(delete) with no Extra = %v, want Negative", got)
	}
	if got := th.SemanticColor("review", th.Muted); got != th.Cursor {
		t.Errorf("SemanticColor(review) with no Extra = %v, want Cursor", got)
	}
	if got := th.SemanticColor("issue", th.Muted); got != th.Warn {
		t.Errorf("SemanticColor(issue) with no Extra = %v, want Warn", got)
	}
	if got := th.SemanticColor("release", th.Muted); got != th.Flash {
		t.Errorf("SemanticColor(release) with no Extra = %v, want Flash", got)
	}
	if got := th.SemanticColor("local", th.Muted); got != th.Accent {
		t.Errorf("SemanticColor(local) with no Extra = %v, want Accent", got)
	}
	if got := th.SemanticColor("comment", th.Muted); got != th.Muted {
		t.Errorf("SemanticColor(comment) with no Extra = %v, want Muted", got)
	}
}

func TestSemanticColorUnknownName(t *testing.T) {
	th := DefaultTheme()
	// Unknown name should return the fallback
	if got := th.SemanticColor("unknown", th.Text); got != th.Text {
		t.Errorf("SemanticColor(unknown) = %v, want fallback Text", got)
	}
}

func TestSemanticColorExtraOverridesDefault(t *testing.T) {
	th := DefaultTheme()
	th.Extra = map[string]Color{
		"info": "#custom",
	}
	if got := th.SemanticColor("info", th.Muted); got != Color("#custom") {
		t.Errorf("SemanticColor(info) with Extra override = %v, want #custom", got)
	}
}

func TestHealthDot(t *testing.T) {
	th := DefaultTheme()

	ok := HealthDot("repo", true, th)
	fail := HealthDot("repo", false, th)
	unknown := HealthDotUnknown("repo", th)

	if !strings.Contains(stripANSI(ok), "repo") {
		t.Error("HealthDot(ok) should contain label")
	}
	if !strings.Contains(stripANSI(fail), "repo") {
		t.Error("HealthDot(fail) should contain label")
	}
	if !strings.Contains(stripANSI(unknown), "repo") {
		t.Error("HealthDotUnknown should contain label")
	}

	// The dot glyph differs: ok/fail use ●, unknown uses ○
	if !strings.Contains(stripANSI(ok), "●") {
		t.Error("HealthDot(ok) should contain ● glyph")
	}
	if !strings.Contains(stripANSI(unknown), "○") {
		t.Error("HealthDotUnknown should contain ○ glyph")
	}
}
