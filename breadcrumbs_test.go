package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
)

func newTestBreadcrumbs(segs []string, maxWidth int) *blit.Breadcrumbs {
	b := blit.NewBreadcrumbs(segs)
	b.MaxWidth = maxWidth
	b.SetTheme(blit.DefaultTheme())
	b.SetSize(maxWidth, 1)
	return b
}

func TestBreadcrumbsComponentInterface(t *testing.T) {
	var _ blit.Component = blit.NewBreadcrumbs(nil)
}

func TestBreadcrumbsThemedInterface(t *testing.T) {
	var _ blit.Themed = blit.NewBreadcrumbs(nil)
}

func TestBreadcrumbsRenderAllSegments(t *testing.T) {
	b := newTestBreadcrumbs([]string{"home", "docs", "api"}, 0)
	view := b.View()
	if !strings.Contains(view, "home") {
		t.Errorf("expected 'home' in view, got: %q", view)
	}
	if !strings.Contains(view, "docs") {
		t.Errorf("expected 'docs' in view, got: %q", view)
	}
	if !strings.Contains(view, "api") {
		t.Errorf("expected 'api' in view, got: %q", view)
	}
}

func TestBreadcrumbsCustomSeparator(t *testing.T) {
	b := blit.NewBreadcrumbs([]string{"a", "b"})
	b.Separator = " > "
	b.MaxWidth = 0
	b.SetTheme(blit.DefaultTheme())
	view := b.View()
	if !strings.Contains(view, ">") {
		t.Errorf("expected custom separator '>' in view, got: %q", view)
	}
}

func TestBreadcrumbsTruncationEllipsis(t *testing.T) {
	// Each segment is 10 chars; separator is 3; 5 segs = ~10*5 + 3*4 = 62 chars.
	// Set MaxWidth=20 to force truncation.
	segs := []string{"segment-one", "segment-two", "segment-three", "segment-four", "segment-five"}
	b := newTestBreadcrumbs(segs, 20)
	view := b.View()
	if !strings.Contains(view, "…") {
		t.Errorf("expected ellipsis '…' in truncated breadcrumbs, got: %q", view)
	}
}

func TestBreadcrumbsTruncationLastSegmentPreserved(t *testing.T) {
	segs := []string{"very-long-segment-one", "very-long-segment-two", "leaf"}
	b := newTestBreadcrumbs(segs, 15)
	view := b.View()
	// The last segment "leaf" must always be shown.
	if !strings.Contains(view, "leaf") {
		t.Errorf("expected last segment 'leaf' preserved in view, got: %q", view)
	}
}

func TestBreadcrumbsEmptySegments(t *testing.T) {
	b := newTestBreadcrumbs(nil, 80)
	view := b.View()
	// Should not panic; view may be empty.
	_ = view
}

func TestBreadcrumbsSingleSegment(t *testing.T) {
	b := newTestBreadcrumbs([]string{"only"}, 80)
	view := b.View()
	if !strings.Contains(view, "only") {
		t.Errorf("expected 'only' in single-segment view, got: %q", view)
	}
}

func TestBreadcrumbsNoTruncationWhenFits(t *testing.T) {
	segs := []string{"a", "b", "c"}
	b := newTestBreadcrumbs(segs, 200)
	view := b.View()
	if strings.Contains(view, "…") {
		t.Errorf("did not expect ellipsis when content fits, got: %q", view)
	}
}

func TestBreadcrumbsDefaultSeparator(t *testing.T) {
	b := blit.NewBreadcrumbs([]string{"x", "y"})
	b.SetTheme(blit.DefaultTheme())
	view := b.View()
	if !strings.Contains(view, "/") {
		t.Errorf("expected default separator '/' in view, got: %q", view)
	}
}
