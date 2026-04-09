package tuikit_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func testTheme() tuikit.Theme {
	return tuikit.Theme{
		Accent:   lipgloss.Color("#3b82f6"),
		Muted:    lipgloss.Color("#6b7280"),
		Text:     lipgloss.Color("#e5e7eb"),
		Positive: lipgloss.Color("#22c55e"),
		Negative: lipgloss.Color("#ef4444"),
		Flash:    lipgloss.Color("#facc15"),
	}
}

// TestMarkdown_HeadingColor verifies that a rendered heading contains the
// Accent color token from the theme (G4: heading color == theme.Accent).
func TestMarkdown_HeadingColor(t *testing.T) {
	theme := testTheme()
	out := tuikit.Markdown("# Hello World", theme)
	// The rendered output should contain the accent hex color somewhere in the
	// ANSI sequence. We check for the hex value stripped of '#'.
	accentHex := strings.TrimPrefix(string(theme.Accent), "#")
	if !strings.Contains(out, accentHex) {
		t.Errorf("heading output does not contain accent color %q:\n%s", accentHex, out)
	}
}

// TestMarkdown_CodeBlockBackground verifies that a fenced code block uses the
// Muted token as background (G4: code block bg == theme.Muted).
func TestMarkdown_CodeBlockBackground(t *testing.T) {
	theme := testTheme()
	md := "```\nfoo := bar\n```"
	out := tuikit.Markdown(md, theme)
	mutedHex := strings.TrimPrefix(string(theme.Muted), "#")
	if !strings.Contains(out, mutedHex) {
		t.Errorf("code block output does not contain muted color %q:\n%s", mutedHex, out)
	}
}

// TestMarkdown_FallbackOnEmpty ensures Markdown("", theme) returns something
// renderable (glamour returns a blank/whitespace string, not an error panic).
func TestMarkdown_FallbackOnEmpty(t *testing.T) {
	out := tuikit.Markdown("", testTheme())
	// Should not panic; result may be empty or whitespace — just not an error string.
	_ = out
}

// TestMarkdown_PlainText verifies that plain text (no markdown) passes through
// and the output contains the original words.
func TestMarkdown_PlainText(t *testing.T) {
	theme := testTheme()
	out := tuikit.Markdown("hello world", theme)
	if !strings.Contains(out, "hello world") {
		t.Errorf("plain text not preserved in output:\n%s", out)
	}
}

// TestMarkdown_MultipleHeadingLevels ensures H1–H3 all render without panic
// and contain the accent hex.
func TestMarkdown_MultipleHeadingLevels(t *testing.T) {
	theme := testTheme()
	accentHex := strings.TrimPrefix(string(theme.Accent), "#")
	for _, md := range []string{"# H1", "## H2", "### H3"} {
		out := tuikit.Markdown(md, theme)
		if !strings.Contains(out, accentHex) {
			t.Errorf("%q: accent color %q not found in output:\n%s", md, accentHex, out)
		}
	}
}

// TestMarkdown_ReleaseNotesHighlight verifies that BREAKING and SECURITY
// section headings are highlighted in UpdateNotesOverlay using theme colors.
func TestMarkdown_ReleaseNotesHighlight(t *testing.T) {
	theme := testTheme()
	notes := "## BREAKING CHANGE\n\nThis is a breaking change.\n\n## SECURITY FIX\n\nPatched CVE."
	o := tuikit.NewReleaseNotesOverlayThemed("v1.0.0", notes, theme)

	// The rendered lines should contain the negative color for BREAKING and
	// flash color for SECURITY.
	joined := strings.Join(o.Lines, "\n")
	negHex := strings.TrimPrefix(string(theme.Negative), "#")
	flashHex := strings.TrimPrefix(string(theme.Flash), "#")

	if !strings.Contains(joined, negHex) {
		t.Errorf("BREAKING section should contain negative color %q:\n%s", negHex, joined)
	}
	if !strings.Contains(joined, flashHex) {
		t.Errorf("SECURITY section should contain flash color %q:\n%s", flashHex, joined)
	}
}

// TestMarkdown_SetThemeReRendersLines confirms that SetTheme on an overlay
// re-renders lines through Markdown.
func TestMarkdown_SetThemeReRendersLines(t *testing.T) {
	notes := "# Release v1.0\n\nSome notes."
	o := tuikit.NewReleaseNotesOverlay("v1.0.0", notes)

	// Before SetTheme: lines should be plain (no ANSI for accent color).
	theme := testTheme()
	accentHex := strings.TrimPrefix(string(theme.Accent), "#")
	plainJoined := strings.Join(o.Lines, "\n")
	if strings.Contains(plainJoined, accentHex) {
		t.Logf("pre-theme lines already contain accent — that is acceptable")
	}

	o.SetTheme(theme)
	themedJoined := strings.Join(o.Lines, "\n")
	if !strings.Contains(themedJoined, accentHex) {
		t.Errorf("after SetTheme, lines should contain accent color %q:\n%s", accentHex, themedJoined)
	}
}
