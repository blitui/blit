# Charts

The `charts` sub-package provides terminal chart components for data visualization. All chart types implement `Component` and `Themed`.

```go
import "github.com/moneycaringcoder/tuikit-go/charts"
```

## Chart Types

### Bar

Horizontal or vertical bar chart with labeled categories.

```go
bar := charts.NewBar(charts.BarOpts{
    Data:   []charts.BarItem{{Label: "Go", Value: 42}, {Label: "Rust", Value: 28}},
    Height: 12,
})
```

### Line

Line chart with optional multi-series support.

```go
line := charts.NewLine(charts.LineOpts{
    Series: []charts.Series{{Label: "CPU", Data: cpuData}},
    Height: 10,
    Width:  60,
})
```

### Ring

Donut/ring chart for proportional data.

```go
ring := charts.NewRing(charts.RingOpts{
    Segments: []charts.Segment{{Label: "Used", Value: 72}, {Label: "Free", Value: 28}},
    Radius:   6,
})
```

### Gauge

Single-value gauge with configurable range.

```go
gauge := charts.NewGauge(charts.GaugeOpts{
    Value: 75,
    Max:   100,
    Width: 40,
    Label: "Memory",
})
```

### Heatmap

Grid heatmap for matrix data.

```go
heatmap := charts.NewHeatmap(charts.HeatmapOpts{
    Data:    matrix,
    XLabels: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
    YLabels: []string{"Morning", "Afternoon", "Evening"},
})
```

## Sparkline (Root Package)

The root `tuikit` package also provides an inline sparkline for embedding in table cells or status bars:

```go
tuikit.Sparkline([]float64{1, 4, 2, 8, 5, 7, 3}, 10)
```

## Example

```bash
go run ./examples/charts/
```

## API Reference

Full API documentation: [pkg.go.dev/github.com/moneycaringcoder/tuikit-go/charts](https://pkg.go.dev/github.com/moneycaringcoder/tuikit-go/charts)
