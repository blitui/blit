// Coverage overlay for btest.
//
// G1: runCoverage runs "go test -coverprofile=.blit/coverage.out ./..." and
//
//	parses the resulting coverage profile.
//
// G2: renderCoveragePanel prints a summary to the terminal: total %, top 5
//
//	least-covered files, and a mini histogram of per-file coverage buckets.
//
// G3: readCoverage reads the most recent .blit/coverage.out without re-running
//
//	tests and renders the same panel.
//
// G4: Both render paths dogfood the charts subpackage (Bar histogram + Gauge).
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/blitui/blit/charts"
)

const coverageOut = ".blit/coverage.out"

// fileCoverage holds coverage data for a single source file.
type fileCoverage struct {
	Name    string
	Covered int // statements covered
	Total   int // total statements
}

// Pct returns the coverage percentage (0–100), or 100 when Total == 0.
func (f fileCoverage) Pct() float64 {
	if f.Total == 0 {
		return 100
	}
	return float64(f.Covered) / float64(f.Total) * 100
}

// runCoverage implements G1: runs go test with -coverprofile and then renders
// the coverage panel (G2). packages defaults to "./...".
func runCoverage(packages []string) int {
	if err := os.MkdirAll(".blit", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[blit] coverage: mkdir .blit: %v\n", err)
		return 1
	}

	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	args := append([]string{"test", "-coverprofile=" + coverageOut}, packages...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			// Still try to render whatever profile was written.
			_ = printCoveragePanel(coverageOut)
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "[blit] coverage: go test failed: %v\n", err)
		return 1
	}

	if err := printCoveragePanel(coverageOut); err != nil {
		fmt.Fprintf(os.Stderr, "[blit] coverage: render failed: %v\n", err)
		return 1
	}
	return 0
}

// readCoverage implements G3: renders the panel from the most recent profile
// without re-running tests.
func readCoverage() int {
	path := latestCoverageFile()
	if path == "" {
		fmt.Fprintln(os.Stderr, "[blit] coverage: no coverage file found — run blit -coverage first")
		return 1
	}
	if err := printCoveragePanel(path); err != nil {
		fmt.Fprintf(os.Stderr, "[blit] coverage: render failed: %v\n", err)
		return 1
	}
	return 0
}

// latestCoverageFile returns the path of the most recently modified *.out file
// under .blit, falling back to the canonical coverageOut path.
func latestCoverageFile() string {
	if _, err := os.Stat(coverageOut); err == nil {
		return coverageOut
	}
	// Search .blit subtree for any .out file.
	var best string
	var bestMod int64
	_ = filepath.WalkDir(".blit", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".out") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().UnixNano() > bestMod {
			bestMod = info.ModTime().UnixNano()
			best = path
		}
		return nil
	})
	return best
}

// parseCoverProfile parses a go cover profile and returns per-file coverage.
// The profile format is:
//
//	mode: set|count|atomic
//	<file>:<startLine>.<startCol>,<endLine>.<endCol> <stmts> <count>
func parseCoverProfile(path string) ([]fileCoverage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	type accum struct{ covered, total int }
	byFile := map[string]*accum{}
	var order []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		// <block> <stmts> <count>
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}
		block := parts[0]
		stmts, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		// Extract filename: everything before the last ':'
		colonIdx := strings.LastIndex(block, ":")
		if colonIdx < 0 {
			continue
		}
		fileName := block[:colonIdx]

		a, ok := byFile[fileName]
		if !ok {
			a = &accum{}
			byFile[fileName] = a
			order = append(order, fileName)
		}
		a.total += stmts
		if count > 0 {
			a.covered += stmts
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	result := make([]fileCoverage, 0, len(order))
	for _, name := range order {
		a := byFile[name]
		result = append(result, fileCoverage{Name: name, Covered: a.covered, Total: a.total})
	}
	return result, nil
}

// totalCoverage returns the overall coverage percentage across all files.
func totalCoverage(files []fileCoverage) float64 {
	var covered, total int
	for _, f := range files {
		covered += f.Covered
		total += f.Total
	}
	if total == 0 {
		return 100
	}
	return float64(covered) / float64(total) * 100
}

// printCoveragePanel implements G2/G4: renders a coverage summary using
// the charts subpackage (Bar histogram + Gauge).
func printCoveragePanel(profilePath string) error {
	files, err := parseCoverProfile(profilePath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("[blit] coverage: profile is empty")
		return nil
	}

	total := totalCoverage(files)

	// ── Gauge (G4: dogfoods charts.Gauge) ────────────────────────────────────
	gauge := charts.NewGauge(total, 100, []float64{50, 80}, "coverage")
	gauge.SetSize(40, 5)
	fmt.Println(gauge.View())
	fmt.Printf("  Total coverage: %.1f%%\n\n", total)

	// ── Top 5 least-covered files ─────────────────────────────────────────────
	sorted := make([]fileCoverage, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Pct() < sorted[j].Pct()
	})

	n := 5
	if len(sorted) < n {
		n = len(sorted)
	}

	fmt.Println("  Least covered files:")
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	for i := 0; i < n; i++ {
		fc := sorted[i]
		name := shortName(fc.Name)
		pct := fc.Pct()
		color := lipgloss.Color("196")
		if pct >= 80 {
			color = lipgloss.Color("46")
		} else if pct >= 50 {
			color = lipgloss.Color("226")
		}
		pctStyle := lipgloss.NewStyle().Foreground(color)
		fmt.Printf("  %s %s\n",
			labelStyle.Render(fmt.Sprintf("%-50s", name)),
			pctStyle.Render(fmt.Sprintf("%5.1f%%", pct)),
		)
	}
	fmt.Println()

	// ── Histogram of per-file coverage (G4: dogfoods charts.Bar) ─────────────
	// Bucket files into 10% bands: [0,10), [10,20), ..., [90,100]
	buckets := make([]float64, 10)
	for _, fc := range files {
		idx := int(fc.Pct() / 10)
		if idx >= 10 {
			idx = 9
		}
		buckets[idx]++
	}
	labels := []string{"0", "10", "20", "30", "40", "50", "60", "70", "80", "90"}

	bar := charts.NewBar(buckets, labels, false)
	bar.SetSize(60, 8)
	fmt.Println("  Per-file coverage distribution (% band → file count):")
	fmt.Println(bar.View())

	return nil
}

// shortName trims the module path prefix from a coverage file name for display.
func shortName(name string) string {
	// Coverage profiles use "module/pkg/file.go" — strip leading module path.
	if idx := strings.Index(name, "/"); idx >= 0 {
		// Keep everything after the first path segment (module name).
		return name[idx+1:]
	}
	return name
}
