package blit_test

import (
	"testing"

	blit "github.com/blitui/blit"
)

func TestDefaultGlyphs_NonEmpty(t *testing.T) {
	t.Parallel()
	g := blit.DefaultGlyphs()
	fields := []struct {
		name  string
		value string
	}{
		{"TreeBranch", g.TreeBranch},
		{"TreeLast", g.TreeLast},
		{"TreePipe", g.TreePipe},
		{"TreeEmpty", g.TreeEmpty},
		{"CursorMarker", g.CursorMarker},
		{"FlashMarker", g.FlashMarker},
		{"SelectedBullet", g.SelectedBullet},
		{"UnselectedBullet", g.UnselectedBullet},
		{"CollapsedArrow", g.CollapsedArrow},
		{"ExpandedArrow", g.ExpandedArrow},
		{"BarFilled", g.BarFilled},
		{"BarEmpty", g.BarEmpty},
		{"Check", g.Check},
		{"Cross", g.Cross},
		{"Info", g.Info},
		{"Warn", g.Warn},
		{"Star", g.Star},
		{"Dot", g.Dot},
	}
	for _, f := range fields {
		if f.value == "" {
			t.Errorf("DefaultGlyphs().%s is empty", f.name)
		}
	}
}

func TestDefaultGlyphs_SpinnerFrames(t *testing.T) {
	t.Parallel()
	g := blit.DefaultGlyphs()
	if len(g.SpinnerFrames) == 0 {
		t.Fatal("DefaultGlyphs().SpinnerFrames should not be empty")
	}
	for i, f := range g.SpinnerFrames {
		if f == "" {
			t.Errorf("SpinnerFrames[%d] is empty", i)
		}
	}
}

func TestAsciiGlyphs_NonEmpty(t *testing.T) {
	t.Parallel()
	g := blit.AsciiGlyphs()
	fields := []struct {
		name  string
		value string
	}{
		{"TreeBranch", g.TreeBranch},
		{"TreeLast", g.TreeLast},
		{"TreePipe", g.TreePipe},
		{"CursorMarker", g.CursorMarker},
		{"SelectedBullet", g.SelectedBullet},
		{"UnselectedBullet", g.UnselectedBullet},
		{"CollapsedArrow", g.CollapsedArrow},
		{"ExpandedArrow", g.ExpandedArrow},
		{"BarFilled", g.BarFilled},
		{"BarEmpty", g.BarEmpty},
		{"Check", g.Check},
		{"Cross", g.Cross},
		{"Info", g.Info},
		{"Warn", g.Warn},
		{"Star", g.Star},
		{"Dot", g.Dot},
	}
	for _, f := range fields {
		if f.value == "" {
			t.Errorf("AsciiGlyphs().%s is empty", f.name)
		}
	}
}

func TestAsciiGlyphs_SpinnerFrames(t *testing.T) {
	t.Parallel()
	g := blit.AsciiGlyphs()
	if len(g.SpinnerFrames) == 0 {
		t.Fatal("AsciiGlyphs().SpinnerFrames should not be empty")
	}
}

func TestAsciiGlyphs_AsciiOnly(t *testing.T) {
	t.Parallel()
	g := blit.AsciiGlyphs()
	check := func(name, val string) {
		for _, r := range val {
			if r > 127 {
				t.Errorf("AsciiGlyphs().%s contains non-ASCII rune %q", name, r)
			}
		}
	}
	check("TreeBranch", g.TreeBranch)
	check("TreeLast", g.TreeLast)
	check("TreePipe", g.TreePipe)
	check("CursorMarker", g.CursorMarker)
	check("Check", g.Check)
	check("Cross", g.Cross)
	for i, f := range g.SpinnerFrames {
		for _, r := range f {
			if r > 127 {
				t.Errorf("AsciiGlyphs().SpinnerFrames[%d] contains non-ASCII rune %q", i, r)
			}
		}
	}
}

func TestGlyphs_DefaultAndAsciiHaveSameFieldCount(t *testing.T) {
	t.Parallel()
	d := blit.DefaultGlyphs()
	a := blit.AsciiGlyphs()

	// Both should have spinner frames.
	if len(d.SpinnerFrames) == 0 || len(a.SpinnerFrames) == 0 {
		t.Fatal("both glyph sets should have spinner frames")
	}
}
