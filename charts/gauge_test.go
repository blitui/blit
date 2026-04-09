package charts_test

import (
	"strings"
	"testing"

	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/charts"
)

func TestGaugeRender(t *testing.T) {
	g := charts.NewGauge(75, 100, []float64{60, 80}, "Load")
	g.SetSize(30, 10)
	g.SetTheme(tuikit.DefaultTheme())

	out := g.View()
	if out == "" {
		t.Fatal("expected non-empty output for gauge chart")
	}
}

func TestGaugeZeroValue(t *testing.T) {
	g := charts.NewGauge(0, 100, nil, "")
	g.SetSize(30, 10)
	g.SetTheme(tuikit.DefaultTheme())
	out := g.View()
	if out == "" {
		t.Fatal("expected non-empty output for zero-value gauge")
	}
}

func TestGaugeMaxValue(t *testing.T) {
	g := charts.NewGauge(100, 100, nil, "")
	g.SetSize(30, 10)
	g.SetTheme(tuikit.DefaultTheme())
	out := g.View()
	if out == "" {
		t.Fatal("expected non-empty output for max-value gauge")
	}
}

func TestGaugeTooSmall(t *testing.T) {
	g := charts.NewGauge(50, 100, nil, "")
	g.SetSize(2, 1)
	out := g.View()
	if out != "" {
		t.Fatalf("expected empty output for too-small gauge, got %q", out)
	}
}

func TestGaugeLabelInOutput(t *testing.T) {
	g := charts.NewGauge(42, 100, []float64{60, 80}, "Speed")
	g.SetSize(40, 12)
	g.SetTheme(tuikit.DefaultTheme())
	out := g.View()
	if !strings.Contains(out, "Speed") {
		t.Errorf("expected label 'Speed' in gauge output")
	}
}

func TestGaugeDeterministic(t *testing.T) {
	g := charts.NewGauge(65, 100, []float64{50, 75}, "CPU")
	g.SetSize(40, 12)
	g.SetTheme(tuikit.DefaultTheme())

	out1 := g.View()
	out2 := g.View()
	if out1 != out2 {
		t.Fatal("gauge chart rendering is not deterministic")
	}
}
