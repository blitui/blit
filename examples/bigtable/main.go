// Package main demonstrates the virtualized Table over 1,000,000 rows.
//
// The example defines a TableRowProvider that lazily formats rows on demand
// — no 1M-element backing slice is ever allocated. Scrolling through the full
// dataset stays smooth because only the visible window (~20 rows) is
// materialized per frame. The built-in "/" search integrates with the
// provider via OnFilterChange: typing a query swaps the provider into a
// filtered mode that walks the logical dataset and returns matching rows.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	blit "github.com/blitui/blit"
	"github.com/charmbracelet/lipgloss"
)

// totalRows is the logical dataset size. A million rows is ~40 MB if you were
// to materialize every cell as a string; here we materialize only what's on
// screen, so memory stays tiny regardless of this value.
const totalRows = 1_000_000

// statuses cycles through a small set of status labels so row styling has
// something to react to.
var statuses = []string{"online", "pending", "offline", "error"}

// bigtableProvider lazily generates rows for the virtualized Table. When
// filter is non-empty, Rows walks the synthetic dataset and returns only
// matching rows for the requested window.
type bigtableProvider struct {
	filter string
}

func (p *bigtableProvider) rowAt(i int) blit.Row {
	status := statuses[i%len(statuses)]
	return blit.Row{
		strconv.Itoa(i),
		"user-" + strconv.Itoa(i),
		status,
		strconv.Itoa((i * 37) % 10_000),
	}
}

func (p *bigtableProvider) matches(r blit.Row) bool {
	if p.filter == "" {
		return true
	}
	q := strings.ToLower(p.filter)
	for _, cell := range r {
		if strings.Contains(strings.ToLower(cell), q) {
			return true
		}
	}
	return false
}

// Len — when unfiltered this is totalRows. When a filter is active we cap the
// reported total at the number of matches found while walking. For the
// example we walk the full synthetic set once per filter change to compute
// match count; in a real app you would cache this or ask your data source.
func (p *bigtableProvider) Len() int {
	if p.filter == "" {
		return totalRows
	}
	// Synthetic dataset is cheap to scan — count matches on demand. A real
	// provider would hit its own index.
	n := 0
	for i := 0; i < totalRows; i++ {
		if p.matches(p.rowAt(i)) {
			n++
		}
	}
	return n
}

func (p *bigtableProvider) Rows(offset, limit int) []blit.Row {
	out := make([]blit.Row, 0, limit)
	if p.filter == "" {
		for i := offset; i < offset+limit && i < totalRows; i++ {
			out = append(out, p.rowAt(i))
		}
		return out
	}
	// Filtered path: walk until we reach offset matches, then collect limit.
	seen := 0
	for i := 0; i < totalRows && len(out) < limit; i++ {
		r := p.rowAt(i)
		if !p.matches(r) {
			continue
		}
		if seen >= offset {
			out = append(out, r)
		}
		seen++
	}
	return out
}

func main() {
	cols := []blit.Column{
		{Title: "ID", Width: 8, Align: blit.Right, MaxWidth: 10},
		{Title: "User", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Score", Width: 10, Align: blit.Right},
	}

	provider := &bigtableProvider{}

	// Color the status cell per value. Cell renderers still run in virtual
	// mode — only on the visible slice, so they stay fast.
	cellRenderer := func(row blit.Row, colIdx int, isCursor bool, theme blit.Theme) string {
		if colIdx >= len(row) {
			return ""
		}
		cell := row[colIdx]
		if colIdx == 2 {
			color := theme.Muted
			switch cell {
			case "online":
				color = theme.Positive
			case "pending":
				color = theme.Accent
			case "error":
				color = theme.Negative
			}
			return lipgloss.NewStyle().Foreground(color).Render(cell)
		}
		return cell
	}

	var tbl *blit.Table
	tbl = blit.NewTable(cols, nil, blit.TableOpts{
		Virtual:      true,
		RowProvider:  provider,
		Filterable:   true,
		CellRenderer: cellRenderer,
		OnFilterChange: func(q string) {
			// Update the provider filter in place. The next View() frame will
			// re-query the provider with the new filter.
			provider.filter = q
			tbl.SetCursor(0)
		},
	})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("bigtable", tbl),
		blit.WithStatusBar(
			func() string { return " ↑/↓ j/k  g/G  / search  ? help  q quit" },
			func() string {
				return fmt.Sprintf(" %d / %d rows ",
					tbl.VisibleRowCount(),
					totalRows,
				)
			},
		),
		blit.WithHelp(),
	)

	if _, err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
