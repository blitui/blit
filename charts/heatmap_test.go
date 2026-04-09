package charts_test

import (
	"strings"
	"testing"

	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/charts"
)

func TestHeatmapSequential(t *testing.T) {
	grid := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.SetSize(30, 8)
	h.SetTheme(tuikit.DefaultTheme())

	out := h.View()
	if out == "" {
		t.Fatal("expected non-empty output for sequential heatmap")
	}
}

func TestHeatmapDivergent(t *testing.T) {
	grid := [][]float64{
		{-5, 0, 5},
		{-3, 1, 4},
	}
	h := charts.NewHeatmap(grid, charts.PaletteDivergent)
	h.SetSize(30, 8)
	h.SetTheme(tuikit.DefaultTheme())

	out := h.View()
	if out == "" {
		t.Fatal("expected non-empty output for divergent heatmap")
	}
}

func TestHeatmapWithLabels(t *testing.T) {
	grid := [][]float64{
		{1, 2},
		{3, 4},
	}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.Labels = []string{"Col A", "Col B"}
	h.RowLabels = []string{"Row 1", "Row 2"}
	h.SetSize(30, 8)
	h.SetTheme(tuikit.DefaultTheme())

	out := h.View()
	if !strings.Contains(out, "Col A") {
		t.Error("expected column label 'Col A' in heatmap output")
	}
}

func TestHeatmapEmpty(t *testing.T) {
	h := charts.NewHeatmap(nil, charts.PaletteSequential)
	h.SetSize(20, 8)
	out := h.View()
	if out != "" {
		t.Fatalf("expected empty output for nil grid, got %q", out)
	}
}

func TestHeatmapTooSmall(t *testing.T) {
	grid := [][]float64{{1, 2}, {3, 4}}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.SetSize(0, 0)
	out := h.View()
	if out != "" {
		t.Fatalf("expected empty output for zero size, got %q", out)
	}
}

func TestHeatmapDeterministic(t *testing.T) {
	grid := [][]float64{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 8, 7, 6},
	}
	h := charts.NewHeatmap(grid, charts.PaletteSequential)
	h.SetSize(40, 10)
	h.SetTheme(tuikit.DefaultTheme())

	out1 := h.View()
	out2 := h.View()
	if out1 != out2 {
		t.Fatal("heatmap rendering is not deterministic")
	}
}
