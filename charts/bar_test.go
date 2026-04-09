package charts_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/charts"
)

func TestBarVerticalRender(t *testing.T) {
	b := charts.NewBar([]float64{1, 2, 3, 4, 5}, []string{"a", "b", "c", "d", "e"}, false)
	b.SetSize(20, 10)
	b.SetTheme(tuikit.DefaultTheme())

	out := b.View()
	if out == "" {
		t.Fatal("expected non-empty output for vertical bar chart")
	}
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d", len(lines))
	}
}

func TestBarHorizontalRender(t *testing.T) {
	b := charts.NewBar([]float64{10, 50, 100}, []string{"low", "mid", "high"}, true)
	b.SetSize(30, 5)
	b.SetTheme(tuikit.DefaultTheme())

	out := b.View()
	if out == "" {
		t.Fatal("expected non-empty output for horizontal bar chart")
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for 3 data points, got %d", len(lines))
	}
}

func TestBarGradient(t *testing.T) {
	g := &tuikit.Gradient{
		Start: lipgloss.Color("#ff0000"),
		End:   lipgloss.Color("#0000ff"),
	}
	b := charts.NewBar([]float64{1, 5, 3}, nil, false)
	b.Gradient = g
	b.SetSize(15, 8)
	b.SetTheme(tuikit.DefaultTheme())

	out := b.View()
	if out == "" {
		t.Fatal("expected non-empty output for gradient bar chart")
	}
}

func TestBarEmptyData(t *testing.T) {
	b := charts.NewBar(nil, nil, false)
	b.SetSize(20, 10)
	out := b.View()
	if out != "" {
		t.Fatalf("expected empty output for nil data, got %q", out)
	}
}

func TestBarZeroSize(t *testing.T) {
	b := charts.NewBar([]float64{1, 2, 3}, nil, false)
	b.SetSize(0, 0)
	out := b.View()
	if out != "" {
		t.Fatalf("expected empty output for zero size, got %q", out)
	}
}

func TestBarCustomColors(t *testing.T) {
	b := charts.NewBar([]float64{1, 2, 3}, nil, false)
	b.Colors = []lipgloss.Color{"#ff0000", "#00ff00"}
	b.SetSize(15, 8)
	b.SetTheme(tuikit.DefaultTheme())
	out := b.View()
	if out == "" {
		t.Fatal("expected non-empty output with custom colors")
	}
}

func TestBarDeterministic(t *testing.T) {
	b := charts.NewBar([]float64{3, 1, 4, 1, 5}, []string{"a", "b", "c", "d", "e"}, false)
	b.SetSize(25, 10)
	b.SetTheme(tuikit.DefaultTheme())

	out1 := b.View()
	out2 := b.View()
	if out1 != out2 {
		t.Fatal("bar chart rendering is not deterministic")
	}
}
