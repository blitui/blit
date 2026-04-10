package charts_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/charts"
)

func TestLineRender(t *testing.T) {
	series := [][]float64{
		{1, 3, 2, 5, 4},
		{5, 4, 3, 2, 1},
	}
	l := charts.NewLine(series, []string{"#ff0000", "#00ff00"}, false)
	l.SetSize(40, 10)
	l.SetTheme(blit.DefaultTheme())

	out := l.View()
	if out == "" {
		t.Fatal("expected non-empty output for line chart")
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines (height), got %d", len(lines))
	}
}

func TestLineSmoothRender(t *testing.T) {
	series := [][]float64{{0, 5, 2, 8, 3}}
	l := charts.NewLine(series, nil, true)
	l.SetSize(30, 8)
	l.SetTheme(blit.DefaultTheme())

	out := l.View()
	if out == "" {
		t.Fatal("expected non-empty output for smooth line chart")
	}
}

func TestLineEmptyData(t *testing.T) {
	l := charts.NewLine(nil, nil, false)
	l.SetSize(30, 8)
	out := l.View()
	if out != "" {
		t.Fatalf("expected empty output for nil series, got %q", out)
	}
}

func TestLineSinglePoint(t *testing.T) {
	series := [][]float64{{42}}
	l := charts.NewLine(series, nil, false)
	l.SetSize(20, 5)
	l.SetTheme(blit.DefaultTheme())
	// Single point series should render without panic
	_ = l.View()
}

func TestLineDeterministic(t *testing.T) {
	series := [][]float64{{1, 4, 2, 8, 5, 7, 1, 3}}
	l := charts.NewLine(series, nil, false)
	l.SetSize(30, 8)
	l.SetTheme(blit.DefaultTheme())

	out1 := l.View()
	out2 := l.View()
	if out1 != out2 {
		t.Fatal("line chart rendering is not deterministic")
	}
}
